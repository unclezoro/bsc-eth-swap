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

	"github.com/jinzhu/gorm"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	sabi "github.com/binance-chain/bsc-eth-swap/abi"
	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

func NewSwapPairEngine(db *gorm.DB, cfg *util.Config, bscClient *ethclient.Client, swapEngine *SwapEngine) (*SwapPairEngine, error) {
	keyConfig, err := GetKeyConfig(cfg)
	if err != nil {
		return nil, err
	}
	bscChainID, err := bscClient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	bscSwapAgentAbi, err := abi.JSON(strings.NewReader(sabi.BSCSwapAgentABI))
	if err != nil {
		return nil, err
	}
	swapPairEngine := &SwapPairEngine{
		db:                    db,
		config:                cfg,
		hmacKey:               keyConfig.HMACKey,
		tssClientSecureConfig: NewClientSecureConfig(keyConfig),
		bscClient:             bscClient,
		bscChainID:            bscChainID.Int64(),
		bscTxSender:           ethcom.HexToAddress(cfg.KeyManagerConfig.BSCAccountAddr),
		bscSwapAgentABi:       &bscSwapAgentAbi,
		bscSwapAgent:          ethcom.HexToAddress(cfg.ChainConfig.BSCSwapAgentAddr),
		swapEngine:            swapEngine,
	}
	return swapPairEngine, nil
}

func (engine *SwapPairEngine) Start() {
	go engine.monitorSwapRequestDaemon()
	go engine.confirmSwapRequestDaemon()
	go engine.swapPairInstanceDaemon()
	go engine.trackSwapPairTxDaemon()
}

func (engine *SwapPairEngine) monitorSwapRequestDaemon() {
	for {
		swapPairRegisterTxLogs := make([]model.SwapPairRegisterTxLog, 0)
		engine.db.Where("phase = ?", model.SeenRequest).Order("height asc").Limit(BatchSize).Find(&swapPairRegisterTxLogs)

		if len(swapPairRegisterTxLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		for _, swapPairEventLog := range swapPairRegisterTxLogs {
			swapSM := engine.createSwapPairSM(&swapPairEventLog)
			writeDBErr := func() error {
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if err := engine.insertSwapPairSM(tx, swapSM); err != nil {
					tx.Rollback()
					return err
				}
				tx.Model(model.SwapPairRegisterTxLog{}).Where("tx_hash = ?", swapSM.PairRegisterTxHash).Updates(
					map[string]interface{}{
						"phase":       model.ConfirmRequest,
						"update_time": time.Now().Unix(),
					})
				return tx.Commit().Error
			}()

			if writeDBErr != nil {
				util.Logger.Errorf("write db error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
			}
		}
	}
}

func (engine *SwapPairEngine) createSwapPairSM(txEventLog *model.SwapPairRegisterTxLog) *model.SwapPairStateMachine {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	swapPairRegisterTxHash := txEventLog.TxHash

	ethContractAddr := ethcom.HexToAddress(txEventLog.ERC20Addr)
	swapPairStatus := SwapPairReceived
	// TODO, duplicate check
	swapSM := &model.SwapPairStateMachine{
		Status:    swapPairStatus,
		ERC20Addr: ethContractAddr.String(),
		BEP20Addr: "",
		Sponsor:   txEventLog.Sponsor,
		Symbol:    txEventLog.Symbol,
		Name:      txEventLog.Name,
		Decimals:  txEventLog.Decimals,

		PairRegisterTxHash: swapPairRegisterTxHash,
		PairCreatTxHash:    "",
		Log:                "",
	}

	return swapSM
}

func (engine *SwapPairEngine) insertSwapPairSM(tx *gorm.DB, swapSM *model.SwapPairStateMachine) error {
	swapSM.RecordHash = engine.getSwapPairSMHMAC(swapSM)
	return tx.Create(swapSM).Error
}

func (engine *SwapPairEngine) insertSwapPair(tx *gorm.DB, swapPair *model.SwapPair) error {
	swapPair.RecordHash = engine.getSwapPairHMAC(swapPair)
	return tx.Create(swapPair).Error
}

func (engine *SwapPairEngine) updateSwapPairSM(tx *gorm.DB, swapPairSM *model.SwapPairStateMachine) {
	swapPairSM.RecordHash = engine.getSwapPairSMHMAC(swapPairSM)

	tx.Save(swapPairSM)
}

func (engine *SwapPairEngine) verifySwapPairSM(swapPairSM *model.SwapPairStateMachine) bool {
	return swapPairSM.RecordHash == engine.getSwapPairSMHMAC(swapPairSM)
}

func (engine *SwapPairEngine) getSwapPairSMHMAC(swapSM *model.SwapPairStateMachine) string {
	material := fmt.Sprintf("%s#%s#%s#%s#%d#%s#%s#%s",
		swapSM.Status, swapSM.ERC20Addr, swapSM.BEP20Addr, swapSM.Symbol, swapSM.Decimals, swapSM.Name, swapSM.PairRegisterTxHash, swapSM.PairCreatTxHash)
	mac := hmac.New(sha256.New, []byte(engine.hmacKey))
	mac.Write([]byte(material))
	return hex.EncodeToString(mac.Sum(nil))
}

func (engine *SwapPairEngine) getSwapPairHMAC(swapSM *model.SwapPair) string {
	material := fmt.Sprintf("#%s#%s#%s#%d#%s",
		swapSM.ERC20Addr, swapSM.BEP20Addr, swapSM.Symbol, swapSM.Decimals, swapSM.Name)
	mac := hmac.New(sha256.New, []byte(engine.hmacKey))
	mac.Write([]byte(material))
	return hex.EncodeToString(mac.Sum(nil))
}

func (engine *SwapPairEngine) confirmSwapRequestDaemon() {
	for {
		swapPairRegisterEventLogs := make([]model.SwapPairRegisterTxLog, 0)
		engine.db.Where("status = ? and phase = ?", model.TxStatusConfirmed, model.ConfirmRequest).
			Order("height asc").Limit(BatchSize).Find(&swapPairRegisterEventLogs)

		if len(swapPairRegisterEventLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		util.Logger.Infof("found %d confirmed event logs", len(swapPairRegisterEventLogs))

		for _, pairRegisterEventLog := range swapPairRegisterEventLogs {
			swapPairSM := model.SwapPairStateMachine{}
			engine.db.Where("pair_register_tx_hash = ?", pairRegisterEventLog.TxHash).First(&swapPairSM)
			if swapPairSM.PairRegisterTxHash == "" {
				util.Logger.Errorf("unexpected error, can't find swapPairSM by register hash: %s", pairRegisterEventLog.TxHash)
				util.SendTelegramMessage(fmt.Sprintf("unexpected error, can't find swapPairSM by register hash: %s", pairRegisterEventLog.TxHash))
				continue
			}

			if !engine.verifySwapPairSM(&swapPairSM) {
				util.Logger.Errorf("verify hmac of swapPairSM failed: %s", swapPairSM.PairRegisterTxHash)
				util.SendTelegramMessage(fmt.Sprintf("Urgent alert: verify hmac of swapPairSM failed: %s", swapPairSM.PairRegisterTxHash))
				continue
			}

			writeDBErr := func() error {
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if swapPairSM.Status == SwapPairReceived {
					swapPairSM.Status = SwapPairConfirmed
					engine.updateSwapPairSM(tx, &swapPairSM)
				}

				tx.Model(model.SwapPairRegisterTxLog{}).Where("id = ?", pairRegisterEventLog.Id).Updates(
					map[string]interface{}{
						"phase":       model.AckRequest,
						"update_time": time.Now().Unix(),
					})
				return tx.Commit().Error
			}()

			if writeDBErr != nil {
				util.Logger.Errorf("write db error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
			}
		}
	}
}

func (engine *SwapPairEngine) swapPairInstanceDaemon() {
	for {

		swapPairSMs := make([]model.SwapPairStateMachine, 0)
		engine.db.Where("status in (?)", []common.SwapPairStatus{SwapPairConfirmed, SwapPairSending}).Order("id asc").Limit(BatchSize).Find(&swapPairSMs)

		if len(swapPairSMs) == 0 {
			time.Sleep(SwapSleepSecond * time.Second)
			continue
		}

		util.Logger.Infof("found %d confirmed swapPairSM pair register requests", len(swapPairSMs))

		for _, swapPairSM := range swapPairSMs {
			if !engine.verifySwapPairSM(&swapPairSM) {
				util.Logger.Errorf("verify hmac of swapPairSM failed: %s", swapPairSM.PairRegisterTxHash)
				util.SendTelegramMessage(fmt.Sprintf("Urgent alert: verify hmac of swapPairSM failed: %s", swapPairSM.PairRegisterTxHash))
				continue
			}

			skip, writeDBErr := func() (bool, error) {
				isSkip := false
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return false, err
				}
				if swapPairSM.Status == SwapPairSending {
					var swapPairTx model.SwapPairCreatTx
					engine.db.Where("swap_pair_register_tx_hash = ?", swapPairSM.PairRegisterTxHash).First(&swapPairTx)
					if swapPairTx.SwapPairRegisterTxHash == "" {
						util.Logger.Infof("retry swapPairSM, start tx hash %s", swapPairSM.PairRegisterTxHash)
						tx.Model(model.Swap{}).Where("id = ?", swapPairSM.ID).Updates(
							map[string]interface{}{
								"log":        "retry swapPairSM",
								"updated_at": time.Now().Unix(),
							})
					} else {
						tx.Model(model.SwapFillTx{}).Where("id = ?", swapPairTx.ID).Updates(
							map[string]interface{}{
								"status":     model.FillTxSent,
								"updated_at": time.Now().Unix(),
							})

						// update swapPairSM
						swapPairSM.Status = SwapPairSent
						swapPairSM.PairCreatTxHash = swapPairTx.SwapPairCreatTxHash
						engine.updateSwapPairSM(tx, &swapPairSM)
						isSkip = true
					}
				} else {
					swapPairSM.Status = SwapPairSending
					engine.updateSwapPairSM(tx, &swapPairSM)
				}
				return isSkip, tx.Commit().Error
			}()
			if writeDBErr != nil {
				util.Logger.Errorf("write db error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
				continue
			}
			if skip {
				util.Logger.Infof("skip this swapPairSM, start tx hash %s", swapPairSM.PairRegisterTxHash)
				continue
			}

			util.Logger.Infof("do swapPairSM, erc20 address %s, symbol %s", swapPairSM.ERC20Addr, swapPairSM.Symbol)
			swapPairCreateTx, swapErr := engine.doCreateSwapPair(&swapPairSM)

			writeDBErr = func() error {
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if swapErr != nil {
					util.Logger.Errorf("do swapPairSM failed: %s, start hash %s", swapErr.Error(), swapPairSM.PairRegisterTxHash)
					if swapErr.Error() == core.ErrReplaceUnderpriced.Error() {
						// retry this swapPairSM
						swapPairSM.Status = SwapPairConfirmed
						swapPairSM.Log = fmt.Sprintf("do swapPairSM failure: %s", swapErr.Error())

						engine.updateSwapPairSM(tx, &swapPairSM)
					} else {
						util.SendTelegramMessage(fmt.Sprintf("do swapPairSM failed: %s, start hash %s", swapErr.Error(), swapPairSM.PairRegisterTxHash))
						createPairTxHash := ""
						if swapPairCreateTx != nil {
							createPairTxHash = swapPairCreateTx.SwapPairCreatTxHash
						}

						swapPairSM.Status = SwapPairSendFailed
						swapPairSM.PairCreatTxHash = createPairTxHash
						swapPairSM.Log = fmt.Sprintf("do swapPairSM failure: %s", swapErr.Error())
						engine.updateSwapPairSM(tx, &swapPairSM)
					}
				} else {
					tx.Model(model.SwapPairCreatTx{}).Where("id = ?", swapPairCreateTx.ID).Updates(
						map[string]interface{}{
							"status":     model.FillTxSent,
							"updated_at": time.Now().Unix(),
						})

					swapPairSM.Status = SwapPairSent
					swapPairSM.PairCreatTxHash = swapPairCreateTx.SwapPairCreatTxHash
					engine.updateSwapPairSM(tx, &swapPairSM)
				}

				return tx.Commit().Error
			}()

			if writeDBErr != nil {
				util.Logger.Errorf("write db error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
			}
			time.Sleep(time.Duration(engine.config.ChainConfig.BSCWaitMilliSecBetweenSwaps) * time.Millisecond)
		}
	}
}

func (engine *SwapPairEngine) doCreateSwapPair(swapPairSM *model.SwapPairStateMachine) (*model.SwapPairCreatTx, error) {

	bscClientMutex.Lock()
	defer bscClientMutex.Unlock()
	data, err := abiEncodeCreateSwapPair(ethcom.HexToHash(swapPairSM.PairRegisterTxHash), ethcom.HexToAddress(swapPairSM.ERC20Addr), swapPairSM.Name, swapPairSM.Symbol, uint8(swapPairSM.Decimals), engine.bscSwapAgentABi)
	if err != nil {
		return nil, err
	}
	signedTx, err := buildSignedTransaction(common.ChainBSC, engine.bscTxSender, engine.bscSwapAgent, engine.bscClient, data, engine.tssClientSecureConfig, engine.config.KeyManagerConfig.Endpoint)
	if err != nil {
		return nil, err
	}
	swapTx := &model.SwapPairCreatTx{
		SwapPairRegisterTxHash: swapPairSM.PairRegisterTxHash,
		SwapPairCreatTxHash:    signedTx.Hash().String(),
		ERC20Addr:              swapPairSM.ERC20Addr,
		Symbol:                 swapPairSM.Symbol,
		Name:                   swapPairSM.Name,
		Decimals:               swapPairSM.Decimals,
		GasPrice:               signedTx.GasPrice().String(),
		Status:                 model.FillTxCreated,
	}
	err = engine.insertSwapPairTxToDB(swapTx)
	if err != nil {
		return nil, err
	}
	err = engine.bscClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		util.Logger.Errorf("broadcast tx to BSC error: %s", err.Error())
		return nil, err
	}
	util.Logger.Infof("Send transaction to BSC, %s/%s", engine.config.ChainConfig.BSCExplorerUrl, signedTx.Hash().String())
	return swapTx, nil
}

func (engine *SwapPairEngine) insertSwapPairTxToDB(data *model.SwapPairCreatTx) error {
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

func (engine *SwapPairEngine) trackSwapPairTxDaemon() {
	go func() {
		for {
			time.Sleep(SleepTime * time.Second)

			swapPairCreateTxs := make([]model.SwapPairCreatTx, 0)
			engine.db.Where("status = ? and track_retry_counter >= ?", model.FillTxSent, engine.config.ChainConfig.BSCMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&swapPairCreateTxs)

			if len(swapPairCreateTxs) > 0 {
				util.Logger.Infof("%d fill tx are missing, mark these swaps as failed", len(swapPairCreateTxs))
			}

			for _, swapPairTx := range swapPairCreateTxs {

				maxRetry := engine.config.ChainConfig.BSCMaxTrackRetry

				util.Logger.Errorf("The create swap pair tx is sent, however, after %d seconds its status is still uncertain. Mark tx as missing and mark swap as failed, crete swap hash %s", SleepTime*maxRetry, swapPairTx.SwapPairCreatTxHash)
				util.SendTelegramMessage(fmt.Sprintf("The create swap tx is sent, however, after %d seconds its status is still uncertain. Mark tx as missing and mark swap as failed, create swap %s", SleepTime*maxRetry, swapPairTx.SwapPairCreatTxHash))

				writeDBErr := func() error {
					tx := engine.db.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					tx.Model(model.SwapPairCreatTx{}).Where("id = ?", swapPairTx.ID).Updates(
						map[string]interface{}{
							"status":     model.FillTxMissing,
							"updated_at": time.Now().Unix(),
						})

					swapPairSM, err := engine.getSwapPairSMByRegisterTxHash(tx, swapPairTx.SwapPairRegisterTxHash)
					if err != nil {
						tx.Rollback()
						return err
					}
					swapPairSM.Status = SwapPairSendFailed
					swapPairSM.Log = fmt.Sprintf("track create swap tx for more than %d times, the fill tx status is still uncertain", maxRetry)
					engine.updateSwapPairSM(tx, swapPairSM)

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

			swapPairTxs := make([]model.SwapPairCreatTx, 0)
			engine.db.Where("status = ? and track_retry_counter < ?", model.FillTxSent, engine.config.ChainConfig.ETHMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&swapPairTxs)

			if len(swapPairTxs) > 0 {
				util.Logger.Infof("Track %d non-finalized swap txs", len(swapPairTxs))
			}

			for _, swapPairTx := range swapPairTxs {
				gasPrice := big.NewInt(0)
				gasPrice.SetString(swapPairTx.GasPrice, 10)

				client := engine.bscClient

				var txRecipient *types.Receipt
				queryTxStatusErr := func() error {
					block, err := client.BlockByNumber(context.Background(), nil)
					if err != nil {
						util.Logger.Debugf("query block failed: %s", err.Error())
						return err
					}
					txRecipient, err = client.TransactionReceipt(context.Background(), ethcom.HexToHash(swapPairTx.SwapPairCreatTxHash))
					if err != nil {
						util.Logger.Debugf("query tx failed: %s", err.Error())
						return err
					}
					if block.Number().Int64() < txRecipient.BlockNumber.Int64()+engine.config.ChainConfig.ETHConfirmNum {
						return fmt.Errorf("swap tx is still not finalized")
					}
					return nil
				}()

				writeDBErr := func() error {
					tx := engine.db.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					if queryTxStatusErr != nil {
						tx.Model(model.SwapPairCreatTx{}).Where("id = ?", swapPairTx.ID).Updates(
							map[string]interface{}{
								"track_retry_counter": gorm.Expr("track_retry_counter + 1"),
								"updated_at":          time.Now().Unix(),
							})
					} else {
						txFee := big.NewInt(1).Mul(gasPrice, big.NewInt(int64(txRecipient.GasUsed))).String()
						if txRecipient.Status == TxFailedStatus {
							util.Logger.Infof(fmt.Sprintf("create swapPairSM pair tx is failed, txHash: %s", txRecipient.TxHash))
							util.SendTelegramMessage(fmt.Sprintf("create swapPairSM pair tx is failed, chain %s, txHash: %s", txRecipient.TxHash))
							tx.Model(model.SwapPairCreatTx{}).Where("id = ?", swapPairTx.ID).Updates(
								map[string]interface{}{
									"status":              model.FillTxFailed,
									"height":              txRecipient.BlockNumber.Int64(),
									"consumed_fee_amount": txFee,
									"updated_at":          time.Now().Unix(),
								})

							swapPairSM, err := engine.getSwapPairSMByRegisterTxHash(tx, swapPairTx.SwapPairRegisterTxHash)
							if err != nil {
								tx.Rollback()
								return err
							}
							swapPairSM.Status = SwapPairSendFailed
							swapPairSM.Log = "create swapPairSM pair tx is failed"
							engine.updateSwapPairSM(tx, swapPairSM)
						} else {
							bep20ContractAddr, err := queryDeployedBEP20ContractAddr(
								ethcom.HexToAddress(swapPairTx.ERC20Addr),
								ethcom.HexToAddress(engine.config.ChainConfig.BSCSwapAgentAddr),
								txRecipient,
								engine.bscClient)
							if err != nil {
								tx.Rollback()
								return err
							}
							tx.Model(model.SwapPairCreatTx{}).Where("id = ?", swapPairTx.ID).Updates(
								map[string]interface{}{
									"status":              model.FillTxSuccess,
									"height":              txRecipient.BlockNumber.Int64(),
									"consumed_fee_amount": txFee,
									"updated_at":          time.Now().Unix(),
								})

							swapPairSM, err := engine.getSwapPairSMByRegisterTxHash(tx, swapPairTx.SwapPairRegisterTxHash)
							if err != nil {
								tx.Rollback()
								return err
							}
							swapPairSM.Status = SwapPairSuccess
							swapPairSM.BEP20Addr = bep20ContractAddr.String()
							engine.updateSwapPairSM(tx, swapPairSM)
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

	go func() {
		for {
			time.Sleep(SleepTime * time.Second)

			swapPairSMs := make([]model.SwapPairStateMachine, 0)
			engine.db.Where("status = ? ", SwapPairSuccess).
				Order("id asc").Limit(TrackSwapPairSMBatchSize).Find(&swapPairSMs)

			if len(swapPairSMs) > 0 {
				util.Logger.Infof("Track %d success swap pair created ", len(swapPairSMs))
			}
			for _, swapPairSM := range swapPairSMs {
				if !engine.verifySwapPairSM(&swapPairSM) {
					util.Logger.Errorf("verify hmac of swapPairSM failed: %s", swapPairSM.PairRegisterTxHash)
					util.SendTelegramMessage(fmt.Sprintf("Urgent alert: verify hmac of swapPairSM failed: %s", swapPairSM.PairRegisterTxHash))
					continue
				}
				swapPair := model.SwapPair{
					Sponsor:    swapPairSM.Sponsor,
					Symbol:     swapPairSM.Symbol,
					Name:       swapPairSM.Name,
					Decimals:   swapPairSM.Decimals,
					BEP20Addr:  swapPairSM.BEP20Addr,
					ERC20Addr:  swapPairSM.ERC20Addr,
					Available:  true,
					LowBound:   "0",
					UpperBound: MaxUpperBound,
					IconUrl:    "",
				}
				writeDBErr := func() error {
					tx := engine.db.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					if err := engine.insertSwapPair(tx, &swapPair); err != nil {
						tx.Rollback()
						return err
					}
					swapPairSM.Status = SwapPairFinalized
					engine.updateSwapPairSM(tx, &swapPairSM)
					return tx.Commit().Error
				}()

				if writeDBErr != nil {
					util.Logger.Errorf("write db error: %s", writeDBErr.Error())
					util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
				}
				engine.swapEngine.AddSwapPairInstance(&swapPair)
			}

		}
	}()

}

func (engine *SwapPairEngine) getSwapPairSMByRegisterTxHash(tx *gorm.DB, txHash string) (*model.SwapPairStateMachine, error) {
	swapPairSM := model.SwapPairStateMachine{}
	err := tx.Where("pair_register_tx_hash = ?", txHash).First(&swapPairSM).Error
	return &swapPairSM, err
}
