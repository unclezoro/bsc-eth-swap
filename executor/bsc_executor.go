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
	contractabi "github.com/binance-chain/bsc-eth-swap/abi"
	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type BscExecutor struct {
	Chain  string
	Config *util.Config

	SwapAgentAddr    ethcmm.Address
	BSCSwapAgentInst *contractabi.ETHSwapAgent
	SwapAgentAbi     abi.ABI
	Client           *ethclient.Client
}

func NewBSCExecutor(ethClient *ethclient.Client, swapAddr string, config *util.Config) *BscExecutor {
	agentAbi, err := abi.JSON(strings.NewReader(agent.BSCSwapAgentABI))
	if err != nil {
		panic("marshal abi error")
	}

	bscSwapAgentInst, err := contractabi.NewETHSwapAgent(ethcmm.HexToAddress(swapAddr), ethClient)
	if err != nil {
		panic(err.Error())
	}

	return &BscExecutor{
		Chain:            common.ChainBSC,
		Config:           config,
		SwapAgentAddr:    ethcmm.HexToAddress(swapAddr),
		BSCSwapAgentInst: bscSwapAgentInst,
		SwapAgentAbi:     agentAbi,
		Client:           ethClient,
	}
}

func (e *BscExecutor) GetChainName() string {
	return e.Chain
}

func (e *BscExecutor) GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error) {
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
func (e *BscExecutor) GetLogs(header *types.Header) ([]interface{}, error) {
	return e.GetSwapStartLogs(header)
}

func (e *BscExecutor) GetSwapStartLogs(header *types.Header) ([]interface{}, error) {
	topics := [][]ethcmm.Hash{{BSC2ETHSwapStartedEventHash}}

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
		event, err := ParseBSC2ETHSwapStartEvent(&e.SwapAgentAbi, &log)
		if err != nil {
			util.Logger.Errorf("parse event log error, er=%s", err.Error())
			continue
		}
		if event == nil {
			continue
		}

		eventModel := event.ToSwapStartTxLog(&log)
		eventModel.Chain = e.Chain
		util.Logger.Debugf("Found BSC2ETH swap, txHash: %s, token address: %s, amount: %s, fee amount: %s",
			eventModel.TxHash, eventModel.TokenAddr, eventModel.Amount, eventModel.FeeAmount)
		eventModels = append(eventModels, eventModel)
	}
	return eventModels, nil
}
