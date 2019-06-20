// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package main

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

// ZkPoDExchangeABI is the input ABI used to generate the binding from.
const ZkPoDExchangeABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"publicVar_\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"bobDeposits_\",\"outputs\":[{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"unDepositAt\",\"type\":\"uint256\"},{\"name\":\"pendingCnt\",\"type\":\"uint256\"},{\"name\":\"stat\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"s_\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"t3\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"bulletins_\",\"outputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"size\",\"type\":\"uint64\"},{\"name\":\"s\",\"type\":\"uint64\"},{\"name\":\"n\",\"type\":\"uint64\"},{\"name\":\"sigma_mkl_root\",\"type\":\"uint256\"},{\"name\":\"vrf_meta_digest\",\"type\":\"uint256\"},{\"name\":\"pledge_value\",\"type\":\"uint256\"},{\"name\":\"unDepositAt\",\"type\":\"uint256\"},{\"name\":\"blt_type\",\"type\":\"uint8\"},{\"name\":\"stat\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"t1\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"t4\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_publicVar\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_a\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_b\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_sid\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_mode\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"OnDeal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_a\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_b\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_sid\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_seed0\",\"type\":\"bytes32\"}],\"name\":\"OnComplaintKey\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_a\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_b\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_sid\",\"type\":\"uint256\"}],\"name\":\"OnComplaintClaim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_a\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_b\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_sid\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"OnComplaintDeal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_a\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_b\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_sid\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_seed0\",\"type\":\"bytes32\"}],\"name\":\"OnAtomicSwapDeal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_a\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_b\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_sid\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_r\",\"type\":\"uint256\"}],\"name\":\"OnVRFDeal\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_size\",\"type\":\"uint64\"},{\"name\":\"_s\",\"type\":\"uint64\"},{\"name\":\"_n\",\"type\":\"uint64\"},{\"name\":\"_sigma_mkl_root\",\"type\":\"uint256\"},{\"name\":\"_vrf_meta_digest\",\"type\":\"uint256\"},{\"name\":\"_blt_type\",\"type\":\"uint256\"}],\"name\":\"publish\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_bltKey\",\"type\":\"bytes32\"}],\"name\":\"unPublish\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"}],\"name\":\"bobDeposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"}],\"name\":\"bobUnDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_bltKey\",\"type\":\"bytes32\"}],\"name\":\"withdrawA\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"}],\"name\":\"withdrawB\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_seed0\",\"type\":\"bytes32\"},{\"name\":\"_sid\",\"type\":\"uint256\"},{\"name\":\"_b\",\"type\":\"address\"},{\"name\":\"_seed2\",\"type\":\"bytes32\"},{\"name\":\"_k_mkl_root\",\"type\":\"bytes32\"},{\"name\":\"_count\",\"type\":\"uint64\"},{\"name\":\"_price\",\"type\":\"uint256\"},{\"name\":\"_expireAt\",\"type\":\"uint256\"},{\"name\":\"_sig\",\"type\":\"bytes\"}],\"name\":\"submitProofComplaint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_a\",\"type\":\"address\"},{\"name\":\"_sid\",\"type\":\"uint256\"},{\"name\":\"_i\",\"type\":\"uint64\"},{\"name\":\"_j\",\"type\":\"uint64\"},{\"name\":\"_tx\",\"type\":\"uint256\"},{\"name\":\"_ty\",\"type\":\"uint256\"},{\"name\":\"_mkl_path\",\"type\":\"bytes32[]\"},{\"name\":\"_sCnt\",\"type\":\"uint64\"}],\"name\":\"claimComplaint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_a\",\"type\":\"address\"},{\"name\":\"_b\",\"type\":\"address\"},{\"name\":\"_sid\",\"type\":\"uint256\"}],\"name\":\"settleComplaintDeal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_seed0\",\"type\":\"bytes32\"},{\"name\":\"_sCnt\",\"type\":\"uint64\"},{\"name\":\"_sid\",\"type\":\"uint256\"},{\"name\":\"_b\",\"type\":\"address\"},{\"name\":\"_seed2\",\"type\":\"bytes32\"},{\"name\":\"_sigma_vw\",\"type\":\"uint256\"},{\"name\":\"_count\",\"type\":\"uint64\"},{\"name\":\"_price\",\"type\":\"uint256\"},{\"name\":\"_expireAt\",\"type\":\"uint256\"},{\"name\":\"_sig\",\"type\":\"bytes\"}],\"name\":\"submitProofAtomicSwap\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_s_r\",\"type\":\"uint256\"},{\"name\":\"_sid\",\"type\":\"uint256\"},{\"name\":\"_b\",\"type\":\"address\"},{\"name\":\"_g_exp_r\",\"type\":\"uint256[2]\"},{\"name\":\"_price\",\"type\":\"uint256\"},{\"name\":\"_expireAt\",\"type\":\"uint256\"},{\"name\":\"_sig\",\"type\":\"bytes\"}],\"name\":\"submitProofVRF\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_a\",\"type\":\"address\"},{\"name\":\"_b\",\"type\":\"address\"},{\"name\":\"_sid\",\"type\":\"uint256\"}],\"name\":\"getSessionRecord\",\"outputs\":[{\"name\":\"submitAt\",\"type\":\"uint256\"},{\"name\":\"mode\",\"type\":\"uint8\"},{\"name\":\"stat\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_a\",\"type\":\"address\"},{\"name\":\"_b\",\"type\":\"address\"},{\"name\":\"_sid\",\"type\":\"uint256\"}],\"name\":\"getRecordComplaint\",\"outputs\":[{\"name\":\"seed0\",\"type\":\"bytes32\"},{\"name\":\"seed2\",\"type\":\"bytes32\"},{\"name\":\"k_mkl_root\",\"type\":\"bytes32\"},{\"name\":\"count\",\"type\":\"uint64\"},{\"name\":\"price\",\"type\":\"uint256\"},{\"name\":\"expireAt\",\"type\":\"uint256\"},{\"name\":\"submitAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_a\",\"type\":\"address\"},{\"name\":\"_b\",\"type\":\"address\"},{\"name\":\"_sid\",\"type\":\"uint256\"}],\"name\":\"getRecordAtomicSwap\",\"outputs\":[{\"name\":\"seed0\",\"type\":\"bytes32\"},{\"name\":\"submitAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_a\",\"type\":\"address\"},{\"name\":\"_b\",\"type\":\"address\"},{\"name\":\"_sid\",\"type\":\"uint256\"}],\"name\":\"getRecordVRF\",\"outputs\":[{\"name\":\"r\",\"type\":\"uint256\"},{\"name\":\"submitAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ZkPoDExchange is an auto generated Go binding around an Ethereum contract.
type ZkPoDExchange struct {
	ZkPoDExchangeCaller     // Read-only binding to the contract
	ZkPoDExchangeTransactor // Write-only binding to the contract
	ZkPoDExchangeFilterer   // Log filterer for contract events
}

// ZkPoDExchangeCaller is an auto generated read-only Go binding around an Ethereum contract.
type ZkPoDExchangeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZkPoDExchangeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ZkPoDExchangeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZkPoDExchangeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ZkPoDExchangeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZkPoDExchangeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ZkPoDExchangeSession struct {
	Contract     *ZkPoDExchange    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ZkPoDExchangeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ZkPoDExchangeCallerSession struct {
	Contract *ZkPoDExchangeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ZkPoDExchangeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ZkPoDExchangeTransactorSession struct {
	Contract     *ZkPoDExchangeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ZkPoDExchangeRaw is an auto generated low-level Go binding around an Ethereum contract.
type ZkPoDExchangeRaw struct {
	Contract *ZkPoDExchange // Generic contract binding to access the raw methods on
}

// ZkPoDExchangeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ZkPoDExchangeCallerRaw struct {
	Contract *ZkPoDExchangeCaller // Generic read-only contract binding to access the raw methods on
}

// ZkPoDExchangeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ZkPoDExchangeTransactorRaw struct {
	Contract *ZkPoDExchangeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewZkPoDExchange creates a new instance of ZkPoDExchange, bound to a specific deployed contract.
func NewZkPoDExchange(address common.Address, backend bind.ContractBackend) (*ZkPoDExchange, error) {
	contract, err := bindZkPoDExchange(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchange{ZkPoDExchangeCaller: ZkPoDExchangeCaller{contract: contract}, ZkPoDExchangeTransactor: ZkPoDExchangeTransactor{contract: contract}, ZkPoDExchangeFilterer: ZkPoDExchangeFilterer{contract: contract}}, nil
}

// NewZkPoDExchangeCaller creates a new read-only instance of ZkPoDExchange, bound to a specific deployed contract.
func NewZkPoDExchangeCaller(address common.Address, caller bind.ContractCaller) (*ZkPoDExchangeCaller, error) {
	contract, err := bindZkPoDExchange(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeCaller{contract: contract}, nil
}

// NewZkPoDExchangeTransactor creates a new write-only instance of ZkPoDExchange, bound to a specific deployed contract.
func NewZkPoDExchangeTransactor(address common.Address, transactor bind.ContractTransactor) (*ZkPoDExchangeTransactor, error) {
	contract, err := bindZkPoDExchange(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeTransactor{contract: contract}, nil
}

// NewZkPoDExchangeFilterer creates a new log filterer instance of ZkPoDExchange, bound to a specific deployed contract.
func NewZkPoDExchangeFilterer(address common.Address, filterer bind.ContractFilterer) (*ZkPoDExchangeFilterer, error) {
	contract, err := bindZkPoDExchange(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeFilterer{contract: contract}, nil
}

// bindZkPoDExchange binds a generic wrapper to an already deployed contract.
func bindZkPoDExchange(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ZkPoDExchangeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ZkPoDExchange *ZkPoDExchangeRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ZkPoDExchange.Contract.ZkPoDExchangeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ZkPoDExchange *ZkPoDExchangeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.ZkPoDExchangeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ZkPoDExchange *ZkPoDExchangeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.ZkPoDExchangeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ZkPoDExchange *ZkPoDExchangeCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ZkPoDExchange.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ZkPoDExchange *ZkPoDExchangeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ZkPoDExchange *ZkPoDExchangeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.contract.Transact(opts, method, params...)
}

// BobDeposits is a free data retrieval call binding the contract method 0x98033565.
//
// Solidity: function bobDeposits_(address , address ) constant returns(uint256 value, uint256 unDepositAt, uint256 pendingCnt, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeCaller) BobDeposits(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (struct {
	Value       *big.Int
	UnDepositAt *big.Int
	PendingCnt  *big.Int
	Stat        uint8
}, error) {
	ret := new(struct {
		Value       *big.Int
		UnDepositAt *big.Int
		PendingCnt  *big.Int
		Stat        uint8
	})
	out := ret
	err := _ZkPoDExchange.contract.Call(opts, out, "bobDeposits_", arg0, arg1)
	return *ret, err
}

// BobDeposits is a free data retrieval call binding the contract method 0x98033565.
//
// Solidity: function bobDeposits_(address , address ) constant returns(uint256 value, uint256 unDepositAt, uint256 pendingCnt, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeSession) BobDeposits(arg0 common.Address, arg1 common.Address) (struct {
	Value       *big.Int
	UnDepositAt *big.Int
	PendingCnt  *big.Int
	Stat        uint8
}, error) {
	return _ZkPoDExchange.Contract.BobDeposits(&_ZkPoDExchange.CallOpts, arg0, arg1)
}

// BobDeposits is a free data retrieval call binding the contract method 0x98033565.
//
// Solidity: function bobDeposits_(address , address ) constant returns(uint256 value, uint256 unDepositAt, uint256 pendingCnt, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) BobDeposits(arg0 common.Address, arg1 common.Address) (struct {
	Value       *big.Int
	UnDepositAt *big.Int
	PendingCnt  *big.Int
	Stat        uint8
}, error) {
	return _ZkPoDExchange.Contract.BobDeposits(&_ZkPoDExchange.CallOpts, arg0, arg1)
}

// Bulletins is a free data retrieval call binding the contract method 0xd9417785.
//
// Solidity: function bulletins_(bytes32 ) constant returns(address owner, uint64 size, uint64 s, uint64 n, uint256 sigma_mkl_root, uint256 vrf_meta_digest, uint256 pledge_value, uint256 unDepositAt, uint8 blt_type, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeCaller) Bulletins(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Owner         common.Address
	Size          uint64
	S             uint64
	N             uint64
	SigmaMklRoot  *big.Int
	VrfMetaDigest *big.Int
	PledgeValue   *big.Int
	UnDepositAt   *big.Int
	BltType       uint8
	Stat          uint8
}, error) {
	ret := new(struct {
		Owner         common.Address
		Size          uint64
		S             uint64
		N             uint64
		SigmaMklRoot  *big.Int
		VrfMetaDigest *big.Int
		PledgeValue   *big.Int
		UnDepositAt   *big.Int
		BltType       uint8
		Stat          uint8
	})
	out := ret
	err := _ZkPoDExchange.contract.Call(opts, out, "bulletins_", arg0)
	return *ret, err
}

// Bulletins is a free data retrieval call binding the contract method 0xd9417785.
//
// Solidity: function bulletins_(bytes32 ) constant returns(address owner, uint64 size, uint64 s, uint64 n, uint256 sigma_mkl_root, uint256 vrf_meta_digest, uint256 pledge_value, uint256 unDepositAt, uint8 blt_type, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeSession) Bulletins(arg0 [32]byte) (struct {
	Owner         common.Address
	Size          uint64
	S             uint64
	N             uint64
	SigmaMklRoot  *big.Int
	VrfMetaDigest *big.Int
	PledgeValue   *big.Int
	UnDepositAt   *big.Int
	BltType       uint8
	Stat          uint8
}, error) {
	return _ZkPoDExchange.Contract.Bulletins(&_ZkPoDExchange.CallOpts, arg0)
}

// Bulletins is a free data retrieval call binding the contract method 0xd9417785.
//
// Solidity: function bulletins_(bytes32 ) constant returns(address owner, uint64 size, uint64 s, uint64 n, uint256 sigma_mkl_root, uint256 vrf_meta_digest, uint256 pledge_value, uint256 unDepositAt, uint8 blt_type, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) Bulletins(arg0 [32]byte) (struct {
	Owner         common.Address
	Size          uint64
	S             uint64
	N             uint64
	SigmaMklRoot  *big.Int
	VrfMetaDigest *big.Int
	PledgeValue   *big.Int
	UnDepositAt   *big.Int
	BltType       uint8
	Stat          uint8
}, error) {
	return _ZkPoDExchange.Contract.Bulletins(&_ZkPoDExchange.CallOpts, arg0)
}

// GetRecordAtomicSwap is a free data retrieval call binding the contract method 0x7c7c75a8.
//
// Solidity: function getRecordAtomicSwap(address _a, address _b, uint256 _sid) constant returns(bytes32 seed0, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeCaller) GetRecordAtomicSwap(opts *bind.CallOpts, _a common.Address, _b common.Address, _sid *big.Int) (struct {
	Seed0    [32]byte
	SubmitAt *big.Int
}, error) {
	ret := new(struct {
		Seed0    [32]byte
		SubmitAt *big.Int
	})
	out := ret
	err := _ZkPoDExchange.contract.Call(opts, out, "getRecordAtomicSwap", _a, _b, _sid)
	return *ret, err
}

// GetRecordAtomicSwap is a free data retrieval call binding the contract method 0x7c7c75a8.
//
// Solidity: function getRecordAtomicSwap(address _a, address _b, uint256 _sid) constant returns(bytes32 seed0, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeSession) GetRecordAtomicSwap(_a common.Address, _b common.Address, _sid *big.Int) (struct {
	Seed0    [32]byte
	SubmitAt *big.Int
}, error) {
	return _ZkPoDExchange.Contract.GetRecordAtomicSwap(&_ZkPoDExchange.CallOpts, _a, _b, _sid)
}

// GetRecordAtomicSwap is a free data retrieval call binding the contract method 0x7c7c75a8.
//
// Solidity: function getRecordAtomicSwap(address _a, address _b, uint256 _sid) constant returns(bytes32 seed0, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) GetRecordAtomicSwap(_a common.Address, _b common.Address, _sid *big.Int) (struct {
	Seed0    [32]byte
	SubmitAt *big.Int
}, error) {
	return _ZkPoDExchange.Contract.GetRecordAtomicSwap(&_ZkPoDExchange.CallOpts, _a, _b, _sid)
}

// GetRecordComplaint is a free data retrieval call binding the contract method 0x1f31e95e.
//
// Solidity: function getRecordComplaint(address _a, address _b, uint256 _sid) constant returns(bytes32 seed0, bytes32 seed2, bytes32 k_mkl_root, uint64 count, uint256 price, uint256 expireAt, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeCaller) GetRecordComplaint(opts *bind.CallOpts, _a common.Address, _b common.Address, _sid *big.Int) (struct {
	Seed0    [32]byte
	Seed2    [32]byte
	KMklRoot [32]byte
	Count    uint64
	Price    *big.Int
	ExpireAt *big.Int
	SubmitAt *big.Int
}, error) {
	ret := new(struct {
		Seed0    [32]byte
		Seed2    [32]byte
		KMklRoot [32]byte
		Count    uint64
		Price    *big.Int
		ExpireAt *big.Int
		SubmitAt *big.Int
	})
	out := ret
	err := _ZkPoDExchange.contract.Call(opts, out, "getRecordComplaint", _a, _b, _sid)
	return *ret, err
}

// GetRecordComplaint is a free data retrieval call binding the contract method 0x1f31e95e.
//
// Solidity: function getRecordComplaint(address _a, address _b, uint256 _sid) constant returns(bytes32 seed0, bytes32 seed2, bytes32 k_mkl_root, uint64 count, uint256 price, uint256 expireAt, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeSession) GetRecordComplaint(_a common.Address, _b common.Address, _sid *big.Int) (struct {
	Seed0    [32]byte
	Seed2    [32]byte
	KMklRoot [32]byte
	Count    uint64
	Price    *big.Int
	ExpireAt *big.Int
	SubmitAt *big.Int
}, error) {
	return _ZkPoDExchange.Contract.GetRecordComplaint(&_ZkPoDExchange.CallOpts, _a, _b, _sid)
}

// GetRecordComplaint is a free data retrieval call binding the contract method 0x1f31e95e.
//
// Solidity: function getRecordComplaint(address _a, address _b, uint256 _sid) constant returns(bytes32 seed0, bytes32 seed2, bytes32 k_mkl_root, uint64 count, uint256 price, uint256 expireAt, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) GetRecordComplaint(_a common.Address, _b common.Address, _sid *big.Int) (struct {
	Seed0    [32]byte
	Seed2    [32]byte
	KMklRoot [32]byte
	Count    uint64
	Price    *big.Int
	ExpireAt *big.Int
	SubmitAt *big.Int
}, error) {
	return _ZkPoDExchange.Contract.GetRecordComplaint(&_ZkPoDExchange.CallOpts, _a, _b, _sid)
}

// GetRecordVRF is a free data retrieval call binding the contract method 0xb79881e1.
//
// Solidity: function getRecordVRF(address _a, address _b, uint256 _sid) constant returns(uint256 r, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeCaller) GetRecordVRF(opts *bind.CallOpts, _a common.Address, _b common.Address, _sid *big.Int) (struct {
	R        *big.Int
	SubmitAt *big.Int
}, error) {
	ret := new(struct {
		R        *big.Int
		SubmitAt *big.Int
	})
	out := ret
	err := _ZkPoDExchange.contract.Call(opts, out, "getRecordVRF", _a, _b, _sid)
	return *ret, err
}

// GetRecordVRF is a free data retrieval call binding the contract method 0xb79881e1.
//
// Solidity: function getRecordVRF(address _a, address _b, uint256 _sid) constant returns(uint256 r, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeSession) GetRecordVRF(_a common.Address, _b common.Address, _sid *big.Int) (struct {
	R        *big.Int
	SubmitAt *big.Int
}, error) {
	return _ZkPoDExchange.Contract.GetRecordVRF(&_ZkPoDExchange.CallOpts, _a, _b, _sid)
}

// GetRecordVRF is a free data retrieval call binding the contract method 0xb79881e1.
//
// Solidity: function getRecordVRF(address _a, address _b, uint256 _sid) constant returns(uint256 r, uint256 submitAt)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) GetRecordVRF(_a common.Address, _b common.Address, _sid *big.Int) (struct {
	R        *big.Int
	SubmitAt *big.Int
}, error) {
	return _ZkPoDExchange.Contract.GetRecordVRF(&_ZkPoDExchange.CallOpts, _a, _b, _sid)
}

// GetSessionRecord is a free data retrieval call binding the contract method 0xacff2f99.
//
// Solidity: function getSessionRecord(address _a, address _b, uint256 _sid) constant returns(uint256 submitAt, uint8 mode, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeCaller) GetSessionRecord(opts *bind.CallOpts, _a common.Address, _b common.Address, _sid *big.Int) (struct {
	SubmitAt *big.Int
	Mode     uint8
	Stat     uint8
}, error) {
	ret := new(struct {
		SubmitAt *big.Int
		Mode     uint8
		Stat     uint8
	})
	out := ret
	err := _ZkPoDExchange.contract.Call(opts, out, "getSessionRecord", _a, _b, _sid)
	return *ret, err
}

// GetSessionRecord is a free data retrieval call binding the contract method 0xacff2f99.
//
// Solidity: function getSessionRecord(address _a, address _b, uint256 _sid) constant returns(uint256 submitAt, uint8 mode, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeSession) GetSessionRecord(_a common.Address, _b common.Address, _sid *big.Int) (struct {
	SubmitAt *big.Int
	Mode     uint8
	Stat     uint8
}, error) {
	return _ZkPoDExchange.Contract.GetSessionRecord(&_ZkPoDExchange.CallOpts, _a, _b, _sid)
}

// GetSessionRecord is a free data retrieval call binding the contract method 0xacff2f99.
//
// Solidity: function getSessionRecord(address _a, address _b, uint256 _sid) constant returns(uint256 submitAt, uint8 mode, uint8 stat)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) GetSessionRecord(_a common.Address, _b common.Address, _sid *big.Int) (struct {
	SubmitAt *big.Int
	Mode     uint8
	Stat     uint8
}, error) {
	return _ZkPoDExchange.Contract.GetSessionRecord(&_ZkPoDExchange.CallOpts, _a, _b, _sid)
}

// PublicVar is a free data retrieval call binding the contract method 0x0a6c93e0.
//
// Solidity: function publicVar_() constant returns(address)
func (_ZkPoDExchange *ZkPoDExchangeCaller) PublicVar(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ZkPoDExchange.contract.Call(opts, out, "publicVar_")
	return *ret0, err
}

// PublicVar is a free data retrieval call binding the contract method 0x0a6c93e0.
//
// Solidity: function publicVar_() constant returns(address)
func (_ZkPoDExchange *ZkPoDExchangeSession) PublicVar() (common.Address, error) {
	return _ZkPoDExchange.Contract.PublicVar(&_ZkPoDExchange.CallOpts)
}

// PublicVar is a free data retrieval call binding the contract method 0x0a6c93e0.
//
// Solidity: function publicVar_() constant returns(address)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) PublicVar() (common.Address, error) {
	return _ZkPoDExchange.Contract.PublicVar(&_ZkPoDExchange.CallOpts)
}

// S is a free data retrieval call binding the contract method 0xc6b46963.
//
// Solidity: function s_() constant returns(uint64)
func (_ZkPoDExchange *ZkPoDExchangeCaller) S(opts *bind.CallOpts) (uint64, error) {
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _ZkPoDExchange.contract.Call(opts, out, "s_")
	return *ret0, err
}

// S is a free data retrieval call binding the contract method 0xc6b46963.
//
// Solidity: function s_() constant returns(uint64)
func (_ZkPoDExchange *ZkPoDExchangeSession) S() (uint64, error) {
	return _ZkPoDExchange.Contract.S(&_ZkPoDExchange.CallOpts)
}

// S is a free data retrieval call binding the contract method 0xc6b46963.
//
// Solidity: function s_() constant returns(uint64)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) S() (uint64, error) {
	return _ZkPoDExchange.Contract.S(&_ZkPoDExchange.CallOpts)
}

// T1 is a free data retrieval call binding the contract method 0xfb5343f3.
//
// Solidity: function t1() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeCaller) T1(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ZkPoDExchange.contract.Call(opts, out, "t1")
	return *ret0, err
}

// T1 is a free data retrieval call binding the contract method 0xfb5343f3.
//
// Solidity: function t1() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeSession) T1() (*big.Int, error) {
	return _ZkPoDExchange.Contract.T1(&_ZkPoDExchange.CallOpts)
}

// T1 is a free data retrieval call binding the contract method 0xfb5343f3.
//
// Solidity: function t1() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) T1() (*big.Int, error) {
	return _ZkPoDExchange.Contract.T1(&_ZkPoDExchange.CallOpts)
}

// T3 is a free data retrieval call binding the contract method 0xcfad78b1.
//
// Solidity: function t3() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeCaller) T3(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ZkPoDExchange.contract.Call(opts, out, "t3")
	return *ret0, err
}

// T3 is a free data retrieval call binding the contract method 0xcfad78b1.
//
// Solidity: function t3() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeSession) T3() (*big.Int, error) {
	return _ZkPoDExchange.Contract.T3(&_ZkPoDExchange.CallOpts)
}

// T3 is a free data retrieval call binding the contract method 0xcfad78b1.
//
// Solidity: function t3() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) T3() (*big.Int, error) {
	return _ZkPoDExchange.Contract.T3(&_ZkPoDExchange.CallOpts)
}

// T4 is a free data retrieval call binding the contract method 0xfeae062d.
//
// Solidity: function t4() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeCaller) T4(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ZkPoDExchange.contract.Call(opts, out, "t4")
	return *ret0, err
}

// T4 is a free data retrieval call binding the contract method 0xfeae062d.
//
// Solidity: function t4() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeSession) T4() (*big.Int, error) {
	return _ZkPoDExchange.Contract.T4(&_ZkPoDExchange.CallOpts)
}

// T4 is a free data retrieval call binding the contract method 0xfeae062d.
//
// Solidity: function t4() constant returns(uint256)
func (_ZkPoDExchange *ZkPoDExchangeCallerSession) T4() (*big.Int, error) {
	return _ZkPoDExchange.Contract.T4(&_ZkPoDExchange.CallOpts)
}

// BobDeposit is a paid mutator transaction binding the contract method 0x6bc76fdb.
//
// Solidity: function bobDeposit(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) BobDeposit(opts *bind.TransactOpts, _to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "bobDeposit", _to)
}

// BobDeposit is a paid mutator transaction binding the contract method 0x6bc76fdb.
//
// Solidity: function bobDeposit(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) BobDeposit(_to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.BobDeposit(&_ZkPoDExchange.TransactOpts, _to)
}

// BobDeposit is a paid mutator transaction binding the contract method 0x6bc76fdb.
//
// Solidity: function bobDeposit(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) BobDeposit(_to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.BobDeposit(&_ZkPoDExchange.TransactOpts, _to)
}

// BobUnDeposit is a paid mutator transaction binding the contract method 0x30a84ef3.
//
// Solidity: function bobUnDeposit(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) BobUnDeposit(opts *bind.TransactOpts, _to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "bobUnDeposit", _to)
}

// BobUnDeposit is a paid mutator transaction binding the contract method 0x30a84ef3.
//
// Solidity: function bobUnDeposit(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) BobUnDeposit(_to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.BobUnDeposit(&_ZkPoDExchange.TransactOpts, _to)
}

// BobUnDeposit is a paid mutator transaction binding the contract method 0x30a84ef3.
//
// Solidity: function bobUnDeposit(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) BobUnDeposit(_to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.BobUnDeposit(&_ZkPoDExchange.TransactOpts, _to)
}

// ClaimComplaint is a paid mutator transaction binding the contract method 0xf0e2e4e6.
//
// Solidity: function claimComplaint(address _a, uint256 _sid, uint64 _i, uint64 _j, uint256 _tx, uint256 _ty, bytes32[] _mkl_path, uint64 _sCnt) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) ClaimComplaint(opts *bind.TransactOpts, _a common.Address, _sid *big.Int, _i uint64, _j uint64, _tx *big.Int, _ty *big.Int, _mkl_path [][32]byte, _sCnt uint64) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "claimComplaint", _a, _sid, _i, _j, _tx, _ty, _mkl_path, _sCnt)
}

// ClaimComplaint is a paid mutator transaction binding the contract method 0xf0e2e4e6.
//
// Solidity: function claimComplaint(address _a, uint256 _sid, uint64 _i, uint64 _j, uint256 _tx, uint256 _ty, bytes32[] _mkl_path, uint64 _sCnt) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) ClaimComplaint(_a common.Address, _sid *big.Int, _i uint64, _j uint64, _tx *big.Int, _ty *big.Int, _mkl_path [][32]byte, _sCnt uint64) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.ClaimComplaint(&_ZkPoDExchange.TransactOpts, _a, _sid, _i, _j, _tx, _ty, _mkl_path, _sCnt)
}

// ClaimComplaint is a paid mutator transaction binding the contract method 0xf0e2e4e6.
//
// Solidity: function claimComplaint(address _a, uint256 _sid, uint64 _i, uint64 _j, uint256 _tx, uint256 _ty, bytes32[] _mkl_path, uint64 _sCnt) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) ClaimComplaint(_a common.Address, _sid *big.Int, _i uint64, _j uint64, _tx *big.Int, _ty *big.Int, _mkl_path [][32]byte, _sCnt uint64) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.ClaimComplaint(&_ZkPoDExchange.TransactOpts, _a, _sid, _i, _j, _tx, _ty, _mkl_path, _sCnt)
}

// Publish is a paid mutator transaction binding the contract method 0x56e25912.
//
// Solidity: function publish(uint64 _size, uint64 _s, uint64 _n, uint256 _sigma_mkl_root, uint256 _vrf_meta_digest, uint256 _blt_type) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) Publish(opts *bind.TransactOpts, _size uint64, _s uint64, _n uint64, _sigma_mkl_root *big.Int, _vrf_meta_digest *big.Int, _blt_type *big.Int) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "publish", _size, _s, _n, _sigma_mkl_root, _vrf_meta_digest, _blt_type)
}

// Publish is a paid mutator transaction binding the contract method 0x56e25912.
//
// Solidity: function publish(uint64 _size, uint64 _s, uint64 _n, uint256 _sigma_mkl_root, uint256 _vrf_meta_digest, uint256 _blt_type) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) Publish(_size uint64, _s uint64, _n uint64, _sigma_mkl_root *big.Int, _vrf_meta_digest *big.Int, _blt_type *big.Int) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.Publish(&_ZkPoDExchange.TransactOpts, _size, _s, _n, _sigma_mkl_root, _vrf_meta_digest, _blt_type)
}

// Publish is a paid mutator transaction binding the contract method 0x56e25912.
//
// Solidity: function publish(uint64 _size, uint64 _s, uint64 _n, uint256 _sigma_mkl_root, uint256 _vrf_meta_digest, uint256 _blt_type) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) Publish(_size uint64, _s uint64, _n uint64, _sigma_mkl_root *big.Int, _vrf_meta_digest *big.Int, _blt_type *big.Int) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.Publish(&_ZkPoDExchange.TransactOpts, _size, _s, _n, _sigma_mkl_root, _vrf_meta_digest, _blt_type)
}

// SettleComplaintDeal is a paid mutator transaction binding the contract method 0x629eee0a.
//
// Solidity: function settleComplaintDeal(address _a, address _b, uint256 _sid) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) SettleComplaintDeal(opts *bind.TransactOpts, _a common.Address, _b common.Address, _sid *big.Int) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "settleComplaintDeal", _a, _b, _sid)
}

// SettleComplaintDeal is a paid mutator transaction binding the contract method 0x629eee0a.
//
// Solidity: function settleComplaintDeal(address _a, address _b, uint256 _sid) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) SettleComplaintDeal(_a common.Address, _b common.Address, _sid *big.Int) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.SettleComplaintDeal(&_ZkPoDExchange.TransactOpts, _a, _b, _sid)
}

// SettleComplaintDeal is a paid mutator transaction binding the contract method 0x629eee0a.
//
// Solidity: function settleComplaintDeal(address _a, address _b, uint256 _sid) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) SettleComplaintDeal(_a common.Address, _b common.Address, _sid *big.Int) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.SettleComplaintDeal(&_ZkPoDExchange.TransactOpts, _a, _b, _sid)
}

// SubmitProofAtomicSwap is a paid mutator transaction binding the contract method 0x71c9f5dc.
//
// Solidity: function submitProofAtomicSwap(bytes32 _seed0, uint64 _sCnt, uint256 _sid, address _b, bytes32 _seed2, uint256 _sigma_vw, uint64 _count, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) SubmitProofAtomicSwap(opts *bind.TransactOpts, _seed0 [32]byte, _sCnt uint64, _sid *big.Int, _b common.Address, _seed2 [32]byte, _sigma_vw *big.Int, _count uint64, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "submitProofAtomicSwap", _seed0, _sCnt, _sid, _b, _seed2, _sigma_vw, _count, _price, _expireAt, _sig)
}

// SubmitProofAtomicSwap is a paid mutator transaction binding the contract method 0x71c9f5dc.
//
// Solidity: function submitProofAtomicSwap(bytes32 _seed0, uint64 _sCnt, uint256 _sid, address _b, bytes32 _seed2, uint256 _sigma_vw, uint64 _count, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) SubmitProofAtomicSwap(_seed0 [32]byte, _sCnt uint64, _sid *big.Int, _b common.Address, _seed2 [32]byte, _sigma_vw *big.Int, _count uint64, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.SubmitProofAtomicSwap(&_ZkPoDExchange.TransactOpts, _seed0, _sCnt, _sid, _b, _seed2, _sigma_vw, _count, _price, _expireAt, _sig)
}

// SubmitProofAtomicSwap is a paid mutator transaction binding the contract method 0x71c9f5dc.
//
// Solidity: function submitProofAtomicSwap(bytes32 _seed0, uint64 _sCnt, uint256 _sid, address _b, bytes32 _seed2, uint256 _sigma_vw, uint64 _count, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) SubmitProofAtomicSwap(_seed0 [32]byte, _sCnt uint64, _sid *big.Int, _b common.Address, _seed2 [32]byte, _sigma_vw *big.Int, _count uint64, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.SubmitProofAtomicSwap(&_ZkPoDExchange.TransactOpts, _seed0, _sCnt, _sid, _b, _seed2, _sigma_vw, _count, _price, _expireAt, _sig)
}

// SubmitProofComplaint is a paid mutator transaction binding the contract method 0x13efc6f8.
//
// Solidity: function submitProofComplaint(bytes32 _seed0, uint256 _sid, address _b, bytes32 _seed2, bytes32 _k_mkl_root, uint64 _count, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) SubmitProofComplaint(opts *bind.TransactOpts, _seed0 [32]byte, _sid *big.Int, _b common.Address, _seed2 [32]byte, _k_mkl_root [32]byte, _count uint64, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "submitProofComplaint", _seed0, _sid, _b, _seed2, _k_mkl_root, _count, _price, _expireAt, _sig)
}

// SubmitProofComplaint is a paid mutator transaction binding the contract method 0x13efc6f8.
//
// Solidity: function submitProofComplaint(bytes32 _seed0, uint256 _sid, address _b, bytes32 _seed2, bytes32 _k_mkl_root, uint64 _count, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) SubmitProofComplaint(_seed0 [32]byte, _sid *big.Int, _b common.Address, _seed2 [32]byte, _k_mkl_root [32]byte, _count uint64, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.SubmitProofComplaint(&_ZkPoDExchange.TransactOpts, _seed0, _sid, _b, _seed2, _k_mkl_root, _count, _price, _expireAt, _sig)
}

// SubmitProofComplaint is a paid mutator transaction binding the contract method 0x13efc6f8.
//
// Solidity: function submitProofComplaint(bytes32 _seed0, uint256 _sid, address _b, bytes32 _seed2, bytes32 _k_mkl_root, uint64 _count, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) SubmitProofComplaint(_seed0 [32]byte, _sid *big.Int, _b common.Address, _seed2 [32]byte, _k_mkl_root [32]byte, _count uint64, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.SubmitProofComplaint(&_ZkPoDExchange.TransactOpts, _seed0, _sid, _b, _seed2, _k_mkl_root, _count, _price, _expireAt, _sig)
}

// SubmitProofVRF is a paid mutator transaction binding the contract method 0x50f9b445.
//
// Solidity: function submitProofVRF(uint256 _s_r, uint256 _sid, address _b, uint256[2] _g_exp_r, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) SubmitProofVRF(opts *bind.TransactOpts, _s_r *big.Int, _sid *big.Int, _b common.Address, _g_exp_r [2]*big.Int, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "submitProofVRF", _s_r, _sid, _b, _g_exp_r, _price, _expireAt, _sig)
}

// SubmitProofVRF is a paid mutator transaction binding the contract method 0x50f9b445.
//
// Solidity: function submitProofVRF(uint256 _s_r, uint256 _sid, address _b, uint256[2] _g_exp_r, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) SubmitProofVRF(_s_r *big.Int, _sid *big.Int, _b common.Address, _g_exp_r [2]*big.Int, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.SubmitProofVRF(&_ZkPoDExchange.TransactOpts, _s_r, _sid, _b, _g_exp_r, _price, _expireAt, _sig)
}

// SubmitProofVRF is a paid mutator transaction binding the contract method 0x50f9b445.
//
// Solidity: function submitProofVRF(uint256 _s_r, uint256 _sid, address _b, uint256[2] _g_exp_r, uint256 _price, uint256 _expireAt, bytes _sig) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) SubmitProofVRF(_s_r *big.Int, _sid *big.Int, _b common.Address, _g_exp_r [2]*big.Int, _price *big.Int, _expireAt *big.Int, _sig []byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.SubmitProofVRF(&_ZkPoDExchange.TransactOpts, _s_r, _sid, _b, _g_exp_r, _price, _expireAt, _sig)
}

// UnPublish is a paid mutator transaction binding the contract method 0x52bb0780.
//
// Solidity: function unPublish(bytes32 _bltKey) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) UnPublish(opts *bind.TransactOpts, _bltKey [32]byte) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "unPublish", _bltKey)
}

// UnPublish is a paid mutator transaction binding the contract method 0x52bb0780.
//
// Solidity: function unPublish(bytes32 _bltKey) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) UnPublish(_bltKey [32]byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.UnPublish(&_ZkPoDExchange.TransactOpts, _bltKey)
}

// UnPublish is a paid mutator transaction binding the contract method 0x52bb0780.
//
// Solidity: function unPublish(bytes32 _bltKey) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) UnPublish(_bltKey [32]byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.UnPublish(&_ZkPoDExchange.TransactOpts, _bltKey)
}

// WithdrawA is a paid mutator transaction binding the contract method 0x9ecd2944.
//
// Solidity: function withdrawA(bytes32 _bltKey) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) WithdrawA(opts *bind.TransactOpts, _bltKey [32]byte) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "withdrawA", _bltKey)
}

// WithdrawA is a paid mutator transaction binding the contract method 0x9ecd2944.
//
// Solidity: function withdrawA(bytes32 _bltKey) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) WithdrawA(_bltKey [32]byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.WithdrawA(&_ZkPoDExchange.TransactOpts, _bltKey)
}

// WithdrawA is a paid mutator transaction binding the contract method 0x9ecd2944.
//
// Solidity: function withdrawA(bytes32 _bltKey) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) WithdrawA(_bltKey [32]byte) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.WithdrawA(&_ZkPoDExchange.TransactOpts, _bltKey)
}

// WithdrawB is a paid mutator transaction binding the contract method 0x2927251f.
//
// Solidity: function withdrawB(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactor) WithdrawB(opts *bind.TransactOpts, _to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.contract.Transact(opts, "withdrawB", _to)
}

// WithdrawB is a paid mutator transaction binding the contract method 0x2927251f.
//
// Solidity: function withdrawB(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeSession) WithdrawB(_to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.WithdrawB(&_ZkPoDExchange.TransactOpts, _to)
}

// WithdrawB is a paid mutator transaction binding the contract method 0x2927251f.
//
// Solidity: function withdrawB(address _to) returns()
func (_ZkPoDExchange *ZkPoDExchangeTransactorSession) WithdrawB(_to common.Address) (*types.Transaction, error) {
	return _ZkPoDExchange.Contract.WithdrawB(&_ZkPoDExchange.TransactOpts, _to)
}

// ZkPoDExchangeOnAtomicSwapDealIterator is returned from FilterOnAtomicSwapDeal and is used to iterate over the raw logs and unpacked data for OnAtomicSwapDeal events raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnAtomicSwapDealIterator struct {
	Event *ZkPoDExchangeOnAtomicSwapDeal // Event containing the contract specifics and raw log

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
func (it *ZkPoDExchangeOnAtomicSwapDealIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkPoDExchangeOnAtomicSwapDeal)
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
		it.Event = new(ZkPoDExchangeOnAtomicSwapDeal)
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
func (it *ZkPoDExchangeOnAtomicSwapDealIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkPoDExchangeOnAtomicSwapDealIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkPoDExchangeOnAtomicSwapDeal represents a OnAtomicSwapDeal event raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnAtomicSwapDeal struct {
	A     common.Address
	B     common.Address
	Sid   *big.Int
	Seed0 [32]byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOnAtomicSwapDeal is a free log retrieval operation binding the contract event 0xff2c2f55ea8e89ebf103b88cb5f7806ec36789e2e8ffb192727fa47ccfec5d63.
//
// Solidity: event OnAtomicSwapDeal(address indexed _a, address indexed _b, uint256 indexed _sid, bytes32 _seed0)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) FilterOnAtomicSwapDeal(opts *bind.FilterOpts, _a []common.Address, _b []common.Address, _sid []*big.Int) (*ZkPoDExchangeOnAtomicSwapDealIterator, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.FilterLogs(opts, "OnAtomicSwapDeal", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeOnAtomicSwapDealIterator{contract: _ZkPoDExchange.contract, event: "OnAtomicSwapDeal", logs: logs, sub: sub}, nil
}

// WatchOnAtomicSwapDeal is a free log subscription operation binding the contract event 0xff2c2f55ea8e89ebf103b88cb5f7806ec36789e2e8ffb192727fa47ccfec5d63.
//
// Solidity: event OnAtomicSwapDeal(address indexed _a, address indexed _b, uint256 indexed _sid, bytes32 _seed0)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) WatchOnAtomicSwapDeal(opts *bind.WatchOpts, sink chan<- *ZkPoDExchangeOnAtomicSwapDeal, _a []common.Address, _b []common.Address, _sid []*big.Int) (event.Subscription, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.WatchLogs(opts, "OnAtomicSwapDeal", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkPoDExchangeOnAtomicSwapDeal)
				if err := _ZkPoDExchange.contract.UnpackLog(event, "OnAtomicSwapDeal", log); err != nil {
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

// ZkPoDExchangeOnComplaintClaimIterator is returned from FilterOnComplaintClaim and is used to iterate over the raw logs and unpacked data for OnComplaintClaim events raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnComplaintClaimIterator struct {
	Event *ZkPoDExchangeOnComplaintClaim // Event containing the contract specifics and raw log

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
func (it *ZkPoDExchangeOnComplaintClaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkPoDExchangeOnComplaintClaim)
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
		it.Event = new(ZkPoDExchangeOnComplaintClaim)
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
func (it *ZkPoDExchangeOnComplaintClaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkPoDExchangeOnComplaintClaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkPoDExchangeOnComplaintClaim represents a OnComplaintClaim event raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnComplaintClaim struct {
	A   common.Address
	B   common.Address
	Sid *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterOnComplaintClaim is a free log retrieval operation binding the contract event 0xacba611c3ca0197a61bfbe86ec078a16c436808bdda2c1c248c534a595b1ad90.
//
// Solidity: event OnComplaintClaim(address indexed _a, address indexed _b, uint256 indexed _sid)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) FilterOnComplaintClaim(opts *bind.FilterOpts, _a []common.Address, _b []common.Address, _sid []*big.Int) (*ZkPoDExchangeOnComplaintClaimIterator, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.FilterLogs(opts, "OnComplaintClaim", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeOnComplaintClaimIterator{contract: _ZkPoDExchange.contract, event: "OnComplaintClaim", logs: logs, sub: sub}, nil
}

// WatchOnComplaintClaim is a free log subscription operation binding the contract event 0xacba611c3ca0197a61bfbe86ec078a16c436808bdda2c1c248c534a595b1ad90.
//
// Solidity: event OnComplaintClaim(address indexed _a, address indexed _b, uint256 indexed _sid)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) WatchOnComplaintClaim(opts *bind.WatchOpts, sink chan<- *ZkPoDExchangeOnComplaintClaim, _a []common.Address, _b []common.Address, _sid []*big.Int) (event.Subscription, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.WatchLogs(opts, "OnComplaintClaim", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkPoDExchangeOnComplaintClaim)
				if err := _ZkPoDExchange.contract.UnpackLog(event, "OnComplaintClaim", log); err != nil {
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

// ZkPoDExchangeOnComplaintDealIterator is returned from FilterOnComplaintDeal and is used to iterate over the raw logs and unpacked data for OnComplaintDeal events raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnComplaintDealIterator struct {
	Event *ZkPoDExchangeOnComplaintDeal // Event containing the contract specifics and raw log

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
func (it *ZkPoDExchangeOnComplaintDealIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkPoDExchangeOnComplaintDeal)
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
		it.Event = new(ZkPoDExchangeOnComplaintDeal)
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
func (it *ZkPoDExchangeOnComplaintDealIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkPoDExchangeOnComplaintDealIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkPoDExchangeOnComplaintDeal represents a OnComplaintDeal event raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnComplaintDeal struct {
	A     common.Address
	B     common.Address
	Sid   *big.Int
	Price *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOnComplaintDeal is a free log retrieval operation binding the contract event 0xbb7539d148611cf3f9068b9e92a49185664f507b60523ae71e03216bcfc7e505.
//
// Solidity: event OnComplaintDeal(address indexed _a, address indexed _b, uint256 indexed _sid, uint256 _price)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) FilterOnComplaintDeal(opts *bind.FilterOpts, _a []common.Address, _b []common.Address, _sid []*big.Int) (*ZkPoDExchangeOnComplaintDealIterator, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.FilterLogs(opts, "OnComplaintDeal", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeOnComplaintDealIterator{contract: _ZkPoDExchange.contract, event: "OnComplaintDeal", logs: logs, sub: sub}, nil
}

// WatchOnComplaintDeal is a free log subscription operation binding the contract event 0xbb7539d148611cf3f9068b9e92a49185664f507b60523ae71e03216bcfc7e505.
//
// Solidity: event OnComplaintDeal(address indexed _a, address indexed _b, uint256 indexed _sid, uint256 _price)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) WatchOnComplaintDeal(opts *bind.WatchOpts, sink chan<- *ZkPoDExchangeOnComplaintDeal, _a []common.Address, _b []common.Address, _sid []*big.Int) (event.Subscription, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.WatchLogs(opts, "OnComplaintDeal", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkPoDExchangeOnComplaintDeal)
				if err := _ZkPoDExchange.contract.UnpackLog(event, "OnComplaintDeal", log); err != nil {
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

// ZkPoDExchangeOnComplaintKeyIterator is returned from FilterOnComplaintKey and is used to iterate over the raw logs and unpacked data for OnComplaintKey events raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnComplaintKeyIterator struct {
	Event *ZkPoDExchangeOnComplaintKey // Event containing the contract specifics and raw log

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
func (it *ZkPoDExchangeOnComplaintKeyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkPoDExchangeOnComplaintKey)
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
		it.Event = new(ZkPoDExchangeOnComplaintKey)
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
func (it *ZkPoDExchangeOnComplaintKeyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkPoDExchangeOnComplaintKeyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkPoDExchangeOnComplaintKey represents a OnComplaintKey event raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnComplaintKey struct {
	A     common.Address
	B     common.Address
	Sid   *big.Int
	Seed0 [32]byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOnComplaintKey is a free log retrieval operation binding the contract event 0x85218455b9d63667af97af61458f54558d6907d3e0b0eb772f6d4736b0c3eb52.
//
// Solidity: event OnComplaintKey(address indexed _a, address indexed _b, uint256 indexed _sid, bytes32 _seed0)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) FilterOnComplaintKey(opts *bind.FilterOpts, _a []common.Address, _b []common.Address, _sid []*big.Int) (*ZkPoDExchangeOnComplaintKeyIterator, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.FilterLogs(opts, "OnComplaintKey", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeOnComplaintKeyIterator{contract: _ZkPoDExchange.contract, event: "OnComplaintKey", logs: logs, sub: sub}, nil
}

// WatchOnComplaintKey is a free log subscription operation binding the contract event 0x85218455b9d63667af97af61458f54558d6907d3e0b0eb772f6d4736b0c3eb52.
//
// Solidity: event OnComplaintKey(address indexed _a, address indexed _b, uint256 indexed _sid, bytes32 _seed0)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) WatchOnComplaintKey(opts *bind.WatchOpts, sink chan<- *ZkPoDExchangeOnComplaintKey, _a []common.Address, _b []common.Address, _sid []*big.Int) (event.Subscription, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.WatchLogs(opts, "OnComplaintKey", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkPoDExchangeOnComplaintKey)
				if err := _ZkPoDExchange.contract.UnpackLog(event, "OnComplaintKey", log); err != nil {
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

// ZkPoDExchangeOnDealIterator is returned from FilterOnDeal and is used to iterate over the raw logs and unpacked data for OnDeal events raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnDealIterator struct {
	Event *ZkPoDExchangeOnDeal // Event containing the contract specifics and raw log

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
func (it *ZkPoDExchangeOnDealIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkPoDExchangeOnDeal)
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
		it.Event = new(ZkPoDExchangeOnDeal)
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
func (it *ZkPoDExchangeOnDealIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkPoDExchangeOnDealIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkPoDExchangeOnDeal represents a OnDeal event raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnDeal struct {
	A     common.Address
	B     common.Address
	Sid   *big.Int
	Mode  uint8
	Price *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOnDeal is a free log retrieval operation binding the contract event 0x37a5d6eac0a1d706c2eeea2c832256ebf3253a7b42d5a6cedd60e67f4195449b.
//
// Solidity: event OnDeal(address indexed _a, address indexed _b, uint256 indexed _sid, uint8 _mode, uint256 _price)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) FilterOnDeal(opts *bind.FilterOpts, _a []common.Address, _b []common.Address, _sid []*big.Int) (*ZkPoDExchangeOnDealIterator, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.FilterLogs(opts, "OnDeal", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeOnDealIterator{contract: _ZkPoDExchange.contract, event: "OnDeal", logs: logs, sub: sub}, nil
}

// WatchOnDeal is a free log subscription operation binding the contract event 0x37a5d6eac0a1d706c2eeea2c832256ebf3253a7b42d5a6cedd60e67f4195449b.
//
// Solidity: event OnDeal(address indexed _a, address indexed _b, uint256 indexed _sid, uint8 _mode, uint256 _price)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) WatchOnDeal(opts *bind.WatchOpts, sink chan<- *ZkPoDExchangeOnDeal, _a []common.Address, _b []common.Address, _sid []*big.Int) (event.Subscription, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.WatchLogs(opts, "OnDeal", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkPoDExchangeOnDeal)
				if err := _ZkPoDExchange.contract.UnpackLog(event, "OnDeal", log); err != nil {
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

// ZkPoDExchangeOnVRFDealIterator is returned from FilterOnVRFDeal and is used to iterate over the raw logs and unpacked data for OnVRFDeal events raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnVRFDealIterator struct {
	Event *ZkPoDExchangeOnVRFDeal // Event containing the contract specifics and raw log

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
func (it *ZkPoDExchangeOnVRFDealIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkPoDExchangeOnVRFDeal)
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
		it.Event = new(ZkPoDExchangeOnVRFDeal)
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
func (it *ZkPoDExchangeOnVRFDealIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkPoDExchangeOnVRFDealIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkPoDExchangeOnVRFDeal represents a OnVRFDeal event raised by the ZkPoDExchange contract.
type ZkPoDExchangeOnVRFDeal struct {
	A   common.Address
	B   common.Address
	Sid *big.Int
	R   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterOnVRFDeal is a free log retrieval operation binding the contract event 0x0cf813cbe764040d8f6a8bd61c65524fefa1b75e383c20d3673881e921173a52.
//
// Solidity: event OnVRFDeal(address indexed _a, address indexed _b, uint256 indexed _sid, uint256 _r)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) FilterOnVRFDeal(opts *bind.FilterOpts, _a []common.Address, _b []common.Address, _sid []*big.Int) (*ZkPoDExchangeOnVRFDealIterator, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.FilterLogs(opts, "OnVRFDeal", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return &ZkPoDExchangeOnVRFDealIterator{contract: _ZkPoDExchange.contract, event: "OnVRFDeal", logs: logs, sub: sub}, nil
}

// WatchOnVRFDeal is a free log subscription operation binding the contract event 0x0cf813cbe764040d8f6a8bd61c65524fefa1b75e383c20d3673881e921173a52.
//
// Solidity: event OnVRFDeal(address indexed _a, address indexed _b, uint256 indexed _sid, uint256 _r)
func (_ZkPoDExchange *ZkPoDExchangeFilterer) WatchOnVRFDeal(opts *bind.WatchOpts, sink chan<- *ZkPoDExchangeOnVRFDeal, _a []common.Address, _b []common.Address, _sid []*big.Int) (event.Subscription, error) {

	var _aRule []interface{}
	for _, _aItem := range _a {
		_aRule = append(_aRule, _aItem)
	}
	var _bRule []interface{}
	for _, _bItem := range _b {
		_bRule = append(_bRule, _bItem)
	}
	var _sidRule []interface{}
	for _, _sidItem := range _sid {
		_sidRule = append(_sidRule, _sidItem)
	}

	logs, sub, err := _ZkPoDExchange.contract.WatchLogs(opts, "OnVRFDeal", _aRule, _bRule, _sidRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkPoDExchangeOnVRFDeal)
				if err := _ZkPoDExchange.contract.UnpackLog(event, "OnVRFDeal", log); err != nil {
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
