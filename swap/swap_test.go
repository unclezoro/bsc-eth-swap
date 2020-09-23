package swap

import (
	"strings"
	"testing"

	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/require"
)

func TestSwap_SwapInstanceCheck(t *testing.T) {
	config := getTestConfig()

	bscClient, err := ethclient.Dial(config.ChainConfig.BSCProvider)
	require.NoError(t, err)

	ethClient, err := ethclient.Dial(config.ChainConfig.ETHProvider)
	require.NoError(t, err)

	db, err := prepareDB(config)
	require.NoError(t, err)

	err = insertABCToken(db)
	require.NoError(t, err)

	err = insertDEFToken(db)
	require.NoError(t, err)

	swapInstance, err := prepareTest(db, bscClient, ethClient, config)
	require.NoError(t, err)

	require.Equal(t, 2, len(swapInstance.TokenInstances))

	tokens := make([]model.Token, 0)
	db.Find(&tokens)

	for _, token := range tokens {
		require.Equal(t, token.Symbol, swapInstance.BSCContractAddrToSymbol[strings.ToLower(token.BSCTokenContractAddr)])
		require.Equal(t, token.Symbol, swapInstance.ETHContractAddrToSymbol[strings.ToLower(token.ETHTokenContractAddr)])
	}

}
