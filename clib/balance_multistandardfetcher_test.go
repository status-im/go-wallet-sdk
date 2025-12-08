package main

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/balance/multistandardfetcher"
)

func TestCollectibleID_String_RoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		contractAddr   string
		tokenID        string
		expectedString string
	}{
		{
			name:           "small token ID",
			contractAddr:   "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:        "1",
			expectedString: "0x495f947276749Ce646f68AC8c248420045cb7b5e:1",
		},
		{
			name:           "zero token ID",
			contractAddr:   "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:        "0",
			expectedString: "0x495f947276749Ce646f68AC8c248420045cb7b5e:0",
		},
		{
			name:           "large token ID",
			contractAddr:   "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:        "80725417304363601833901294931248829736335313578142415985923215537121414611044",
			expectedString: "0x495f947276749Ce646f68AC8c248420045cb7b5e:80725417304363601833901294931248829736335313578142415985923215537121414611044",
		},
		{
			name:           "different contract address",
			contractAddr:   "0x76BE3b62873462d2142405439777e971754E8E77",
			tokenID:        "1",
			expectedString: "0x76BE3b62873462d2142405439777e971754E8E77:1",
		},
		{
			name:           "maximum 256-bit token ID",
			contractAddr:   "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:        "115792089237316195423570985008687907853269984665640564039457584007913129639935",
			expectedString: "0x495f947276749Ce646f68AC8c248420045cb7b5e:115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse token ID
			tokenID, ok := new(big.Int).SetString(tt.tokenID, 10)
			require.True(t, ok, "failed to parse token ID")

			// Create CollectibleID
			original := multistandardfetcher.CollectibleID{
				ContractAddress: common.HexToAddress(tt.contractAddr),
				TokenID:         tokenID,
			}

			// Convert to string
			str := collectibleIDToString(original)
			assert.Equal(t, tt.expectedString, str)

			// Parse back from string
			parsed, err := collectibleIDFromString(str)
			require.NoError(t, err)

			// Verify round trip
			assert.Equal(t, original.ContractAddress, parsed.ContractAddress)
			assert.Equal(t, 0, original.TokenID.Cmp(parsed.TokenID), "token IDs should match")
			assert.Equal(t, collectibleIDToString(original), collectibleIDToString(parsed))
		})
	}
}

func TestHashableCollectibleID_String_RoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		contractAddr string
		tokenID      string
	}{
		{
			name:         "small token ID",
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "1",
		},
		{
			name:         "zero token ID",
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "0",
		},
		{
			name:         "large token ID",
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "80725417304363601833901294931248829736335313578142415985923215537121414611044",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create from string via CollectibleID
			cid, err := collectibleIDFromString(
				tt.contractAddr + ":" + tt.tokenID,
			)
			require.NoError(t, err)

			// Convert to HashableCollectibleID
			hashable := cid.ToHashableCollectibleID()

			// Convert to string
			str := hashableCollectibleIDToString(hashable)

			// Parse back from string to CollectibleID
			parsed, err := collectibleIDFromString(str)
			require.NoError(t, err)

			// Convert back to HashableCollectibleID
			hashableFromString := parsed.ToHashableCollectibleID()

			// Verify round trip
			assert.Equal(t, hashable.ContractAddress, hashableFromString.ContractAddress)
			assert.Equal(t, hashable.TokenID, hashableFromString.TokenID)
			assert.Equal(t, hashableCollectibleIDToString(hashable), hashableCollectibleIDToString(hashableFromString))
		})
	}
}

func TestCollectibleID_FullRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		contractAddr string
		tokenID      string
	}{
		{
			name:         "small token ID",
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "1",
		},
		{
			name:         "zero token ID",
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "0",
		},
		{
			name:         "large token ID from C example",
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "80725417304363601833901294931248829736335313578142415985923215537121414611044",
		},
		{
			name:         "different contract",
			contractAddr: "0x76BE3b62873462d2142405439777e971754E8E77",
			tokenID:      "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalString := tt.contractAddr + ":" + tt.tokenID

			// String -> CollectibleID
			cid, err := collectibleIDFromString(originalString)
			require.NoError(t, err)

			// CollectibleID -> HashableCollectibleID
			hashable := cid.ToHashableCollectibleID()

			// HashableCollectibleID -> CollectibleID
			cid2 := hashable.ToCollectibleID()

			// CollectibleID -> String
			finalString := collectibleIDToString(cid2)

			// Verify full round trip
			assert.Equal(t, originalString, finalString)
			assert.Equal(t, cid.ContractAddress, cid2.ContractAddress)
			assert.Equal(t, 0, cid.TokenID.Cmp(cid2.TokenID))
		})
	}
}

func TestCollectibleIDFromString_InvalidFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing colon",
			input: "0x495f947276749Ce646f68AC8c248420045cb7b5e1",
		},
		{
			name:  "too many colons",
			input: "0x495f947276749Ce646f68AC8c248420045cb7b5e:1:extra",
		},
		{
			name:  "invalid token ID format",
			input: "0x495f947276749Ce646f68AC8c248420045cb7b5e:notanumber",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "only contract address",
			input: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
		},
		{
			name:  "only token ID (empty contract address)",
			input: ":1",
		},
		{
			name:  "only contract address (empty token ID)",
			input: "0x495f947276749Ce646f68AC8c248420045cb7b5e:",
		},
		{
			name:  "both parts empty",
			input: ":",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := collectibleIDFromString(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestHashableCollectibleIDFromString(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectError  bool
		contractAddr string
		tokenID      string
	}{
		{
			name:         "valid small token ID",
			input:        "0x495f947276749Ce646f68AC8c248420045cb7b5e:1",
			expectError:  false,
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "1",
		},
		{
			name:         "valid large token ID",
			input:        "0x495f947276749Ce646f68AC8c248420045cb7b5e:80725417304363601833901294931248829736335313578142415985923215537121414611044",
			expectError:  false,
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "80725417304363601833901294931248829736335313578142415985923215537121414611044",
		},
		{
			name:        "invalid format",
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashable, err := hashableCollectibleIDFromString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, common.HexToAddress(tt.contractAddr), hashable.ContractAddress)

				// Convert back and verify token ID
				cid := hashable.ToCollectibleID()
				expectedTokenID, ok := new(big.Int).SetString(tt.tokenID, 10)
				require.True(t, ok)
				assert.Equal(t, 0, expectedTokenID.Cmp(cid.TokenID))
			}
		})
	}
}
