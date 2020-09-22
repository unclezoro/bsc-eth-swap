package swap

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/swap/erc20"
	"github.com/binance-chain/bsc-eth-swap/util"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func buildTokenInstance(tokens []model.Token, cfg *util.Config) (map[string]*TokenInstance, error) {
	tokenInstances := make(map[string]*TokenInstance, len(tokens))

	tokenKeys, err := GetAllTokenKeys(cfg)
	if err != nil {
		return nil, err
	}

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

		tokenKey, ok := tokenKeys[token.Symbol]
		if !ok {
			panic(fmt.Sprintf("Missing private key for %s", token.Symbol))
		}

		tokenInstances[token.Symbol] = &TokenInstance{
			Symbol:               token.Symbol,
			Name:                 token.Name,
			Decimals:             token.Decimals,
			LowBound:             lowBound,
			UpperBound:           upperBound,
			CloseSignal:          make(chan bool),
			BSCPrivateKey:        tokenKey.BSCPrivateKey,
			BSCTokenContractAddr: ethcom.HexToAddress(token.BSCTokenContractAddr),
			BSCTxSender:          GetAddress(tokenKey.BSCPublicKey),
			BSCERC20Threshold:    bscERC20Threshold,
			ETHPrivateKey:        tokenKey.ETHPrivateKey,
			ETHTokenContractAddr: ethcom.HexToAddress(token.ETHTokenContractAddr),
			ETHTxSender:          GetAddress(tokenKey.ETHPublicKey),
			ETHERC20Threshold:    ethERC20Threshold,
		}
	}

	return tokenInstances, nil
}

func GetAllTokenKeys(cfg *util.Config) (map[string]*TokenKey, error) {
	var tokenSecretKeys []util.TokenSecretKey
	if cfg.KeyManagerConfig.KeyType == common.AWSPrivateKey {
		result, err := util.GetSecret(cfg.KeyManagerConfig.AWSSecretName, cfg.KeyManagerConfig.AWSRegion)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(result), &tokenSecretKeys)
		if err != nil {
			return nil, err
		}
	} else {
		tokenSecretKeys = cfg.KeyManagerConfig.LocalKeys
	}

	keys := make(map[string]*TokenKey)
	for _, secretKey := range tokenSecretKeys {
		tokenKey := TokenKey{}

		priKey, publicKey, err := BuildKeys(secretKey.BSCPrivateKey)
		if err != nil {
			return nil, err
		}
		tokenKey.BSCPrivateKey = priKey
		tokenKey.BSCPublicKey = publicKey

		priKey, publicKey, err = BuildKeys(secretKey.ETHPrivateKey)
		if err != nil {
			return nil, err
		}
		tokenKey.ETHPrivateKey = priKey
		tokenKey.ETHPublicKey = publicKey

		keys[secretKey.Symbol] = &tokenKey
	}
	return keys, nil
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

func buildSignedTransaction(ethClient *ethclient.Client, privateKey *ecdsa.PrivateKey, contract ethcom.Address, input []byte) (*types.Transaction, error) {
	txOpts := bind.NewKeyedTransactor(privateKey)

	nonce, err := ethClient.PendingNonceAt(context.Background(), txOpts.From)
	if err != nil {
		return nil, err
	}
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	value := big.NewInt(0)
	msg := ethereum.CallMsg{From: txOpts.From, To: &contract, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err := ethClient.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}

	rawTx := types.NewTransaction(nonce, contract, value, gasLimit, gasPrice, input)
	signedTx, err := txOpts.Signer(types.HomesteadSigner{}, txOpts.From, rawTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
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
