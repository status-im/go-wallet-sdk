package mnemonic_test

import (
	"testing"

	"github.com/status-im/extkeys"
	"github.com/status-im/go-wallet-sdk/pkg/accounts/mnemonic"
	"github.com/stretchr/testify/assert"
)

func TestCreateRandomMnemonic(t *testing.T) {

	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{"valid length 12", 12, false},
		{"valid length 15", 15, false},
		{"valid length 18", 18, false},
		{"valid length 21", 21, false},
		{"valid length 24", 24, false},
		{"invalid length 11", 11, true},
		{"invalid length 13", 13, true},
		{"invalid length 25", 25, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := mnemonic.CreateRandomMnemonic(tt.length)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, mnemonic.ErrInvalidMnemonicPhraseLength, err)
				assert.Empty(t, m)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, m)
			}
		})
	}
}

func TestCreateRandomMnemonicWithDefaultLength(t *testing.T) {
	m, err := mnemonic.CreateRandomMnemonicWithDefaultLength()
	assert.NoError(t, err)
	assert.NotEmpty(t, m)
}

func TestCreateExtendedKeyFromMnemonic(t *testing.T) {
	tests := []struct {
		name        string
		phrase      string
		passphrase  string
		expectError bool
	}{
		{
			name:        "valid mnemonic",
			phrase:      "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			passphrase:  "",
			expectError: false,
		},
		{
			name:        "invalid mnemonic",
			phrase:      "invalid mnemonic phrase",
			passphrase:  "",
			expectError: false,
		},
		{
			name:        "empty mnemonic",
			phrase:      "",
			passphrase:  "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := mnemonic.CreateExtendedKeyFromMnemonic(tt.phrase, tt.passphrase)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
			}
		})
	}
}

func TestLengthToEntropyStrength(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{"valid length 12", 12, false},
		{"valid length 15", 15, false},
		{"valid length 18", 18, false},
		{"valid length 21", 21, false},
		{"valid length 24", 24, false},
		{"invalid length 11", 11, true},
		{"invalid length 13", 13, true},
		{"invalid length 25", 25, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength, err := mnemonic.LengthToEntropyStrength(tt.length)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, mnemonic.ErrInvalidMnemonicPhraseLength, err)
				assert.Equal(t, extkeys.EntropyStrength(0), strength)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, strength)
			}
		})
	}
}
