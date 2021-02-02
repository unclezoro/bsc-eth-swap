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
		bscContractAddrToEthContractAddr[ethcom.HexToAddress(token.BSCTokenContractAddr)] = ethcom.HexToAddress(token.ETHTokenContractAddr)
		ethContractAddrToBscContractAddr[ethcom.HexToAddress(token.ETHTokenContractAddr)] = ethcom.HexToAddress(token.BSCTokenContractAddr)
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
		db:                    db,
		config:                cfg,
		hmacCKey:              keyConfig.HMACKey,
		tssClientSecureConfig: NewClientSecureConfig(keyConfig),
		bscClient:             bscClient,
		ethClient:             ethClient,
		bscChainID:            bscChainID.Int64(),
		ethChainID:            ethChainID.Int64(),
		bscTxSender:           ethcom.HexToAddress(cfg.KeyManagerConfig.BSCAccountAddr),
		ethTxSender:           ethcom.HexToAddress(cfg.KeyManagerConfig.ETHAccountAddr),
		swapPairs:             swapPairInstances,
		bscToEthContractAddr:  bscContractAddrToEthContractAddr,
		ethToBscContractAddr:  ethContractAddrToBscContractAddr,
		newSwapPairSignal:     make(chan ethcom.Address),
		ethSwapAgentAbi:       &ethSwapAgentAbi,
		bscSwapAgentABi:       &bscSwapAgentAbi,
		ethSwapAgent:          ethcom.HexToAddress(cfg.ChainConfig.ETHSwapAgentAddr),
		bscSwapAgent:          ethcom.HexToAddress(cfg.ChainConfig.BSCSwapAgentAddr),
	}

	return swapEngine, nil
}

func (engine *SwapEngine) Start() {
	go engine.monitorSwapRequestDaemon()
	go engine.confirmSwapRequestDaemon()
	go engine.createSwapDaemon()
	go engine.trackSwapTxDaemon()
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
	material := fmt.Sprintf("%s#%s#%s#%s#%s#%s#%d#%s#%s#%s#%s",
		swap.Status, swap.Sponsor, swap.BscContractAddr, swap.EThContractAddr, swap.Symbol, swap.Amount, swap.Decimals, swap.Direction, swap.StartTxHash, swap.FillTxHash, swap.RefundTxHash)
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

func (engine *SwapEngine) createSwap(txEventLog *model.SwapStartTxLog) *model.Swap {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	sponsor := txEventLog.FromAddress
	amount := txEventLog.Amount
	swapStartTxHash := txEventLog.TxHash
	swapDirection := SwapEth2BSC
	if txEventLog.Chain == common.ChainBSC {
		swapDirection = SwapBSC2Eth
	}

	var bscContractAddr ethcom.Address
	var ethContractAddr ethcom.Address
	var ok bool
	decimals := 0
	var symbol string
	swapStatus := SwapQuoteRejected
	err := func() error {
		if txEventLog.Chain == common.ChainETH {
			ethContractAddr = ethcom.HexToAddress(txEventLog.ContractAddress)
			if bscContractAddr, ok = engine.ethToBscContractAddr[ethcom.HexToAddress(txEventLog.ContractAddress)]; !ok {
				return fmt.Errorf("unsupported eth token contract address: %s", txEventLog.ContractAddress)
			}
		} else {
			bscContractAddr = ethcom.HexToAddress(txEventLog.ContractAddress)
			if ethContractAddr, ok = engine.ethToBscContractAddr[ethcom.HexToAddress(txEventLog.ContractAddress)]; !ok {
				return fmt.Errorf("unsupported bsc token contract address: %s", txEventLog.ContractAddress)
			}
		}
		pairInstance, ok := engine.swapPairs[bscContractAddr]
		if !ok {
			return fmt.Errorf("unsupported swap pair %s", bscContractAddr.String())
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
		Status:          swapStatus,
		Sponsor:         sponsor,
		BscContractAddr: bscContractAddr.String(),
		EThContractAddr: ethContractAddr.String(),
		Symbol:          symbol,
		Amount:          amount,
		Decimals:        decimals,
		Direction:       swapDirection,
		StartTxHash:     swapStartTxHash,
		FillTxHash:      "",
		Log:             log,
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

func (engine *SwapEngine) updateSwap(tx *gorm.DB, swap *model.Swap) {
	swap.RecordHash = engine.getSwapHMAC(swap)

	tx.Save(swap)
}

func (engine *SwapEngine) createSwapDaemon() {
	// start initial swap pair daemon
	for _, swapPairInstance := range engine.swapPairs {
		go engine.swapInstanceDaemon(SwapEth2BSC, swapPairInstance)
		go engine.swapInstanceDaemon(SwapBSC2Eth, swapPairInstance)
	}
	// start new swap daemon for admin
	for bscContractAddr := range engine.newSwapPairSignal {
		func() {
			engine.mutex.RLock()
			defer engine.mutex.RUnlock()

			util.Logger.Infof("start new swap daemon for %s", bscContractAddr)
			tokenInstance, ok := engine.swapPairs[bscContractAddr]
			if !ok {
				util.Logger.Errorf("unexpected error, can't find token install for bsc contract %s", bscContractAddr)
				util.SendTelegramMessage(fmt.Sprintf("unexpected error, can't find token install for bsc contract %s", bscContractAddr))
			} else {
				go engine.swapInstanceDaemon(SwapEth2BSC, tokenInstance)
				go engine.swapInstanceDaemon(SwapBSC2Eth, tokenInstance)
			}
		}()
	}
	select {}
}

func (engine *SwapEngine) swapInstanceDaemon(direction common.SwapDirection, swapPairInstance *SwapPairIns) {
	bscContract := swapPairInstance.BSCTokenContractAddr
	util.Logger.Infof("start swap daemon for bsc token %s, direction %s", bscContract, direction)
	for {

		swaps := make([]model.Swap, 0)
		engine.db.Where("status in (?) and bsc_contract_addr = ? and direction = ?", []common.SwapStatus{SwapConfirmed, SwapSending}, bscContract.String(), direction).Order("id asc").Limit(BatchSize).Find(&swaps)

		if len(swaps) == 0 {
			time.Sleep(SwapSleepSecond * time.Second)
			continue
		}

		util.Logger.Infof("found %d confirmed swap requests on token %s", len(swaps), bscContract.String())

		for _, swap := range swaps {
			if !engine.verifySwap(&swap) {
				util.Logger.Errorf("verify hmac of swap failed: %s", swap.StartTxHash)
				util.SendTelegramMessage(fmt.Sprintf("Urgent alert: verify hmac of swap failed: %s", swap.StartTxHash))
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
					if swapTx.StartSwapTxHash == "" {
						util.Logger.Infof("retry swap, start tx hash %s", swap.StartTxHash)
						tx.Model(model.Swap{}).Where("id = ?", swap.ID).Updates(
							map[string]interface{}{
								"log":        "retry swap",
								"updated_at": time.Now().Unix(),
							})
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
				util.Logger.Infof("skip this swap, start tx hash %s", swap.StartTxHash)
				continue
			}

			util.Logger.Infof("do swap token %s , direction %s, sponsor: %s, amount %s, decimals %d,", bscContract.String(), direction, swap.Sponsor, swap.Amount, swap.Decimals)
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
		data, err := abiEncodeFillETH2BSCSwap(ethcom.HexToHash(swap.StartTxHash), swapPairInstance.BSCTokenContractAddr, ethcom.HexToAddress(swap.Sponsor), amount, engine.bscSwapAgentABi)
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
		data, err := abiEncodeFillBSC2ETHSwap(ethcom.HexToHash(swap.StartTxHash), swapPairInstance.ETHTokenContractAddr, ethcom.HexToAddress(swap.Sponsor), amount, engine.ethSwapAgentAbi)
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

			ethSwapTxs := make([]model.SwapFillTx, 0)
			engine.db.Where("status = ? and direction = ? and track_retry_counter >= ?", model.FillTxSent, SwapBSC2Eth, engine.config.ChainConfig.ETHMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&ethSwapTxs)

			bscSwapTxs := make([]model.SwapFillTx, 0)
			engine.db.Where("status = ? and direction = ? and track_retry_counter >= ?", model.FillTxSent, SwapEth2BSC, engine.config.ChainConfig.BSCMaxTrackRetry).
				Order("id asc").Limit(TrackSentTxBatchSize).Find(&bscSwapTxs)

			swapTxs := append(ethSwapTxs, bscSwapTxs...)

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

func (engine *SwapEngine) insertSwapToDB(data *model.Swap) error {
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
	engine.swapPairs[ethcom.HexToAddress(swapPair.BSCTokenContractAddr)] = &SwapPairIns{
		Symbol:               swapPair.Symbol,
		Name:                 swapPair.Name,
		Decimals:             swapPair.Decimals,
		LowBound:             lowBound,
		UpperBound:           upperBound,
		BSCTokenContractAddr: ethcom.HexToAddress(swapPair.BSCTokenContractAddr),
		ETHTokenContractAddr: ethcom.HexToAddress(swapPair.ETHTokenContractAddr),
	}
	engine.bscToEthContractAddr[ethcom.HexToAddress(swapPair.BSCTokenContractAddr)] = ethcom.HexToAddress(swapPair.ETHTokenContractAddr)
	engine.ethToBscContractAddr[ethcom.HexToAddress(swapPair.ETHTokenContractAddr)] = ethcom.HexToAddress(swapPair.BSCTokenContractAddr)

	engine.newSwapPairSignal <- ethcom.HexToAddress(swapPair.BSCTokenContractAddr)
	return nil
}

func (engine *SwapEngine) GetSwapPairInstance(bscTokenAddr ethcom.Address) *SwapPairIns {
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()

	tokenInstance, ok := engine.swapPairs[bscTokenAddr]
	if !ok {
		return nil
	}
	return tokenInstance
}

func (engine *SwapEngine) UpdateSwapInstance(swapPair *model.SwapPair) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	bscTokenAddr := ethcom.HexToAddress(swapPair.BSCTokenContractAddr)
	tokenInstance, ok := engine.swapPairs[bscTokenAddr]
	if !ok {
		return
	}

	upperBound := big.NewInt(0)
	_, ok = upperBound.SetString(swapPair.UpperBound, 10)
	tokenInstance.UpperBound = upperBound

	lowBound := big.NewInt(0)
	_, ok = upperBound.SetString(swapPair.LowBound, 10)
	tokenInstance.LowBound = lowBound

	engine.swapPairs[bscTokenAddr] = tokenInstance
}
