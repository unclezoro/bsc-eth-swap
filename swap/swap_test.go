package swap

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/require"
)

func TestSwap_InsertAndConfirmSwapRequest(t *testing.T) {
	config := getTestConfig()

	db, err := prepareDB(config)
	require.NoError(t, err)

	err = insertABCToken(db)
	require.NoError(t, err)

	bscClient, err := ethclient.Dial(config.ChainConfig.BSCProvider)
	require.NoError(t, err)

	ethClient, err := ethclient.Dial(config.ChainConfig.ETHProvider)
	require.NoError(t, err)

	swapInstance, err := NewSwapper(db, config, bscClient, ethClient)
	require.NoError(t, err)

	clientPrivateKey := "580c4a31fffda2006462183b6ecc82a67a7fc772568c5b7191e3ed5e4be1bf04"

	{
		go swapInstance.monitorSwapRequestDaemon()
		go swapInstance.confirmSwapRequestDaemon()
		go swapInstance.createSwapDaemon()
		go swapInstance.trackSwapTxDaemon()
	}

}
