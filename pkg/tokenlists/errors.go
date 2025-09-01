package tokenlists

import (
	"errors"
)

var (
	ErrConfigNotProvided                           = errors.New("config not provided")
	ErrLoggerNotProvided                           = errors.New("logger not provided")
	ErrMainListNotProvided                         = errors.New("main list not provided")
	ErrMainListIDNotProvided                       = errors.New("main list ID not provided")
	ErrMainListParserNotFound                      = errors.New("main list parser not found")
	ErrMainListIDCannotBeUsedAsInitialListID       = errors.New("main list ID cannot be used as an initial list ID")
	ErrInitialListParserNotFound                   = errors.New("initial list parser not found")
	ErrChainsNotProvided                           = errors.New("chains not provided")
	ErrAutoRefreshCheckIntervalGreaterThanInterval = errors.New("check interval must be <= refresh interval")
	ErrPrivacyGuardNotProvided                     = errors.New("privacy guard not provided")
	ErrLastTokenListsUpdateTimeStoreNotProvided    = errors.New("last token lists update time store not provided")
	ErrContentStoreNotProvided                     = errors.New("content store not provided")
	ErrChainNotAllowed                             = errors.New("chain not allowed")
	ErrInvalidAddressLength                        = errors.New("invalid address length")
	ErrSymbolCannotBeEmpty                         = errors.New("symbol cannot be empty")
	ErrDecimalsExceedsMaximum                      = errors.New("decimals exceeds maximum")
	ErrInvalidLogoURI                              = errors.New("invalid logo URI")
	ErrTokenListDoesNotMatchSchema                 = errors.New("token list does not match schema")
	ErrChannelClosed                               = errors.New("channel is closed")
	ErrTokenNotProvided                            = errors.New("token not provided")
)
