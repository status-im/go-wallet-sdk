package multicall_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	"github.com/status-im/go-wallet-sdk/pkg/multicall"
	mock_multicall "github.com/status-im/go-wallet-sdk/pkg/multicall/mock"

	"github.com/stretchr/testify/assert"

	"go.uber.org/mock/gomock"
)

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
			calls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

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
			allCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

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
	chunk1 := calls[0:2]
	results1 := expectedResults[0:2]
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			chunk1,
		).
		Return(expectedBlockNumber, expectedBlockHash, results1, nil)

	// Mock second chunk (ViewTryAggregate)
	chunk2 := calls[2:4]
	results2 := expectedResults[2:4]
	mockCaller.EXPECT().
		ViewTryAggregate(
			gomock.Any(),
			false,
			chunk2,
		).
		Return(results2, nil)

	// Mock third chunk (ViewTryAggregate)
	chunk3 := calls[4:5]
	results3 := expectedResults[4:5]
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
			allCalls,
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
	chunk1 := calls[0:2]
	results1 := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: []byte("result1")},
		{Success: true, ReturnData: []byte("result2")},
	}
	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			chunk1,
		).
		Return(expectedBlockNumber, expectedBlockHash, results1, nil)

	// Mock second chunk (ViewTryAggregate) - fails
	chunk2 := calls[2:3]
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
			calls,
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
			calls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

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
			allCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

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
			calls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

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
			calls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

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
