package swap

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

// NewSwapper returns the Swapper instance
func NewSwapper(db *gorm.DB, cfg *util.Config, bscClient, ethClient *ethclient.Client) (*Swapper, error) {
	tokens := make([]model.Token, 0)
	db.Where("available = ?", true).Find(&tokens)

	tokenInstances, err := buildTokenInstance(tokens)
	if err != nil {
		return nil, err
	}
	bscContractAddrToSymbol := make(map[string]string)
	ethContractAddrToSymbol := make(map[string]string)
	for _, token := range tokens {
		bscContractAddrToSymbol[token.BSCContractAddr] = token.Symbol
		ethContractAddrToSymbol[token.ETHContractAddr] = token.Symbol
	}

	swapper := &Swapper{
		DB:                      db,
		Config:                  cfg,
		BSCClient:               bscClient,
		ETHClient:               ethClient,
		TokenInstances:          tokenInstances,
		BSCContractAddrToSymbol: bscContractAddrToSymbol,
		ETHContractAddrToSymbol: ethContractAddrToSymbol,
	}
	err = swapper.syncTokenSendAddress()
	if err != nil {
		panic(err)
	}

	return swapper, nil
}

func (swapper *Swapper) syncTokenSendAddress() error {
	tokens := make([]model.Token, 0)
	swapper.DB.Where("available = ?", true).Find(&tokens)
	for _, token := range tokens {
		ethTxSender, err := getAddress(token.ETHPrivateKey)
		if err != nil {
			return err
		}
		bscTxSender, err := getAddress(token.BSCPrivateKey)
		if err != nil {
			return err
		}
		err = swapper.DB.Model(model.Token{}).Where("symbol = ?", token.Symbol).Updates(
			map[string]interface{}{
				"bsc_send_addr": bscTxSender.String(),
				"eth_send_addr": ethTxSender.String(),
				"updated_at":    time.Now().Unix(),
			}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (swapper *Swapper) Start() {
	go swapper.monitorSwapRequestDaemon()
	go swapper.confirmSwapRequestDaemon()
	go swapper.handleSwapDaemon()
	go swapper.monitorUncertainSwapTxDaemon()
	go swapper.trackSwapTxDaemon()
	go swapper.alertDaemon()
}

func (swapper *Swapper) monitorSwapRequestDaemon() {
	for {
		txEventLogs := make([]model.TxEventLog, BatchSize)
		swapper.DB.Where("phase = ?", model.SeenSwapRequest).Order("height asc").Limit(BatchSize).Find(&txEventLogs)

		if len(txEventLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		for _, txEventLog := range txEventLogs {
			swap := swapper.createSwap(&txEventLog)

			err := swapper.insertSwapToDB(swap)
			if err != nil {
				util.Logger.Errorf("failed to persistent swap: %s", err.Error())
				util.SendTelegramMessage(fmt.Sprintf("failed to persistent swap: %s", err.Error()))
				continue
			}
			err = swapper.DB.Model(model.TxEventLog{}).Where("tx_hash = ?", swap.DepositTxHash).Updates(
				map[string]interface{}{
					"phase":       model.ConfirmSwapRequest,
					"update_time": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update %s table failed: %s", model.TxEventLog{}.TableName(), err.Error())
			}
		}
	}
}

func (swapper *Swapper) createSwap(txEventLog *model.TxEventLog) *model.Swap {
	sponsor := txEventLog.FromAddress
	amount := txEventLog.Amount
	depositTxHash := txEventLog.TxHash
	swapDirection := SwapEth2BSC
	if txEventLog.Chain == common.ChainBSC {
		swapDirection = SwapBSC2Eth
	}

	symbol := ""
	decimals := 0
	swapStatus := SwapQuoteRejected
	err := func() error {
		var ok bool
		if txEventLog.Chain == common.ChainETH {
			symbol, ok = swapper.ETHContractAddrToSymbol[strings.ToLower(txEventLog.ContractAddress)]
			if !ok {
				return fmt.Errorf("unsupported eth token contract address: %s", txEventLog.ContractAddress)
			}
		} else {
			symbol, ok = swapper.BSCContractAddrToSymbol[strings.ToLower(txEventLog.ContractAddress)]
			if !ok {
				return fmt.Errorf("unsupported bsc token contract address: %s", txEventLog.ContractAddress)
			}
		}
		tokenInstance, ok := swapper.TokenInstances[symbol]
		if !ok {
			return fmt.Errorf("unsupported token symbol %s", symbol)
		}
		decimals = tokenInstance.Decimals

		swapAmount := big.NewInt(0)
		_, ok = swapAmount.SetString(txEventLog.Amount, 10)
		if !ok {
			return fmt.Errorf("unrecongnized swap amount: %s", txEventLog.Amount)
		}
		if swapAmount.Cmp(tokenInstance.LowBound) < 0 || swapAmount.Cmp(tokenInstance.UpperBound) > 0 {
			return fmt.Errorf("swap amount is out of bound, expected bound [%s, %s]", tokenInstance.LowBound.String(), tokenInstance.UpperBound.String())
		}

		swapStatus = SwapTokenReceived
		return nil
	}()

	log := ""
	if err != nil {
		log = err.Error()
	}

	swap := &model.Swap{
		Status:         swapStatus,
		Sponsor:        sponsor,
		Symbol:         symbol,
		Amount:         amount,
		Decimals:       decimals,
		Direction:      swapDirection,
		DepositTxHash:  depositTxHash,
		WithdrawTxHash: "",
		Log:            log,
	}

	return swap
}

func (swapper *Swapper) confirmSwapRequestDaemon() {
	for {
		txEventLogs := make([]model.TxEventLog, BatchSize)
		swapper.DB.Where("status = ? and phase = ?", model.TxStatusConfirmed, model.ConfirmSwapRequest).
			Order("height asc").Limit(BatchSize).Find(&txEventLogs)

		if len(txEventLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		util.Logger.Infof("found %d confirmed event logs", len(txEventLogs))

		for _, txEventLog := range txEventLogs {
			swap := model.Swap{}
			swapper.DB.Where("deposit_tx_hash = ?", txEventLog.TxHash).First(&swap)
			if swap.DepositTxHash == "" {
				util.Logger.Errorf(fmt.Sprintf("unexpected error, can't find swap by deposit hash: %s", txEventLog.TxHash))
				util.SendTelegramMessage(fmt.Sprintf("unexpected error, can't find swap by deposit hash: %s", txEventLog.TxHash))
				continue
			}
			if swap.Status == SwapQuoteRejected {
				util.Logger.Debugf("swap is rejected, deposit txHash: %s", swap.DepositTxHash)
			} else {
				err := swapper.DB.Model(model.Swap{}).Where("deposit_tx_hash = ?", txEventLog.TxHash).Updates(
					map[string]interface{}{
						"status":     SwapQuoteConfirmed,
						"updated_at": time.Now().Unix(),
					}).Error
				if err != nil {
					util.Logger.Errorf("update %s table failed: %s", model.Swap{}.TableName(), err.Error())
					util.SendTelegramMessage(fmt.Sprintf("update %s table failed: %s", model.Swap{}.TableName(), err.Error()))
				}
			}

			err := swapper.DB.Model(model.TxEventLog{}).Where("tx_hash = ?", swap.DepositTxHash).Updates(
				map[string]interface{}{
					"phase":       model.AckSwapRequest,
					"update_time": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update table %s failed: %s", model.TxEventLog{}.TableName(), err.Error())
			}
		}
	}
}

func (swapper *Swapper) handleSwapDaemon() {
	for {
		swaps := make([]model.Swap, BatchSize)
		swapper.DB.Where("status = ?", SwapQuoteConfirmed).Order("id asc").Limit(BatchSize).Find(&swaps)

		if len(swaps) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		util.Logger.Infof("found %d confirmed swap requests", len(swaps))

		for _, swap := range swaps {
			err := swapper.DB.Model(model.Swap{}).Where("id = ?", swap.ID).Updates(
				map[string]interface{}{
					"status":     SwapQuoteSending,
					"updated_at": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update %s table failed: %s", model.Swap{}.TableName(), err.Error())
				util.SendTelegramMessage(fmt.Sprintf("update %s table failed: %s", model.Swap{}.TableName(), err.Error()))
			}

			swapTx, err := swapper.doSwap(&swap)
			if err != nil {
				util.Logger.Errorf("do swap failure: %s", err.Error())
				util.SendTelegramMessage(fmt.Sprintf("do swap failure: %s", err.Error()))
				continue
			}
			err = swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swap.DepositTxHash).Updates(
				map[string]interface{}{
					"status":     model.WithdrawTxSent,
					"updated_at": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
			}
			err = swapper.DB.Model(model.Swap{}).Where("id = ?", swap.ID).Updates(
				map[string]interface{}{
					"status":           SwapSent,
					"withdraw_tx_hash": swapTx.WithdrawTxHash,
					"updated_at":       time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update %s table failed: %s", model.Swap{}.TableName(), err.Error())
				util.SendTelegramMessage(fmt.Sprintf("update %s table failed: %s", model.Swap{}.TableName(), err.Error()))
			}
		}
	}
}

func (swapper *Swapper) doSwap(swap *model.Swap) (*model.SwapTx, error) {
	tokenInstance, ok := swapper.TokenInstances[swap.Symbol]
	if !ok {
		return nil, fmt.Errorf("unsupported token %s", swap.Symbol)
	}
	amount := big.NewInt(0)
	_, ok = amount.SetString(swap.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid swap amount: %s", swap.Amount)
	}
	txInput, err := abiEncodeTransfer(ethcom.HexToAddress(swap.Sponsor), amount)
	if err != nil {
		return nil, err
	}

	if swap.Direction == SwapEth2BSC {
		signedTx, err := buildSignedTransaction(swapper.BSCClient, tokenInstance.BSCPrivateKey, tokenInstance.BSCContractAddr, txInput)
		if err != nil {
			return nil, err
		}
		swapTx := &model.SwapTx{
			Direction:      SwapEth2BSC,
			DepositTxHash:  swap.DepositTxHash,
			WithdrawTxHash: signedTx.Hash().String(),
			GasPrice:       signedTx.GasPrice().String(),
			Status:         model.WithdrawTxCreated,
		}
		err = swapper.insertSwapTxToDB(swapTx)
		if err != nil {
			return nil, err
		}
		err = swapper.BSCClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			util.Logger.Errorf("broadcast tx to BSC error: %s", err.Error())
			return nil, err
		}
		util.Logger.Infof("Send transaction to BSC, %s/%s", swapper.Config.ChainConfig.BSCExplorerUrl, signedTx.Hash().String())
		return swapTx, nil
	} else {
		signedTx, err := buildSignedTransaction(swapper.ETHClient, tokenInstance.ETHPrivateKey, tokenInstance.ETHContractAddr, txInput)
		if err != nil {
			return nil, err
		}
		swapTx := &model.SwapTx{
			Direction:      SwapBSC2Eth,
			DepositTxHash:  swap.DepositTxHash,
			GasPrice:       signedTx.GasPrice().String(),
			WithdrawTxHash: signedTx.Hash().String(),
			Status:         model.WithdrawTxCreated,
		}
		err = swapper.insertSwapTxToDB(swapTx)
		if err != nil {
			return nil, err
		}
		err = swapper.ETHClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			util.Logger.Errorf("broadcast tx to ETH error: %s", err.Error())
			return nil, err
		} else {
			util.Logger.Infof("Send transaction to ETH, %s/%s", swapper.Config.ChainConfig.ETHExplorerUrl, signedTx.Hash().String())
		}
		return swapTx, nil
	}
}

func (swapper *Swapper) monitorUncertainSwapTxDaemon() {
	for {
		time.Sleep(SleepTime * time.Second)

		swapTxs := make([]model.SwapTx, BatchSize)
		swapper.DB.Where("status = ? and track_retry_counter >= ?", model.WithdrawTxCreated, MaxTrackerRetry).Order("id asc").Limit(BatchSize).Find(&swapTxs)
		if len(swapTxs) > 0 {
			util.Logger.Infof("%d withdraw tx are missing, retry to handle these swap", len(swapTxs))
		}

		for _, swapTx := range swapTxs {
			util.Logger.Errorf("the withdraw tx is missing, retry this swap, deposit hash %s", swapTx.DepositTxHash)
			util.SendTelegramMessage(fmt.Sprintf("the withdraw tx is missing, retry this swap, deposit hash %s", swapTx.DepositTxHash))
			err := swapper.DB.Model(model.Swap{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
				map[string]interface{}{
					"status":     SwapQuoteConfirmed, // SwapQuoteSending -> SwapQuoteConfirmed, retry
					"updated_at": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update %s table failed: %s", model.Swap{}.TableName(), err.Error())
				util.SendTelegramMessage(fmt.Sprintf("update %s table failed: %s", model.Swap{}.TableName(), err.Error()))
			}
		}

		swapTxs = swapTxs[:0]
		swapper.DB.Where("status = ? and track_retry_counter < ?", model.WithdrawTxCreated, MaxTrackerRetry).Order("id asc").Limit(BatchSize).Find(&swapTxs)
		if len(swapTxs) > 0 {
			util.Logger.Infof("Track uncertain withdraw tx %d", len(swapTxs))
		}

		for _, swapTx := range swapTxs {
			var tx *types.Transaction
			if swapTx.Direction == SwapEth2BSC {
				tx, _, _ = swapper.BSCClient.TransactionByHash(context.Background(), ethcom.HexToHash(swapTx.WithdrawTxHash))
			} else {
				tx, _, _ = swapper.ETHClient.TransactionByHash(context.Background(), ethcom.HexToHash(swapTx.WithdrawTxHash))
			}
			if tx != nil {
				err := swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
					map[string]interface{}{
						"status":     model.WithdrawTxSent,
						"updated_at": time.Now().Unix(),
					}).Error
				if err != nil {
					util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
				}

				err = swapper.DB.Model(model.Swap{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
					map[string]interface{}{
						"status":           SwapSent,
						"withdraw_tx_hash": swapTx.WithdrawTxHash,
						"updated_at":       time.Now().Unix(),
					}).Error
				if err != nil {
					util.Logger.Errorf("update %s table failed: %s", model.Swap{}.TableName(), err.Error())
					util.SendTelegramMessage(fmt.Sprintf("update %s table failed: %s", model.Swap{}.TableName(), err.Error()))
				}
			} else {
				err := swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
					map[string]interface{}{
						"track_retry_counter": gorm.Expr("track_retry_counter + 1"),
						"updated_at":          time.Now().Unix(),
					}).Error
				if err != nil {
					util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
				}
			}
		}
	}
}

func (swapper *Swapper) trackSwapTxDaemon() {
	for {
		time.Sleep(SleepTime * time.Second)

		swapTxs := make([]model.SwapTx, BatchSize)
		swapper.DB.Where("status = ?", model.WithdrawTxSent).Order("id asc").Limit(BatchSize).Find(&swapTxs)

		if len(swapTxs) > 0 {
			util.Logger.Infof("Track %d swap txs", len(swapTxs))
		}

		for _, swapTx := range swapTxs {
			gasPrice := big.NewInt(0)
			gasPrice.SetString(swapTx.GasPrice, 10)

			var client *ethclient.Client
			if swapTx.Direction == SwapBSC2Eth {
				client = swapper.ETHClient
			} else {
				client = swapper.BSCClient
			}
			block, err := client.BlockByNumber(context.Background(), nil)
			if err != nil {
				util.Logger.Debugf("ETH, query block failed: %s", err.Error())
				continue
			}
			txRecipient, err := client.TransactionReceipt(context.Background(), ethcom.HexToHash(swapTx.WithdrawTxHash))
			if err != nil {
				util.Logger.Debugf("ETH, query tx failed: %s", err.Error())
				continue
			}

			if block.Number().Int64() < txRecipient.BlockNumber.Int64()+swapper.Config.ChainConfig.ETHConfirmNum {
				continue
			}

			txFee := big.NewInt(1).Mul(gasPrice, big.NewInt(int64(txRecipient.GasUsed))).String()
			if txRecipient.Status == TxFailedStatus {
				util.SendTelegramMessage(fmt.Sprintf("withdraw tx is failed, txHash: %s", txRecipient.TxHash))
				err = swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
					map[string]interface{}{
						"status":              model.WithdrawTxFailed,
						"height":              txRecipient.BlockNumber.Int64(),
						"consumed_fee_amount": txFee,
						"updated_at":          time.Now().Unix(),
					}).Error
				if err != nil {
					util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
					util.SendTelegramMessage(fmt.Sprintf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error()))
				}
				err = swapper.DB.Model(model.Swap{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
					map[string]interface{}{
						"status":     SwapSendFailed,
						"log":        "withdraw tx is failed",
						"updated_at": time.Now().Unix(),
					}).Error
				if err != nil {
					util.Logger.Errorf("update table %s failed: %s", model.Swap{}.TableName(), err.Error())
					util.SendTelegramMessage(fmt.Sprintf("update table %s failed: %s", model.Swap{}.TableName(), err.Error()))
				}
				continue
			}

			err = swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
				map[string]interface{}{
					"status":              model.WithdrawTxSuccess,
					"height":              txRecipient.BlockNumber.Int64(),
					"consumed_fee_amount": txFee,
					"updated_at":          time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
				util.SendTelegramMessage(fmt.Sprintf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error()))
			}

			err = swapper.DB.Model(model.Swap{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
				map[string]interface{}{
					"status":     SwapSuccess,
					"updated_at": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update table %s failed: %s", model.Swap{}.TableName(), err.Error())
				util.SendTelegramMessage(fmt.Sprintf("update table %s failed: %s", model.Swap{}.TableName(), err.Error()))
			}
		}
	}
}

func (swapper *Swapper) alertDaemon() {
	// TODO
}

func (swapper *Swapper) insertSwapToDB(data *model.Swap) error {
	tx := swapper.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(data).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (swapper *Swapper) insertSwapTxToDB(data *model.SwapTx) error {
	tx := swapper.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(data).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (swapper *Swapper) AddToken(token *model.Token) error {
	bscPriKey, err := getBSCPrivateKey(token)
	if err != nil {
		return err
	}

	ethPriKey, err := getETHPrivateKey(token)
	if err != nil {
		return err
	}
	lowBound := big.NewInt(0)
	_, ok := lowBound.SetString(token.LowBound, 10)
	if !ok {
		return fmt.Errorf("invalid lowBound amount: %s", token.LowBound)
	}
	upperBound := big.NewInt(0)
	_, ok = upperBound.SetString(token.UpperBound, 10)
	if !ok {
		return fmt.Errorf("invalid upperBound amount: %s", token.LowBound)
	}

	swapper.TokenInstances[token.Symbol] = &TokenInstance{
		Symbol:          token.Symbol,
		Name:            token.Name,
		Decimals:        token.Decimals,
		LowBound:        lowBound,
		UpperBound:      upperBound,
		BSCPrivateKey:   bscPriKey,
		BSCContractAddr: ethcom.HexToAddress(token.BSCContractAddr),
		ETHPrivateKey:   ethPriKey,
		ETHContractAddr: ethcom.HexToAddress(token.ETHContractAddr),
	}
	return nil
}

func (swapper *Swapper) RemoveToken(token *model.Token) {
	delete(swapper.TokenInstances, token.Symbol)
	delete(swapper.BSCContractAddrToSymbol, strings.ToLower(token.BSCContractAddr))
	delete(swapper.ETHContractAddrToSymbol, strings.ToLower(token.ETHContractAddr))
}
