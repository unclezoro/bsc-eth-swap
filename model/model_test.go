package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestInsertTokenConfig(t *testing.T)  {
	db, err := gorm.Open("sqlite3", "/Users/liuhaoyang/workspace/bsc-eth-swap/bsc-eth-swap/build/test.db")
	if err != nil {
		panic(fmt.Sprintf("open db error, err=%s", err.Error()))
	}
	defer db.Close()
	InitTables(db)

	token := Token{
		Symbol:              "FT",
		Name:                "FT TOKEN",
		BSCContractAddr:     "0x6e491b5569a30935bc961377957212e27cD85Ba5",
		ETHContractAddr:     "0x8A1a84726AbE38764D34c848021F8860691FdDB3",
		LowBound:            "0",
		UpperBound:          "1000000000000000000000000",
		BSCKeyType:          "local",
		BSCKeyAWSRegion:     "",
		BSCKeyAWSSecretName: "",
		BSCPrivateKey:       "26ca57a5b8e622c87b1f5816b54bed6b8f49357531929c4e29f1cd381c210678",
		ETHKeyType:          "local",
		ETHKeyAWSRegion:     "",
		ETHKeyAWSSecretName: "",
		ETHPrivateKey:       "26ca57a5b8e622c87b1f5816b54bed6b8f49357531929c4e29f1cd381c210678",
		UpdateTime:          time.Now().Unix(),
		CreateTime:          time.Now().Unix(),
	}

	tx := db.Begin()
	require.NoError(t, tx.Error)

	if err := tx.Create(&token).Error; err != nil {
		tx.Rollback()
		require.NoError(t, tx.Error)
	}
	require.NoError(t, tx.Commit().Error)
}
