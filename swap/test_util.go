package swap

import (
	"io/ioutil"
	"strings"

	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
	"github.com/jinzhu/gorm"
)

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

func insertABCToken(db *gorm.DB) error {
	token := model.Token{
		Symbol:               "ABC",
		Name:                 "ABC TOKEN",
		Decimals:             18,
		BSCTokenContractAddr: strings.ToLower("0xCCE0532FE1029f1A6B7ccca4C522cF9870a6a8Ed"),
		ETHTokenContractAddr: strings.ToLower("0x055d208b90DA0E3A431CA7E0fba326888Ef8a822"),
		LowBound:             "0",
		UpperBound:          "1000000000000000000000000",
		BSCSenderAddr:       "0x0C5006c9322b6dC49BC475d7635659F7147326d3",
		BSCERC20Threshold:   "1000000000000000000",
		ETHSenderAddr:       "0x0C5006c9322b6dC49BC475d7635659F7147326d3",
		ETHERC20Threshold:   "1000000000000000000",
		Available:           true,
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
		UpperBound:          "1000000000000000000000000",
		BSCSenderAddr:       "0xb0438919ABefa48a43635279d822735d31b5d762",
		BSCERC20Threshold:   "1000000000000000000",
		ETHSenderAddr:       "0xb0438919ABefa48a43635279d822735d31b5d762",
		ETHERC20Threshold:   "1000000000000000000",
		Available:           true,
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
