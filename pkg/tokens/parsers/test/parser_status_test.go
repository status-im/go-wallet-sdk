package parsers_test

import (
	"strings"
	"testing"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"

	"github.com/stretchr/testify/assert"
)

func TestStatusTokenListParser_Parse(t *testing.T) {
	parser := &parsers.StatusTokenListParser{}

	tests := []struct {
		name              string
		raw               []byte
		fetchedTokenList  fetcher.FetchedData
		expectedTokenList types.TokenList
	}{
		{
			name:              "valid status token list",
			raw:               []byte(statusTokenListJsonResponse),
			fetchedTokenList:  fetchedStatusTokenList,
			expectedTokenList: statusTokenList,
		},
		{
			name:              "invalid JSON",
			raw:               []byte(statusTokenListInvalidTokensJsonResponse),
			fetchedTokenList:  fetchedStatusTokenListInvalidTokens,
			expectedTokenList: statusTokenListInvalidTokens,
		},
		{
			name:              "empty tokens list",
			raw:               []byte(statusEmptyTokensJsonResponse),
			fetchedTokenList:  fetchedStatusTokenListEmpty,
			expectedTokenList: statusTokenListEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.raw, common.AllChains)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.expectedTokenList.Name, got.Name)
			assert.Equal(t, tt.expectedTokenList.Timestamp, got.Timestamp)
			assert.Equal(t, tt.expectedTokenList.FetchedTimestamp, got.FetchedTimestamp)
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
	parser := &parsers.StatusTokenListParser{}

	raw := []byte(`{invalid json`)
	_, err := parser.Parse(raw, common.AllChains)
	assert.Error(t, err)
}

func TestStatusTokenListParser_Parse_MissingFields(t *testing.T) {
	parser := &parsers.StatusTokenListParser{}

	raw := []byte(``)
	got, err := parser.Parse(raw, common.AllChains)
	assert.Error(t, err)
	assert.Nil(t, got)

	raw = []byte(`{}`)
	got, err = parser.Parse(raw, common.AllChains)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got.Tokens)

	raw = []byte(`{
		"name": "Status Token List"
	}`)
	got, err = parser.Parse(raw, common.AllChains)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "Status Token List", got.Name)
	assert.Empty(t, got.Tokens)
}
