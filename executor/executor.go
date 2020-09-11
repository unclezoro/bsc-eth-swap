package executor

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/binance-chain/bsc-eth-swap/common"
	swapproxy "github.com/binance-chain/bsc-eth-swap/executor/abi"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type BscExecutor interface {
	GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error)
}

type ChainExecutor struct {
	Config *util.Config

	CrossChainAbi abi.ABI
	Client        *ethclient.Client
}

func NewExecutor(provider string, config *util.Config) *ChainExecutor {
	crossChainAbi, err := abi.JSON(strings.NewReader(swapproxy.SwapProxyABI))
	if err != nil {
		panic("marshal abi error")
	}

	client, err := ethclient.Dial(provider)
	if err != nil {
		panic("new eth client error")
	}

	return &ChainExecutor{
		Config:        config,
		CrossChainAbi: crossChainAbi,
		Client:        client,
	}
}
