# Balance Refresh Policy

Decides when wallet balances are read and which ones. The policy is pure: it takes events and the time they happen at, and answers with the reads to make now and when to be called again. The caller owns every timer and every RPC call, so the rules are tested on a fake clock.

Terms (held set, held refresh, full refresh, transfer refresh, detection gap) are defined in [CONTEXT.md](../../../CONTEXT.md); the design and its measurements are in [ADR 0001](../../../docs/adr/0001-held-refresh-with-transfer-detection.md).

## Use it when

- You keep balances of many accounts on many chains current, and reading every listed token every period costs too much.
- You want the refresh rules (periods, debounce, what to read after which event) in one place, apart from the code that reads balances.

## Key entrypoints

- `refresh.NewPolicy(config) *Policy`
- `(*Policy).Handle(now, event) Decision`
- `refresh.DefaultConfig()`: 2 min on screen, 15 min hidden, full refresh hourly and at most every 10 min on request, 10 s debounce capped at 1 min
- `refresh.ChangeSource`: where `TransferDetected` and `DetectionGap` come from, scanned with every periodic refresh
- `refresh.TransfersFromLogs(chain, accounts, logs, known)`: transfer logs to `TransferDetected` events

## Schedules

| | `FullEveryTick` (zero value) | `HeldEveryTick` |
|---|---|---|
| Every period | full refresh | held refresh |
| Start | full refresh | full refresh |
| Resume | full refresh | held refresh after a pause under an hour unless the hourly one is due on screen, full refresh otherwise |
| Hourly | — | full refresh while on screen; while hidden, as soon as balances are back on screen |
| Token lists updated | full refresh; at once on the first update after a cold start | full refresh of what was never read; at once on the first update after a cold start |
| Accounts or chains changed | full refresh of balances older than the period | full refresh of what was never read |
| Refresh requested | full refresh | held refresh at once; full refresh at once if the last one is over 10 min old |
| Reconnected | full refresh | held refresh at once; full refresh at once if the hourly one is due on screen |
| Transfer detected | transfer refresh of what moved | transfer refresh of what moved |
| Detection gap | full refresh of that chain | full refresh of that chain |
| Full refresh of a chain failed | covered by the next period's full refresh | retried at the next period |
| Transaction ended | full refresh of its accounts and chains | transfer refresh of what it moved and the native coin |

`FullEveryTick` keeps the schedule balances had before the held set existed.

## Driving it

```go
policy := refresh.NewPolicy(refresh.Config{
    Schedule:            refresh.HeldEveryTick,
    OnScreenPeriod:      2 * time.Minute,
    HiddenPeriod:        15 * time.Minute,
    FullPeriod:          time.Hour,
    RequestedFullPeriod: 10 * time.Minute,
    Debounce:            10 * time.Second,
    MaxDebounce:         time.Minute,
})

handle := func(ev refresh.Event) {
    d := policy.Handle(time.Now(), ev)
    for _, read := range d.Reads {
        go readBalances(read) // resolve read.Tokens and read.Scope against your storage
    }
    if !d.WakeAt.IsZero() {
        timer.Reset(time.Until(d.WakeAt)) // call handle(refresh.Tick{}) when it fires
    }
}

handle(refresh.Started{})
```

The policy is not safe for concurrent use: feed it from one goroutine. A `Tick` that comes early changes nothing, so the timer does not need to be exact.

Before `Started` and after `Stopped` every `Decision` is empty. The policy still hears events: it keeps whether balances are on screen and the reads that events ask for, and makes them at the next `Started`. A periodic refresh takes the reads waiting for the debounce, so they never wait longer than a period, and `MaxDebounce` bounds the wait when events keep coming.

Each `Read` names which tokens to read and where:

- `Tokens`: `Full` (every token of the lists), `Held` (the held set) or `Transfer` (`Contracts`, plus the native coin when `Kinds` has `Native`).
- `Scope`: every account on every chain when zero; narrowed by `Pair` (a zero field means every account or every chain), `Unread` (only what was never read) or `OlderThan`.
- `Kinds`: native, ERC20, ERC721, ERC1155; zero means every kind.

The policy does not know which tokens are held or unread: the caller resolves that from what it has read before.

## Change sources

A `ChangeSource` finds where balances may have moved between refreshes and tells the caller with `TransferDetected` and `DetectionGap`, which the caller passes to the policy. The caller calls `Scan` with every periodic refresh and when the user asks for one, so the source looks for changes as often as balances are read and right after a read has told it the latest block.

Transfer detection turns the `Transfer` logs of the user's accounts into events with `TransfersFromLogs`. It keeps the tokens the caller knows (its token lists and custom tokens), so spam tokens do not grow the held set. A native transfer log (EIP-7708, emitted by `NativeTransferLogAddress`) is a transfer of the native coin, never a token. A source tells `DetectionGap` for a block range it could not cover; the full refresh of that chain closes it.
