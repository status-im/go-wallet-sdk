package refresh_test

import (
	"math/big"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/balance/refresh"
	"github.com/status-im/go-wallet-sdk/pkg/eventlog"
)

var (
	carol = common.HexToAddress("0xca201")
	usdc  = common.HexToAddress("0x05")
	dai   = common.HexToAddress("0x06")
	spam  = common.HexToAddress("0x5BA3")
	punks = common.HexToAddress("0x9C")
)

func known(token common.Address) bool {
	return slices.Contains([]common.Address{usdc, dai, punks}, token)
}

func transferLog(contract, from, to common.Address) eventlog.Event {
	events := eventlog.ParseLog(types.Log{
		Address: contract,
		Topics:  []common.Hash{eventlog.ERC20TransferID, common.BytesToHash(from[:]), common.BytesToHash(to[:])},
		Data:    common.BigToHash(big.NewInt(1)).Bytes(),
	})
	if len(events) != 1 {
		panic("not one ERC20 transfer")
	}
	return events[0]
}

func nftTransferLog(contract, from, to common.Address) eventlog.Event {
	events := eventlog.ParseLog(types.Log{
		Address: contract,
		Topics: []common.Hash{eventlog.ERC20TransferID, common.BytesToHash(from[:]), common.BytesToHash(to[:]),
			common.BigToHash(big.NewInt(7))},
	})
	if len(events) != 1 {
		panic("not one ERC721 transfer")
	}
	return events[0]
}

func TestTransferOfAKnownTokenIsDetectedForTheWatchedAccount(t *testing.T) {
	logs := []eventlog.Event{transferLog(usdc, carol, alice)}

	require.Equal(t, []refresh.TransferDetected{
		{Pair: refresh.Pair{Account: alice, Chain: 10}, Contracts: []common.Address{usdc}},
	}, refresh.TransfersFromLogs(10, []common.Address{alice}, logs, known))
}

func TestTransferBetweenWatchedAccountsIsDetectedForBoth(t *testing.T) {
	logs := []eventlog.Event{transferLog(usdc, alice, bob)}

	require.Equal(t, []refresh.TransferDetected{
		{Pair: refresh.Pair{Account: alice, Chain: 1}, Contracts: []common.Address{usdc}},
		{Pair: refresh.Pair{Account: bob, Chain: 1}, Contracts: []common.Address{usdc}},
	}, refresh.TransfersFromLogs(1, []common.Address{alice, bob}, logs, known))
}

func TestTransferOfAnUnknownTokenIsSpam(t *testing.T) {
	logs := []eventlog.Event{transferLog(spam, carol, alice)}

	require.Empty(t, refresh.TransfersFromLogs(1, []common.Address{alice}, logs, known))
}

func TestNativeTransferLogIsATransferOfTheNativeCoin(t *testing.T) {
	logs := []eventlog.Event{transferLog(refresh.NativeTransferLogAddress, carol, alice)}

	require.Equal(t, []refresh.TransferDetected{
		{Pair: refresh.Pair{Account: alice, Chain: 1}, Native: true},
	}, refresh.TransfersFromLogs(1, []common.Address{alice}, logs, known))
}

func TestTransfersOfOneAccountMakeOneEvent(t *testing.T) {
	logs := []eventlog.Event{
		transferLog(usdc, carol, alice),
		transferLog(refresh.NativeTransferLogAddress, alice, carol),
		transferLog(dai, alice, carol),
		transferLog(usdc, alice, carol),
	}

	require.Equal(t, []refresh.TransferDetected{
		{Pair: refresh.Pair{Account: alice, Chain: 1}, Contracts: []common.Address{usdc, dai}, Native: true},
	}, refresh.TransfersFromLogs(1, []common.Address{alice}, logs, known))
}

func TestCollectibleTransferIsNotABalanceTransfer(t *testing.T) {
	logs := []eventlog.Event{nftTransferLog(punks, carol, alice)}

	require.Empty(t, refresh.TransfersFromLogs(1, []common.Address{alice}, logs, known))
}

func TestTransferOfOtherAccountsIsIgnored(t *testing.T) {
	logs := []eventlog.Event{transferLog(usdc, carol, bob)}

	require.Empty(t, refresh.TransfersFromLogs(1, []common.Address{alice}, logs, known))
}
