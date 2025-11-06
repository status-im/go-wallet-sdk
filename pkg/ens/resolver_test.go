package ens

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		ensName   string
		wantError bool
		errorType error
	}{
		{
			name:      "valid name with .eth",
			ensName:   "vitalik.eth",
			wantError: false,
		},
		{
			name:      "valid name with subdomain",
			ensName:   "wallet.vitalik.eth",
			wantError: false,
		},
		{
			name:      "empty name",
			ensName:   "",
			wantError: true,
			errorType: ErrEmptyName,
		},
		{
			name:      "name without dot",
			ensName:   "vitalik",
			wantError: true,
			errorType: ErrInvalidName,
		},
		{
			name:      "name starting with dot",
			ensName:   ".vitalik.eth",
			wantError: true,
			errorType: ErrInvalidName,
		},
		{
			name:      "name ending with dot",
			ensName:   "vitalik.eth.",
			wantError: true,
			errorType: ErrInvalidName,
		},
		{
			name:      "name with spaces gets trimmed and is valid",
			ensName:   "  vitalik.eth  ",
			wantError: false,
		},
		{
			name:      "name with special characters (allowed, go-ens handles validation)",
			ensName:   "vitalik@.eth",
			wantError: false,
		},
		{
			name:      "unicode name (allowed, go-ens handles validation)",
			ensName:   "münchen.eth",
			wantError: false,
		},
		{
			name:      "name with uppercase (valid, will be normalized)",
			ensName:   "Vitalik.ETH",
			wantError: false,
		},
		{
			name:      "name with numbers",
			ensName:   "wallet123.eth",
			wantError: false,
		},
		{
			name:      "name with hyphens",
			ensName:   "my-wallet.eth",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.ensName)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name      string
		address   common.Address
		wantError bool
		errorType error
	}{
		{
			name:      "valid address",
			address:   common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"),
			wantError: false,
		},
		{
			name:      "zero address",
			address:   common.Address{},
			wantError: true,
			errorType: ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAddress(tt.address)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewResolver_NilClient(t *testing.T) {
	resolver, err := NewResolver(nil)
	require.Error(t, err)
	assert.Nil(t, resolver)
	assert.Contains(t, err.Error(), "client cannot be nil")
}

// Note: Full integration tests for AddressOf and GetName would require a live
// Ethereum connection and are better suited for integration test suites.
// The validation logic is covered by the unit tests above.

func TestAddressOf_ValidationErrors(t *testing.T) {
	// This test only covers validation errors, not actual resolution
	// Actual resolution requires a live client and is tested in integration tests

	tests := []struct {
		name      string
		ensName   string
		wantError bool
	}{
		{
			name:      "empty name",
			ensName:   "",
			wantError: true,
		},
		{
			name:      "invalid name without dot",
			ensName:   "vitalik",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock resolver with nil client for validation testing
			// Note: This won't work for actual resolution, only validation
			r := &Resolver{}

			_, err := r.AddressOf(tt.ensName)
			if tt.wantError {
				require.Error(t, err)
			}
		})
	}
}

func TestGetName_ValidationErrors(t *testing.T) {
	// This test only covers validation errors, not actual resolution

	tests := []struct {
		name      string
		address   common.Address
		wantError bool
	}{
		{
			name:      "zero address",
			address:   common.Address{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock resolver with nil client for validation testing
			r := &Resolver{}

			_, err := r.GetName(tt.address)
			if tt.wantError {
				require.Error(t, err)
			}
		})
	}
}
