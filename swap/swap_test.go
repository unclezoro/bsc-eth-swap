package swap

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func TestRLP(t *testing.T) {
	bytesStr := "0x0f8a81a843b9aca00828df9948a1a84726abe38764d34c848021f8860691fddb380b844a9059cbb00000000000000000000000037b8516a0f88e65d677229b402ec6c1e0e33300400000000000000000000000000000000000000000000000017979cfe362a00001ca076b4f4c3b7709940cfb08148bde0b1b9ae819055d73"
	//bytes , err := hexutil.Decode(bytesStr)
	//require.NoError(t, err)
	//
	//var signTx types.Transaction
	//err = rlp.DecodeBytes(bytes, &signTx)
	//require.NoError(t, err)

	rpcClient, _ := rpc.DialContext(context.Background(), "https://ropsten.infura.io/v3/16a60635f67b474cb24ce5ee843b96be")
	err := rpcClient.CallContext(context.Background(), nil, "eth_sendRawTransaction", bytesStr)
	require.NoError(t, err)
}
