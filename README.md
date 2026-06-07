# mscoin_go

`mscoin_go` is the new standard `go-zero` refactor target for the original
MSCoin system.

Current status:

- repository skeleton is created
- contracts are centralized under `idl/`
- `market-api + market-rpc` are migrated and buildable
- `exchange-api + exchange-rpc` are migrated and buildable
- `ucenter-api + ucenter-rpc` minimum viable slice is migrated and buildable
- `jobcenter` async worker and goroutine-based periodic task slice are migrated
- each microservice has its own Dockerfile and the full local stack is
  containerized through Compose

## Layout

```text
mscoin_go/
  app/
  docs/
  idl/
  pkg/
```

Current migrated business slices:

- `market`: market symbol query and coin information query
- `exchange`: order list and order detail query
- `ucenter`: send register code, register by phone, login, token validation, member-by-id, security setting, wallet list, wallet-by-symbol, wallet reset-address, transaction history, withdraw support coin info, withdraw verification code send, withdraw apply, withdraw record query
- `jobcenter`: withdraw Kafka consume, BTC withdraw execution, USD/CNY sync, OKX K-line sync, latest 1m price cache and optional Kafka publish

Infrastructure choices enforced by this refactor:

- MySQL access is unified on `sqlx`
- Redis access is unified on `go-redis`
- Kafka is the standard asynchronous message queue
- MongoDB is wrapped behind dedicated helper packages for document workloads
- `jobcenter` scheduled work is implemented with native goroutines plus
  `time.Ticker`, not an external scheduler framework

## Commands

```bash
make tidy
make fmt
make vet
make test
make build
make docker-up
make docker-down
make docker-logs
```

## Docker

The repository now ships service-local Dockerfiles under each microservice
directory plus a full [docker-compose.yml](/Volumes/移动卷宗/学习/go/mscoin_go/docker-compose.yml).

What Compose starts:

- business services: `market-api`, `market-rpc`, `exchange-api`,
  `exchange-rpc`, `ucenter-api`, `ucenter-rpc`, `jobcenter`
- infrastructure: `mysql`, `redis`, `etcd`, `mongo`, `zookeeper`, `kafka`,
  `kafdrop`, `bitcoin`
- one-shot bootstrap: `bitcoin-wallet-init`

Why `bitcoin-wallet-init` exists:

- `ucenter-rpc` now allocates BTC addresses from Bitcoin Core through
  `getnewaddress`
- `jobcenter` signs BTC withdraw transactions with
  `signrawtransactionwithwallet`
- both operations must use the same wallet, so Compose creates/loads
  `wallet/mscoin` before `ucenter-rpc` and `jobcenter` start

The MySQL initialization scripts live under
[deploy/mysql/init](/Volumes/移动卷宗/学习/go/mscoin_go/deploy/mysql/init) and
create the `market`, `ucenter`, and `exchange` databases together with the
minimum schema and seed data required by the migrated services.

## Notes

This refactor intentionally moves business code toward:

- contract-driven API and RPC layers
- `repository + domain/service + logic` layering
- `sqlx` for MySQL access
- `go-redis` for Redis access
- Kafka and MongoDB through dedicated infrastructure wrappers

Current `ucenter` write-side status:

- `/uc/withdraw/apply/code` is now migrated with explicit MySQL transaction
  orchestration for balance freeze and Kafka event dispatch for downstream async
  processing
- BTC reset-address is now backed by Bitcoin Core wallet-managed address
  allocation instead of local private-key generation, so downstream withdraw
  signing and upstream address allocation use the same wallet authority
- the first async follow-up target is already migrated in `jobcenter`, including
  the Kafka consumer and goroutine-based periodic market data tasks
