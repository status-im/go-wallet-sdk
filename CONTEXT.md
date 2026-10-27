# Glossary

Terms with a precise meaning in go-wallet-sdk and in status-go, which builds on
it. Use them exactly; if a new term conflicts with one below, resolve the
conflict here first.

## Wallet balances

Every balance read is a **refresh**, named after the tokens it reads: the held
set (held refresh), every listed token (full refresh), or the tokens a
transfer moved (transfer refresh).

**Active chain** — a chain the user turned on for the whole app; balances are
read on active chains only. Not the wallet's chain filter, which hides chains
from view without changing what is read.
_Avoid_: enabled chain — in status-app "enabled" is that filter.

**Held set** — per account and chain, the tokens worth re-reading often: the
native coin, always; tokens with a non-zero last balance; tokens seen in a
detected transfer; the user's custom tokens; tokens of pending transactions;
and the collectibles the account is known to own. A token other than the
native coin leaves the set once it reads zero.

**Held refresh** — reading every balance in the held set. Frequent and cheap:
its cost follows what the user holds, not the size of the token lists.
Rebasing tokens the user holds are kept current by it.

**Full refresh** — reading every token of the token lists within a scope:
everywhere (every account on every active chain), one chain (after a detection
gap or a failed read), or what has never been read (a new account, a newly
activated chain, tokens a token list update added). Rare: the safety net that catches
whatever the other refreshes miss.
_Avoid_: sweep — in wallets it means moving all funds off a key.

**Transfer refresh** — reading only what a transfer moved: the tokens of a
transfer that transfer detection found, or of a pending transaction that
ended, for the accounts and chains involved.

**Transfer detection** — periodic search of new blocks for token `Transfer`
logs that involve one of the user's addresses. A signal that a balance may have
moved, not a source of balances or of history.

**Detection gap** — a block range transfer detection did not cover: a failed
query, a range over the provider's limit, or a pause too long to catch up. A
gap is never read back; it is closed by a full refresh of that chain.

**Native transfer log** — the log a chain with EIP-7708 emits for a transfer of
the native coin, shaped like a token `Transfer` but coming from a system
address rather than a token contract. A signal for the native balance, never a
token.

**Activity** — the user's transaction history. Not a balance signal; distinct
from transfer detection even though both look at transfers.
_Avoid_ for balance triggers: "refresh on activity" reads as refreshing on new
history.
