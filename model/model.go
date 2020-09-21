package model

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/common"
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

type DepositTxStatus int
type TxPhase int
type WithdrawTxStatus int

const (
	TxStatusInit      DepositTxStatus = 0
	TxStatusConfirmed DepositTxStatus = 1

	SeenSwapRequest    TxPhase = 0
	ConfirmSwapRequest TxPhase = 1
	AckSwapRequest     TxPhase = 2

	WithdrawTxCreated WithdrawTxStatus = 0
	WithdrawTxSent    WithdrawTxStatus = 1
	WithdrawTxSuccess WithdrawTxStatus = 2
	WithdrawTxFailed  WithdrawTxStatus = 3
	WithdrawTxMissing WithdrawTxStatus = 4
)

type TxEventLog struct {
	Id    int64
	Chain string `gorm:"not null;index:tx_event_log_chain"`

	ContractAddress string `gorm:"not null"`
	FromAddress     string `gorm:"not null"`
	ToAddress       string `gorm:"not null"`
	Amount          string `gorm:"not null"`
	FeeAmount       string `gorm:"not null"`

	Status       DepositTxStatus `gorm:"not null;index:tx_event_log_status"`
	TxHash       string          `gorm:"not null;index:tx_event_log_tx_hash"`
	BlockHash    string          `gorm:"not null"`
	Height       int64           `gorm:"not null"`
	ConfirmedNum int64           `gorm:"not null"`

	Phase TxPhase `gorm:"not null;index:tx_event_log_phase"`

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
	gorm.Model

	Direction         common.SwapDirection `gorm:"not null"`
	DepositTxHash     string               `gorm:"not null;index:swap_tx_deposit_tx_hash"`
	WithdrawTxHash    string               `gorm:"not null;index:swap_tx_withdraw_tx_hash"`
	GasPrice          string               `gorm:"not null"`
	ConsumedFeeAmount string
	Height            int64
	Status            WithdrawTxStatus `gorm:"not null"`
	TrackRetryCounter int64
}

func (SwapTx) TableName() string {
	return "swap_txs"
}

type Swap struct {
	gorm.Model

	Status common.SwapStatus `gorm:"not null;index:swap_status"`
	// the user addreess who start this swap
	Sponsor string `gorm:"not null;index:sponsor"`

	Symbol    string               `gorm:"not null;index:swap_symbol"`
	Amount    string               `gorm:"not null;index:swap_amount"`
	Decimals  int                  `gorm:"not null"`
	Direction common.SwapDirection `gorm:"not null"`

	// The tx hash confirmed deposit
	DepositTxHash string `gorm:"not null;index:swap_deposit_tx_hash"`
	// The tx hash confirmed withdraw
	WithdrawTxHash string `gorm:"not null;index:swap_withdraw_tx_hash"`

	// The tx hash of refund
	RefundTxHash string

	// used to log more message about how this swap failed or invalid
	Log string
}

func (Swap) TableName() string {
	return "swaps"
}

type Token struct {
	gorm.Model

	Symbol               string `gorm:"unique;not null;index:symbol"`
	Name                 string `gorm:"not null"`
	Decimals             int    `gorm:"not null"`
	BSCTokenContractAddr string `gorm:"unique;not null"`
	ETHTokenContractAddr string `gorm:"unique;not null"`
	Available            bool   `gorm:"not null;index:available"`
	LowBound             string `gorm:"not null"`
	UpperBound           string `gorm:"not null"`

	IconUrl string

	BSCKeyType          string `gorm:"not null"`
	BSCKeyAWSRegion     string
	BSCKeyAWSSecretName string
	BSCPrivateKey       string // won't present in production environment
	BSCSenderAddr       string `gorm:"not null"`
	BSCERC20Threshold   string `gorm:"not null"`

	ETHKeyType          string `gorm:"not null"`
	ETHKeyAWSRegion     string
	ETHKeyAWSSecretName string
	ETHPrivateKey       string // won't present in production environment
	ETHSenderAddr       string `gorm:"not null"`
	ETHERC20Threshold   string `gorm:"not null"`
}

func (Token) TableName() string {
	return "tokens"
}

func InitTables(db *gorm.DB) {
	db.AutoMigrate(&Token{}, &SwapTx{}, &Swap{}, &TxEventLog{}, &BlockLog{})
}
