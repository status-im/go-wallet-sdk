package manager

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
)

var (
	ErrMainListIDNotProvided          = errors.New("main list ID is not provided")
	ErrMainListNotProvided            = errors.New("main list is not provided")
	ErrInitialListIDsNotProvided      = errors.New("initial list IDs are not provided")
	ErrInitialListProviderNotProvided = errors.New("initial list provider is not provided")
	ErrChainsNotProvided              = errors.New("chains are not provided")
)

// InitialListProvider provides the raw bytes for an initial token list by its ID.
// Implementations can load from disk, embedded assets, a database, etc.
type InitialListProvider func(id string) ([]byte, error)

// Config holds the configuration for manager.
type Config struct {
	AutoFetcherConfig *autofetcher.ConfigRemoteListOfTokenLists

	MainListID string // used to select the main list from the initial lists and process it first

	// initial lists are processed in alphabetical order of their IDs after the main list is processed
	InitialListIDs      []string                           // list of initial list IDs to process
	InitialListProvider InitialListProvider                // provider of initial list bytes by ID
	CustomParsers       map[string]parsers.TokenListParser // key: list ID, value: parser, is no match for the list ID, the StandardTokenList parser will be used

	Chains []uint64

	// SkippedTokenKeys is a list of token keys (format: "{chainID}-{lowercaseAddress}") that should be excluded from the
	// manager's *unique token collection* (e.g. UniqueTokens / GetTokenByChainAddress / GetTokensByChain / GetTokensByKeys).
	//
	// Note: this does NOT modify token lists themselves; `TokenList` / `TokenLists` still return the original lists as loaded.
	SkippedTokenKeys []string

	// AdditionalAddressesForNativeToken registers extra addresses that resolve to a chain's native token. Some chains
	// expose the native token at more than one address (e.g. on zkSync Era the native token is reachable both at the
	// zero address and at the system-contract alias 0x000000000000000000000000000000000000800a).
	//
	// Each registered address is surfaced through the manager's *unique token collection* APIs (UniqueTokens /
	// GetTokenByChainAddress / GetTokensByChain / GetTokensByKeys) as a distinct entry whose fields match the chain's
	// native token, except that Address is set to the registered address. It is NOT added to the "native" token list
	// returned by `TokenList` / `TokenLists`.
	//
	// Entries whose chain is not in `Chains` and entries equal to the zero address are ignored.
	AdditionalAddressesForNativeToken map[uint64][]common.Address
}

func (c *Config) Validate() error {
	if c.AutoFetcherConfig != nil {
		if err := c.AutoFetcherConfig.Validate(); err != nil {
			return err
		}
	}

	if c.MainListID == "" {
		return ErrMainListIDNotProvided
	}

	if len(c.InitialListIDs) == 0 {
		return ErrInitialListIDsNotProvided
	}

	if c.InitialListProvider == nil {
		return ErrInitialListProviderNotProvided
	}

	foundMain := false
	for _, id := range c.InitialListIDs {
		if id == c.MainListID {
			foundMain = true
			break
		}
	}
	if !foundMain {
		return ErrMainListNotProvided
	}

	if len(c.Chains) == 0 {
		return ErrChainsNotProvided
	}

	return nil
}
