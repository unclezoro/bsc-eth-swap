package swap

import "github.com/binance-chain/bsc-eth-swap/common"

const (
	SwapTokenReceived common.SwapStatus = "received"
	SwapQuoteRejected common.SwapStatus = "rejected"
	SwapSent          common.SwapStatus = "sent"
	SwapSendFailed    common.SwapStatus = "sent_fail"
	SwapSuccess       common.SwapStatus = "sent_success"

	SwapEth2BSC common.SwapDirection = "eth_bsc"
	SwapBSC2Eth common.SwapDirection = "bsc_eth"

	BatchSize = 10
	SleepSecond = 10

	MaxTryGenerateUUID = 10
)

