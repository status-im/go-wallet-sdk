package multicall

//go:generate mockgen -destination=mock/caller.go . Caller

import (
	"context"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

type Caller interface {
	ViewTryBlockAndAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error)
	ViewTryAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error)
}

type JobResult struct {
	Results     []multicall3.IMulticall3Result
	Err         error
	BlockNumber *big.Int
	BlockHash   common.Hash
}

type JobRunner struct {
	Job      []multicall3.IMulticall3Call
	ResultCh chan<- JobResult
}

// Collects all jobs and runs them in batches in a non-blocking manner.
// Once finished, returns a JobResult for each job.
func RunSync(ctx context.Context, jobs [][]multicall3.IMulticall3Call, atBlock *big.Int, caller Caller, batchsize int) []JobResult {
	jobRunners := make([]JobRunner, 0, len(jobs))
	resultChs := make([]chan JobResult, 0, len(jobs))
	results := make([]JobResult, 0, len(jobs))
	for _, job := range jobs {
		resultCh := make(chan JobResult, 1) // A single result is sent for each job, buffered to avoid blocking
		resultChs = append(resultChs, resultCh)
		jobRunners = append(jobRunners, JobRunner{
			Job:      job,
			ResultCh: resultCh,
		})
	}

	defer func() {
		for _, resultCh := range resultChs {
			close(resultCh)
		}
	}()

	ProcessJobRunners(ctx, jobRunners, atBlock, caller, batchsize)

	for _, resultCh := range resultChs {
		results = append(results, <-resultCh)
	}

	return results
}

// Collects all jobs and runs them in batches.
// A JobResult is sent to each channel as soon as its associated job is finished.
func ProcessJobRunners(ctx context.Context, jobRunners []JobRunner, atBlock *big.Int, caller Caller, batchsize int) {
	flatCalls := make([]multicall3.IMulticall3Call, 0, len(jobRunners))
	for _, jobRunner := range jobRunners {
		flatCalls = append(flatCalls, jobRunner.Job...)
	}

	if len(flatCalls) == 0 {
		// No jobs to run, send empty result to all job runners
		for _, jobRunner := range jobRunners {
			jobRunner.ResultCh <- JobResult{
				Err: nil,
			}
		}
		return
	}

	results := make([]multicall3.IMulticall3Result, 0, len(flatCalls))

	var blockNumber *big.Int
	var blockHash common.Hash
	const requireSuccess = false // Don't revert if any individual call fails
	lastProcessedJobIdx := 0

	// Handle errors
	var err error
	defer func() {
		if err == nil || lastProcessedJobIdx >= len(jobRunners) {
			return
		}
		// Report error to unprocessed jobs
		jobResult := JobResult{
			Err: err,
		}
		for _, job := range jobRunners[lastProcessedJobIdx:] {
			job.ResultCh <- jobResult
		}
	}()

	first := true
	for chunk := range slices.Chunk(flatCalls, batchsize) {
		var chunkResults []multicall3.IMulticall3Result
		if first {
			first = false
			// First chunk, we need to get the block number and hash
			blockNumber, blockHash, chunkResults, err = caller.ViewTryBlockAndAggregate(&bind.CallOpts{
				Context:     ctx,
				BlockNumber: atBlock,
			}, requireSuccess, chunk)
		} else {
			// Subsequent chunks, we use the block number from the previous chunk
			chunkResults, err = caller.ViewTryAggregate(&bind.CallOpts{
				Context:     ctx,
				BlockNumber: blockNumber,
			}, requireSuccess, chunk)
		}
		if err != nil {
			return
		}
		results = append(results, chunkResults...)

		// Process results for any finished jobs
		for {
			pendingCallCount := len(jobRunners[lastProcessedJobIdx].Job)
			if len(results) < pendingCallCount {
				break
			}
			jobResult := JobResult{
				Results:     results[:len(jobRunners[lastProcessedJobIdx].Job)],
				BlockNumber: blockNumber,
				BlockHash:   blockHash,
			}
			jobRunners[lastProcessedJobIdx].ResultCh <- jobResult
			results = results[pendingCallCount:]
			lastProcessedJobIdx++
			if lastProcessedJobIdx >= len(jobRunners) {
				break
			}
		}
	}
}
