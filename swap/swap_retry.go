package swap

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

func (engine *SwapEngine) getRetrySwapHMAC(retrySwap *model.RetrySwap) string {
	material := fmt.Sprintf("%d#%s#%s#%s#%s#%s#%s#%s#%s#%d#%v",
		retrySwap.SwapID, retrySwap.Direction, retrySwap.StartTxHash, retrySwap.FillTxHash, retrySwap.Sponsor,
		retrySwap.BEP20Addr, retrySwap.ERC20Addr, retrySwap.Symbol, retrySwap.Amount, retrySwap.Decimals, retrySwap.Done)
	mac := hmac.New(sha256.New, []byte(engine.hmacCKey))
	mac.Write([]byte(material))

	return hex.EncodeToString(mac.Sum(nil))
}

func (engine *SwapEngine) verifyRetrySwap(retrySwap *model.RetrySwap) bool {
	return retrySwap.RecordHash == engine.getRetrySwapHMAC(retrySwap)
}

func (engine *SwapEngine) insertRetrySwap(tx *gorm.DB, swap *model.RetrySwap) error {
	swap.RecordHash = engine.getRetrySwapHMAC(swap)
	return tx.Create(swap).Error
}

func (engine *SwapEngine) updateRetrySwap(tx *gorm.DB, retrySwap *model.RetrySwap) {
	retrySwap.RecordHash = engine.getRetrySwapHMAC(retrySwap)
	tx.Save(retrySwap)
}

func (engine *SwapEngine) insertRetrySwapTxsToDB(data *model.RetrySwapTx) error {
	tx := engine.db.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(data).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (engine *SwapEngine) doRetrySwap(retrySwap *model.RetrySwap, swapPairInstance *SwapPairIns) error {
	amount := big.NewInt(0)
	_, ok := amount.SetString(retrySwap.Amount, 10)
	if !ok {
		return fmt.Errorf("invalid swap amount: %s", retrySwap.Amount)
	}

	if retrySwap.Direction == SwapEth2BSC {
		bscClientMutex.Lock()
		defer bscClientMutex.Unlock()
		data, err := abiEncodeFillETH2BSCSwap(ethcom.HexToHash(retrySwap.StartTxHash), swapPairInstance.ERC20Addr, ethcom.HexToAddress(retrySwap.Sponsor), amount, engine.bscSwapAgentABI)
		if err != nil {
			return err
		}
		signedTx, err := buildSignedTransaction(common.ChainBSC, engine.bscTxSender, engine.bscSwapAgent, engine.bscClient, data, engine.tssClientSecureConfig, engine.config.KeyManagerConfig.Endpoint)
		if err != nil {
			return err
		}
		retrySwapTx := &model.RetrySwapTx{
			SwapID:              retrySwap.SwapID,
			StartTxHash:         retrySwap.StartTxHash,
			Direction:           retrySwap.Direction,
			RetryFillSwapTxHash: signedTx.Hash().String(),
			Status:              model.FillRetryTxCreated,
		}
		err = engine.insertRetrySwapTxsToDB(retrySwapTx)
		if err != nil {
			return err
		}
		err = engine.bscClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			util.Logger.Errorf("broadcast tx to BSC error: %s", err.Error())
			return err
		}
		util.Logger.Infof("Send transaction to BSC, %s/%s", engine.config.ChainConfig.BSCExplorerUrl, signedTx.Hash().String())
		return nil
	} else {
		ethClientMutex.Lock()
		defer ethClientMutex.Unlock()
		data, err := abiEncodeFillBSC2ETHSwap(ethcom.HexToHash(retrySwap.StartTxHash), swapPairInstance.ERC20Addr, ethcom.HexToAddress(retrySwap.Sponsor), amount, engine.ethSwapAgentABI)
		signedTx, err := buildSignedTransaction(common.ChainETH, engine.ethTxSender, engine.ethSwapAgent, engine.ethClient, data, engine.tssClientSecureConfig, engine.config.KeyManagerConfig.Endpoint)
		if err != nil {
			return err
		}
		retrySwapTx := &model.RetrySwapTx{
			SwapID:              retrySwap.SwapID,
			StartTxHash:         retrySwap.StartTxHash,
			Direction:           retrySwap.Direction,
			RetryFillSwapTxHash: signedTx.Hash().String(),
		}
		err = engine.insertRetrySwapTxsToDB(retrySwapTx)
		if err != nil {
			return err
		}
		err = engine.ethClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			util.Logger.Errorf("broadcast tx to ETH error: %s", err.Error())
			return err
		} else {
			util.Logger.Infof("Send transaction to ETH, %s/%s", engine.config.ChainConfig.ETHExplorerUrl, signedTx.Hash().String())
		}
		return nil
	}
}

func (engine *SwapEngine) retryFailedSwapsDaemon() {
	for {
		retrySwaps := make([]model.RetrySwap, 0)
		engine.db.Where("done = ?", false).Order("id asc").Limit(BatchSize).Find(&retrySwaps)

		for _, retrySwap := range retrySwaps {
			var swapPairInstance *SwapPairIns
			var err error
			retryCheckErr := func() error {
				valid := engine.verifyRetrySwap(&retrySwap)
				if !valid {
					return fmt.Errorf("verify hmac of retry swap failed: %s", retrySwap.StartTxHash)
				}
				swapPairInstance, err = engine.GetSwapPairInstance(ethcom.HexToAddress(retrySwap.ERC20Addr))
				if err != nil {
					return fmt.Errorf("failed to get swap instance for erc20 %s, err: %s, skip this swap", retrySwap.ERC20Addr, err.Error())
				}
				return nil
			} ()
			if retryCheckErr != nil {
				writeDBErr := func() error {
					tx := engine.db.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					retrySwap.Done = true
					retrySwap.ErrorMsg = retryCheckErr.Error()
					engine.updateRetrySwap(tx, &retrySwap)
					return tx.Commit().Error
				}()
				if writeDBErr != nil {
					util.Logger.Errorf("write db error: %s", writeDBErr.Error())
					util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
				}
				continue
			}

			util.Logger.Infof("Retry to handle swap, id: %d, direction %s, symbol %s, bep20 address %s, erc20 address %s, amount %s, sponsor %s",
				retrySwap.ID, retrySwap.Direction, retrySwap.Symbol, retrySwap.BEP20Addr, retrySwap.ERC20Addr, retrySwap.Amount, retrySwap.Sponsor)

			doRetrySwapErr := engine.doRetrySwap(&retrySwap, swapPairInstance)
			writeDBErr := func() error {
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if doRetrySwapErr == nil {
					tx.Model(model.RetrySwapTx{}).Where("swap_id = ?", retrySwap.SwapID).Updates(
						map[string]interface{}{
							"status":     model.FillRetryTxSent,
							"updated_at": time.Now().Unix(),
						})
					retrySwap.Done = true
					engine.updateRetrySwap(tx, &retrySwap)
				}
				return tx.Commit().Error
			}()
			if writeDBErr != nil {
				util.Logger.Errorf("write db error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
			}
		}
	}
}

func (engine *SwapEngine) trackRetrySwapTxDaemon() {
	go func() {
		for {
			time.Sleep(SleepTime * time.Second)

			retrySwapTxs := make([]model.RetrySwapTx, 0)
			engine.db.Where("status = ? and track_retry_counter >= ?", model.FillRetryTxSent, engine.config.ChainConfig.ETHMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&retrySwapTxs)

			if len(retrySwapTxs) > 0 {
				util.Logger.Infof("%d retry fill tx are missing, mark these retry swaps as failed", len(retrySwapTxs))
			}

			for _, retrySwapTx := range retrySwapTxs {
				chainName := "ETH"
				maxRetry := engine.config.ChainConfig.ETHMaxTrackRetry
				if retrySwapTx.Direction == SwapEth2BSC {
					chainName = "BSC"
					maxRetry = engine.config.ChainConfig.BSCMaxTrackRetry
				}
				util.Logger.Errorf("The retry fill tx is sent, however, after %d seconds its status is still uncertain. Mark tx as missing and mark swap as failed, chain %s, fill hash %s", SleepTime*maxRetry, chainName, retrySwapTx.RetryFillSwapTxHash)
				util.SendTelegramMessage(fmt.Sprintf("The retry fill tx is sent, however, after %d seconds its status is still uncertain. Mark tx as missing and mark swap as failed, chain %s, start hash %s", SleepTime*maxRetry, chainName, retrySwapTx.RetryFillSwapTxHash))

				writeDBErr := func() error {
					tx := engine.db.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					tx.Model(model.RetrySwapTx{}).Where("id = ?", retrySwapTx.ID).Updates(
						map[string]interface{}{
							"status":     model.FillRetryTxMissing,
							"updated_at": time.Now().Unix(),
						})

					return tx.Commit().Error
				}()
				if writeDBErr != nil {
					util.Logger.Errorf("write db error: %s", writeDBErr.Error())
					util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
				}
			}
		}
	}()

	go func() {
		for {
			time.Sleep(SleepTime * time.Second)

			retrySwapTxs := make([]model.RetrySwapTx, 0)
			engine.db.Where("status = ? and track_retry_counter < ?", model.FillRetryTxSent, engine.config.ChainConfig.ETHMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&retrySwapTxs)

			if len(retrySwapTxs) > 0 {
				util.Logger.Infof("Track %d non-finalized retry swap txs", len(retrySwapTxs))
			}

			for _, retrySwapTx := range retrySwapTxs {
				var client *ethclient.Client
				var chainName string
				if retrySwapTx.Direction == SwapBSC2Eth {
					client = engine.ethClient
					chainName = "ETH"
				} else {
					client = engine.bscClient
					chainName = "BSC"
				}
				var txRecipient *types.Receipt
				queryTxStatusErr := func() error {
					block, err := client.BlockByNumber(context.Background(), nil)
					if err != nil {
						util.Logger.Debugf("%s, query block failed: %s", chainName, err.Error())
						return err
					}
					txRecipient, err = client.TransactionReceipt(context.Background(), ethcom.HexToHash(retrySwapTx.RetryFillSwapTxHash))
					if err != nil {
						util.Logger.Debugf("%s, query tx failed: %s", chainName, err.Error())
						return err
					}
					if block.Number().Int64() < txRecipient.BlockNumber.Int64()+engine.config.ChainConfig.ETHConfirmNum {
						return fmt.Errorf("%s, swap tx is still not finalized", chainName)
					}
					return nil
				}()

				writeDBErr := func() error {
					tx := engine.db.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					if queryTxStatusErr != nil {
						tx.Model(model.RetrySwapTx{}).Where("id = ?", retrySwapTx.ID).Updates(
							map[string]interface{}{
								"track_retry_counter": gorm.Expr("track_retry_counter + 1"),
								"updated_at":          time.Now().Unix(),
							})
					} else {
						if txRecipient.Status == TxFailedStatus {
							util.Logger.Infof(fmt.Sprintf("fill tx is failed, chain %s, txHash: %s", chainName, txRecipient.TxHash))
							util.SendTelegramMessage(fmt.Sprintf("fill tx is failed, chain %s, txHash: %s", chainName, txRecipient.TxHash))
							tx.Model(model.RetrySwapTx{}).Where("id = ?", retrySwapTx.ID).Updates(
								map[string]interface{}{
									"status":              model.FillRetryTxFailed,
									"updated_at":          time.Now().Unix(),
								})
						} else {
							tx.Model(model.RetrySwapTx{}).Where("id = ?", retrySwapTx.ID).Updates(
								map[string]interface{}{
									"status":              model.FillRetryTxSuccess,
									"updated_at":          time.Now().Unix(),
								})

							swap, err := engine.getSwapByStartTxHash(tx, retrySwapTx.StartTxHash)
							if err != nil {
								tx.Rollback()
								return err
							}
							swap.Status = SwapSuccess
							engine.updateSwap(tx, swap)
						}
					}
					return tx.Commit().Error
				}()
				if writeDBErr != nil {
					util.Logger.Errorf("update db failure: %s", writeDBErr.Error())
					util.SendTelegramMessage(fmt.Sprintf("Upgent alert: update db failure: %s", writeDBErr.Error()))
				}
			}
		}
	}()
}

func (engine *SwapEngine) InsertRetryFailedSwaps(swapIDList []uint) ([]uint, error) {
	swaps := make([]model.Swap, 0)
	engine.db.Where("id  in (?)", swapIDList).Find(&swaps)

	if len(swaps) == 0 {
		return nil, fmt.Errorf("no matched swap")
	}

	retrySwapList := make([]uint, 0, len(swapIDList))
	writeDBErr := func() error {
		tx := engine.db.Begin()
		if err := tx.Error; err != nil {
			return err
		}
		for _, swap := range swaps {
			if swap.Status != SwapSendFailed {
				util.Logger.Infof("Skip non-failed swap, swapID %d", swap.ID)
				continue
			}
			retrySwapList = append(retrySwapList, swap.ID)
			retrySwap := &model.RetrySwap{
				SwapID:      swap.ID,
				Direction:   swap.Direction,
				StartTxHash: swap.StartTxHash,
				FillTxHash:  swap.FillTxHash,
				Sponsor:     swap.Sponsor,
				BEP20Addr:   swap.BEP20Addr,
				ERC20Addr:   swap.ERC20Addr,
				Symbol:      swap.Symbol,
				Amount:      swap.Amount,
				Decimals:    swap.Decimals,
			}
			if err := engine.insertRetrySwap(tx, retrySwap); err != nil {
				tx.Rollback()
				return err
			}
		}
		return tx.Commit().Error
	}()
	return retrySwapList, writeDBErr
}
