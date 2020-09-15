package executor

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/binance-chain/bsc-eth-swap/common"
	swapproxy "github.com/binance-chain/bsc-eth-swap/executor/abi"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type Executor interface {
	GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error)
}

type ChainExecutor struct {
	Chain  string
	Config *util.Config

	SwapProxyAddr ethcmm.Address
	SwapProxyAbi  abi.ABI
	Client        *ethclient.Client
}

func NewExecutor(chain string, provider string, swapAddr ethcmm.Address, config *util.Config) *ChainExecutor {
	proxyAbi, err := abi.JSON(strings.NewReader(swapproxy.SwapProxyABI))
	if err != nil {
		panic("marshal abi error")
	}

	client, err := ethclient.Dial(provider)
	if err != nil {
		panic("new eth client error")
	}

	return &ChainExecutor{
		Chain:         chain,
		Config:        config,
		SwapProxyAddr: swapAddr,
		SwapProxyAbi:  proxyAbi,
		Client:        client,
	}
}

func (e *ChainExecutor) GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	header, err := e.Client.HeaderByNumber(ctxWithTimeout, big.NewInt(height))
	if err != nil {
		return nil, err
	}

	packageLogs, err := e.GetLogs(header)
	if err != nil {
		return nil, err
	}

	return &common.BlockAndEventLogs{
		Height:          height,
		BlockHash:       header.Hash().String(),
		ParentBlockHash: header.ParentHash.String(),
		BlockTime:       int64(header.Time),
		Events:          packageLogs,
	}, nil
}

func (e *ChainExecutor) GetLogs(header *types.Header) ([]interface{}, error) {
	topics := [][]ethcmm.Hash{{TokenTransferEventHash}}

	blockHash := header.Hash()

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logs, err := e.Client.FilterLogs(ctxWithTimeout, ethereum.FilterQuery{
		BlockHash: &blockHash,
		Topics:    topics,
		Addresses: []ethcmm.Address{e.SwapProxyAddr},
	})
	if err != nil {
		return nil, err
	}

	eventModels := make([]interface{}, 0, len(logs))
	for _, log := range logs {
		util.Logger.Infof("get log: %d, %s, %s", log.BlockNumber, log.Topics[0].String(), log.TxHash.String())

		event, err := ParseTokenTransferEvent(&e.SwapProxyAbi, &log)
		if err != nil {
			util.Logger.Errorf("parse event log error, er=%s", err.Error())
			continue
		}

		if event == nil {
			continue
		}

		eventModel := event.ToTxLog(&log)
		eventModel.Chain = e.Chain

		eventModels = append(eventModels, eventModel)
	}
	return eventModels, nil
}
