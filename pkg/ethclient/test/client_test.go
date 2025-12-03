package ethclient_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	mock_ethclient "github.com/status-im/go-wallet-sdk/pkg/ethclient/mock"
)

func TestRPCCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)
	client := ethclient.NewClient(mockRPC)

	t.Run("successful RPC call", func(t *testing.T) {
		// Test data
		method := "eth_getBalance"
		address := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
		blockNumber := "latest"
		expectedResponse := `"0x1bc16d674ec80000"` // 2 ETH in wei

		// Set up mock expectation
		mockRPC.EXPECT().
			CallContext(
				gomock.Any(),
				gomock.Any(),
				method,
				address,
				blockNumber,
			).
			DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
				// Verify args were passed correctly
				assert.Equal(t, address, args[0])
				assert.Equal(t, blockNumber, args[1])
				// Unmarshal the response into result
				return json.Unmarshal([]byte(expectedResponse), result)
			})

		// Call RPCCall
		var result string
		err := client.RPCCall(context.Background(), &result, method, address, blockNumber)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, "0x1bc16d674ec80000", result)
	})

	t.Run("RPC call with JSON object response", func(t *testing.T) {
		// Test data
		method := "eth_getBlockByNumber"
		blockNumber := "0x1"
		includeTxs := false
		expectedResponse := `{"number":"0x1","hash":"0x123","parentHash":"0x456"}`

		// Set up mock expectation
		mockRPC.EXPECT().
			CallContext(
				gomock.Any(),
				gomock.Any(),
				method,
				blockNumber,
				includeTxs,
			).
			DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
				// Verify args were passed correctly
				assert.Equal(t, blockNumber, args[0])
				assert.Equal(t, includeTxs, args[1])
				// Unmarshal the response into result
				return json.Unmarshal([]byte(expectedResponse), result)
			})

		// Call RPCCall with a map result
		var result map[string]interface{}
		err := client.RPCCall(context.Background(), &result, method, blockNumber, includeTxs)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, "0x1", result["number"])
		assert.Equal(t, "0x123", result["hash"])
		assert.Equal(t, "0x456", result["parentHash"])
	})

	t.Run("RPC call with error", func(t *testing.T) {
		// Test data
		method := "eth_invalidMethod"
		expectedError := errors.New("method not found")

		// Set up mock expectation to return error
		mockRPC.EXPECT().
			CallContext(
				gomock.Any(),
				gomock.Any(),
				method,
			).
			Return(expectedError)

		// Call RPCCall
		var result interface{}
		err := client.RPCCall(context.Background(), &result, method)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("RPC call with no arguments", func(t *testing.T) {
		// Test data
		method := "net_version"
		expectedResponse := `"1"`

		// Set up mock expectation
		mockRPC.EXPECT().
			CallContext(
				gomock.Any(),
				gomock.Any(),
				method,
			).
			DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
				// Verify no args were passed
				assert.Empty(t, args)
				// Unmarshal the response into result
				return json.Unmarshal([]byte(expectedResponse), result)
			})

		// Call RPCCall with no args
		var result string
		err := client.RPCCall(context.Background(), &result, method)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, "1", result)
	})
}
