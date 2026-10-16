// Package refresh decides when wallet balances are read and which ones. It is
// pure: it takes events and the time they happen at and answers with the reads
// to make now and when to be called again, so the caller owns every timer and
// every RPC call. The terms follow CONTEXT.md at the repository root; the
// design is docs/adr/0001-held-refresh-with-transfer-detection.md.
package refresh

import (
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Schedule is what the periodic refresh reads.
type Schedule int

const (
	// FullEveryTick reads every listed token every period.
	FullEveryTick Schedule = iota
	// HeldEveryTick reads the held set every period and every listed token
	// rarely.
	HeldEveryTick
)

// Config holds the schedule and the periods of the policy.
type Config struct {
	Schedule Schedule
	// OnScreenPeriod is the refresh period while balances are on screen.
	OnScreenPeriod time.Duration
	// HiddenPeriod is the refresh period while they are not.
	HiddenPeriod time.Duration
	// FullPeriod is how often HeldEveryTick reads everything while balances
	// are on screen.
	FullPeriod time.Duration
	// Debounce is how long reads wait for more events, so that a burst of
	// events makes one read; each event restarts the wait.
	Debounce time.Duration
}

func DefaultConfig() Config {
	return Config{
		OnScreenPeriod: 2 * time.Minute,
		HiddenPeriod:   15 * time.Minute,
		FullPeriod:     time.Hour,
		Debounce:       10 * time.Second,
	}
}

// Tokens names which tokens a read covers.
type Tokens int

const (
	// Full is every token of the token lists.
	Full Tokens = iota + 1
	// Held is the held set.
	Held
	// Transfer is what a transfer moved: Read.Contracts, and the native coin
	// when Read.Kinds has Native.
	Transfer
)

// Kinds is a set of balance kinds; the zero Kinds is every kind.
type Kinds uint8

const (
	Native Kinds = 1 << iota
	ERC20
	ERC721
	ERC1155
)

// Pair is an account on a chain; a zero field stands for every account or
// every chain.
type Pair struct {
	Account common.Address
	Chain   uint64
}

// Read is one refresh to make.
type Read struct {
	Tokens Tokens
	Scope  Scope
	Kinds  Kinds
	// Contracts are the tokens a Transfer read covers.
	Contracts []common.Address
}

// Scope narrows a read; the zero Scope is every account on every chain.
type Scope struct {
	Pair Pair
	// Unread keeps only what has never been read: new accounts, newly
	// enabled chains, tokens new to the lists.
	Unread bool
	// OlderThan keeps only the balances last read longer ago than this.
	OlderThan time.Duration
}

// Decision is the answer to an event: the reads to make now, and when to call
// Handle with Tick next.
type Decision struct {
	Reads  []Read
	WakeAt time.Time
}

// Event is something the policy reacts to.
type Event interface{ event() }

// Started is the start of balance refreshing, including a resume.
type Started struct{}

func (Started) event() {}

// Stopped is the end of balance refreshing, such as a pause in background.
type Stopped struct{}

func (Stopped) event() {}

// Tick is the call the caller makes at the WakeAt of a decision. A tick may
// come early; it then changes nothing.
type Tick struct{}

func (Tick) event() {}

// OnScreenChanged tells whether balances are in front of the user.
type OnScreenChanged struct{ OnScreen bool }

func (OnScreenChanged) event() {}

// TokenListsUpdated tells that the token lists have been refreshed.
type TokenListsUpdated struct{}

func (TokenListsUpdated) event() {}

// AccountsOrChainsChanged tells that accounts were added or the active chains
// changed.
type AccountsOrChainsChanged struct{}

func (AccountsOrChainsChanged) event() {}

// TransactionEnded tells that a pending transaction of the user reached its
// final state.
type TransactionEnded struct {
	// Pairs are the accounts and chains the transaction involved.
	Pairs []Pair
	// Kinds are the balance kinds it may have moved; FullEveryTick reads them.
	Kinds Kinds
	// Contracts are the tokens it moved, when known; HeldEveryTick reads them
	// and the native coin, which paid for gas.
	Contracts []common.Address
	// Confirmed is true when it succeeded: its balances are read at once.
	Confirmed bool
}

func (TransactionEnded) event() {}

// DetectionGap tells that transfer detection left a block range of Chain
// uncovered.
type DetectionGap struct{ Chain uint64 }

func (DetectionGap) event() {}

// TransferDetected tells that transfer detection found a transfer of Pair's
// account on Pair's chain.
type TransferDetected struct {
	Pair Pair
	// Contracts are the tokens that moved.
	Contracts []common.Address
	// Native is true for a native transfer log.
	Native bool
}

func (TransferDetected) event() {}

// FullRefreshFailed tells that a full refresh of Chain did not complete; it is
// retried at the next periodic refresh until it does.
type FullRefreshFailed struct{ Chain uint64 }

func (FullRefreshFailed) event() {}

// Policy is not safe for concurrent use: its caller owns it.
type Policy struct {
	config Config
	hidden bool
	// lastPeriodic is when the last periodic refresh was made, lastFull the
	// last one that read everything.
	lastPeriodic time.Time
	lastFull     time.Time
	// stopped is when refreshing stopped; zero while it runs.
	stopped time.Time
	// retries are the chains whose full refresh failed.
	retries []uint64
	// pending are the reads waiting for the debounce to end at debounced.
	pending   []Read
	debounced time.Time
}

func NewPolicy(config Config) *Policy {
	return &Policy{config: config}
}

// Handle takes ev, which happened at now, and says what to read now and when
// to be called with Tick next.
func (p *Policy) Handle(now time.Time, ev Event) Decision {
	var reads []Read
	switch ev := ev.(type) {
	case Started:
		reads = p.readStarted(now)
	case Stopped:
		p.stopped = now
		p.pending = nil
		return Decision{}
	case Tick:
		if reads = p.readPeriodicIfDue(now); reads == nil {
			reads = p.readDebouncedIfDue(now)
		}
	case OnScreenChanged:
		if p.hidden != ev.OnScreen {
			break
		}
		p.hidden = !ev.OnScreen
		reads = p.readPeriodicIfDue(now)
	case TokenListsUpdated:
		if p.config.Schedule == HeldEveryTick {
			p.debounce(now, Read{Tokens: Full, Scope: Scope{Unread: true}})
		} else {
			p.debounce(now, Read{Tokens: Full})
		}
	case AccountsOrChainsChanged:
		if p.config.Schedule == HeldEveryTick {
			p.debounce(now, Read{Tokens: Full, Scope: Scope{Unread: true}})
		} else {
			p.debounce(now, Read{Tokens: Full, Scope: Scope{OlderThan: p.period()}})
		}
	case TransferDetected:
		read := Read{Tokens: Transfer, Scope: Scope{Pair: ev.Pair}, Contracts: slices.Clone(ev.Contracts)}
		if len(ev.Contracts) > 0 {
			read.Kinds |= ERC20
		}
		if ev.Native {
			read.Kinds |= Native
		}
		p.debounce(now, read)
	case DetectionGap:
		p.debounce(now, Read{Tokens: Full, Scope: Scope{Pair: Pair{Chain: ev.Chain}}})
	case FullRefreshFailed:
		if !slices.Contains(p.retries, ev.Chain) {
			p.retries = append(p.retries, ev.Chain)
		}
	case TransactionEnded:
		for _, pair := range ev.Pairs {
			read := Read{Tokens: Full, Scope: Scope{Pair: pair}, Kinds: ev.Kinds}
			if p.config.Schedule == HeldEveryTick {
				read = Read{Tokens: Transfer, Scope: Scope{Pair: pair}, Kinds: Native, Contracts: slices.Clone(ev.Contracts)}
				if len(ev.Contracts) > 0 {
					read.Kinds |= ERC20
				}
			}
			if ev.Confirmed {
				reads = append(reads, read)
			} else {
				p.debounce(now, read)
			}
		}
	}
	return Decision{Reads: reads, WakeAt: p.wakeAt()}
}

func (p *Policy) wakeAt() time.Time {
	next := p.nextPeriodic()
	if len(p.pending) > 0 && p.debounced.Before(next) {
		return p.debounced
	}
	return next
}

func (p *Policy) debounce(now time.Time, read Read) {
	switch {
	case slices.ContainsFunc(p.pending, readsEverything):
	case readsEverything(read):
		p.pending = []Read{read}
	default:
		p.pending = merge(p.pending, read)
	}
	p.debounced = now.Add(p.config.Debounce)
}

func readsEverything(r Read) bool {
	return r.Tokens == Full && r.Scope == Scope{} && r.Kinds == 0
}

// merge adds read to reads, folding it into a read of the same tokens and
// scope.
func merge(reads []Read, read Read) []Read {
	for i := range reads {
		r := &reads[i]
		if r.Tokens != read.Tokens || r.Scope != read.Scope {
			continue
		}
		if read.Tokens == Transfer {
			r.Kinds |= read.Kinds
			for _, c := range read.Contracts {
				if !slices.Contains(r.Contracts, c) {
					r.Contracts = append(r.Contracts, c)
				}
			}
			return reads
		}
		if r.Kinds == read.Kinds {
			return reads
		}
	}
	return append(reads, read)
}

func (p *Policy) readDebouncedIfDue(now time.Time) []Read {
	if len(p.pending) == 0 || now.Before(p.debounced) {
		return nil
	}
	reads := p.pending
	p.pending = nil
	return reads
}

// readStarted reads everything, except after a pause shorter than FullPeriod
// under HeldEveryTick: transfer detection catches up on the paused blocks.
func (p *Policy) readStarted(now time.Time) []Read {
	shortPause := !p.stopped.IsZero() && now.Sub(p.stopped) < p.config.FullPeriod
	p.stopped = time.Time{}
	if p.config.Schedule == HeldEveryTick && shortPause {
		return p.readPeriodic(now, Held)
	}
	return p.readPeriodic(now, Full)
}

func (p *Policy) period() time.Duration {
	if p.hidden {
		return p.config.HiddenPeriod
	}
	return p.config.OnScreenPeriod
}

func (p *Policy) nextPeriodic() time.Time {
	return p.lastPeriodic.Add(p.period())
}

func (p *Policy) readPeriodicIfDue(now time.Time) []Read {
	if p.config.Schedule == HeldEveryTick {
		// The hourly full refresh waits while balances are hidden, and runs as
		// soon as they are back on screen.
		if !p.hidden && !now.Before(p.lastFull.Add(p.config.FullPeriod)) {
			return p.readPeriodic(now, Full)
		}
		if now.Before(p.nextPeriodic()) {
			return nil
		}
		return p.readPeriodic(now, Held)
	}
	if now.Before(p.nextPeriodic()) {
		return nil
	}
	return p.readPeriodic(now, Full)
}

func (p *Policy) readPeriodic(now time.Time, tokens Tokens) []Read {
	p.lastPeriodic = now
	reads := []Read{{Tokens: tokens}}
	if tokens == Full {
		p.lastFull = now
		// Reading everything covers whatever was waiting.
		p.pending = nil
	} else {
		for _, chain := range p.retries {
			reads = append(reads, Read{Tokens: Full, Scope: Scope{Pair: Pair{Chain: chain}}})
		}
	}
	p.retries = nil
	return reads
}
