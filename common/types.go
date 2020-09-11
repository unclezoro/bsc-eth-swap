package common

import "time"

const (
	ObserverMaxBlockNumber = 10000
	ObserverPruneInterval  = 10 * time.Second
	ObserverAlertInterval  = 5 * time.Second
	ObserverFetchInterval  = 1 * time.Second
)

const (
	ChainBSC = "bsc" // binance smart chain
	ChainBBC = "bbc" // binance beacon chain

	CoinBNB   = "BNB"
	CoinOther = "OTHER"

	ChannelIdTransferIn  = 3
	ChannelIdTransferOut = 2
)

const (
	DBDialectMysql   = "mysql"
	DBDialectSqlite3 = "sqlite3"
)

type BlockAndEventLogs struct {
	Height          int64
	Chain           string
	BlockHash       string
	ParentBlockHash string
	BlockTime       int64
	Events          []interface{}
}
