package executor

import (
	common "github.com/binance-chain/bsc-eth-swap/common"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/binance-chain/bsc-eth-swap/model"
)

type Executor interface {
	GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error)
	GetChainName() string
}

// ===================  SwapStarted =============
var (
	SwapStartedEventName = "SwapStarted"
	SwapStartedEventHash = ethcmm.HexToHash("0xf60309f865a6aa297da5fac6188136a02e5acfdf6e8f6d35257a9f4e9653170f")
)

type SwapStartedEvent struct {
	ContractAddr ethcmm.Address
	FromAddr     ethcmm.Address
	ToAddr       ethcmm.Address
	Amount       *big.Int
	FeeAmount    *big.Int
}

func (ev *SwapStartedEvent) ToSwapStartTxLog(log *types.Log) *model.SwapStartTxLog {
	pack := &model.SwapStartTxLog{
		ContractAddress: ev.ContractAddr.String(),
		FromAddress:     ev.FromAddr.String(),
		ToAddress:       ev.ToAddr.String(),
		Amount:          ev.Amount.String(),

		FeeAmount: ev.FeeAmount.String(),
		BlockHash: log.BlockHash.Hex(),
		TxHash:    log.TxHash.String(),
		Height:    int64(log.BlockNumber),
	}
	return pack
}

func ParseSwapStartEvent(abi *abi.ABI, log *types.Log) (*SwapStartedEvent, error) {
	var ev SwapStartedEvent

	err := abi.Unpack(&ev, SwapStartedEventName, log.Data)
	if err != nil {
		return nil, err
	}

	ev.ContractAddr = ethcmm.BytesToAddress(log.Topics[1].Bytes())
	ev.FromAddr = ethcmm.BytesToAddress(log.Topics[2].Bytes())
	ev.ToAddr = ethcmm.BytesToAddress(log.Topics[3].Bytes())

	return &ev, nil
}

// =================  SwapPairRegister ===================
var (
	SwapPairRegisterEventName = "SwapPairRegister"
	SwapPairRegisterEventHash = ethcmm.HexToHash("0xfe3bd005e346323fa452df8cafc28c55b99e3766ba8750571d139c6cf5bc08a0")
)

type SwapPairRegisterEvent struct {
	Sponsor      ethcmm.Address
	ContractAddr ethcmm.Address
	Name         string
	Symbol       string
	Decimals     uint8
}

func (ev *SwapPairRegisterEvent) ToSwapPairRegisterLog(log *types.Log) *model.SwapPairRegisterTxLog {
	pack := &model.SwapPairRegisterTxLog{
		ETHTokenContractAddr: ev.ContractAddr.String(),
		Sponsor:              ev.Sponsor.String(),
		Symbol:               ev.Symbol,
		Name:                 ev.Name,
		Decimals:             int(ev.Decimals),

		BlockHash: log.BlockHash.Hex(),
		TxHash:    log.TxHash.String(),
		Height:    int64(log.BlockNumber),
	}
	return pack
}

func ParseSwapPairRegisterEvent(abi *abi.ABI, log *types.Log) (*SwapPairRegisterEvent, error) {
	var ev SwapPairRegisterEvent

	err := abi.Unpack(&ev, SwapPairRegisterEventName, log.Data)
	if err != nil {
		return nil, err
	}
	ev.Sponsor = ethcmm.BytesToAddress(log.Topics[1].Bytes())
	ev.ContractAddr = ethcmm.BytesToAddress(log.Topics[2].Bytes())

	return &ev, nil
}
