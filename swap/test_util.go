package swap

//
//import (
//	"crypto/ecdsa"
//	"fmt"
//	"golang.org/x/crypto/sha3"
//	"io/ioutil"
//	"math/big"
//	"math/rand"
//	"strings"
//	"time"
//
//	"github.com/binance-chain/bsc-eth-swap/model"
//	"github.com/binance-chain/bsc-eth-swap/util"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
//	ethcom "github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/ethclient"
//	"github.com/jinzhu/gorm"
//)
//
//var characterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
//
//func prepareTest(db *gorm.db, bscClient, ethClient *ethclient.Client, config *util.config) (*SwapEngine, error) {
//	util.InitLogger(config.LogConfig)
//	swapEngine, err := NewSwapEngine(db, config, bscClient, ethClient)
//	if err != nil {
//		return nil, err
//	}
//	{
//		go swapEngine.monitorSwapRequestDaemon()
//		go swapEngine.confirmSwapRequestDaemon()
//		go swapEngine.createSwapDaemon()
//		go swapEngine.trackSwapTxDaemon()
//	}
//	return swapEngine, nil
//}
//
//func getTestConfig() *util.config {
//	config := util.ParseConfigFromFile("../config/config.json")
//	return config
//}
//
//func prepareDB(config *util.config) (*gorm.db, error) {
//	config.DBConfig.DBPath = "tmp.db"
//	tmpDBFile, err := ioutil.TempFile("", config.DBConfig.DBPath)
//	if err != nil {
//		return nil, err
//	}
//
//	db, err := gorm.Open(config.DBConfig.Dialect, tmpDBFile.Name())
//	if err != nil {
//		return nil, err
//	}
//	model.InitTables(db)
//	return db, nil
//}
//
//func getClientAccount() (*ecdsa.PrivateKey, ethcom.Address, error) {
//	privateKey, publicKey, err := BuildKeys("580c4a31fffda2006462183b6ecc82a67a7fc772568c5b7191e3ed5e4be1bf04")
//	if err != nil {
//		return nil, ethcom.Address{}, err
//	}
//	return privateKey, GetAddress(publicKey), nil
//}
//
//func insertABCToken(db *gorm.db) error {
//	token := model.SwapPair{
//		Symbol:               "ABC",
//		Name:                 "ABC TOKEN",
//		Decimals:             18,
//		BEP20Addr: strings.ToLower("0xCCE0532FE1029f1A6B7ccca4C522cF9870a6a8Ed"),
//		ERC20Addr: strings.ToLower("0x055d208b90DA0E3A431CA7E0fba326888Ef8a822"),
//		LowBound:             "10000",
//		UpperBound:           "1000000000000000000000000",
//		Available:            true,
//	}
//
//	tx := db.Begin()
//	if tx.Error != nil {
//		return tx.Error
//	}
//
//	if err := tx.Create(&token).Error; err != nil {
//		tx.Rollback()
//		return err
//	}
//	return tx.Commit().Error
//}
//
//func insertDEFToken(db *gorm.db) error {
//	token := model.SwapPair{
//		Symbol:               "DEF",
//		Name:                 "DEF TOKEN",
//		Decimals:             18,
//		BEP20Addr: strings.ToLower("0x8f36F4A709409a95a0df90cbc43ED9a658E11E4A"),
//		ERC20Addr: strings.ToLower("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"),
//		LowBound:             "0",
//		UpperBound:           "1000000000000000000000000",
//		Available:            true,
//	}
//
//	tx := db.Begin()
//	if tx.Error != nil {
//		return tx.Error
//	}
//
//	if err := tx.Create(&token).Error; err != nil {
//		tx.Rollback()
//		return err
//	}
//	return tx.Commit().Error
//}
//
//func buildTransactor(privateKey *ecdsa.PrivateKey, value *big.Int) *bind.TransactOpts {
//	txOpt := bind.NewKeyedTransactor(privateKey)
//	txOpt.Value = value
//	return txOpt
//}
//
//func swapBSC2ETH(client *ethclient.Client, privateKey *ecdsa.PrivateKey, bscSwapProxyContractAddr, tokenAddr ethcom.Address, amount *big.Int) (ethcom.Hash, error) {
//	bscERC20Instance, _ := erc20.NewErc20(tokenAddr, client)
//	bscSwapProxy, _ := swapagent.NewSwap(bscSwapProxyContractAddr, client)
//
//	approveTx, err := bscERC20Instance.Approve(buildTransactor(privateKey, nil), bscSwapProxyContractAddr, amount)
//	if err != nil {
//		return ethcom.Hash{}, err
//	}
//	util.Logger.Infof("approveTx Hash: https://testnet.bscscan.com/tx/%s", approveTx.Hash().String())
//	time.Sleep(500 * time.Millisecond)
//
//	swapTx, err := bscSwapProxy.Swap(buildTransactor(privateKey, big.NewInt(100000)), tokenAddr, amount)
//	if err != nil {
//		return ethcom.Hash{}, err
//	}
//	util.Logger.Infof("https://testnet.bscscan.com/tx/%s", swapTx.Hash().String())
//	return swapTx.Hash(), nil
//}
//
//func swapETH2BSC(client *ethclient.Client, privateKey *ecdsa.PrivateKey, ethSwapProxyContractAddr, tokenAddr ethcom.Address, amount *big.Int) (ethcom.Hash, error) {
//	ethERC20Instance, _ := erc20.NewErc20(tokenAddr, client)
//	ethSwapProxy, _ := swapagent.NewSwap(ethSwapProxyContractAddr, client)
//	approveTx, err := ethERC20Instance.Approve(buildTransactor(privateKey, nil), ethSwapProxyContractAddr, amount)
//	if err != nil {
//		return ethcom.Hash{}, err
//	}
//	util.Logger.Infof("approveTx Hash: https://rinkeby.etherscan.io/tx/%s", approveTx.Hash().String())
//	time.Sleep(500 * time.Millisecond)
//
//	swapTx, err := ethSwapProxy.Swap(buildTransactor(privateKey, big.NewInt(100000)), tokenAddr, amount)
//	if err != nil {
//		return ethcom.Hash{}, err
//	}
//	util.Logger.Infof("https://rinkeby.etherscan.io/tx/%s", swapTx.Hash().String())
//	return swapTx.Hash(), nil
//}
//
//func insertTxEventLogToDB(db *gorm.db, data *model.SwapStartTxLog) error {
//	tx := db.Begin()
//	if err := tx.Error; err != nil {
//		return err
//	}
//
//	if err := tx.Create(data).Error; err != nil {
//		tx.Rollback()
//		return err
//	}
//
//	return tx.Commit().Error
//}
//
//func NewSHA1Hash(n ...int) string {
//	noRandomCharacters := 64
//
//	if len(n) > 0 {
//		noRandomCharacters = n[0]
//	}
//
//	randString := RandomString(noRandomCharacters)
//
//	hash := sha3.NewLegacyKeccak256()
//	hash.Write([]byte(randString))
//	bs := hash.Sum(nil)
//
//	return fmt.Sprintf("0x%x", bs)
//}
//
//// RandomString generates a random string of n length
//func RandomString(n int) string {
//	b := make([]rune, n)
//	for i := range b {
//		b[i] = characterRunes[rand.Intn(len(characterRunes))]
//	}
//	return string(b)
//}
