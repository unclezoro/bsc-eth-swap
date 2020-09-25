package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	ethcom "github.com/ethereum/go-ethereum/common"

	"github.com/binance-chain/bsc-eth-swap/common"
)

type Config struct {
	KeyManagerConfig KeyManagerConfig `json:"key_manager_config"`
	DBConfig         DBConfig         `json:"db_config"`
	ChainConfig      ChainConfig      `json:"chain_config"`
	LogConfig        LogConfig        `json:"log_config"`
	AlertConfig      AlertConfig      `json:"alert_config"`
	AdminConfig      AdminConfig      `json:"admin_config"`
}

func (cfg *Config) Validate() {
	cfg.DBConfig.Validate()
	cfg.ChainConfig.Validate()
	cfg.LogConfig.Validate()
	cfg.AlertConfig.Validate()
}

type AlertConfig struct {
	TelegramBotId  string `json:"telegram_bot_id"`
	TelegramChatId string `json:"telegram_chat_id"`

	BlockUpdateTimeout int64 `json:"block_update_timeout"`
}

func (cfg AlertConfig) Validate() {
	if cfg.BlockUpdateTimeout <= 0 {
		panic(fmt.Sprintf("block_update_timeout should be larger than 0"))
	}
}

type KeyManagerConfig struct {
	KeyType       string `json:"key_type"`
	AWSRegion     string `json:"aws_region"`
	AWSSecretName string `json:"aws_secret_name"`

	// local keys
	LocalKeys           []TokenSecretKey `json:"local_keys"`
	LocalHMACKey        string           `json:"local_hmac_key"`
	LocalAdminApiKey    string           `json:"local_admin_api_key"`
	LocalAdminSecretKey string           `json:"local_admin_secret_key"`
}

type KeyConfig struct {
	TokenKeys      []TokenSecretKey `json:"token_keys"`
	HMACKey        string           `json:"hmac_key"`
	AdminApiKey    string           `json:"admin_api_key"`
	AdminSecretKey string           `json:"admin_secret_key"`
}

func (cfg KeyManagerConfig) Validate() {
	if cfg.KeyType == common.LocalPrivateKey && len(cfg.LocalKeys) == 0 {
		panic("missing local private key")
	}
	if cfg.KeyType == common.LocalPrivateKey && len(cfg.LocalHMACKey) == 0 {
		panic("missing local hmac key")
	}
	if cfg.KeyType == common.LocalPrivateKey && len(cfg.LocalAdminApiKey) == 0 {
		panic("missing local admin api key")
	}
	if cfg.KeyType == common.LocalPrivateKey && len(cfg.LocalAdminSecretKey) == 0 {
		panic("missing local admin secret key")
	}

	if cfg.KeyType == common.AWSPrivateKey && (cfg.AWSRegion == "" || cfg.AWSSecretName == "") {
		panic("Missing aws key region or name")
	}
}

type TokenSecretKey struct {
	Symbol        string `json:"symbol"`
	BSCPrivateKey string `json:"bsc_private_key"`
	ETHPrivateKey string `json:"eth_private_key"`
}

type DBConfig struct {
	Dialect string `json:"dialect"`
	DBPath  string `json:"db_path"`
}

func (cfg DBConfig) Validate() {
	if cfg.Dialect != common.DBDialectMysql && cfg.Dialect != common.DBDialectSqlite3 {
		panic(fmt.Sprintf("only %s and %s supported", common.DBDialectMysql, common.DBDialectSqlite3))
	}
	if cfg.DBPath == "" {
		panic("db path should not be empty")
	}
}

type ChainConfig struct {
	BalanceMonitorInterval int64 `json:"balance_monitor_interval"`

	BSCObserverFetchInterval    int64  `json:"bsc_observer_fetch_interval"`
	BSCStartHeight              int64  `json:"bsc_start_height"`
	BSCProvider                 string `json:"bsc_provider"`
	BSCConfirmNum               int64  `json:"bsc_confirm_num"`
	BSCSwapContractAddr         string `json:"bsc_swap_contract_addr"`
	BSCExplorerUrl              string `json:"bsc_explorer_url"`
	BSCMaxTrackRetry            int64  `json:"bsc_max_track_retry"`
	BNBAlertThreshold           string `json:"bnb_alert_threshold"`
	BSCWaitMilliSecBetweenSwaps int64  `json:"bsc_wait_milli_sec_between_swaps"`

	ETHObserverFetchInterval    int64  `json:"eth_observer_fetch_interval"`
	ETHStartHeight              int64  `json:"eth_start_height"`
	ETHProvider                 string `json:"eth_provider"`
	ETHConfirmNum               int64  `json:"eth_confirm_num"`
	ETHSwapContractAddr         string `json:"eth_swap_contract_addr"`
	ETHExplorerUrl              string `json:"eth_explorer_url"`
	ETHMaxTrackRetry            int64  `json:"eth_max_track_retry"`
	ETHAlertThreshold           string `json:"eth_alert_threshold"`
	ETHWaitMilliSecBetweenSwaps int64  `json:"eth_wait_milli_sec_between_swaps"`
}

func (cfg ChainConfig) Validate() {
	if cfg.BSCStartHeight < 0 {
		panic("bsc_start_height should not be less than 0")
	}
	if cfg.BSCProvider == "" {
		panic("bsc_provider should not be empty")
	}
	if cfg.BSCConfirmNum <= 0 {
		panic("bsc_confirm_num should be larger than 0")
	}
	if !ethcom.IsHexAddress(cfg.BSCSwapContractAddr) {
		panic(fmt.Sprintf("invalid bsc_swap_contract_addr: %s", cfg.BSCSwapContractAddr))
	}
	if cfg.BSCMaxTrackRetry <= 0 {
		panic("bsc_max_track_retry should be larger than 0")
	}

	if cfg.ETHStartHeight < 0 {
		panic("bsc_start_height should not be less than 0")
	}
	if cfg.ETHProvider == "" {
		panic("bsc_provider should not be empty")
	}
	if !ethcom.IsHexAddress(cfg.ETHSwapContractAddr) {
		panic(fmt.Sprintf("invalid eth_swap_contract_addr: %s", cfg.ETHSwapContractAddr))
	}
	if cfg.ETHConfirmNum <= 0 {
		panic("bsc_confirm_num should be larger than 0")
	}
	if cfg.ETHMaxTrackRetry <= 0 {
		panic("eth_max_track_retry should be larger than 0")
	}
}

type LogConfig struct {
	Level                        string `json:"level"`
	Filename                     string `json:"filename"`
	MaxFileSizeInMB              int    `json:"max_file_size_in_mb"`
	MaxBackupsOfLogFiles         int    `json:"max_backups_of_log_files"`
	MaxAgeToRetainLogFilesInDays int    `json:"max_age_to_retain_log_files_in_days"`
	UseConsoleLogger             bool   `json:"use_console_logger"`
	UseFileLogger                bool   `json:"use_file_logger"`
	Compress                     bool   `json:"compress"`
}

func (cfg LogConfig) Validate() {
	if cfg.UseFileLogger {
		if cfg.Filename == "" {
			panic("filename should not be empty if use file logger")
		}
		if cfg.MaxFileSizeInMB <= 0 {
			panic("max_file_size_in_mb should be larger than 0 if use file logger")
		}
		if cfg.MaxBackupsOfLogFiles <= 0 {
			panic("max_backups_off_log_files should be larger than 0 if use file logger")
		}
	}
}

type AdminConfig struct {
	ListenAddr string `json:"listen_addr"`
}

func ParseConfigFromFile(filePath string) *Config {
	bz, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var config Config
	if err := json.Unmarshal(bz, &config); err != nil {
		panic(err)
	}
	return &config
}

func ParseConfigFromJson(content string) *Config {
	var config Config
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		panic(err)
	}
	return &config
}
