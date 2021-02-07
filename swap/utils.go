package swap

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
	"strconv"

	"github.com/binance-chain/tss-crypto-toolkit/ec"
	rsaTool "github.com/binance-chain/tss-crypto-toolkit/rsa"
	nt "github.com/binance-chain/tss-zerotrust-sdk/network"
	tsssdksecure "github.com/binance-chain/tss-zerotrust-sdk/secure"
	tsssdktypes "github.com/binance-chain/tss-zerotrust-sdk/types"

	contractabi "github.com/binance-chain/bsc-eth-swap/abi"
	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

func buildSwapPairInstance(pairs []model.SwapPair) (map[ethcom.Address]*SwapPairIns, error) {
	swapPairInstances := make(map[ethcom.Address]*SwapPairIns, len(pairs))

	for _, pair := range pairs {

		lowBound := big.NewInt(0)
		_, ok := lowBound.SetString(pair.LowBound, 10)
		if !ok {
			panic(fmt.Sprintf("invalid lowBound amount: %s", pair.LowBound))
		}
		upperBound := big.NewInt(0)
		_, ok = upperBound.SetString(pair.UpperBound, 10)
		if !ok {
			panic(fmt.Sprintf("invalid upperBound amount: %s", pair.LowBound))
		}

		swapPairInstances[ethcom.HexToAddress(pair.ERC20Addr)] = &SwapPairIns{
			Symbol:     pair.Symbol,
			Name:       pair.Name,
			Decimals:   pair.Decimals,
			LowBound:   lowBound,
			UpperBound: upperBound,
			BEP20Addr:  ethcom.HexToAddress(pair.BEP20Addr),
			ERC20Addr:  ethcom.HexToAddress(pair.ERC20Addr),
		}

		util.Logger.Infof("Load swap pair, symbol %s, bep20 address %s, erc20 address %s", pair.Symbol, pair.BEP20Addr, pair.ERC20Addr)
	}

	return swapPairInstances, nil
}

func GetKeyConfig(cfg *util.Config) (*util.KeyConfig, error) {
	if cfg.KeyManagerConfig.KeyType == common.AWSPrivateKey {
		result, err := util.GetSecret(cfg.KeyManagerConfig.AWSSecretName, cfg.KeyManagerConfig.AWSRegion)
		if err != nil {
			return nil, err
		}

		keyConfig := util.KeyConfig{}
		err = json.Unmarshal([]byte(result), &keyConfig)
		if err != nil {
			return nil, err
		}
		return &keyConfig, nil
	} else {
		return &util.KeyConfig{
			HMACKey:               cfg.KeyManagerConfig.LocalHMACKey,
			AdminApiKey:           cfg.KeyManagerConfig.LocalAdminApiKey,
			AdminSecretKey:        cfg.KeyManagerConfig.LocalAdminSecretKey,
			P521PrvB64:            cfg.KeyManagerConfig.LocalP521PrvB64,
			P521PrvForServerPub:   cfg.KeyManagerConfig.LocalP521PrvForServerPub,
			RSAPrvB64:             cfg.KeyManagerConfig.LocalRSAPrvB64,
			RSAPrvB64ForServerPub: cfg.KeyManagerConfig.LocalRSAPrvB64ForServerPub,
		}, nil
	}
}

func abiEncodeFillETH2BSCSwap(ethTxHash ethcom.Hash, erc20Addr ethcom.Address, toAddress ethcom.Address, amount *big.Int, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("fillETH2BSCSwap", ethTxHash, erc20Addr, toAddress, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func abiEncodeERC20Transfer(recipient ethcom.Address, amount *big.Int, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("transfer", recipient, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func abiEncodeFillBSC2ETHSwap(ethTxHash ethcom.Hash, erc20Addr ethcom.Address, toAddress ethcom.Address, amount *big.Int, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("fillBSC2ETHSwap", ethTxHash, erc20Addr, toAddress, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func abiEncodeCreateSwapPair(registerTxHash ethcom.Hash, erc20Addr ethcom.Address, name, symbol string, decimals uint8, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("createSwapPair", registerTxHash, erc20Addr, name, symbol, decimals)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func NewClientSecureConfig(keyCfg *util.KeyConfig) *tsssdksecure.ClientSecureConfig {
	rsaPrvBz, err := base64.StdEncoding.DecodeString(keyCfg.RSAPrvB64)
	if err != nil {
		panic(err.Error())
	}

	rsaPrv, err := rsaTool.UnmarshalPKCS8PrivateKey(rsaPrvBz)
	if err != nil {
		panic(err.Error())
	}

	p521PrvBz, err := base64.StdEncoding.DecodeString(keyCfg.P521PrvB64)
	if err != nil {
		panic(err.Error())
	}

	p521Prv, err := ec.UnmarshalPrivateKey(p521PrvBz)
	if err != nil {
		panic(err.Error())
	}

	peerRsaPubBz, err := base64.StdEncoding.DecodeString(keyCfg.RSAPrvB64ForServerPub)
	if err != nil {
		panic(err.Error())
	}

	peerRsaPub, err := rsaTool.UnmarshalPKIXPublicKey(peerRsaPubBz)
	if err != nil {
		panic(err.Error())
	}

	peerP521PubBz, err := base64.StdEncoding.DecodeString(keyCfg.P521PrvForServerPub)
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
	chainId int64, nonce int64, gasPrice string, gasLimit string, data []byte, secureMode bool) (*tsssdktypes.SignResponse, error) {
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
		Data:     hex.EncodeToString(data),
	}

	nt.BinanceSC.SetEndPoint(endPoint)
	secureConfig.SecureMode = secureMode
	nt.BinanceSC.SetSecureConfig(secureConfig)

	return nt.BinanceSC.Sign(request)
}

func signETH(secureConfig *tsssdksecure.ClientSecureConfig, endPoint string, from, to, amount string, contract, value string,
	chainId int64, nonce int64, gasPrice string, gasLimit string, data []byte, secureMode bool) (*tsssdktypes.SignResponse, error) {
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
		Data:     hex.EncodeToString(data),
	}

	nt.Ethereum.SetEndPoint(endPoint)
	secureConfig.SecureMode = secureMode
	nt.Ethereum.SetSecureConfig(secureConfig)

	return nt.Ethereum.Sign(request)
}

func buildSignedTransaction(network string, txSender, contract ethcom.Address, ethClient *ethclient.Client, txInput []byte, tssConfig *tsssdksecure.ClientSecureConfig, endpoint string) (*types.Transaction, error) {

	nonce, err := ethClient.PendingNonceAt(context.Background(), txSender)
	if err != nil {
		return nil, err
	}
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	value := big.NewInt(0)
	msg := ethereum.CallMsg{From: txSender, To: &contract, GasPrice: gasPrice, Value: value, Data: txInput}
	gasLimit, err := ethClient.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	gasLimit = gasLimit * 2

	chainId, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chainid: %v", err)
	}

	var signedRawTx *tsssdktypes.SignResponse
	if network == common.ChainBSC {
		signedRawTx, err = signBSC(tssConfig, endpoint, txSender.String(),
			contract.String(), "0", "", "", chainId.Int64(), int64(nonce), "0x"+strconv.FormatInt(gasPrice.Int64(), 16), "0x"+strconv.FormatInt(int64(gasLimit), 16), txInput, true)
		if err != nil {
			return nil, fmt.Errorf("TSS server failure: %v", err)
		}
	} else {
		signedRawTx, err = signETH(tssConfig, endpoint, txSender.String(),
			contract.String(), "0", "", "", chainId.Int64(), int64(nonce), "0x"+strconv.FormatInt(gasPrice.Int64(), 16), "0x"+strconv.FormatInt(int64(gasLimit), 16), txInput, true)
		if err != nil {
			return nil, fmt.Errorf("TSS server failure: %v", err)
		}
	}

	var signedTx types.Transaction
	err = rlp.DecodeBytes(ethcom.FromHex(signedRawTx.RawTransaction), &signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TSS signed result: %v", err)
	}

	return &signedTx, nil
}

func buildNativeCoinTransferTx(network string, txSender, recipient ethcom.Address, amount *big.Int, ethClient *ethclient.Client, tssConfig *tsssdksecure.ClientSecureConfig, endpoint string) (*types.Transaction, error) {
	nonce, err := ethClient.PendingNonceAt(context.Background(), txSender)
	if err != nil {
		return nil, err
	}
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	msg := ethereum.CallMsg{From: txSender, To: &recipient, GasPrice: gasPrice, Value: amount}
	gasLimit, err := ethClient.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	gasLimit = gasLimit * 2

	chainId, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chainid: %v", err)
	}

	var signedRawTx *tsssdktypes.SignResponse
	if network == common.ChainBSC {
		signedRawTx, err = signBSC(tssConfig, endpoint, txSender.String(),
			recipient.String(), amount.String(), "BNB", "", chainId.Int64(), int64(nonce), "0x"+strconv.FormatInt(gasPrice.Int64(), 16), "0x"+strconv.FormatInt(int64(gasLimit), 16), nil, true)
		if err != nil {
			return nil, fmt.Errorf("TSS server failure: %v", err)
		}
	} else {
		signedRawTx, err = signETH(tssConfig, endpoint, txSender.String(),
			recipient.String(), amount.String(), "ETH", "", chainId.Int64(), int64(nonce), "0x"+strconv.FormatInt(gasPrice.Int64(), 16), "0x"+strconv.FormatInt(int64(gasLimit), 16), nil, true)
		if err != nil {
			return nil, fmt.Errorf("TSS server failure: %v", err)
		}
	}

	var signedTx types.Transaction
	err = rlp.DecodeBytes(ethcom.FromHex(signedRawTx.RawTransaction), &signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TSS signed result: %v", err)
	}

	return &signedTx, nil
}

func queryDeployedBEP20ContractAddr(erc20Addr ethcom.Address, bscSwapAgentAddr ethcom.Address, txRecipient *types.Receipt, bscClient *ethclient.Client) (ethcom.Address, error) {
	swapAgentInstance, err := contractabi.NewBSCSwapAgent(bscSwapAgentAddr, bscClient)
	if err != nil {
		return ethcom.Address{}, err
	}
	if len(txRecipient.Logs) != 2 {
		return ethcom.Address{}, fmt.Errorf("Expected tx logs length in recipient is 2, actual it is %d", len(txRecipient.Logs))
	}
	createSwapEvent, err := swapAgentInstance.ParseSwapPairCreated(*txRecipient.Logs[1])
	if err != nil || createSwapEvent == nil {
		return ethcom.Address{}, err
	}

	util.Logger.Debugf("Deployed bep20 contact %s for register erc20 %s", createSwapEvent.Bep20Addr.String(), erc20Addr.String())
	return createSwapEvent.Bep20Addr, nil
}
