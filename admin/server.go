package admin

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	scmn "github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/executor"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/swap"
	"github.com/binance-chain/bsc-eth-swap/util"
)

const (
	DefaultListenAddr = "0.0.0.0:8080"

	MaxTokenLength   = 20
	MaxIconUrlLength = 400
)

var isAlphaNumFunc = regexp.MustCompile(`^[[:alnum:]]+$`).MatchString

type Admin struct {
	DB *gorm.DB

	Config *util.Config

	Swapper *swap.Swapper

	BSCExecutor executor.Executor
	ETHExecutor executor.Executor
}

func NewAdmin(config *util.Config, db *gorm.DB, swapper *swap.Swapper, bscExecutor executor.Executor, ethExecutor executor.Executor) *Admin {
	return &Admin{
		DB:          db,
		Config:      config,
		Swapper:     swapper,
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

	IconUrl string `json:"icon_url"`

	BSCKeyType          string `json:"bsc_key_type"`
	BSCKeyAWSRegion     string `json:"bsc_key_aws_region"`
	BSCKeyAWSSecretName string `json:"bsc_key_aws_secret_name"`
	BSCPrivateKey       string `json:"bsc_private_key"`
	BSCSendAddr         string `json:"bsc_sender"`
	BSCERC20Threshold   string `json:"bsc_erc20_threshold"`

	ETHKeyType          string `json:"eth_key_type"`
	ETHKeyAWSRegion     string `json:"eth_aws_region"`
	ETHKeyAWSSecretName string `json:"eth_key_aws_secret_name"`
	ETHPrivateKey       string `json:"eth_private_key"`
	ETHSendAddr         string `json:"eth_send_addr"`
	ETHERC20Threshold   string `json:"eth_erc20_threshold"`
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

	// check symbol
	bscSymbol, err := admin.BSCExecutor.GetContractSymbol(common.HexToAddress(newToken.BSCContractAddr))
	if err != nil {
		http.Error(w, fmt.Sprintf("get bsc symbol error, addr=%s, err=%s", newToken.BSCContractAddr, err.Error()), http.StatusBadRequest)
		return
	}

	ethSymbol, err := admin.ETHExecutor.GetContractSymbol(common.HexToAddress(newToken.ETHContractAddr))
	if err != nil {
		http.Error(w, fmt.Sprintf("get eth symbol error, addr=%s, err=%s", newToken.ETHContractAddr, err.Error()), http.StatusBadRequest)
		return
	}

	if bscSymbol != ethSymbol || bscSymbol != newToken.Symbol {
		http.Error(w, fmt.Sprintf("symbol is wrong, bsc_symbol=%s, eth_symbol=%s", bscSymbol, ethSymbol), http.StatusBadRequest)
		return
	}

	// check decimals
	bscDecimals, err := admin.BSCExecutor.GetContractDecimals(common.HexToAddress(newToken.BSCContractAddr))
	if err != nil {
		http.Error(w, fmt.Sprintf("get bsc decimals error, addr=%s, err=%s", newToken.BSCContractAddr, err.Error()), http.StatusBadRequest)
		return
	}

	ethDecimals, err := admin.ETHExecutor.GetContractDecimals(common.HexToAddress(newToken.ETHContractAddr))
	if err != nil {
		http.Error(w, fmt.Sprintf("get eth decimals error, addr=%s, err=%s", newToken.ETHContractAddr, err.Error()), http.StatusBadRequest)
		return
	}

	if bscDecimals != ethDecimals || bscDecimals != newToken.Decimals {
		http.Error(w, fmt.Sprintf("decimals is wrong, bsc_decimals=%d, eth_decimals=%d", bscDecimals, ethDecimals), http.StatusBadRequest)
		return
	}

	bscPriKey, bscPubKey, err := swap.GetPrivateKey(newToken.BSCKeyType, newToken.BSCKeyAWSSecretName, newToken.BSCKeyAWSRegion, newToken.BSCPrivateKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get bsc private key: %s", err.Error()), http.StatusBadRequest)
		return
	}

	ethPriKey, ethPubKey, err := swap.GetPrivateKey(newToken.ETHKeyType, newToken.ETHKeyAWSSecretName, newToken.ETHKeyAWSRegion, newToken.ETHPrivateKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get eth private key: %s", err.Error()), http.StatusBadRequest)
		return
	}

	tokenModel := model.Token{
		Symbol:               newToken.Symbol,
		Name:                 newToken.Name,
		Decimals:             newToken.Decimals,
		BSCTokenContractAddr: strings.ToLower(common.HexToAddress(newToken.BSCContractAddr).String()),
		ETHTokenContractAddr: strings.ToLower(common.HexToAddress(newToken.ETHContractAddr).String()),
		LowBound:             newToken.LowerBound,
		UpperBound:           newToken.UpperBound,
		IconUrl:              newToken.IconUrl,
		BSCKeyType:           newToken.BSCKeyType,
		BSCKeyAWSRegion:      newToken.BSCKeyAWSRegion,
		BSCKeyAWSSecretName:  newToken.BSCKeyAWSSecretName,
		BSCPrivateKey:        newToken.BSCPrivateKey,
		BSCSendAddr:          strings.ToLower(swap.GetAddress(bscPubKey).String()),
		BSCERC20Threshold:    newToken.BSCERC20Threshold,
		ETHKeyType:           newToken.ETHKeyType,
		ETHKeyAWSRegion:      newToken.ETHKeyAWSRegion,
		ETHKeyAWSSecretName:  newToken.ETHKeyAWSSecretName,
		ETHPrivateKey:        newToken.ETHPrivateKey,
		ETHSendAddr:          strings.ToLower(swap.GetAddress(ethPubKey).String()),
		ETHERC20Threshold:    newToken.ETHERC20Threshold,
		Available:            false,
	}

	err = admin.DB.Create(&tokenModel).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get token
	token := model.Token{}
	err = admin.DB.Where("symbol = ?", tokenModel.Symbol).First(&token).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("token %s is not found", tokenModel.Symbol), http.StatusBadRequest)
		return
	}

	// add token in swapper
	err = admin.Swapper.AddToken(&token, bscPriKey, bscPubKey, ethPriKey, ethPubKey)
	if err != nil {
		dbErr := admin.DB.Where("symbol = ?", tokenModel.Symbol).Unscoped().Delete(&model.Token{}).Error
		if dbErr != nil {
			http.Error(w, fmt.Sprintf("delete token error, err=%s", dbErr.Error()), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.MarshalIndent(token, "", "  ")
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

func tokenBasicCheck(token *NewTokenRequest) error {
	if len(token.Symbol) == 0 || len(token.Symbol) > MaxTokenLength {
		return fmt.Errorf("symbol length invalid")
	}
	if !isAlphaNumFunc(token.Symbol) {
		return fmt.Errorf("symbol contains invalid character")
	}
	if len(token.IconUrl) > MaxIconUrlLength {
		return fmt.Errorf("icon length exceed limit")
	}
	if token.Name == "" {
		return fmt.Errorf("name should not be empty")
	}
	if token.Decimals <= 0 {
		return fmt.Errorf("decimals should be larger than 0")
	}
	if token.LowerBound == "" {
		return fmt.Errorf("lower_bound should not be empty")
	}
	if token.UpperBound == "" {
		return fmt.Errorf("upper_bound should not be empty")
	}

	if _, ok := big.NewInt(0).SetString(token.UpperBound, 10); !ok {
		return fmt.Errorf("invalid upperBound amount: %s", token.UpperBound)
	}

	if _, ok := big.NewInt(0).SetString(token.LowerBound, 10); !ok {
		return fmt.Errorf("invalid lowerBound amount: %s", token.LowerBound)
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
	if !common.IsHexAddress(token.BSCSendAddr) {
		return fmt.Errorf("bse_sender_addr is wrong")
	}
	// check bsc key
	if token.BSCKeyType != scmn.LocalPrivateKey && token.BSCKeyType != scmn.AWSPrivateKey {
		return fmt.Errorf("bsc_key_type should be %s or %s", scmn.LocalPrivateKey, scmn.AWSPrivateKey)
	}
	if token.BSCKeyType == scmn.AWSPrivateKey {
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
	if token.ETHKeyType != scmn.LocalPrivateKey && token.ETHKeyType != scmn.AWSPrivateKey {
		return fmt.Errorf("eth_key_type should be %s or %s", scmn.LocalPrivateKey, scmn.AWSPrivateKey)
	}
	if token.ETHKeyType == scmn.AWSPrivateKey {
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

type UpdateTokenRequest struct {
	Symbol string `json:"symbol"`

	Available bool `json:"available"`

	LowerBound string `json:"lower_bound"`
	UpperBound string `json:"upper_bound"`

	IconUrl string `json:"icon_url"`

	BSCKeyAWSSecretName string `json:"bsc_key_aws_secret_name"`
	BSCSendAddr         string `json:"bsc_sender"`

	ETHKeyAWSSecretName string `json:"eth_key_aws_secret_name"`
	ETHSendAddr         string `json:"eth_send_addr"`
}

func updateCheck(update *UpdateTokenRequest) error {
	if len(update.Symbol) == 0 || len(update.Symbol) > MaxTokenLength {
		return fmt.Errorf("symbol length invalid")
	}
	if !isAlphaNumFunc(update.Symbol) {
		return fmt.Errorf("symbol contains invalid character")
	}
	if update.UpperBound != "" {
		if _, ok := big.NewInt(0).SetString(update.UpperBound, 10); !ok {
			return fmt.Errorf("invalid upperBound amount: %s", update.UpperBound)
		}
	}
	if update.LowerBound != "" {
		if _, ok := big.NewInt(0).SetString(update.LowerBound, 10); !ok {
			return fmt.Errorf("invalid lowerBound amount: %s", update.LowerBound)
		}
	}
	if len(update.IconUrl) > MaxIconUrlLength {
		return fmt.Errorf("icon length exceed limit")
	}
	if update.ETHSendAddr != "" {
		if !common.IsHexAddress(update.ETHSendAddr) {
			return fmt.Errorf("eth_sender_addr is wrong")
		}
	}
	if update.BSCSendAddr != "" {
		if !common.IsHexAddress(update.BSCSendAddr) {
			return fmt.Errorf("bse_sender_addr is wrong")
		}
	}
	return nil
}

func (admin *Admin) UpdateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var updateToken UpdateTokenRequest

	err := json.NewDecoder(r.Body).Decode(&updateToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updateCheck(&updateToken); err != nil {
		http.Error(w, fmt.Sprintf("parameters is invalid, %v", err), http.StatusBadRequest)
		return
	}

	token := model.Token{}
	err = admin.DB.Where("symbol = ?", updateToken.Symbol).First(&token).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("token %s is not found", updateToken.Symbol), http.StatusBadRequest)
		return
	}

	toUpdate := map[string]interface{}{
		"available": updateToken.Available,
	}

	if updateToken.LowerBound != "" {
		toUpdate["low_bound"] = updateToken.LowerBound
	}
	if updateToken.UpperBound != "" {
		toUpdate["upper_bound"] = updateToken.UpperBound
	}
	if updateToken.BSCKeyAWSSecretName != "" {
		toUpdate["bsc_key_aws_secret_name"] = updateToken.BSCKeyAWSSecretName
	}
	if updateToken.BSCSendAddr != "" {
		toUpdate["bsc_send_addr"] = strings.ToLower(common.HexToAddress(updateToken.BSCSendAddr).String())
	}
	if updateToken.ETHKeyAWSSecretName != "" {
		toUpdate["eth_key_aws_secret_name"] = updateToken.ETHKeyAWSSecretName
	}
	if updateToken.ETHSendAddr != "" {
		toUpdate["eth_send_addr"] = strings.ToLower(common.HexToAddress(updateToken.ETHSendAddr).String())
	}
	if updateToken.IconUrl != "" {
		toUpdate["icon_url"] = updateToken.IconUrl
	}

	err = admin.DB.Model(model.Token{}).Where("symbol = ?", updateToken.Symbol).Updates(toUpdate).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("update token error, err=%s", err.Error()), http.StatusInternalServerError)
		return
	}

	// get token
	token = model.Token{}
	err = admin.DB.Where("symbol = ?", updateToken.Symbol).First(&token).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("token %s is not found", updateToken.Symbol), http.StatusBadRequest)
		return
	}
	jsonBytes, err := json.MarshalIndent(token, "", "  ")
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

func (admin *Admin) DeleteTokenHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	symbol := params["symbol"]
	if symbol == "" {
		http.Error(w, "required parameter 'symbol' is missing", http.StatusBadRequest)
		return
	}

	// check symbol
	if len(symbol) == 0 || len(symbol) > MaxTokenLength {
		http.Error(w, "symbol length invalid", http.StatusBadRequest)
		return
	}
	if !isAlphaNumFunc(symbol) {
		http.Error(w, "symbol contains invalid character", http.StatusBadRequest)
		return
	}

	token := model.Token{}
	err := admin.DB.Where("symbol = ?", symbol).First(&token).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("token %s is not found", symbol), http.StatusBadRequest)
		return
	}

	// check ongoing swaps
	var ongoingSwapCount = 0
	err = admin.DB.Model(model.Swap{}).Where("status not in (?)", []scmn.SwapStatus{swap.SwapQuoteRejected, swap.SwapSendFailed, swap.SwapSuccess}).Count(&ongoingSwapCount).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("find ongoing swaps error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if ongoingSwapCount > 0 {
		http.Error(w, fmt.Sprintf("there are onging swaps, can not delete token"), http.StatusBadRequest)
		return
	}

	// delete token
	err = admin.DB.Where("symbol = ?", symbol).Unscoped().Delete(&model.Token{}).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("delete token error, err=%s", err.Error()), http.StatusInternalServerError)
		return
	}

	// remove token in swapper
	admin.Swapper.RemoveToken(&token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (admin *Admin) Endpoints(w http.ResponseWriter, r *http.Request) {
	endpoints := struct {
		Endpoints []string `json:"endpoints"`
	}{
		Endpoints: []string{
			"/add_token",
			"/update_token",
			"/delete_token/{symbol}",
			"/healthz",
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

func (admin *Admin) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (admin *Admin) Serve() {
	router := mux.NewRouter()

	router.HandleFunc("/", admin.Endpoints).Methods("GET")
	router.HandleFunc("/healthz", admin.Healthz).Methods("GET")
	router.HandleFunc("/add_token", admin.AddToken).Methods("POST")
	router.HandleFunc("/update_token", admin.UpdateTokenHandler).Methods("PUT")
	router.HandleFunc("/delete_token/{symbol}", admin.DeleteTokenHandler).Methods("DELETE")

	listenAddr := DefaultListenAddr
	if admin.Config.AdminConfig.ListenAddr != "" {
		listenAddr = admin.Config.AdminConfig.ListenAddr
	}
	srv := &http.Server{
		Handler:      router,
		Addr:         listenAddr,
		WriteTimeout: 3 * time.Second,
		ReadTimeout:  3 * time.Second,
	}

	util.Logger.Infof("start admin server at %s", srv.Addr)

	err := srv.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("start admin server error, err=%s", err.Error()))
	}
}
