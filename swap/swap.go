package swap

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/binance-chain/bsc-eth-swap/swap/erc20"
	"math/big"
	"time"

	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-uuid"
	"github.com/jinzhu/gorm"

	"github.com/binance-chain/bsc-eth-swap/common"
	"github.com/binance-chain/bsc-eth-swap/model"
	"github.com/binance-chain/bsc-eth-swap/util"
)

type TokenInstance struct {
	Symbol     string
	Name       string
	LowBound   string
	UpperBound string

	BSCPrivateKey   *ecdsa.PrivateKey
	BSCTxSender     ethcom.Address
	BSCContractAddr ethcom.Address
	ETHPrivateKey   *ecdsa.PrivateKey
	ETHContractAddr ethcom.Address
	ETHTxSender     ethcom.Address
}

type Swapper struct {
	DB                      *gorm.DB
	Config                  *util.Config
	TokenInstances          map[string]*TokenInstance
	ETHClient               *ethclient.Client
	BSCClient               *ethclient.Client
	BSCContractAddrToSymbol map[string]string
	ETHContractAddrToSymbol map[string]string
}

// NewSwapper returns the Swapper instance
func NewSwapper(db *gorm.DB, cfg *util.Config, bscClient, ethClient *ethclient.Client) (*Swapper, error) {
	tokens := make([]model.Token, 0)
	db.Find(&tokens)

	tokenInstances, err := buildTokenInstance(tokens)
	if err != nil {
		return nil, err
	}
	bscContractAddrToSymbol := make(map[string]string)
	ethContractAddrToSymbol := make(map[string]string)
	for _, token := range tokens {
		bscContractAddrToSymbol[token.BSCContractAddr] = token.Symbol
		ethContractAddrToSymbol[token.ETHContractAddr] = token.Symbol
	}

	return &Swapper{
		DB:                      db,
		Config:                  cfg,
		BSCClient:               bscClient,
		ETHClient:               ethClient,
		TokenInstances:          tokenInstances,
		BSCContractAddrToSymbol: bscContractAddrToSymbol,
		ETHContractAddrToSymbol: ethContractAddrToSymbol,
	}, nil
}

func (swapper *Swapper) Start() {
	go swapper.handleSwapDaemon()
	go swapper.sendTokenDaemon()
	go swapper.trackSwapDaemon()
	go swapper.alertDaemon()
}

func (swapper *Swapper) handleSwapDaemon() {
	for {
		txEventLogs := make([]model.TxEventLog, BatchSize)
		swapper.DB.Where("phase = ?", model.SeenSwapRequest).
			Order("height asc").Limit(BatchSize).Find(&txEventLogs)

		if len(txEventLogs) == 0 {
			time.Sleep(SleepSecond * time.Second)
			continue
		}

		util.Logger.Infof("found swap tx log")
		for _, txEventLog := range txEventLogs {
			var symbol string
			var ok bool
			if txEventLog.Chain == common.ChainETH {
				symbol, ok = swapper.ETHContractAddrToSymbol[txEventLog.ContractAddress]
				if !ok {
					// log and mark the swap is failed because the token is not supported yet
					continue
				}
			} else {
				symbol, ok = swapper.BSCContractAddrToSymbol[txEventLog.ContractAddress]
				if !ok {
					// log and mark the swap is failed because the token is not supported yet
					continue
				}
			}

			tokenInstance, ok := swapper.TokenInstances[symbol]
			if !ok {
				// log and mark the swap is failed for missing private key
				continue
			}

			generatedUUID, err := swapper.generateUUID()
			if err != nil {
				// log and mark the swap is failed for missing private key
				continue
			}

			swapStatus := SwapTokenReceived
			if txEventLog.Amount < tokenInstance.LowBound || txEventLog.Amount > tokenInstance.UpperBound {
				swapStatus = SwapQuoteRejected
			}
			swapDirection := SwapEth2BSC
			if txEventLog.Chain == common.ChainBSC {
				swapDirection = SwapBSC2Eth
			}
			swap := &model.Swap{
				UUID:           generatedUUID,
				Status:         swapStatus,
				Sponsor:        txEventLog.FromAddress,
				Symbol:         symbol,
				Amount:         txEventLog.Amount,
				Direction:      swapDirection,
				DepositTxHash:  txEventLog.TxHash,
				WithdrawTxHash: "",
				Log:            "",
			}

			err = swapper.insertSwapToDB(swap)
			if err != nil {
				// log and mark the swap is failed for missing private key
			}

			// mark tx_event_log as processed
			err = swapper.DB.Model(model.TxEventLog{}).Where("tx_hash = ?", swap.DepositTxHash).Updates(
				map[string]interface{}{
					"phase":      model.ConfirmSwapRequest,
					"update_time": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update tx_event_log table failed: %s", err.Error())
			}
		}
	}
}

func (swapper *Swapper) sendTokenDaemon() {
	for {
		txEventLogs := make([]model.TxEventLog, BatchSize)
		swapper.DB.Where("status = ? and phase = ?", model.TxStatusConfirmed, model.ConfirmSwapRequest).
			Order("height asc").Limit(BatchSize).Find(&txEventLogs)

		if len(txEventLogs) == 0 {
			time.Sleep(SleepSecond * time.Second)
			continue
		}

		util.Logger.Infof("found confirmed swap tx log")

		for _, txEventLog := range txEventLogs {
			swap := model.Swap{}
			swapper.DB.Where("deposit_tx_hash = ?", txEventLog.TxHash).First(&swap)
			if swap.UUID == "" {
				util.Logger.Errorf("unexpected error: found empty swap")
				continue
			}

			err := swapper.doSwap(&swap)
			if err != nil {
				err = swapper.DB.Model(model.Swap{}).Where("uuid = ?", swap.UUID).Updates(
					map[string]interface{}{
						"status":      SwapSendFailed,
						"log":         err.Error(),
						"updated_at": time.Now().Unix(),
					}).Error
				if err != nil {
					util.Logger.Errorf("update swap table failed: %s", err.Error())
				}
			} else {
				err = swapper.DB.Model(model.Swap{}).Where("uuid = ?", swap.UUID).Updates(
					map[string]interface{}{
						"status":      SwapSent,
						"updated_at": time.Now().Unix(),
					}).Error
				if err != nil {
					util.Logger.Errorf("update swap table failed: %s", err.Error())
				}
			}

			// mark tx_event_log as processed
			err = swapper.DB.Model(model.TxEventLog{}).Where("tx_hash = ?", swap.DepositTxHash).Updates(
				map[string]interface{}{
					"phase":       model.AckSwapRequest,
					"update_time": time.Now().Unix(),
				}).Error
			if err != nil {
				util.Logger.Errorf("update tx_event_log table failed: %s", err.Error())
			}
		}
	}
}
func (swapper *Swapper) trackSwapDaemon() {

}

func (swapper *Swapper) alertDaemon() {

}

func (swapper *Swapper) doSwap(swap *model.Swap) error {
	tokenInstance, ok := swapper.TokenInstances[swap.Symbol]
	if !ok {
		return fmt.Errorf("unsupported token %s", swap.Symbol)
	}
	amount := big.NewInt(0)
	_, ok = amount.SetString(swap.Amount, 10)
	if !ok {
		return fmt.Errorf("invalid swap amount: %s", swap.Amount)
	}
	//txInput, err := abiEncodeTransfer(ethcom.HexToAddress(swap.Sponsor), amount)
	//if err != nil {
	//	return err
	//}

	if swap.Direction == SwapEth2BSC {
		erc20Instance, err := erc20.NewErc20(tokenInstance.BSCContractAddr, swapper.BSCClient)
		if err != nil {
			return err
		}
		tx, err := erc20Instance.Transfer(generateTxOpt(tokenInstance.BSCPrivateKey), ethcom.HexToAddress(swap.Sponsor), amount)
		if err != nil {
			return err
		}
		swapTx := &model.SwapTx{
			SourceChain:       common.ChainETH,
			SwapRequestTxHash: swap.DepositTxHash,
			Symbol:            swap.Symbol,
			Amount:            swap.Amount,

			DestiChain:         common.ChainBSC,
			DestiAssetContract: tokenInstance.BSCContractAddr.String(),
			TxHash:             tx.Hash().String(),
			Status:             model.TxStatusInit,
		}
		err = swapper.insertSwapTxToDB(swapTx)
		if err != nil {
			return err
		}
		//err = swapper.BSCClient.SendTransaction(context.Background(), signedTx)
		//if err != nil {
		//	return err
		//}
	} else {
		erc20Instance, err := erc20.NewErc20(tokenInstance.ETHContractAddr, swapper.ETHClient)
		if err != nil {
			return err
		}
		tx, err := erc20Instance.Transfer(generateTxOpt(tokenInstance.ETHPrivateKey), ethcom.HexToAddress(swap.Sponsor), amount)
		if err != nil {
			return err
		}
		swapTx := &model.SwapTx{
			SourceChain:       common.ChainBSC,
			SwapRequestTxHash: swap.DepositTxHash,
			Symbol:            swap.Symbol,
			Amount:            swap.Amount,

			DestiChain:         common.ChainETH,
			DestiAssetContract: tokenInstance.BSCContractAddr.String(),
			TxHash:             tx.Hash().String(),
			Status:             model.TxStatusInit,
		}
		err = swapper.insertSwapTxToDB(swapTx)
		if err != nil {
			return err
		}
		//err = swapper.ETHClient.SendTransaction(context.Background(), t)
		//if err != nil {
		//	return err
		//}
	}
	return nil
}

func (swapper *Swapper) insertSwapToDB(data *model.Swap) error {
	tx := swapper.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(data).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (swapper *Swapper) insertSwapTxToDB(data *model.SwapTx) error {
	tx := swapper.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(data).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (swapper *Swapper) getSwapByUUID(uuid string) model.Swap {
	swap := model.Swap{}
	swapper.DB.Where("uuid = ?", uuid).First(&swap)
	return swap
}

func (swapper *Swapper) generateUUID() (string, error) {
	for idx := 0; idx < MaxTryGenerateUUID; idx++ {
		id, err := uuid.GenerateUUID()
		if err != nil {
			return "", err
		}

		swap := swapper.getSwapByUUID(id)
		if swap.UUID != "" {
			continue
		} else {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique uuid")
}
