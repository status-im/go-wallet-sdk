package multicall

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

// Call for ERC20 function "balanceOf(owner)"
func BuildERC20BalanceCall(accountAddress common.Address, tokenAddress common.Address) multicall3.IMulticall3Call {
	abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	callData, err := abi.Pack("balanceOf", accountAddress)
	if err != nil {
		panic(err)
	}

	call := multicall3.IMulticall3Call{
		Target:   tokenAddress,
		CallData: callData,
	}

	return call
}

// balanceOf(address) returns a single uint256, i.e. exactly one 32-byte word.
const erc20BalanceReturnDataLen = 32

// ProcessERC20BalanceResult decodes a balanceOf sub-call of a multicall. A
// sub-call that "succeeded" with anything but one uint256 word is reported as
// an error rather than a zero balance: Multicall3 marks a call to an address
// without code as successful with empty return data, which is what a node that
// lacks the contract's state (a lagging or misrouted RPC upstream) answers.
// Decoding that as 0 would replace a real balance with a false zero.
func ProcessERC20BalanceResult(result multicall3.IMulticall3Result) (*big.Int, error) {
	if !result.Success {
		return nil, errors.New(string(result.ReturnData))
	}
	if len(result.ReturnData) != erc20BalanceReturnDataLen {
		return nil, fmt.Errorf("unexpected balanceOf return data length: %d", len(result.ReturnData))
	}
	return new(big.Int).SetBytes(result.ReturnData), nil
}
