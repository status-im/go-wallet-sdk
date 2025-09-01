package tokenlists

import (
	"strings"
	"testing"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"github.com/stretchr/testify/assert"
)

// Uniswap token list follows the StandardTokenList format, so we can use the StandardTokenListParser to parse it.
func TestStandardTokenListParser_Parse(t *testing.T) {
	parser := &StandardTokenListParser{}

	tests := []struct {
		name                string
		raw                 []byte
		sourceURL           string
		useFetchedTimestamp bool
		fetchedTokenList    fetchedTokenList
		expectedTokenList   TokenList
	}{
		{
			name:                "valid uniswap token list with fetched timestamp",
			raw:                 []byte(uniswapTokenListJsonResponse2),
			sourceURL:           "https://example.com/uniswap-token-list.json",
			useFetchedTimestamp: true,
			fetchedTokenList:    fetchedUniswapTokenList2,
			expectedTokenList:   uniswapTokenList2,
		},
		{
			name:                "valid status token list without fetched timestamp",
			raw:                 []byte(uniswapTokenListJsonResponse2),
			sourceURL:           "https://example.com/uniswap-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedUniswapTokenList2,
			expectedTokenList:   uniswapTokenList2,
		},
		{
			name:                "invalid JSON",
			raw:                 []byte(uniswapTokenListInvalidTokensJsonResponse),
			sourceURL:           "https://example.com/uniswap-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedUniswapTokenListInvalidTokens,
			expectedTokenList:   uniswapTokenListInvalidTokens,
		},
		{
			name:                "empty tokens list",
			raw:                 []byte(uniswapTokenListEmptyTokensJsonResponse),
			sourceURL:           "https://example.com/uniswap-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedUniswapTokenListEmpty,
			expectedTokenList:   uniswapTokenListEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp := time.Time{}
			if tt.useFetchedTimestamp {
				timestamp = tt.fetchedTokenList.Fetched
			}
			got, err := parser.Parse(tt.raw, tt.sourceURL, timestamp, common.AllChains)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.expectedTokenList.Name, got.Name)
			assert.Equal(t, tt.expectedTokenList.Timestamp, got.Timestamp)
			if tt.useFetchedTimestamp {
				assert.Equal(t, tt.expectedTokenList.FetchedTimestamp, got.FetchedTimestamp)
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

func TestStandardTokenListParser_Parse_InvalidJSON(t *testing.T) {
	parser := &StandardTokenListParser{}

	raw := []byte(`{invalid json`)
	_, err := parser.Parse(raw, "https://example.com/invalid.json", time.Time{}, common.AllChains)
	assert.Error(t, err)
}

func TestStandardTokenListParser_Parse_MissingFields(t *testing.T) {
	parser := &StandardTokenListParser{}

	raw := []byte(``)
	got, err := parser.Parse(raw, "https://example.com/missing-fields.json", time.Time{}, common.AllChains)
	assert.Error(t, err)
	assert.Nil(t, got)

	raw = []byte(`{}`)
	got, err = parser.Parse(raw, "https://example.com/missing-fields.json", time.Time{}, common.AllChains)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got.Tokens)

	raw = []byte(`{
		"name": "Uniswap Labs Default"
	}`)
	got, err = parser.Parse(raw, "https://example.com/missing-fields.json", time.Time{}, common.AllChains)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "Uniswap Labs Default", got.Name)
	assert.Empty(t, got.Tokens)
}
