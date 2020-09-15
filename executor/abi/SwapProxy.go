// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package swap_proxy

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

// SwapProxyABI is the input ABI used to generate the binding from.
const SwapProxyABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"fromAddr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"toAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"tokenTransfer\",\"type\":\"event\"}]"

// SwapProxy is an auto generated Go binding around an Ethereum contract.
type SwapProxy struct {
	SwapProxyCaller     // Read-only binding to the contract
	SwapProxyTransactor // Write-only binding to the contract
	SwapProxyFilterer   // Log filterer for contract events
}

// SwapProxyCaller is an auto generated read-only Go binding around an Ethereum contract.
type SwapProxyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapProxyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SwapProxyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapProxyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SwapProxyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapProxySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SwapProxySession struct {
	Contract     *SwapProxy        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapProxyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SwapProxyCallerSession struct {
	Contract *SwapProxyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// SwapProxyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SwapProxyTransactorSession struct {
	Contract     *SwapProxyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// SwapProxyRaw is an auto generated low-level Go binding around an Ethereum contract.
type SwapProxyRaw struct {
	Contract *SwapProxy // Generic contract binding to access the raw methods on
}

// SwapProxyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SwapProxyCallerRaw struct {
	Contract *SwapProxyCaller // Generic read-only contract binding to access the raw methods on
}

// SwapProxyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SwapProxyTransactorRaw struct {
	Contract *SwapProxyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSwapProxy creates a new instance of SwapProxy, bound to a specific deployed contract.
func NewSwapProxy(address common.Address, backend bind.ContractBackend) (*SwapProxy, error) {
	contract, err := bindSwapProxy(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SwapProxy{SwapProxyCaller: SwapProxyCaller{contract: contract}, SwapProxyTransactor: SwapProxyTransactor{contract: contract}, SwapProxyFilterer: SwapProxyFilterer{contract: contract}}, nil
}

// NewSwapProxyCaller creates a new read-only instance of SwapProxy, bound to a specific deployed contract.
func NewSwapProxyCaller(address common.Address, caller bind.ContractCaller) (*SwapProxyCaller, error) {
	contract, err := bindSwapProxy(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SwapProxyCaller{contract: contract}, nil
}

// NewSwapProxyTransactor creates a new write-only instance of SwapProxy, bound to a specific deployed contract.
func NewSwapProxyTransactor(address common.Address, transactor bind.ContractTransactor) (*SwapProxyTransactor, error) {
	contract, err := bindSwapProxy(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SwapProxyTransactor{contract: contract}, nil
}

// NewSwapProxyFilterer creates a new log filterer instance of SwapProxy, bound to a specific deployed contract.
func NewSwapProxyFilterer(address common.Address, filterer bind.ContractFilterer) (*SwapProxyFilterer, error) {
	contract, err := bindSwapProxy(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SwapProxyFilterer{contract: contract}, nil
}

// bindSwapProxy binds a generic wrapper to an already deployed contract.
func bindSwapProxy(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapProxyABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapProxy *SwapProxyRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SwapProxy.Contract.SwapProxyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapProxy *SwapProxyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapProxy.Contract.SwapProxyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapProxy *SwapProxyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapProxy.Contract.SwapProxyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwapProxy *SwapProxyCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SwapProxy.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwapProxy *SwapProxyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwapProxy.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwapProxy *SwapProxyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwapProxy.Contract.contract.Transact(opts, method, params...)
}

// SwapProxyTokenTransferIterator is returned from FilterTokenTransfer and is used to iterate over the raw logs and unpacked data for TokenTransfer events raised by the SwapProxy contract.
type SwapProxyTokenTransferIterator struct {
	Event *SwapProxyTokenTransfer // Event containing the contract specifics and raw log

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
func (it *SwapProxyTokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapProxyTokenTransfer)
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
		it.Event = new(SwapProxyTokenTransfer)
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
func (it *SwapProxyTokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapProxyTokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapProxyTokenTransfer represents a TokenTransfer event raised by the SwapProxy contract.
type SwapProxyTokenTransfer struct {
	ContractAddr common.Address
	FromAddr     common.Address
	ToAddr       common.Address
	Amount       *big.Int
	FeeAmount    *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTokenTransfer is a free log retrieval operation binding the contract event 0x05e8bebd9fcb5eb8e77fbd53c65340bcda78c0ff916583b5eff776e21316dced.
//
// Solidity: event tokenTransfer(address indexed contractAddr, address indexed fromAddr, address indexed toAddr, uint256 amount, uint256 feeAmount)
func (_SwapProxy *SwapProxyFilterer) FilterTokenTransfer(opts *bind.FilterOpts, contractAddr []common.Address, fromAddr []common.Address, toAddr []common.Address) (*SwapProxyTokenTransferIterator, error) {

	var contractAddrRule []interface{}
	for _, contractAddrItem := range contractAddr {
		contractAddrRule = append(contractAddrRule, contractAddrItem)
	}
	var fromAddrRule []interface{}
	for _, fromAddrItem := range fromAddr {
		fromAddrRule = append(fromAddrRule, fromAddrItem)
	}
	var toAddrRule []interface{}
	for _, toAddrItem := range toAddr {
		toAddrRule = append(toAddrRule, toAddrItem)
	}

	logs, sub, err := _SwapProxy.contract.FilterLogs(opts, "tokenTransfer", contractAddrRule, fromAddrRule, toAddrRule)
	if err != nil {
		return nil, err
	}
	return &SwapProxyTokenTransferIterator{contract: _SwapProxy.contract, event: "tokenTransfer", logs: logs, sub: sub}, nil
}

// WatchTokenTransfer is a free log subscription operation binding the contract event 0x05e8bebd9fcb5eb8e77fbd53c65340bcda78c0ff916583b5eff776e21316dced.
//
// Solidity: event tokenTransfer(address indexed contractAddr, address indexed fromAddr, address indexed toAddr, uint256 amount, uint256 feeAmount)
func (_SwapProxy *SwapProxyFilterer) WatchTokenTransfer(opts *bind.WatchOpts, sink chan<- *SwapProxyTokenTransfer, contractAddr []common.Address, fromAddr []common.Address, toAddr []common.Address) (event.Subscription, error) {

	var contractAddrRule []interface{}
	for _, contractAddrItem := range contractAddr {
		contractAddrRule = append(contractAddrRule, contractAddrItem)
	}
	var fromAddrRule []interface{}
	for _, fromAddrItem := range fromAddr {
		fromAddrRule = append(fromAddrRule, fromAddrItem)
	}
	var toAddrRule []interface{}
	for _, toAddrItem := range toAddr {
		toAddrRule = append(toAddrRule, toAddrItem)
	}

	logs, sub, err := _SwapProxy.contract.WatchLogs(opts, "tokenTransfer", contractAddrRule, fromAddrRule, toAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapProxyTokenTransfer)
				if err := _SwapProxy.contract.UnpackLog(event, "tokenTransfer", log); err != nil {
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

// ParseTokenTransfer is a log parse operation binding the contract event 0x05e8bebd9fcb5eb8e77fbd53c65340bcda78c0ff916583b5eff776e21316dced.
//
// Solidity: event tokenTransfer(address indexed contractAddr, address indexed fromAddr, address indexed toAddr, uint256 amount, uint256 feeAmount)
func (_SwapProxy *SwapProxyFilterer) ParseTokenTransfer(log types.Log) (*SwapProxyTokenTransfer, error) {
	event := new(SwapProxyTokenTransfer)
	if err := _SwapProxy.contract.UnpackLog(event, "tokenTransfer", log); err != nil {
		return nil, err
	}
	return event, nil
}
