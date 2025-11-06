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
	ErrInvalidName    = errors.New("invalid ENS name format")
	ErrEmptyName      = errors.New("ENS name cannot be empty")
	ErrInvalidAddress = errors.New("invalid Ethereum address")
)

// ENSRegistryAddress is the ENS registry contract address (same on all supported chains)
const ENSRegistryAddress = "0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e"

// Resolver handles ENS name resolution operations
type Resolver struct {
	client *ethclient.Client
}

// NewResolver creates a new ENS resolver instance
func NewResolver(client *ethclient.Client) (*Resolver, error) {
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}

	return &Resolver{
		client: client,
	}, nil
}

// ENSContractExists checks if the ENS registry contract is deployed on the connected chain
func ENSContractExists(ctx context.Context, client *ethclient.Client) (bool, error) {
	addr := common.HexToAddress(ENSRegistryAddress)
	code, err := client.CodeAt(ctx, addr, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check ENS registry: %w", err)
	}
	return len(code) > 0, nil
}

// AddressOf performs forward ENS resolution (name → address)
// Returns the Ethereum address associated with the ENS name
func (r *Resolver) AddressOf(name string) (common.Address, error) {
	if err := validateName(name); err != nil {
		return common.Address{}, err
	}

	// Normalize the name to lowercase
	name = strings.ToLower(strings.TrimSpace(name))

	address, err := ens.Resolve(r.client, name)
	if err != nil {
		return common.Address{}, err
	}

	return address, nil
}

// GetName performs reverse ENS resolution (address → name)
// Returns the ENS name associated with the Ethereum address
func (r *Resolver) GetName(address common.Address) (string, error) {
	if err := validateAddress(address); err != nil {
		return "", err
	}

	name, err := ens.ReverseResolve(r.client, address)
	if err != nil {
		return "", err
	}

	return name, nil
}

// validateName checks if the ENS name has valid structure
func validateName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrEmptyName
	}

	if !strings.Contains(name, ".") {
		return fmt.Errorf("%w: name must contain at least one dot (e.g., vitalik.eth)", ErrInvalidName)
	}

	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("%w: name cannot start or end with a dot", ErrInvalidName)
	}

	return nil
}

// validateAddress checks if the Ethereum address is valid
func validateAddress(address common.Address) error {
	if address == (common.Address{}) {
		return fmt.Errorf("%w: zero address", ErrInvalidAddress)
	}
	return nil
}
