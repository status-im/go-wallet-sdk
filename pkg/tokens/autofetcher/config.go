package autofetcher

import (
	"errors"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

var (
	ErrAutoRefreshCheckIntervalGreaterThanInterval = errors.New("check interval must be <= refresh interval")
	ErrRemoteListOfTokenListsParserNotProvided     = errors.New("remote list of token lists parser is required")
	ErrTokenListsNotProvided                       = errors.New("token lists are required")
)

type Config struct {
	LastUpdate               time.Time
	AutoRefreshInterval      time.Duration
	AutoRefreshCheckInterval time.Duration
}

type ConfigRemoteListOfTokenLists struct {
	Config
	RemoteListOfTokenListsFetchDetails types.ListDetails
	RemoteListOfTokenListsParser       parsers.ListOfTokenListsParser
}

type ConfigTokenLists struct {
	Config
	TokenLists []types.ListDetails
}

func (c *Config) Validate() error {
	if c.AutoRefreshCheckInterval > c.AutoRefreshInterval {
		return ErrAutoRefreshCheckIntervalGreaterThanInterval
	}
	return nil
}

func (c *ConfigRemoteListOfTokenLists) Validate() error {
	if err := c.RemoteListOfTokenListsFetchDetails.Validate(); err != nil {
		return err
	}

	if c.RemoteListOfTokenListsParser == nil {
		return ErrRemoteListOfTokenListsParserNotProvided
	}

	return c.Config.Validate()
}

func (c *ConfigTokenLists) Validate() error {
	if len(c.TokenLists) == 0 {
		return ErrTokenListsNotProvided
	}

	for _, tokenList := range c.TokenLists {
		if err := tokenList.Validate(); err != nil {
			return err
		}
	}

	return c.Config.Validate()
}
