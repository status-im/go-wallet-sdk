package parsers_test

import (
	"testing"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// #nosec G101
	validListOfTokenListsJson = `{
  "timestamp": "2025-09-01T00:00:00.000Z",
  "version": {
    "major": 0,
    "minor": 1,
    "patch": 0
  },
  "tokenLists": [
    {
      "id": "status",
      "sourceUrl": "https://example.com/status-token-list.json",
      "schema": "https://example.com/status-schema.json"
    },
    {
      "id": "uniswap",
      "sourceUrl": "https://example.com/uniswap.json",
      "schema": ""
    }
  ]
}`

	// #nosec G101
	emptyTokenListsJson = `{
  "timestamp": "2025-09-01T00:00:00.000Z",
  "version": {
    "major": 1,
    "minor": 0,
    "patch": 0
  },
  "tokenLists": []
}`

	minimalValidJson = `{
  "tokenLists": [
    {
      "id": "minimal",
      "sourceUrl": "https://example.com/minimal.json"
    }
  ]
}`

	invalidJsonMissingBrace = `{
  "timestamp": "2025-09-01T00:00:00.000Z",
  "version": {
    "major": 0,
    "minor": 1,
    "patch": 0
  },
  "tokenLists": [
    {
      "id": "status",
      "sourceUrl": "https://example.com/status-token-list.json"
    }
  `

	invalidJsonWrongType = `{
  "timestamp": "2025-09-01T00:00:00.000Z",
  "version": {
    "major": "zero",
    "minor": 1,
    "patch": 0
  },
  "tokenLists": []
}`
)

func TestStatusListOfTokenListsParser_Parse(t *testing.T) {
	parser := &parsers.StatusListOfTokenListsParser{}

	tests := []struct {
		name        string
		raw         []byte
		expectError bool
		expected    *types.ListOfTokenLists
	}{
		{
			name:        "valid list of token lists with multiple entries",
			raw:         []byte(validListOfTokenListsJson),
			expectError: false,
			expected: &types.ListOfTokenLists{
				Timestamp: "2025-09-01T00:00:00.000Z",
				Version: types.Version{
					Major: 0,
					Minor: 1,
					Patch: 0,
				},
				TokenLists: []types.ListDetails{
					{
						ID:        "status",
						SourceURL: "https://example.com/status-token-list.json",
						Schema:    "https://example.com/status-schema.json",
					},
					{
						ID:        "uniswap",
						SourceURL: "https://example.com/uniswap.json",
						Schema:    "",
					},
				},
			},
		},
		{
			name:        "empty token lists",
			raw:         []byte(emptyTokenListsJson),
			expectError: false,
			expected: &types.ListOfTokenLists{
				Timestamp: "2025-09-01T00:00:00.000Z",
				Version: types.Version{
					Major: 1,
					Minor: 0,
					Patch: 0,
				},
				TokenLists: []types.ListDetails{},
			},
		},
		{
			name:        "minimal valid JSON",
			raw:         []byte(minimalValidJson),
			expectError: false,
			expected: &types.ListOfTokenLists{
				Timestamp: "",
				Version: types.Version{
					Major: 0,
					Minor: 0,
					Patch: 0,
				},
				TokenLists: []types.ListDetails{
					{
						ID:        "minimal",
						SourceURL: "https://example.com/minimal.json",
						Schema:    "",
					},
				},
			},
		},
		{
			name:        "empty JSON object",
			raw:         []byte(`{}`),
			expectError: false,
			expected: &types.ListOfTokenLists{
				Timestamp:  "",
				Version:    types.Version{},
				TokenLists: nil,
			},
		},
		{
			name:        "invalid JSON - missing brace",
			raw:         []byte(invalidJsonMissingBrace),
			expectError: true,
			expected:    nil,
		},
		{
			name:        "invalid JSON - wrong type for version.major",
			raw:         []byte(invalidJsonWrongType),
			expectError: true,
			expected:    nil,
		},
		{
			name:        "invalid JSON - completely malformed",
			raw:         []byte(`{invalid json`),
			expectError: true,
			expected:    nil,
		},
		{
			name:        "empty input",
			raw:         []byte(""),
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.raw)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expected.Timestamp, result.Timestamp)
			assert.Equal(t, tt.expected.Version.Major, result.Version.Major)
			assert.Equal(t, tt.expected.Version.Minor, result.Version.Minor)
			assert.Equal(t, tt.expected.Version.Patch, result.Version.Patch)

			assert.Len(t, result.TokenLists, len(tt.expected.TokenLists))
			for i, expectedTokenList := range tt.expected.TokenLists {
				found := false
				for i := range result.TokenLists {
					if expectedTokenList.ID == result.TokenLists[i].ID {
						found = true
						break
					}
				}
				assert.True(t, found)
				assert.Equal(t, expectedTokenList.SourceURL, result.TokenLists[i].SourceURL)
				assert.Equal(t, expectedTokenList.Schema, result.TokenLists[i].Schema)
			}
		})
	}
}

func TestStatusListOfTokenListsParser_Parse_NilInput(t *testing.T) {
	parser := &parsers.StatusListOfTokenListsParser{}

	result, err := parser.Parse(nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestStatusListOfTokenListsParser_Parse_LargeInput(t *testing.T) {
	parser := &parsers.StatusListOfTokenListsParser{}

	// Create a large but valid JSON with many token lists
	largeJson := `{
  "timestamp": "2025-09-01T00:00:00.000Z",
  "version": {
    "major": 0,
    "minor": 1,
    "patch": 0
  },
  "tokenLists": [`

	// Add 100 token list entries
	for i := range 100 {
		if i > 0 {
			largeJson += ","
		}
		largeJson += `
    {
      "id": "list` + string(rune('0'+i%10)) + `",
      "sourceUrl": "https://example.com/list` + string(rune('0'+i%10)) + `.json",
      "schema": "https://example.com/schema` + string(rune('0'+i%10)) + `.json"
    }`
	}

	largeJson += `
  ]
}`

	result, err := parser.Parse([]byte(largeJson))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.TokenLists, 100)
	assert.Equal(t, "2025-09-01T00:00:00.000Z", result.Timestamp)
	assert.Equal(t, 0, result.Version.Major)
	assert.Equal(t, 1, result.Version.Minor)
	assert.Equal(t, 0, result.Version.Patch)
}
