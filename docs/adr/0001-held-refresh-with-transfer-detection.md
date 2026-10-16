---
status: proposed
---

# Balances kept current by held refresh, with transfer detection as a signal

Wallet balances are kept current by three refreshes. A held refresh reads the
held set every 2 min while balances are on screen and every 15 min otherwise.
A transfer refresh reads what transfer detection or a pending transaction
points at, 10 s after the event. A full refresh reads every listed token:
everywhere on start, on resume after more than an hour, once an hour while
balances are on screen and on the user's request; for one chain after a detection gap or a failed read;
and for what has never been read, such as a new account or the tokens a token
list update added.

Four rules keep the full refresh rare, each backed by a measurement on a
3-account, 5-chain profile, where one full refresh costs about 2.45 MB:

- While balances are hidden, the hourly full refresh waits for them to come
  back on screen and runs then if the last one is over an hour old. Hourly,
  it alone would leave hidden traffic only 3-4x lower than today's 13 MB/h.
- A pause shorter than an hour is caught up by transfer detection reading the
  paused blocks. A full refresh on every resume would be the largest cost on
  mobile.
- A token list update reads only the tokens it added. The lists refresh about
  every half hour, so a full refresh each time would cost more than everything
  else together.
- A refresh the user asks for reads the held set, and everything only if the
  last full refresh is over 10 min old. The reload button allows a press a
  minute, so a full refresh on every press could cost up to 147 MB/h.

Transfer detection only adds tokens to the held set and triggers a transfer
refresh; it is never the sole source of a balance. Before this, every tick
read every token of every list for every account, which made the per-tick cost
follow the size of the token lists (about 20k tokens per account) instead of
what the user holds.

## Considered Options

- **Provider balance API (Alchemy `getTokenBalances`)** — fewer bytes, but ties
  balances to one provider; dropping the provider would leave no balances.
  Multicall3 works against any RPC node.
- **Purely event-driven balances from transfer logs** — native coin moves
  without logs (internal transfers, until EIP-7708 ships on every chain we
  support), rebasing tokens change without a `Transfer`, and `eth_getLogs`
  fails or is range-limited on some providers. Every such miss would leave a
  balance stale with no error.
- **WebSocket log subscriptions** — the RPC layer has no WebSocket transport,
  and mobile pauses the wallet service in background, so a subscription still
  needs the same gap handling as polling.

## Consequences

- Correctness rests on one rule: every block range is either covered by
  transfer detection or followed by a full refresh of that chain.
- Reads of one chain run one after another, so a later result never carries
  an older block; balance storage merges a read that lands in the same block
  as the stored one, since partial reads of a chain often do.
- Activity is triggered by causes (detected transfers, native balance changes,
  pending transactions), not by every balance change, so rebasing tokens do not
  re-query the history provider.
- Transfer detection is built for EIP-7708 from the start: native transfer
  logs pass the token-list filter and refresh the native balance, so chains
  can later read the native coin less often once 7708 is live on them.
