# jobcenter

`jobcenter` is the asynchronous worker center in the refactored `go-zero`
layout.

Current migrated scope:

- standard go-zero background process bootstrap through `core/service`
- goroutine-based interval task service built with `time.Ticker`
- shared Kafka consumer wrapper built on `kafka-go`
- `withdraw` topic consumption from `ucenter-rpc`
- OKX-driven USD/CNY rate synchronization into Redis
- OKX-driven K-line synchronization into MongoDB
- 1m latest-price cache refresh and optional `kline_1m` Kafka publish
- BTC withdraw execution through Bitcoin Core JSON-RPC
- direct `withdraw_record` success finalization in the `ucenter` database
- Redis-based txid recovery cache to reduce duplicate-broadcast risk after
  chain success but before MySQL update completion
- dead-letter support for non-retryable poison messages
- Bitcoin Core wallet-based address/signing alignment with `ucenter-rpc`
- Docker Compose bootstrap support for the shared `mscoin` Bitcoin wallet

Planned next responsibilities:

- more Kafka consumers for order/trade/member events
- K-line aggregation
- chain scanning and compensation tasks
