package multistandardfetcher_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/balance/multistandardfetcher"
)

func TestCollectibleID_ToHashable_RoundTrip(t *testing.T) {
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
		{
			name:         "maximum 256-bit token ID",
			contractAddr: "0x495f947276749Ce646f68AC8c248420045cb7b5e",
			tokenID:      "115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
		{
			name:         "different contract",
			contractAddr: "0x76BE3b62873462d2142405439777e971754E8E77",
			tokenID:      "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse token ID
			tokenID, ok := new(big.Int).SetString(tt.tokenID, 10)
			require.True(t, ok, "failed to parse token ID")

			// Create original CollectibleID
			original := multistandardfetcher.CollectibleID{
				ContractAddress: common.HexToAddress(tt.contractAddr),
				TokenID:         tokenID,
			}

			// Convert to HashableCollectibleID
			hashable := original.ToHashableCollectibleID()

			// Verify hashable structure
			assert.Equal(t, original.ContractAddress, hashable.ContractAddress)

			// Convert back to CollectibleID
			recovered := hashable.ToCollectibleID()

			// Verify round trip
			assert.Equal(t, original.ContractAddress, recovered.ContractAddress)
			assert.Equal(t, 0, original.TokenID.Cmp(recovered.TokenID), "token IDs should match")
		})
	}
}

func TestHashableCollectibleID_Consistency(t *testing.T) {
	// Test that the same CollectibleID always produces the same HashableCollectibleID
	contractAddr := common.HexToAddress("0x495f947276749Ce646f68AC8c248420045cb7b5e")
	tokenID := big.NewInt(42)

	cid := multistandardfetcher.CollectibleID{
		ContractAddress: contractAddr,
		TokenID:         tokenID,
	}

	hashable1 := cid.ToHashableCollectibleID()
	hashable2 := cid.ToHashableCollectibleID()

	// Same input should produce same hashable ID
	assert.Equal(t, hashable1.ContractAddress, hashable2.ContractAddress)
	assert.Equal(t, hashable1.TokenID, hashable2.TokenID)

	// Round trip should produce same result
	cid1 := hashable1.ToCollectibleID()
	cid2 := hashable2.ToCollectibleID()

	assert.Equal(t, cid1.ContractAddress, cid2.ContractAddress)
	assert.Equal(t, 0, cid1.TokenID.Cmp(cid2.TokenID))
}

func TestCollectibleID_TokenIDPadding(t *testing.T) {
	// Test that token IDs are properly padded to 32 bytes in HashableTokenID
	contractAddr := common.HexToAddress("0x495f947276749Ce646f68AC8c248420045cb7b5e")

	// Small token ID (1 byte)
	smallTokenID := big.NewInt(1)
	cidSmall := multistandardfetcher.CollectibleID{
		ContractAddress: contractAddr,
		TokenID:         smallTokenID,
	}
	hashableSmall := cidSmall.ToHashableCollectibleID()

	// Large token ID (32 bytes)
	largeTokenID, ok := new(big.Int).SetString("80725417304363601833901294931248829736335313578142415985923215537121414611044", 10)
	require.True(t, ok)
	cidLarge := multistandardfetcher.CollectibleID{
		ContractAddress: contractAddr,
		TokenID:         largeTokenID,
	}
	hashableLarge := cidLarge.ToHashableCollectibleID()

	// Both should have 32-byte token IDs
	assert.Len(t, hashableSmall.TokenID, 32)
	assert.Len(t, hashableLarge.TokenID, 32)

	// Round trip should work for both
	assert.Equal(t, 0, smallTokenID.Cmp(hashableSmall.ToCollectibleID().TokenID))
	assert.Equal(t, 0, largeTokenID.Cmp(hashableLarge.ToCollectibleID().TokenID))
}
