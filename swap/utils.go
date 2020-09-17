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

func buildTokenInstance(tokens []model.Token) (map[string]*TokenInstance, error) {
	tokenPrivateKeys := make(map[string]*TokenInstance, len(tokens))
	for _, token := range tokens {
		bscPriKey, err := getBSCPrivateKey(&token)
		if err != nil {
			return nil, err
		}

		ethPriKey, err := getETHPrivateKey(&token)
		if err != nil {
			return nil, err
		}
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

		tokenPrivateKeys[token.Symbol] = &TokenInstance{
			Symbol:          token.Symbol,
			Name:            token.Name,
			Decimals:        token.Decimals,
			LowBound:        lowBound,
			UpperBound:      upperBound,
			BSCPrivateKey:   bscPriKey,
			BSCContractAddr: ethcom.HexToAddress(token.BSCContractAddr),
			ETHPrivateKey:   ethPriKey,
			ETHContractAddr: ethcom.HexToAddress(token.ETHContractAddr),
		}
	}

	return tokenPrivateKeys, nil
}

func getBSCPrivateKey(token *model.Token) (*ecdsa.PrivateKey, error) {
	var bscPrivateKey string
	if token.BSCKeyType == common.AWSPrivateKey {
		result, err := util.GetSecret(token.BSCKeyAWSSecretName, token.BSCKeyAWSRegion)
		if err != nil {
			return nil, err
		}
		type AwsPrivateKey struct {
			PrivateKey string `json:"private_key"`
		}
		var awsPrivateKey AwsPrivateKey
		err = json.Unmarshal([]byte(result), &awsPrivateKey)
		if err != nil {
			return nil, err
		}
		bscPrivateKey = awsPrivateKey.PrivateKey
	} else {
		bscPrivateKey = token.BSCPrivateKey
	}

	return crypto.HexToECDSA(bscPrivateKey)
}

func getETHPrivateKey(token *model.Token) (*ecdsa.PrivateKey, error) {
	var ethPrivateKey string
	if token.ETHKeyType == common.AWSPrivateKey {
		result, err := util.GetSecret(token.ETHKeyAWSSecretName, token.ETHKeyAWSRegion)
		if err != nil {
			return nil, err
		}
		type AwsPrivateKey struct {
			PrivateKey string `json:"private_key"`
		}
		var awsPrivateKey AwsPrivateKey
		err = json.Unmarshal([]byte(result), &awsPrivateKey)
		if err != nil {
			return nil, err
		}
		ethPrivateKey = awsPrivateKey.PrivateKey
	} else {
		ethPrivateKey = token.ETHPrivateKey
	}

	return crypto.HexToECDSA(ethPrivateKey)
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

func getCallOpts() (*bind.CallOpts, error) {
	callOpts := &bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	return callOpts, nil
}


