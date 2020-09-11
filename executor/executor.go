package executor

import (
	"github.com/binance-chain/bsc-eth-swap/common"
)

type BscExecutor interface {
	GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error)
}
