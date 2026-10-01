package multicall_test

import (
	"context"
	"errors"
	"math/big"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	"github.com/status-im/go-wallet-sdk/pkg/multicall"
	mock_multicall "github.com/status-im/go-wallet-sdk/pkg/multicall/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"
)

// The first request of a run carries the chain block number call after the job calls.
func withBlockNumberCall(calls []multicall3.IMulticall3Call) []multicall3.IMulticall3Call {
	return append(calls[:len(calls):len(calls)], multicall.BuildChainBlockNumberCall())
}

// Nothing answers the chain block number call on a chain that reports its own block number.
func withBlockNumberResult(results []multicall3.IMulticall3Result) []multicall3.IMulticall3Result {
	return append(results[:len(results):len(results)], multicall3.IMulticall3Result{Success: true})
}

func TestRunSync_SingleJob_SingleChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create test data
	calls := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
		{Target: common.HexToAddress("0x2"), CallData: []byte("call2")},
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: []byte("result1")},
		{Success: true, ReturnData: []byte("result2")},
	}
	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}

	// Mock the caller to return expected results
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false, // requireSuccess
			withBlockNumberCall(calls),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult(expectedResults), nil)

	// Create job with call result function
	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	// Run the sync function
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, 10)

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.NoError(t, result.Err)
	assert.Len(t, result.Results, 2)
	assert.Equal(t, expectedResults[0], result.Results[0].Value)
	assert.NoError(t, result.Results[0].Err)
	assert.Equal(t, expectedResults[1], result.Results[1].Value)
	assert.NoError(t, result.Results[1].Err)
	assert.Equal(t, expectedBlockNumber, result.BlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), result.BlockHash)
}

func TestRunSync_MultipleJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create multiple jobs
	calls1 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
	}
	calls2 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x2"), CallData: []byte("call2")},
		{Target: common.HexToAddress("0x3"), CallData: []byte("call3")},
	}

	// Expected combined calls and results
	allCalls := append(calls1, calls2...)
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: []byte("result1")},
		{Success: true, ReturnData: []byte("result2")},
		{Success: true, ReturnData: []byte("result3")},
	}
	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}

	// Mock the caller
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(allCalls),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult(expectedResults), nil)

	// Create jobs with call result functions
	jobs := []multicall.Job{
		{
			Calls: calls1,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
		{
			Calls: calls2,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
	}

	// Run the sync function
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, jobs, atBlock, mockCaller, 10)

	// Verify results
	assert.Len(t, results, 2)

	// Verify results for job 1
	result1 := results[0]
	assert.NoError(t, result1.Err)
	assert.Len(t, result1.Results, 1)
	assert.Equal(t, expectedResults[0], result1.Results[0].Value)
	assert.NoError(t, result1.Results[0].Err)
	assert.Equal(t, expectedBlockNumber, result1.BlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), result1.BlockHash)

	// Verify results for job 2
	result2 := results[1]
	assert.NoError(t, result2.Err)
	assert.Len(t, result2.Results, 2)
	assert.Equal(t, expectedResults[1], result2.Results[0].Value)
	assert.NoError(t, result2.Results[0].Err)
	assert.Equal(t, expectedResults[2], result2.Results[1].Value)
	assert.NoError(t, result2.Results[1].Err)
	assert.Equal(t, expectedBlockNumber, result2.BlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), result2.BlockHash)
}

func TestRunSync_Batching_MultipleChunks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create calls that will be split into multiple chunks
	calls := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
		{Target: common.HexToAddress("0x2"), CallData: []byte("call2")},
		{Target: common.HexToAddress("0x3"), CallData: []byte("call3")},
		{Target: common.HexToAddress("0x4"), CallData: []byte("call4")},
		{Target: common.HexToAddress("0x5"), CallData: []byte("call5")},
	}

	// Expected results for all calls
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: []byte("result1")},
		{Success: true, ReturnData: []byte("result2")},
		{Success: true, ReturnData: []byte("result3")},
		{Success: true, ReturnData: []byte("result4")},
		{Success: true, ReturnData: []byte("result5")},
	}
	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}

	// Mock first chunk (ViewTryBlockAndAggregate)
	// The first chunk leaves room for the chain block number call
	chunk1 := calls[0:1]
	results1 := expectedResults[0:1]
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(chunk1),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult(results1), nil)

	// Mock second chunk (ViewTryAggregate)
	chunk2 := calls[1:3]
	results2 := expectedResults[1:3]
	mockCaller.EXPECT().
		ViewTryAggregate(
			gomock.Any(),
			false,
			chunk2,
		).
		Return(results2, nil)

	// Mock third chunk (ViewTryAggregate)
	chunk3 := calls[3:5]
	results3 := expectedResults[3:5]
	mockCaller.EXPECT().
		ViewTryAggregate(
			gomock.Any(),
			false,
			chunk3,
		).
		Return(results3, nil)

	// Create job with call result function
	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	// Run the sync function with small batch size to force batching
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, 2)

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.NoError(t, result.Err)
	assert.Len(t, result.Results, 5)
	for i, expectedResult := range expectedResults {
		assert.Equal(t, expectedResult, result.Results[i].Value)
		assert.NoError(t, result.Results[i].Err)
	}
	assert.Equal(t, expectedBlockNumber, result.BlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), result.BlockHash)
}

func TestRunSync_ErrorHandling_FirstChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create test data
	calls1 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
	}
	calls2 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x2"), CallData: []byte("call2")},
	}

	// Mock the caller to return an error
	expectedError := errors.New("network error")
	allCalls := append(calls1, calls2...)
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(allCalls),
		).
		Return(nil, [32]byte{}, nil, expectedError)

	// Create jobs with call result functions
	jobs := []multicall.Job{
		{
			Calls: calls1,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
		{
			Calls: calls2,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
	}

	// Run the sync function
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, jobs, atBlock, mockCaller, 10)

	// Verify both jobs received the error
	assert.Len(t, results, 2)

	result1 := results[0]
	assert.Equal(t, expectedError, result1.Err)
	assert.Nil(t, result1.Results)
	assert.Nil(t, result1.BlockNumber)
	assert.Equal(t, common.Hash{}, result1.BlockHash)

	result2 := results[1]
	assert.Equal(t, expectedError, result2.Err)
	assert.Nil(t, result2.Results)
	assert.Nil(t, result2.BlockNumber)
	assert.Equal(t, common.Hash{}, result2.BlockHash)
}

func TestRunSync_ErrorHandling_SubsequentChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create calls that will be split into multiple chunks
	calls := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
		{Target: common.HexToAddress("0x2"), CallData: []byte("call2")},
		{Target: common.HexToAddress("0x3"), CallData: []byte("call3")},
	}

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}
	expectedError := errors.New("network error")

	// Mock first chunk (ViewTryBlockAndAggregate) - succeeds
	chunk1 := calls[0:1]
	results1 := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: []byte("result1")},
	}
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(chunk1),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult(results1), nil)

	// Mock second chunk (ViewTryAggregate) - fails
	chunk2 := calls[1:3]
	mockCaller.EXPECT().
		ViewTryAggregate(
			gomock.Any(),
			false,
			chunk2,
		).
		Return(nil, expectedError)

	// Create job with call result function
	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	// Run the sync function with small batch size to force batching
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, 2)

	// Verify job received the error
	assert.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, expectedError, result.Err)
	assert.Nil(t, result.Results)
	assert.Nil(t, result.BlockNumber)
	assert.Equal(t, common.Hash{}, result.BlockHash)
}

func TestRunSync_ErrorHandling_MultipleJobs_SubsequentChunkFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	callsJob0 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
		{Target: common.HexToAddress("0x2"), CallData: []byte("call2")},
	}
	callsJob1 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x3"), CallData: []byte("call3")},
		{Target: common.HexToAddress("0x4"), CallData: []byte("call4")},
	}
	callsJob2 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x5"), CallData: []byte("call5")},
		{Target: common.HexToAddress("0x6"), CallData: []byte("call6")},
	}

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}
	expectedError := errors.New("network error")

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(callsJob0[0:1]),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult([]multicall3.IMulticall3Result{
			{Success: true, ReturnData: []byte("result1")},
		}), nil)

	mockCaller.EXPECT().
		ViewTryAggregate(
			gomock.Any(),
			false,
			[]multicall3.IMulticall3Call{callsJob0[1], callsJob1[0]},
		).
		Return([]multicall3.IMulticall3Result{
			{Success: true, ReturnData: []byte("result2")},
			{Success: true, ReturnData: []byte("result3")},
		}, nil)

	mockCaller.EXPECT().
		ViewTryAggregate(
			gomock.Any(),
			false,
			[]multicall3.IMulticall3Call{callsJob1[1], callsJob2[0]},
		).
		Return([]multicall3.IMulticall3Result{
			{Success: true, ReturnData: []byte("result4")},
			{Success: true, ReturnData: []byte("result5")},
		}, nil)

	mockCaller.EXPECT().
		ViewTryAggregate(
			gomock.Any(),
			false,
			callsJob2[1:2],
		).
		Return(nil, expectedError)

	jobs := []multicall.Job{
		{
			Calls: callsJob0,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
		{
			Calls: callsJob1,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
		{
			Calls: callsJob2,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
	}

	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, jobs, atBlock, mockCaller, 2)

	assert.Len(t, results, 3)
	assert.NoError(t, results[0].Err)
	assert.Len(t, results[0].Results, 2)
	assert.Equal(t, expectedBlockNumber, results[0].BlockNumber)

	assert.NoError(t, results[1].Err)
	assert.Len(t, results[1].Results, 2)
	assert.Equal(t, expectedBlockNumber, results[1].BlockNumber)

	assert.Equal(t, expectedError, results[2].Err)
	assert.Nil(t, results[2].Results)
}

func TestRunSync_ChunkRetry_SplitsOnFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	const callCount = 2 * multicall.DefaultMinChunkSize
	calls := make([]multicall3.IMulticall3Call, callCount)
	expectedResults := make([]multicall3.IMulticall3Result, callCount)
	for i := 0; i < callCount; i++ {
		calls[i] = multicall3.IMulticall3Call{
			Target:   common.HexToAddress("0x1"),
			CallData: []byte{byte(i)},
		}
		expectedResults[i] = multicall3.IMulticall3Result{
			Success:    true,
			ReturnData: []byte{byte(i)},
		}
	}

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).
		DoAndReturn(func(_ *bind.CallOpts, _ bool, chunk []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			if len(chunk) >= callCount {
				return nil, [32]byte{}, nil, errors.New("chunk too large")
			}
			results := make([]multicall3.IMulticall3Result, len(chunk))
			for i, call := range chunk {
				results[i] = multicall3.IMulticall3Result{
					Success:    true,
					ReturnData: call.CallData,
				}
			}
			return expectedBlockNumber, expectedBlockHash, results, nil
		}).
		AnyTimes()

	mockCaller.EXPECT().
		ViewTryAggregate(gomock.Any(), false, gomock.Any()).
		DoAndReturn(func(_ *bind.CallOpts, _ bool, chunk []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error) {
			if len(chunk) >= callCount {
				return nil, errors.New("chunk too large")
			}
			results := make([]multicall3.IMulticall3Result, len(chunk))
			for i, call := range chunk {
				results[i] = multicall3.IMulticall3Result{
					Success:    true,
					ReturnData: call.CallData,
				}
			}
			return results, nil
		}).
		AnyTimes()

	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, callCount)

	assert.Len(t, results, 1)
	result := results[0]
	assert.NoError(t, result.Err)
	assert.Len(t, result.Results, callCount)
	for i, callResult := range result.Results {
		assert.NoError(t, callResult.Err)
		parsed, ok := callResult.Value.(multicall3.IMulticall3Result)
		assert.True(t, ok)
		assert.Equal(t, expectedResults[i].ReturnData, parsed.ReturnData)
	}
}

func TestRunSync_ChunkRetry_MinSizeFailurePropagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	const callCount = multicall.DefaultMinChunkSize
	calls := make([]multicall3.IMulticall3Call, callCount)
	for i := 0; i < callCount; i++ {
		calls[i] = multicall3.IMulticall3Call{
			Target:   common.HexToAddress("0x1"),
			CallData: []byte{byte(i)},
		}
	}

	expectedError := errors.New("rpc unavailable")
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(gomock.Any(), false, withBlockNumberCall(calls[:callCount-1])).
		Return(nil, [32]byte{}, nil, expectedError)

	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, callCount)

	assert.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, expectedError, result.Err)
	assert.Nil(t, result.Results)
}

func TestRunSync_ChunkRetry_ContextCanceledNoSplit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	callCount := 2 * multicall.DefaultMinChunkSize
	calls := make([]multicall3.IMulticall3Call, callCount)
	for i := 0; i < callCount; i++ {
		calls[i] = multicall3.IMulticall3Call{
			Target:   common.HexToAddress("0x1"),
			CallData: []byte{byte(i)},
		}
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).
		Return(nil, [32]byte{}, nil, context.Canceled).
		Times(1)

	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	results := multicall.RunSync(context.Background(), []multicall.Job{job}, big.NewInt(12345), mockCaller, callCount)

	assert.Len(t, results, 1)
	assert.ErrorIs(t, results[0].Err, context.Canceled)
}

func TestRunSync_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create test data
	calls := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Mock the caller to return context cancellation error
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(calls),
		).
		DoAndReturn(func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			cancel() // Cancel the context during the call
			return nil, [32]byte{}, nil, context.Canceled
		})

	// Create job with call result function
	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	// Run the sync function
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, 10)

	// Verify job received the error
	assert.Len(t, results, 1)
	result := results[0]
	assert.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "context canceled")
}

func TestRunSync_EmptyJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Run the sync function with empty jobs
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{}, atBlock, mockCaller, 10)

	// Verify no results
	assert.Empty(t, results)
}

func TestRunSync_EmptyJobCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create job with empty calls
	emptyCalls := []multicall3.IMulticall3Call{}
	job := multicall.Job{
		Calls: emptyCalls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	// Run the sync function
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, 10)

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.NoError(t, result.Err)
	assert.Empty(t, result.Results)
	assert.Nil(t, result.BlockNumber)
	assert.Equal(t, common.Hash{}, result.BlockHash)
}

func TestRunSync_RequireSuccessFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create test data
	calls := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
	}

	expectedResults := []multicall3.IMulticall3Result{
		{Success: false, ReturnData: []byte("error")}, // Failed call
	}
	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}

	// Mock the caller - verify requireSuccess is false
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false, // This should be false
			withBlockNumberCall(calls),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult(expectedResults), nil)

	// Create job with call result function
	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return result, nil
		},
	}

	// Run the sync function
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, 10)

	// Verify results (even with failed individual calls)
	assert.Len(t, results, 1)
	result := results[0]
	assert.NoError(t, result.Err)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, expectedResults[0], result.Results[0].Value)
	assert.NoError(t, result.Results[0].Err)
	assert.Equal(t, expectedBlockNumber, result.BlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), result.BlockHash)
}

func TestProcessJobRunners_AsyncExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create test data
	calls1 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
	}
	calls2 := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x2"), CallData: []byte("call2")},
	}

	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: []byte("result1")},
		{Success: true, ReturnData: []byte("result2")},
	}
	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}

	// Mock the caller
	allCalls := append(calls1, calls2...)
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(allCalls),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult(expectedResults), nil)

	jobs := []multicall.Job{
		{
			Calls: calls1,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
		{
			Calls: calls2,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return result, nil
			},
		},
	}

	// Run ProcessJobRunners
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	resultsCh := multicall.RunAsync(ctx, jobs, atBlock, mockCaller, 10)

	// Verify results
	processedJob1 := false
	processedJob2 := false
	for result := range resultsCh {
		switch result.JobIdx {
		case 0:
			processedJob1 = true
			assert.NoError(t, result.JobResult.Err)
			assert.Len(t, result.JobResult.Results, 1)
			assert.Equal(t, expectedResults[0], result.JobResult.Results[0].Value)
			assert.NoError(t, result.JobResult.Results[0].Err)
			assert.Equal(t, expectedBlockNumber, result.JobResult.BlockNumber)
			assert.Equal(t, common.Hash(expectedBlockHash), result.JobResult.BlockHash)
		case 1:
			processedJob2 = true
			assert.NoError(t, result.JobResult.Err)
			assert.Len(t, result.JobResult.Results, 1)
			assert.Equal(t, expectedResults[1], result.JobResult.Results[0].Value)
			assert.NoError(t, result.JobResult.Results[0].Err)
			assert.Equal(t, expectedBlockNumber, result.JobResult.BlockNumber)
			assert.Equal(t, common.Hash(expectedBlockHash), result.JobResult.BlockHash)
		}
	}

	assert.True(t, processedJob1, "Job 1 not processed")
	assert.True(t, processedJob2, "Job 2 not processed")
}

func TestRunSync_NilCallResultFn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create test data
	calls := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: []byte("result1")},
	}
	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}

	// Mock the caller to return expected results
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(calls),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult(expectedResults), nil)

	// Create job with nil CallResultFn
	job := multicall.Job{
		Calls:        calls,
		CallResultFn: nil, // This should cause an error
	}

	// Run the sync function
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, 10)

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "call result function is nil")
	assert.Nil(t, result.Results)
	assert.Equal(t, expectedBlockNumber, result.BlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), result.BlockHash)
}

func TestRunSync_DataTransformation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Create test data
	calls := []multicall3.IMulticall3Call{
		{Target: common.HexToAddress("0x1"), CallData: []byte("call1")},
		{Target: common.HexToAddress("0x2"), CallData: []byte("call2")},
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: []byte("result1")},
		{Success: false, ReturnData: []byte("error2")},
	}
	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4}

	// Mock the caller to return expected results
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			withBlockNumberCall(calls),
		).
		Return(expectedBlockNumber, expectedBlockHash, withBlockNumberResult(expectedResults), nil)

	// Create job with a transformation function that converts results to strings
	job := multicall.Job{
		Calls: calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			if !result.Success {
				return nil, errors.New("call failed: " + string(result.ReturnData))
			}
			// Transform successful results to a custom format
			return map[string]interface{}{
				"success":    result.Success,
				"data":       string(result.ReturnData),
				"dataLength": len(result.ReturnData),
			}, nil
		},
	}

	// Run the sync function
	ctx := context.Background()
	atBlock := big.NewInt(12345)
	results := multicall.RunSync(ctx, []multicall.Job{job}, atBlock, mockCaller, 10)

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.NoError(t, result.Err)
	assert.Len(t, result.Results, 2)

	// Verify first call result (successful)
	firstResult := result.Results[0]
	assert.NoError(t, firstResult.Err)
	assert.NotNil(t, firstResult.Value)

	// Check the transformed data structure
	transformedData, ok := firstResult.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, transformedData["success"])
	assert.Equal(t, "result1", transformedData["data"])
	assert.Equal(t, 7, transformedData["dataLength"])

	// Verify second call result (failed)
	secondResult := result.Results[1]
	assert.Error(t, secondResult.Err)
	assert.Contains(t, secondResult.Err.Error(), "call failed: error2")
	assert.Nil(t, secondResult.Value)

	assert.Equal(t, expectedBlockNumber, result.BlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), result.BlockHash)
}

// A chain as the runner sees it: what its Multicall3 reports as the block number,
// and what answers the chain block number call (nil when nothing does).
type fakeChain struct {
	reported     *big.Int
	arbSys       *multicall3.IMulticall3Result
	failFrom     int // requests of this many calls or more fail
	readAt       []*big.Int
	firstChunks  [][]multicall3.IMulticall3Call
	requestSizes []int
}

func (c *fakeChain) answer(chunk []multicall3.IMulticall3Call) []multicall3.IMulticall3Result {
	probe := multicall.BuildChainBlockNumberCall()
	results := make([]multicall3.IMulticall3Result, len(chunk))
	for i, call := range chunk {
		if call.Target == probe.Target {
			if c.arbSys != nil {
				results[i] = *c.arbSys
			} else {
				results[i] = multicall3.IMulticall3Result{Success: true} // no code at the address
			}
			continue
		}
		results[i] = multicall3.IMulticall3Result{Success: true, ReturnData: call.CallData}
	}
	return results
}

func (c *fakeChain) expect(mockCaller *mock_multicall.MockCaller) {
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).
		DoAndReturn(func(opts *bind.CallOpts, _ bool, chunk []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			if c.failFrom > 0 && len(chunk) >= c.failFrom {
				return nil, [32]byte{}, nil, errors.New("chunk too large")
			}
			c.readAt = append(c.readAt, opts.BlockNumber)
			c.firstChunks = append(c.firstChunks, chunk)
			c.requestSizes = append(c.requestSizes, len(chunk))
			return c.reported, [32]byte{}, c.answer(chunk), nil
		}).
		AnyTimes()
	mockCaller.EXPECT().
		ViewTryAggregate(gomock.Any(), false, gomock.Any()).
		DoAndReturn(func(opts *bind.CallOpts, _ bool, chunk []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error) {
			if c.failFrom > 0 && len(chunk) >= c.failFrom {
				return nil, errors.New("chunk too large")
			}
			c.readAt = append(c.readAt, opts.BlockNumber)
			c.requestSizes = append(c.requestSizes, len(chunk))
			return c.answer(chunk), nil
		}).
		AnyTimes()
}

func blockNumberResult(blockNumber int64) *multicall3.IMulticall3Result {
	return &multicall3.IMulticall3Result{Success: true, ReturnData: common.LeftPadBytes(big.NewInt(blockNumber).Bytes(), 32)}
}

func jobOfCalls(count int) multicall.Job {
	job := multicall.Job{CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
		return result, nil
	}}
	for i := 0; i < count; i++ {
		job.Calls = append(job.Calls, multicall3.IMulticall3Call{Target: common.HexToAddress("0x1"), CallData: []byte{byte(i)}})
	}
	return job
}

func requireJobAnswered(t *testing.T, result multicall.JobResult, job multicall.Job) {
	t.Helper()
	require.NoError(t, result.Err)
	require.Len(t, result.Results, len(job.Calls), "the chain block number call is not part of the job results")
	for i, call := range job.Calls {
		assert.Equal(t, call.CallData, result.Results[i].Value.(multicall3.IMulticall3Result).ReturnData)
	}
}

const (
	l1BlockNumber    = 26_096_991 // what Multicall3 reports on an Arbitrum-stack chain
	chainBlockNumber = 77_317_482 // the block of the chain itself
)

func TestRunSync_ChainReportingL1BlockNumber_LaterChunksReadTheBlockOfTheFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCaller := mock_multicall.NewMockCaller(ctrl)
	chain := &fakeChain{reported: big.NewInt(l1BlockNumber), arbSys: blockNumberResult(chainBlockNumber)}
	chain.expect(mockCaller)

	job := jobOfCalls(5)
	results := multicall.RunSync(context.Background(), []multicall.Job{job}, nil, mockCaller, 2)

	require.Len(t, results, 1)
	requireJobAnswered(t, results[0], job)
	require.Len(t, chain.readAt, 3)
	assert.Nil(t, chain.readAt[0], "the first chunk is read at the latest block")
	for _, blockNumber := range chain.readAt[1:] {
		require.NotNil(t, blockNumber)
		assert.Equal(t, int64(chainBlockNumber), blockNumber.Int64())
	}
	assert.Equal(t, int64(chainBlockNumber), results[0].BlockNumber.Int64())
}

func TestRunSync_ChainReportingOwnBlockNumber_LaterChunksReadTheReportedBlock(t *testing.T) {
	answers := map[string]*multicall3.IMulticall3Result{
		"nothing at the address":    nil,
		"the call fails":            {Success: false},
		"an answer of another size": {Success: true, ReturnData: []byte{1, 2, 3, 4}},
	}
	for name, arbSys := range answers {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockCaller := mock_multicall.NewMockCaller(ctrl)
			chain := &fakeChain{reported: big.NewInt(12345), arbSys: arbSys}
			chain.expect(mockCaller)

			job := jobOfCalls(5)
			results := multicall.RunSync(context.Background(), []multicall.Job{job}, nil, mockCaller, 2)

			require.Len(t, results, 1)
			requireJobAnswered(t, results[0], job)
			require.Len(t, chain.readAt, 3)
			for _, blockNumber := range chain.readAt[1:] {
				require.NotNil(t, blockNumber)
				assert.Equal(t, int64(12345), blockNumber.Int64())
			}
			assert.Equal(t, int64(12345), results[0].BlockNumber.Int64())
		})
	}
}

func TestRunSync_ChainBlockNumberCallIsSentOnceWithTheFirstChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCaller := mock_multicall.NewMockCaller(ctrl)
	chain := &fakeChain{reported: big.NewInt(12345)}
	chain.expect(mockCaller)

	job := jobOfCalls(5)
	atBlock := big.NewInt(12345)
	multicall.RunSync(context.Background(), []multicall.Job{job}, atBlock, mockCaller, 2)

	require.Len(t, chain.firstChunks, 1)
	assert.Equal(t, append(job.Calls[0:1:1], multicall.BuildChainBlockNumberCall()), chain.firstChunks[0])
	assert.Equal(t, atBlock, chain.readAt[0], "the first chunk is read at the requested block")
}

func TestRunSync_NoRequestExceedsTheBatchSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCaller := mock_multicall.NewMockCaller(ctrl)
	chain := &fakeChain{reported: big.NewInt(12345)}
	chain.expect(mockCaller)

	job := jobOfCalls(7)
	results := multicall.RunSync(context.Background(), []multicall.Job{job}, nil, mockCaller, 3)

	require.Len(t, results, 1)
	requireJobAnswered(t, results[0], job)
	// 2 calls + the chain block number call, then 3, then 2
	assert.Equal(t, []int{3, 3, 2}, chain.requestSizes)
}

func TestRunSync_FirstRequestAnsweredWithAnotherNumberOfResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCaller := mock_multicall.NewMockCaller(ctrl)
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).
		Return(big.NewInt(12345), [32]byte{}, []multicall3.IMulticall3Result{{Success: true}}, nil)

	results := multicall.RunSync(context.Background(), []multicall.Job{jobOfCalls(2)}, nil, mockCaller, 10)

	require.Len(t, results, 1)
	assert.ErrorContains(t, results[0].Err, "expected 3 call results, got 1")
}

func TestBuildChainBlockNumberCall_OwnsItsCallData(t *testing.T) {
	call := multicall.BuildChainBlockNumberCall()
	expected := slices.Clone(call.CallData)

	call.CallData[0] = 0

	assert.Equal(t, expected, multicall.BuildChainBlockNumberCall().CallData)
}

func TestRunSync_ChunkRetry_ChainReportingL1BlockNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCaller := mock_multicall.NewMockCaller(ctrl)

	const callCount = 2 * multicall.DefaultMinChunkSize
	chain := &fakeChain{reported: big.NewInt(l1BlockNumber), arbSys: blockNumberResult(chainBlockNumber), failFrom: callCount}
	chain.expect(mockCaller)

	job := jobOfCalls(callCount)
	results := multicall.RunSync(context.Background(), []multicall.Job{job}, nil, mockCaller, callCount)

	require.Len(t, results, 1)
	requireJobAnswered(t, results[0], job)
	// the first request (999 calls + the chain block number call) fails and is
	// split in two halves; the last call goes in a request of its own
	require.Len(t, chain.readAt, 3)
	assert.Nil(t, chain.readAt[0], "the first half is read at the latest block")
	for _, blockNumber := range chain.readAt[1:] {
		require.NotNil(t, blockNumber)
		assert.Equal(t, int64(chainBlockNumber), blockNumber.Int64(), "the rest is read at the block of the first half")
	}
	assert.Equal(t, int64(chainBlockNumber), results[0].BlockNumber.Int64())
}
