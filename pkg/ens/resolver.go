package ens

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ens "github.com/wealdtech/go-ens/v3"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

// Package-level errors
var (
	ErrInvalidName      = errors.New("invalid ENS name format")
	ErrEmptyName        = errors.New("ENS name cannot be empty")
	ErrInvalidAddress   = errors.New("invalid Ethereum address")
	ErrUnsupportedChain = errors.New("ENS not supported on this chain")
	ErrNoReverseRecord  = errors.New("no reverse record found for address")
)

// Supported chain IDs for ENS
const (
	MainnetChainID = 1
	SepoliaChainID = 11155111
)

// Resolver handles ENS name resolution operations
type Resolver struct {
	client  *ethclient.Client
	chainID uint64
}

// NewResolver creates a new ENS resolver instance
// The ethclient must be connected to a supported ENS chain (Ethereum Mainnet or Sepolia)
func NewResolver(client *ethclient.Client) (*Resolver, error) {
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}

	// Get chain ID to validate it's a supported ENS chain
	ctx := context.Background()
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Validate chain supports ENS
	chainIDUint := chainID.Uint64()
	if !isSupportedChain(chainIDUint) {
		return nil, fmt.Errorf("%w: chainID %d (supported: %d Mainnet, %d Sepolia)",
			ErrUnsupportedChain, chainIDUint, MainnetChainID, SepoliaChainID)
	}

	return &Resolver{
		client:  client,
		chainID: chainIDUint,
	}, nil
}

// AddressOf performs forward ENS resolution (name → address)
// Returns the Ethereum address associated with the ENS name
func (r *Resolver) AddressOf(ctx context.Context, name string) (common.Address, error) {
	if err := validateName(name); err != nil {
		return common.Address{}, err
	}

	// Normalize the name to lowercase
	name = strings.ToLower(strings.TrimSpace(name))

	// Resolve the ENS name to an address
	address, err := ens.Resolve(r.client, name)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to resolve ENS name %s: %w", name, err)
	}

	// Check if we got a valid address
	if address == (common.Address{}) {
		return common.Address{}, fmt.Errorf("no address found for ENS name: %s", name)
	}

	return address, nil
}

// GetName performs reverse ENS resolution (address → name)
// Returns the ENS name associated with the Ethereum address
func (r *Resolver) GetName(ctx context.Context, address common.Address) (string, error) {
	if err := validateAddress(address); err != nil {
		return "", err
	}

	// Perform reverse resolution
	name, err := ens.ReverseResolve(r.client, address)
	if err != nil {
		return "", fmt.Errorf("failed to reverse resolve address %s: %w", address.Hex(), err)
	}

	// Check if we got a valid name
	if name == "" {
		return "", ErrNoReverseRecord
	}

	return name, nil
}

// validateName checks if the ENS name is valid
func validateName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrEmptyName
	}

	// Basic validation: ENS names should contain at least one dot
	// and shouldn't start or end with a dot
	if !strings.Contains(name, ".") {
		return fmt.Errorf("%w: name must contain at least one dot (e.g., vitalik.eth)", ErrInvalidName)
	}

	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("%w: name cannot start or end with a dot", ErrInvalidName)
	}

	// Check for invalid characters (basic validation)
	// ENS names should only contain lowercase letters, numbers, and hyphens
	lowerName := strings.ToLower(name)
	for _, char := range lowerName {
		if (char < 'a' || char > 'z') &&
			(char < '0' || char > '9') &&
			char != '.' &&
			char != '-' {
			return fmt.Errorf("%w: name contains invalid character '%c'", ErrInvalidName, char)
		}
	}

	return nil
}

// validateAddress checks if the Ethereum address is valid
func validateAddress(address common.Address) error {
	// Check for zero address
	if address == (common.Address{}) {
		return fmt.Errorf("%w: zero address", ErrInvalidAddress)
	}
	return nil
}

// isSupportedChain checks if the chain ID is supported for ENS
func isSupportedChain(chainID uint64) bool {
	return chainID == MainnetChainID || chainID == SepoliaChainID
}
