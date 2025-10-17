package infura

import (
	"fmt"
	"slices"
)

// GasPrice represents the gas price for a given network
type GasPrice struct {
	SuggestedMaxPriorityFeePerGas string `json:"suggestedMaxPriorityFeePerGas"` // in gwei
	SuggestedMaxFeePerGas         string `json:"suggestedMaxFeePerGas"`         // in gwei
	MinWaitTimeEstimate           int    `json:"minWaitTimeEstimate"`           // in seconds
	MaxWaitTimeEstimate           int    `json:"maxWaitTimeEstimate"`           // in seconds
}

// GasResponse represents Infura's Gas API response format
type GasResponse struct {
	Low                        GasPrice `json:"low"`
	Medium                     GasPrice `json:"medium"`
	High                       GasPrice `json:"high"`
	EstimatedBaseFee           string   `json:"estimatedBaseFee"`
	NetworkCongestion          float64  `json:"networkCongestion"` // [0-1]
	LatestPriorityFeeRange     []string `json:"latestPriorityFeeRange,omitempty"`
	HistoricalPriorityFeeRange []string `json:"historicalPriorityFeeRange,omitempty"`
	HistoricalBaseFeeRange     []string `json:"historicalBaseFeeRange,omitempty"`
	PriorityFeeTrend           string   `json:"priorityFeeTrend,omitempty"` // "up", "down"
	BaseFeeTrend               string   `json:"baseFeeTrend,omitempty"`     // "up", "down"
	Version                    string   `json:"version,omitempty"`
}

// Supported network IDs for Infura Gas API
// Complete list from Infura's Gas API documentation
const (
	// Ethereum Networks
	Ethereum        = 1
	Optimism        = 10
	Cronos          = 25
	BNB             = 56
	BNBTestnet      = 97
	Gnosis          = 100
	Unichain        = 130
	Polygon         = 137
	Monad           = 143
	Sonic           = 146
	Manta           = 169
	Mint            = 185
	OpBNB           = 204
	Mind            = 228
	Lens            = 232
	Fantom          = 250
	Fraxtal         = 252
	Orderly         = 291
	Filecoin        = 314
	ZkSync          = 324
	Redstone        = 690
	Matchain        = 698
	Flow            = 747
	Rivalz          = 752
	Lyra            = 957
	Lisk            = 1135
	UnichainSepolia = 1301
	Sei             = 1329
	Gravity         = 1625
	Reya            = 1729
	Playblock       = 1829
	Soneium         = 1868
	Lydia           = 1989
	Sanko           = 1996
	Edgeless        = 2026
	Game7           = 2187
	Dogelon         = 2420
	Injective       = 2525
	Abstract        = 2741
	Hytopia         = 2911
	BotanixTestnet  = 3636
	Botanix         = 3637
	Cometh          = 4078
	SXRollup        = 4162
	Trumpchain      = 4547
	API3            = 4913
	Mantle          = 5000
	Ham             = 5112
	Duck            = 5545
	OpBNBTestnet    = 5611
	MegaethTestnet  = 6342
	ZetaChain       = 7000
	ZetaChainTest   = 7001
	Kinto           = 7887
	Base            = 8453
	Clink           = 8818
	MonadTestnet    = 10143
	Huddle01        = 12323
	Immutable       = 13371
	ImmutableTest   = 13473
	EthereumHolesky = 17000
	OnchainPoints   = 17071
	Everclear       = 25327
	SlingshotDAO    = 33401
	Mode            = 34443
	AlephZero       = 41455
	Educhain        = 41923
	ArbitrumOne     = 42161
	ArbitrumNova    = 42170
	Avalanche       = 43114
	Blessnet        = 45513
	Chainbounty     = 51828
	Dodo            = 53456
	Superposition   = 55244
	LineaSepolia    = 59141
	Linea           = 59144
	ProofOfPlayApex = 70700
	ProofOfPlayBoss = 70701
	Fusion          = 75978
	PolygonMumbai   = 80001
	PolygonAmoy     = 80002
	Berachain       = 80094
	GeoGenesis      = 80451
	Onyx            = 80888
	Forta           = 80931
	Blast           = 81457
	Vemp            = 82614
	BaseSepolia     = 84532
	Unite           = 88899
	Henez           = 91111
	Miracle         = 92278
	Lumiterra       = 94168
	Idex            = 94524
	PlumeTestnet    = 98864
	Plume           = 98866
	Real            = 111188
	Eventum         = 161803
	Taiko           = 167000
	TaikoKatla      = 167008
	TaikoHekla      = 167009
	BlastSepolia    = 168587773
	Blockfit        = 202424
	Cheese          = 383353
	LayerK          = 529375
	ScrollSepolia   = 534351
	Scroll          = 534352
	Hoodi           = 560048
	Xai             = 660279
	Conwai          = 668668
	Katana          = 747474
	Winr            = 777777
	Logx            = 936369
	Scorekount      = 1000080
	Zora            = 7777777
	Fluence         = 9999999
	Spotlight       = 10058111
	AlienX          = 10241024
	AlienXTestnet   = 10241025
	EthereumSepolia = 11155111
	OptimismSepolia = 11155420
	Deri            = 20231119
	Corn            = 21000000
	DegenChain      = 666666666
	Anxient         = 8888888888
	Rarible         = 1380012617
)

// GetNetworkName returns the human-readable name for a network ID
func GetNetworkName(networkID int) string {
	networks := map[int]string{
		// Major Networks
		Ethereum:        "Ethereum",
		EthereumSepolia: "Ethereum Sepolia",
		EthereumHolesky: "Ethereum Holesky",
		Optimism:        "Optimism",
		OptimismSepolia: "Optimism Sepolia",
		Polygon:         "Polygon",
		PolygonMumbai:   "Polygon Mumbai",
		PolygonAmoy:     "Polygon Amoy",
		ArbitrumOne:     "Arbitrum One",
		ArbitrumNova:    "Arbitrum Nova",
		Base:            "Base",
		BaseSepolia:     "Base Sepolia",
		Linea:           "Linea",
		LineaSepolia:    "Linea Sepolia",
		BNB:             "BNB Smart Chain",
		BNBTestnet:      "BNB Testnet",
		OpBNB:           "opBNB",
		OpBNBTestnet:    "opBNB Testnet",
		Avalanche:       "Avalanche",
		Fantom:          "Fantom",
		Gnosis:          "Gnosis",
		Cronos:          "Cronos",

		// Layer 2 & Scaling
		Scroll:        "Scroll",
		ScrollSepolia: "Scroll Sepolia",
		ZkSync:        "zkSync Era",
		Mantle:        "Mantle",
		Blast:         "Blast",
		BlastSepolia:  "Blast Sepolia",
		Taiko:         "Taiko",
		TaikoKatla:    "Taiko Katla",
		TaikoHekla:    "Taiko Hekla",
		Zora:          "Zora",

		// Gaming & Entertainment
		Immutable:     "Immutable",
		ImmutableTest: "Immutable Testnet",
		Xai:           "Xai",
		SlingshotDAO:  "SlingShot",

		DegenChain: "Degen Chain",

		// DeFi & Trading
		Deri:      "Deri",
		Logx:      "LogX",
		Injective: "Injective",

		// Specialized Networks
		Unichain:        "Unichain",
		UnichainSepolia: "Unichain Sepolia",
		Sei:             "Sei",
		Flow:            "Flow",
		Filecoin:        "Filecoin",
		ZetaChain:       "ZetaChain",
		ZetaChainTest:   "ZetaChain Testnet",

		// Additional Networks
		Abstract:       "Abstract",
		AlienX:         "AlienX",
		AlienXTestnet:  "AlienX Testnet",
		AlephZero:      "Aleph Zero",
		Anxient:        "Anxient",
		API3:           "API3",
		Berachain:      "Berachain",
		Blessnet:       "Blessnet",
		Blockfit:       "Blockfit",
		Botanix:        "Botanix",
		BotanixTestnet: "Botanix Testnet",
		Chainbounty:    "Chainbounty",
		Cheese:         "Cheese",
		// Citronus network not in constants list
		Clink:           "Clink",
		Cometh:          "Cometh",
		Conwai:          "Conwai",
		Corn:            "Corn",
		Dodo:            "Dodo",
		Duck:            "Duck",
		Edgeless:        "Edgeless",
		Educhain:        "Educhain",
		Eventum:         "Eventum",
		Everclear:       "Everclear",
		Fluence:         "Fluence",
		Forta:           "Forta",
		Fraxtal:         "Fraxtal",
		Fusion:          "Fusion",
		Game7:           "Game7",
		GeoGenesis:      "Geo Genesis",
		Gravity:         "Gravity",
		Ham:             "Ham",
		Henez:           "Henez",
		Hoodi:           "Hoodi",
		Huddle01:        "Huddle01",
		Hytopia:         "Hytopia",
		Katana:          "Katana",
		Kinto:           "Kinto",
		LayerK:          "Layer K",
		Lens:            "Lens",
		Lisk:            "Lisk",
		Lumiterra:       "Lumiterra",
		Lydia:           "Lydia",
		Lyra:            "Lyra",
		Manta:           "Manta",
		Matchain:        "Matchain",
		MegaethTestnet:  "Megaeth Testnet",
		Mind:            "Mind",
		Mint:            "Mint",
		Miracle:         "Miracle",
		Mode:            "Mode",
		Monad:           "Monad",
		MonadTestnet:    "Monad Testnet",
		OnchainPoints:   "Onchain Points",
		Onyx:            "Onyx",
		Orderly:         "Orderly",
		Playblock:       "Playblock",
		Plume:           "Plume",
		PlumeTestnet:    "Plume Testnet",
		ProofOfPlayApex: "Proof of Play Apex",
		ProofOfPlayBoss: "Proof of Play Boss",
		Rarible:         "Rarible",
		Real:            "Real",
		Redstone:        "Redstone",
		Reya:            "Reya",
		Rivalz:          "Rivalz",
		Sanko:           "Sanko",
		Scorekount:      "Scorekount",
		Soneium:         "Soneium",
		Sonic:           "Sonic",
		Spotlight:       "Spotlight",
		Superposition:   "Superposition",
		SXRollup:        "SX Rollup",
		Trumpchain:      "Trumpchain",
		Unite:           "Unite",
		Vemp:            "Vemp",
		Winr:            "Winr",
	}

	if name, exists := networks[networkID]; exists {
		return name
	}
	return fmt.Sprintf("Network %d", networkID)
}

// IsSupported checks if a network ID is supported by Infura's Gas API
func IsSupported(networkID int) bool {
	supportedNetworks := []int{
		Ethereum, Optimism, Cronos, BNB, BNBTestnet, Gnosis, Unichain, Polygon, Monad, Sonic,
		Manta, Mint, OpBNB, Mind, Lens, Fantom, Fraxtal, Orderly, Filecoin, ZkSync,
		Redstone, Matchain, Flow, Rivalz, Lyra, Lisk, UnichainSepolia, Sei, Gravity, Reya,
		Playblock, Soneium, Lydia, Sanko, Edgeless, Game7, Dogelon, Injective, Abstract, Hytopia,
		BotanixTestnet, Botanix, Cometh, SXRollup, Trumpchain, API3, Mantle, Ham, Duck, OpBNBTestnet,
		MegaethTestnet, ZetaChain, ZetaChainTest, Kinto, Base, Clink, MonadTestnet, Huddle01,
		Immutable, ImmutableTest, EthereumHolesky, OnchainPoints, Everclear, SlingshotDAO, Mode,
		AlephZero, Educhain, ArbitrumOne, ArbitrumNova, Avalanche, Blessnet, Chainbounty, Dodo,
		Superposition, LineaSepolia, Linea, ProofOfPlayApex, ProofOfPlayBoss, Fusion, PolygonMumbai,
		PolygonAmoy, Berachain, GeoGenesis, Onyx, Forta, Blast, Vemp, BaseSepolia, Unite,
		Henez, Miracle, Lumiterra, Idex, PlumeTestnet, Plume, Real, Eventum, Taiko,
		TaikoKatla, TaikoHekla, BlastSepolia, Blockfit, Cheese, LayerK, ScrollSepolia, Scroll,
		Hoodi, Xai, Conwai, Katana, Winr, Logx, Scorekount, Zora, Fluence, Spotlight,
		AlienX, AlienXTestnet, EthereumSepolia, OptimismSepolia, Deri, Corn, DegenChain, Anxient, Rarible,
	}

	return slices.Contains(supportedNetworks, networkID)
}
