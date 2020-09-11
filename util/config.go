package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/binance-chain/bsc-eth-swap/common"
)

type Config struct {
	DBConfig    *DBConfig    `json:"db_config"`
	ChainConfig *ChainConfig `json:"chain_config"`
	LogConfig   *LogConfig   `json:"log_config"`
	AlertConfig *AlertConfig `json:"alert_config"`
	AdminConfig *AdminConfig `json:"admin_config"`
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

func (cfg *AlertConfig) Validate() {
	if cfg.BlockUpdateTimeout <= 0 {
		panic(fmt.Sprintf("block_update_timeout should be larger than 0"))
	}
}

type DBConfig struct {
	Dialect string `json:"dialect"`
	DBPath  string `json:"db_path"`
}

func (cfg *DBConfig) Validate() {
	if cfg.Dialect != common.DBDialectMysql && cfg.Dialect != common.DBDialectSqlite3 {
		panic(fmt.Sprintf("only %s and %s supported", common.DBDialectMysql, common.DBDialectSqlite3))
	}
	if cfg.DBPath == "" {
		panic("db path should not be empty")
	}
}

type ChainConfig struct {
	BSCStartHeight int64  `json:"bsc_start_height"`
	BSCProvider    string `json:"bsc_provider"`
	BSCConfirmNum  int64  `json:"bsc_confirm_num"`
	BSCChainId     string `json:"bsc_chain_id"`

	BBCStartHeight          int64  `json:"bbc_start_height"`
	BBCRpcAddr              string `json:"bbc_rpc_addr"`
	BBCBreatheBlockInterval int64  `json:"bbc_breathe_block_interval"`
	GameStartHeightOnBC     int64  `json:"game_start_height_on_bc"`
	GameEndHeightOnBC       int64  `json:"game_end_height_on_bc"`

	AccumulatedValidatorRewardScoreLimit int64    `json:"accumulated_validator_reward_score_limit"`
	BinanceValidators                    []string `json:"binance_validators"`
}

func (cfg *ChainConfig) Validate() {
	if cfg.BSCStartHeight < 0 {
		panic("bsc_start_height should not be less than 0")
	}
	if cfg.BSCProvider == "" {
		panic("bsc_provider should not be empty")
	}
	if cfg.BSCConfirmNum <= 0 {
		panic("bsc_confirm_num should be larger than 0")
	}

	if cfg.BBCRpcAddr == "" {
		panic("bbc_rpc_addr should not be empty")
	}
	if cfg.BBCRpcAddr == "" {
		panic("bbc_rpc_addr should not be empty")
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

func (cfg *LogConfig) Validate() {
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
