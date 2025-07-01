package balancescanner

import (
	"github.com/ethereum/go-ethereum/common"

	sdkcommon "github.com/status-im/go-wallet-sdk/pkg/common"
)

type ContractData struct {
	Address        common.Address
	CreatedAtBlock uint
}

var contractDataByChainID = map[uint64]ContractData{
	sdkcommon.EthereumMainnet:      {common.HexToAddress("0x08A8fDBddc160A7d5b957256b903dCAb1aE512C5"), 12_194_222},
	sdkcommon.OptimismMainnet:      {common.HexToAddress("0x9e5076df494fc949abc4461f4e57592b81517d81"), 34_421_097},
	sdkcommon.ArbitrumMainnet:      {common.HexToAddress("0xbb85398092b83a016935a17fc857507b7851a071"), 70_031_945},
	sdkcommon.BaseMainnet:          {common.HexToAddress("0xc68c1e011cfE059EB94C8915c291502288704D89"), 24_567_587},
	sdkcommon.BSCMainnet:           {common.HexToAddress("0x71cfeb2ab5a3505f80b4c86f8ccd0a4b29f62447"), 47_746_468},
	sdkcommon.EthereumSepolia:      {common.HexToAddress("0xec21ebe1918e8975fc0cd0c7747d318c00c0acd5"), 4_366_506},
	sdkcommon.ArbitrumSepolia:      {common.HexToAddress("0xec21Ebe1918E8975FC0CD0c7747D318C00C0aCd5"), 553_947},
	sdkcommon.OptimismSepolia:      {common.HexToAddress("0xec21ebe1918e8975fc0cd0c7747d318c00c0acd5"), 7_362_011},
	sdkcommon.BaseSepolia:          {common.HexToAddress("0xc68c1e011cfE059EB94C8915c291502288704D89"), 20_078_235},
	sdkcommon.StatusNetworkSepolia: {common.HexToAddress("0xc68c1e011cfE059EB94C8915c291502288704D89"), 1_753_813},
	sdkcommon.BSCTestnet:           {common.HexToAddress("0x71cfeb2ab5a3505f80b4c86f8ccd0a4b29f62447"), 49_365_870},
}

func GetContractData(chainID uint64) *ContractData {
	contractData, exists := contractDataByChainID[chainID]
	if !exists {
		return nil
	}

	return &contractData
}
