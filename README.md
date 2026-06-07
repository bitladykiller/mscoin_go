# mscoin_go

`mscoin_go` 是原始 MSCoin 系统的新标准 `go-zero` 重构目标。

当前状态：

- 仓库骨架已创建
- 契约集中在 `idl/` 目录下
- `market-api + market-rpc` 已迁移并可构建
- `exchange-api + exchange-rpc` 已迁移并可构建
- `ucenter-api + ucenter-rpc` 最小可行切片已迁移并可构建
- `jobcenter` 异步工作器和基于 goroutine 的周期性任务切片已迁移
- 每个微服务都有自己的 Dockerfile，完整本地栈通过 Compose 容器化

## 目录结构

```text
mscoin_go/
  app/      # 应用服务目录
  docs/     # 文档目录
  idl/      # 接口定义语言（API和RPC契约）
  pkg/      # 公共基础设施包
```

当前已迁移的业务切片：

- `market`：市场符号查询和币种信息查询
- `exchange`：订单列表和订单详情查询
- `ucenter`：发送注册验证码、手机号注册、登录、令牌验证、按ID查询会员、安全设置、钱包列表、按符号查询钱包、钱包重置地址、交易历史、提现支持币种信息、提现验证码发送、提现申请、提现记录查询
- `jobcenter`：提现 Kafka 消费、BTC 提现执行、USD/CNY 汇率同步、OKX K线同步、最新1分钟价格缓存及可选的 Kafka 发布

本次重构采用的基础设施选型：

- MySQL 访问统一使用 `sqlx`
- Redis 访问统一使用 `go-redis`
- Kafka 作为标准的异步消息队列
- MongoDB 通过专用辅助包封装用于文档工作负载
- `jobcenter` 定时任务使用原生 goroutine 配合 `time.Ticker` 实现，不依赖外部调度框架

## 命令

```bash
make tidy    # 整理依赖
make fmt     # 格式化代码
make vet     # 静态检查
make test    # 运行测试
make build   # 构建项目
make docker-up    # 启动 Docker 容器
make docker-down  # 停止 Docker 容器
make docker-logs  # 查看 Docker 日志
```

## Docker

仓库现在在每个微服务目录下提供服务级别的 Dockerfile，以及完整的 [docker-compose.yml](/Volumes/移动卷宗/学习/go/mscoin_go/docker-compose.yml)。

Compose 启动的服务：

- 业务服务：`market-api`、`market-rpc`、`exchange-api`、`exchange-rpc`、`ucenter-api`、`ucenter-rpc`、`jobcenter`
- 基础设施：`mysql`、`redis`、`etcd`、`mongo`、`zookeeper`、`kafka`、`kafdrop`、`bitcoin`
- 一次性引导：`bitcoin-wallet-init`

`bitcoin-wallet-init` 存在的原因：

- `ucenter-rpc` 现在通过 Bitcoin Core 的 `getnewaddress` 分配 BTC 地址
- `jobcenter` 使用 `signrawtransactionwithwallet` 签署 BTC 提现交易
- 两个操作必须使用同一个钱包，因此 Compose 在 `ucenter-rpc` 和 `jobcenter` 启动前创建/加载 `wallet/mscoin`

MySQL 初始化脚本位于 [deploy/mysql/init](/Volumes/移动卷宗/学习/go/mscoin_go/deploy/mysql/init)，创建 `market`、`ucenter` 和 `exchange` 数据库以及迁移服务所需的最小模式和种子数据。

## 备注

本次重构有意将业务代码推向：

- 契约驱动的 API 和 RPC 层
- `repository + domain/service + logic` 分层架构
- MySQL 访问使用 `sqlx`
- Redis 访问使用 `go-redis`
- Kafka 和 MongoDB 通过专用基础设施封装

当前 `ucenter` 写端状态：

- `/uc/withdraw/apply/code` 已迁移，使用显式 MySQL 事务编排进行余额冻结，并通过 Kafka 事件分发进行下游异步处理
- BTC 重置地址现在由 Bitcoin Core 钱包管理的地址分配支持，而非本地私钥生成，因此下游提现签名和上游地址分配使用相同的钱包授权
- 第一个异步后续处理目标已在 `jobcenter` 中迁移，包括 Kafka 消费者和基于 goroutine 的周期性市场数据任务