package tokenlists

import (
	"strings"
	"testing"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"github.com/stretchr/testify/assert"
)

func TestStatusTokenListParser_Parse(t *testing.T) {
	parser := &StatusTokenListParser{}

	tests := []struct {
		name                string
		raw                 []byte
		sourceURL           string
		useFetchedTimestamp bool
		fetchedTokenList    fetchedTokenList
		expectedTokenList   TokenList
	}{
		{
			name:                "valid status token list with fetched timestamp",
			raw:                 []byte(statusTokenListJsonResponse),
			sourceURL:           "https://example.com/status-token-list.json",
			useFetchedTimestamp: true,
			fetchedTokenList:    fetchedStatusTokenList,
			expectedTokenList:   statusTokenList,
		},
		{
			name:                "valid status token list without fetched timestamp",
			raw:                 []byte(statusTokenListJsonResponse),
			sourceURL:           "https://example.com/status-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedStatusTokenList,
			expectedTokenList:   statusTokenList,
		},
		{
			name:                "invalid JSON",
			raw:                 []byte(statusTokenListInvalidTokensJsonResponse),
			sourceURL:           "https://example.com/status-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedStatusTokenListInvalidTokens,
			expectedTokenList:   statusTokenListInvalidTokens,
		},
		{
			name:                "empty tokens list",
			raw:                 []byte(statusEmptyTokensJsonResponse),
			sourceURL:           "https://example.com/status-token-list.json",
			useFetchedTimestamp: false,
			fetchedTokenList:    fetchedStatusTokenListEmpty,
			expectedTokenList:   statusTokenListEmpty,
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

func TestStatusTokenListParser_Parse_InvalidJSON(t *testing.T) {
	parser := &StatusTokenListParser{}

	raw := []byte(`{invalid json`)
	_, err := parser.Parse(raw, "https://example.com/invalid.json", time.Time{}, common.AllChains)
	assert.Error(t, err)
}

func TestStatusTokenListParser_Parse_MissingFields(t *testing.T) {
	parser := &StatusTokenListParser{}

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
		"name": "Status Token List"
	}`)
	got, err = parser.Parse(raw, "https://example.com/missing-fields.json", time.Time{}, common.AllChains)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "Status Token List", got.Name)
	assert.Empty(t, got.Tokens)
}
