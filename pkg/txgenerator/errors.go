package txgenerator

import "errors"

var (
	// ErrInvalidParams is returned when transaction parameters are invalid
	ErrInvalidParams = errors.New("invalid transaction parameters")
	// ErrMissingGasPrice is returned when GasPrice is required but not provided
	ErrMissingGasPrice = errors.New("gas price is required for legacy transactions")
	// ErrMissingMaxFeePerGas is returned when MaxFeePerGas is required but not provided
	ErrMissingMaxFeePerGas = errors.New("max fee per gas is required for EIP-1559 transactions")
	// ErrMissingMaxPriorityFeePerGas is returned when MaxPriorityFeePerGas is required but not provided
	ErrMissingMaxPriorityFeePerGas = errors.New("max priority fee per gas is required for EIP-1559 transactions")
)
