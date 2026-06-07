# Architecture

The refactored project uses a standard `go-zero` service split:

- `app/*/api` for HTTP adapters
- `app/*/rpc` for domain execution
- `pkg/` for infrastructure and cross-service helpers
- `idl/` for API and RPC source-of-truth contracts

The current migrated slices are:

- `market-api -> market-rpc -> MySQL/MongoDB`
- `exchange-api -> exchange-rpc -> MySQL`
- `ucenter-api -> ucenter-rpc -> MySQL + Redis + market-rpc`
- `jobcenter -> Kafka + Redis + ucenter MySQL + market-rpc + ucenter-rpc`

The currently migrated `ucenter` flows now cover:

- register, login, token check, member profile/security reads
- wallet list and wallet-by-symbol reads
- BTC wallet reset-address generation
- transaction history reads
- withdraw supported-coin aggregation
- withdraw verification code send
- withdraw apply write flow with transactional balance freeze and Kafka publish
- withdraw history reads

The first migrated `jobcenter` slice now covers the downstream withdraw Kafka
consumer:

- consume `withdraw` events emitted by `ucenter-rpc`
- re-load the persisted withdraw record from MySQL for idempotent status checks
- query coin metadata from `market-rpc`
- query member wallet source address from `ucenter-rpc`
- execute BTC withdraw broadcast through Bitcoin Core JSON-RPC
- mark the withdraw record successful after a txid is obtained
- cache the txid in Redis before final MySQL update so retries can finish the
  record without rebroadcasting when the chain step already succeeded
- share the same Bitcoin Core wallet authority with `ucenter-rpc`, so reset
  addresses and withdraw signing are consistent at runtime

The current `jobcenter` task side now also covers:

- goroutine-based interval scheduling with `time.Ticker`
- USD/CNY exchange-rate synchronization from OKX into Redis
- K-line synchronization from OKX into MongoDB for visible market pairs
- latest 1m pair price refresh into Redis and optional `kline_1m` Kafka publish

The standard internal layering for each RPC service is:

- `internal/repository`: all `sqlx` data access and persistence details
- `internal/domain/service`: orchestration rules and reusable business behavior
- `internal/logic`: transport-facing use cases called by the RPC server
- `internal/server`: protobuf server adapters registered into go-zero `zrpc`
- `internal/svc`: dependency graph composition and client wiring

The standard internal layering for each API service is:

- `internal/handler`: HTTP parsing, route binding, and response adaptation
- `internal/logic`: use-case orchestration for each HTTP endpoint
- `internal/middleware`: JWT and request pipeline middleware
- `internal/svc`: RPC client wiring and shared middleware construction

The standard internal layering for the current `jobcenter` worker is:

- `internal/consumer`: Kafka message adaptation and error classification
- `internal/domain/service`: async orchestration, idempotency, and chain rules
- `internal/repository`: narrow SQL status-finalization access
- `internal/task`: goroutine-based periodic task lifecycle and scheduling
- `internal/svc`: dependency graph composition and RPC client wiring

Container delivery now follows one Dockerfile per microservice plus
`docker-compose.yml`. The local stack also includes a one-shot Bitcoin wallet
initializer so the shared `mscoin` wallet exists before `ucenter-rpc` allocates
BTC addresses and before `jobcenter` signs BTC withdraw transactions.
