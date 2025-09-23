package manager

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

//go:generate mockgen -destination=mock/manager.go . CustomTokenStore

// Manager is the public interface for managing token lists.
type Manager interface {
	// Start begins the Manager service, if notify channel is provided, it will be notified when the token lists are refreshed.
	// Once the manager is started, the initial state is built and then the manager will start to manage the refresh of the token lists
	// if auto refresh is enabled.
	Start(ctx context.Context, autoRefreshEnabled bool, notifyCh chan struct{}) error
	// Stop stops the Manager service.
	Stop() error

	// EnableAutoRefresh enables auto refresh of the token lists.
	EnableAutoRefresh(ctx context.Context) error
	// DisableAutoRefresh disables auto refresh of the token lists.
	DisableAutoRefresh(ctx context.Context) error
	// TriggerRefresh triggers a manual refresh of the token lists.
	TriggerRefresh(ctx context.Context) error

	// UniqueTokens returns all unique tokens.
	UniqueTokens() []*types.Token
	// GetTokenByChainAddress retrieves a token by chain ID and address.
	GetTokenByChainAddress(chainID uint64, addr common.Address) (*types.Token, bool)
	// GetTokensByChain returns all tokens for a specific chain.
	GetTokensByChain(chainID uint64) []*types.Token
	// GetTokensByKeys returns tokens by keys.
	GetTokensByKeys(keys []string) ([]*types.Token, error)

	// TokenLists returns all token lists.
	TokenLists() []*types.TokenList
	// TokenList returns a token list by ID.
	TokenList(id string) (*types.TokenList, bool)
}

// CustomTokenStore interface for storing and retrieving custom tokens.
type CustomTokenStore interface {
	// GetAll returns all custom tokens.
	GetAll() ([]*types.Token, error)
}
