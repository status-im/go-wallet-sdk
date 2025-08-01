package ethclient_test

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	mock_ethclient "github.com/status-im/go-wallet-sdk/pkg/ethclient/mock"
)

func TestLineaEstimateGas(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	// Test data
	callMsg := ethereum.CallMsg{
		From:  common.HexToAddress("0xa7d9ddbe1f17865597fbd27ec712455208b6b76d"),
		To:    &common.Address{0xf0, 0x2c, 0x1c, 0x8e, 0x61, 0x14, 0xb1, 0xdb, 0xe8, 0x93, 0x7a, 0x39, 0x26, 0x0b, 0x5b, 0x0a, 0x37, 0x44, 0x32, 0xbb},
		Value: big.NewInt(1000000000000000000),
		Data:  []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}, // "hello"
	}

	// Expected response JSON
	responseJSON := `{
		"baseFeePerGas": "0x59682f00",
		"gasLimit": "0x5208",
		"priorityFeePerGas": "0x3b9aca00"
	}`

	// Set up mock expectation
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "linea_estimateGas", gomock.Any()).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(responseJSON), result)
		})

	// Call the method
	result, err := client.LineaEstimateGas(context.Background(), callMsg)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, big.NewInt(0x59682f00), result.BaseFeePerGas)
	assert.Equal(t, big.NewInt(0x5208), result.GasLimit)
	assert.Equal(t, big.NewInt(0x3b9aca00), result.PriorityFeePerGas)
}
