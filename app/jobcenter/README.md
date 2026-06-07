# jobcenter

`jobcenter` 是重构后 `go-zero` 布局中的异步工作中心。

当前已迁移的范围：

- 通过 `core/service` 的标准 go-zero 后台进程引导
- 使用 `time.Ticker` 构建的基于 goroutine 的间隔任务服务
- 基于 `kafka-go` 构建的共享 Kafka 消费者封装
- 来自 `ucenter-rpc` 的 `withdraw` 主题消费
- OKX 驱动的 USD/CNY 汇率同步到 Redis
- OKX 驱动的 K 线同步到 MongoDB
- 1 分钟最新价格缓存刷新及可选的 `kline_1m` Kafka 发布
- 通过 Bitcoin Core JSON-RPC 的 BTC 提现执行
- 在 `ucenter` 数据库中直接进行 `withdraw_record` 成功最终化
- 基于 Redis 的 txid 恢复缓存，用于减少链上成功但 MySQL 更新完成前的重复广播风险
- 不可重试的毒消息死信支持
- 与 `ucenter-rpc` 的 Bitcoin Core 钱包地址/签名对齐
- 共享 `mscoin` Bitcoin 钱包的 Docker Compose 引导支持

计划的下一步职责：

- 更多订单/交易/会员事件的 Kafka 消费者
- K 线聚合
- 链扫描和补偿任务