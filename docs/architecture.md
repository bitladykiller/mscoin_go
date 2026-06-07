# 架构

重构后的项目采用标准的 `go-zero` 服务拆分：

- `app/*/api` 用于 HTTP 适配器
- `app/*/rpc` 用于领域执行
- `pkg/` 用于基础设施和跨服务辅助工具
- `idl/` 用于 API 和 RPC 的真实来源契约

当前已迁移的切片：

- `market-api -> market-rpc -> MySQL/MongoDB`
- `exchange-api -> exchange-rpc -> MySQL`
- `ucenter-api -> ucenter-rpc -> MySQL + Redis + market-rpc`
- `jobcenter -> Kafka + Redis + ucenter MySQL + market-rpc + ucenter-rpc`

当前已迁移的 `ucenter` 流程现已覆盖：

- 注册、登录、令牌检查、会员资料/安全设置读取
- 钱包列表和按符号查询钱包读取
- BTC 钱包重置地址生成
- 交易历史读取
- 提现支持币种聚合
- 提现验证码发送
- 提现申请写入流程，包含事务性余额冻结和 Kafka 发布
- 提现历史读取

第一个已迁移的 `jobcenter` 切片现已覆盖下游提现 Kafka 消费者：

- 消费 `ucenter-rpc` 发出的 `withdraw` 事件
- 从 MySQL 重新加载持久化的提现记录以进行幂等状态检查
- 从 `market-rpc` 查询币种元数据
- 从 `ucenter-rpc` 查询会员钱包源地址
- 通过 Bitcoin Core JSON-RPC 执行 BTC 提现广播
- 获取 txid 后标记提现记录成功
- 在最终 MySQL 更新前将 txid 缓存到 Redis，以便链上步骤已成功时重试可以在不重新广播的情况下完成记录
- 与 `ucenter-rpc` 共享相同的 Bitcoin Core 钱包授权，因此重置地址和提现签名在运行时保持一致

当前 `jobcenter` 任务端也已覆盖：

- 使用 `time.Ticker` 的基于 goroutine 的间隔调度
- 从 OKX 同步 USD/CNY 汇率到 Redis
- 从 OKX 同步 K 线到 MongoDB（针对可见的市场交易对）
- 刷新最新 1 分钟交易对价格到 Redis，可选发布到 `kline_1m` Kafka 主题

每个 RPC 服务的标准内部分层：

- `internal/repository`：所有 `sqlx` 数据访问和持久化细节
- `internal/domain/service`：编排规则和可复用的业务行为
- `internal/logic`：面向传输的用例，由 RPC 服务器调用
- `internal/server`：注册到 go-zero `zrpc` 的 protobuf 服务器适配器
- `internal/svc`：依赖图组合和客户端连接

每个 API 服务的标准内部分层：

- `internal/handler`：HTTP 解析、路由绑定和响应适配
- `internal/logic`：每个 HTTP 端点的用例编排
- `internal/middleware`：JWT 和请求管道中间件
- `internal/svc`：RPC 客户端连接和共享中间件构建

当前 `jobcenter` 工作器的标准内部分层：

- `internal/consumer`：Kafka 消息适配和错误分类
- `internal/domain/service`：异步编排、幂等性和链上规则
- `internal/repository`：窄范围 SQL 状态最终化访问
- `internal/task`：基于 goroutine 的周期性任务生命周期和调度
- `internal/svc`：依赖图组合和 RPC 客户端连接

容器交付现在遵循每个微服务一个 Dockerfile 加上 `docker-compose.yml` 的方式。本地栈还包括一个一次性 Bitcoin 钱包初始化器，确保在 `ucenter-rpc` 分配 BTC 地址和 `jobcenter` 签署 BTC 提现交易之前，共享的 `mscoin` 钱包已存在。