# Migration Plan

Completed:

- root repository skeleton
- centralized contracts under `idl/`
- `market-api` and `market-rpc` migration sample
- `exchange-api` and `exchange-rpc` migration sample
- `ucenter-api` and `ucenter-rpc` minimum viable migration

Current `ucenter` scope in the migrated repository:

- send register code
- register by phone
- login
- check login token
- member by id
- security setting
- wallet list
- wallet by symbol
- wallet reset-address
- transaction history
- withdraw support coin info
- send withdraw code
- withdraw apply
- withdraw record

Completed async follow-up:

- `jobcenter` standard go-zero worker bootstrap
- shared Kafka consumer abstraction
- withdraw event consumption and BTC withdraw execution
- goroutine-based interval task scheduler in `jobcenter`
- OKX-driven USD/CNY rate synchronization
- OKX-driven K-line synchronization
- Redis-backed `market-rpc` fiat-rate lookup
- Bitcoin Core wallet-backed BTC address allocation in `ucenter-rpc`
- service-local Dockerfiles, Compose stack, MySQL init scripts, and Bitcoin
  wallet bootstrap

Next:

1. add remaining Kafka consumer/compensation refactors
2. expand integration tests around cross-service chains
3. continue filling the remaining business modules on top of the standard skeleton
