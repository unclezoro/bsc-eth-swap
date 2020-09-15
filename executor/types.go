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

	TokenTransferEventHash = common.HexToHash("0x05e8bebd9fcb5eb8e77fbd53c65340bcda78c0ff916583b5eff776e21316dced")
)

type TokenTransferEvent struct {
	ContractAddr common.Address
	FromAddr     common.Address
	ToAddr       common.Address
	Amount       *big.Int
	FeeAmount    *big.Int
}

func (ev *TokenTransferEvent) ToTxLog(log *types.Log) *model.TxEventLog {
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

	ev.ContractAddr = common.BytesToAddress(log.Topics[1].Bytes())
	ev.FromAddr = common.BytesToAddress(log.Topics[2].Bytes())
	ev.ToAddr = common.BytesToAddress(log.Topics[3].Bytes())

	return &ev, nil
}
