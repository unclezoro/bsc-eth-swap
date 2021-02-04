package model

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/common"
)

type SwapPair struct {
	gorm.Model
	Sponsor    string `gorm:"not null;index:sponsor"`
	Symbol     string `gorm:"not null;index:symbol"`
	Name       string `gorm:"not null"`
	Decimals   int    `gorm:"not null"`
	BEP20Addr  string `gorm:"not null"`
	ERC20Addr  string `gorm:"not null"`
	Available  bool   `gorm:"not null;index:available"`
	LowBound   string `gorm:"not null"`
	UpperBound string `gorm:"not null"`
	IconUrl    string

	RecordHash string `gorm:"not null"`
}

func (SwapPair) TableName() string {
	return "swap_pairs"
}

type SwapPairRegisterTxLog struct {
	Id    int64
	Chain string `gorm:"not null;index:swappair_register_tx_log_chain"`

	Sponsor   string `gorm:"not null"`
	ERC20Addr string `gorm:"not null"`
	Symbol    string `gorm:"not null;index:swappair_register_tx_log_symbol"`
	Name      string `gorm:"not null"`
	Decimals  int    `gorm:"not null"`

	Status       TxStatus `gorm:"not null;index:swappair_register_tx_log_status"`
	TxHash       string   `gorm:"not null;index:swappair_register_tx_log_tx_hash"`
	BlockHash    string   `gorm:"not null"`
	Height       int64    `gorm:"not null"`
	ConfirmedNum int64    `gorm:"not null"`

	Phase TxPhase `gorm:"not null;index:swappair_register_tx_log_phase"`

	UpdateTime int64
	CreateTime int64
}

func (SwapPairRegisterTxLog) TableName() string {
	return "swap_pair_register_tx"
}

func (l *SwapPairRegisterTxLog) BeforeCreate() (err error) {
	l.CreateTime = time.Now().Unix()
	l.UpdateTime = time.Now().Unix()
	return nil
}

type SwapPairCreatTx struct {
	gorm.Model

	SwapPairRegisterTxHash string `gorm:"unique;not null"`
	SwapPairCreatTxHash    string `gorm:"unique;not null"`

	ERC20Addr string `gorm:"not null"`

	Symbol   string `gorm:"not null;index:swap_pair_creat_tx_symbol"`
	Name     string `gorm:"not null"`
	Decimals int    `gorm:"not null"`

	GasPrice          string `gorm:"not null"`
	ConsumedFeeAmount string
	Height            int64
	Status            FillTxStatus `gorm:"not null"`
	TrackRetryCounter int64
}

func (SwapPairCreatTx) TableName() string {
	return "swap_pair_creat_tx"
}

type SwapPairStateMachine struct {
	gorm.Model

	Status common.SwapPairStatus `gorm:"not null;index:swap_pair_sm_status"`

	ERC20Addr string `gorm:"not null"`
	BEP20Addr string

	Sponsor  string `gorm:"not null"`
	Symbol   string `gorm:"not null;index:swap_pair_sm_symbol"`
	Name     string `gorm:"not null"`
	Decimals int    `gorm:"not null"`

	PairRegisterTxHash string `gorm:"not null"`
	PairCreatTxHash    string

	// used to log more message about how this swap_pair failed or invalid
	Log string

	RecordHash string `gorm:"not null"`
}

func (SwapPairStateMachine) TableName() string {
	return "swap_pair_sm"
}
