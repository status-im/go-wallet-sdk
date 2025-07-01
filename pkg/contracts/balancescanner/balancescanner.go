// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package balancescanner

import (
	"errors"
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
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// BalanceScannerResult is an auto generated low-level Go binding around an user-defined struct.
type BalanceScannerResult struct {
	Success bool
	Data    []byte
}

// BalancescannerMetaData contains all meta data concerning the Balancescanner contract.
var BalancescannerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"contracts\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"gas\",\"type\":\"uint256\"}],\"name\":\"call\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"contracts\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"}],\"name\":\"call\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"}],\"name\":\"etherBalances\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"tokenBalances\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"contracts\",\"type\":\"address[]\"}],\"name\":\"tokensBalance\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// BalancescannerABI is the input ABI used to generate the binding from.
// Deprecated: Use BalancescannerMetaData.ABI instead.
var BalancescannerABI = BalancescannerMetaData.ABI

// Balancescanner is an auto generated Go binding around an Ethereum contract.
type Balancescanner struct {
	BalancescannerCaller     // Read-only binding to the contract
	BalancescannerTransactor // Write-only binding to the contract
	BalancescannerFilterer   // Log filterer for contract events
}

// BalancescannerCaller is an auto generated read-only Go binding around an Ethereum contract.
type BalancescannerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalancescannerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BalancescannerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalancescannerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BalancescannerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalancescannerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BalancescannerSession struct {
	Contract     *Balancescanner   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BalancescannerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BalancescannerCallerSession struct {
	Contract *BalancescannerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// BalancescannerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BalancescannerTransactorSession struct {
	Contract     *BalancescannerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// BalancescannerRaw is an auto generated low-level Go binding around an Ethereum contract.
type BalancescannerRaw struct {
	Contract *Balancescanner // Generic contract binding to access the raw methods on
}

// BalancescannerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BalancescannerCallerRaw struct {
	Contract *BalancescannerCaller // Generic read-only contract binding to access the raw methods on
}

// BalancescannerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BalancescannerTransactorRaw struct {
	Contract *BalancescannerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBalancescanner creates a new instance of Balancescanner, bound to a specific deployed contract.
func NewBalancescanner(address common.Address, backend bind.ContractBackend) (*Balancescanner, error) {
	contract, err := bindBalancescanner(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Balancescanner{BalancescannerCaller: BalancescannerCaller{contract: contract}, BalancescannerTransactor: BalancescannerTransactor{contract: contract}, BalancescannerFilterer: BalancescannerFilterer{contract: contract}}, nil
}

// NewBalancescannerCaller creates a new read-only instance of Balancescanner, bound to a specific deployed contract.
func NewBalancescannerCaller(address common.Address, caller bind.ContractCaller) (*BalancescannerCaller, error) {
	contract, err := bindBalancescanner(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BalancescannerCaller{contract: contract}, nil
}

// NewBalancescannerTransactor creates a new write-only instance of Balancescanner, bound to a specific deployed contract.
func NewBalancescannerTransactor(address common.Address, transactor bind.ContractTransactor) (*BalancescannerTransactor, error) {
	contract, err := bindBalancescanner(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BalancescannerTransactor{contract: contract}, nil
}

// NewBalancescannerFilterer creates a new log filterer instance of Balancescanner, bound to a specific deployed contract.
func NewBalancescannerFilterer(address common.Address, filterer bind.ContractFilterer) (*BalancescannerFilterer, error) {
	contract, err := bindBalancescanner(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BalancescannerFilterer{contract: contract}, nil
}

// bindBalancescanner binds a generic wrapper to an already deployed contract.
func bindBalancescanner(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BalancescannerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Balancescanner *BalancescannerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Balancescanner.Contract.BalancescannerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Balancescanner *BalancescannerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Balancescanner.Contract.BalancescannerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Balancescanner *BalancescannerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Balancescanner.Contract.BalancescannerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Balancescanner *BalancescannerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Balancescanner.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Balancescanner *BalancescannerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Balancescanner.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Balancescanner *BalancescannerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Balancescanner.Contract.contract.Transact(opts, method, params...)
}

// Call is a free data retrieval call binding the contract method 0x36738374.
//
// Solidity: function call(address[] contracts, bytes[] data, uint256 gas) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCaller) Call(opts *bind.CallOpts, contracts []common.Address, data [][]byte, gas *big.Int) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _Balancescanner.contract.Call(opts, &out, "call", contracts, data, gas)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// Call is a free data retrieval call binding the contract method 0x36738374.
//
// Solidity: function call(address[] contracts, bytes[] data, uint256 gas) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerSession) Call(contracts []common.Address, data [][]byte, gas *big.Int) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.Call(&_Balancescanner.CallOpts, contracts, data, gas)
}

// Call is a free data retrieval call binding the contract method 0x36738374.
//
// Solidity: function call(address[] contracts, bytes[] data, uint256 gas) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCallerSession) Call(contracts []common.Address, data [][]byte, gas *big.Int) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.Call(&_Balancescanner.CallOpts, contracts, data, gas)
}

// Call0 is a free data retrieval call binding the contract method 0x458b3a7c.
//
// Solidity: function call(address[] contracts, bytes[] data) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCaller) Call0(opts *bind.CallOpts, contracts []common.Address, data [][]byte) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _Balancescanner.contract.Call(opts, &out, "call0", contracts, data)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// Call0 is a free data retrieval call binding the contract method 0x458b3a7c.
//
// Solidity: function call(address[] contracts, bytes[] data) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerSession) Call0(contracts []common.Address, data [][]byte) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.Call0(&_Balancescanner.CallOpts, contracts, data)
}

// Call0 is a free data retrieval call binding the contract method 0x458b3a7c.
//
// Solidity: function call(address[] contracts, bytes[] data) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCallerSession) Call0(contracts []common.Address, data [][]byte) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.Call0(&_Balancescanner.CallOpts, contracts, data)
}

// EtherBalances is a free data retrieval call binding the contract method 0xdbdbb51b.
//
// Solidity: function etherBalances(address[] addresses) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCaller) EtherBalances(opts *bind.CallOpts, addresses []common.Address) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _Balancescanner.contract.Call(opts, &out, "etherBalances", addresses)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// EtherBalances is a free data retrieval call binding the contract method 0xdbdbb51b.
//
// Solidity: function etherBalances(address[] addresses) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerSession) EtherBalances(addresses []common.Address) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.EtherBalances(&_Balancescanner.CallOpts, addresses)
}

// EtherBalances is a free data retrieval call binding the contract method 0xdbdbb51b.
//
// Solidity: function etherBalances(address[] addresses) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCallerSession) EtherBalances(addresses []common.Address) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.EtherBalances(&_Balancescanner.CallOpts, addresses)
}

// TokenBalances is a free data retrieval call binding the contract method 0xaad33091.
//
// Solidity: function tokenBalances(address[] addresses, address token) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCaller) TokenBalances(opts *bind.CallOpts, addresses []common.Address, token common.Address) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _Balancescanner.contract.Call(opts, &out, "tokenBalances", addresses, token)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// TokenBalances is a free data retrieval call binding the contract method 0xaad33091.
//
// Solidity: function tokenBalances(address[] addresses, address token) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerSession) TokenBalances(addresses []common.Address, token common.Address) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.TokenBalances(&_Balancescanner.CallOpts, addresses, token)
}

// TokenBalances is a free data retrieval call binding the contract method 0xaad33091.
//
// Solidity: function tokenBalances(address[] addresses, address token) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCallerSession) TokenBalances(addresses []common.Address, token common.Address) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.TokenBalances(&_Balancescanner.CallOpts, addresses, token)
}

// TokensBalance is a free data retrieval call binding the contract method 0xe5da1b68.
//
// Solidity: function tokensBalance(address owner, address[] contracts) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCaller) TokensBalance(opts *bind.CallOpts, owner common.Address, contracts []common.Address) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _Balancescanner.contract.Call(opts, &out, "tokensBalance", owner, contracts)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// TokensBalance is a free data retrieval call binding the contract method 0xe5da1b68.
//
// Solidity: function tokensBalance(address owner, address[] contracts) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerSession) TokensBalance(owner common.Address, contracts []common.Address) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.TokensBalance(&_Balancescanner.CallOpts, owner, contracts)
}

// TokensBalance is a free data retrieval call binding the contract method 0xe5da1b68.
//
// Solidity: function tokensBalance(address owner, address[] contracts) view returns((bool,bytes)[] results)
func (_Balancescanner *BalancescannerCallerSession) TokensBalance(owner common.Address, contracts []common.Address) ([]BalanceScannerResult, error) {
	return _Balancescanner.Contract.TokensBalance(&_Balancescanner.CallOpts, owner, contracts)
}
