package swap

import (
	"io/ioutil"

	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
	"github.com/jinzhu/gorm"
)

func GetTestConfig() *util.Config {
	config := util.ParseConfigFromFile("../config/config.json")
	return config
}

func PrepareDB(config *util.Config) (*gorm.DB, error) {
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
