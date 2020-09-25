package swap

import (
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/require"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
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

func TestSwap_ETH2BSC(t *testing.T) {
	config := getTestConfig()

	bscClient, err := ethclient.Dial(config.ChainConfig.BSCProvider)
	require.NoError(t, err)

	ethClient, err := ethclient.Dial(config.ChainConfig.ETHProvider)
	require.NoError(t, err)

	db, err := prepareDB(config)
	require.NoError(t, err)

	err = insertABCToken(db)
	require.NoError(t, err)

	_, err = prepareTest(db, bscClient, ethClient, config)
	require.NoError(t, err)

	_, clientAccount, err := getClientAccount()
	require.NoError(t, err)

	txEventLog := model.TxEventLog{
		Chain: common.ChainETH,

		ContractAddress: "0x055d208b90DA0E3A431CA7E0fba326888Ef8a822",
		FromAddress:     clientAccount.String(),
		ToAddress:       "",
		Amount:          "1000",
		FeeAmount:       "100000",

		Status:       model.TxStatusInit,
		TxHash:       NewSHA1Hash(),
		BlockHash:    "",
		Height:       0,
		ConfirmedNum: 0,

		Phase: model.SeenSwapRequest,

		UpdateTime: time.Now().Unix(),
		CreateTime: time.Now().Unix(),
	}

	err = insertTxEventLogToDB(db, &txEventLog)
	require.NoError(t, err)

	time.Sleep(SleepTime * time.Second)
	time.Sleep(1 * time.Second)

	txEventLogs := make([]model.TxEventLog, 0)
	db.Order("id desc").Find(&txEventLogs)
	require.Equal(t, 1, len(txEventLogs))
	require.Equal(t, model.ConfirmSwapRequest, txEventLogs[0].Phase)

	swaps := make([]model.Swap, 0)
	db.Order("id desc").Find(&swaps)
	require.Equal(t, 1, len(swaps))
	require.Equal(t, txEventLog.TxHash, swaps[0].DepositTxHash)
	require.Equal(t, SwapQuoteRejected, swaps[0].Status)

	txEventLog = model.TxEventLog{
		Chain: common.ChainETH,

		ContractAddress: "0x055d208b90DA0E3A431CA7E0fba326888Ef8a822",
		FromAddress:     clientAccount.String(),
		ToAddress:       "",
		Amount:          "1000000000000000000",
		FeeAmount:       "100000",

		Status:       model.TxStatusInit,
		TxHash:       NewSHA1Hash(),
		BlockHash:    "",
		Height:       0,
		ConfirmedNum: 0,

		Phase: model.SeenSwapRequest,

		UpdateTime: time.Now().Unix(),
		CreateTime: time.Now().Unix(),
	}
	err = insertTxEventLogToDB(db, &txEventLog)
	require.NoError(t, err)

	time.Sleep((SleepTime + 1) * time.Second)

	txEventLogs = make([]model.TxEventLog, 0)
	db.Order("id desc").Find(&txEventLogs)
	require.Equal(t, 2, len(txEventLogs))
	require.Equal(t, model.ConfirmSwapRequest, txEventLogs[0].Phase)

	swaps = make([]model.Swap, 0)
	db.Order("id desc").Find(&swaps)
	require.Equal(t, 2, len(swaps))
	require.Equal(t, txEventLog.TxHash, swaps[0].DepositTxHash)
	require.Equal(t, SwapTokenReceived, swaps[0].Status)

	err = db.Model(model.TxEventLog{}).Where("tx_hash = ?", txEventLog.TxHash).Updates(
		map[string]interface{}{
			"status":      model.TxStatusConfirmed,
			"update_time": time.Now().Unix(),
		}).Error
	require.NoError(t, err)

	time.Sleep(SleepTime * time.Second)
	time.Sleep(1 * time.Second)

	txEventLogs = make([]model.TxEventLog, 0)
	db.Order("id desc").Find(&txEventLogs)
	require.Equal(t, 2, len(txEventLogs))
	require.Equal(t, model.AckSwapRequest, txEventLogs[0].Phase)

	time.Sleep((SwapSleepSecond + 1) * time.Second)

	swapTxs := make([]model.SwapTx, 0)
	db.Order("id desc").Find(&swapTxs)
	require.Equal(t, 1, len(swapTxs))
	require.Equal(t, model.WithdrawTxSent, swapTxs[0].Status)
	depositTxHash := swapTxs[0].DepositTxHash

	swap := model.Swap{}
	db.Where("deposit_tx_hash = ?", depositTxHash).First(&swap)
	require.Equal(t, SwapSent, swap.Status)

	t.Log("wait to withdraw tx finalization")
	time.Sleep(15 * time.Second)

	swapTx := model.SwapTx{}
	swap = model.Swap{}
	db.Where("deposit_tx_hash = ?", depositTxHash).First(&swap)
	db.Where("deposit_tx_hash = ?", depositTxHash).First(&swapTx)
	require.Equal(t, SwapSuccess, swap.Status)
	require.Equal(t, model.WithdrawTxSuccess, swapTx.Status)
}

func TestSwap_BSC2ETH(t *testing.T) {
	config := getTestConfig()

	bscClient, err := ethclient.Dial(config.ChainConfig.BSCProvider)
	require.NoError(t, err)

	ethClient, err := ethclient.Dial(config.ChainConfig.ETHProvider)
	require.NoError(t, err)

	db, err := prepareDB(config)
	require.NoError(t, err)

	err = insertABCToken(db)
	require.NoError(t, err)

	_, err = prepareTest(db, bscClient, ethClient, config)
	require.NoError(t, err)

	_, clientAccount, err := getClientAccount()
	require.NoError(t, err)

	txEventLog := model.TxEventLog{
		Chain: common.ChainBSC,

		ContractAddress: "0xCCE0532FE1029f1A6B7ccca4C522cF9870a6a8Ed",
		FromAddress:     clientAccount.String(),
		ToAddress:       "",
		Amount:          "1000",
		FeeAmount:       "100000",

		Status:       model.TxStatusInit,
		TxHash:       NewSHA1Hash(),
		BlockHash:    "",
		Height:       0,
		ConfirmedNum: 0,

		Phase: model.SeenSwapRequest,

		UpdateTime: time.Now().Unix(),
		CreateTime: time.Now().Unix(),
	}

	err = insertTxEventLogToDB(db, &txEventLog)
	require.NoError(t, err)

	time.Sleep(SleepTime * time.Second)
	time.Sleep(1 * time.Second)

	txEventLogs := make([]model.TxEventLog, 0)
	db.Order("id desc").Find(&txEventLogs)
	require.Equal(t, 1, len(txEventLogs))
	require.Equal(t, model.ConfirmSwapRequest, txEventLogs[0].Phase)

	swaps := make([]model.Swap, 0)
	db.Order("id desc").Find(&swaps)
	require.Equal(t, 1, len(swaps))
	require.Equal(t, txEventLog.TxHash, swaps[0].DepositTxHash)
	require.Equal(t, SwapQuoteRejected, swaps[0].Status)

	txEventLog = model.TxEventLog{
		Chain: common.ChainBSC,

		ContractAddress: "0xCCE0532FE1029f1A6B7ccca4C522cF9870a6a8Ed",
		FromAddress:     clientAccount.String(),
		ToAddress:       "",
		Amount:          "1000000000000000000",
		FeeAmount:       "100000",

		Status:       model.TxStatusInit,
		TxHash:       NewSHA1Hash(),
		BlockHash:    "",
		Height:       0,
		ConfirmedNum: 0,

		Phase: model.SeenSwapRequest,

		UpdateTime: time.Now().Unix(),
		CreateTime: time.Now().Unix(),
	}
	depositTxHash := txEventLog.TxHash
	err = insertTxEventLogToDB(db, &txEventLog)
	require.NoError(t, err)

	time.Sleep((SleepTime + 1) * time.Second)

	err = db.Model(model.TxEventLog{}).Where("tx_hash = ?", depositTxHash).Updates(
		map[string]interface{}{
			"status":      model.TxStatusConfirmed,
			"update_time": time.Now().Unix(),
		}).Error
	require.NoError(t, err)

	time.Sleep((SleepTime + 1) * time.Second)

	txEventLog = model.TxEventLog{}
	err = db.Where("tx_hash = ?", depositTxHash).First(&txEventLog).Error
	require.NoError(t, err)
	require.Equal(t, model.AckSwapRequest, txEventLog.Phase)

	time.Sleep(SwapSleepSecond * 3 * time.Second)

	swap := model.Swap{}
	err = db.Where("deposit_tx_hash = ?", depositTxHash).First(&swap).Error
	require.NoError(t, err)
	require.Equal(t, SwapSent, swap.Status)

	swapTx := model.SwapTx{}
	err = db.Model(model.SwapTx{}).Where("deposit_tx_hash = ?", depositTxHash).First(&swapTx).Error
	require.NoError(t, err)
	require.Equal(t, model.WithdrawTxSent, swapTx.Status)

	t.Log("wait to withdraw tx finalization")
	time.Sleep(60 * time.Second)

	swapTx = model.SwapTx{}
	swap = model.Swap{}
	db.Where("deposit_tx_hash = ?", depositTxHash).First(&swap)
	db.Where("deposit_tx_hash = ?", depositTxHash).First(&swapTx)
	require.Equal(t, SwapSuccess, swap.Status)
	require.Equal(t, model.WithdrawTxSuccess, swapTx.Status)
}

func TestSwap_UnsupportedToken(t *testing.T) {
	config := getTestConfig()

	bscClient, err := ethclient.Dial(config.ChainConfig.BSCProvider)
	require.NoError(t, err)

	ethClient, err := ethclient.Dial(config.ChainConfig.ETHProvider)
	require.NoError(t, err)

	db, err := prepareDB(config)
	require.NoError(t, err)

	err = insertABCToken(db)
	require.NoError(t, err)

	_, err = prepareTest(db, bscClient, ethClient, config)
	require.NoError(t, err)

	_, clientAccount, err := getClientAccount()
	require.NoError(t, err)

	txEventLog := model.TxEventLog{
		Chain: common.ChainBSC,

		ContractAddress: "0x8f36F4A709409a95a0df90cbc43ED9a658E11E4A", // DEF token
		FromAddress:     clientAccount.String(),
		ToAddress:       "",
		Amount:          "1000000000000000000",
		FeeAmount:       "100000",

		Status:       model.TxStatusInit,
		TxHash:       NewSHA1Hash(),
		BlockHash:    "",
		Height:       0,
		ConfirmedNum: 0,

		Phase: model.SeenSwapRequest,

		UpdateTime: time.Now().Unix(),
		CreateTime: time.Now().Unix(),
	}
	depositTxHash := txEventLog.TxHash
	err = insertTxEventLogToDB(db, &txEventLog)
	require.NoError(t, err)

	time.Sleep((SleepTime + 1) * time.Second)

	err = db.Model(model.TxEventLog{}).Where("tx_hash = ?", depositTxHash).Updates(
		map[string]interface{}{
			"status":      model.TxStatusConfirmed,
			"update_time": time.Now().Unix(),
		}).Error
	require.NoError(t, err)

	time.Sleep((SleepTime + 1) * time.Second)

	txEventLog = model.TxEventLog{}
	err = db.Where("tx_hash = ?", depositTxHash).First(&txEventLog).Error
	require.NoError(t, err)
	require.Equal(t, model.AckSwapRequest, txEventLog.Phase)

	time.Sleep(SwapSleepSecond * 3 * time.Second)

	swap := model.Swap{}
	err = db.Where("deposit_tx_hash = ?", depositTxHash).First(&swap).Error
	require.NoError(t, err)
	require.Equal(t, SwapQuoteRejected, swap.Status)
}
