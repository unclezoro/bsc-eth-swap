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

// ETHSwapAgentABI is the input ABI used to generate the binding from.
const ETHSwapAgentABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bscTxHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"toAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"SwapFilled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sponsor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"decimals\",\"type\":\"uint8\"}],\"name\":\"SwapPairRegister\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"fromAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"SwapStarted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"filledBSCTx\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"registeredERC20\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"swapFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"addresspayable\",\"name\":\"ownerAddr\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"setSwapFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"}],\"name\":\"registerSwapPairToBSC\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bscTxHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"toAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"fillBSC2ETHSwap\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"erc20Addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"swapETH2BSC\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"}]"

// ETHSwapAgent is an auto generated Go binding around an Ethereum contract.
type ETHSwapAgent struct {
	ETHSwapAgentCaller     // Read-only binding to the contract
	ETHSwapAgentTransactor // Write-only binding to the contract
	ETHSwapAgentFilterer   // Log filterer for contract events
}

// ETHSwapAgentCaller is an auto generated read-only Go binding around an Ethereum contract.
type ETHSwapAgentCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHSwapAgentTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ETHSwapAgentTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHSwapAgentFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ETHSwapAgentFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHSwapAgentSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ETHSwapAgentSession struct {
	Contract     *ETHSwapAgent     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ETHSwapAgentCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ETHSwapAgentCallerSession struct {
	Contract *ETHSwapAgentCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// ETHSwapAgentTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ETHSwapAgentTransactorSession struct {
	Contract     *ETHSwapAgentTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ETHSwapAgentRaw is an auto generated low-level Go binding around an Ethereum contract.
type ETHSwapAgentRaw struct {
	Contract *ETHSwapAgent // Generic contract binding to access the raw methods on
}

// ETHSwapAgentCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ETHSwapAgentCallerRaw struct {
	Contract *ETHSwapAgentCaller // Generic read-only contract binding to access the raw methods on
}

// ETHSwapAgentTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ETHSwapAgentTransactorRaw struct {
	Contract *ETHSwapAgentTransactor // Generic write-only contract binding to access the raw methods on
}

// NewETHSwapAgent creates a new instance of ETHSwapAgent, bound to a specific deployed contract.
func NewETHSwapAgent(address common.Address, backend bind.ContractBackend) (*ETHSwapAgent, error) {
	contract, err := bindETHSwapAgent(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ETHSwapAgent{ETHSwapAgentCaller: ETHSwapAgentCaller{contract: contract}, ETHSwapAgentTransactor: ETHSwapAgentTransactor{contract: contract}, ETHSwapAgentFilterer: ETHSwapAgentFilterer{contract: contract}}, nil
}

// NewETHSwapAgentCaller creates a new read-only instance of ETHSwapAgent, bound to a specific deployed contract.
func NewETHSwapAgentCaller(address common.Address, caller bind.ContractCaller) (*ETHSwapAgentCaller, error) {
	contract, err := bindETHSwapAgent(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ETHSwapAgentCaller{contract: contract}, nil
}

// NewETHSwapAgentTransactor creates a new write-only instance of ETHSwapAgent, bound to a specific deployed contract.
func NewETHSwapAgentTransactor(address common.Address, transactor bind.ContractTransactor) (*ETHSwapAgentTransactor, error) {
	contract, err := bindETHSwapAgent(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ETHSwapAgentTransactor{contract: contract}, nil
}

// NewETHSwapAgentFilterer creates a new log filterer instance of ETHSwapAgent, bound to a specific deployed contract.
func NewETHSwapAgentFilterer(address common.Address, filterer bind.ContractFilterer) (*ETHSwapAgentFilterer, error) {
	contract, err := bindETHSwapAgent(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ETHSwapAgentFilterer{contract: contract}, nil
}

// bindETHSwapAgent binds a generic wrapper to an already deployed contract.
func bindETHSwapAgent(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ETHSwapAgentABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ETHSwapAgent *ETHSwapAgentRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ETHSwapAgent.Contract.ETHSwapAgentCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ETHSwapAgent *ETHSwapAgentRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.ETHSwapAgentTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ETHSwapAgent *ETHSwapAgentRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.ETHSwapAgentTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ETHSwapAgent *ETHSwapAgentCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ETHSwapAgent.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ETHSwapAgent *ETHSwapAgentTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ETHSwapAgent *ETHSwapAgentTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.contract.Transact(opts, method, params...)
}

// FilledBSCTx is a free data retrieval call binding the contract method 0x50877c77.
//
// Solidity: function filledBSCTx(bytes32 ) constant returns(bool)
func (_ETHSwapAgent *ETHSwapAgentCaller) FilledBSCTx(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ETHSwapAgent.contract.Call(opts, out, "filledBSCTx", arg0)
	return *ret0, err
}

// FilledBSCTx is a free data retrieval call binding the contract method 0x50877c77.
//
// Solidity: function filledBSCTx(bytes32 ) constant returns(bool)
func (_ETHSwapAgent *ETHSwapAgentSession) FilledBSCTx(arg0 [32]byte) (bool, error) {
	return _ETHSwapAgent.Contract.FilledBSCTx(&_ETHSwapAgent.CallOpts, arg0)
}

// FilledBSCTx is a free data retrieval call binding the contract method 0x50877c77.
//
// Solidity: function filledBSCTx(bytes32 ) constant returns(bool)
func (_ETHSwapAgent *ETHSwapAgentCallerSession) FilledBSCTx(arg0 [32]byte) (bool, error) {
	return _ETHSwapAgent.Contract.FilledBSCTx(&_ETHSwapAgent.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ETHSwapAgent *ETHSwapAgentCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ETHSwapAgent.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ETHSwapAgent *ETHSwapAgentSession) Owner() (common.Address, error) {
	return _ETHSwapAgent.Contract.Owner(&_ETHSwapAgent.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ETHSwapAgent *ETHSwapAgentCallerSession) Owner() (common.Address, error) {
	return _ETHSwapAgent.Contract.Owner(&_ETHSwapAgent.CallOpts)
}

// RegisteredERC20 is a free data retrieval call binding the contract method 0x89b15604.
//
// Solidity: function registeredERC20(address ) constant returns(bool)
func (_ETHSwapAgent *ETHSwapAgentCaller) RegisteredERC20(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ETHSwapAgent.contract.Call(opts, out, "registeredERC20", arg0)
	return *ret0, err
}

// RegisteredERC20 is a free data retrieval call binding the contract method 0x89b15604.
//
// Solidity: function registeredERC20(address ) constant returns(bool)
func (_ETHSwapAgent *ETHSwapAgentSession) RegisteredERC20(arg0 common.Address) (bool, error) {
	return _ETHSwapAgent.Contract.RegisteredERC20(&_ETHSwapAgent.CallOpts, arg0)
}

// RegisteredERC20 is a free data retrieval call binding the contract method 0x89b15604.
//
// Solidity: function registeredERC20(address ) constant returns(bool)
func (_ETHSwapAgent *ETHSwapAgentCallerSession) RegisteredERC20(arg0 common.Address) (bool, error) {
	return _ETHSwapAgent.Contract.RegisteredERC20(&_ETHSwapAgent.CallOpts, arg0)
}

// SwapFee is a free data retrieval call binding the contract method 0x54cf2aeb.
//
// Solidity: function swapFee() constant returns(uint256)
func (_ETHSwapAgent *ETHSwapAgentCaller) SwapFee(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ETHSwapAgent.contract.Call(opts, out, "swapFee")
	return *ret0, err
}

// SwapFee is a free data retrieval call binding the contract method 0x54cf2aeb.
//
// Solidity: function swapFee() constant returns(uint256)
func (_ETHSwapAgent *ETHSwapAgentSession) SwapFee() (*big.Int, error) {
	return _ETHSwapAgent.Contract.SwapFee(&_ETHSwapAgent.CallOpts)
}

// SwapFee is a free data retrieval call binding the contract method 0x54cf2aeb.
//
// Solidity: function swapFee() constant returns(uint256)
func (_ETHSwapAgent *ETHSwapAgentCallerSession) SwapFee() (*big.Int, error) {
	return _ETHSwapAgent.Contract.SwapFee(&_ETHSwapAgent.CallOpts)
}

// FillBSC2ETHSwap is a paid mutator transaction binding the contract method 0x9867df11.
//
// Solidity: function fillBSC2ETHSwap(bytes32 bscTxHash, address erc20Addr, address toAddress, uint256 amount) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentTransactor) FillBSC2ETHSwap(opts *bind.TransactOpts, bscTxHash [32]byte, erc20Addr common.Address, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.contract.Transact(opts, "fillBSC2ETHSwap", bscTxHash, erc20Addr, toAddress, amount)
}

// FillBSC2ETHSwap is a paid mutator transaction binding the contract method 0x9867df11.
//
// Solidity: function fillBSC2ETHSwap(bytes32 bscTxHash, address erc20Addr, address toAddress, uint256 amount) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentSession) FillBSC2ETHSwap(bscTxHash [32]byte, erc20Addr common.Address, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.FillBSC2ETHSwap(&_ETHSwapAgent.TransactOpts, bscTxHash, erc20Addr, toAddress, amount)
}

// FillBSC2ETHSwap is a paid mutator transaction binding the contract method 0x9867df11.
//
// Solidity: function fillBSC2ETHSwap(bytes32 bscTxHash, address erc20Addr, address toAddress, uint256 amount) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentTransactorSession) FillBSC2ETHSwap(bscTxHash [32]byte, erc20Addr common.Address, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.FillBSC2ETHSwap(&_ETHSwapAgent.TransactOpts, bscTxHash, erc20Addr, toAddress, amount)
}

// Initialize is a paid mutator transaction binding the contract method 0xda35a26f.
//
// Solidity: function initialize(uint256 fee, address ownerAddr) returns()
func (_ETHSwapAgent *ETHSwapAgentTransactor) Initialize(opts *bind.TransactOpts, fee *big.Int, ownerAddr common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.contract.Transact(opts, "initialize", fee, ownerAddr)
}

// Initialize is a paid mutator transaction binding the contract method 0xda35a26f.
//
// Solidity: function initialize(uint256 fee, address ownerAddr) returns()
func (_ETHSwapAgent *ETHSwapAgentSession) Initialize(fee *big.Int, ownerAddr common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.Initialize(&_ETHSwapAgent.TransactOpts, fee, ownerAddr)
}

// Initialize is a paid mutator transaction binding the contract method 0xda35a26f.
//
// Solidity: function initialize(uint256 fee, address ownerAddr) returns()
func (_ETHSwapAgent *ETHSwapAgentTransactorSession) Initialize(fee *big.Int, ownerAddr common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.Initialize(&_ETHSwapAgent.TransactOpts, fee, ownerAddr)
}

// RegisterSwapPairToBSC is a paid mutator transaction binding the contract method 0x5c13c151.
//
// Solidity: function registerSwapPairToBSC(address erc20Addr) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentTransactor) RegisterSwapPairToBSC(opts *bind.TransactOpts, erc20Addr common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.contract.Transact(opts, "registerSwapPairToBSC", erc20Addr)
}

// RegisterSwapPairToBSC is a paid mutator transaction binding the contract method 0x5c13c151.
//
// Solidity: function registerSwapPairToBSC(address erc20Addr) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentSession) RegisterSwapPairToBSC(erc20Addr common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.RegisterSwapPairToBSC(&_ETHSwapAgent.TransactOpts, erc20Addr)
}

// RegisterSwapPairToBSC is a paid mutator transaction binding the contract method 0x5c13c151.
//
// Solidity: function registerSwapPairToBSC(address erc20Addr) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentTransactorSession) RegisterSwapPairToBSC(erc20Addr common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.RegisterSwapPairToBSC(&_ETHSwapAgent.TransactOpts, erc20Addr)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ETHSwapAgent *ETHSwapAgentTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHSwapAgent.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ETHSwapAgent *ETHSwapAgentSession) RenounceOwnership() (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.RenounceOwnership(&_ETHSwapAgent.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ETHSwapAgent *ETHSwapAgentTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.RenounceOwnership(&_ETHSwapAgent.TransactOpts)
}

// SetSwapFee is a paid mutator transaction binding the contract method 0x34e19907.
//
// Solidity: function setSwapFee(uint256 fee) returns()
func (_ETHSwapAgent *ETHSwapAgentTransactor) SetSwapFee(opts *bind.TransactOpts, fee *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.contract.Transact(opts, "setSwapFee", fee)
}

// SetSwapFee is a paid mutator transaction binding the contract method 0x34e19907.
//
// Solidity: function setSwapFee(uint256 fee) returns()
func (_ETHSwapAgent *ETHSwapAgentSession) SetSwapFee(fee *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.SetSwapFee(&_ETHSwapAgent.TransactOpts, fee)
}

// SetSwapFee is a paid mutator transaction binding the contract method 0x34e19907.
//
// Solidity: function setSwapFee(uint256 fee) returns()
func (_ETHSwapAgent *ETHSwapAgentTransactorSession) SetSwapFee(fee *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.SetSwapFee(&_ETHSwapAgent.TransactOpts, fee)
}

// SwapETH2BSC is a paid mutator transaction binding the contract method 0xb9927a9c.
//
// Solidity: function swapETH2BSC(address erc20Addr, uint256 amount) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentTransactor) SwapETH2BSC(opts *bind.TransactOpts, erc20Addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.contract.Transact(opts, "swapETH2BSC", erc20Addr, amount)
}

// SwapETH2BSC is a paid mutator transaction binding the contract method 0xb9927a9c.
//
// Solidity: function swapETH2BSC(address erc20Addr, uint256 amount) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentSession) SwapETH2BSC(erc20Addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.SwapETH2BSC(&_ETHSwapAgent.TransactOpts, erc20Addr, amount)
}

// SwapETH2BSC is a paid mutator transaction binding the contract method 0xb9927a9c.
//
// Solidity: function swapETH2BSC(address erc20Addr, uint256 amount) returns(bool)
func (_ETHSwapAgent *ETHSwapAgentTransactorSession) SwapETH2BSC(erc20Addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.SwapETH2BSC(&_ETHSwapAgent.TransactOpts, erc20Addr, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ETHSwapAgent *ETHSwapAgentTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ETHSwapAgent *ETHSwapAgentSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.TransferOwnership(&_ETHSwapAgent.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ETHSwapAgent *ETHSwapAgentTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ETHSwapAgent.Contract.TransferOwnership(&_ETHSwapAgent.TransactOpts, newOwner)
}

// ETHSwapAgentOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ETHSwapAgent contract.
type ETHSwapAgentOwnershipTransferredIterator struct {
	Event *ETHSwapAgentOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ETHSwapAgentOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHSwapAgentOwnershipTransferred)
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
		it.Event = new(ETHSwapAgentOwnershipTransferred)
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
func (it *ETHSwapAgentOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHSwapAgentOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHSwapAgentOwnershipTransferred represents a OwnershipTransferred event raised by the ETHSwapAgent contract.
type ETHSwapAgentOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ETHSwapAgent *ETHSwapAgentFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ETHSwapAgentOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ETHSwapAgent.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ETHSwapAgentOwnershipTransferredIterator{contract: _ETHSwapAgent.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ETHSwapAgent *ETHSwapAgentFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ETHSwapAgentOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ETHSwapAgent.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHSwapAgentOwnershipTransferred)
				if err := _ETHSwapAgent.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ETHSwapAgent *ETHSwapAgentFilterer) ParseOwnershipTransferred(log types.Log) (*ETHSwapAgentOwnershipTransferred, error) {
	event := new(ETHSwapAgentOwnershipTransferred)
	if err := _ETHSwapAgent.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ETHSwapAgentSwapFilledIterator is returned from FilterSwapFilled and is used to iterate over the raw logs and unpacked data for SwapFilled events raised by the ETHSwapAgent contract.
type ETHSwapAgentSwapFilledIterator struct {
	Event *ETHSwapAgentSwapFilled // Event containing the contract specifics and raw log

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
func (it *ETHSwapAgentSwapFilledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHSwapAgentSwapFilled)
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
		it.Event = new(ETHSwapAgentSwapFilled)
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
func (it *ETHSwapAgentSwapFilledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHSwapAgentSwapFilledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHSwapAgentSwapFilled represents a SwapFilled event raised by the ETHSwapAgent contract.
type ETHSwapAgentSwapFilled struct {
	Erc20Addr common.Address
	BscTxHash [32]byte
	ToAddress common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSwapFilled is a free log retrieval operation binding the contract event 0x3bebd9a738291e69898b5dbfadb6329b4b09fc648bdef68762928e521463abd9.
//
// Solidity: event SwapFilled(address indexed erc20Addr, bytes32 indexed bscTxHash, address indexed toAddress, uint256 amount)
func (_ETHSwapAgent *ETHSwapAgentFilterer) FilterSwapFilled(opts *bind.FilterOpts, erc20Addr []common.Address, bscTxHash [][32]byte, toAddress []common.Address) (*ETHSwapAgentSwapFilledIterator, error) {

	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}
	var bscTxHashRule []interface{}
	for _, bscTxHashItem := range bscTxHash {
		bscTxHashRule = append(bscTxHashRule, bscTxHashItem)
	}
	var toAddressRule []interface{}
	for _, toAddressItem := range toAddress {
		toAddressRule = append(toAddressRule, toAddressItem)
	}

	logs, sub, err := _ETHSwapAgent.contract.FilterLogs(opts, "SwapFilled", erc20AddrRule, bscTxHashRule, toAddressRule)
	if err != nil {
		return nil, err
	}
	return &ETHSwapAgentSwapFilledIterator{contract: _ETHSwapAgent.contract, event: "SwapFilled", logs: logs, sub: sub}, nil
}

// WatchSwapFilled is a free log subscription operation binding the contract event 0x3bebd9a738291e69898b5dbfadb6329b4b09fc648bdef68762928e521463abd9.
//
// Solidity: event SwapFilled(address indexed erc20Addr, bytes32 indexed bscTxHash, address indexed toAddress, uint256 amount)
func (_ETHSwapAgent *ETHSwapAgentFilterer) WatchSwapFilled(opts *bind.WatchOpts, sink chan<- *ETHSwapAgentSwapFilled, erc20Addr []common.Address, bscTxHash [][32]byte, toAddress []common.Address) (event.Subscription, error) {

	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}
	var bscTxHashRule []interface{}
	for _, bscTxHashItem := range bscTxHash {
		bscTxHashRule = append(bscTxHashRule, bscTxHashItem)
	}
	var toAddressRule []interface{}
	for _, toAddressItem := range toAddress {
		toAddressRule = append(toAddressRule, toAddressItem)
	}

	logs, sub, err := _ETHSwapAgent.contract.WatchLogs(opts, "SwapFilled", erc20AddrRule, bscTxHashRule, toAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHSwapAgentSwapFilled)
				if err := _ETHSwapAgent.contract.UnpackLog(event, "SwapFilled", log); err != nil {
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
// Solidity: event SwapFilled(address indexed erc20Addr, bytes32 indexed bscTxHash, address indexed toAddress, uint256 amount)
func (_ETHSwapAgent *ETHSwapAgentFilterer) ParseSwapFilled(log types.Log) (*ETHSwapAgentSwapFilled, error) {
	event := new(ETHSwapAgentSwapFilled)
	if err := _ETHSwapAgent.contract.UnpackLog(event, "SwapFilled", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ETHSwapAgentSwapPairRegisterIterator is returned from FilterSwapPairRegister and is used to iterate over the raw logs and unpacked data for SwapPairRegister events raised by the ETHSwapAgent contract.
type ETHSwapAgentSwapPairRegisterIterator struct {
	Event *ETHSwapAgentSwapPairRegister // Event containing the contract specifics and raw log

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
func (it *ETHSwapAgentSwapPairRegisterIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHSwapAgentSwapPairRegister)
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
		it.Event = new(ETHSwapAgentSwapPairRegister)
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
func (it *ETHSwapAgentSwapPairRegisterIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHSwapAgentSwapPairRegisterIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHSwapAgentSwapPairRegister represents a SwapPairRegister event raised by the ETHSwapAgent contract.
type ETHSwapAgentSwapPairRegister struct {
	Sponsor   common.Address
	Erc20Addr common.Address
	Name      string
	Symbol    string
	Decimals  uint8
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSwapPairRegister is a free log retrieval operation binding the contract event 0xfe3bd005e346323fa452df8cafc28c55b99e3766ba8750571d139c6cf5bc08a0.
//
// Solidity: event SwapPairRegister(address indexed sponsor, address indexed erc20Addr, string name, string symbol, uint8 decimals)
func (_ETHSwapAgent *ETHSwapAgentFilterer) FilterSwapPairRegister(opts *bind.FilterOpts, sponsor []common.Address, erc20Addr []common.Address) (*ETHSwapAgentSwapPairRegisterIterator, error) {

	var sponsorRule []interface{}
	for _, sponsorItem := range sponsor {
		sponsorRule = append(sponsorRule, sponsorItem)
	}
	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}

	logs, sub, err := _ETHSwapAgent.contract.FilterLogs(opts, "SwapPairRegister", sponsorRule, erc20AddrRule)
	if err != nil {
		return nil, err
	}
	return &ETHSwapAgentSwapPairRegisterIterator{contract: _ETHSwapAgent.contract, event: "SwapPairRegister", logs: logs, sub: sub}, nil
}

// WatchSwapPairRegister is a free log subscription operation binding the contract event 0xfe3bd005e346323fa452df8cafc28c55b99e3766ba8750571d139c6cf5bc08a0.
//
// Solidity: event SwapPairRegister(address indexed sponsor, address indexed erc20Addr, string name, string symbol, uint8 decimals)
func (_ETHSwapAgent *ETHSwapAgentFilterer) WatchSwapPairRegister(opts *bind.WatchOpts, sink chan<- *ETHSwapAgentSwapPairRegister, sponsor []common.Address, erc20Addr []common.Address) (event.Subscription, error) {

	var sponsorRule []interface{}
	for _, sponsorItem := range sponsor {
		sponsorRule = append(sponsorRule, sponsorItem)
	}
	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}

	logs, sub, err := _ETHSwapAgent.contract.WatchLogs(opts, "SwapPairRegister", sponsorRule, erc20AddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHSwapAgentSwapPairRegister)
				if err := _ETHSwapAgent.contract.UnpackLog(event, "SwapPairRegister", log); err != nil {
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

// ParseSwapPairRegister is a log parse operation binding the contract event 0xfe3bd005e346323fa452df8cafc28c55b99e3766ba8750571d139c6cf5bc08a0.
//
// Solidity: event SwapPairRegister(address indexed sponsor, address indexed erc20Addr, string name, string symbol, uint8 decimals)
func (_ETHSwapAgent *ETHSwapAgentFilterer) ParseSwapPairRegister(log types.Log) (*ETHSwapAgentSwapPairRegister, error) {
	event := new(ETHSwapAgentSwapPairRegister)
	if err := _ETHSwapAgent.contract.UnpackLog(event, "SwapPairRegister", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ETHSwapAgentSwapStartedIterator is returned from FilterSwapStarted and is used to iterate over the raw logs and unpacked data for SwapStarted events raised by the ETHSwapAgent contract.
type ETHSwapAgentSwapStartedIterator struct {
	Event *ETHSwapAgentSwapStarted // Event containing the contract specifics and raw log

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
func (it *ETHSwapAgentSwapStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHSwapAgentSwapStarted)
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
		it.Event = new(ETHSwapAgentSwapStarted)
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
func (it *ETHSwapAgentSwapStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHSwapAgentSwapStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHSwapAgentSwapStarted represents a SwapStarted event raised by the ETHSwapAgent contract.
type ETHSwapAgentSwapStarted struct {
	Erc20Addr common.Address
	FromAddr  common.Address
	Amount    *big.Int
	FeeAmount *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSwapStarted is a free log retrieval operation binding the contract event 0xf60309f865a6aa297da5fac6188136a02e5acfdf6e8f6d35257a9f4e9653170f.
//
// Solidity: event SwapStarted(address indexed erc20Addr, address indexed fromAddr, uint256 amount, uint256 feeAmount)
func (_ETHSwapAgent *ETHSwapAgentFilterer) FilterSwapStarted(opts *bind.FilterOpts, erc20Addr []common.Address, fromAddr []common.Address) (*ETHSwapAgentSwapStartedIterator, error) {

	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}
	var fromAddrRule []interface{}
	for _, fromAddrItem := range fromAddr {
		fromAddrRule = append(fromAddrRule, fromAddrItem)
	}

	logs, sub, err := _ETHSwapAgent.contract.FilterLogs(opts, "SwapStarted", erc20AddrRule, fromAddrRule)
	if err != nil {
		return nil, err
	}
	return &ETHSwapAgentSwapStartedIterator{contract: _ETHSwapAgent.contract, event: "SwapStarted", logs: logs, sub: sub}, nil
}

// WatchSwapStarted is a free log subscription operation binding the contract event 0xf60309f865a6aa297da5fac6188136a02e5acfdf6e8f6d35257a9f4e9653170f.
//
// Solidity: event SwapStarted(address indexed erc20Addr, address indexed fromAddr, uint256 amount, uint256 feeAmount)
func (_ETHSwapAgent *ETHSwapAgentFilterer) WatchSwapStarted(opts *bind.WatchOpts, sink chan<- *ETHSwapAgentSwapStarted, erc20Addr []common.Address, fromAddr []common.Address) (event.Subscription, error) {

	var erc20AddrRule []interface{}
	for _, erc20AddrItem := range erc20Addr {
		erc20AddrRule = append(erc20AddrRule, erc20AddrItem)
	}
	var fromAddrRule []interface{}
	for _, fromAddrItem := range fromAddr {
		fromAddrRule = append(fromAddrRule, fromAddrItem)
	}

	logs, sub, err := _ETHSwapAgent.contract.WatchLogs(opts, "SwapStarted", erc20AddrRule, fromAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHSwapAgentSwapStarted)
				if err := _ETHSwapAgent.contract.UnpackLog(event, "SwapStarted", log); err != nil {
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

// ParseSwapStarted is a log parse operation binding the contract event 0xf60309f865a6aa297da5fac6188136a02e5acfdf6e8f6d35257a9f4e9653170f.
//
// Solidity: event SwapStarted(address indexed erc20Addr, address indexed fromAddr, uint256 amount, uint256 feeAmount)
func (_ETHSwapAgent *ETHSwapAgentFilterer) ParseSwapStarted(log types.Log) (*ETHSwapAgentSwapStarted, error) {
	event := new(ETHSwapAgentSwapStarted)
	if err := _ETHSwapAgent.contract.UnpackLog(event, "SwapStarted", log); err != nil {
		return nil, err
	}
	return event, nil
}
