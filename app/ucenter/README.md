# ucenter

This directory now contains the migrated `go-zero` implementation for the
minimum viable `ucenter` slice.

Implemented submodules:

- `api/`
- `rpc/`

Implemented capabilities:

- register SMS-style verification code caching
- phone registration with captcha verification
- user login with captcha passthrough
- login-token verification
- member lookup by user id
- security setting query with legacy boolean-string flags
- wallet list query
- wallet lookup by coin symbol
- wallet address reset for BTC through Bitcoin Core wallet-managed address allocation
- transaction history query with legacy pagination envelope
- withdraw supported coin aggregation
- withdraw verification code caching
- withdraw apply with transactional balance freeze and Kafka event publish
- withdraw history query

Technical rules used in this submodule:

- API layer only handles HTTP parsing, JWT middleware, and RPC adaptation
- RPC layer owns business orchestration and repository access
- MySQL persistence is implemented with `sqlx`
- register and withdraw verification codes are cached through `go-redis`
- transaction history reads directly from MySQL through repository filtering and paging
- cross-service coin metadata lookup is delegated to `market-rpc`
- BTC address allocation is isolated in `pkg/btcx` and delegated to Bitcoin Core
  so `ucenter-rpc` and `jobcenter` use the same node wallet

Current async follow-up:

- the `ucenter` side now persists withdraw applications and publishes Kafka
  events in the standard refactor layout
- the downstream Kafka consumer / `jobcenter` chain execution side is already
  migrated, including BTC withdraw finalization
