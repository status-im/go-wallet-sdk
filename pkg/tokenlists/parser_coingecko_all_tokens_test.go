package tokenlists

import (
	"strings"
	"testing"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"github.com/stretchr/testify/assert"
)

func TestNewCoinGeckoAllTokensParser(t *testing.T) {
	chainsMapper := map[string]uint64{
		"ethereum": 1,
		"polygon":  137,
	}

	parser := NewCoinGeckoAllTokensParser(chainsMapper)
	assert.NotNil(t, parser)
	assert.Equal(t, chainsMapper, parser.chainsMapper)
}

func TestCoinGeckoAllTokensParser_Parse(t *testing.T) {
	parser := NewCoinGeckoAllTokensParser(DefaultCoinGeckoChainsMapper)

	tests := []struct {
		name                string
		raw                 []byte
		sourceURL           string
		useFetchedTimestamp bool
		fetchedTokenList    fetchedTokenList
		expectedTokenList   TokenList
	}{
		{
			name:                "valid coingecko token list with fetched timestamp",
			raw:                 []byte(coingeckoTokensJsonResponse),
			sourceURL:           "https://example.com/coingecko-token-list.json",
			useFetchedTimestamp: true,
			fetchedTokenList:    fetchedCoingeckoTokenList,
			expectedTokenList:   coingeckoTokenList,
		},
		{
			name:                "valid coingecko token list without fetched timestamp",
			raw:                 []byte(coingeckoTokensJsonResponse),
			sourceURL:           "https://example.com/coingecko-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedCoingeckoTokenList,
			expectedTokenList:   coingeckoTokenList,
		},
		{
			name:                "invalid JSON",
			raw:                 []byte(coingeckoTokensJsonResponseInvalidTokens),
			sourceURL:           "https://example.com/coingecko-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedCoingeckoTokenListInvalidTokens,
			expectedTokenList:   coingeckoTokenListInvalidTokens,
		},
		{
			name:                "empty tokens list",
			raw:                 []byte("[]"),
			sourceURL:           "https://example.com/coingecko-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedTokenList{tokenList: tokenList{SourceURL: "https://example.com/coingecko-token-list.json"}},
			expectedTokenList:   TokenList{Source: "https://example.com/coingecko-token-list.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp := time.Time{}
			if tt.useFetchedTimestamp {
				timestamp = time.Now()
			}
			got, err := parser.Parse(tt.raw, tt.sourceURL, timestamp, common.AllChains)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.expectedTokenList.Name, got.Name)
			assert.Equal(t, tt.expectedTokenList.Timestamp, got.Timestamp)
			if tt.useFetchedTimestamp {
				assert.Equal(t, timestamp.Format(time.RFC3339), got.FetchedTimestamp)
			} else {
				assert.Equal(t, tt.expectedTokenList.Timestamp, got.Timestamp)
			}
			assert.Equal(t, tt.expectedTokenList.Source, got.Source)
			assert.Equal(t, tt.expectedTokenList.Version, got.Version)
			assert.Equal(t, tt.expectedTokenList.Tags, got.Tags)
			assert.Equal(t, tt.expectedTokenList.LogoURI, got.LogoURI)
			assert.Equal(t, tt.expectedTokenList.Keywords, got.Keywords)
			assert.Len(t, got.Tokens, len(tt.expectedTokenList.Tokens))

			for _, expectedToken := range tt.expectedTokenList.Tokens {
				found := false
				for _, actualToken := range got.Tokens {
					if actualToken.ChainID == expectedToken.ChainID && actualToken.Address == expectedToken.Address {
						found = true
						assert.Equal(t, expectedToken.CrossChainID, actualToken.CrossChainID)
						assert.Equal(t, expectedToken.ChainID, actualToken.ChainID)
						assert.Equal(t, strings.ToLower(expectedToken.Address.String()), strings.ToLower(actualToken.Address.String()))
						assert.Equal(t, expectedToken.Name, actualToken.Name)
						assert.Equal(t, expectedToken.Symbol, actualToken.Symbol)
						assert.Equal(t, expectedToken.Decimals, actualToken.Decimals)
						assert.Equal(t, strings.ToLower(expectedToken.LogoURI), strings.ToLower(actualToken.LogoURI))
						break
					}
				}
				assert.True(t, found)
			}
		})
	}
}

func TestCoinGeckoAllTokensParser_Parse_InvalidJSON(t *testing.T) {
	parser := &CoinGeckoAllTokensParser{}

	raw := []byte(`{invalid json`)
	_, err := parser.Parse(raw, "https://example.com/invalid.json", time.Time{}, common.AllChains)
	assert.Error(t, err)
}

func TestCoinGeckoAllTokensParser_Parse_MissingFields(t *testing.T) {
	parser := &CoinGeckoAllTokensParser{}

	raw := []byte(``)
	got, err := parser.Parse(raw, "https://example.com/missing-fields.json", time.Time{}, common.AllChains)
	assert.Error(t, err)
	assert.Nil(t, got)

	raw = []byte(`[]`)
	got, err = parser.Parse(raw, "https://example.com/missing-fields.json", time.Time{}, common.AllChains)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got.Tokens)

	raw = []byte(`[{
		"id": "usd-coin",
		"symbol-wrong": "usdc",
		"name-wrong": "USDC",
		"platforms": {}}]`)
	got, err = parser.Parse(raw, "https://example.com/missing-fields.json", time.Time{}, common.AllChains)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got.Tokens)
}
