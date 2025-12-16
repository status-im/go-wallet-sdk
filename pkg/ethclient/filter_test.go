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

func TestFilterQuery(t *testing.T) {
	// Test filter query creation and marshaling
	_ = ethereum.FilterQuery{
		BlockHash: &common.Hash{0x01, 0x02, 0x03},
		FromBlock: big.NewInt(1000),
		ToBlock:   big.NewInt(2000),
		Addresses: []common.Address{
			common.HexToAddress("0x1234567890123456789012345678901234567890"),
			common.HexToAddress("0x0987654321098765432109876543210987654321"),
		},
		Topics: [][]common.Hash{
			{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
			{common.HexToHash("0x000000000000000000000000a7d9ddbe1f17865597fbd27ec712455208b6b76d")},
		},
	}

	// Test filter query with real data
	realFilterQuery := ethereum.FilterQuery{
		FromBlock: big.NewInt(436),
		ToBlock:   big.NewInt(436),
		Addresses: []common.Address{
			common.HexToAddress("0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae"),
		},
		Topics: [][]common.Hash{
			{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
		},
	}

	// Test that the query can be marshaled to JSON
	queryJSON, err := json.Marshal(realFilterQuery)
	assert.NoError(t, err)
	assert.NotEmpty(t, queryJSON)

	// Test filter query with nil values
	nilQuery := ethereum.FilterQuery{
		FromBlock: nil,
		ToBlock:   nil,
		Addresses: nil,
		Topics:    nil,
	}

	nilQueryJSON, err := json.Marshal(nilQuery)
	assert.NoError(t, err)
	assert.NotEmpty(t, nilQueryJSON)
}

func TestFilterMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)
	client := ethclient.NewClient(mockRPC)

	// Test EthNewBlockFilter
	filterIDJSON := `"0x1"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_newBlockFilter").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(filterIDJSON), result)
		})

	filterID, err := client.EthNewBlockFilter(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, ethclient.FilterID("0x1"), filterID)

	// Test EthNewFilter
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(436),
		ToBlock:   big.NewInt(500),
		Addresses: []common.Address{
			common.HexToAddress("0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae"),
		},
	}

	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_newFilter", gomock.Any()).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(filterIDJSON), result)
		})

	filterID, err = client.EthNewFilter(context.Background(), query)
	assert.NoError(t, err)
	assert.Equal(t, ethclient.FilterID("0x1"), filterID)

	// Test EthGetFilterChanges
	var logsJSONData = `[{"address":"0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"],"data":"0x0000000000000000000000000000000000000000000000000000000000000000","blockNumber":"0x1b4","transactionHash":"0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b","transactionIndex":"0x0","blockHash":"0xe670ec64341771606e55d6b4ca35a1a6b75ee3d28e5ec2e4b0f1b2b3c4d5e6f7","logIndex":"0x0","removed":false}]`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getFilterChanges", ethclient.FilterID("0x1")).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(logsJSONData), result)
		})

	logs, err := client.EthGetFilterChanges(context.Background(), ethclient.FilterID("0x1"))
	assert.NoError(t, err)
	assert.NotNil(t, logs)

	// Test EthGetFilterLogs
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getFilterLogs", ethclient.FilterID("0x1")).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(logsJSONData), result)
		})

	logs2, err := client.EthGetFilterLogs(context.Background(), ethclient.FilterID("0x1"))
	assert.NoError(t, err)
	assert.NotNil(t, logs2)

	// Test EthUninstallFilter
	uninstallResultJSON := `true`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_uninstallFilter", ethclient.FilterID("0x1")).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(uninstallResultJSON), result)
		})

	uninstalled, err := client.EthUninstallFilter(context.Background(), ethclient.FilterID("0x1"))
	assert.NoError(t, err)
	assert.True(t, uninstalled)
}

func TestEthNewFilterWithBlockHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)
	client := ethclient.NewClient(mockRPC)

	blockHash := common.HexToHash("0xa917fcc721a5465a484e9be17cda0cc5493933dd3bc70c9adbee192cb419c9d7")
	query := ethereum.FilterQuery{
		BlockHash: &blockHash,
		Addresses: []common.Address{
			common.HexToAddress("0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae"),
		},
	}

	filterIDJSON := `"0x2"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_newFilter", gomock.Any()).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(filterIDJSON), result)
		})

	filterID, err := client.EthNewFilter(context.Background(), query)
	assert.NoError(t, err)
	assert.Equal(t, ethclient.FilterID("0x2"), filterID)
}
