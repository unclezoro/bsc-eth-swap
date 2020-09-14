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
	TxStatusPending   TxStatus = 1
	TxStatusConfirmed TxStatus = 2
)

type TxEventLog struct {
	gorm.Model

	Chain string

	ContractAddress string
	FromAddress     string
	ToAddress       string
	Amount          string
	FeeAmount       string

	Status       TxStatus
	TxHash       string
	BlockHash    string
	Height       int64
	ConfirmedNum int64
}

func (TxEventLog) TableName() string {
	return "tx_event_log"
}

func (l *TxEventLog) BeforeCreate() (err error) {
	l.Model.CreatedAt = time.Now()
	l.Model.UpdatedAt = time.Now()
	return nil
}

type SwapTx struct {
	gorm.Model

	SourceChain         string `gorm:"not null;index:source_chain"`
	SwapRequestTxHash   string `gorm:"not null;index:swap_request_tx_hash"`
	SourceAssetContract string `gorm:"not null;index:source_asset_contract"`
	Symbol              string `gorm:"not null"`
	Decimals            int8   `gorm:"not null"`
	Amount              string `gorm:"not null"`
	Recipient           string `gorm:"not null"`

	DestiChain         string `gorm:"not null;index:desti_chain"`
	DestiAssetContract string `gorm:"not null;index:desti_chain"`
	TxHash             string `gorm:"not null"`
	ConsumedFeeAmount  string
	BlockHash          string
	Height             int64
	ConfirmedNum       int64
	Status             TxStatus `gorm:"not null"`
}

func (SwapTx) TableName() string {
	return "swap_tx_log"
}

func (l *SwapTx) BeforeCreate() (err error) {
	l.Model.CreatedAt = time.Now()
	l.Model.UpdatedAt = time.Now()
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

	if !db.HasTable(&SwapTx{}) {
		db.CreateTable(&SwapTx{})
	}
}
