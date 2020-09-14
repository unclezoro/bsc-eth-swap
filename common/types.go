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
	ChainETH = "eth" // ethereum
)

const (
	DBDialectMysql   = "mysql"
	DBDialectSqlite3 = "sqlite3"

	LocalPrivateKey = "local_private_key"
	AWSPrivateKey   = "aws_private_key"
)

type BlockAndEventLogs struct {
	Height          int64
	Chain           string
	BlockHash       string
	ParentBlockHash string
	BlockTime       int64
	Events          []interface{}
}
