package swap

import (
	"math/big"
	"sync"

	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	tsssdksecure "github.com/binance-chain/tss-zerotrust-sdk/secure"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/util"
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
	SleepTime            = 5
	SwapSleepSecond      = 2

	TxFailedStatus = 0x00
)

type Swapper struct {
	Mutex                   sync.RWMutex
	DB                      *gorm.DB
	HMACKey                 string
	Config                  *util.Config
	TokenInstances          map[string]*TokenInstance
	TSSClientSecureConfig   *tsssdksecure.ClientSecureConfig
	ETHClient               *ethclient.Client
	BSCClient               *ethclient.Client
	ETHChainID              int64
	BSCChainID				int64
	ETHTxSender             ethcom.Address
	BSCTxSender             ethcom.Address
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

	BSCTokenContractAddr ethcom.Address
	BSCERC20Threshold    *big.Int
	ETHTokenContractAddr ethcom.Address
	ETHERC20Threshold    *big.Int
}
