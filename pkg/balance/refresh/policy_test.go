package refresh_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/balance/refresh"
)

var (
	t0    = time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	alice = common.HexToAddress("0xa11ce")
	bob   = common.HexToAddress("0xb0b")
)

func TestStartReadsEverythingAndWakesAfterOnScreenPeriod(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())

	d := p.Handle(t0, refresh.Started{})

	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, d.Reads)
	require.Equal(t, t0.Add(2*time.Minute), d.WakeAt)
}

func TestTickReadsEverythingOncePerOnScreenPeriod(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	early := p.Handle(t0.Add(time.Minute), refresh.Tick{})
	require.Empty(t, early.Reads)
	require.Equal(t, t0.Add(2*time.Minute), early.WakeAt)

	due := p.Handle(t0.Add(2*time.Minute), refresh.Tick{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, due.Reads)
	require.Equal(t, t0.Add(4*time.Minute), due.WakeAt)
}

func TestHiddenBalancesAreReadEveryHiddenPeriodSinceTheLastPeriodicRead(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	hidden := p.Handle(t0.Add(30*time.Second), refresh.OnScreenChanged{OnScreen: false})
	require.Empty(t, hidden.Reads)
	require.Equal(t, t0.Add(15*time.Minute), hidden.WakeAt)

	require.Empty(t, p.Handle(t0.Add(2*time.Minute), refresh.Tick{}).Reads)

	due := p.Handle(t0.Add(15*time.Minute), refresh.Tick{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, due.Reads)
	require.Equal(t, t0.Add(30*time.Minute), due.WakeAt)
}

func TestBackOnScreenReadsAtOnceOnlyWhenTheLastPeriodicReadIsStale(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})
	p.Handle(t0.Add(30*time.Second), refresh.OnScreenChanged{OnScreen: false})

	soon := p.Handle(t0.Add(time.Minute), refresh.OnScreenChanged{OnScreen: true})
	require.Empty(t, soon.Reads)
	require.Equal(t, t0.Add(2*time.Minute), soon.WakeAt)

	p.Handle(t0.Add(90*time.Second), refresh.OnScreenChanged{OnScreen: false})
	late := p.Handle(t0.Add(10*time.Minute), refresh.OnScreenChanged{OnScreen: true})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, late.Reads)
	require.Equal(t, t0.Add(12*time.Minute), late.WakeAt)
}

func TestTokenListsUpdateReadsEverythingAfterDebounceWithoutMovingThePeriod(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	updated := p.Handle(t0.Add(30*time.Second), refresh.TokenListsUpdated{})
	require.Empty(t, updated.Reads)
	require.Equal(t, t0.Add(40*time.Second), updated.WakeAt)

	debounced := p.Handle(t0.Add(40*time.Second), refresh.Tick{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, debounced.Reads)
	require.Equal(t, t0.Add(2*time.Minute), debounced.WakeAt)
}

func TestBurstOfEventsMakesOneReadAfterTheLastOne(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	p.Handle(t0.Add(30*time.Second), refresh.TokenListsUpdated{})
	second := p.Handle(t0.Add(35*time.Second), refresh.TokenListsUpdated{})
	require.Equal(t, t0.Add(45*time.Second), second.WakeAt)

	require.Empty(t, p.Handle(t0.Add(40*time.Second), refresh.Tick{}).Reads)
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, p.Handle(t0.Add(45*time.Second), refresh.Tick{}).Reads)
}

func TestNewAccountsOrChainsReadWhatIsOlderThanThePeriod(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	changed := p.Handle(t0.Add(30*time.Second), refresh.AccountsOrChainsChanged{})
	require.Empty(t, changed.Reads)
	require.Equal(t, t0.Add(40*time.Second), changed.WakeAt)

	stale := refresh.Read{Tokens: refresh.Full, Scope: refresh.Scope{OlderThan: 2 * time.Minute}}
	require.Equal(t, []refresh.Read{stale}, p.Handle(t0.Add(40*time.Second), refresh.Tick{}).Reads)
}

func TestRequestedRefreshReadsEverythingAfterDebounce(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	requested := p.Handle(t0.Add(30*time.Second), refresh.RefreshRequested{})
	require.Empty(t, requested.Reads)
	require.Equal(t, t0.Add(40*time.Second), requested.WakeAt)

	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, p.Handle(t0.Add(40*time.Second), refresh.Tick{}).Reads)
}

func TestReadingEverythingCoversNarrowerReadsOfTheSameBurst(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	p.Handle(t0.Add(30*time.Second), refresh.AccountsOrChainsChanged{})
	p.Handle(t0.Add(31*time.Second), refresh.TokenListsUpdated{})
	p.Handle(t0.Add(32*time.Second), refresh.AccountsOrChainsChanged{})

	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, p.Handle(t0.Add(42*time.Second), refresh.Tick{}).Reads)
}

func TestLateTickMakesOneReadWhenPeriodicAndDebouncedReadsAreBothDue(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})
	p.Handle(t0.Add(100*time.Second), refresh.AccountsOrChainsChanged{})

	late := p.Handle(t0.Add(3*time.Minute), refresh.Tick{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, late.Reads)
	require.Equal(t, t0.Add(5*time.Minute), late.WakeAt)
}

func TestEndedTransactionReadsItsAccountsAndChainsAfterDebounce(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	ended := p.Handle(t0.Add(30*time.Second), refresh.TransactionEnded{
		Pairs: []refresh.Pair{{Account: alice, Chain: 1}, {Account: bob, Chain: 10}},
	})
	require.Empty(t, ended.Reads)
	require.Equal(t, t0.Add(40*time.Second), ended.WakeAt)

	require.Equal(t, []refresh.Read{
		{Tokens: refresh.Full, Scope: refresh.Scope{Pair: refresh.Pair{Account: alice, Chain: 1}}},
		{Tokens: refresh.Full, Scope: refresh.Scope{Pair: refresh.Pair{Account: bob, Chain: 10}}},
	}, p.Handle(t0.Add(40*time.Second), refresh.Tick{}).Reads)
}

func TestConfirmedTransactionReadsItsBalancesAtOnce(t *testing.T) {
	p := refresh.NewPolicy(refresh.DefaultConfig())
	p.Handle(t0, refresh.Started{})

	confirmed := p.Handle(t0.Add(30*time.Second), refresh.TransactionEnded{
		Pairs:     []refresh.Pair{{Account: alice, Chain: 1}},
		Kinds:     refresh.Native | refresh.ERC20,
		Confirmed: true,
	})
	require.Equal(t, []refresh.Read{{
		Tokens: refresh.Full,
		Scope:  refresh.Scope{Pair: refresh.Pair{Account: alice, Chain: 1}},
		Kinds:  refresh.Native | refresh.ERC20,
	}}, confirmed.Reads)
	require.Equal(t, t0.Add(2*time.Minute), confirmed.WakeAt)
}

func heldConfig() refresh.Config {
	c := refresh.DefaultConfig()
	c.Schedule = refresh.HeldEveryTick
	return c
}

func TestHeldScheduleReadsEverythingOnStartThenTheHeldSetEveryTick(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())

	start := p.Handle(t0, refresh.Started{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, start.Reads)
	require.Equal(t, t0.Add(2*time.Minute), start.WakeAt)

	tick := p.Handle(t0.Add(2*time.Minute), refresh.Tick{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Held}}, tick.Reads)
	require.Equal(t, t0.Add(4*time.Minute), tick.WakeAt)
}

func TestHeldScheduleReadsEverythingHourlyWhileOnScreen(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})

	for at := 2 * time.Minute; at < time.Hour; at += 2 * time.Minute {
		require.Equal(t, []refresh.Read{{Tokens: refresh.Held}}, p.Handle(t0.Add(at), refresh.Tick{}).Reads, at)
	}
	hourly := p.Handle(t0.Add(time.Hour), refresh.Tick{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, hourly.Reads)
	require.Equal(t, t0.Add(62*time.Minute), hourly.WakeAt)
}

func TestHeldScheduleRequestedRefreshReadsTheHeldSetSoonAfterAFullRefresh(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})
	p.Handle(t0.Add(2*time.Minute), refresh.Tick{})
	p.Handle(t0.Add(4*time.Minute), refresh.Tick{})

	requested := p.Handle(t0.Add(5*time.Minute), refresh.RefreshRequested{})
	require.Empty(t, requested.Reads)
	require.Equal(t, t0.Add(5*time.Minute+10*time.Second), requested.WakeAt)

	require.Equal(t, []refresh.Read{{Tokens: refresh.Held}}, p.Handle(t0.Add(5*time.Minute+10*time.Second), refresh.Tick{}).Reads)
}

func TestHeldScheduleRequestedRefreshReadsEverythingOnceTheLastFullIsStaleAndMovesTheHourlyOne(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})
	for at := 2 * time.Minute; at <= 10*time.Minute; at += 2 * time.Minute {
		p.Handle(t0.Add(at), refresh.Tick{})
	}

	p.Handle(t0.Add(11*time.Minute), refresh.RefreshRequested{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, p.Handle(t0.Add(11*time.Minute+10*time.Second), refresh.Tick{}).Reads)

	for at := 12 * time.Minute; at <= 70*time.Minute; at += 2 * time.Minute {
		require.Equal(t, []refresh.Read{{Tokens: refresh.Held}}, p.Handle(t0.Add(at), refresh.Tick{}).Reads, at)
	}
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, p.Handle(t0.Add(72*time.Minute), refresh.Tick{}).Reads,
		"the hourly full refresh counts from the requested one")
}

func TestHiddenHeldScheduleRequestedRefreshReadsEverythingOnceTheLastFullIsStale(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})
	p.Handle(t0.Add(time.Minute), refresh.OnScreenChanged{OnScreen: false})
	p.Handle(t0.Add(15*time.Minute), refresh.Tick{})

	p.Handle(t0.Add(20*time.Minute), refresh.RefreshRequested{})

	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, p.Handle(t0.Add(20*time.Minute+10*time.Second), refresh.Tick{}).Reads)
}

func TestHiddenHeldScheduleSkipsTheHourlyReadUntilBalancesAreBackOnScreen(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})
	p.Handle(t0.Add(time.Minute), refresh.OnScreenChanged{OnScreen: false})

	for at := 15 * time.Minute; at <= 75*time.Minute; at += 15 * time.Minute {
		require.Equal(t, []refresh.Read{{Tokens: refresh.Held}}, p.Handle(t0.Add(at), refresh.Tick{}).Reads, at)
	}
	back := p.Handle(t0.Add(76*time.Minute), refresh.OnScreenChanged{OnScreen: true})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, back.Reads)
	require.Equal(t, t0.Add(78*time.Minute), back.WakeAt)
}

func TestHeldScheduleResumesWithTheHeldSetAfterAShortPauseAndEverythingAfterALongOne(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})

	stopped := p.Handle(t0.Add(10*time.Minute), refresh.Stopped{})
	require.Empty(t, stopped.Reads)
	require.Zero(t, stopped.WakeAt)
	short := p.Handle(t0.Add(40*time.Minute), refresh.Started{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Held}}, short.Reads)
	require.Equal(t, t0.Add(42*time.Minute), short.WakeAt)

	p.Handle(t0.Add(50*time.Minute), refresh.Stopped{})
	long := p.Handle(t0.Add(3*time.Hour), refresh.Started{})
	require.Equal(t, []refresh.Read{{Tokens: refresh.Full}}, long.Reads)
}

func TestHeldScheduleReadsOnlyWhatWasNeverReadAfterListsOrAccountsChange(t *testing.T) {
	unread := refresh.Read{Tokens: refresh.Full, Scope: refresh.Scope{Unread: true}}
	for _, ev := range []refresh.Event{refresh.TokenListsUpdated{}, refresh.AccountsOrChainsChanged{}} {
		p := refresh.NewPolicy(heldConfig())
		p.Handle(t0, refresh.Started{})

		changed := p.Handle(t0.Add(30*time.Second), ev)
		require.Empty(t, changed.Reads)
		require.Equal(t, t0.Add(40*time.Second), changed.WakeAt)
		require.Equal(t, []refresh.Read{unread}, p.Handle(t0.Add(40*time.Second), refresh.Tick{}).Reads, "%T", ev)
	}
}

func TestDetectionGapReadsEverythingOnThatChainAfterDebounce(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})

	gap := p.Handle(t0.Add(30*time.Second), refresh.DetectionGap{Chain: 10})
	require.Empty(t, gap.Reads)
	require.Equal(t, t0.Add(40*time.Second), gap.WakeAt)

	chain := refresh.Read{Tokens: refresh.Full, Scope: refresh.Scope{Pair: refresh.Pair{Chain: 10}}}
	require.Equal(t, []refresh.Read{chain}, p.Handle(t0.Add(40*time.Second), refresh.Tick{}).Reads)
}

func TestFailedFullRefreshOfAChainIsRetriedAtTheNextTick(t *testing.T) {
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})

	failed := p.Handle(t0.Add(30*time.Second), refresh.FullRefreshFailed{Chain: 10})
	require.Empty(t, failed.Reads)
	require.Equal(t, t0.Add(2*time.Minute), failed.WakeAt)

	retry := refresh.Read{Tokens: refresh.Full, Scope: refresh.Scope{Pair: refresh.Pair{Chain: 10}}}
	require.Equal(t, []refresh.Read{{Tokens: refresh.Held}, retry}, p.Handle(t0.Add(2*time.Minute), refresh.Tick{}).Reads)
	require.Equal(t, []refresh.Read{{Tokens: refresh.Held}}, p.Handle(t0.Add(4*time.Minute), refresh.Tick{}).Reads)
}

func TestDetectedTransfersOfAPairMakeOneTransferRefreshOfWhatMoved(t *testing.T) {
	usdc := common.HexToAddress("0x05dc")
	dai := common.HexToAddress("0xda1")
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})

	p.Handle(t0.Add(30*time.Second), refresh.TransferDetected{Pair: refresh.Pair{Account: alice, Chain: 1}, Contracts: []common.Address{usdc}})
	p.Handle(t0.Add(31*time.Second), refresh.TransferDetected{Pair: refresh.Pair{Account: bob, Chain: 1}, Contracts: []common.Address{usdc}})
	p.Handle(t0.Add(32*time.Second), refresh.TransferDetected{Pair: refresh.Pair{Account: alice, Chain: 1}, Contracts: []common.Address{dai, usdc}})
	last := p.Handle(t0.Add(33*time.Second), refresh.TransferDetected{Pair: refresh.Pair{Account: alice, Chain: 1}, Native: true})
	require.Equal(t, t0.Add(43*time.Second), last.WakeAt)

	require.Equal(t, []refresh.Read{
		{Tokens: refresh.Transfer, Scope: refresh.Scope{Pair: refresh.Pair{Account: alice, Chain: 1}}, Kinds: refresh.Native | refresh.ERC20, Contracts: []common.Address{usdc, dai}},
		{Tokens: refresh.Transfer, Scope: refresh.Scope{Pair: refresh.Pair{Account: bob, Chain: 1}}, Kinds: refresh.ERC20, Contracts: []common.Address{usdc}},
	}, p.Handle(t0.Add(43*time.Second), refresh.Tick{}).Reads)
}

func TestHeldScheduleReadsWhatAnEndedTransactionMoved(t *testing.T) {
	usdc := common.HexToAddress("0x05dc")
	p := refresh.NewPolicy(heldConfig())
	p.Handle(t0, refresh.Started{})

	confirmed := p.Handle(t0.Add(30*time.Second), refresh.TransactionEnded{
		Pairs:     []refresh.Pair{{Account: alice, Chain: 1}},
		Contracts: []common.Address{usdc},
		Confirmed: true,
	})
	require.Equal(t, []refresh.Read{{
		Tokens:    refresh.Transfer,
		Scope:     refresh.Scope{Pair: refresh.Pair{Account: alice, Chain: 1}},
		Kinds:     refresh.Native | refresh.ERC20,
		Contracts: []common.Address{usdc},
	}}, confirmed.Reads)
}
