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

// Fetch starts the main routine for fetching blocks of BSC
func (ob *Observer) Fetch(startHeight int64) {
	for {
		curBlockLog, err := ob.GetCurrentBlockLog()
		if err != nil {
			util.Logger.Errorf("get current block log error, err=%s", err.Error())
			time.Sleep(common.ObserverFetchInterval)
			continue
		}

		nextHeight := curBlockLog.Height + 1
		if curBlockLog.Height == 0 && startHeight != 0 {
			nextHeight = startHeight
		}

		util.Logger.Infof("fetch block, height=%d", nextHeight)
		err = ob.fetchBlock(curBlockLog.Height, nextHeight, curBlockLog.BlockHash)
		if err != nil {
			util.Logger.Errorf("fetch block error, err=%s", err.Error())
			time.Sleep(common.ObserverFetchInterval)
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
		}

		err := ob.SaveBlockAndTxEvents(&nextBlockLog, blockAndEventLogs.Events)
		if err != nil {
			return err
		}

		err = ob.UpdateConfirmedNum(nextBlockLog.Height)
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

	if err := tx.Where("height = ? and status = ?", height, model.TxStatusInit).Delete(model.TxEventLog{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (ob *Observer) UpdateConfirmedNum(height int64) error {
	err := ob.DB.Model(model.TxEventLog{}).Where("status = ?", model.TxStatusInit).Updates(
		map[string]interface{}{
			"confirmed_num": gorm.Expr("? - height", height+1),
		}).Error
	if err != nil {
		return err
	}

	err = ob.DB.Model(model.TxEventLog{}).Where("status = ? and confirmed_num >= ?",
		model.TxStatusInit, ob.ConfirmNum).Updates(
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
		err = ob.DB.Where("height < ?", curBlockLog.Height-common.ObserverMaxBlockNumber).Delete(model.BlockLog{}).Error
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
	err := ob.DB.Order("height desc").First(&blockLog).Error
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
				msg := fmt.Sprintf("last block fetched at %s, height=%d",
					time.Unix(curOtherChainBlockLog.CreateTime, 0).String(), curOtherChainBlockLog.Height)
				util.SendTelegramMessage(msg)
			}
		}

		time.Sleep(common.ObserverAlertInterval)
	}
}
