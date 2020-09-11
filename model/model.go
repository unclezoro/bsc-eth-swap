package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type BlockLog struct {
	Id         int64
	Chain      string
	BlockHash  string
	ParentHash string
	Height     int64
	BlockTime  int64
	CreateTime int64
}

func (BlockLog) TableName() string {
	return "block_log"
}

func (l *BlockLog) BeforeCreate() (err error) {
	l.CreateTime = time.Now().Unix()
	return nil
}

type TxStatus int

const (
	TxStatusInit      TxStatus = 0
	TxStatusConfirmed TxStatus = 1
)

type TxEventLog struct {
	Id    int64
	Chain string

	ContractAddress string
	FromAddress     string
	ToAddress       string
	Amount          string

	Status       TxStatus
	TxHash       string
	BlockHash    string
	Height       int64
	ConfirmedNum int64
	CreateTime   int64
	UpdateTime   int64
}

func (TxEventLog) TableName() string {
	return "tx_event_log"
}

func (l *TxEventLog) BeforeCreate() (err error) {
	l.CreateTime = time.Now().Unix()
	l.UpdateTime = time.Now().Unix()
	return nil
}

func InitTables(db *gorm.DB) {
	if !db.HasTable(&BlockLog{}) {
		db.CreateTable(&BlockLog{})
		db.Model(&BlockLog{}).AddIndex("idx_block_log_height", "height")
		db.Model(&BlockLog{}).AddIndex("idx_block_log_create_time", "create_time")
	}

	if !db.HasTable(&TxEventLog{}) {
		db.CreateTable(&TxEventLog{})
		db.Model(&TxEventLog{}).AddIndex("idx_event_log_tx_hash", "tx_hash")
		db.Model(&TxEventLog{}).AddIndex("idx_event_log_height", "height")
		db.Model(&TxEventLog{}).AddIndex("idx_event_log_create_time", "create_time")
	}
}
