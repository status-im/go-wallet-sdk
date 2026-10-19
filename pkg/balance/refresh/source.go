package refresh

import (
	"slices"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
	"github.com/status-im/go-wallet-sdk/pkg/eventlog"
)

// ChangeSource tells where balances may have moved between refreshes, as
// TransferDetected and DetectionGap events. Transfer detection is one;
// native transfer logs or log subscriptions can be others.
type ChangeSource interface {
	// Watch tells emit what the source finds until stop is called.
	Watch(emit func(Event)) (stop func())
	// Scan makes the source look for changes since its last scan. The caller
	// scans with every periodic refresh and when the user asks for one.
	Scan()
}

// NativeTransferLogAddress is the system address that emits native transfer
// logs (EIP-7708).
var NativeTransferLogAddress = common.HexToAddress("0xfffffffffffffffffffffffffffffffffffffffe")

// TransfersFromLogs are the transfers that logs found on chain show for
// accounts, one event per account. Only the tokens known says are kept, the
// rest is spam; a native transfer log is a transfer of the native coin.
func TransfersFromLogs(chain uint64, accounts []common.Address, logs []eventlog.Event, known func(common.Address) bool) []TransferDetected {
	var transfers []TransferDetected
	for _, account := range accounts {
		transfer := TransferDetected{Pair: Pair{Account: account, Chain: chain}}
		for _, log := range logs {
			t, ok := log.Unpacked.(erc20.Erc20Transfer)
			if !ok || log.EventKey != eventlog.ERC20Transfer || (t.From != account && t.To != account) {
				continue
			}
			switch contract := t.Raw.Address; {
			case contract == NativeTransferLogAddress:
				transfer.Native = true
			case known(contract) && !slices.Contains(transfer.Contracts, contract):
				transfer.Contracts = append(transfer.Contracts, contract)
			}
		}
		if transfer.Native || len(transfer.Contracts) > 0 {
			transfers = append(transfers, transfer)
		}
	}
	return transfers
}
