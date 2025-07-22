package old

import (
	"fmt"
	"math/big"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethParams "github.com/ethereum/go-ethereum/params"
)

const (
	UnknownChainID       uint64 = 0
	EthereumMainnet      uint64 = 1
	EthereumSepolia      uint64 = 11155111
	OptimismMainnet      uint64 = 10
	OptimismSepolia      uint64 = 11155420
	ArbitrumMainnet      uint64 = 42161
	ArbitrumSepolia      uint64 = 421614
	BSCMainnet           uint64 = 56
	BSCTestnet           uint64 = 97
	AnvilMainnet         uint64 = 31337
	BaseMainnet          uint64 = 8453
	BaseSepolia          uint64 = 84532
	StatusNetworkSepolia uint64 = 1660990954
	TestnetChainID       uint64 = 777333
)

var AverageBlockDurationForChain = map[uint64]time.Duration{
	UnknownChainID:       time.Duration(12000) * time.Millisecond,
	EthereumMainnet:      time.Duration(12000) * time.Millisecond,
	OptimismMainnet:      time.Duration(2000) * time.Millisecond,
	ArbitrumMainnet:      time.Duration(250) * time.Millisecond,
	BaseMainnet:          time.Duration(2000) * time.Millisecond,
	BSCMainnet:           time.Duration(3000) * time.Millisecond,
	StatusNetworkSepolia: time.Duration(2000) * time.Millisecond,
}

func GweiToEth(val *big.Float) *big.Float {
	return new(big.Float).Quo(val, big.NewFloat(1000000000))
}

func WeiToGwei(val *big.Int) *big.Float {
	result := new(big.Float)
	result.SetInt(val)

	unit := new(big.Int)
	unit.SetInt64(gethParams.GWei)

	return result.Quo(result, new(big.Float).SetInt(unit))
}

func GetBlockCreationTimeForChain(chainID uint64) time.Duration {
	blockDuration, found := AverageBlockDurationForChain[chainID]
	if !found {
		blockDuration = AverageBlockDurationForChain[UnknownChainID]
	}
	return blockDuration
}

// Special functions to hardcode the nature of some special chains (eg. Status Network), where we cannot deduce EIP-1559 compatibility in a generic way

// IsPartiallyOrFullyGaslessChain returns true if the chain is fully or partially (no base or no priority fee) gasless
func IsPartiallyOrFullyGaslessChain(chainID uint64) bool {
	return chainID == StatusNetworkSepolia
}

// IsPartiallyOrFullyGaslessChainEIP1559Compatible throws an error if the chain is not partially or fully gasless, if it is, returns true if the chain is EIP-1559 compatible
func IsPartiallyOrFullyGaslessChainEIP1559Compatible(chainID uint64) (bool, error) {
	if !IsPartiallyOrFullyGaslessChain(chainID) {
		return false, fmt.Errorf("chain %d is not supposed to be gasless", chainID) // for non-gasless chains, we should not use this function
	}
	return chainID == StatusNetworkSepolia, nil
}

func ZeroAddress() ethCommon.Address {
	return ethCommon.Address{}
}

func ZeroBigIntValue() *big.Int {
	return big.NewInt(0)
}

func ZeroHash() ethCommon.Hash {
	return ethCommon.Hash{}
}

func ToCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}
