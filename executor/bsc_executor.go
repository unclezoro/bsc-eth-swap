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

type BscExecutor struct {
	Chain  string
	Config *util.Config

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
	blockHeight := header.Number.Uint64()
	swapStartedIterator, err := e.BSCSwapAgentInst.FilterSwapStarted(&bind.FilterOpts{
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
