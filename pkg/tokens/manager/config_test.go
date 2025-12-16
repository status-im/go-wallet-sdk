package manager_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/common"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/manager"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	mock_parsers "github.com/status-im/go-wallet-sdk/pkg/tokens/parsers/mock"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "main-list",
			InitialLists: map[string][]byte{
				"main-list":  []byte(`{"name": "Main List", "tokens": []}`),
				"other-list": []byte(`{"name": "Other List", "tokens": []}`),
			},
			Chains: []uint64{common.EthereumMainnet, common.BSCMainnet},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty main list ID", func(t *testing.T) {
		config := &manager.Config{
			MainListID:    "",
			InitialLists:  map[string][]byte{},
			CustomParsers: map[string]parsers.TokenListParser{},
			Chains:        []uint64{common.EthereumMainnet},
		}

		err := config.Validate()
		assert.ErrorIs(t, err, manager.ErrMainListIDNotProvided)
	})

	t.Run("main list not in initial lists", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "main-list",
			InitialLists: map[string][]byte{
				"other-list": []byte(`{"name": "Other List", "tokens": []}`),
			},
			Chains: []uint64{common.EthereumMainnet},
		}

		err := config.Validate()
		assert.ErrorIs(t, err, manager.ErrMainListNotProvided)
	})

	t.Run("missing custom parser defaults to standard parser", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "main-list",
			InitialLists: map[string][]byte{
				"main-list":  []byte(`{"name": "Main List", "tokens": []}`),
				"other-list": []byte(`{"name": "Other List", "tokens": []}`),
			},
			CustomParsers: map[string]parsers.TokenListParser{
				"main-list": &parsers.CoinGeckoAllTokensParser{},
				// no custom parser for "other-list" - should default to StandardTokenListParser
			},
			Chains: []uint64{1},
		}

		err := config.Validate()
		assert.NoError(t, err) // Should not fail, defaults to StandardTokenListParser
	})

	t.Run("empty chains", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "main-list",
			InitialLists: map[string][]byte{
				"main-list": []byte(`{"name": "Main List", "tokens": []}`),
			},
			Chains: []uint64{}, // empty chains
		}

		err := config.Validate()
		assert.ErrorIs(t, err, manager.ErrChainsNotProvided)
	})

	t.Run("nil chains", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "main-list",
			InitialLists: map[string][]byte{
				"main-list": []byte(`{"name": "Main List", "tokens": []}`),
			},
			Chains: nil, // nil chains
		}

		err := config.Validate()
		assert.ErrorIs(t, err, manager.ErrChainsNotProvided)
	})

	t.Run("valid config with autofetcher", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)

		config := &manager.Config{
			MainListID: "main-list",
			InitialLists: map[string][]byte{
				"main-list": []byte(`{"name": "Main List", "tokens": []}`),
			},
			Chains: []uint64{common.EthereumMainnet, common.BSCMainnet},
			AutoFetcherConfig: &autofetcher.ConfigRemoteListOfTokenLists{
				Config: autofetcher.Config{
					AutoRefreshInterval:      time.Hour,
					AutoRefreshCheckInterval: time.Minute,
				},
				RemoteListOfTokenListsFetchDetails: types.ListDetails{
					ID:        "remote-list",
					SourceURL: "https://example.com/token-lists.json",
					Schema:    "status-list-of-token-lists",
				},
				RemoteListOfTokenListsParser: mockParser,
			},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid autofetcher config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)

		config := &manager.Config{
			MainListID: "main-list",
			InitialLists: map[string][]byte{
				"main-list": []byte(`{"name": "Main List", "tokens": []}`),
			},
			Chains: []uint64{common.EthereumMainnet},
			AutoFetcherConfig: &autofetcher.ConfigRemoteListOfTokenLists{
				Config: autofetcher.Config{
					AutoRefreshInterval:      time.Minute,
					AutoRefreshCheckInterval: time.Hour,
				},
				RemoteListOfTokenListsFetchDetails: types.ListDetails{
					ID:        "remote-list",
					SourceURL: "https://example.com/token-lists.json",
					Schema:    "status-list-of-token-lists",
				},
				RemoteListOfTokenListsParser: mockParser,
			},
		}

		err := config.Validate()
		assert.Error(t, err)
	})

	t.Run("complex valid config", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "uniswap-default",
			InitialLists: map[string][]byte{
				"uniswap-default": []byte(`{"name": "Uniswap Default List", "tokens": []}`),
				"compound":        []byte(`{"name": "Compound Token List", "tokens": []}`),
				"aave":            []byte(`{"name": "Aave Token List", "tokens": []}`),
				"status":          []byte(`{"name": "Status Token List", "tokens": {}}`),
			},
			CustomParsers: map[string]parsers.TokenListParser{
				"status": &parsers.StatusTokenListParser{},
			},
			Chains: []uint64{common.EthereumMainnet, common.BSCMainnet, common.OptimismMainnet, common.ArbitrumMainnet},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})
}

func TestConfig_ValidationEdgeCases(t *testing.T) {
	t.Run("only main list", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "only-list",
			InitialLists: map[string][]byte{
				"only-list": []byte(`{"name": "Only List", "tokens": []}`),
			},
			Chains: []uint64{common.EthereumMainnet},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("main list with different parsers", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "main-list",
			InitialLists: map[string][]byte{
				"main-list": []byte(`{"name": "Main List", "tokens": []}`),
				"standard":  []byte(`{"name": "Standard List", "tokens": []}`),
				"status":    []byte(`{"name": "Status List", "tokens": {}}`),
				"coingecko": []byte(`{"bitcoin": {"id": "bitcoin", "platforms": {}}}`),
			},
			CustomParsers: map[string]parsers.TokenListParser{
				"status":    &parsers.StatusTokenListParser{},
				"coingecko": &parsers.CoinGeckoAllTokensParser{},
			},
			Chains: []uint64{common.EthereumMainnet, common.BSCMainnet},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("single chain configuration", func(t *testing.T) {
		config := &manager.Config{
			MainListID: "eth-only",
			InitialLists: map[string][]byte{
				"eth-only": []byte(`{"name": "Ethereum Only", "tokens": []}`),
			},
			Chains: []uint64{common.EthereumMainnet},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("many chains configuration", func(t *testing.T) {
		manyChains := []uint64{
			common.EthereumMainnet,
			common.BSCMainnet,
			common.OptimismMainnet,
			common.ArbitrumMainnet,
			common.BaseMainnet,
			common.StatusNetworkSepolia,
		}

		config := &manager.Config{
			MainListID: "multi-chain",
			InitialLists: map[string][]byte{
				"multi-chain": []byte(`{"name": "Multi Chain List", "tokens": []}`),
			},
			Chains: manyChains,
		}

		err := config.Validate()
		assert.NoError(t, err)
	})
}
