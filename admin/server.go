package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	common2 "github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/executor"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

const (
	DefaultListenAddr = "0.0.0.0:8080"
)

type Admin struct {
	DB *gorm.DB

	Config *util.Config

	BSCExecutor executor.Executor
	ETHExecutor executor.Executor
}

func NewAdmin(config *util.Config, db *gorm.DB, bscExecutor executor.Executor, ethExecutor executor.Executor) *Admin {
	return &Admin{
		DB:          db,
		Config:      config,
		BSCExecutor: bscExecutor,
		ETHExecutor: ethExecutor,
	}
}

type NewTokenRequest struct {
	Symbol          string `json:"symbol"`
	Name            string `json:"name"`
	Decimals        int    `json:"decimals"`
	BSCContractAddr string `json:"bsc_contract_addr"`
	ETHContractAddr string `json:"eth_contract_addr"`
	LowerBound      string `json:"lower_bound"`
	UpperBound      string `json:"upper_bound"`

	BSCKeyType          string `json:"bsc_key_type"`
	BSCKeyAWSRegion     string `json:"bsc_key_aws_region"`
	BSCKeyAWSSecretName string `json:"bsc_key_aws_secret_name"`
	BSCPrivateKey       string `json:"bsc_private_key"`
	BSCSendAddr         string `json:"bsc_sender"`

	ETHKeyType          string `json:"eth_key_type"`
	ETHKeyAWSRegion     string `json:"eth_aws_region"`
	ETHKeyAWSSecretName string `json:"eth_key_aws_secret_name"`
	ETHPrivateKey       string `json:"eth_private_key"`
	ETHSendAddr         string `json:"eth_send_addr"`
}

func (admin *Admin) AddToken(w http.ResponseWriter, r *http.Request) {
	var newToken NewTokenRequest

	err := json.NewDecoder(r.Body).Decode(&newToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = tokenBasicCheck(&newToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check decimals
	bscDecimals, err := admin.BSCExecutor.GetContractDecimals(common.HexToAddress(newToken.BSCContractAddr))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ethDecimals, err := admin.BSCExecutor.GetContractDecimals(common.HexToAddress(newToken.ETHContractAddr))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if bscDecimals != ethDecimals || bscDecimals != newToken.Decimals {
		http.Error(w, fmt.Sprintf("decimals is wrong, bsc_decimals=%d, eth_decimals=%d", bscDecimals, ethDecimals), http.StatusInternalServerError)
		return
	}

	tokenModel := model.Token{
		Symbol:              newToken.Symbol,
		Name:                newToken.Name,
		Decimals:            newToken.Decimals,
		BSCContractAddr:     strings.ToLower(newToken.BSCContractAddr),
		ETHContractAddr:     strings.ToLower(newToken.ETHContractAddr),
		LowBound:            newToken.LowerBound,
		UpperBound:          newToken.UpperBound,
		BSCKeyType:          newToken.BSCKeyType,
		BSCKeyAWSRegion:     newToken.BSCKeyAWSRegion,
		BSCKeyAWSSecretName: newToken.BSCKeyAWSSecretName,
		BSCPrivateKey:       newToken.BSCPrivateKey,
		BSCSendAddr:         strings.ToLower(newToken.BSCSendAddr),
		ETHKeyType:          newToken.ETHKeyType,
		ETHKeyAWSRegion:     newToken.ETHKeyAWSRegion,
		ETHKeyAWSSecretName: newToken.ETHKeyAWSSecretName,
		ETHPrivateKey:       newToken.ETHPrivateKey,
		ETHSendAddr:         strings.ToLower(newToken.ETHSendAddr),
	}

	err = admin.DB.Create(&tokenModel).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func tokenBasicCheck(token *NewTokenRequest) error {
	if token.Symbol == "" {
		return fmt.Errorf("symbol should not be empty")
	}
	if token.Name == "" {
		return fmt.Errorf("name should not be empty")
	}
	if token.Decimals <= 0 {
		return fmt.Errorf("decimals should be larger than 0")
	}
	if token.BSCContractAddr == "" {
		return fmt.Errorf("bsc_contract_addr should not be empty")
	}
	if token.ETHContractAddr == "" {
		return fmt.Errorf("eth_contract_addr should not be empty")
	}
	if token.LowerBound == "" {
		return fmt.Errorf("lower_bound should not be empty")
	}
	if token.UpperBound == "" {
		return fmt.Errorf("upper_bound should not be empty")
	}

	// check addresses
	if !common.IsHexAddress(token.BSCContractAddr) {
		return fmt.Errorf("bsc_contract_addr is wrong")
	}
	if !common.IsHexAddress(token.ETHContractAddr) {
		return fmt.Errorf("eth_contract_addr is wrong")
	}
	if !common.IsHexAddress(token.ETHSendAddr) {
		return fmt.Errorf("eth_sender_addr is wrong")
	}

	// check bsc key
	if token.BSCKeyType != common2.LocalPrivateKey && token.BSCKeyType != common2.AWSPrivateKey {
		return fmt.Errorf("bsc_key_type should be %s or %s", common2.LocalPrivateKey, common2.AWSPrivateKey)
	}
	if token.BSCKeyType == common2.AWSPrivateKey {
		if token.BSCKeyAWSRegion == "" {
			return fmt.Errorf("bsc_key_aws_region should not be empty")
		}
		if token.BSCKeyAWSSecretName == "" {
			return fmt.Errorf("bsc_key_aws_secret_name should not be empty")
		}
	} else {
		if token.BSCPrivateKey == "" {
			return fmt.Errorf("bsc_private_key should not be empty")
		}
	}

	// check eth key
	if token.ETHKeyType != common2.LocalPrivateKey && token.ETHKeyType != common2.AWSPrivateKey {
		return fmt.Errorf("eth_key_type should be %s or %s", common2.LocalPrivateKey, common2.AWSPrivateKey)
	}
	if token.ETHKeyType == common2.AWSPrivateKey {
		if token.ETHKeyAWSRegion == "" {
			return fmt.Errorf("eth_key_aws_region should not be empty")
		}
		if token.ETHKeyAWSSecretName == "" {
			return fmt.Errorf("eth_key_aws_secret_name should not be empty")
		}
	} else {
		if token.ETHPrivateKey == "" {
			return fmt.Errorf("eth_private_key should not be empty")
		}
	}

	return nil
}

func (admin *Admin) Endpoints(w http.ResponseWriter, r *http.Request) {
	endpoints := struct {
		Endpoints []string `json:"endpoints"`
	}{
		Endpoints: []string{
			"/add_token",
		},
	}

	jsonBytes, err := json.MarshalIndent(endpoints, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		util.Logger.Errorf("write response error, err=%s", err.Error())
	}
}

func (admin *Admin) Serve() {
	router := mux.NewRouter()

	router.HandleFunc("/", admin.Endpoints)
	router.HandleFunc("/add_token", admin.AddToken)

	listenAddr := DefaultListenAddr
	if admin.Config.AdminConfig.ListenAddr != "" {
		listenAddr = admin.Config.AdminConfig.ListenAddr
	}
	srv := &http.Server{
		Handler:      router,
		Addr:         listenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	util.Logger.Infof("start admin server at %s", srv.Addr)

	err := srv.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("start admin server error, err=%s", err.Error()))
	}
}
