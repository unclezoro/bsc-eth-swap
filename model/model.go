package model

import (
	"time"

	"github.com/binance-chain/bsc-eth-swap/common"

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
type TxPhase int

const (
	TxStatusInit      TxStatus = 0
	TxStatusConfirmed TxStatus = 1

	SeenSwapRequest    TxPhase = 0
	ConfirmSwapRequest TxPhase = 1
	AckSwapRequest     TxPhase = 2
)

type TxEventLog struct {
	Id    int64
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

	Phase TxPhase

	UpdateTime int64
	CreateTime int64
}

func (TxEventLog) TableName() string {
	return "tx_event_log"
}

func (l *TxEventLog) BeforeCreate() (err error) {
	l.CreateTime = time.Now().Unix()
	l.UpdateTime = time.Now().Unix()
	return nil
}

type SwapTx struct {
	Id int64

	SourceChain       string `gorm:"not null;index:source_chain"`
	SwapRequestTxHash string `gorm:"not null;index:swap_request_tx_hash"`
	Symbol            string `gorm:"not null"`
	Amount            string `gorm:"not null"`

	DestiChain         string `gorm:"not null;index:desti_chain"`
	DestiAssetContract string `gorm:"not null;index:desti_chain"`
	TxHash             string `gorm:"not null"`
	ConsumedFeeAmount  string
	BlockHash          string
	Height             int64
	ConfirmedNum       int64
	Status             TxStatus `gorm:"not null"`

	UpdateTime int64
	CreateTime int64
}

func (SwapTx) TableName() string {
	return "swap_tx"
}

func (l *SwapTx) BeforeCreate() (err error) {
	l.CreateTime = time.Now().Unix()
	l.UpdateTime = time.Now().Unix()
	return nil
}

type Swap struct {
	gorm.Model

	UUID string `gorm:"unique;not null;index:swap_uuid"`

	Status common.SwapStatus `gorm:"not null;index:swap_status"`
	// the user addreess who start this swap
	Sponsor string `gorm:"not null;index:sponsor"`

	Symbol    string               `gorm:"not null;index:swap_symbol"`
	Amount    string               `gorm:"not null;index:swap_amount"`
	Direction common.SwapDirection `gorm:"not null"`

	// The tx hash confirmed deposit
	DepositTxHash string `gorm:"not null"`
	// The tx hash confirmed withdraw
	WithdrawTxHash string

	// used to log more message about how this swap failed or invalid
	Log string
}

func (Swap) TableName() string {
	return "swaps"
}


type Token struct {
	Id int64

	Symbol          string
	Name            string
	BSCContractAddr string
	ETHContractAddr string
	LowBound        string
	UpperBound      string

	BSCKeyType          string
	BSCKeyAWSRegion     string
	BSCKeyAWSSecretName string
	BSCPrivateKey       string

	ETHKeyType          string
	ETHKeyAWSRegion     string
	ETHKeyAWSSecretName string
	ETHPrivateKey       string

	UpdateTime int64
	CreateTime int64
}

func (Token) TableName() string {
	return "tokens"
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
		db.Model(&TxEventLog{}).AddIndex("idx_event_log_update_time", "update_time")
		db.Model(&TxEventLog{}).AddIndex("idx_event_log_status", "status")
		db.Model(&TxEventLog{}).AddIndex("idx_event_log_phase", "phase")
	}

	if !db.HasTable(&SwapTx{}) {
		db.CreateTable(&SwapTx{})
	}

	if !db.HasTable(&Token{}) {
		db.CreateTable(&Token{})
	}

	if !db.HasTable(&Swap{}) {
		db.CreateTable(&Swap{})
	}
}
