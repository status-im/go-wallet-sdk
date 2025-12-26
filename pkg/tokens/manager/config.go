package manager

import (
	"errors"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
)

var (
	ErrMainListIDNotProvided = errors.New("main list ID is not provided")
	ErrMainListNotProvided   = errors.New("main list is not provided")
	ErrChainsNotProvided     = errors.New("chains are not provided")
)

// Config holds the configuration for manager.
type Config struct {
	AutoFetcherConfig *autofetcher.ConfigRemoteListOfTokenLists

	MainListID string // used to select the main list from the initial lists and process it first

	// initial lists are processed in alphabetical order of their IDs after the main list is processed
	InitialLists  map[string][]byte                  // key: list ID, value: list data
	CustomParsers map[string]parsers.TokenListParser // key: list ID, value: parser, is no match for the list ID, the StandardTokenList parser will be used

	Chains []uint64

	SkippedTokenKeys []string // list of token keys (in the format: "{chainID}-{lowercaseAddress}") that should be excluded from token list
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
	_, existsInInitialLists := c.InitialLists[c.MainListID]
	if !existsInInitialLists {
		return ErrMainListNotProvided
	}

	if len(c.Chains) == 0 {
		return ErrChainsNotProvided
	}

	return nil
}
