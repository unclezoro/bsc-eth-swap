package observer

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/executor"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type Observer struct {
	DB *gorm.DB

	StartHeight int64
	ConfirmNum  int64

	Config   *util.Config
	Executor executor.Executor
}

// NewObserver returns the observer instance
func NewObserver(db *gorm.DB, startHeight, confirmNum int64, cfg *util.Config, executor executor.Executor) *Observer {
	return &Observer{
		DB: db,

		StartHeight: startHeight,
		ConfirmNum:  confirmNum,

		Config:   cfg,
		Executor: executor,
	}
}

// Start starts the routines of observer
func (ob *Observer) Start() {
	go ob.Fetch(ob.StartHeight)
	go ob.Prune()
	go ob.Alert()
}

func (ob *Observer) fetchSleep() {
	if ob.Executor.GetChainName() == common.ChainBSC {
		time.Sleep(time.Duration(ob.Config.ChainConfig.BSCObserverFetchInterval) * time.Second)
	} else if ob.Executor.GetChainName() == common.ChainETH {
		time.Sleep(time.Duration(ob.Config.ChainConfig.ETHObserverFetchInterval) * time.Second)
	}

}

// Fetch starts the main routine for fetching blocks of BSC
func (ob *Observer) Fetch(startHeight int64) {
	for {
		curBlockLog, err := ob.GetCurrentBlockLog()
		if err != nil {
			util.Logger.Errorf("get current block log from db error: %s", err.Error())
			ob.fetchSleep()
			continue
		}

		nextHeight := curBlockLog.Height + 1
		if curBlockLog.Height == 0 && startHeight != 0 {
			nextHeight = startHeight
		}

		util.Logger.Debugf("fetch %s block, height=%d", ob.Executor.GetChainName(), nextHeight)
		err = ob.fetchBlock(curBlockLog.Height, nextHeight, curBlockLog.BlockHash)
		if err != nil {
			util.Logger.Debugf("fetch %s block error, err=%s", ob.Executor.GetChainName(), err.Error())
			ob.fetchSleep()
		}
	}
}

// fetchBlock fetches the next block of BSC and saves it to database. if the next block hash
// does not match to the parent hash, the current block will be deleted for there is a fork.
func (ob *Observer) fetchBlock(curHeight, nextHeight int64, curBlockHash string) error {
	blockAndEventLogs, err := ob.Executor.GetBlockAndTxEvents(nextHeight)
	if err != nil {
		return fmt.Errorf("get block info error, height=%d, err=%s", nextHeight, err.Error())
	}

	parentHash := blockAndEventLogs.ParentBlockHash
	if curHeight != 0 && parentHash != curBlockHash {
		return ob.DeleteBlockAndTxEvents(curHeight)
	} else {
		nextBlockLog := model.BlockLog{
			BlockHash:  blockAndEventLogs.BlockHash,
			ParentHash: parentHash,
			Height:     blockAndEventLogs.Height,
			BlockTime:  blockAndEventLogs.BlockTime,
			Chain:      blockAndEventLogs.Chain,
		}

		err := ob.SaveBlockAndTxEvents(&nextBlockLog, blockAndEventLogs.Events)
		if err != nil {
			return err
		}

		err = ob.UpdateSwapStartConfirmedNum(nextBlockLog.Height)
		if err != nil {
			return err
		}
		err = ob.UpdateSwapPairRegisterConfirmedNum(nextBlockLog.Height)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteBlockAndTxEvents deletes the block and txs of the given height
func (ob *Observer) DeleteBlockAndTxEvents(height int64) error {
	tx := ob.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Where("height = ?", height).Delete(model.BlockLog{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	txEventLogList := make([]model.SwapStartTxLog, 0)
	ob.DB.Where("chain = ? and height = ? and status = ?", ob.Executor.GetChainName(), height, model.TxStatusInit).Find(&txEventLogList)
	for _, txEventLog := range txEventLogList {
		if err := tx.Where("start_tx_hash = ?", txEventLog.TxHash).Delete(model.Swap{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Where("chain = ? and height = ? and status = ?", ob.Executor.GetChainName(), height, model.TxStatusInit).Delete(model.SwapStartTxLog{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	registerLogList := make([]model.SwapPairRegisterTxLog, 0)
	ob.DB.Where("chain = ? and height = ? and status = ?", ob.Executor.GetChainName(), height, model.TxStatusInit).Find(&registerLogList)
	for _, registerLog := range registerLogList {
		if err := tx.Where("swap_pair_register_tx_hash = ?", registerLog.TxHash).Delete(model.SwapPairStateMachine{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Where("chain = ? and height = ? and status = ?", ob.Executor.GetChainName(), height, model.TxStatusInit).Delete(model.SwapPairRegisterTxLog{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (ob *Observer) UpdateSwapStartConfirmedNum(height int64) error {
	err := ob.DB.Model(model.SwapStartTxLog{}).Where("chain = ? and status = ?", ob.Executor.GetChainName(), model.TxStatusInit).Updates(
		map[string]interface{}{
			"confirmed_num": gorm.Expr("? - height", height+1),
		}).Error
	if err != nil {
		return err
	}

	err = ob.DB.Model(model.SwapStartTxLog{}).Where("chain = ? and status = ? and confirmed_num >= ?",
		ob.Executor.GetChainName(), model.TxStatusInit, ob.ConfirmNum).Updates(
		map[string]interface{}{
			"status": model.TxStatusConfirmed,
		}).Error
	if err != nil {
		return err
	}

	return nil
}

func (ob *Observer) UpdateSwapPairRegisterConfirmedNum(height int64) error {
	err := ob.DB.Model(model.SwapPairRegisterTxLog{}).Where("chain = ? and status = ?", ob.Executor.GetChainName(), model.TxStatusInit).Updates(
		map[string]interface{}{
			"confirmed_num": gorm.Expr("? - height", height+1),
		}).Error
	if err != nil {
		return err
	}

	err = ob.DB.Model(model.SwapPairRegisterTxLog{}).Where("chain = ? and status = ? and confirmed_num >= ?",
		ob.Executor.GetChainName(), model.TxStatusInit, ob.ConfirmNum).Updates(
		map[string]interface{}{
			"status": model.TxStatusConfirmed,
		}).Error
	if err != nil {
		return err
	}

	return nil
}

// Prune prunes the outdated blocks
func (ob *Observer) Prune() {
	for {
		curBlockLog, err := ob.GetCurrentBlockLog()
		if err != nil {
			util.Logger.Errorf("get current block log error, err=%s", err.Error())
			time.Sleep(common.ObserverPruneInterval)

			continue
		}
		err = ob.DB.Where("chain = ? and height < ?", ob.Executor.GetChainName(), curBlockLog.Height-common.ObserverMaxBlockNumber).Delete(model.BlockLog{}).Error
		if err != nil {
			util.Logger.Infof("prune block logs error, err=%s", err.Error())
		}
		time.Sleep(common.ObserverPruneInterval)
	}
}

func (ob *Observer) SaveBlockAndTxEvents(blockLog *model.BlockLog, packages []interface{}) error {
	tx := ob.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(blockLog).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, pack := range packages {
		if err := tx.Create(pack).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// GetCurrentBlockLog returns the highest block log
func (ob *Observer) GetCurrentBlockLog() (*model.BlockLog, error) {
	blockLog := model.BlockLog{}
	err := ob.DB.Where("chain = ?", ob.Executor.GetChainName()).Order("height desc").First(&blockLog).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &blockLog, nil
}

// Alert sends alerts to tg group if there is no new block fetched in a specific time
func (ob *Observer) Alert() {
	for {
		curOtherChainBlockLog, err := ob.GetCurrentBlockLog()
		if err != nil {
			util.Logger.Errorf("get current block log error, err=%s", err.Error())
			time.Sleep(common.ObserverAlertInterval)

			continue
		}
		if curOtherChainBlockLog.Height > 0 {
			if time.Now().Unix()-curOtherChainBlockLog.CreateTime > ob.Config.AlertConfig.BlockUpdateTimeout {
				msg := fmt.Sprintf("last block fetched at %s, chain=%s, height=%d",
					time.Unix(curOtherChainBlockLog.CreateTime, 0).String(), ob.Executor.GetChainName(), curOtherChainBlockLog.Height)
				util.SendTelegramMessage(msg)
			}
		}

		time.Sleep(common.ObserverAlertInterval)
	}
}
