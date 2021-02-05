package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type TxPhase int
type TxStatus int
type FillTxStatus int
type FillRetryTxStatus int

const (
	SeenRequest    TxPhase = 0
	ConfirmRequest TxPhase = 1
	AckRequest     TxPhase = 2

	TxStatusInit      TxStatus = 0
	TxStatusConfirmed TxStatus = 1

	FillTxCreated FillTxStatus = 0
	FillTxSent    FillTxStatus = 1
	FillTxSuccess FillTxStatus = 2
	FillTxFailed  FillTxStatus = 3
	FillTxMissing FillTxStatus = 4

	FillRetryTxCreated FillRetryTxStatus = 0
	FillRetryTxSent    FillRetryTxStatus = 1
	FillRetryTxSuccess FillRetryTxStatus = 2
	FillRetryTxFailed  FillRetryTxStatus = 3
	FillRetryTxMissing FillRetryTxStatus = 4
)

type BlockLog struct {
	Id         int64
	Chain      string `gorm:"not null;index:block_log_chain"`
	BlockHash  string `gorm:"not null;index:block_log_block_hash"`
	ParentHash string `gorm:"not null;index:block_log_parent_hash"`
	Height     int64  `gorm:"not null;index:block_log_height"`
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

func InitTables(db *gorm.DB) {
	db.AutoMigrate(&SwapPair{})
	db.AutoMigrate(&SwapFillTx{})
	db.AutoMigrate(&Swap{})
	db.AutoMigrate(&SwapStartTxLog{})
	db.AutoMigrate(&BlockLog{})
	db.AutoMigrate(&SwapPairCreatTx{})
	db.AutoMigrate(&SwapPairRegisterTxLog{})
	db.AutoMigrate(&SwapPairStateMachine{})
	db.AutoMigrate(&RetrySwap{})
	db.AutoMigrate(&RetrySwapTx{})
}
