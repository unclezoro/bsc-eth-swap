package swap

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
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
	go swapper.handleSwapDaemon()
	go swapper.sendTokenDaemon()
	go swapper.broadcastRetrySwapTxDaemon()
	go swapper.trackSwapTxDaemon()
	go swapper.alertDaemon()
}

func (swapper *Swapper) handleSwapDaemon() {
	for {
		txEventLogs := make([]model.TxEventLog, BatchSize)
		swapper.DB.Where("phase = ?", model.SeenSwapRequest).Order("height asc").Limit(BatchSize).Find(&txEventLogs)

		if len(txEventLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		for _, txEventLog := range txEventLogs {
			swap, err := swapper.createSwap(&txEventLog)
			if err != nil {
				util.Logger.Errorf("Encounter failure in create swap: %s", err.Error())
			}
			if swap != nil {
				// mark tx_event_log as processed
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
}

func (swapper *Swapper) createSwap(txEventLog *model.TxEventLog) (*model.Swap, error) {
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

	err = swapper.insertSwapToDB(swap)
	if err != nil {
		return nil, fmt.Errorf("failed to persistent swap: %s", err.Error())
	}
	return swap, nil
}

func (swapper *Swapper) sendTokenDaemon() {
	for {
		txEventLogs := make([]model.TxEventLog, BatchSize)
		swapper.DB.Where("status = ? and phase = ?", model.TxStatusConfirmed, model.ConfirmSwapRequest).
			Order("height asc").Limit(BatchSize).Find(&txEventLogs)

		if len(txEventLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		util.Logger.Infof("found %d confirmed swap request", len(txEventLogs))

		for _, txEventLog := range txEventLogs {
			swap := model.Swap{}
			swapper.DB.Where("deposit_tx_hash = ?", txEventLog.TxHash).First(&swap)
			if swap.Status == SwapQuoteRejected {
				util.Logger.Debugf("swap is rejected, deposit txHash: %s", swap.DepositTxHash)
			} else {
				if swap.Direction == SwapBSC2Eth {
					util.Logger.Infof("%s swap %s:%s from BSC to ETH", swap.Sponsor, swap.Amount, swap.Symbol)
				} else {
					util.Logger.Infof("%s swap %s:%s from ETH to BSC", swap.Sponsor, swap.Amount, swap.Symbol)
				}
				err := swapper.doSwap(&swap)
				if err != nil {
					util.Logger.Errorf("doSwap failed: %s", err.Error())
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

func (swapper *Swapper) broadcastRetrySwapTxDaemon() {
	for {

		time.Sleep(SleepTime * time.Second)

		swapTxs := make([]model.SwapTx, BatchSize)
		swapper.DB.Where("status = ? and retry_counter < ?", model.WithdrawTxCreated, MaxBroadcastRetry).Order("id asc").Limit(BatchSize).Find(&swapTxs)

		if len(swapTxs) > 0 {
			util.Logger.Infof("Try to broadcast %d swap txs", len(swapTxs))
		}

		for _, swapTx := range swapTxs {
			func() {
				var signedTx types.Transaction
				txData, err := hexutil.Decode(swapTx.TxData)
				if err != nil {
					util.Logger.Errorf("txData hex decoding error: %s", err.Error())
					return
				}
				err = rlp.DecodeBytes(txData, &signedTx)
				if err != nil {
					util.Logger.Errorf("txData rlp decoding error: %s", err.Error())
					return
				}
				if swapTx.Direction == SwapEth2BSC {
					err = swapper.BSCClient.SendTransaction(context.Background(), &signedTx)
					if err != nil {
						util.Logger.Errorf("broadcast tx to BSC error: %s", err.Error())
						return
					} else {
						util.Logger.Infof("Send transaction to BSC, %s/%s", swapper.Config.ChainConfig.BSCExplorerUrl, signedTx.Hash().String())
						return
					}
				} else {
					err = swapper.ETHClient.SendTransaction(context.Background(), &signedTx)
					if err != nil {
						util.Logger.Errorf("broadcast tx to ETH error: %s", err.Error())
						return
					} else {
						util.Logger.Infof("Send transaction to ETH, %s/%s", swapper.Config.ChainConfig.ETHExplorerUrl, signedTx.Hash().String())
						return
					}
				}
			}()
			err := swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
				map[string]interface{}{
					"status":        model.WithdrawTxSent,
					"retry_counter": gorm.Expr("retry_counter + 1"),
					"updated_at":    time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
			}

			err = swapper.DB.Model(model.Swap{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
				map[string]interface{}{
					"status":     SwapSent,
					"updated_at": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update %s table failed: %s", model.Swap{}, err.Error())
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
			if swapTx.Direction == SwapBSC2Eth {
				block, err := swapper.ETHClient.BlockByNumber(context.Background(), nil)
				if err != nil {
					util.Logger.Debugf("ETH, query block failed: %s", err.Error())
					continue
				}
				txRecipient, err := swapper.ETHClient.TransactionReceipt(context.Background(), ethcom.HexToHash(swapTx.WithdrawTxHash))
				if err != nil {
					util.Logger.Debugf("ETH, query tx failed: %s", err.Error())
					continue
				}
				if txRecipient.Status == TxFailedStatus {
					err = swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
						map[string]interface{}{
							"status":     model.WithdrawTxFailed,
							"updated_at": time.Now().Unix(),
						}).Error
					if err != nil {
						util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
					}
					err = swapper.DB.Model(model.Swap{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
						map[string]interface{}{
							"status":     SwapSendFailed,
							"updated_at": time.Now().Unix(),
						}).Error
					if err != nil {
						util.Logger.Errorf("update table %s failed: %s", model.Swap{}.TableName(), err.Error())
					}
					continue
				}

				if block.Number().Int64() >= txRecipient.BlockNumber.Int64()+swapper.Config.ChainConfig.ETHConfirmNum {
					err = swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
						map[string]interface{}{
							"status":     model.WithdrawTxSuccess,
							"updated_at": time.Now().Unix(),
						}).Error
					if err != nil {
						util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
					}

					err = swapper.DB.Model(model.Swap{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
						map[string]interface{}{
							"status":     SwapSuccess,
							"updated_at": time.Now().Unix(),
						}).Error
					if err != nil {
						util.Logger.Errorf("update table %s failed: %s", model.Swap{}.TableName(), err.Error())
					}
				}
			}
		}
	}
}

func (swapper *Swapper) alertDaemon() {
	// TODO
}

func (swapper *Swapper) doSwap(swap *model.Swap) error {
	tokenInstance, ok := swapper.TokenInstances[swap.Symbol]
	if !ok {
		return fmt.Errorf("unsupported token %s", swap.Symbol)
	}
	amount := big.NewInt(0)
	_, ok = amount.SetString(swap.Amount, 10)
	if !ok {
		return fmt.Errorf("invalid swap amount: %s", swap.Amount)
	}
	txInput, err := abiEncodeTransfer(ethcom.HexToAddress(swap.Sponsor), amount)
	if err != nil {
		return err
	}

	if swap.Direction == SwapEth2BSC {
		signedTx, err := buildSignedTransaction(swapper.BSCClient, tokenInstance.BSCPrivateKey, tokenInstance.BSCContractAddr, txInput)
		if err != nil {
			return err
		}
		txData, err := rlp.EncodeToBytes(signedTx)
		if err != nil {
			return err
		}
		swapTx := &model.SwapTx{
			Direction:      SwapEth2BSC,
			DepositTxHash:  swap.DepositTxHash,
			WithdrawTxHash: signedTx.Hash().String(),
			TxData:         hexutil.Encode(txData),
			Status:         model.WithdrawTxSent,
		}
		err = swapper.insertSwapTxToDB(swapTx)
		if err != nil {
			return err
		}
		err = swapper.BSCClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			util.Logger.Errorf("broadcast tx to BSC error: %s", err.Error())
			err = swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
				map[string]interface{}{
					"status":     model.WithdrawTxCreated,
					"updated_at": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
			}
		} else {
			util.Logger.Infof("Send transaction to BSC, %s/%s", swapper.Config.ChainConfig.BSCExplorerUrl, signedTx.Hash().String())
		}
	} else {
		signedTx, err := buildSignedTransaction(swapper.ETHClient, tokenInstance.ETHPrivateKey, tokenInstance.ETHContractAddr, txInput)
		if err != nil {
			return err
		}
		txData, err := rlp.EncodeToBytes(signedTx)
		if err != nil {
			return err
		}
		swapTx := &model.SwapTx{
			Direction:      SwapBSC2Eth,
			DepositTxHash:  swap.DepositTxHash,
			WithdrawTxHash: signedTx.Hash().String(),
			TxData:         hexutil.Encode(txData),
			Status:         model.WithdrawTxSent,
		}
		err = swapper.insertSwapTxToDB(swapTx)
		if err != nil {
			return err
		}
		err = swapper.ETHClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			util.Logger.Errorf("broadcast tx to ETH error: %s", err.Error())
			err = swapper.DB.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", swapTx.DepositTxHash).Updates(
				map[string]interface{}{
					"status":     model.WithdrawTxCreated,
					"updated_at": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update table %s failed: %s", model.SwapTx{}.TableName(), err.Error())
			}
		} else {
			util.Logger.Infof("Send transaction to ETH, %s/%s", swapper.Config.ChainConfig.ETHExplorerUrl, signedTx.Hash().String())
		}
	}
	return nil
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
