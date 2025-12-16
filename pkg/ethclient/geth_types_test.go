package ethclient

import (
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestRpcProgressToSyncProgress(t *testing.T) {
	testCases := []struct {
		name     string
		progress *rpcProgress
		expected *ethereum.SyncProgress
	}{
		{
			name:     "nil progress",
			progress: nil,
			expected: nil,
		},
		{
			name: "valid progress",
			progress: &rpcProgress{
				StartingBlock:          hexutil.Uint64(0x100),
				CurrentBlock:           hexutil.Uint64(0x200),
				HighestBlock:           hexutil.Uint64(0x300),
				PulledStates:           hexutil.Uint64(0x400),
				KnownStates:            hexutil.Uint64(0x500),
				SyncedAccounts:         hexutil.Uint64(0x600),
				SyncedAccountBytes:     hexutil.Uint64(0x700),
				SyncedBytecodes:        hexutil.Uint64(0x800),
				SyncedBytecodeBytes:    hexutil.Uint64(0x900),
				SyncedStorage:          hexutil.Uint64(0xa00),
				SyncedStorageBytes:     hexutil.Uint64(0xb00),
				HealedTrienodes:        hexutil.Uint64(0xc00),
				HealedTrienodeBytes:    hexutil.Uint64(0xd00),
				HealedBytecodes:        hexutil.Uint64(0xe00),
				HealedBytecodeBytes:    hexutil.Uint64(0xf00),
				HealingTrienodes:       hexutil.Uint64(0x1000),
				HealingBytecode:        hexutil.Uint64(0x1100),
				TxIndexFinishedBlocks:  hexutil.Uint64(0x1200),
				TxIndexRemainingBlocks: hexutil.Uint64(0x1300),
			},
			expected: &ethereum.SyncProgress{
				StartingBlock:          0x100,
				CurrentBlock:           0x200,
				HighestBlock:           0x300,
				PulledStates:           0x400,
				KnownStates:            0x500,
				SyncedAccounts:         0x600,
				SyncedAccountBytes:     0x700,
				SyncedBytecodes:        0x800,
				SyncedBytecodeBytes:    0x900,
				SyncedStorage:          0xa00,
				SyncedStorageBytes:     0xb00,
				HealedTrienodes:        0xc00,
				HealedTrienodeBytes:    0xd00,
				HealedBytecodes:        0xe00,
				HealedBytecodeBytes:    0xf00,
				HealingTrienodes:       0x1000,
				HealingBytecode:        0x1100,
				TxIndexFinishedBlocks:  0x1200,
				TxIndexRemainingBlocks: 0x1300,
			},
		},
		{
			name: "zero values",
			progress: &rpcProgress{
				StartingBlock:          hexutil.Uint64(0),
				CurrentBlock:           hexutil.Uint64(0),
				HighestBlock:           hexutil.Uint64(0),
				PulledStates:           hexutil.Uint64(0),
				KnownStates:            hexutil.Uint64(0),
				SyncedAccounts:         hexutil.Uint64(0),
				SyncedAccountBytes:     hexutil.Uint64(0),
				SyncedBytecodes:        hexutil.Uint64(0),
				SyncedBytecodeBytes:    hexutil.Uint64(0),
				SyncedStorage:          hexutil.Uint64(0),
				SyncedStorageBytes:     hexutil.Uint64(0),
				HealedTrienodes:        hexutil.Uint64(0),
				HealedTrienodeBytes:    hexutil.Uint64(0),
				HealedBytecodes:        hexutil.Uint64(0),
				HealedBytecodeBytes:    hexutil.Uint64(0),
				HealingTrienodes:       hexutil.Uint64(0),
				HealingBytecode:        hexutil.Uint64(0),
				TxIndexFinishedBlocks:  hexutil.Uint64(0),
				TxIndexRemainingBlocks: hexutil.Uint64(0),
			},
			expected: &ethereum.SyncProgress{
				StartingBlock:          0,
				CurrentBlock:           0,
				HighestBlock:           0,
				PulledStates:           0,
				KnownStates:            0,
				SyncedAccounts:         0,
				SyncedAccountBytes:     0,
				SyncedBytecodes:        0,
				SyncedBytecodeBytes:    0,
				SyncedStorage:          0,
				SyncedStorageBytes:     0,
				HealedTrienodes:        0,
				HealedTrienodeBytes:    0,
				HealedBytecodes:        0,
				HealedBytecodeBytes:    0,
				HealingTrienodes:       0,
				HealingBytecode:        0,
				TxIndexFinishedBlocks:  0,
				TxIndexRemainingBlocks: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.progress.toSyncProgress()
			assert.Equal(t, tc.expected, result)
		})
	}
}
