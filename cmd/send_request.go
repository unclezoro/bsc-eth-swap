package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/binance-chain/bsc-eth-swap/util"
)

type Request struct {
	ApiKey      string      `json:"api_key"`
	ApiSecret   string      `json:"api_secret"`
	Endpoint    string      `json:"endpoint"`
	Method      string      `json:"method"`
	RequestBody interface{} `json:"request_body"`
}

const (
	flagReqPath = "request-path"
)

func initFlags() {
	flag.String(flagReqPath, "", "request path")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(fmt.Sprintf("bind flags error, err=%s", err))
	}
}

func printUsage() {
	fmt.Print("usage: ./send_request --request-path request_file_path\n")
}

func main() {
	initFlags()

	reqFilePath := viper.GetString(flagReqPath)
	if reqFilePath == "" {
		printUsage()
		return
	}

	bz, err := ioutil.ReadFile(reqFilePath)
	if err != nil {
		panic(err)
	}

	var req Request
	if err := json.Unmarshal(bz, &req); err != nil {
		panic(err)
	}

	if req.ApiKey == "" {
		println("api_key should not be empty")
		return
	}
	if req.ApiSecret == "" {
		println("api_secret should not be empty")
		return
	}
	if req.Endpoint == "" {
		println("endpoint should not be empty")
		return
	}
	if req.Method == "" {
		println("method should not be empty")
		return
	}
	if req.RequestBody == nil {
		println("request body should not be empty")
		return
	}

	body, err := json.Marshal(req.RequestBody)
	if err != nil {
		println("marshal request body error")
		return
	}

	signer := util.NewHmacSigner(req.ApiKey, req.ApiSecret)
	hash := signer.Sign(body)

	httpReq, err := http.NewRequest(req.Method, req.Endpoint, bytes.NewReader(body))
	if err != nil {
		println("new request error")
		return
	}
	httpReq.Header.Set("ApiKey", req.ApiKey)
	httpReq.Header.Set("Authorization", hash)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		println(fmt.Sprintf("send request error, err=%s", err.Error()))
		return
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println(fmt.Sprintf("get response body error, err=%s", err.Error()))
		return
	}
	fmt.Printf("Response: %s\n", string(resBody))
}
