package admin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/swap"
	"github.com/binance-chain/bsc-eth-swap/util"
	cmm "github.com/binance-chain/bsc-eth-swap/common"
)

const (
	DefaultListenAddr = "0.0.0.0:8080"

	MaxIconUrlLength = 400
)

type Admin struct {
	DB *gorm.DB

	cfg *util.Config

	hmacSigner *util.HmacSigner
	swapEngine *swap.SwapEngine
}

func NewAdmin(config *util.Config, db *gorm.DB, signer *util.HmacSigner, swapEngine *swap.SwapEngine) *Admin {
	return &Admin{
		DB:         db,
		cfg:        config,
		hmacSigner: signer,
		swapEngine: swapEngine,
	}
}

func updateCheck(update *updateSwapPairRequest) error {
	if update.ERC20Addr == "" {
		return fmt.Errorf("bsc_token_contract_addr can't be empty")
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
	return nil
}

func (admin *Admin) UpdateSwapPairHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := admin.checkAuth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var updateSwapPair updateSwapPairRequest
	err = json.Unmarshal(reqBody, &updateSwapPair)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updateCheck(&updateSwapPair); err != nil {
		http.Error(w, fmt.Sprintf("parameters is invalid, %v", err), http.StatusBadRequest)
		return
	}

	swapPair := model.SwapPair{}
	err = admin.DB.Where("erc20_addr = ?", updateSwapPair.ERC20Addr).First(&swapPair).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("swapPair %s is not found", updateSwapPair.ERC20Addr), http.StatusBadRequest)
		return
	}

	toUpdate := map[string]interface{}{
		"available": updateSwapPair.Available,
	}

	if updateSwapPair.LowerBound != "" {
		toUpdate["low_bound"] = updateSwapPair.LowerBound
	}
	if updateSwapPair.UpperBound != "" {
		toUpdate["upper_bound"] = updateSwapPair.UpperBound
	}
	if updateSwapPair.IconUrl != "" {
		toUpdate["icon_url"] = updateSwapPair.IconUrl
	}

	err = admin.DB.Model(model.SwapPair{}).Where("erc20_addr = ?", updateSwapPair.ERC20Addr).Updates(toUpdate).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("update swapPair error, err=%s", err.Error()), http.StatusInternalServerError)
		return
	}

	// get swapPair
	swapPair = model.SwapPair{}
	err = admin.DB.Where("erc20_addr = ?", updateSwapPair.ERC20Addr).First(&swapPair).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("swapPair %s is not found", updateSwapPair.ERC20Addr), http.StatusBadRequest)
		return
	}

	swapPairIns, err := admin.swapEngine.GetSwapPairInstance(common.HexToAddress(updateSwapPair.ERC20Addr))
	// disable is only for frontend, do not affect backend
	// if we want to disable it in backend, set the low_bound and upper_bound to be zero
	if err != nil && updateSwapPair.Available {
		// add swapPair in swapper
		err = admin.swapEngine.AddSwapPairInstance(&swapPair)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if swapPairIns != nil {
		admin.swapEngine.UpdateSwapInstance(&swapPair)
	}

	jsonBytes, err := json.MarshalIndent(swapPair, "", "  ")
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

func (admin *Admin) Endpoints(w http.ResponseWriter, r *http.Request) {
	endpoints := struct {
		Endpoints []string `json:"endpoints"`
	}{
		Endpoints: []string{
			"/update_swap_pair",
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

func (admin *Admin) WithdrawToken(w http.ResponseWriter, r *http.Request) {
	reqBody, err := admin.checkAuth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var withdrawToken withdrawTokenRequest
	err = json.Unmarshal(reqBody, &withdrawToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = withdrawCheck(&withdrawToken); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	amount := big.NewInt(0)
	amount.SetString(withdrawToken.Amount, 10)

	var withdrawResp withdrawTokenResponse
	withdrawResp.TxHash, err = admin.swapEngine.WithdrawToken(withdrawToken.Chain,
		common.HexToAddress(withdrawToken.TokenAddr),
		common.HexToAddress(withdrawToken.Recipient), amount)
	if err != nil {
		withdrawResp.ErrMsg = err.Error()
	}

	jsonBytes, err := json.MarshalIndent(withdrawResp, "", "    ")
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

func withdrawCheck(withdraw *withdrawTokenRequest) error {
	if strings.ToUpper(withdraw.Chain) != cmm.ChainBSC && strings.ToUpper(withdraw.Chain) != cmm.ChainETH {
		return fmt.Errorf("bsc_token_contract_addr can't be empty")
	}
	if !common.IsHexAddress(withdraw.TokenAddr) {
		return fmt.Errorf("token address is not a valid address")
	}
	if !common.IsHexAddress(withdraw.Recipient) {
		return fmt.Errorf("recipient is not a valid address")
	}
	_, ok := big.NewInt(0).SetString(withdraw.Amount, 10)
	if !ok {
		return fmt.Errorf("invalid input, expected big integer")
	}
	return nil
}

func (admin *Admin) RetryFailedSwaps(w http.ResponseWriter, r *http.Request) {
	reqBody, err := admin.checkAuth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var retryFailedSwaps retryFailedSwapsRequest
	err = json.Unmarshal(reqBody, &retryFailedSwaps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var retryFailedSwapsResp retryFailedSwapsResponse
	retryFailedSwapsResp.SwapIDList, retryFailedSwapsResp.RejectedSwapIDList, err = admin.swapEngine.InsertRetryFailedSwaps(retryFailedSwaps.SwapIDList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		retryFailedSwapsResp.ErrMsg = err.Error()
	} else {
		w.WriteHeader(http.StatusOK)
	}

	jsonBytes, err := json.MarshalIndent(retryFailedSwapsResp, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonBytes)
	if err != nil {
		util.Logger.Errorf("write response error, err=%s", err.Error())
	}
}

func (admin *Admin) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (admin *Admin) checkAuth(r *http.Request) ([]byte, error) {
	apiKey := r.Header.Get("ApiKey")
	hash := r.Header.Get("Authorization")

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if admin.hmacSigner.ApiKey != apiKey {
		return nil, fmt.Errorf("api key mismatch")
	}

	if !admin.hmacSigner.Verify(payload, hash) {
		return nil, fmt.Errorf("invalud auth")
	}
	return payload, nil
}

func (admin *Admin) Serve() {
	router := mux.NewRouter()

	router.HandleFunc("/", admin.Endpoints).Methods("GET")
	router.HandleFunc("/healthz", admin.Healthz).Methods("GET")
	router.HandleFunc("/update_swap_pair", admin.UpdateSwapPairHandler).Methods("PUT")
	router.HandleFunc("/withdraw_token", admin.WithdrawToken).Methods("POST")
	router.HandleFunc("/retry_failed_swaps", admin.RetryFailedSwaps).Methods("POST")

	listenAddr := DefaultListenAddr
	if admin.cfg.AdminConfig.ListenAddr != "" {
		listenAddr = admin.cfg.AdminConfig.ListenAddr
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
