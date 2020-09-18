package executor

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/binance-chain/bsc-eth-swap/common"
	swapproxy "github.com/binance-chain/bsc-eth-swap/executor/abi"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type Executor interface {
	GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error)
	GetChainName() string
	GetContractDecimals(address ethcmm.Address) (int, error)
	GetContractSymbol(address ethcmm.Address) (string, error)
}

type ChainExecutor struct {
	Chain  string
	Config *util.Config

	SwapProxyAddr ethcmm.Address
	SwapProxyAbi  abi.ABI
	Client        *ethclient.Client
}

func NewExecutor(chain string, ethClient *ethclient.Client, swapAddr string, config *util.Config) *ChainExecutor {
	proxyAbi, err := abi.JSON(strings.NewReader(swapproxy.SwapProxyABI))
	if err != nil {
		panic("marshal abi error")
	}

	return &ChainExecutor{
		Chain:         chain,
		Config:        config,
		SwapProxyAddr: ethcmm.HexToAddress(swapAddr),
		SwapProxyAbi:  proxyAbi,
		Client:        ethClient,
	}
}

func (e *ChainExecutor) GetChainName() string {
	return e.Chain
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
		Chain:           e.Chain,
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
		util.Logger.Debugf("get log: %d, %s, %s", log.BlockNumber, log.Topics[0].String(), log.TxHash.String())

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

func (e *ChainExecutor) GetContractDecimals(address ethcmm.Address) (int, error) {
	instance, err := swapproxy.NewERC20(address, e.Client)
	if err != nil {
		return 0, err
	}
	decimals, err := instance.Decimals(&bind.CallOpts{})
	return int(decimals), err
}

func (e *ChainExecutor) GetContractSymbol(address ethcmm.Address) (string, error) {
	instance, err := swapproxy.NewERC20(address, e.Client)
	if err != nil {
		return "", err
	}
	symbol, err := instance.Symbol(&bind.CallOpts{})
	return symbol, err
}
