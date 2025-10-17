package old

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	eth "github.com/ethereum/go-ethereum"
	ethereum "github.com/ethereum/go-ethereum"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

const (
	RewardPercentiles1 = 10.0
	RewardPercentiles2 = 45.0
	RewardPercentiles3 = 90.0
)

type GasFeeMode int

const (
	GasFeeLow GasFeeMode = iota
	GasFeeMedium
	GasFeeHigh
	GasFeeCustom
)

var (
	ErrCustomFeeModeNotAvailableInSuggestedFees = errors.New("custom fee mode is not available in suggested fees")
	ErrEIP1559IncompaibleChain                  = errors.New("EIP-1559 is not supported on this chain")
	ErrInvalidRewardData                        = errors.New("invalid reward data")
)

// NonEIP1559Fees represents the fees for non EIP-1559 compatible chains
type NonEIP1559Fees struct {
	GasPrice      *hexutil.Big `json:"gasPrice"`      // Gas price for the transaction used for non EIP-1559 compatible chains (in base unit of the chain eg. WEI for ETH or BNB)
	EstimatedTime uint         `json:"estimatedTime"` // Estimated time for the transaction in seconds, used for non EIP-1559 compatible chains
}

// MaxFeesLevels represents the max fees levels for low, medium and high fee modes and should be used for EIP-1559 compatible chains
type MaxFeesLevels struct {
	Low                 *hexutil.Big `json:"low"`                 // Low max fee per gas in WEI
	LowPriority         *hexutil.Big `json:"lowPriority"`         // Low priority fee in WEI
	LowEstimatedTime    uint         `json:"lowEstimatedTime"`    // Estimated time for low fees in seconds
	Medium              *hexutil.Big `json:"medium"`              // Medium max fee per gas in WEI
	MediumPriority      *hexutil.Big `json:"mediumPriority"`      // Medium priority fee in WEI
	MediumEstimatedTime uint         `json:"mediumEstimatedTime"` // Estimated time for medium fees in seconds
	High                *hexutil.Big `json:"high"`                // High max fee per gas in WEI
	HighPriority        *hexutil.Big `json:"highPriority"`        // High priority fee in WEI
	HighEstimatedTime   uint         `json:"highEstimatedTime"`   // Estimated time for high fees in seconds
}

type MaxPriorityFeesSuggestedBounds struct {
	Lower *big.Int // Lower bound for priority fee per gas in WEI
	Upper *big.Int // Upper bound for priority fee per gas in WEI
}

type SuggestedFees struct {
	// Fields that need to be removed once clients stop using them
	GasPrice             *big.Int   // TODO: remove once clients stop using this field, used for EIP-1559 incompatible chains, not in use anymore
	BaseFee              *big.Int   // TODO: remove once clients stop using this field, current network base fee (in ETH WEI), kept for backward compatibility
	MaxPriorityFeePerGas *big.Int   // TODO: remove once clients stop using this field, kept for backward compatibility
	L1GasFee             *big.Float // TODO: remove once clients stop using this field, not in use anymore

	// Fields in use
	NonEIP1559Fees                *NonEIP1559Fees                 // Fees for non EIP-1559 compatible chains
	MaxFeesLevels                 *MaxFeesLevels                  // Max fees levels for low, medium and high fee modes, should be used for EIP-1559 compatible chains
	MaxPriorityFeeSuggestedBounds *MaxPriorityFeesSuggestedBounds // Lower and upper bounds for priority fee per gas in WEI
	CurrentBaseFee                *big.Int                        // Current network base fee (in ETH WEI)
	EIP1559Enabled                bool                            // TODO: remove it since all chains we have support EIP-1559
}

type EthClient interface {
	FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error)
	BlockNumber(ctx context.Context) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethclient.BlockWithFullTxs, error)
	LineaEstimateGas(ctx context.Context, msg eth.CallMsg) (*ethclient.LineaEstimateGasResult, error)
}

type LineaEstimateGasResponse struct {
	BaseFeePerGas     *big.Int `json:"baseFeePerGas"`
	GasLimit          *big.Int `json:"gasLimit"`
	PriorityFeePerGas *big.Int `json:"priorityFeePerGas"`
}

type LineaEthClient interface {
	LineaEstimateGas(ctx context.Context, msg eth.CallMsg) (*LineaEstimateGasResponse, error)
}

type FeeManager struct {
	ethClient EthClient
}

func NewFeeManager(ethClient EthClient) *FeeManager {
	return &FeeManager{ethClient: ethClient}
}

func (f *FeeManager) IsEIP1559Enabled(ctx context.Context, chainID uint64) (bool, error) {
	eip1559Enabled, err := IsPartiallyOrFullyGaslessChainEIP1559Compatible(chainID)
	if err == nil {
		return eip1559Enabled, nil
	}

	block, err := f.ethClient.BlockByNumber(ctx, nil)
	if err != nil {
		return false, err
	}
	return block.BaseFeePerGas != nil && block.BaseFeePerGas.Cmp(big.NewInt(0)) > 0, nil
}

func (f *FeeManager) SuggestedFees(ctx context.Context, chainID uint64, address ethCommon.Address) (suggestedFees *SuggestedFees, noBaseFee bool, noPriorityFee bool, err error) {
	feeHistory, err := f.getFeeHistory(ctx, chainID, nil, []float64{RewardPercentiles1, RewardPercentiles2, RewardPercentiles3})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		suggestedFees, err = f.getNonEIP1559SuggestedFees(ctx, chainID)
		return
	}

	var (
		lowPriorityFeePerGasLowerBound *big.Int
		mediumPriorityFeePerGas        *big.Int
		maxPriorityFeePerGasUpperBound *big.Int
		baseFee                        *big.Int
	)

	if chainID == StatusNetworkSepolia {
		baseFee, lowPriorityFeePerGasLowerBound, err = f.getGaslessParamsForAccount(ctx, chainID, address)
		if err != nil {
			return
		}

		mediumPriorityFeePerGas = new(big.Int).Set(lowPriorityFeePerGasLowerBound)
		maxPriorityFeePerGasUpperBound = new(big.Int).Set(lowPriorityFeePerGasLowerBound)

		noBaseFee = baseFee == nil || baseFee.Cmp(ZeroBigIntValue()) == 0
		noPriorityFee = lowPriorityFeePerGasLowerBound == nil || lowPriorityFeePerGasLowerBound.Cmp(ZeroBigIntValue()) == 0
	} else {
		lowPriorityFeePerGasLowerBound, mediumPriorityFeePerGas, maxPriorityFeePerGasUpperBound, baseFee, err = getEIP1559SuggestedFees(chainID, feeHistory)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			suggestedFees, err = f.getNonEIP1559SuggestedFees(ctx, chainID)
			return
		}
	}

	suggestedFees = &SuggestedFees{
		GasPrice:             big.NewInt(0),
		BaseFee:              baseFee,
		CurrentBaseFee:       baseFee,
		MaxPriorityFeePerGas: mediumPriorityFeePerGas,
		MaxPriorityFeeSuggestedBounds: &MaxPriorityFeesSuggestedBounds{
			Lower: lowPriorityFeePerGasLowerBound,
			Upper: maxPriorityFeePerGasUpperBound,
		},
		EIP1559Enabled: true,
	}

	if chainID == EthereumMainnet || chainID == EthereumSepolia || chainID == AnvilMainnet {
		networkCongestion := calculateNetworkCongestion(feeHistory)

		baseFeeFloat := new(big.Float).SetUint64(baseFee.Uint64())
		baseFeeFloat.Mul(baseFeeFloat, big.NewFloat(networkCongestion))
		additionBasedOnCongestion := new(big.Int)
		baseFeeFloat.Int(additionBasedOnCongestion)

		mediumBaseFee := new(big.Int).Add(baseFee, additionBasedOnCongestion)

		highBaseFee := new(big.Int).Mul(baseFee, big.NewInt(2))
		highBaseFee.Add(highBaseFee, additionBasedOnCongestion)

		suggestedFees.MaxFeesLevels = &MaxFeesLevels{
			Low:            (*hexutil.Big)(new(big.Int).Add(baseFee, lowPriorityFeePerGasLowerBound)),
			LowPriority:    (*hexutil.Big)(new(big.Int).Set(lowPriorityFeePerGasLowerBound)),
			Medium:         (*hexutil.Big)(new(big.Int).Add(mediumBaseFee, mediumPriorityFeePerGas)),
			MediumPriority: (*hexutil.Big)(new(big.Int).Set(mediumPriorityFeePerGas)),
			High:           (*hexutil.Big)(new(big.Int).Add(highBaseFee, maxPriorityFeePerGasUpperBound)),
			HighPriority:   (*hexutil.Big)(new(big.Int).Set(maxPriorityFeePerGasUpperBound)),
		}
	} else {
		suggestedFees.MaxFeesLevels = &MaxFeesLevels{
			Low:            (*hexutil.Big)(new(big.Int).Add(baseFee, lowPriorityFeePerGasLowerBound)),
			LowPriority:    (*hexutil.Big)(new(big.Int).Set(lowPriorityFeePerGasLowerBound)),
			Medium:         (*hexutil.Big)(new(big.Int).Add(new(big.Int).Mul(baseFee, big.NewInt(4)), mediumPriorityFeePerGas)),
			MediumPriority: (*hexutil.Big)(new(big.Int).Set(mediumPriorityFeePerGas)),
			High:           (*hexutil.Big)(new(big.Int).Add(new(big.Int).Mul(baseFee, big.NewInt(10)), maxPriorityFeePerGasUpperBound)),
			HighPriority:   (*hexutil.Big)(new(big.Int).Set(maxPriorityFeePerGasUpperBound)),
		}
	}

	suggestedFees.MaxFeesLevels.LowEstimatedTime = estimatedTimeV2(feeHistory, suggestedFees.MaxFeesLevels.Low.ToInt(), suggestedFees.MaxFeesLevels.LowPriority.ToInt(), chainID, 1)
	suggestedFees.MaxFeesLevels.MediumEstimatedTime = estimatedTimeV2(feeHistory, suggestedFees.MaxFeesLevels.Medium.ToInt(), suggestedFees.MaxFeesLevels.MediumPriority.ToInt(), chainID, 1)
	suggestedFees.MaxFeesLevels.HighEstimatedTime = estimatedTimeV2(feeHistory, suggestedFees.MaxFeesLevels.High.ToInt(), suggestedFees.MaxFeesLevels.HighPriority.ToInt(), chainID, 1)

	return
}
