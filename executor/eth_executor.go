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

	agent "github.com/binance-chain/bsc-eth-swap/abi"
	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type EthExecutor struct {
	Chain  string
	Config *util.Config

	SwapAgentAddr ethcmm.Address
	SwapAgentAbi  abi.ABI
	Client        *ethclient.Client
}

func NewEthExecutor(ethClient *ethclient.Client, swapAddr string, config *util.Config) *EthExecutor {
	agentAbi, err := abi.JSON(strings.NewReader(agent.ETHSwapAgentABI))
	if err != nil {
		panic("marshal abi error")
	}

	return &EthExecutor{
		Chain:         common.ChainETH,
		Config:        config,
		SwapAgentAddr: ethcmm.HexToAddress(swapAddr),
		SwapAgentAbi:  agentAbi,
		Client:        ethClient,
	}
}

func (e *EthExecutor) GetChainName() string {
	return e.Chain
}

func (e *EthExecutor) GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error) {
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

func (e *EthExecutor) GetLogs(header *types.Header) ([]interface{}, error) {
	startEvs, err := e.GetSwapStartLogs(header)
	if err != nil {
		return nil, err
	}
	regiserEvs, err := e.GetSwapStartLogs(header)
	if err != nil {
		return nil, err
	}
	var res = make([]interface{}, 0, len(startEvs)+len(regiserEvs))
	res = append(append(res, startEvs), regiserEvs)
	return res, nil

}

func (e *EthExecutor) GetSwapPairRegisterLogs(header *types.Header) ([]interface{}, error) {
	topics := [][]ethcmm.Hash{{SwapPairRegisterEventHash}}

	blockHash := header.Hash()

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logs, err := e.Client.FilterLogs(ctxWithTimeout, ethereum.FilterQuery{
		BlockHash: &blockHash,
		Topics:    topics,
		Addresses: []ethcmm.Address{e.SwapAgentAddr},
	})
	if err != nil {
		return nil, err
	}

	eventModels := make([]interface{}, 0, len(logs))
	for _, log := range logs {
		util.Logger.Debugf("get log: %d, %s, %s", log.BlockNumber, log.Topics[0].String(), log.TxHash.String())

		event, err := ParseSwapPairRegisterEvent(&e.SwapAgentAbi, &log)
		if err != nil {
			util.Logger.Errorf("parse event log error, er=%s", err.Error())
			continue
		}

		if event == nil {
			continue
		}

		eventModel := event.ToSwapPairRegisterLog(&log)
		eventModel.Chain = e.Chain

		eventModels = append(eventModels, eventModel)
	}
	return eventModels, nil
}

func (e *EthExecutor) GetSwapStartLogs(header *types.Header) ([]interface{}, error) {
	topics := [][]ethcmm.Hash{{SwapStartedEventHash}}

	blockHash := header.Hash()

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logs, err := e.Client.FilterLogs(ctxWithTimeout, ethereum.FilterQuery{
		BlockHash: &blockHash,
		Topics:    topics,
		Addresses: []ethcmm.Address{e.SwapAgentAddr},
	})
	if err != nil {
		return nil, err
	}

	eventModels := make([]interface{}, 0, len(logs))
	for _, log := range logs {
		util.Logger.Debugf("get log: %d, %s, %s", log.BlockNumber, log.Topics[0].String(), log.TxHash.String())

		event, err := ParseSwapStartEvent(&e.SwapAgentAbi, &log)
		if err != nil {
			util.Logger.Errorf("parse event log error, er=%s", err.Error())
			continue
		}

		if event == nil {
			continue
		}

		eventModel := event.ToSwapStartTxLog(&log)
		eventModel.Chain = e.Chain

		eventModels = append(eventModels, eventModel)
	}
	return eventModels, nil
}
