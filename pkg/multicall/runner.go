package multicall

//go:generate mockgen -destination=mock/caller.go . Caller

import (
	"context"
	"errors"
	"math/big"
	"slices"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

const DefaultMinChunkSize = 500

type Caller interface {
	ViewTryBlockAndAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error)
	ViewTryAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error)
}

type CallResult struct {
	Value any
	Err   error
}
type JobResult struct {
	Results     []CallResult
	Err         error
	BlockNumber *big.Int
	BlockHash   common.Hash
}

type JobsResult struct {
	JobIdx    int
	JobResult JobResult
}

type Job struct {
	Calls        []multicall3.IMulticall3Call
	CallResultFn func(multicall3.IMulticall3Result) (any, error)
}

// Collects all jobs and runs them in batches in a blocking manner.
// Once finished, returns a JobResult for each job.
// The output JobResult index matches the input Job index.
func RunSync(ctx context.Context, jobs []Job, atBlock *big.Int, caller Caller, batchsize int) []JobResult {
	resultsCh := RunAsync(ctx, jobs, atBlock, caller, batchsize)

	results := make([]JobResult, len(jobs))
	for result := range resultsCh {
		results[result.JobIdx] = result.JobResult
	}

	return results
}

// Collects all jobs and runs them in batches in a non-blocking manner.
// Returns immediately with a channel where a single JobResult will be sent for each job.
// The received JobResult index matches the input Job index.
// The channel is closed when all results have been sent.
func RunAsync(ctx context.Context, jobs []Job, atBlock *big.Int, caller Caller, batchsize int) <-chan JobsResult {
	resultsCh := make(chan JobsResult, len(jobs))

	go func() {
		defer func() {
			close(resultsCh)
		}()

		ProcessJobs(ctx, jobs, resultsCh, atBlock, caller, batchsize)
	}()
	return resultsCh
}

// ArbSys precompile of Arbitrum-stack chains (Arbitrum, Robinhood).
var arbSysAddress = common.HexToAddress("0x0000000000000000000000000000000000000064")

// BuildChainBlockNumberCall builds the call the runner adds to the first request
// of a run to learn the number of the block the request is read at.
//
// Multicall3 reports block.number, which on Arbitrum-stack chains is the L1 block
// number. Their ArbSys precompile reports the chain's own block number. On other
// chains nothing answers at that address and the Multicall3 number is used.
func BuildChainBlockNumberCall() multicall3.IMulticall3Call {
	return multicall3.IMulticall3Call{
		Target:   arbSysAddress,
		CallData: []byte{0xa3, 0xb1, 0xb3, 0x1d}, // arbBlockNumber()
	}
}

// failedChunkResults answers every call of a request answered with another
// number of results than calls: the results cannot be matched to the calls.
func failedChunkResults(calls, expected, got int) []multicall3.IMulticall3Result {
	reason := []byte("expected " + strconv.Itoa(expected) + " call results, got " + strconv.Itoa(got))
	results := make([]multicall3.IMulticall3Result, calls)
	for i := range results {
		results[i] = multicall3.IMulticall3Result{Success: false, ReturnData: reason}
	}
	return results
}

func chainBlockNumber(reported *big.Int, result multicall3.IMulticall3Result) *big.Int {
	if !result.Success || len(result.ReturnData) != 32 {
		return reported
	}
	return new(big.Int).SetBytes(result.ReturnData)
}

// Runs the first request of a run: the calls plus the chain block number call.
// Returns the number of the block the request was read at and the results of the calls.
func tryBlockAndAggregate(
	ctx context.Context,
	caller Caller,
	atBlock *big.Int,
	requireSuccess bool,
	calls []multicall3.IMulticall3Call,
) (*big.Int, common.Hash, []multicall3.IMulticall3Result, error) {
	callsAndBlockNumber := make([]multicall3.IMulticall3Call, 0, len(calls)+1)
	callsAndBlockNumber = append(callsAndBlockNumber, calls...)
	callsAndBlockNumber = append(callsAndBlockNumber, BuildChainBlockNumberCall())

	reported, blockHash, results, err := caller.ViewTryBlockAndAggregate(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: atBlock,
	}, requireSuccess, callsAndBlockNumber)
	if err != nil {
		return nil, common.Hash{}, nil, err
	}
	if len(results) != len(callsAndBlockNumber) {
		// The chain block number call cannot be told apart either: ask for it alone.
		blockNumber, blockHash, err := readChainBlockNumber(ctx, caller, atBlock)
		if err != nil {
			return nil, common.Hash{}, nil, err
		}
		return blockNumber, blockHash, failedChunkResults(len(calls), len(callsAndBlockNumber), len(results)), nil
	}
	return chainBlockNumber(reported, results[len(calls)]), blockHash, results[:len(calls)], nil
}

func readChainBlockNumber(ctx context.Context, caller Caller, atBlock *big.Int) (*big.Int, common.Hash, error) {
	reported, blockHash, results, err := caller.ViewTryBlockAndAggregate(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: atBlock,
	}, false, []multicall3.IMulticall3Call{BuildChainBlockNumberCall()})
	if err != nil {
		return nil, common.Hash{}, err
	}
	if len(results) != 1 {
		return nil, common.Hash{}, errors.New("expected 1 call result, got " + strconv.Itoa(len(results)))
	}
	return chainBlockNumber(reported, results[0]), blockHash, nil
}

// Splits the calls into requests of at most batchsize calls. The first request
// also carries the chain block number call, so its chunk is one call shorter
// (but never empty: with a batchsize of 1 the first request has two calls).
func splitIntoChunks(calls []multicall3.IMulticall3Call, batchsize int) [][]multicall3.IMulticall3Call {
	first := min(max(batchsize-1, 1), len(calls))
	chunks := [][]multicall3.IMulticall3Call{calls[:first:first]}
	for chunk := range slices.Chunk(calls[first:], batchsize) {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func executeChunkWithRetry(
	ctx context.Context,
	caller Caller,
	atBlock, blockNumber *big.Int,
	requireSuccess bool,
	minChunkSize int,
	calls []multicall3.IMulticall3Call,
) (*big.Int, common.Hash, []multicall3.IMulticall3Result, error) {
	if len(calls) == 0 {
		return blockNumber, common.Hash{}, nil, nil
	}

	var (
		bn      = blockNumber
		bh      common.Hash
		results []multicall3.IMulticall3Result
		err     error
	)
	if blockNumber == nil {
		bn, bh, results, err = tryBlockAndAggregate(ctx, caller, atBlock, requireSuccess, calls)
	} else {
		results, err = caller.ViewTryAggregate(&bind.CallOpts{
			Context:     ctx,
			BlockNumber: blockNumber,
		}, requireSuccess, calls)
		if err == nil && len(results) != len(calls) {
			results = failedChunkResults(len(calls), len(calls), len(results))
		}
	}
	if err == nil {
		return bn, bh, results, nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, common.Hash{}, nil, err
	}
	requestSize := len(calls)
	if blockNumber == nil {
		requestSize++ // the chain block number call
	}
	if requestSize <= minChunkSize || len(calls) < 2 {
		return nil, common.Hash{}, nil, err
	}

	mid := len(calls) / 2
	lbn, lbh, left, err := executeChunkWithRetry(ctx, caller, atBlock, blockNumber, requireSuccess, minChunkSize, calls[:mid])
	if err != nil {
		return nil, common.Hash{}, nil, err
	}
	_, _, right, err := executeChunkWithRetry(ctx, caller, atBlock, lbn, requireSuccess, minChunkSize, calls[mid:])
	if err != nil {
		return nil, common.Hash{}, nil, err
	}
	return lbn, lbh, append(left, right...), nil
}

// Collects all jobs and runs them in batches.
// A single JobResult will be sent on each JobRunner's channel,
// as soon as each individual job is finished.
// The first batch is read at atBlock (the latest block when nil) and the following
// ones at the block the first batch was read at, so a run reads a single block.
// JobResult.BlockNumber is that block. No batch exceeds batchsize; the first one
// includes the chain block number call, see BuildChainBlockNumberCall.
func ProcessJobs(ctx context.Context, jobs []Job, resultsCh chan<- JobsResult, atBlock *big.Int, caller Caller, batchsize int) {
	flatCalls := make([]multicall3.IMulticall3Call, 0, len(jobs))
	for _, job := range jobs {
		flatCalls = append(flatCalls, job.Calls...)
	}

	if len(flatCalls) == 0 {
		// No jobs to run, send empty result for each job
		for i := range jobs {
			resultsCh <- JobsResult{
				JobIdx: i,
				JobResult: JobResult{
					Err: nil,
				},
			}
		}
		return
	}

	rawCallResults := make([]multicall3.IMulticall3Result, 0, len(flatCalls))

	var blockNumber *big.Int
	var blockHash common.Hash
	const requireSuccess = false // Don't revert if any individual call fails
	lastProcessedJobIdx := 0

	// Handle errors
	var err error
	defer func() {
		if err == nil || lastProcessedJobIdx >= len(jobs) {
			return
		}
		// Report error to unprocessed jobs
		for i := range jobs[lastProcessedJobIdx:] {
			resultsCh <- JobsResult{
				JobIdx: lastProcessedJobIdx + i,
				JobResult: JobResult{
					Err: err,
				},
			}
		}
	}()

	for _, chunk := range splitIntoChunks(flatCalls, batchsize) {
		var chunkBlockNumber *big.Int
		var chunkBlockHash common.Hash
		var chunkResults []multicall3.IMulticall3Result

		chunkBlockNumber, chunkBlockHash, chunkResults, err = executeChunkWithRetry(
			ctx, caller, atBlock, blockNumber, requireSuccess, DefaultMinChunkSize, chunk,
		)
		if err != nil {
			return
		}

		if blockNumber == nil {
			blockNumber = chunkBlockNumber
			blockHash = chunkBlockHash
		}

		rawCallResults = append(rawCallResults, chunkResults...)

		// Process results for any finished jobs
		for {
			pendingJob := jobs[lastProcessedJobIdx]
			pendingCallCount := len(pendingJob.Calls)
			if len(rawCallResults) < pendingCallCount {
				break
			}

			jobResult := JobResult{
				BlockNumber: blockNumber,
				BlockHash:   blockHash,
			}

			callResultFn := pendingJob.CallResultFn
			if callResultFn == nil {
				jobResult.Err = errors.New("call result function is nil")
			} else {
				results := make([]CallResult, 0, pendingCallCount)
				for _, rawCallResult := range rawCallResults[:pendingCallCount] {
					result, err := callResultFn(rawCallResult)
					results = append(results, CallResult{
						Value: result,
						Err:   err,
					})
				}
				jobResult.Results = results
			}

			resultsCh <- JobsResult{
				JobIdx:    lastProcessedJobIdx,
				JobResult: jobResult,
			}

			rawCallResults = rawCallResults[pendingCallCount:]
			lastProcessedJobIdx++
			if lastProcessedJobIdx >= len(jobs) {
				break
			}
		}
	}
}
