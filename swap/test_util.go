package swap

import (
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/swap/erc20"
	"github.com/binance-chain/bsc-eth-swap/swap/swapproxy"
	"github.com/binance-chain/bsc-eth-swap/util"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
)

func prepareTest(db *gorm.DB, bscClient, ethClient *ethclient.Client, config *util.Config) (*Swapper, error) {
	swapInstance, err := NewSwapper(db, config, bscClient, ethClient)
	if err != nil {
		return nil, err
	}
	{
		go swapInstance.monitorSwapRequestDaemon()
		go swapInstance.confirmSwapRequestDaemon()
		go swapInstance.createSwapDaemon()
		go swapInstance.trackSwapTxDaemon()
	}
	return swapInstance, nil
}

func getTestConfig() *util.Config {
	config := util.ParseConfigFromFile("../config/config.json")
	return config
}

func prepareDB(config *util.Config) (*gorm.DB, error) {
	config.DBConfig.DBPath = "tmp.db"
	tmpDBFile, err := ioutil.TempFile("", config.DBConfig.DBPath)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(config.DBConfig.Dialect, tmpDBFile.Name())
	if err != nil {
		return nil, err
	}
	model.InitTables(db)
	return db, nil
}

func getClientAccount() (*ecdsa.PrivateKey, ethcom.Address, error) {
	privateKey, publicKey, err := BuildKeys("580c4a31fffda2006462183b6ecc82a67a7fc772568c5b7191e3ed5e4be1bf04")
	if err != nil {
		return nil, ethcom.Address{}, err
	}
	return privateKey, GetAddress(publicKey), nil
}

func insertABCToken(db *gorm.DB) error {
	token := model.Token{
		Symbol:               "ABC",
		Name:                 "ABC TOKEN",
		Decimals:             18,
		BSCTokenContractAddr: strings.ToLower("0xCCE0532FE1029f1A6B7ccca4C522cF9870a6a8Ed"),
		ETHTokenContractAddr: strings.ToLower("0x055d208b90DA0E3A431CA7E0fba326888Ef8a822"),
		LowBound:             "0",
		UpperBound:           "1000000000000000000000000",
		BSCSenderAddr:        "0x0C5006c9322b6dC49BC475d7635659F7147326d3",
		BSCERC20Threshold:    "1000000000000000000",
		ETHSenderAddr:        "0x0C5006c9322b6dC49BC475d7635659F7147326d3",
		ETHERC20Threshold:    "1000000000000000000",
		Available:            true,
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Create(&token).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func insertDEFToken(db *gorm.DB) error {
	token := model.Token{
		Symbol:               "DEF",
		Name:                 "DEF TOKEN",
		Decimals:             18,
		BSCTokenContractAddr: strings.ToLower("0x8f36F4A709409a95a0df90cbc43ED9a658E11E4A"),
		ETHTokenContractAddr: strings.ToLower("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"),
		LowBound:             "0",
		UpperBound:           "1000000000000000000000000",
		BSCSenderAddr:        "0xb0438919ABefa48a43635279d822735d31b5d762",
		BSCERC20Threshold:    "1000000000000000000",
		ETHSenderAddr:        "0xb0438919ABefa48a43635279d822735d31b5d762",
		ETHERC20Threshold:    "1000000000000000000",
		Available:            true,
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Create(&token).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func buildTransactor(privateKey *ecdsa.PrivateKey, value *big.Int) *bind.TransactOpts {
	txOpt := bind.NewKeyedTransactor(privateKey)
	txOpt.Value = value
	return txOpt
}

func swapBSC2ETH(client *ethclient.Client, privateKey *ecdsa.PrivateKey, bscSwapProxyContractAddr, tokenAddr ethcom.Address, amount *big.Int) (ethcom.Hash, error) {
	bscERC20Instance, _ := erc20.NewErc20(tokenAddr, client)
	bscSwapProxy, _ := swapproxy.NewSwap(bscSwapProxyContractAddr, client)

	approveTx, err := bscERC20Instance.Approve(buildTransactor(privateKey, nil), bscSwapProxyContractAddr, amount)
	if err != nil {
		return ethcom.Hash{}, err
	}
	util.Logger.Infof("approveTx Hash: https://testnet.bscscan.com/tx/%s", approveTx.Hash().String())
	time.Sleep(500 * time.Millisecond)

	swapTx, err := bscSwapProxy.Swap(buildTransactor(privateKey, big.NewInt(100000)), tokenAddr, amount)
	if err != nil {
		return ethcom.Hash{}, err
	}
	util.Logger.Infof("https://testnet.bscscan.com/tx/%s", swapTx.Hash().String())
	return swapTx.Hash(), nil
}

func swapETH2BSC(client *ethclient.Client, privateKey *ecdsa.PrivateKey, ethSwapProxyContractAddr, tokenAddr ethcom.Address, amount *big.Int) (ethcom.Hash, error) {
	ethERC20Instance, _ := erc20.NewErc20(tokenAddr, client)
	ethSwapProxy, _ := swapproxy.NewSwap(ethSwapProxyContractAddr, client)
	approveTx, err := ethERC20Instance.Approve(buildTransactor(privateKey, nil), ethSwapProxyContractAddr, amount)
	if err != nil {
		return ethcom.Hash{}, err
	}
	util.Logger.Infof("approveTx Hash: https://rinkeby.etherscan.io/tx/%s", approveTx.Hash().String())
	time.Sleep(500 * time.Millisecond)

	swapTx, err := ethSwapProxy.Swap(buildTransactor(privateKey, big.NewInt(100000)), tokenAddr, amount)
	if err != nil {
		return ethcom.Hash{}, err
	}
	util.Logger.Infof("https://rinkeby.etherscan.io/tx/%s", swapTx.Hash().String())
	return swapTx.Hash(), nil
}
