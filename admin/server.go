package admin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/swap"
	"github.com/binance-chain/bsc-eth-swap/util"
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

type updateSwapPairRequest struct {
	BSCTokenContractAddr string `json:"bsc_token_contract_addr"`

	Available bool `json:"available"`

	LowerBound string `json:"lower_bound"`
	UpperBound string `json:"upper_bound"`

	IconUrl string `json:"icon_url"`
}

func updateCheck(update *updateSwapPairRequest) error {
	if update.BSCTokenContractAddr == "" {
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
	apiKey := r.Header.Get("ApiKey")
	auth := r.Header.Get("Authorization")

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !admin.checkAuth(apiKey, reqBody, auth) {
		http.Error(w, "auth is not correct", http.StatusUnauthorized)
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
	err = admin.DB.Where("bsc_token_contract_addr = ?", updateSwapPair.BSCTokenContractAddr).First(&swapPair).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("swapPair %s is not found", updateSwapPair.BSCTokenContractAddr), http.StatusBadRequest)
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

	err = admin.DB.Model(model.SwapPair{}).Where("bsc_token_contract_addr = ?", updateSwapPair.BSCTokenContractAddr).Updates(toUpdate).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("update swapPair error, err=%s", err.Error()), http.StatusInternalServerError)
		return
	}

	// get swapPair
	swapPair = model.SwapPair{}
	err = admin.DB.Where("bsc_token_contract_addr = ?", updateSwapPair.BSCTokenContractAddr).First(&swapPair).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("swapPair %s is not found", updateSwapPair.BSCTokenContractAddr), http.StatusBadRequest)
		return
	}

	swapPairIns := admin.swapEngine.GetSwapPairInstance(common.HexToAddress(updateSwapPair.BSCTokenContractAddr))
	// disable is only for frontend, do not affect backend
	// if we want to disable it in backend, set the low_bound and upper_bound to be zero
	if swapPairIns == nil {
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

func (admin *Admin) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (admin *Admin) checkAuth(apiKey string, payload []byte, hash string) bool {
	if admin.hmacSigner.ApiKey != apiKey {
		return false
	}

	return admin.hmacSigner.Verify(payload, hash)
}

func (admin *Admin) Serve() {
	router := mux.NewRouter()

	router.HandleFunc("/", admin.Endpoints).Methods("GET")
	router.HandleFunc("/healthz", admin.Healthz).Methods("GET")
	router.HandleFunc("/update_swap_pair", admin.UpdateSwapPairHandler).Methods("PUT")

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
