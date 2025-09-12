package eventfilter

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestToFilterQueries_QueryCount(t *testing.T) {
	testAddr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	testAddr2 := common.HexToAddress("0x9876543210987654321098765432109876543210")
	addresses := []common.Address{testAddr1, testAddr2}
	fromBlock := big.NewInt(18000000)
	toBlock := big.NewInt(18001000)

	tests := []struct {
		name          string
		config        TransferQueryConfig
		expectedCount int
	}{
		{
			name: "ERC20 Send only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20},
				Direction:     Send,
			},
			expectedCount: 1, // ERC20 Send
		},
		{
			name: "ERC20 Receive only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20},
				Direction:     Receive,
			},
			expectedCount: 1, // ERC20 Receive
		},
		{
			name: "ERC20 Both directions",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20},
				Direction:     Both,
			},
			expectedCount: 2, // ERC20 Send + ERC20 Receive
		},
		{
			name: "ERC721 Send only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC721},
				Direction:     Send,
			},
			expectedCount: 1, // ERC721 Send
		},
		{
			name: "ERC721 Receive only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC721},
				Direction:     Receive,
			},
			expectedCount: 1, // ERC721 Receive
		},
		{
			name: "ERC721 Both directions",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC721},
				Direction:     Both,
			},
			expectedCount: 2, // ERC721 Send + ERC721 Receive
		},
		{
			name: "ERC1155 Send only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC1155},
				Direction:     Send,
			},
			expectedCount: 1, // ERC1155 Send (merged TransferSingle + TransferBatch)
		},
		{
			name: "ERC1155 Receive only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC1155},
				Direction:     Receive,
			},
			expectedCount: 1, // ERC1155 Receive (merged TransferSingle + TransferBatch)
		},
		{
			name: "ERC1155 Both directions",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC1155},
				Direction:     Both,
			},
			expectedCount: 2, // ERC1155 Send + ERC1155 Receive
		},
		{
			name: "ERC20 + ERC721 Send only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC721},
				Direction:     Send,
			},
			expectedCount: 1, // ERC20/ERC721 Send (shared event signature)
		},
		{
			name: "ERC20 + ERC721 Receive only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC721},
				Direction:     Receive,
			},
			expectedCount: 1, // ERC20/ERC721 Receive (shared event signature)
		},
		{
			name: "ERC20 + ERC721 Both directions",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC721},
				Direction:     Both,
			},
			expectedCount: 2, // ERC20/ERC721 Send + ERC20/ERC721 Receive
		},
		{
			name: "ERC20 + ERC1155 Send only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC1155},
				Direction:     Send,
			},
			expectedCount: 2, // ERC20 Send + ERC1155 Send
		},
		{
			name: "ERC20 + ERC1155 Receive only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC1155},
				Direction:     Receive,
			},
			expectedCount: 2, // ERC20 Receive + ERC1155 Receive
		},
		{
			name: "ERC20 + ERC1155 Both directions",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC1155},
				Direction:     Both,
			},
			expectedCount: 3, // ERC20 Send + merged (ERC20 Receive + ERC1155 Send) + ERC1155 Receive
		},
		{
			name: "ERC721 + ERC1155 Send only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC721, TransferTypeERC1155},
				Direction:     Send,
			},
			expectedCount: 2, // ERC721 Send + ERC1155 Send
		},
		{
			name: "ERC721 + ERC1155 Receive only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC721, TransferTypeERC1155},
				Direction:     Receive,
			},
			expectedCount: 2, // ERC721 Receive + ERC1155 Receive
		},
		{
			name: "ERC721 + ERC1155 Both directions",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC721, TransferTypeERC1155},
				Direction:     Both,
			},
			expectedCount: 3, // ERC721 Send + merged (ERC721 Receive + ERC1155 Send) + ERC1155 Receive
		},
		{
			name: "All types Send only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC721, TransferTypeERC1155},
				Direction:     Send,
			},
			expectedCount: 2, // ERC20/ERC721 Send + ERC1155 Send
		},
		{
			name: "All types Receive only",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC721, TransferTypeERC1155},
				Direction:     Receive,
			},
			expectedCount: 2, // ERC20/ERC721 Receive + ERC1155 Receive
		},
		{
			name: "All types Both directions",
			config: TransferQueryConfig{
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				Accounts:      addresses,
				TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC721, TransferTypeERC1155},
				Direction:     Both,
			},
			expectedCount: 3, // ERC20/ERC721 Send + merged (ERC20/ERC721 Receive + ERC1155 Send) + ERC1155 Receive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := tt.config.ToFilterQueries()
			assert.Equal(t, tt.expectedCount, len(queries),
				"Expected %d queries, got %d for config: %+v", tt.expectedCount, len(queries), tt.config)
		})
	}
}

func TestToFilterQueries_QueryStructure(t *testing.T) {
	testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addresses := []common.Address{testAddr}
	fromBlock := big.NewInt(18000000)
	toBlock := big.NewInt(18001000)

	t.Run("ERC20 Send query structure", func(t *testing.T) {
		config := TransferQueryConfig{
			FromBlock:     fromBlock,
			ToBlock:       toBlock,
			Accounts:      addresses,
			TransferTypes: []TransferType{TransferTypeERC20},
			Direction:     Send,
		}

		queries := config.ToFilterQueries()
		assert.Len(t, queries, 1)

		query := queries[0]
		assert.Equal(t, fromBlock, query.FromBlock)
		assert.Equal(t, toBlock, query.ToBlock)
		assert.Len(t, query.Topics, 2)      // [eventSignature, address]
		assert.NotEmpty(t, query.Topics[0]) // Event signature
		assert.NotEmpty(t, query.Topics[1]) // Address topics
	})

	t.Run("ERC20 Receive query structure", func(t *testing.T) {
		config := TransferQueryConfig{
			FromBlock:     fromBlock,
			ToBlock:       toBlock,
			Accounts:      addresses,
			TransferTypes: []TransferType{TransferTypeERC20},
			Direction:     Receive,
		}

		queries := config.ToFilterQueries()
		assert.Len(t, queries, 1)

		query := queries[0]
		assert.Equal(t, fromBlock, query.FromBlock)
		assert.Equal(t, toBlock, query.ToBlock)
		assert.Len(t, query.Topics, 3)      // [eventSignature, {}, address]
		assert.NotEmpty(t, query.Topics[0]) // Event signature
		assert.Empty(t, query.Topics[1])    // Empty (any from address)
		assert.NotEmpty(t, query.Topics[2]) // Address topics
	})

	t.Run("ERC1155 Send query structure", func(t *testing.T) {
		config := TransferQueryConfig{
			FromBlock:     fromBlock,
			ToBlock:       toBlock,
			Accounts:      addresses,
			TransferTypes: []TransferType{TransferTypeERC1155},
			Direction:     Send,
		}

		queries := config.ToFilterQueries()
		assert.Len(t, queries, 1)

		query := queries[0]
		assert.Equal(t, fromBlock, query.FromBlock)
		assert.Equal(t, toBlock, query.ToBlock)
		assert.Len(t, query.Topics, 3)      // [eventSignature, {}, address] (omitted empty last topic)
		assert.NotEmpty(t, query.Topics[0]) // Event signature (TransferSingle + TransferBatch)
		assert.Empty(t, query.Topics[1])    // Empty (any operator)
		assert.NotEmpty(t, query.Topics[2]) // Address topics
	})

	t.Run("ERC1155 Receive query structure", func(t *testing.T) {
		config := TransferQueryConfig{
			FromBlock:     fromBlock,
			ToBlock:       toBlock,
			Accounts:      addresses,
			TransferTypes: []TransferType{TransferTypeERC1155},
			Direction:     Receive,
		}

		queries := config.ToFilterQueries()
		assert.Len(t, queries, 1)

		query := queries[0]
		assert.Equal(t, fromBlock, query.FromBlock)
		assert.Equal(t, toBlock, query.ToBlock)
		assert.Len(t, query.Topics, 4)      // [eventSignature, {}, {}, address]
		assert.NotEmpty(t, query.Topics[0]) // Event signature (TransferSingle + TransferBatch)
		assert.Empty(t, query.Topics[1])    // Empty (any operator)
		assert.Empty(t, query.Topics[2])    // Empty (any from address)
		assert.NotEmpty(t, query.Topics[3]) // Address topics
	})

	t.Run("Merged query structure (Both direction)", func(t *testing.T) {
		config := TransferQueryConfig{
			FromBlock:     fromBlock,
			ToBlock:       toBlock,
			Accounts:      addresses,
			TransferTypes: []TransferType{TransferTypeERC20, TransferTypeERC1155},
			Direction:     Both,
		}

		queries := config.ToFilterQueries()
		assert.Len(t, queries, 3)

		// First query: ERC20 Send (2 topics)
		query1 := queries[0]
		assert.Len(t, query1.Topics, 2)

		// Second query: Merged ERC20 Receive + ERC1155 Send (3 topics)
		query2 := queries[1]
		assert.Len(t, query2.Topics, 3)
		assert.NotEmpty(t, query2.Topics[0]) // Event signatures (ERC20 + ERC1155)
		assert.Empty(t, query2.Topics[1])    // Empty (any from/operator)
		assert.NotEmpty(t, query2.Topics[2]) // Address topics

		// Third query: ERC1155 Receive (4 topics)
		query3 := queries[2]
		assert.Len(t, query3.Topics, 4)
	})
}
