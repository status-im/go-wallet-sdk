# Balance Refresh Policy

Decides when wallet balances are read and which ones. The policy is pure: it takes events and the time they happen at, and answers with the reads to make now and when to be called again. The caller owns every timer and every RPC call, so the rules are tested on a fake clock.

Terms (held set, held refresh, full refresh, transfer refresh, detection gap) are defined in [CONTEXT.md](../../../CONTEXT.md); the design and its measurements are in [ADR 0001](../../../docs/adr/0001-held-refresh-with-transfer-detection.md).

## Use it when

- You keep balances of many accounts on many chains current, and reading every listed token every period costs too much.
- You want the refresh rules (periods, debounce, what to read after which event) in one place, apart from the code that reads balances.

## Key entrypoints

- `refresh.NewPolicy(config) *Policy`
- `(*Policy).Handle(now, event) Decision`
- `refresh.DefaultConfig()`: 2 min on screen, 15 min hidden, full refresh hourly and at most every 10 min on request, 10 s debounce

## Schedules

| | `FullEveryTick` (zero value) | `HeldEveryTick` |
|---|---|---|
| Every period | full refresh | held refresh |
| Start | full refresh | full refresh |
| Resume | full refresh | held refresh after a pause under an hour, full refresh after a longer one |
| Hourly | — | full refresh while on screen; while hidden, as soon as balances are back on screen |
| Token lists updated | full refresh | full refresh of what was never read |
| Accounts or chains changed | full refresh of balances older than the period | full refresh of what was never read |
| Refresh requested | full refresh | held refresh; full refresh if the last one is over 10 min old |
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

Each `Read` names which tokens to read and where:

- `Tokens`: `Full` (every token of the lists), `Held` (the held set) or `Transfer` (`Contracts`, plus the native coin when `Kinds` has `Native`).
- `Scope`: every account on every chain when zero; narrowed by `Pair` (a zero field means every account or every chain), `Unread` (only what was never read) or `OlderThan`.
- `Kinds`: native, ERC20, ERC721, ERC1155; zero means every kind.

The policy does not know which tokens are held or unread: the caller resolves that from what it has read before.
