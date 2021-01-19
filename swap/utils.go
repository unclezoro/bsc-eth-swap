package swap

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/binance-chain/tss-crypto-toolkit/ec"
	rsaTool "github.com/binance-chain/tss-crypto-toolkit/rsa"
	nt "github.com/binance-chain/tss-zerotrust-sdk/network"
	tsssdksecure "github.com/binance-chain/tss-zerotrust-sdk/secure"
	tsssdktypes "github.com/binance-chain/tss-zerotrust-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/swap/erc20"
	"github.com/binance-chain/bsc-eth-swap/util"
)

func buildTokenInstance(tokens []model.Token, cfg *util.Config) (map[string]*TokenInstance, error) {
	tokenInstances := make(map[string]*TokenInstance, len(tokens))

	for _, token := range tokens {

		lowBound := big.NewInt(0)
		_, ok := lowBound.SetString(token.LowBound, 10)
		if !ok {
			panic(fmt.Sprintf("invalid lowBound amount: %s", token.LowBound))
		}
		upperBound := big.NewInt(0)
		_, ok = upperBound.SetString(token.UpperBound, 10)
		if !ok {
			panic(fmt.Sprintf("invalid upperBound amount: %s", token.LowBound))
		}

		bscERC20Threshold := big.NewInt(0)
		bscERC20Threshold.SetString(token.BSCERC20Threshold, 10)

		ethERC20Threshold := big.NewInt(0)
		ethERC20Threshold.SetString(token.ETHERC20Threshold, 10)

		tokenInstances[token.Symbol] = &TokenInstance{
			Symbol:               token.Symbol,
			Name:                 token.Name,
			Decimals:             token.Decimals,
			LowBound:             lowBound,
			UpperBound:           upperBound,
			CloseSignal:          make(chan bool),
			BSCTokenContractAddr: ethcom.HexToAddress(token.BSCTokenContractAddr),
			BSCERC20Threshold:    bscERC20Threshold,
			ETHTokenContractAddr: ethcom.HexToAddress(token.ETHTokenContractAddr),
			ETHERC20Threshold:    ethERC20Threshold,
		}
	}

	return tokenInstances, nil
}

func GetHMACKey(cfg *util.Config) (string, error) {
	if cfg.KeyManagerConfig.KeyType == common.AWSPrivateKey {
		result, err := util.GetSecret(cfg.KeyManagerConfig.AWSSecretName, cfg.KeyManagerConfig.AWSRegion)
		if err != nil {
			return "", err
		}

		keyConfig := util.KeyConfig{}
		err = json.Unmarshal([]byte(result), &keyConfig)
		if err != nil {
			return "", err
		}
		return keyConfig.HMACKey, nil
	} else {
		return cfg.KeyManagerConfig.LocalHMACKey, nil
	}
}

func abiEncodeTransfer(recipient ethcom.Address, amount *big.Int) ([]byte, error) {
	erc20ABI, err := abi.JSON(strings.NewReader(erc20.Erc20ABI))
	if err != nil {
		return nil, err
	}

	data, err := erc20ABI.Pack("transfer", recipient, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func generateTxOpt(privateKey *ecdsa.PrivateKey) *bind.TransactOpts {
	txOpts := bind.NewKeyedTransactor(privateKey)
	txOpts.Value = big.NewInt(0)
	return txOpts
}

func getCallOpts() *bind.CallOpts {
	callOpts := &bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	return callOpts
}

func GetAddress(pubKey *ecdsa.PublicKey) ethcom.Address {
	return crypto.PubkeyToAddress(*pubKey)
}

func BuildKeys(privateKeyStr string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	if strings.HasPrefix(privateKeyStr, "0x") {
		privateKeyStr = privateKeyStr[2:]
	}
	priKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, nil, err
	}
	publicKey, ok := priKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, fmt.Errorf("get public key error")
	}
	return priKey, publicKey, nil
}

func NewClientSecureConfig(cfg *util.Config) *tsssdksecure.ClientSecureConfig {
	rsaPrvBz, err := base64.StdEncoding.DecodeString(cfg.KeyManagerConfig.TSSCfg.RSAPrvB64)
	if err != nil {
		panic(err.Error())
	}

	rsaPrv, err := rsaTool.UnmarshalPKCS8PrivateKey(rsaPrvBz)
	if err != nil {
		panic(err.Error())
	}

	p521PrvBz, err := base64.StdEncoding.DecodeString(cfg.KeyManagerConfig.TSSCfg.P521PrvB64)
	if err != nil {
		panic(err.Error())
	}

	p521Prv, err := ec.UnmarshalPrivateKey(p521PrvBz)
	if err != nil {
		panic(err.Error())
	}

	peerRsaPubBz, err := base64.StdEncoding.DecodeString(cfg.KeyManagerConfig.TSSCfg.RSAPrvB64ForServerPub)
	if err != nil {
		panic(err.Error())
	}

	peerRsaPub, err := rsaTool.UnmarshalPKIXPublicKey(peerRsaPubBz)
	if err != nil {
		panic(err.Error())
	}

	peerP521PubBz, err := base64.StdEncoding.DecodeString(cfg.KeyManagerConfig.TSSCfg.P521PrvForServerPub)
	if err != nil {
		panic(err.Error())
	}

	peerP521Pub, err := ec.UnmarshalPKIXPublicKey(peerP521PubBz)
	if err != nil {
		panic(err.Error())
	}

	return &tsssdksecure.ClientSecureConfig{
		SecureMode: true,
		RsaPrv:     rsaPrv,
		EcPrv:      p521Prv,
		PeerRsaPub: peerRsaPub,
		PeerEcPub:  peerP521Pub,
	}
}

func signBSC(secureConfig *tsssdksecure.ClientSecureConfig, endPoint string, from, to, amount string, contract, value string,
	chainId int64, nonce int64, gasPrice string, gasLimit string, secureMode bool) (*tsssdktypes.SignResponse, error) {
	request := &tsssdktypes.SignETHRequest{
		From:     from,
		To:       to,
		Amount:   amount,
		Contract: contract,
		Value:    value,
		ChainId:  chainId,
		Nonce:    nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	nt.BinanceSC.SetEndPoint(endPoint)
	secureConfig.SecureMode = secureMode
	nt.BinanceSC.SetSecureConfig(secureConfig)

	return nt.BinanceSC.Sign(request)
}

func signETH(secureConfig *tsssdksecure.ClientSecureConfig, endPoint string, from, to, amount string, contract, value string,
	chainId int64, nonce int64, gasPrice string, gasLimit string, secureMode bool) (*tsssdktypes.SignResponse, error) {
	request := &tsssdktypes.SignETHRequest{
		From:     from,
		To:       to,
		Amount:   amount,
		Contract: contract,
		Value:    value,
		ChainId:  chainId,
		Nonce:    nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	nt.Ethereum.SetEndPoint(endPoint)
	secureConfig.SecureMode = secureMode
	nt.Ethereum.SetSecureConfig(secureConfig)

	return nt.Ethereum.Sign(request)
}
