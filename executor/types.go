package executor

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	common2 "github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
)

var (
	TokenTransferEventName = "tokenTransfer"

	TokenTransferEventHash = common.HexToHash("0xfb08937c18a8d4b15e559e41ae0e3a6be8c85434f744c0743f224e9b48fdc4e5")
)

type TokenTransferEvent struct {
	ContractAddr common.Address
	FromAddr     common.Address
	ToAddr       common.Address
	Amount       *big.Int
	FeeAmount    *big.Int
}

func (ev *TokenTransferEvent) ToTxLog(log *types.Log) interface{} {
	pack := &model.TxEventLog{
		Chain: common2.ChainETH,

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

func ParseTokenTransferEvent(abi *abi.ABI, log *types.Log) (*TokenTransferEvent, error) {
	var ev TokenTransferEvent

	err := abi.Unpack(&ev, TokenTransferEventName, log.Data)
	if err != nil {
		return nil, err
	}

	ev.Amount = big.NewInt(0).SetBytes(log.Topics[4].Bytes())
	ev.FeeAmount = big.NewInt(0).SetBytes(log.Topics[5].Bytes())

	return &ev, nil
}
