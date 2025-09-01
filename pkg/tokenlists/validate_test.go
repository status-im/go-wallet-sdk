package tokenlists

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr error
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: ErrConfigNotProvided,
		},
		{
			name: "missing logger",
			config: &Config{
				MainList:   []byte("{}"),
				MainListID: StatusListID,
				Chains:     []uint64{1},
			},
			wantErr: ErrLoggerNotProvided,
		},
		{
			name: "missing main list",
			config: &Config{
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				PrivacyGuard:                  &defaultPrivacyGuard{},
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  &defaultContentStore{},
				logger:                        zap.NewNop(),
			},
			wantErr: ErrMainListNotProvided,
		},
		{
			name: "missing main list ID",
			config: &Config{
				MainList:                      []byte("{}"),
				Chains:                        []uint64{1},
				PrivacyGuard:                  &defaultPrivacyGuard{},
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  &defaultContentStore{},
				logger:                        zap.NewNop(),
			},
			wantErr: ErrMainListIDNotProvided,
		},
		{
			name: "missing chains",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				PrivacyGuard:                  &defaultPrivacyGuard{},
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  &defaultContentStore{},
				logger:                        zap.NewNop(),
			},
			wantErr: ErrChainsNotProvided,
		},
		{
			name: "missing privacy guard",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  &defaultContentStore{},
				logger:                        zap.NewNop(),
			},
			wantErr: ErrPrivacyGuardNotProvided,
		},
		{
			name: "missing last update time store",
			config: &Config{
				MainList:     []byte("{}"),
				MainListID:   StatusListID,
				Chains:       []uint64{1},
				PrivacyGuard: &defaultPrivacyGuard{},
				ContentStore: &defaultContentStore{},
				logger:       zap.NewNop(),
			},
			wantErr: ErrLastTokenListsUpdateTimeStoreNotProvided,
		},
		{
			name: "missing content store",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				PrivacyGuard:                  &defaultPrivacyGuard{},
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				logger:                        zap.NewNop(),
			},
			wantErr: ErrContentStoreNotProvided,
		},
		{
			name: "invalid refresh intervals",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				AutoRefreshInterval:           1,
				AutoRefreshCheckInterval:      2,
				PrivacyGuard:                  &defaultPrivacyGuard{},
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  &defaultContentStore{},
				logger:                        zap.NewNop(),
			},
			wantErr: ErrAutoRefreshCheckIntervalGreaterThanInterval,
		},
		{
			name: "valid config",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				PrivacyGuard:                  &defaultPrivacyGuard{},
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  &defaultContentStore{},
				logger:                        zap.NewNop(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsChainAllowed(t *testing.T) {
	allowedChains := []uint64{1, 2, 3}

	tests := []struct {
		name     string
		chainID  uint64
		allowed  []uint64
		expected bool
	}{
		{
			name:     "chain allowed",
			chainID:  1,
			allowed:  allowedChains,
			expected: true,
		},
		{
			name:     "chain not allowed",
			chainID:  4,
			allowed:  allowedChains,
			expected: false,
		},
		{
			name:     "empty allowed chains",
			chainID:  1,
			allowed:  []uint64{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isChainAllowed(tt.chainID, tt.allowed)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidLogoURI(t *testing.T) {
	tests := []struct {
		name     string
		logoURI  string
		expected bool
	}{
		{
			name:     "empty logo URI",
			logoURI:  "",
			expected: true,
		},
		{
			name:     "data URI",
			logoURI:  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
			expected: true,
		},
		{
			name:     "IPFS URI",
			logoURI:  "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYjPjoiQX5cL1T3bqgm1",
			expected: true,
		},
		{
			name:     "HTTP URI",
			logoURI:  "http://example.com/logo.png",
			expected: true,
		},
		{
			name:     "HTTPS URI",
			logoURI:  "https://example.com/logo.png",
			expected: true,
		},
		{
			name:     "invalid URI",
			logoURI:  "ftp://example.com/logo.png",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidLogoURI(tt.logoURI)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateToken(t *testing.T) {
	allowedChains := []uint64{1, 2, 3}

	tests := []struct {
		name    string
		token   Token
		allowed []uint64
		wantErr error
	}{
		{
			name: "valid token",
			token: Token{
				ChainID:  1,
				Address:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Symbol:   "TEST",
				Decimals: 18,
				LogoURI:  "https://example.com/logo.png",
			},
			allowed: allowedChains,
			wantErr: nil,
		},
		{
			name: "chain not allowed",
			token: Token{
				ChainID:  4,
				Address:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Symbol:   "TEST",
				Decimals: 18,
			},
			allowed: allowedChains,
			wantErr: ErrChainNotAllowed,
		},
		{
			name: "empty symbol",
			token: Token{
				ChainID:  1,
				Address:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Symbol:   "",
				Decimals: 18,
			},
			allowed: allowedChains,
			wantErr: ErrSymbolCannotBeEmpty,
		},
		{
			name: "decimals too high",
			token: Token{
				ChainID:  1,
				Address:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Symbol:   "TEST",
				Decimals: 19,
			},
			allowed: allowedChains,
			wantErr: ErrDecimalsExceedsMaximum,
		},
		{
			name: "invalid logo URI",
			token: Token{
				ChainID:  1,
				Address:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Symbol:   "TEST",
				Decimals: 18,
				LogoURI:  "ftp://example.com/logo.png",
			},
			allowed: allowedChains,
			wantErr: ErrInvalidLogoURI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToken(&tt.token, tt.allowed)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateJsonAgainstSchema(t *testing.T) {
	validJSON := `{"name": "Test Token List", "tokens": []}`
	invalidJSON := `{"name": "Test Token List"}`

	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"tokens": {"type": "array"}
		},
		"required": ["name", "tokens"]
	}`

	schemaLoader := gojsonschema.NewStringLoader(schema)

	err := validateJsonAgainstSchema(validJSON, schemaLoader)
	assert.NoError(t, err)

	err = validateJsonAgainstSchema(invalidJSON, schemaLoader)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenListDoesNotMatchSchema)
}
