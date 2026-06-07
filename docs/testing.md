# Testing

Recommended commands:

```bash
make tidy
make fmt
make vet
make test
make build
make docker-up
```

The current migration baseline should at least pass for:

- `./app/market/api`
- `./app/market/rpc`
- `./app/exchange/api`
- `./app/exchange/rpc`
- `./app/ucenter/api`
- `./app/ucenter/rpc`
- `./app/jobcenter`

Recent `ucenter` additions with direct unit coverage:

- withdraw supported-coin API aggregation
- withdraw verification code API mapping
- withdraw apply domain orchestration, including verification-code validation,
  balance freeze, record persistence, and Kafka publish failure handling
- withdraw history API mapping
- asset reset-address API mapping
- asset reset-address RPC logic
- asset get-address RPC logic
- withdraw record and address-book model mapping
- withdraw record apply-model construction
- MySQL transaction-manager guard paths
- Kafka producer parameter validation
- BTC address generation helper

Recent `jobcenter` additions with direct unit coverage:

- Kafka consumer parameter validation, retry policy, commit behavior, and
  dead-letter handling
- BTC withdraw sender validation and transaction-flow orchestration
- OKX market-data client validation, exchange-rate parsing, and candle parsing
- goroutine-based task scheduler immediate-run behavior
- exchange-rate synchronization domain orchestration
- K-line synchronization domain orchestration
- withdraw async domain orchestration, including:
  verification of committed record visibility
  Redis txid recovery path
  BTC-only execution routing
  final MySQL success update path
- withdraw consumer error classification

Recent `market-rpc` additions with direct unit coverage:

- Redis-backed fiat-rate lookup with graceful fallback behavior

Recent BTC wallet alignment additions with direct unit coverage:

- `ucenter-rpc` reset-address logic now verifies Bitcoin Core-backed address
  allocation instead of local private-key generation
- `pkg/btcx` now covers both withdraw-sender validation and address-allocator
  bootstrap validation paths

Container verification notes:

- `docker-compose.yml` expects Bitcoin Core RPC under `wallet/mscoin`
- `deploy/bitcoin/init-wallet.sh` creates or loads that wallet before
  `ucenter-rpc` and `jobcenter` start
- `deploy/mysql/init/*.sql` create the minimal databases, schema, and seed rows
  needed for the migrated services to answer queries and run K-line sync
- each microservice image is built from its own service-local Dockerfile under
  `app/*/{api,rpc,}/Dockerfile`
