package swap

import (
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/util"
	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
)

const (
	SwapTokenReceived  common.SwapStatus = "received"
	SwapQuoteRejected  common.SwapStatus = "rejected"
	SwapQuoteConfirmed common.SwapStatus = "confirmed"
	SwapQuoteSending   common.SwapStatus = "sending"
	SwapSent           common.SwapStatus = "sent"
	SwapSendFailed     common.SwapStatus = "sent_fail"
	SwapSuccess        common.SwapStatus = "sent_success"

	SwapEth2BSC common.SwapDirection = "eth_bsc"
	SwapBSC2Eth common.SwapDirection = "bsc_eth"

	BatchSize            = 50
	TrackSentTxBatchSize = 100
	SleepTime            = 10
	SwapSleepSecond      = 5

	TxFailedStatus = 0x00
)

type Swapper struct {
	Mutex                   sync.RWMutex
	DB                      *gorm.DB
	Config                  *util.Config
	TokenInstances          map[string]*TokenInstance
	ETHClient               *ethclient.Client
	BSCClient               *ethclient.Client
	BSCContractAddrToSymbol map[string]string
	ETHContractAddrToSymbol map[string]string
	NewTokenSignal          chan string
}

type TokenInstance struct {
	Symbol      string
	Name        string
	Decimals    int
	LowBound    *big.Int
	UpperBound  *big.Int
	CloseSignal chan bool

	BSCPrivateKey   *ecdsa.PrivateKey
	BSCTxSender     ethcom.Address
	BSCContractAddr ethcom.Address
	ETHPrivateKey   *ecdsa.PrivateKey
	ETHContractAddr ethcom.Address
	ETHTxSender     ethcom.Address
}
