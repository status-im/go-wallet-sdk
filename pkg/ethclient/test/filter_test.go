package ethclient_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
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
