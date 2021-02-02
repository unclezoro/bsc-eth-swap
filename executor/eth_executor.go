package executor

import (
	"bytes"
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	agent "github.com/binance-chain/bsc-eth-swap/abi"
	contractabi "github.com/binance-chain/bsc-eth-swap/abi"
	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type EthExecutor struct {
	Chain  string
	Config *util.Config

	ethSwapAgentInst *contractabi.ETHSwapAgent
	SwapAgentAbi     abi.ABI
	Client           *ethclient.Client
}

func NewEthExecutor(ethClient *ethclient.Client, swapAddr string, config *util.Config) *EthExecutor {
	agentAbi, err := abi.JSON(strings.NewReader(agent.ETHSwapAgentABI))
	if err != nil {
		panic("marshal abi error")
	}
	ethSwapAgentInst, err := contractabi.NewETHSwapAgent(ethcmm.HexToAddress(swapAddr), ethClient)
	if err != nil {
		panic(err.Error())
	}

	return &EthExecutor{
		Chain:            common.ChainETH,
		Config:           config,
		ethSwapAgentInst: ethSwapAgentInst,
		SwapAgentAbi:     agentAbi,
		Client:           ethClient,
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
	regiserEvs, err := e.GetSwapPairRegisterLogs(header)
	if err != nil {
		return nil, err
	}
	var res = make([]interface{}, 0, len(startEvs)+len(regiserEvs))
	res = append(append(res, startEvs...), regiserEvs...)
	return res, nil

}

func (e *EthExecutor) GetSwapPairRegisterLogs(header *types.Header) ([]interface{}, error) {
	blockHeight := header.Number.Uint64()
	registerEventsIterator, err := e.ethSwapAgentInst.FilterSwapPairRegister(&bind.FilterOpts{
		Start:   blockHeight,
		End:     &blockHeight,
		Context: context.Background(),
	}, nil, nil)
	if err != nil || registerEventsIterator == nil {
		return nil, err
	}

	blockHash := header.Hash()
	eventModels := make([]interface{}, 0)
	for registerEventsIterator.Next() {
		if !bytes.Equal(registerEventsIterator.Event.Raw.BlockHash[:], blockHash[:]) {
			util.Logger.Debugf("Block hash mismatch, height %d, expected block hash %s, got block hash %s", header.Number.Uint64(), blockHash.String(), registerEventsIterator.Event.Raw.BlockHash.String())
			continue
		}
		util.Logger.Debugf("Got register event log: %d, %s", header.Number, registerEventsIterator.Event.Raw.TxHash.String())
		eventModel := &model.SwapPairRegisterTxLog{
			ETHTokenContractAddr: registerEventsIterator.Event.Erc20Addr.String(),
			Symbol:               registerEventsIterator.Event.Symbol,
			Name:                 registerEventsIterator.Event.Name,
			Decimals:             int(registerEventsIterator.Event.Decimals),

			BlockHash: header.Hash().Hex(),
			TxHash:    registerEventsIterator.Event.Raw.TxHash.String(),
			Height:    header.Number.Int64(),
		}
		eventModel.Chain = e.Chain
		eventModels = append(eventModels, eventModel)
	}
	return eventModels, nil
}

func (e *EthExecutor) GetSwapStartLogs(header *types.Header) ([]interface{}, error) {
	blockHeight := header.Number.Uint64()
	swapStartedIterator, err := e.ethSwapAgentInst.FilterSwapStarted(&bind.FilterOpts{
		Start:   blockHeight,
		End:     &blockHeight,
		Context: context.Background(),
	}, nil, nil)
	if err != nil || swapStartedIterator == nil {
		return nil, err
	}

	blockHash := header.Hash()
	eventModels := make([]interface{}, 0)
	for swapStartedIterator.Next() {
		if !bytes.Equal(swapStartedIterator.Event.Raw.BlockHash[:], blockHash[:]) {
			util.Logger.Debugf("Block hash mismatch, height %d, expected block hash %s, got block hash %s", header.Number.Uint64(), blockHash.String(), swapStartedIterator.Event.Raw.BlockHash.String())
			continue
		}
		util.Logger.Debugf("Get swap start event log: %d, %s, %s", header.Number, swapStartedIterator.Event.Raw.TxHash.String())
		eventModel := &model.SwapStartTxLog{
			ContractAddress: swapStartedIterator.Event.Erc20Addr.String(),
			FromAddress:     swapStartedIterator.Event.FromAddr.String(),
			Amount:          swapStartedIterator.Event.Amount.String(),
			FeeAmount:       swapStartedIterator.Event.FeeAmount.String(),
			BlockHash:       header.Hash().Hex(),

			TxHash: swapStartedIterator.Event.Raw.TxHash.String(),
			Height: header.Number.Int64(),
		}
		eventModel.Chain = e.Chain

		eventModels = append(eventModels, eventModel)
	}
	return eventModels, nil
}
