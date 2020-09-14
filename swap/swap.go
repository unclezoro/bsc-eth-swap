package swap

import (
	"crypto/ecdsa"
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/executor"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type Swapper struct {
	DB            *gorm.DB
	Config        *util.Config
	BSCPrivateKey *ecdsa.PrivateKey
	ETHPrivateKey *ecdsa.PrivateKey
}

func getBSCPrivateKey(cfg *util.Config) (*ecdsa.PrivateKey, error) {
	var bscPrivateKey string
	if cfg.SecretKeyConfig.BSCKeyType == common.AWSPrivateKey {
		result, err := util.GetSecret(cfg.SecretKeyConfig.BSCKeyAWSSecretName, cfg.SecretKeyConfig.BSCKeyAWSRegion)
		if err != nil {
			return nil, err
		}
		type AwsPrivateKey struct {
			PrivateKey string `json:"private_key"`
		}
		var awsPrivateKey AwsPrivateKey
		err = json.Unmarshal([]byte(result), &awsPrivateKey)
		if err != nil {
			return nil, err
		}
		bscPrivateKey = awsPrivateKey.PrivateKey
	} else {
		bscPrivateKey = cfg.SecretKeyConfig.BSCPrivateKey
	}

	return crypto.HexToECDSA(bscPrivateKey)
}

func getETHPrivateKey(cfg *util.Config) (*ecdsa.PrivateKey, error) {
	var ethPrivateKey string
	if cfg.SecretKeyConfig.ETHKeyType == common.AWSPrivateKey {
		result, err := util.GetSecret(cfg.SecretKeyConfig.ETHKeyAWSSecretName, cfg.SecretKeyConfig.ETHKeyAWSRegion)
		if err != nil {
			return nil, err
		}
		type AwsPrivateKey struct {
			PrivateKey string `json:"private_key"`
		}
		var awsPrivateKey AwsPrivateKey
		err = json.Unmarshal([]byte(result), &awsPrivateKey)
		if err != nil {
			return nil, err
		}
		ethPrivateKey = awsPrivateKey.PrivateKey
	} else {
		ethPrivateKey = cfg.SecretKeyConfig.ETHPrivateKey
	}

	return crypto.HexToECDSA(ethPrivateKey)
}

// NewSwapper returns the Swapper instance
func NewSwapper(db *gorm.DB, cfg *util.Config, bscExecutor executor.Executor) (*Swapper, error) {
	bscPriKey, err := getBSCPrivateKey(cfg)
	if err != nil {
		return nil, err
	}

	ethPriKey, err := getETHPrivateKey(cfg)
	if err != nil {
		return nil, err
	}

	return &Swapper{
		DB:            db,
		Config:        cfg,
		BSCPrivateKey: bscPriKey,
		ETHPrivateKey: ethPriKey,
	}, nil
}

func HandleSwapDaemon() {

}

func TrackSwapDaemon() {

}

func AlertDaemon() {

}
