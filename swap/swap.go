package swap

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/swap/erc20"
	"github.com/binance-chain/bsc-eth-swap/util"
)

// NewSwapper returns the Swapper instance
func NewSwapper(db *gorm.DB, cfg *util.Config, bscClient, ethClient *ethclient.Client) (*Swapper, error) {
	tokens := make([]model.Token, 0)
	db.Find(&tokens)

	tokenInstances, err := buildTokenInstance(tokens, cfg)
	if err != nil {
		return nil, err
	}
	bscContractAddrToSymbol := make(map[string]string)
	ethContractAddrToSymbol := make(map[string]string)
	for _, token := range tokens {
		bscContractAddrToSymbol[token.BSCTokenContractAddr] = token.Symbol
		ethContractAddrToSymbol[token.ETHTokenContractAddr] = token.Symbol
	}

	hmacKey, err := GetHMACKey(cfg)
	if err != nil {
		return nil, err

	}
	swapper := &Swapper{
		DB:                      db,
		Config:                  cfg,
		HMACKey:                 hmacKey,
		BSCClient:               bscClient,
		ETHClient:               ethClient,
		TokenInstances:          tokenInstances,
		BSCContractAddrToSymbol: bscContractAddrToSymbol,
		ETHContractAddrToSymbol: ethContractAddrToSymbol,
		NewTokenSignal:          make(chan string),
	}

	return swapper, nil
}

func (swapper *Swapper) Start() {
	go swapper.monitorSwapRequestDaemon()
	go swapper.confirmSwapRequestDaemon()
	go swapper.createSwapDaemon()
	go swapper.trackSwapTxDaemon()
	go swapper.alertDaemon()
}

func (swapper *Swapper) monitorSwapRequestDaemon() {
	for {
		txEventLogs := make([]model.TxEventLog, 0)
		swapper.DB.Where("phase = ?", model.SeenSwapRequest).Order("height asc").Limit(BatchSize).Find(&txEventLogs)

		if len(txEventLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		for _, txEventLog := range txEventLogs {
			swap := swapper.createSwap(&txEventLog)
			writeDBErr := func() error {
				tx := swapper.DB.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if err := swapper.insertSwap(tx, swap); err != nil {
					tx.Rollback()
					return err
				}
				tx.Model(model.TxEventLog{}).Where("tx_hash = ?", swap.DepositTxHash).Updates(
					map[string]interface{}{
						"phase":       model.ConfirmSwapRequest,
						"update_time": time.Now().Unix(),
					})
				return tx.Commit().Error
			}()

			if writeDBErr != nil {
				util.Logger.Errorf("write DB error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write DB error: %s", writeDBErr.Error()))
			}
		}
	}
}

func (swapper *Swapper) getSwapHMAC(swap *model.Swap) string {
	material := fmt.Sprintf("%s#%s#%s#%s#%d#%s#%s#%s#%s",
		swap.Status, swap.Sponsor, swap.Symbol, swap.Amount, swap.Decimals, swap.Direction, swap.DepositTxHash, swap.WithdrawTxHash, swap.RefundTxHash)
	mac := hmac.New(sha256.New, []byte(swapper.HMACKey))
	mac.Write([]byte(material))

	return hex.EncodeToString(mac.Sum(nil))
}

func (swapper *Swapper) verifySwap(swap *model.Swap) bool {
	return swap.RecordHash == swapper.getSwapHMAC(swap)
}

func (swapper *Swapper) insertSwap(tx *gorm.DB, swap *model.Swap) error {
	swap.RecordHash = swapper.getSwapHMAC(swap)
	return tx.Create(swap).Error
}

func (swapper *Swapper) createSwap(txEventLog *model.TxEventLog) *model.Swap {
	swapper.Mutex.RLock()
	defer swapper.Mutex.RUnlock()

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
		txEventLogs := make([]model.TxEventLog, 0)
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
				util.Logger.Errorf("unexpected error, can't find swap by deposit hash: %s", txEventLog.TxHash)
				util.SendTelegramMessage(fmt.Sprintf("unexpected error, can't find swap by deposit hash: %s", txEventLog.TxHash))
				continue
			}

			if !swapper.verifySwap(&swap) {
				util.Logger.Errorf("verify hmac of swap failed: %s", swap.DepositTxHash)
				util.SendTelegramMessage(fmt.Sprintf("Urgent alert: verify hmac of swap failed: %s", swap.DepositTxHash))
				continue
			}

			writeDBErr := func() error {
				tx := swapper.DB.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if swap.Status == SwapTokenReceived {
					swap.Status = SwapQuoteConfirmed
					swapper.updateSwap(tx, &swap)
				}

				tx.Model(model.TxEventLog{}).Where("id = ?", txEventLog.Id).Updates(
					map[string]interface{}{
						"phase":       model.AckSwapRequest,
						"update_time": time.Now().Unix(),
					})
				return tx.Commit().Error
			}()

			if writeDBErr != nil {
				util.Logger.Errorf("write DB error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write DB error: %s", writeDBErr.Error()))
			}
		}
	}
}

func (swapper *Swapper) updateSwap(tx *gorm.DB, swap *model.Swap) {
	swap.RecordHash = swapper.getSwapHMAC(swap)

	tx.Save(swap)
}

func (swapper *Swapper) createSwapDaemon() {
	// start initial token swap daemon
	for symbol, tokenInstance := range swapper.TokenInstances {
		go swapper.swapInstanceDaemon(symbol, SwapEth2BSC, tokenInstance)
		go swapper.swapInstanceDaemon(symbol, SwapBSC2Eth, tokenInstance)
	}
	// start new swap daemon for admin
	for symbol := range swapper.NewTokenSignal {
		func() {
			swapper.Mutex.RLock()
			defer swapper.Mutex.RUnlock()

			util.Logger.Infof("start new swap daemon for %s", symbol)
			tokenInstance, ok := swapper.TokenInstances[symbol]
			if !ok {
				util.Logger.Errorf("unexpected error, can't find token install for symbol %s", symbol)
				util.SendTelegramMessage(fmt.Sprintf("unexpected error, can't find token install for symbol %s", symbol))
			} else {
				go swapper.swapInstanceDaemon(symbol, SwapEth2BSC, tokenInstance)
				go swapper.swapInstanceDaemon(symbol, SwapBSC2Eth, tokenInstance)
			}
		}()
	}
	select {}
}

func (swapper *Swapper) swapInstanceDaemon(symbol string, direction common.SwapDirection, tokenInstance *TokenInstance) {
	util.Logger.Infof("start swap daemon for %s, direction %s", symbol, direction)
	for {
		select {
		case <-tokenInstance.CloseSignal:
			util.Logger.Infof("close swap daemon for %s, direction %s", symbol, direction)
			util.SendTelegramMessage(fmt.Sprintf("close swap daemon for %s, direction %s", symbol, direction))
			return
		default:
		}

		swaps := make([]model.Swap, 0)
		swapper.DB.Where("status in (?) and symbol = ? and direction = ?", []common.SwapStatus{SwapQuoteConfirmed, SwapQuoteSending}, symbol, direction).Order("id asc").Limit(BatchSize).Find(&swaps)

		if len(swaps) == 0 {
			time.Sleep(SwapSleepSecond * time.Second)
			continue
		}

		util.Logger.Infof("found %d confirmed swap requests on token %s", len(swaps), symbol)

		for _, swap := range swaps {
			if !swapper.verifySwap(&swap) {
				util.Logger.Errorf("verify hmac of swap failed: %s", swap.DepositTxHash)
				util.SendTelegramMessage(fmt.Sprintf("Urgent alert: verify hmac of swap failed: %s", swap.DepositTxHash))
				continue
			}

			skip, writeDBErr := func() (bool, error) {
				isSkip := false
				tx := swapper.DB.Begin()
				if err := tx.Error; err != nil {
					return false, err
				}
				if swap.Status == SwapQuoteSending {
					var swapTx model.SwapTx
					swapper.DB.Where("deposit_tx_hash = ?", swap.DepositTxHash).First(&swapTx)
					if swapTx.DepositTxHash == "" {
						util.Logger.Infof("retry swap, deposit tx hash %s", swap.DepositTxHash)
						tx.Model(model.Swap{}).Where("id = ?", swap.ID).Updates(
							map[string]interface{}{
								"log":        "retry swap",
								"updated_at": time.Now().Unix(),
							})
					} else {
						tx.Model(model.SwapTx{}).Where("id = ?", swapTx.ID).Updates(
							map[string]interface{}{
								"status":     model.WithdrawTxSent,
								"updated_at": time.Now().Unix(),
							})

						// update swap
						swap.Status = SwapSent
						swap.WithdrawTxHash = swapTx.WithdrawTxHash
						swapper.updateSwap(tx, &swap)

						isSkip = true
					}
				} else {
					swap.Status = SwapQuoteSending
					swapper.updateSwap(tx, &swap)
				}
				return isSkip, tx.Commit().Error
			}()
			if writeDBErr != nil {
				util.Logger.Errorf("write DB error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write DB error: %s", writeDBErr.Error()))
				continue
			}
			if skip {
				util.Logger.Infof("skip this swap, deposit tx hash %s", swap.DepositTxHash)
				continue
			}

			util.Logger.Infof("do swap token %s , direction %s, sponsor: %s, amount %s, decimals %d,", symbol, direction, swap.Sponsor, swap.Amount, swap.Decimals)
			swapTx, swapErr := swapper.doSwap(&swap, tokenInstance)

			writeDBErr = func() error {
				tx := swapper.DB.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if swapErr != nil {
					util.Logger.Errorf("do swap failed: %s, deposit hash %s", swapErr.Error(), swap.DepositTxHash)
					if swapErr.Error() == core.ErrReplaceUnderpriced.Error() {
						// retry this swap
						swap.Status = SwapQuoteConfirmed
						swap.Log = fmt.Sprintf("do swap failure: %s", swapErr.Error())

						swapper.updateSwap(tx, &swap)
					} else {
						util.SendTelegramMessage(fmt.Sprintf("do swap failed: %s, deposit hash %s", swapErr.Error(), swap.DepositTxHash))
						withdrawTxHash := ""
						if swapTx != nil {
							withdrawTxHash = swapTx.WithdrawTxHash
						}

						swap.Status = SwapSendFailed
						swap.WithdrawTxHash = withdrawTxHash
						swap.Log = fmt.Sprintf("do swap failure: %s", swapErr.Error())
						swapper.updateSwap(tx, &swap)
					}
				} else {
					tx.Model(model.SwapTx{}).Where("id = ?", swapTx.ID).Updates(
						map[string]interface{}{
							"status":     model.WithdrawTxSent,
							"updated_at": time.Now().Unix(),
						})

					swap.Status = SwapSent
					swap.WithdrawTxHash = swapTx.WithdrawTxHash
					swapper.updateSwap(tx, &swap)
				}

				return tx.Commit().Error
			}()

			if writeDBErr != nil {
				util.Logger.Errorf("write DB error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write DB error: %s", writeDBErr.Error()))
			}

			if swap.Direction == SwapEth2BSC {
				time.Sleep(time.Duration(swapper.Config.ChainConfig.BSCWaitMilliSecBetweenSwaps) * time.Millisecond)
			} else {
				time.Sleep(time.Duration(swapper.Config.ChainConfig.ETHWaitMilliSecBetweenSwaps) * time.Millisecond)
			}
		}
	}
}

func (swapper *Swapper) doSwap(swap *model.Swap, tokenInstance *TokenInstance) (*model.SwapTx, error) {
	amount := big.NewInt(0)
	_, ok := amount.SetString(swap.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid swap amount: %s", swap.Amount)
	}
	txInput, err := abiEncodeTransfer(ethcom.HexToAddress(swap.Sponsor), amount)
	if err != nil {
		return nil, err
	}

	if swap.Direction == SwapEth2BSC {
		signedTx, err := buildSignedTransaction(swapper.BSCClient, tokenInstance.BSCPrivateKey, tokenInstance.BSCTokenContractAddr, txInput)
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
		signedTx, err := buildSignedTransaction(swapper.ETHClient, tokenInstance.ETHPrivateKey, tokenInstance.ETHTokenContractAddr, txInput)
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

func (swapper *Swapper) trackSwapTxDaemon() {
	go func() {
		for {
			time.Sleep(SleepTime * time.Second)

			ethSwapTxs := make([]model.SwapTx, 0)
			swapper.DB.Where("status = ? and direction = ? and track_retry_counter >= ?", model.WithdrawTxSent, SwapBSC2Eth, swapper.Config.ChainConfig.ETHMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&ethSwapTxs)

			bscSwapTxs := make([]model.SwapTx, 0)
			swapper.DB.Where("status = ? and direction = ? and track_retry_counter >= ?", model.WithdrawTxSent, SwapEth2BSC, swapper.Config.ChainConfig.BSCMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&bscSwapTxs)

			swapTxs := append(ethSwapTxs, bscSwapTxs...)

			if len(swapTxs) > 0 {
				util.Logger.Infof("%d withdraw tx are missing, mark these swaps as failed", len(swapTxs))
			}

			for _, swapTx := range swapTxs {
				chainName := "ETH"
				maxRetry := swapper.Config.ChainConfig.ETHMaxTrackRetry
				if swapTx.Direction == SwapEth2BSC {
					chainName = "BSC"
					maxRetry = swapper.Config.ChainConfig.BSCMaxTrackRetry
				}
				util.Logger.Errorf("The withdraw tx is sent, however, after %d seconds its status is still uncertain. Mark tx as missing and mark swap as failed, chain %s, deposit hash %s", SleepTime*maxRetry, chainName, swapTx.DepositTxHash)
				util.SendTelegramMessage(fmt.Sprintf("The withdraw tx is sent, however, after %d seconds its status is still uncertain. Mark tx as missing and mark swap as failed, chain %s, deposit hash %s", SleepTime*maxRetry, chainName, swapTx.DepositTxHash))

				writeDBErr := func() error {
					tx := swapper.DB.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					tx.Model(model.SwapTx{}).Where("id = ?", swapTx.ID).Updates(
						map[string]interface{}{
							"status":     model.WithdrawTxMissing,
							"updated_at": time.Now().Unix(),
						})

					swap, err := swapper.getSwapByDepositTxHash(tx, swapTx.DepositTxHash)
					if err != nil {
						tx.Rollback()
						return err
					}
					swap.Status = SwapSendFailed
					swap.Log = fmt.Sprintf("track withdraw tx for more than %d times, the withdraw tx status is still uncertain", maxRetry)
					swapper.updateSwap(tx, swap)

					return tx.Commit().Error
				}()
				if writeDBErr != nil {
					util.Logger.Errorf("write DB error: %s", writeDBErr.Error())
					util.SendTelegramMessage(fmt.Sprintf("write DB error: %s", writeDBErr.Error()))
				}
			}
		}
	}()

	go func() {
		for {
			time.Sleep(SleepTime * time.Second)

			ethSwapTxs := make([]model.SwapTx, 0)
			swapper.DB.Where("status = ? and direction = ? and track_retry_counter < ?", model.WithdrawTxSent, SwapBSC2Eth, swapper.Config.ChainConfig.ETHMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&ethSwapTxs)

			bscSwapTxs := make([]model.SwapTx, 0)
			swapper.DB.Where("status = ? and direction = ? and track_retry_counter < ?", model.WithdrawTxSent, SwapEth2BSC, swapper.Config.ChainConfig.BSCMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&bscSwapTxs)

			swapTxs := append(ethSwapTxs, bscSwapTxs...)

			if len(swapTxs) > 0 {
				util.Logger.Infof("Track %d non-finalized swap txs", len(swapTxs))
			}

			for _, swapTx := range swapTxs {
				gasPrice := big.NewInt(0)
				gasPrice.SetString(swapTx.GasPrice, 10)

				var client *ethclient.Client
				var chainName string
				if swapTx.Direction == SwapBSC2Eth {
					client = swapper.ETHClient
					chainName = "ETH"
				} else {
					client = swapper.BSCClient
					chainName = "BSC"
				}
				var txRecipient *types.Receipt
				queryTxStatusErr := func() error {
					block, err := client.BlockByNumber(context.Background(), nil)
					if err != nil {
						util.Logger.Debugf("%s, query block failed: %s", chainName, err.Error())
						return err
					}
					txRecipient, err = client.TransactionReceipt(context.Background(), ethcom.HexToHash(swapTx.WithdrawTxHash))
					if err != nil {
						util.Logger.Debugf("%s, query tx failed: %s", chainName, err.Error())
						return err
					}
					if block.Number().Int64() < txRecipient.BlockNumber.Int64()+swapper.Config.ChainConfig.ETHConfirmNum {
						return fmt.Errorf("%s, swap tx is still not finalized", chainName)
					}
					return nil
				}()

				writeDBErr := func() error {
					tx := swapper.DB.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					if queryTxStatusErr != nil {
						tx.Model(model.SwapTx{}).Where("id = ?", swapTx.ID).Updates(
							map[string]interface{}{
								"track_retry_counter": gorm.Expr("track_retry_counter + 1"),
								"updated_at":          time.Now().Unix(),
							})
					} else {
						txFee := big.NewInt(1).Mul(gasPrice, big.NewInt(int64(txRecipient.GasUsed))).String()
						if txRecipient.Status == TxFailedStatus {
							util.Logger.Infof(fmt.Sprintf("withdraw tx is failed, chain %s, txHash: %s", chainName, txRecipient.TxHash))
							util.SendTelegramMessage(fmt.Sprintf("withdraw tx is failed, chain %s, txHash: %s", chainName, txRecipient.TxHash))
							tx.Model(model.SwapTx{}).Where("id = ?", swapTx.ID).Updates(
								map[string]interface{}{
									"status":              model.WithdrawTxFailed,
									"height":              txRecipient.BlockNumber.Int64(),
									"consumed_fee_amount": txFee,
									"updated_at":          time.Now().Unix(),
								})

							swap, err := swapper.getSwapByDepositTxHash(tx, swapTx.DepositTxHash)
							if err != nil {
								tx.Rollback()
								return err
							}
							swap.Status = SwapSendFailed
							swap.Log = "withdraw tx is failed"
							swapper.updateSwap(tx, swap)
						} else {
							tx.Model(model.SwapTx{}).Where("id = ?", swapTx.ID).Updates(
								map[string]interface{}{
									"status":              model.WithdrawTxSuccess,
									"height":              txRecipient.BlockNumber.Int64(),
									"consumed_fee_amount": txFee,
									"updated_at":          time.Now().Unix(),
								})

							swap, err := swapper.getSwapByDepositTxHash(tx, swapTx.DepositTxHash)
							if err != nil {
								tx.Rollback()
								return err
							}
							swap.Status = SwapSuccess
							swapper.updateSwap(tx, swap)
						}
					}
					return tx.Commit().Error
				}()
				if writeDBErr != nil {
					util.Logger.Errorf("update DB failure: %s", writeDBErr.Error())
					util.SendTelegramMessage(fmt.Sprintf("Upgent alert: update DB failure: %s", writeDBErr.Error()))
				}

			}
		}
	}()
}

func (swapper *Swapper) getSwapByDepositTxHash(tx *gorm.DB, txHash string) (*model.Swap, error) {
	swap := model.Swap{}
	err := tx.Where("deposit_tx_hash = ?", txHash).First(&swap).Error
	return &swap, err
}

func (swapper *Swapper) alertDaemon() {
	bnbAlertThreshold := big.NewInt(1)
	bnbAlertThreshold.SetString(swapper.Config.ChainConfig.BNBAlertThreshold, 10)

	ethAlertThreshold := big.NewInt(1)
	ethAlertThreshold.SetString(swapper.Config.ChainConfig.ETHAlertThreshold, 10)
	for {
		time.Sleep(time.Second * time.Duration(swapper.Config.ChainConfig.BalanceMonitorInterval))

		for symbol, tokenInstance := range swapper.TokenInstances {
			bnbBalance, err := swapper.BSCClient.BalanceAt(context.Background(), tokenInstance.BSCTxSender, nil)
			if err != nil {
				util.Logger.Errorf(fmt.Sprintf("symbol %s, failed to query bsc balance: %s", symbol, err.Error()))
				util.SendTelegramMessage(fmt.Sprintf("symbol %s, failed to query bsc balance: %s", symbol, err.Error()))
			}
			if bnbBalance.Cmp(bnbAlertThreshold) <= 0 {
				util.Logger.Infof(fmt.Sprintf("symbol %s, bsc address %s, bnb balance %s is less than threshold %s", symbol, tokenInstance.BSCTxSender.String(), bnbBalance.String(), bnbAlertThreshold.String()))
				util.SendTelegramMessage(fmt.Sprintf("symbol %s, bsc address %s, bnb balance %s is less than threshold %s", symbol, tokenInstance.BSCTxSender.String(), bnbBalance.String(), bnbAlertThreshold.String()))
			}

			bscERC20, err := erc20.NewErc20(tokenInstance.BSCTokenContractAddr, swapper.BSCClient)
			if err != nil {
				util.Logger.Errorf(fmt.Sprintf("symbol %s, failed to create bsc erc20 instance: %s", symbol, err.Error()))
				util.SendTelegramMessage(fmt.Sprintf("symbol %s, failed to create bsc erc20 instance: %s", symbol, err.Error()))
			}

			bscErc20Balance, err := bscERC20.BalanceOf(getCallOpts(), tokenInstance.BSCTxSender)
			if err != nil {
				util.Logger.Errorf(fmt.Sprintf("symbol %s, failed to query bcs erc20 balance: %s", symbol, err.Error()))
				util.SendTelegramMessage(fmt.Sprintf("symbol %s, failed to query bcs erc20 balance: %s", symbol, err.Error()))
			} else {
				if bscErc20Balance.Cmp(tokenInstance.BSCERC20Threshold) <= 0 {
					util.Logger.Infof(fmt.Sprintf("symbol %s, bsc address %s, erc20 contract addr %s, bsc erc20 balance %s is less than threshold %s",
						symbol, tokenInstance.BSCTxSender.String(), tokenInstance.BSCTokenContractAddr.String(), bscErc20Balance.String(), tokenInstance.BSCERC20Threshold.String()))
					util.SendTelegramMessage(fmt.Sprintf("symbol %s, bsc address %s, erc20 contract addr %s, bsc erc20 balance %s is less than threshold %s",
						symbol, tokenInstance.BSCTxSender.String(), tokenInstance.BSCTokenContractAddr.String(), bscErc20Balance.String(), tokenInstance.BSCERC20Threshold.String()))
				}
			}

			ethBalance, err := swapper.ETHClient.BalanceAt(context.Background(), tokenInstance.ETHTxSender, nil)
			if err != nil {
				util.Logger.Errorf(fmt.Sprintf("symbol %s, failed to query eth balance: %s", symbol, err.Error()))
				util.SendTelegramMessage(fmt.Sprintf("symbol %s, failed to query eth balance: %s", symbol, err.Error()))
			}
			if ethBalance.Cmp(ethAlertThreshold) <= 0 {
				util.Logger.Infof(fmt.Sprintf("symbol %s, eth address %s, eth balance %s is less than threshold %s", symbol, tokenInstance.ETHTxSender.String(), ethBalance.String(), ethAlertThreshold.String()))
				util.SendTelegramMessage(fmt.Sprintf("symbol %s, eth address %s, eth balance %s is less than threshold %s", symbol, tokenInstance.ETHTxSender.String(), ethBalance.String(), ethAlertThreshold.String()))
			}

			ethERC20, err := erc20.NewErc20(tokenInstance.ETHTokenContractAddr, swapper.ETHClient)
			if err != nil {
				util.Logger.Errorf(fmt.Sprintf("symbol %s, failed to create eth erc20 instance: %s", symbol, err.Error()))
				util.SendTelegramMessage(fmt.Sprintf("symbol %s, failed to create eth erc20 instance: %s", symbol, err.Error()))
			}

			ethErc20Balance, err := ethERC20.BalanceOf(getCallOpts(), tokenInstance.ETHTxSender)
			if err != nil {
				util.Logger.Errorf(fmt.Sprintf("symbol %s, failed to query eth erc20 balance: %s", symbol, err.Error()))
				util.SendTelegramMessage(fmt.Sprintf("symbol %s, failed to query eth erc20 balance: %s", symbol, err.Error()))
			} else {
				if ethErc20Balance.Cmp(tokenInstance.ETHERC20Threshold) <= 0 {
					util.Logger.Infof(fmt.Sprintf("symbol %s, eth address %s, erc20 contract addr %s, eth erc20 balance %s is less than threshold %s",
						symbol, tokenInstance.ETHTxSender.String(), tokenInstance.ETHTokenContractAddr.String(), ethErc20Balance.String(), tokenInstance.ETHERC20Threshold.String()))
					util.SendTelegramMessage(fmt.Sprintf("symbol %s, eth address %s, erc20 contract addr %s, eth erc20 balance %s is less than threshold %s",
						symbol, tokenInstance.ETHTxSender.String(), tokenInstance.ETHTokenContractAddr.String(), ethErc20Balance.String(), tokenInstance.ETHERC20Threshold.String()))
				}
			}
		}
	}
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

func (swapper *Swapper) AddToken(token *model.Token, tokenKey *TokenKey) error {
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

	bscERC20Threshold := big.NewInt(0)
	bscERC20Threshold.SetString(token.BSCERC20Threshold, 10)

	ethERC20Threshold := big.NewInt(0)
	ethERC20Threshold.SetString(token.ETHERC20Threshold, 10)

	swapper.Mutex.Lock()
	defer swapper.Mutex.Unlock()
	swapper.TokenInstances[token.Symbol] = &TokenInstance{
		Symbol:               token.Symbol,
		Name:                 token.Name,
		Decimals:             token.Decimals,
		CloseSignal:          make(chan bool),
		LowBound:             lowBound,
		UpperBound:           upperBound,
		BSCPrivateKey:        tokenKey.BSCPrivateKey,
		BSCTokenContractAddr: ethcom.HexToAddress(token.BSCTokenContractAddr),
		BSCTxSender:          GetAddress(tokenKey.BSCPublicKey),
		BSCERC20Threshold:    bscERC20Threshold,
		ETHPrivateKey:        tokenKey.ETHPrivateKey,
		ETHTokenContractAddr: ethcom.HexToAddress(token.ETHTokenContractAddr),
		ETHTxSender:          GetAddress(tokenKey.ETHPublicKey),
		ETHERC20Threshold:    ethERC20Threshold,
	}
	swapper.BSCContractAddrToSymbol[strings.ToLower(token.BSCTokenContractAddr)] = token.Symbol
	swapper.ETHContractAddrToSymbol[strings.ToLower(token.ETHTokenContractAddr)] = token.Symbol

	swapper.NewTokenSignal <- token.Symbol
	return nil
}

func (swapper *Swapper) RemoveToken(token *model.Token) {
	swapper.Mutex.Lock()
	defer swapper.Mutex.Unlock()

	tokenInstance, ok := swapper.TokenInstances[token.Symbol]
	if !ok {
		util.Logger.Errorf("unsupported token %s", token.Symbol)
		return
	}
	close(tokenInstance.CloseSignal)

	delete(swapper.TokenInstances, token.Symbol)
	delete(swapper.BSCContractAddrToSymbol, strings.ToLower(token.BSCTokenContractAddr))
	delete(swapper.ETHContractAddrToSymbol, strings.ToLower(token.ETHTokenContractAddr))
}
