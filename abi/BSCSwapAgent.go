// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abi

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// BSCSwapAgentABI is the input ABI used to generate the binding from.
const BSCSwapAgentABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bep20Addr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"ethTxHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"toAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"SwapFilled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"ethRegisterTxHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bep20Addr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"decimals\",\"type\":\"uint8\"}],\"name\":\"SwapPairCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bep20Addr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"fromAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"SwapStarted\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"bep20Implementation\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bep20ProxyAdmin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"filledETHTx\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"swapFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"swapMappingBSC2ETH\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"swapMappingETH2BSC\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bep20Impl\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"addresspayable\",\"name\":\"ownerAddr\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"bep20ProxyAdminAddr\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"setSwapFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"ethTxHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint8\",\"name\":\"decimals\",\"type\":\"uint8\"}],\"name\":\"createSwapPair\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"ethTxHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"toAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"fillETH2BSCSwap\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bep20Addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"swapBSC2ETH\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"}]"

// BSCSwapAgent is an auto generated Go binding around an Ethereum contract.
type BSCSwapAgent struct {
	BSCSwapAgentCaller     // Read-only binding to the contract
	BSCSwapAgentTransactor // Write-only binding to the contract
	BSCSwapAgentFilterer   // Log filterer for contract events
}

// BSCSwapAgentCaller is an auto generated read-only Go binding around an Ethereum contract.
type BSCSwapAgentCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BSCSwapAgentTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BSCSwapAgentTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BSCSwapAgentFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BSCSwapAgentFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BSCSwapAgentSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BSCSwapAgentSession struct {
	Contract     *BSCSwapAgent     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BSCSwapAgentCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BSCSwapAgentCallerSession struct {
	Contract *BSCSwapAgentCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// BSCSwapAgentTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BSCSwapAgentTransactorSession struct {
	Contract     *BSCSwapAgentTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// BSCSwapAgentRaw is an auto generated low-level Go binding around an Ethereum contract.
type BSCSwapAgentRaw struct {
	Contract *BSCSwapAgent // Generic contract binding to access the raw methods on
}

// BSCSwapAgentCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BSCSwapAgentCallerRaw struct {
	Contract *BSCSwapAgentCaller // Generic read-only contract binding to access the raw methods on
}

// BSCSwapAgentTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BSCSwapAgentTransactorRaw struct {
	Contract *BSCSwapAgentTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBSCSwapAgent creates a new instance of BSCSwapAgent, bound to a specific deployed contract.
func NewBSCSwapAgent(address common.Address, backend bind.ContractBackend) (*BSCSwapAgent, error) {
	contract, err := bindBSCSwapAgent(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BSCSwapAgent{BSCSwapAgentCaller: BSCSwapAgentCaller{contract: contract}, BSCSwapAgentTransactor: BSCSwapAgentTransactor{contract: contract}, BSCSwapAgentFilterer: BSCSwapAgentFilterer{contract: contract}}, nil
}

// NewBSCSwapAgentCaller creates a new read-only instance of BSCSwapAgent, bound to a specific deployed contract.
func NewBSCSwapAgentCaller(address common.Address, caller bind.ContractCaller) (*BSCSwapAgentCaller, error) {
	contract, err := bindBSCSwapAgent(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BSCSwapAgentCaller{contract: contract}, nil
}

// NewBSCSwapAgentTransactor creates a new write-only instance of BSCSwapAgent, bound to a specific deployed contract.
func NewBSCSwapAgentTransactor(address common.Address, transactor bind.ContractTransactor) (*BSCSwapAgentTransactor, error) {
	contract, err := bindBSCSwapAgent(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BSCSwapAgentTransactor{contract: contract}, nil
}

// NewBSCSwapAgentFilterer creates a new log filterer instance of BSCSwapAgent, bound to a specific deployed contract.
func NewBSCSwapAgentFilterer(address common.Address, filterer bind.ContractFilterer) (*BSCSwapAgentFilterer, error) {
	contract, err := bindBSCSwapAgent(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BSCSwapAgentFilterer{contract: contract}, nil
}

// bindBSCSwapAgent binds a generic wrapper to an already deployed contract.
func bindBSCSwapAgent(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BSCSwapAgentABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BSCSwapAgent *BSCSwapAgentRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _BSCSwapAgent.Contract.BSCSwapAgentCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BSCSwapAgent *BSCSwapAgentRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.BSCSwapAgentTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BSCSwapAgent *BSCSwapAgentRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.BSCSwapAgentTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BSCSwapAgent *BSCSwapAgentCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _BSCSwapAgent.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BSCSwapAgent *BSCSwapAgentTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BSCSwapAgent *BSCSwapAgentTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.contract.Transact(opts, method, params...)
}

// Bep20Implementation is a free data retrieval call binding the contract method 0x66fec65c.
//
// Solidity: function bep20Implementation() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCaller) Bep20Implementation(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _BSCSwapAgent.contract.Call(opts, out, "bep20Implementation")
	return *ret0, err
}

// Bep20Implementation is a free data retrieval call binding the contract method 0x66fec65c.
//
// Solidity: function bep20Implementation() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentSession) Bep20Implementation() (common.Address, error) {
	return _BSCSwapAgent.Contract.Bep20Implementation(&_BSCSwapAgent.CallOpts)
}

// Bep20Implementation is a free data retrieval call binding the contract method 0x66fec65c.
//
// Solidity: function bep20Implementation() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCallerSession) Bep20Implementation() (common.Address, error) {
	return _BSCSwapAgent.Contract.Bep20Implementation(&_BSCSwapAgent.CallOpts)
}

// Bep20ProxyAdmin is a free data retrieval call binding the contract method 0x0344165a.
//
// Solidity: function bep20ProxyAdmin() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCaller) Bep20ProxyAdmin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _BSCSwapAgent.contract.Call(opts, out, "bep20ProxyAdmin")
	return *ret0, err
}

// Bep20ProxyAdmin is a free data retrieval call binding the contract method 0x0344165a.
//
// Solidity: function bep20ProxyAdmin() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentSession) Bep20ProxyAdmin() (common.Address, error) {
	return _BSCSwapAgent.Contract.Bep20ProxyAdmin(&_BSCSwapAgent.CallOpts)
}

// Bep20ProxyAdmin is a free data retrieval call binding the contract method 0x0344165a.
//
// Solidity: function bep20ProxyAdmin() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCallerSession) Bep20ProxyAdmin() (common.Address, error) {
	return _BSCSwapAgent.Contract.Bep20ProxyAdmin(&_BSCSwapAgent.CallOpts)
}

// FilledETHTx is a free data retrieval call binding the contract method 0x4e2dc7f1.
//
// Solidity: function filledETHTx(bytes32 ) constant returns(bool)
func (_BSCSwapAgent *BSCSwapAgentCaller) FilledETHTx(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _BSCSwapAgent.contract.Call(opts, out, "filledETHTx", arg0)
	return *ret0, err
}

// FilledETHTx is a free data retrieval call binding the contract method 0x4e2dc7f1.
//
// Solidity: function filledETHTx(bytes32 ) constant returns(bool)
func (_BSCSwapAgent *BSCSwapAgentSession) FilledETHTx(arg0 [32]byte) (bool, error) {
	return _BSCSwapAgent.Contract.FilledETHTx(&_BSCSwapAgent.CallOpts, arg0)
}

// FilledETHTx is a free data retrieval call binding the contract method 0x4e2dc7f1.
//
// Solidity: function filledETHTx(bytes32 ) constant returns(bool)
func (_BSCSwapAgent *BSCSwapAgentCallerSession) FilledETHTx(arg0 [32]byte) (bool, error) {
	return _BSCSwapAgent.Contract.FilledETHTx(&_BSCSwapAgent.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _BSCSwapAgent.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentSession) Owner() (common.Address, error) {
	return _BSCSwapAgent.Contract.Owner(&_BSCSwapAgent.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCallerSession) Owner() (common.Address, error) {
	return _BSCSwapAgent.Contract.Owner(&_BSCSwapAgent.CallOpts)
}

// SwapFee is a free data retrieval call binding the contract method 0x54cf2aeb.
//
// Solidity: function swapFee() constant returns(uint256)
func (_BSCSwapAgent *BSCSwapAgentCaller) SwapFee(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _BSCSwapAgent.contract.Call(opts, out, "swapFee")
	return *ret0, err
}

// SwapFee is a free data retrieval call binding the contract method 0x54cf2aeb.
//
// Solidity: function swapFee() constant returns(uint256)
func (_BSCSwapAgent *BSCSwapAgentSession) SwapFee() (*big.Int, error) {
	return _BSCSwapAgent.Contract.SwapFee(&_BSCSwapAgent.CallOpts)
}

// SwapFee is a free data retrieval call binding the contract method 0x54cf2aeb.
//
// Solidity: function swapFee() constant returns(uint256)
func (_BSCSwapAgent *BSCSwapAgentCallerSession) SwapFee() (*big.Int, error) {
	return _BSCSwapAgent.Contract.SwapFee(&_BSCSwapAgent.CallOpts)
}

// SwapMappingBSC2ETH is a free data retrieval call binding the contract method 0xbe0ace69.
//
// Solidity: function swapMappingBSC2ETH(address ) constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCaller) SwapMappingBSC2ETH(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _BSCSwapAgent.contract.Call(opts, out, "swapMappingBSC2ETH", arg0)
	return *ret0, err
}

// SwapMappingBSC2ETH is a free data retrieval call binding the contract method 0xbe0ace69.
//
// Solidity: function swapMappingBSC2ETH(address ) constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentSession) SwapMappingBSC2ETH(arg0 common.Address) (common.Address, error) {
	return _BSCSwapAgent.Contract.SwapMappingBSC2ETH(&_BSCSwapAgent.CallOpts, arg0)
}

// SwapMappingBSC2ETH is a free data retrieval call binding the contract method 0xbe0ace69.
//
// Solidity: function swapMappingBSC2ETH(address ) constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCallerSession) SwapMappingBSC2ETH(arg0 common.Address) (common.Address, error) {
	return _BSCSwapAgent.Contract.SwapMappingBSC2ETH(&_BSCSwapAgent.CallOpts, arg0)
}

// SwapMappingETH2BSC is a free data retrieval call binding the contract method 0x60b810f1.
//
// Solidity: function swapMappingETH2BSC(address ) constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCaller) SwapMappingETH2BSC(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _BSCSwapAgent.contract.Call(opts, out, "swapMappingETH2BSC", arg0)
	return *ret0, err
}

// SwapMappingETH2BSC is a free data retrieval call binding the contract method 0x60b810f1.
//
// Solidity: function swapMappingETH2BSC(address ) constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentSession) SwapMappingETH2BSC(arg0 common.Address) (common.Address, error) {
	return _BSCSwapAgent.Contract.SwapMappingETH2BSC(&_BSCSwapAgent.CallOpts, arg0)
}

// SwapMappingETH2BSC is a free data retrieval call binding the contract method 0x60b810f1.
//
// Solidity: function swapMappingETH2BSC(address ) constant returns(address)
func (_BSCSwapAgent *BSCSwapAgentCallerSession) SwapMappingETH2BSC(arg0 common.Address) (common.Address, error) {
	return _BSCSwapAgent.Contract.SwapMappingETH2BSC(&_BSCSwapAgent.CallOpts, arg0)
}

// CreateSwapPair is a paid mutator transaction binding the contract method 0x32bd6e31.
//
// Solidity: function createSwapPair(bytes32 ethTxHash, address erc20Addr, string name, string symbol, uint8 decimals) returns(address)
func (_BSCSwapAgent *BSCSwapAgentTransactor) CreateSwapPair(opts *bind.TransactOpts, ethTxHash [32]byte, erc20Addr common.Address, name string, symbol string, decimals uint8) (*types.Transaction, error) {
	return _BSCSwapAgent.contract.Transact(opts, "createSwapPair", ethTxHash, erc20Addr, name, symbol, decimals)
}

// CreateSwapPair is a paid mutator transaction binding the contract method 0x32bd6e31.
//
// Solidity: function createSwapPair(bytes32 ethTxHash, address erc20Addr, string name, string symbol, uint8 decimals) returns(address)
func (_BSCSwapAgent *BSCSwapAgentSession) CreateSwapPair(ethTxHash [32]byte, erc20Addr common.Address, name string, symbol string, decimals uint8) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.CreateSwapPair(&_BSCSwapAgent.TransactOpts, ethTxHash, erc20Addr, name, symbol, decimals)
}

// CreateSwapPair is a paid mutator transaction binding the contract method 0x32bd6e31.
//
// Solidity: function createSwapPair(bytes32 ethTxHash, address erc20Addr, string name, string symbol, uint8 decimals) returns(address)
func (_BSCSwapAgent *BSCSwapAgentTransactorSession) CreateSwapPair(ethTxHash [32]byte, erc20Addr common.Address, name string, symbol string, decimals uint8) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.CreateSwapPair(&_BSCSwapAgent.TransactOpts, ethTxHash, erc20Addr, name, symbol, decimals)
}

// FillETH2BSCSwap is a paid mutator transaction binding the contract method 0xe307b931.
//
// Solidity: function fillETH2BSCSwap(bytes32 ethTxHash, address erc20Addr, address toAddress, uint256 amount) returns(bool)
func (_BSCSwapAgent *BSCSwapAgentTransactor) FillETH2BSCSwap(opts *bind.TransactOpts, ethTxHash [32]byte, erc20Addr common.Address, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.contract.Transact(opts, "fillETH2BSCSwap", ethTxHash, erc20Addr, toAddress, amount)
}

// FillETH2BSCSwap is a paid mutator transaction binding the contract method 0xe307b931.
//
// Solidity: function fillETH2BSCSwap(bytes32 ethTxHash, address erc20Addr, address toAddress, uint256 amount) returns(bool)
func (_BSCSwapAgent *BSCSwapAgentSession) FillETH2BSCSwap(ethTxHash [32]byte, erc20Addr common.Address, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.FillETH2BSCSwap(&_BSCSwapAgent.TransactOpts, ethTxHash, erc20Addr, toAddress, amount)
}

// FillETH2BSCSwap is a paid mutator transaction binding the contract method 0xe307b931.
//
// Solidity: function fillETH2BSCSwap(bytes32 ethTxHash, address erc20Addr, address toAddress, uint256 amount) returns(bool)
func (_BSCSwapAgent *BSCSwapAgentTransactorSession) FillETH2BSCSwap(ethTxHash [32]byte, erc20Addr common.Address, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.FillETH2BSCSwap(&_BSCSwapAgent.TransactOpts, ethTxHash, erc20Addr, toAddress, amount)
}

// Initialize is a paid mutator transaction binding the contract method 0x358394d8.
//
// Solidity: function initialize(address bep20Impl, uint256 fee, address ownerAddr, address bep20ProxyAdminAddr) returns()
func (_BSCSwapAgent *BSCSwapAgentTransactor) Initialize(opts *bind.TransactOpts, bep20Impl common.Address, fee *big.Int, ownerAddr common.Address, bep20ProxyAdminAddr common.Address) (*types.Transaction, error) {
	return _BSCSwapAgent.contract.Transact(opts, "initialize", bep20Impl, fee, ownerAddr, bep20ProxyAdminAddr)
}

// Initialize is a paid mutator transaction binding the contract method 0x358394d8.
//
// Solidity: function initialize(address bep20Impl, uint256 fee, address ownerAddr, address bep20ProxyAdminAddr) returns()
func (_BSCSwapAgent *BSCSwapAgentSession) Initialize(bep20Impl common.Address, fee *big.Int, ownerAddr common.Address, bep20ProxyAdminAddr common.Address) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.Initialize(&_BSCSwapAgent.TransactOpts, bep20Impl, fee, ownerAddr, bep20ProxyAdminAddr)
}

// Initialize is a paid mutator transaction binding the contract method 0x358394d8.
//
// Solidity: function initialize(address bep20Impl, uint256 fee, address ownerAddr, address bep20ProxyAdminAddr) returns()
func (_BSCSwapAgent *BSCSwapAgentTransactorSession) Initialize(bep20Impl common.Address, fee *big.Int, ownerAddr common.Address, bep20ProxyAdminAddr common.Address) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.Initialize(&_BSCSwapAgent.TransactOpts, bep20Impl, fee, ownerAddr, bep20ProxyAdminAddr)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BSCSwapAgent *BSCSwapAgentTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BSCSwapAgent.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BSCSwapAgent *BSCSwapAgentSession) RenounceOwnership() (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.RenounceOwnership(&_BSCSwapAgent.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BSCSwapAgent *BSCSwapAgentTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.RenounceOwnership(&_BSCSwapAgent.TransactOpts)
}

// SetSwapFee is a paid mutator transaction binding the contract method 0x34e19907.
//
// Solidity: function setSwapFee(uint256 fee) returns()
func (_BSCSwapAgent *BSCSwapAgentTransactor) SetSwapFee(opts *bind.TransactOpts, fee *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.contract.Transact(opts, "setSwapFee", fee)
}

// SetSwapFee is a paid mutator transaction binding the contract method 0x34e19907.
//
// Solidity: function setSwapFee(uint256 fee) returns()
func (_BSCSwapAgent *BSCSwapAgentSession) SetSwapFee(fee *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.SetSwapFee(&_BSCSwapAgent.TransactOpts, fee)
}

// SetSwapFee is a paid mutator transaction binding the contract method 0x34e19907.
//
// Solidity: function setSwapFee(uint256 fee) returns()
func (_BSCSwapAgent *BSCSwapAgentTransactorSession) SetSwapFee(fee *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.SetSwapFee(&_BSCSwapAgent.TransactOpts, fee)
}

// SwapBSC2ETH is a paid mutator transaction binding the contract method 0x1ba3b150.
//
// Solidity: function swapBSC2ETH(address bep20Addr, uint256 amount) returns(bool)
func (_BSCSwapAgent *BSCSwapAgentTransactor) SwapBSC2ETH(opts *bind.TransactOpts, bep20Addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.contract.Transact(opts, "swapBSC2ETH", bep20Addr, amount)
}

// SwapBSC2ETH is a paid mutator transaction binding the contract method 0x1ba3b150.
//
// Solidity: function swapBSC2ETH(address bep20Addr, uint256 amount) returns(bool)
func (_BSCSwapAgent *BSCSwapAgentSession) SwapBSC2ETH(bep20Addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.SwapBSC2ETH(&_BSCSwapAgent.TransactOpts, bep20Addr, amount)
}

// SwapBSC2ETH is a paid mutator transaction binding the contract method 0x1ba3b150.
//
// Solidity: function swapBSC2ETH(address bep20Addr, uint256 amount) returns(bool)
func (_BSCSwapAgent *BSCSwapAgentTransactorSession) SwapBSC2ETH(bep20Addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.SwapBSC2ETH(&_BSCSwapAgent.TransactOpts, bep20Addr, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BSCSwapAgent *BSCSwapAgentTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BSCSwapAgent.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BSCSwapAgent *BSCSwapAgentSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.TransferOwnership(&_BSCSwapAgent.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BSCSwapAgent *BSCSwapAgentTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BSCSwapAgent.Contract.TransferOwnership(&_BSCSwapAgent.TransactOpts, newOwner)
}

// BSCSwapAgentOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BSCSwapAgent contract.
type BSCSwapAgentOwnershipTransferredIterator struct {
	Event *BSCSwapAgentOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BSCSwapAgentOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BSCSwapAgentOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BSCSwapAgentOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BSCSwapAgentOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BSCSwapAgentOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BSCSwapAgentOwnershipTransferred represents a OwnershipTransferred event raised by the BSCSwapAgent contract.
type BSCSwapAgentOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BSCSwapAgent *BSCSwapAgentFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BSCSwapAgentOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BSCSwapAgent.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BSCSwapAgentOwnershipTransferredIterator{contract: _BSCSwapAgent.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BSCSwapAgent *BSCSwapAgentFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BSCSwapAgentOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BSCSwapAgent.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BSCSwapAgentOwnershipTransferred)
				if err := _BSCSwapAgent.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BSCSwapAgent *BSCSwapAgentFilterer) ParseOwnershipTransferred(log types.Log) (*BSCSwapAgentOwnershipTransferred, error) {
	event := new(BSCSwapAgentOwnershipTransferred)
	if err := _BSCSwapAgent.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	return event, nil
}

// BSCSwapAgentSwapFilledIterator is returned from FilterSwapFilled and is used to iterate over the raw logs and unpacked data for SwapFilled events raised by the BSCSwapAgent contract.
type BSCSwapAgentSwapFilledIterator struct {
	Event *BSCSwapAgentSwapFilled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BSCSwapAgentSwapFilledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BSCSwapAgentSwapFilled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BSCSwapAgentSwapFilled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BSCSwapAgentSwapFilledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BSCSwapAgentSwapFilledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BSCSwapAgentSwapFilled represents a SwapFilled event raised by the BSCSwapAgent contract.
type BSCSwapAgentSwapFilled struct {
	Bep20Addr common.Address
	EthTxHash [32]byte
	ToAddress common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSwapFilled is a free log retrieval operation binding the contract event 0x3bebd9a738291e69898b5dbfadb6329b4b09fc648bdef68762928e521463abd9.
//
// Solidity: event SwapFilled(address indexed bep20Addr, bytes32 indexed ethTxHash, address indexed toAddress, uint256 amount)
func (_BSCSwapAgent *BSCSwapAgentFilterer) FilterSwapFilled(opts *bind.FilterOpts, bep20Addr []common.Address, ethTxHash [][32]byte, toAddress []common.Address) (*BSCSwapAgentSwapFilledIterator, error) {

	var bep20AddrRule []interface{}
	for _, bep20AddrItem := range bep20Addr {
		bep20AddrRule = append(bep20AddrRule, bep20AddrItem)
	}
	var ethTxHashRule []interface{}
	for _, ethTxHashItem := range ethTxHash {
		ethTxHashRule = append(ethTxHashRule, ethTxHashItem)
	}
	var toAddressRule []interface{}
	for _, toAddressItem := range toAddress {
		toAddressRule = append(toAddressRule, toAddressItem)
	}

	logs, sub, err := _BSCSwapAgent.contract.FilterLogs(opts, "SwapFilled", bep20AddrRule, ethTxHashRule, toAddressRule)
	if err != nil {
		return nil, err
	}
	return &BSCSwapAgentSwapFilledIterator{contract: _BSCSwapAgent.contract, event: "SwapFilled", logs: logs, sub: sub}, nil
}

// WatchSwapFilled is a free log subscription operation binding the contract event 0x3bebd9a738291e69898b5dbfadb6329b4b09fc648bdef68762928e521463abd9.
//
// Solidity: event SwapFilled(address indexed bep20Addr, bytes32 indexed ethTxHash, address indexed toAddress, uint256 amount)
func (_BSCSwapAgent *BSCSwapAgentFilterer) WatchSwapFilled(opts *bind.WatchOpts, sink chan<- *BSCSwapAgentSwapFilled, bep20Addr []common.Address, ethTxHash [][32]byte, toAddress []common.Address) (event.Subscription, error) {

	var bep20AddrRule []interface{}
	for _, bep20AddrItem := range bep20Addr {
		bep20AddrRule = append(bep20AddrRule, bep20AddrItem)
	}
	var ethTxHashRule []interface{}
	for _, ethTxHashItem := range ethTxHash {
		ethTxHashRule = append(ethTxHashRule, ethTxHashItem)
	}
	var toAddressRule []interface{}
	for _, toAddressItem := range toAddress {
		toAddressRule = append(toAddressRule, toAddressItem)
	}

	logs, sub, err := _BSCSwapAgent.contract.WatchLogs(opts, "SwapFilled", bep20AddrRule, ethTxHashRule, toAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BSCSwapAgentSwapFilled)
				if err := _BSCSwapAgent.contract.UnpackLog(event, "SwapFilled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSwapFilled is a log parse operation binding the contract event 0x3bebd9a738291e69898b5dbfadb6329b4b09fc648bdef68762928e521463abd9.
//
// Solidity: event SwapFilled(address indexed bep20Addr, bytes32 indexed ethTxHash, address indexed toAddress, uint256 amount)
func (_BSCSwapAgent *BSCSwapAgentFilterer) ParseSwapFilled(log types.Log) (*BSCSwapAgentSwapFilled, error) {
	event := new(BSCSwapAgentSwapFilled)
	if err := _BSCSwapAgent.contract.UnpackLog(event, "SwapFilled", log); err != nil {
		return nil, err
	}
	return event, nil
}

// BSCSwapAgentSwapPairCreatedIterator is returned from FilterSwapPairCreated and is used to iterate over the raw logs and unpacked data for SwapPairCreated events raised by the BSCSwapAgent contract.
type BSCSwapAgentSwapPairCreatedIterator struct {
	Event *BSCSwapAgentSwapPairCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BSCSwapAgentSwapPairCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BSCSwapAgentSwapPairCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BSCSwapAgentSwapPairCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BSCSwapAgentSwapPairCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BSCSwapAgentSwapPairCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BSCSwapAgentSwapPairCreated represents a SwapPairCreated event raised by the BSCSwapAgent contract.
type BSCSwapAgentSwapPairCreated struct {
	EthRegisterTxHash [32]byte
	Bep20Addr         common.Address
	Erc20Addr         common.Address
	Symbol            string
	Name              string
	Decimals          uint8
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterSwapPairCreated is a free log retrieval operation binding the contract event 0xcc0314763eabceb74cd3d30ae785c09bfe4e204af2088b3bfcdbbe5082133db5.
//
// Solidity: event SwapPairCreated(bytes32 indexed ethRegisterTxHash, address indexed bep20Addr, address indexed erc20Addr, string symbol, string name, uint8 decimals)
func (_BSCSwapAgent *BSCSwapAgentFilterer) FilterSwapPairCreated(opts *bind.FilterOpts, ethRegisterTxHash [][32]byte, bep20Addr []common.Address, erc20Addr []common.Address) (*BSCSwapAgentSwapPairCreatedIterator, error) {

	var ethRegisterTxHashRule []interface{}
	for _, ethRegisterTxHashItem := range ethRegisterTxHash {
		ethRegisterTxHashRule = append(ethRegisterTxHashRule, ethRegisterTxHashItem)
	}
	var bep20AddrRule []interface{}
	for _, bep20AddrItem := range bep20Addr {
		bep20AddrRule = append(bep20AddrRule, bep20AddrItem)
	}
	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}

	logs, sub, err := _BSCSwapAgent.contract.FilterLogs(opts, "SwapPairCreated", ethRegisterTxHashRule, bep20AddrRule, erc20AddrRule)
	if err != nil {
		return nil, err
	}
	return &BSCSwapAgentSwapPairCreatedIterator{contract: _BSCSwapAgent.contract, event: "SwapPairCreated", logs: logs, sub: sub}, nil
}

// WatchSwapPairCreated is a free log subscription operation binding the contract event 0xcc0314763eabceb74cd3d30ae785c09bfe4e204af2088b3bfcdbbe5082133db5.
//
// Solidity: event SwapPairCreated(bytes32 indexed ethRegisterTxHash, address indexed bep20Addr, address indexed erc20Addr, string symbol, string name, uint8 decimals)
func (_BSCSwapAgent *BSCSwapAgentFilterer) WatchSwapPairCreated(opts *bind.WatchOpts, sink chan<- *BSCSwapAgentSwapPairCreated, ethRegisterTxHash [][32]byte, bep20Addr []common.Address, erc20Addr []common.Address) (event.Subscription, error) {

	var ethRegisterTxHashRule []interface{}
	for _, ethRegisterTxHashItem := range ethRegisterTxHash {
		ethRegisterTxHashRule = append(ethRegisterTxHashRule, ethRegisterTxHashItem)
	}
	var bep20AddrRule []interface{}
	for _, bep20AddrItem := range bep20Addr {
		bep20AddrRule = append(bep20AddrRule, bep20AddrItem)
	}
	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}

	logs, sub, err := _BSCSwapAgent.contract.WatchLogs(opts, "SwapPairCreated", ethRegisterTxHashRule, bep20AddrRule, erc20AddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BSCSwapAgentSwapPairCreated)
				if err := _BSCSwapAgent.contract.UnpackLog(event, "SwapPairCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSwapPairCreated is a log parse operation binding the contract event 0xcc0314763eabceb74cd3d30ae785c09bfe4e204af2088b3bfcdbbe5082133db5.
//
// Solidity: event SwapPairCreated(bytes32 indexed ethRegisterTxHash, address indexed bep20Addr, address indexed erc20Addr, string symbol, string name, uint8 decimals)
func (_BSCSwapAgent *BSCSwapAgentFilterer) ParseSwapPairCreated(log types.Log) (*BSCSwapAgentSwapPairCreated, error) {
	event := new(BSCSwapAgentSwapPairCreated)
	if err := _BSCSwapAgent.contract.UnpackLog(event, "SwapPairCreated", log); err != nil {
		return nil, err
	}
	return event, nil
}

// BSCSwapAgentSwapStartedIterator is returned from FilterSwapStarted and is used to iterate over the raw logs and unpacked data for SwapStarted events raised by the BSCSwapAgent contract.
type BSCSwapAgentSwapStartedIterator struct {
	Event *BSCSwapAgentSwapStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BSCSwapAgentSwapStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BSCSwapAgentSwapStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BSCSwapAgentSwapStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BSCSwapAgentSwapStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BSCSwapAgentSwapStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BSCSwapAgentSwapStarted represents a SwapStarted event raised by the BSCSwapAgent contract.
type BSCSwapAgentSwapStarted struct {
	Bep20Addr common.Address
	Erc20Addr common.Address
	FromAddr  common.Address
	Amount    *big.Int
	FeeAmount *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSwapStarted is a free log retrieval operation binding the contract event 0x49c08ff11118922c1e8298915531eff9ef6f8b39b44b3e9952b75d47e1d0cdd0.
//
// Solidity: event SwapStarted(address indexed bep20Addr, address indexed erc20Addr, address indexed fromAddr, uint256 amount, uint256 feeAmount)
func (_BSCSwapAgent *BSCSwapAgentFilterer) FilterSwapStarted(opts *bind.FilterOpts, bep20Addr []common.Address, erc20Addr []common.Address, fromAddr []common.Address) (*BSCSwapAgentSwapStartedIterator, error) {

	var bep20AddrRule []interface{}
	for _, bep20AddrItem := range bep20Addr {
		bep20AddrRule = append(bep20AddrRule, bep20AddrItem)
	}
	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}
	var fromAddrRule []interface{}
	for _, fromAddrItem := range fromAddr {
		fromAddrRule = append(fromAddrRule, fromAddrItem)
	}

	logs, sub, err := _BSCSwapAgent.contract.FilterLogs(opts, "SwapStarted", bep20AddrRule, erc20AddrRule, fromAddrRule)
	if err != nil {
		return nil, err
	}
	return &BSCSwapAgentSwapStartedIterator{contract: _BSCSwapAgent.contract, event: "SwapStarted", logs: logs, sub: sub}, nil
}

// WatchSwapStarted is a free log subscription operation binding the contract event 0x49c08ff11118922c1e8298915531eff9ef6f8b39b44b3e9952b75d47e1d0cdd0.
//
// Solidity: event SwapStarted(address indexed bep20Addr, address indexed erc20Addr, address indexed fromAddr, uint256 amount, uint256 feeAmount)
func (_BSCSwapAgent *BSCSwapAgentFilterer) WatchSwapStarted(opts *bind.WatchOpts, sink chan<- *BSCSwapAgentSwapStarted, bep20Addr []common.Address, erc20Addr []common.Address, fromAddr []common.Address) (event.Subscription, error) {

	var bep20AddrRule []interface{}
	for _, bep20AddrItem := range bep20Addr {
		bep20AddrRule = append(bep20AddrRule, bep20AddrItem)
	}
	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}
	var fromAddrRule []interface{}
	for _, fromAddrItem := range fromAddr {
		fromAddrRule = append(fromAddrRule, fromAddrItem)
	}

	logs, sub, err := _BSCSwapAgent.contract.WatchLogs(opts, "SwapStarted", bep20AddrRule, erc20AddrRule, fromAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BSCSwapAgentSwapStarted)
				if err := _BSCSwapAgent.contract.UnpackLog(event, "SwapStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSwapStarted is a log parse operation binding the contract event 0x49c08ff11118922c1e8298915531eff9ef6f8b39b44b3e9952b75d47e1d0cdd0.
//
// Solidity: event SwapStarted(address indexed bep20Addr, address indexed erc20Addr, address indexed fromAddr, uint256 amount, uint256 feeAmount)
func (_BSCSwapAgent *BSCSwapAgentFilterer) ParseSwapStarted(log types.Log) (*BSCSwapAgentSwapStarted, error) {
	event := new(BSCSwapAgentSwapStarted)
	if err := _BSCSwapAgent.contract.UnpackLog(event, "SwapStarted", log); err != nil {
		return nil, err
	}
	return event, nil
}
