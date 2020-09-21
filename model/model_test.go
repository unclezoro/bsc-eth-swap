package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/require"
)

func TestInsertTokenConfig(t *testing.T) {
	db, err := gorm.Open("mysql", "root:ethswap123123@(localhost:3306)/ethswap?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(fmt.Sprintf("open db error, err=%s", err.Error()))
	}
	defer db.Close()
	InitTables(db)

	token := Token{
		Symbol:               "YT",
		Name:                 "YT TOKEN",
		Decimals:             18,
		BSCTokenContractAddr: strings.ToLower("0x6e491b5569a30935bc961377957212e27cD85Ba5"),
		ETHTokenContractAddr: strings.ToLower("0x8A1a84726AbE38764D34c848021F8860691FdDB3"),
		LowBound:             "0",
		UpperBound:          "1000000000000000000000000",
		BSCKeyType:          "local",
		BSCKeyAWSRegion:     "",
		BSCKeyAWSSecretName: "",
		BSCPrivateKey:       "26ca57a5b8e622c87b1f5816b54bed6b8f49357531929c4e29f1cd381c210678",
		BSCSenderAddr:       "0x277f8E924CfffF8FAbCae123B8e78dDa9e406384",
		BSCERC20Threshold:   "1000000000000000000",
		ETHKeyType:          "local",
		ETHKeyAWSRegion:     "",
		ETHKeyAWSSecretName: "",
		ETHPrivateKey:       "26ca57a5b8e622c87b1f5816b54bed6b8f49357531929c4e29f1cd381c210678",
		ETHSenderAddr:       "0x277f8E924CfffF8FAbCae123B8e78dDa9e406384",
		ETHERC20Threshold:   "1000000000000000000",
		Available:           true,
	}

	tx := db.Begin()
	require.NoError(t, tx.Error)

	if err := tx.Create(&token).Error; err != nil {
		tx.Rollback()
		require.NoError(t, tx.Error)
	}
	require.NoError(t, tx.Commit().Error)
}
