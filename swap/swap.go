package swap

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"

	sabi "github.com/binance-chain/bsc-eth-swap/abi"
	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

// NewSwapEngine returns the swapEngine instance
func NewSwapEngine(db *gorm.DB, cfg *util.Config, bscClient, ethClient *ethclient.Client) (*SwapEngine, error) {
	pairs := make([]model.SwapPair, 0)
	db.Find(&pairs)

	swapPairInstances, err := buildSwapPairInstance(pairs)
	if err != nil {
		return nil, err
	}
	bscContractAddrToEthContractAddr := make(map[ethcom.Address]ethcom.Address)
	ethContractAddrToBscContractAddr := make(map[ethcom.Address]ethcom.Address)
	for _, token := range pairs {
		bscContractAddrToEthContractAddr[ethcom.HexToAddress(token.BEP20Addr)] = ethcom.HexToAddress(token.ERC20Addr)
		ethContractAddrToBscContractAddr[ethcom.HexToAddress(token.ERC20Addr)] = ethcom.HexToAddress(token.BEP20Addr)
	}

	keyConfig, err := GetKeyConfig(cfg)
	if err != nil {
		return nil, err

	}

	bscChainID, err := bscClient.ChainID(context.Background())
	if err != nil {
		return nil, err

	}
	ethChainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	ethSwapAgentAbi, err := abi.JSON(strings.NewReader(sabi.ETHSwapAgentABI))
	if err != nil {
		return nil, err
	}

	bscSwapAgentAbi, err := abi.JSON(strings.NewReader(sabi.BSCSwapAgentABI))
	if err != nil {
		return nil, err
	}

	swapEngine := &SwapEngine{
		db:                     db,
		config:                 cfg,
		hmacCKey:               keyConfig.HMACKey,
		tssClientSecureConfig:  NewClientSecureConfig(keyConfig),
		bscClient:              bscClient,
		ethClient:              ethClient,
		bscChainID:             bscChainID.Int64(),
		ethChainID:             ethChainID.Int64(),
		bscTxSender:            ethcom.HexToAddress(cfg.KeyManagerConfig.BSCAccountAddr),
		ethTxSender:            ethcom.HexToAddress(cfg.KeyManagerConfig.ETHAccountAddr),
		swapPairsFromERC20Addr: swapPairInstances,
		bep20ToERC20:           bscContractAddrToEthContractAddr,
		erc20ToBEP20:           ethContractAddrToBscContractAddr,
		ethSwapAgentABI:        &ethSwapAgentAbi,
		bscSwapAgentABI:        &bscSwapAgentAbi,
		ethSwapAgent:           ethcom.HexToAddress(cfg.ChainConfig.ETHSwapAgentAddr),
		bscSwapAgent:           ethcom.HexToAddress(cfg.ChainConfig.BSCSwapAgentAddr),
	}

	return swapEngine, nil
}

func (engine *SwapEngine) Start() {
	go engine.monitorSwapRequestDaemon()
	go engine.confirmSwapRequestDaemon()
	go engine.swapInstanceDaemon(SwapEth2BSC)
	go engine.swapInstanceDaemon(SwapBSC2Eth)
	go engine.trackSwapTxDaemon()
	go engine.retryFailedSwapsDaemon()
}

func (engine *SwapEngine) monitorSwapRequestDaemon() {
	for {
		swapStartTxLogs := make([]model.SwapStartTxLog, 0)
		engine.db.Where("phase = ?", model.SeenRequest).Order("height asc").Limit(BatchSize).Find(&swapStartTxLogs)

		if len(swapStartTxLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		for _, swapEventLog := range swapStartTxLogs {
			swap := engine.createSwap(&swapEventLog)
			writeDBErr := func() error {
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if err := engine.insertSwap(tx, swap); err != nil {
					tx.Rollback()
					return err
				}
				tx.Model(model.SwapStartTxLog{}).Where("tx_hash = ?", swap.StartTxHash).Updates(
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

func (engine *SwapEngine) getSwapHMAC(swap *model.Swap) string {
	material := fmt.Sprintf("%s#%s#%s#%s#%s#%s#%d#%s#%s#%s",
		swap.Status, swap.Sponsor, swap.BEP20Addr, swap.ERC20Addr, swap.Symbol, swap.Amount, swap.Decimals, swap.Direction, swap.StartTxHash, swap.FillTxHash)
	mac := hmac.New(sha256.New, []byte(engine.hmacCKey))
	mac.Write([]byte(material))

	return hex.EncodeToString(mac.Sum(nil))
}

func (engine *SwapEngine) verifySwap(swap *model.Swap) bool {
	return swap.RecordHash == engine.getSwapHMAC(swap)
}

func (engine *SwapEngine) insertSwap(tx *gorm.DB, swap *model.Swap) error {
	swap.RecordHash = engine.getSwapHMAC(swap)
	return tx.Create(swap).Error
}

func (engine *SwapEngine) updateSwap(tx *gorm.DB, swap *model.Swap) {
	swap.RecordHash = engine.getSwapHMAC(swap)
	tx.Save(swap)
}

func (engine *SwapEngine) createSwap(txEventLog *model.SwapStartTxLog) *model.Swap {
	sponsor := txEventLog.FromAddress
	amount := txEventLog.Amount
	swapStartTxHash := txEventLog.TxHash
	swapDirection := SwapEth2BSC
	if txEventLog.Chain == common.ChainBSC {
		swapDirection = SwapBSC2Eth
	}

	var bep20Addr ethcom.Address
	var erc20Addr ethcom.Address
	var ok bool
	decimals := 0
	var symbol string
	swapStatus := SwapQuoteRejected
	err := func() error {
		if txEventLog.Chain == common.ChainETH {
			erc20Addr = ethcom.HexToAddress(txEventLog.TokenAddr)
			if bep20Addr, ok = engine.erc20ToBEP20[ethcom.HexToAddress(txEventLog.TokenAddr)]; !ok {
				return fmt.Errorf("unsupported eth token contract address: %s", txEventLog.TokenAddr)
			}
		} else {
			bep20Addr = ethcom.HexToAddress(txEventLog.TokenAddr)
			if erc20Addr, ok = engine.bep20ToERC20[ethcom.HexToAddress(txEventLog.TokenAddr)]; !ok {
				return fmt.Errorf("unsupported bsc token contract address: %s", txEventLog.TokenAddr)
			}
		}
		pairInstance, err := engine.GetSwapPairInstance(erc20Addr)
		if err != nil {
			return fmt.Errorf("failed to get swap pair for bep20 %s, error %s", bep20Addr.String(), err.Error())
		}
		decimals = pairInstance.Decimals
		symbol = pairInstance.Symbol
		swapAmount := big.NewInt(0)
		_, ok = swapAmount.SetString(txEventLog.Amount, 10)
		if !ok {
			return fmt.Errorf("unrecongnized swap amount: %s", txEventLog.Amount)
		}
		if swapAmount.Cmp(pairInstance.LowBound) < 0 || swapAmount.Cmp(pairInstance.UpperBound) > 0 {
			return fmt.Errorf("swap amount is out of bound, expected bound [%s, %s]", pairInstance.LowBound.String(), pairInstance.UpperBound.String())
		}

		swapStatus = SwapTokenReceived
		return nil
	}()

	log := ""
	if err != nil {
		log = err.Error()
	}

	swap := &model.Swap{
		Status:      swapStatus,
		Sponsor:     sponsor,
		BEP20Addr:   bep20Addr.String(),
		ERC20Addr:   erc20Addr.String(),
		Symbol:      symbol,
		Amount:      amount,
		Decimals:    decimals,
		Direction:   swapDirection,
		StartTxHash: swapStartTxHash,
		FillTxHash:  "",
		Log:         log,
	}

	return swap
}

func (engine *SwapEngine) confirmSwapRequestDaemon() {
	for {
		txEventLogs := make([]model.SwapStartTxLog, 0)
		engine.db.Where("status = ? and phase = ?", model.TxStatusConfirmed, model.ConfirmRequest).
			Order("height asc").Limit(BatchSize).Find(&txEventLogs)

		if len(txEventLogs) == 0 {
			time.Sleep(SleepTime * time.Second)
			continue
		}

		util.Logger.Infof("found %d confirmed event logs", len(txEventLogs))

		for _, txEventLog := range txEventLogs {
			swap := model.Swap{}
			engine.db.Where("start_tx_hash = ?", txEventLog.TxHash).First(&swap)
			if swap.StartTxHash == "" {
				util.Logger.Errorf("unexpected error, can't find swap by start hash: %s", txEventLog.TxHash)
				util.SendTelegramMessage(fmt.Sprintf("unexpected error, can't find swap by start hash: %s", txEventLog.TxHash))
				continue
			}

			if !engine.verifySwap(&swap) {
				util.Logger.Errorf("verify hmac of swap failed: %s", swap.StartTxHash)
				util.SendTelegramMessage(fmt.Sprintf("Urgent alert: verify hmac of swap failed: %s", swap.StartTxHash))
				continue
			}

			writeDBErr := func() error {
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if swap.Status == SwapTokenReceived {
					swap.Status = SwapConfirmed
					engine.updateSwap(tx, &swap)
				}

				tx.Model(model.SwapStartTxLog{}).Where("id = ?", txEventLog.Id).Updates(
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

func (engine *SwapEngine) swapInstanceDaemon(direction common.SwapDirection) {
	util.Logger.Infof("start swap daemon, direction %s", direction)
	for {

		swaps := make([]model.Swap, 0)
		engine.db.Where("status in (?) and direction = ?", []common.SwapStatus{SwapConfirmed, SwapSending}, direction).Order("id asc").Limit(BatchSize).Find(&swaps)

		if len(swaps) == 0 {
			time.Sleep(SwapSleepSecond * time.Second)
			continue
		}

		util.Logger.Debugf("found %d confirmed swap requests", len(swaps))

		for _, swap := range swaps {
			if !engine.verifySwap(&swap) {
				util.Logger.Errorf("verify hmac of swap failed: %s", swap.StartTxHash)
				util.SendTelegramMessage(fmt.Sprintf("Urgent alert: verify hmac of swap failed: %s", swap.StartTxHash))
				continue
			}
			swapPairInstance, err := engine.GetSwapPairInstance(ethcom.HexToAddress(swap.ERC20Addr))
			if err != nil {
				util.Logger.Debugf("swap instance for bep20 %s doesn't exist, skip this swap", swap.BEP20Addr)
				continue
			}

			skip, writeDBErr := func() (bool, error) {
				isSkip := false
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return false, err
				}
				if swap.Status == SwapSending {
					var swapTx model.SwapFillTx
					engine.db.Where("start_swap_tx_hash = ?", swap.StartTxHash).First(&swapTx)
					if swapTx.FillSwapTxHash == "" {
						util.Logger.Infof("retry swap, start tx hash %s, symbol %s, amount %s, direction",
							swap.StartTxHash, swap.Symbol, swap.Amount, swap.Direction)
						swap.Status = SwapConfirmed
						engine.updateSwap(tx, &swap)
					} else {
						tx.Model(model.SwapFillTx{}).Where("id = ?", swapTx.ID).Updates(
							map[string]interface{}{
								"status":     model.FillTxSent,
								"updated_at": time.Now().Unix(),
							})

						// update swap
						swap.Status = SwapSent
						swap.FillTxHash = swapTx.FillSwapTxHash
						engine.updateSwap(tx, &swap)

						isSkip = true
					}
				} else {
					swap.Status = SwapSending
					engine.updateSwap(tx, &swap)
				}
				return isSkip, tx.Commit().Error
			}()
			if writeDBErr != nil {
				util.Logger.Errorf("write db error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
				continue
			}
			if skip {
				util.Logger.Debugf("skip this swap, start tx hash %s", swap.StartTxHash)
				continue
			}

			util.Logger.Infof("do swap token %s , direction %s, sponsor: %s, amount %s, decimals %d,", swap.BEP20Addr, direction, swap.Sponsor, swap.Amount, swap.Decimals)
			swapTx, swapErr := engine.doSwap(&swap, swapPairInstance)

			writeDBErr = func() error {
				tx := engine.db.Begin()
				if err := tx.Error; err != nil {
					return err
				}
				if swapErr != nil {
					util.Logger.Errorf("do swap failed: %s, start hash %s", swapErr.Error(), swap.StartTxHash)
					if swapErr.Error() == core.ErrReplaceUnderpriced.Error() {
						// retry this swap
						swap.Status = SwapConfirmed
						swap.Log = fmt.Sprintf("do swap failure: %s", swapErr.Error())

						engine.updateSwap(tx, &swap)
					} else {
						util.SendTelegramMessage(fmt.Sprintf("do swap failed: %s, start hash %s", swapErr.Error(), swap.StartTxHash))
						fillTxHash := ""
						if swapTx != nil {
							fillTxHash = swapTx.FillSwapTxHash
						}

						swap.Status = SwapSendFailed
						swap.FillTxHash = fillTxHash
						swap.Log = fmt.Sprintf("do swap failure: %s", swapErr.Error())
						engine.updateSwap(tx, &swap)
					}
				} else {
					tx.Model(model.SwapFillTx{}).Where("id = ?", swapTx.ID).Updates(
						map[string]interface{}{
							"status":     model.FillTxSent,
							"updated_at": time.Now().Unix(),
						})

					swap.Status = SwapSent
					swap.FillTxHash = swapTx.FillSwapTxHash
					engine.updateSwap(tx, &swap)
				}

				return tx.Commit().Error
			}()

			if writeDBErr != nil {
				util.Logger.Errorf("write db error: %s", writeDBErr.Error())
				util.SendTelegramMessage(fmt.Sprintf("write db error: %s", writeDBErr.Error()))
			}

			if swap.Direction == SwapEth2BSC {
				time.Sleep(time.Duration(engine.config.ChainConfig.BSCWaitMilliSecBetweenSwaps) * time.Millisecond)
			} else {
				time.Sleep(time.Duration(engine.config.ChainConfig.ETHWaitMilliSecBetweenSwaps) * time.Millisecond)
			}
		}
	}
}

func (engine *SwapEngine) doSwap(swap *model.Swap, swapPairInstance *SwapPairIns) (*model.SwapFillTx, error) {
	amount := big.NewInt(0)
	_, ok := amount.SetString(swap.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid swap amount: %s", swap.Amount)
	}

	if swap.Direction == SwapEth2BSC {
		bscClientMutex.Lock()
		defer bscClientMutex.Unlock()
		data, err := abiEncodeFillETH2BSCSwap(ethcom.HexToHash(swap.StartTxHash), swapPairInstance.ERC20Addr, ethcom.HexToAddress(swap.Sponsor), amount, engine.bscSwapAgentABI)
		if err != nil {
			return nil, err
		}
		signedTx, err := buildSignedTransaction(common.ChainBSC, engine.bscTxSender, engine.bscSwapAgent, engine.bscClient, data, engine.tssClientSecureConfig, engine.config.KeyManagerConfig.Endpoint)
		if err != nil {
			return nil, err
		}
		swapTx := &model.SwapFillTx{
			Direction:       SwapEth2BSC,
			StartSwapTxHash: swap.StartTxHash,
			FillSwapTxHash:  signedTx.Hash().String(),
			GasPrice:        signedTx.GasPrice().String(),
			Status:          model.FillTxCreated,
		}
		err = engine.insertSwapTxToDB(swapTx)
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
	} else {
		ethClientMutex.Lock()
		defer ethClientMutex.Unlock()
		data, err := abiEncodeFillBSC2ETHSwap(ethcom.HexToHash(swap.StartTxHash), swapPairInstance.ERC20Addr, ethcom.HexToAddress(swap.Sponsor), amount, engine.ethSwapAgentABI)
		signedTx, err := buildSignedTransaction(common.ChainETH, engine.ethTxSender, engine.ethSwapAgent, engine.ethClient, data, engine.tssClientSecureConfig, engine.config.KeyManagerConfig.Endpoint)
		if err != nil {
			return nil, err
		}
		swapTx := &model.SwapFillTx{
			Direction:       SwapBSC2Eth,
			StartSwapTxHash: swap.StartTxHash,
			GasPrice:        signedTx.GasPrice().String(),
			FillSwapTxHash:  signedTx.Hash().String(),
			Status:          model.FillTxCreated,
		}
		err = engine.insertSwapTxToDB(swapTx)
		if err != nil {
			return nil, err
		}
		err = engine.ethClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			util.Logger.Errorf("broadcast tx to ETH error: %s", err.Error())
			return nil, err
		} else {
			util.Logger.Infof("Send transaction to ETH, %s/%s", engine.config.ChainConfig.ETHExplorerUrl, signedTx.Hash().String())
		}
		return swapTx, nil
	}
}

func (engine *SwapEngine) trackSwapTxDaemon() {
	go func() {
		for {
			time.Sleep(SleepTime * time.Second)

			swapTxs := make([]model.SwapFillTx, 0)
			engine.db.Where("status = ? and track_retry_counter >= ?", model.FillTxSent, engine.config.ChainConfig.ETHMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&swapTxs)

			if len(swapTxs) > 0 {
				util.Logger.Infof("%d fill tx are missing, mark these swaps as failed", len(swapTxs))
			}

			for _, swapTx := range swapTxs {
				chainName := "ETH"
				maxRetry := engine.config.ChainConfig.ETHMaxTrackRetry
				if swapTx.Direction == SwapEth2BSC {
					chainName = "BSC"
					maxRetry = engine.config.ChainConfig.BSCMaxTrackRetry
				}
				util.Logger.Errorf("The fill tx is sent, however, after %d seconds its status is still uncertain. Mark tx as missing and mark swap as failed, chain %s, fill hash %s", SleepTime*maxRetry, chainName, swapTx.StartSwapTxHash)
				util.SendTelegramMessage(fmt.Sprintf("The fill tx is sent, however, after %d seconds its status is still uncertain. Mark tx as missing and mark swap as failed, chain %s, start hash %s", SleepTime*maxRetry, chainName, swapTx.StartSwapTxHash))

				writeDBErr := func() error {
					tx := engine.db.Begin()
					if err := tx.Error; err != nil {
						return err
					}
					tx.Model(model.SwapFillTx{}).Where("id = ?", swapTx.ID).Updates(
						map[string]interface{}{
							"status":     model.FillTxMissing,
							"updated_at": time.Now().Unix(),
						})

					swap, err := engine.getSwapByStartTxHash(tx, swapTx.StartSwapTxHash)
					if err != nil {
						tx.Rollback()
						return err
					}
					swap.Status = SwapSendFailed
					swap.Log = fmt.Sprintf("track fill tx for more than %d times, the fill tx status is still uncertain", maxRetry)
					engine.updateSwap(tx, swap)

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

			ethSwapTxs := make([]model.SwapFillTx, 0)
			engine.db.Where("status = ? and direction = ? and track_retry_counter < ?", model.FillTxSent, SwapBSC2Eth, engine.config.ChainConfig.ETHMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&ethSwapTxs)

			bscSwapTxs := make([]model.SwapFillTx, 0)
			engine.db.Where("status = ? and direction = ? and track_retry_counter < ?", model.FillTxSent, SwapEth2BSC, engine.config.ChainConfig.BSCMaxTrackRetry).
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
					txRecipient, err = client.TransactionReceipt(context.Background(), ethcom.HexToHash(swapTx.FillSwapTxHash))
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
						tx.Model(model.SwapFillTx{}).Where("id = ?", swapTx.ID).Updates(
							map[string]interface{}{
								"track_retry_counter": gorm.Expr("track_retry_counter + 1"),
								"updated_at":          time.Now().Unix(),
							})
					} else {
						txFee := big.NewInt(1).Mul(gasPrice, big.NewInt(int64(txRecipient.GasUsed))).String()
						if txRecipient.Status == TxFailedStatus {
							util.Logger.Infof(fmt.Sprintf("fill tx is failed, chain %s, txHash: %s", chainName, txRecipient.TxHash))
							util.SendTelegramMessage(fmt.Sprintf("fill tx is failed, chain %s, txHash: %s", chainName, txRecipient.TxHash))
							tx.Model(model.SwapFillTx{}).Where("id = ?", swapTx.ID).Updates(
								map[string]interface{}{
									"status":              model.FillTxFailed,
									"height":              txRecipient.BlockNumber.Int64(),
									"consumed_fee_amount": txFee,
									"updated_at":          time.Now().Unix(),
								})

							swap, err := engine.getSwapByStartTxHash(tx, swapTx.StartSwapTxHash)
							if err != nil {
								tx.Rollback()
								return err
							}
							swap.Status = SwapSendFailed
							swap.Log = "fill tx is failed"
							engine.updateSwap(tx, swap)
						} else {
							tx.Model(model.SwapFillTx{}).Where("id = ?", swapTx.ID).Updates(
								map[string]interface{}{
									"status":              model.FillTxSuccess,
									"height":              txRecipient.BlockNumber.Int64(),
									"consumed_fee_amount": txFee,
									"updated_at":          time.Now().Unix(),
								})

							swap, err := engine.getSwapByStartTxHash(tx, swapTx.StartSwapTxHash)
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

func (engine *SwapEngine) getSwapByStartTxHash(tx *gorm.DB, txHash string) (*model.Swap, error) {
	swap := model.Swap{}
	err := tx.Where("start_tx_hash = ?", txHash).First(&swap).Error
	return &swap, err
}

func (engine *SwapEngine) insertSwapTxToDB(data *model.SwapFillTx) error {
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

func (engine *SwapEngine) AddSwapPairInstance(swapPair *model.SwapPair) error {
	lowBound := big.NewInt(0)
	_, ok := lowBound.SetString(swapPair.LowBound, 10)
	if !ok {
		return fmt.Errorf("invalid lowBound amount: %s", swapPair.LowBound)
	}
	upperBound := big.NewInt(0)
	_, ok = upperBound.SetString(swapPair.UpperBound, 10)
	if !ok {
		return fmt.Errorf("invalid upperBound amount: %s", swapPair.LowBound)
	}

	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	engine.swapPairsFromERC20Addr[ethcom.HexToAddress(swapPair.ERC20Addr)] = &SwapPairIns{
		Symbol:     swapPair.Symbol,
		Name:       swapPair.Name,
		Decimals:   swapPair.Decimals,
		LowBound:   lowBound,
		UpperBound: upperBound,
		BEP20Addr:  ethcom.HexToAddress(swapPair.BEP20Addr),
		ERC20Addr:  ethcom.HexToAddress(swapPair.ERC20Addr),
	}
	engine.bep20ToERC20[ethcom.HexToAddress(swapPair.BEP20Addr)] = ethcom.HexToAddress(swapPair.ERC20Addr)
	engine.erc20ToBEP20[ethcom.HexToAddress(swapPair.ERC20Addr)] = ethcom.HexToAddress(swapPair.BEP20Addr)

	return nil
}

func (engine *SwapEngine) GetSwapPairInstance(erc20Addr ethcom.Address) (*SwapPairIns, error) {
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()

	tokenInstance, ok := engine.swapPairsFromERC20Addr[erc20Addr]
	if !ok {
		return nil, fmt.Errorf("swap instance doesn't exist")
	}
	return tokenInstance, nil
}

func (engine *SwapEngine) UpdateSwapInstance(swapPair *model.SwapPair) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	bscTokenAddr := ethcom.HexToAddress(swapPair.BEP20Addr)
	tokenInstance, ok := engine.swapPairsFromERC20Addr[bscTokenAddr]
	if !ok {
		return
	}

	if !swapPair.Available {
		delete(engine.swapPairsFromERC20Addr, bscTokenAddr)
		return
	}

	upperBound := big.NewInt(0)
	_, ok = upperBound.SetString(swapPair.UpperBound, 10)
	tokenInstance.UpperBound = upperBound

	lowBound := big.NewInt(0)
	_, ok = upperBound.SetString(swapPair.LowBound, 10)
	tokenInstance.LowBound = lowBound

	engine.swapPairsFromERC20Addr[bscTokenAddr] = tokenInstance
}

func (engine *SwapEngine) WithdrawToken(chain string, tokenAddr, recipient ethcom.Address, amount *big.Int) (string, error) {
	tokenABI, err := abi.JSON(strings.NewReader(sabi.ERC20ABI))
	if err != nil {
		return "", err
	}
	emptyAddr := ethcom.Address{}
	txSender := engine.bscTxSender
	client := engine.bscClient
	explorerUrl := engine.config.ChainConfig.BSCExplorerUrl
	if chain == common.ChainETH {
		txSender = engine.ethTxSender
		client = engine.ethClient
		explorerUrl = engine.config.ChainConfig.ETHExplorerUrl
		ethClientMutex.Lock()
		defer ethClientMutex.Unlock()
	} else {
		bscClientMutex.Lock()
		defer bscClientMutex.Unlock()
	}
	// withdraw native token
	if bytes.Equal(tokenAddr[:], emptyAddr[:]) {
		signedTx, err := buildNativeCoinTransferTx(chain, txSender, recipient, amount, client, engine.tssClientSecureConfig, engine.config.KeyManagerConfig.Endpoint)
		if err != nil {
			util.Logger.Errorf("build native coin transfer error: %s", err.Error())
			return "", err
		}
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			util.Logger.Errorf("broadcast tx to %s error: %s", chain, err.Error())
			return "", err
		}
		util.Logger.Infof("Send transaction to %s, %s/%s", chain, explorerUrl, signedTx.Hash().String())
		return signedTx.Hash().String(), nil
	}
	// withdraw BEP20 or ERC20 token
	data, err := abiEncodeERC20Transfer(recipient, amount, &tokenABI)
	if err != nil {
		return "", err
	}
	signedTx, err := buildSignedTransaction(chain, txSender, tokenAddr, client, data, engine.tssClientSecureConfig, engine.config.KeyManagerConfig.Endpoint)
	if err != nil {
		return "", err
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		util.Logger.Errorf("broadcast tx to %s error: %s", chain, err.Error())
		return "", err
	}
	util.Logger.Infof("Send transaction to %s, %s/%s", chain, explorerUrl, signedTx.Hash().String())
	return signedTx.Hash().String(), nil
}
