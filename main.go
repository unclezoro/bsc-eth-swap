package main

import (
	"flag"
	"fmt"

	"github.com/binance-chain/bsc-eth-swap/admin"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/binance-chain/bsc-eth-swap/executor"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/observer"
	"github.com/binance-chain/bsc-eth-swap/swap"
	"github.com/binance-chain/bsc-eth-swap/util"
)

const (
	flagConfigType         = "config-type"
	flagConfigAwsRegion    = "aws-region"
	flagConfigAwsSecretKey = "aws-secret-key"
	flagConfigPath         = "config-path"
)

const (
	ConfigTypeLocal = "local"
	ConfigTypeAws   = "aws"
)

func initFlags() {
	flag.String(flagConfigPath, "", "config path")
	flag.String(flagConfigType, "", "config type, local or aws")
	flag.String(flagConfigAwsRegion, "", "aws s3 region")
	flag.String(flagConfigAwsSecretKey, "", "aws s3 secret key")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(fmt.Sprintf("bind flags error, err=%s", err))
	}
}

func printUsage() {
	fmt.Print("usage: ./swap --config-type [local or aws] --config-path config_file_path\n")
}

func main() {
	initFlags()

	configType := viper.GetString(flagConfigType)
	if configType == "" {
		printUsage()
		return
	}

	if configType != ConfigTypeAws && configType != ConfigTypeLocal {
		printUsage()
		return
	}

	var config *util.Config
	if configType == ConfigTypeAws {
		awsSecretKey := viper.GetString(flagConfigAwsSecretKey)
		if awsSecretKey == "" {
			printUsage()
			return
		}

		awsRegion := viper.GetString(flagConfigAwsRegion)
		if awsRegion == "" {
			printUsage()
			return
		}

		configContent, err := util.GetSecret(awsSecretKey, awsRegion)
		if err != nil {
			fmt.Printf("get aws config error, err=%s", err.Error())
			return
		}
		config = util.ParseConfigFromJson(configContent)
	} else {
		configFilePath := viper.GetString(flagConfigPath)
		if configFilePath == "" {
			printUsage()
			return
		}
		config = util.ParseConfigFromFile(configFilePath)
	}
	config.Validate()

	// init logger
	util.InitLogger(config.LogConfig)
	util.InitTgAlerter(config.AlertConfig)

	db, err := gorm.Open(config.DBConfig.Dialect, config.DBConfig.DBPath)
	if err != nil {
		panic(fmt.Sprintf("open db error, err=%s", err.Error()))
	}
	defer db.Close()
	model.InitTables(db)

	bscClient, err := ethclient.Dial(config.ChainConfig.BSCProvider)
	if err != nil {
		panic("new eth client error")
	}

	ethClient, err := ethclient.Dial(config.ChainConfig.ETHProvider)
	if err != nil {
		panic("new eth client error")
	}

	bscExecutor := executor.NewBSCExecutor(bscClient, config.ChainConfig.BSCSwapAgentAddr, config)
	bscObserver := observer.NewObserver(db, config.ChainConfig.BSCStartHeight, config.ChainConfig.BSCConfirmNum, config, bscExecutor)
	bscObserver.Start()

	ethExecutor := executor.NewEthExecutor(ethClient, config.ChainConfig.ETHSwapAgentAddr, config)
	ethObserver := observer.NewObserver(db, config.ChainConfig.ETHStartHeight, config.ChainConfig.ETHConfirmNum, config, ethExecutor)
	ethObserver.Start()

	swapEngine, err := swap.NewSwapEngine(db, config, bscClient, ethClient)
	if err != nil {
		panic(fmt.Sprintf("create swap engine error, err=%s", err.Error()))
	}

	swapEngine.Start()

	swapPairEngine, err := swap.NewSwapPairEngine(db, config, bscClient, swapEngine)
	if err != nil {
		panic(fmt.Sprintf("create swap pair engine error, err=%s", err.Error()))
	}
	swapPairEngine.Start()

	signer, err := util.NewHmacSignerFromConfig(config)
	if err != nil {
		panic(fmt.Sprintf("new hmac singer error, err=%s", err.Error()))
	}
	admin := admin.NewAdmin(config, db, signer, swapEngine)
	go admin.Serve()

	select {}
}
