package multistandardfetcher

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	"github.com/status-im/go-wallet-sdk/pkg/multicall"
)

// zeroERC20Balance marks an omitted zero balance; boxing it does not allocate.
type zeroERC20Balance struct{}

func isZeroWord(word []byte) bool {
	for _, b := range word {
		if b != 0 {
			return false
		}
	}
	return true
}

func buildERC20Job(account AccountAddress, contractAddresses []ContractAddress, omitZero bool) multicall.Job {
	job := multicall.Job{
		Calls: make([]multicall3.IMulticall3Call, 0, len(contractAddresses)),
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			if omitZero && result.Success && len(result.ReturnData) == 32 && isZeroWord(result.ReturnData) {
				return zeroERC20Balance{}, nil
			}
			return multicall.ProcessERC20BalanceResult(result)
		},
	}
	for _, contractAddress := range contractAddresses {
		job.Calls = append(job.Calls, multicall.BuildERC20BalanceCall(account, contractAddress))
	}
	return job
}

func processERC20JobResult(account AccountAddress, contractAddresses []ContractAddress, jobResult multicall.JobResult) (result ERC20Result) {
	result = ERC20Result{
		Account: account,
		Results: make(map[ContractAddress]*big.Int),
	}
	if jobResult.Err != nil {
		result.Err = jobResult.Err
		return
	}
	result.AtBlockNumber = jobResult.BlockNumber
	result.AtBlockHash = jobResult.BlockHash

	if len(jobResult.Results) != len(contractAddresses) {
		result.Err = errors.New("expected " + strconv.Itoa(len(contractAddresses)) + " call results, got " + strconv.Itoa(len(jobResult.Results)))
		return
	}

	for i, callResult := range jobResult.Results {
		if callResult.Err != nil {
			result.Failed = append(result.Failed, contractAddresses[i])
			continue
		}
		if _, zero := callResult.Value.(zeroERC20Balance); zero {
			continue
		}
		parsedResult, ok := callResult.Value.(*big.Int)
		if !ok || parsedResult == nil {
			result.Failed = append(result.Failed, contractAddresses[i])
			continue
		}
		result.Results[contractAddresses[i]] = parsedResult
	}

	return
}
