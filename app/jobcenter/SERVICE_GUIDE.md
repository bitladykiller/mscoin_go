# Jobcenter 服务详细指南

## 目录

1. [服务概述](#1-服务概述)
2. [目录结构](#2-目录结构)
3. [文件详细说明](#3-文件详细说明)
4. [Kafka 消费流程](#4-kafka-消费流程)
5. [定时任务机制](#5-定时任务机制)
6. [与其他服务的调用关系](#6-与其他服务的调用关系)
7. [配置说明](#7-配置说明)

---

## 1. 服务概述

### 1.1 服务定位

Jobcenter 是 mscoin 项目的**后台任务处理服务**，负责所有异步、定时、事件驱动的后台工作。它是一个"幕后工作者"，不直接面向用户，而是处理系统内部的后台任务。

### 1.2 核心职责

Jobcenter 服务承担以下核心职责：

#### 1.2.1 Kafka 消息消费

- **提现事件处理**：消费来自 `ucenter-rpc` 的提现申请事件，执行链上转账操作
- **消息可靠性保证**：通过死信队列（Dead Letter Queue）处理不可重试的毒消息
- **幂等性保证**：通过 Redis 缓存检查点防止重复处理

#### 1.2.2 定时任务调度

- **汇率同步**：定时从 OKX API 获取 USDT/CNY 实时汇率，缓存到 Redis
- **K线数据同步**：定时从 OKX API 同步多种周期（1m、5m、1h、1d 等）的 K 线数据到 MongoDB
- **最新价格发布**：将最新 K 线数据发布到 Kafka，供前端 WebSocket 订阅

### 1.3 在整体架构中的位置

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              mscoin 系统架构                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │   用户端    │────▶│  ucenter-api │────▶│  ucenter-rpc│                   │
│  └─────────────┘     └─────────────┘     └──────┬──────┘                   │
│                                                   │                          │
│                              Kafka (withdraw)     │                          │
│                                   │               │                          │
│                                   ▼               │                          │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        jobcenter (本服务)                            │   │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────┐  │   │
│  │  │ WithdrawConsumer│  │  RateSync Task  │  │   KlineSync Task    │  │   │
│  │  └────────┬────────┘  └────────┬────────┘  └──────────┬──────────┘  │   │
│  └───────────┼─────────────────────┼──────────────────────┼─────────────┘   │
│              │                     │                      │                  │
│              ▼                     ▼                      ▼                  │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                          外部依赖                                      │ │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐          │ │
│  │  │Bitcoin Core│  │  OKX API  │  │  MySQL   │  │  MongoDB  │          │ │
│  │  └───────────┘  └───────────┘  └───────────┘  └───────────┘          │ │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐                         │ │
│  │  │   Redis   │  │   Kafka   │  │market-rpc │                         │ │
│  │  └───────────┘  └───────────┘  └───────────┘                         │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

**关键交互说明**：

1. **ucenter-rpc → jobcenter**：用户提交提现申请后，ucenter-rpc 将事件发布到 Kafka 的 `withdraw` Topic，jobcenter 消费并处理
2. **jobcenter → Bitcoin Core**：执行实际的链上转账操作
3. **jobcenter → OKX API**：获取汇率和 K 线市场数据
4. **jobcenter → market-rpc**：查询币种信息和交易对列表
5. **jobcenter → ucenter-rpc**：查询用户钱包地址

### 1.4 技术选型

| 技术组件 | 用途 | 说明 |
|---------|------|------|
| go-zero | 服务框架 | 使用 ServiceGroup 模式统一管理消费者和定时任务 |
| Kafka | 消息队列 | 事件驱动通信，支持重试和死信队列 |
| MySQL | 关系数据库 | 存储提现记录状态 |
| MongoDB | 文档数据库 | 存储 K 线数据 |
| Redis | 缓存 | 恢复检查点、汇率缓存、最新价格缓存 |
| gRPC | RPC 通信 | 与 market-rpc、ucenter-rpc 通信 |

---

## 2. 目录结构

```
app/jobcenter/
├── main.go                          # 服务入口，启动消费者和定时任务
├── Dockerfile                       # Docker 构建文件
├── README.md                        # 服务简介
├── etc/
│   └── jobcenter.yaml               # 服务配置文件
└── internal/                        # 内部实现（不对外暴露）
    ├── config/
    │   └── config.go                # 配置结构定义
    ├── svc/
    │   └── service_context.go       # 依赖注入容器
    ├── consumer/
    │   ├── withdraw_consumer.go     # 提现事件消费者
    │   └── withdraw_consumer_test.go # 消费者单元测试
    ├── task/
    │   ├── service.go               # 定时任务调度服务
    │   └── service_test.go          # 任务调度单元测试
    ├── domain/
    │   └── service/
    │       ├── withdraw_service.go           # 提现处理领域服务
    │       ├── withdraw_service_test.go      # 提现服务测试
    │       ├── kline_service.go              # K线同步领域服务
    │       ├── kline_service_test.go         # K线服务测试
    │       ├── exchange_rate_service.go      # 汇率同步领域服务
    │       └── exchange_rate_service_test.go # 汇率服务测试
    ├── model/
    │   ├── withdraw_record.go        # 提现记录数据模型
    │   └── kline.go                  # K线数据模型
    └── repository/
        ├── withdraw_repository.go    # 提现记录数据访问
        └── kline_repository.go       # K线数据访问
```

### 2.1 目录职责说明

| 目录/文件 | 职责 |
|-----------|------|
| `config/` | 定义配置结构，支持 YAML 配置加载 |
| `svc/` | 依赖注入容器，管理所有外部依赖的生命周期 |
| `consumer/` | Kafka 消费者实现，负责消息适配和错误分类 |
| `task/` | 定时任务调度框架，使用原生 goroutine + Ticker |
| `domain/service/` | 领域服务，封装核心业务逻辑，不依赖基础设施 |
| `model/` | 数据模型定义，包括数据库映射和事件结构 |
| `repository/` | 数据访问层，封装数据库操作 |

### 2.2 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                      main.go (入口)                          │
│                    服务启动和生命周期                         │
├─────────────────────────────────────────────────────────────┤
│                     consumer / task                          │
│              消息适配层 / 任务调度层                          │
│         (仅负责消息反序列化、错误分类、任务注册)               │
├─────────────────────────────────────────────────────────────┤
│                    domain/service                            │
│                      领域服务层                               │
│          (核心业务逻辑，不依赖 Kafka、HTTP 等基础设施)         │
├─────────────────────────────────────────────────────────────┤
│                      repository                              │
│                      数据访问层                               │
│              (封装 MySQL、MongoDB 操作)                       │
├─────────────────────────────────────────────────────────────┤
│                        model                                 │
│                      数据模型层                               │
│           (数据库映射结构体、事件结构、状态常量)               │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. 文件详细说明

### 3.1 main.go - 服务入口

#### 3.1.1 文件职责

`main.go` 是 jobcenter 服务的启动入口，负责：
1. 解析命令行参数，加载配置文件
2. 初始化 ServiceContext，建立所有依赖连接
3. 创建 Kafka 消费者和定时任务服务
4. 使用 go-zero 的 ServiceGroup 统一管理生命周期

#### 3.1.2 启动流程

```go
func main() {
    // 1. 解析命令行参数
    flag.Parse()

    // 2. 加载配置文件
    var c config.Config
    conf.MustLoad(*configFile, &c)
    c.ServiceConf.MustSetUp()

    // 3. 初始化依赖注入容器
    ctx := svc.NewServiceContext(c)
    defer ctx.Close()

    // 4. 创建消费者和任务服务
    withdrawConsumer, err := consumer.NewWithdrawConsumer(ctx)
    if err != nil {
        panic(err)
    }
    taskService := task.NewService(ctx)

    // 5. 加入 ServiceGroup 统一管理
    group := service.NewServiceGroup()
    group.Add(withdrawConsumer)
    group.Add(taskService)

    // 6. 阻塞运行
    fmt.Printf("Starting jobcenter service %s...\n", c.Name)
    group.Start()
}
```

#### 3.1.3 ServiceGroup 工作原理

go-zero 的 `ServiceGroup` 是一个服务生命周期管理器：

- **Add(service)**：添加服务到组，服务需实现 `Service` 接口（`Start()` 和 `Stop()` 方法）
- **Start()**：阻塞运行所有服务，直到收到停止信号
- **Stop()**：通知所有服务停止，等待优雅退出

这种设计允许 Kafka 消费者和定时任务共享同一个进程，统一管理生命周期。

---

### 3.2 config/config.go - 配置定义

#### 3.2.1 文件职责

定义 jobcenter 服务的运行时配置结构，支持从 YAML 文件加载。

#### 3.2.2 配置结构详解

```go
// ScheduleConfig - 定时任务基础调度配置
type ScheduleConfig struct {
    Enabled         bool  // 是否启用该任务
    RunOnStart      bool  // 服务启动时是否立即执行一次
    IntervalSeconds int   // 执行间隔（秒）
}

// KlineTaskConfig - K线任务配置
type KlineTaskConfig struct {
    ScheduleConfig                      // 继承基础调度配置
    Period        string                // K线周期（如 "1m"、"5m"）
    PublishLatest bool                  // 是否发布最新K线到Kafka
    PublishTopic  string                // 发布目标Topic
}

// TasksConfig - 所有定时任务配置
type TasksConfig struct {
    RateSync ScheduleConfig             // 汇率同步任务配置
    Klines   []KlineTaskConfig          // K线同步任务配置列表
}

// Config - 服务总配置
type Config struct {
    service.ServiceConf                 // go-zero 服务配置
    Kafka        kafka.ConsumerConfig   // Kafka 消费者配置
    UcenterMysql mysqlx.Config          // MySQL 配置
    Redis        redisx.Config          // Redis 配置
    Mongo        mongox.Config          // MongoDB 配置
    MarketRPC    zrpc.RpcClientConf     // Market RPC 配置
    UcenterRPC   zrpc.RpcClientConf     // Ucenter RPC 配置
    OKX          okxx.Config            // OKX API 配置
    Tasks        TasksConfig            // 定时任务配置
    Bitcoin      btcx.NodeConfig        // Bitcoin Core 配置
}
```

#### 3.2.3 配置项说明

| 配置项 | 类型 | 说明 |
|--------|------|------|
| `Kafka.Brokers` | []string | Kafka 集群地址 |
| `Kafka.Topic` | string | 消费的 Topic 名称 |
| `Kafka.GroupID` | string | 消费者组 ID |
| `Kafka.DeadLetterTopic` | string | 死信队列 Topic |
| `Tasks.RateSync.Enabled` | bool | 是否启用汇率同步 |
| `Tasks.RateSync.IntervalSeconds` | int | 汇率同步间隔（秒） |
| `Tasks.Klines[].Period` | string | K线周期 |
| `Tasks.Klines[].PublishLatest` | bool | 是否发布到 Kafka |

---

### 3.3 svc/service_context.go - 依赖注入容器

#### 3.3.1 文件职责

`ServiceContext` 是 jobcenter 的依赖注入容器，负责：
1. 管理所有外部依赖的生命周期（数据库、缓存、RPC 客户端）
2. 创建和持有领域服务实例
3. 提供统一的资源释放机制

#### 3.3.2 结构定义

```go
type ServiceContext struct {
    Config                  config.Config           // 服务配置
    UcenterDB               *sqlx.DB                // MySQL 连接
    Cache                   *redisx.Client          // Redis 缓存
    Mongo                   *mongox.Client          // MongoDB 客户端
    Queue                   map[string]kafka.Producer // Kafka 生产者池
    MarketClient            marketpb.MarketClient   // Market RPC 客户端
    AssetClient             assetpb.AssetClient     // Ucenter Asset RPC 客户端
    WithdrawService         *domainservice.WithdrawService        // 提现服务
    ExchangeRateSyncService *domainservice.ExchangeRateSyncService // 汇率服务
    KlineSyncService        *domainservice.KlineSyncService       // K线服务
}
```

#### 3.3.3 初始化流程

```go
func NewServiceContext(c config.Config) *ServiceContext {
    // 1. 建立 MySQL 连接
    db, err := mysqlx.New(c.UcenterMysql)

    // 2. 建立 Redis 连接
    cache := redisx.New(c.Redis)

    // 3. 建立 RPC 客户端
    marketConn := zrpc.MustNewClient(c.MarketRPC)
    ucenterConn := zrpc.MustNewClient(c.UcenterRPC)
    marketClient := marketpb.NewMarketClient(marketConn.Conn())
    assetClient := assetpb.NewAssetClient(ucenterConn.Conn())

    // 4. 创建 Bitcoin Core 发送器
    bitcoinSender, err := btcx.NewWithdrawSender(c.Bitcoin)

    // 5. 创建 OKX 客户端
    okxClient, err := okxx.NewClient(c.OKX)

    // 6. 建立 MongoDB 连接（可选）
    var mongoClient *mongox.Client
    if c.Mongo.URI != "" {
        mongoClient, err = mongox.New(c.Mongo)
    }

    // 7. 创建 Repository
    withdrawRepo := repository.NewWithdrawRepository(db)
    klineRepo := repository.NewKlineRepository(mongoClient.Database())

    // 8. 构建 Kafka 生产者池
    publishers := buildTopicPublishers(c)

    // 9. 创建领域服务
    return &ServiceContext{
        // ... 字段赋值
        WithdrawService: domainservice.NewWithdrawService(...),
        ExchangeRateSyncService: domainservice.NewExchangeRateSyncService(...),
        KlineSyncService: domainservice.NewKlineSyncService(...),
    }
}
```

#### 3.3.4 资源释放

```go
func (s *ServiceContext) Close() {
    // 关闭所有 Kafka 生产者
    for _, producer := range s.Queue {
        _ = producer.Close()
    }
    // 断开 MongoDB 连接
    if s.Mongo != nil {
        _ = s.Mongo.Disconnect(context.Background())
    }
    // 关闭 MySQL 连接
    if s.UcenterDB != nil {
        _ = s.UcenterDB.Close()
    }
}
```

---

### 3.4 consumer/withdraw_consumer.go - 提现消费者

#### 3.4.1 文件职责

`WithdrawConsumer` 负责消费 Kafka 中的提现事件消息，并调用领域服务处理。

#### 3.4.2 消费者创建

```go
func NewWithdrawConsumer(svcCtx *svc.ServiceContext) (coreservice.Service, error) {
    return kafka.NewConsumerService(
        svcCtx.Config.Kafka,
        // 消息处理函数
        func(ctx context.Context, message kafka.Message) error {
            var event model.WithdrawRecordEvent
            if err := json.Unmarshal(message.Value, &event); err != nil {
                return domainservice.NewNonRetryableError(fmt.Errorf("unmarshal withdraw event: %w", err))
            }
            return svcCtx.WithdrawService.ProcessApplied(ctx, &event)
        },
        // 错误分类函数
        classifyWithdrawError,
    )
}
```

#### 3.4.3 错误分类策略

```go
func classifyWithdrawError(err error) kafka.ConsumeAction {
    if err == nil {
        return kafka.ConsumeAck           // 成功，确认消息
    }
    if domainservice.IsNonRetryable(err) {
        return kafka.ConsumeDeadLetter    // 不可重试，发送到死信队列
    }
    return kafka.ConsumeRetry             // 可重试，重新投递
}
```

**错误分类决策树**：

```
错误分类
├── nil → ConsumeAck（确认消息）
├── NonRetryableError → ConsumeDeadLetter（死信队列）
│   ├── 消息格式错误（反序列化失败）
│   ├── 不支持的币种
│   └── 记录状态不支持处理
└── 其他错误 → ConsumeRetry（重试）
    ├── 数据库连接失败
    ├── RPC 调用失败
    └── Bitcoin Core 调用失败
```

---

### 3.5 task/service.go - 定时任务调度

#### 3.5.1 文件职责

`Service` 管理所有基于 goroutine 的周期性任务，使用 Go 原生的 `time.Ticker` 实现。

#### 3.5.2 为什么不用 cron 框架

选择原生实现的原因：
1. 项目只需要长生命周期的间隔任务，不需要 cron 表达式的复杂调度
2. 任务之间相互独立，无复杂依赖关系
3. 需要精细控制任务执行行为（如 RunOnStart）
4. 减少外部依赖，降低维护成本

#### 3.5.3 核心结构

```go
// intervalJob - 周期性任务封装
type intervalJob struct {
    name     string            // 任务名称
    schedule config.ScheduleConfig // 调度配置
    run      jobRunner         // 执行函数
}

// Service - 任务调度服务
type Service struct {
    ctx    context.Context     // 服务上下文
    cancel context.CancelFunc  // 取消函数
    waiter sync.WaitGroup      // 等待组
    jobs   []intervalJob       // 任务列表
}
```

#### 3.5.4 任务注册

```go
func (s *Service) registerJobs(svcCtx *svc.ServiceContext) {
    // 注册汇率同步任务
    s.jobs = append(s.jobs, intervalJob{
        name:     "rate-sync",
        schedule: svcCtx.Config.Tasks.RateSync,
        run:      svcCtx.ExchangeRateSyncService.SyncUSDCNY,
    })

    // 注册 K线同步任务（按周期配置多个）
    for _, item := range svcCtx.Config.Tasks.Klines {
        cfg := item
        period := cfg.Period
        name := fmt.Sprintf("kline-sync-%s", period)
        s.jobs = append(s.jobs, intervalJob{
            name:     name,
            schedule: cfg.ScheduleConfig,
            run: func(ctx context.Context) error {
                return svcCtx.KlineSyncService.SyncPeriod(ctx, period, cfg.PublishLatest, cfg.PublishTopic)
            },
        })
    }
}
```

#### 3.5.5 任务执行循环

```go
func (s *Service) runLoop(job intervalJob) {
    defer s.waiter.Done()

    // 计算执行间隔
    interval := time.Duration(defaultPositive(job.schedule.IntervalSeconds, 60)) * time.Second
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    logx.Infof("starting jobcenter task %s with interval %s", job.name, interval)

    // RunOnStart：启动时立即执行一次
    if job.schedule.RunOnStart {
        s.execute(job)
    }

    // 主循环
    for {
        select {
        case <-s.ctx.Done():
            // 收到停止信号
            logx.Infof("jobcenter task %s stopped", job.name)
            return
        case <-ticker.C:
            // 定时触发
            s.execute(job)
        }
    }
}
```

**调度时序图**：

```
服务启动
    │
    ├─── RunOnStart=true? ───┐
    │                        │
    │                        ▼
    │                    立即执行一次
    │                        │
    ▼                        │
创建 Ticker ◄────────────────┘
    │
    │  ┌──────────────────────────────┐
    │  │                              │
    ▼  ▼                              │
等待 Ticker.C 或 ctx.Done()           │
    │                                 │
    ├─── Ticker.C ──► 执行任务 ────────┤
    │                                 │
    └─── ctx.Done() ──► 退出循环 ─────┘
```

---

### 3.6 domain/service/withdraw_service.go - 提现领域服务

#### 3.6.1 文件职责

`WithdrawService` 是提现处理的核心领域服务，负责完整的提现执行流程。

#### 3.6.2 核心结构

```go
type WithdrawService struct {
    repo        withdrawRepository    // 提现记录 Repository
    market      marketCoinFinder      // Market RPC 客户端
    asset       assetWalletFinder     // Ucenter Asset RPC 客户端
    cache       txCache               // Redis 缓存
    bitcoinSend btcx.WithdrawSender   // Bitcoin Core 发送器
}
```

#### 3.6.3 提现处理流程

```go
func (s *WithdrawService) ProcessApplied(ctx context.Context, event *model.WithdrawRecordEvent) error {
    // 1. 参数校验
    if event == nil || event.Id <= 0 || event.MemberId <= 0 {
        return NewNonRetryableError(...)
    }

    // 2. 查询提现记录
    record, err := s.repo.FindByID(ctx, event.Id)
    if record == nil {
        return fmt.Errorf("withdraw record %d is not committed yet", event.Id)
    }

    // 3. 状态检查
    if record.Status == model.WithdrawStatusSuccess {
        return nil  // 已成功，幂等返回
    }
    if record.Status != model.WithdrawStatusProcessing {
        return NewNonRetryableError(...)  // 状态不支持
    }

    // 4. 恢复检查：尝试从缓存恢复已广播的交易
    if finalized, err := s.finalizeFromCache(ctx, record.Id); finalized {
        return nil  // 从缓存恢复成功
    }

    // 5. 查询币种信息
    coin, err := s.market.FindCoinById(ctx, &marketpb.MarketReq{Id: record.CoinId})
    if coin.Unit != "BTC" {
        return NewNonRetryableError(...)  // 仅支持 BTC
    }

    // 6. 查询用户钱包地址
    wallet, err := s.asset.FindWalletBySymbol(ctx, &assetpb.AssetReq{
        UserId:   record.MemberId,
        CoinName: coin.Unit,
    })

    // 7. 调用 Bitcoin Core 执行链上转账
    txID, err := s.bitcoinSend.Send(ctx, wallet.Address, record.Address, record.TotalAmount, record.ArrivedAmount)

    // 8. 写入缓存检查点
    dealTime := time.Now().UnixMilli()
    s.cache.SetWithExpireCtx(ctx, withdrawTxCacheKey(record.Id), WithdrawTxCacheEntry{
        TxID:     txID,
        DealTime: dealTime,
    }, 24*time.Hour)

    // 9. 更新数据库状态
    updated, err := s.repo.MarkSuccess(ctx, record.Id, txID, dealTime)

    return nil
}
```

#### 3.6.4 提现处理流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           提现处理完整流程                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Kafka 消息                                                                  │
│      │                                                                      │
│      ▼                                                                      │
│  ┌─────────────────┐                                                        │
│  │  1. 参数校验    │                                                        │
│  └────────┬────────┘                                                        │
│           │                                                                 │
│           ▼                                                                 │
│  ┌─────────────────┐     失败     ┌─────────────────┐                      │
│  │ 2. 查询记录     │─────────────▶│   重试等待      │                      │
│  └────────┬────────┘              └─────────────────┘                      │
│           │ 记录存在                                                       │
│           ▼                                                                 │
│  ┌─────────────────┐     已成功     ┌─────────────────┐                      │
│  │  3. 状态检查    │─────────────▶│   幂等返回      │                      │
│  └────────┬────────┘              └─────────────────┘                      │
│           │ Processing                                                       │
│           ▼                                                                 │
│  ┌─────────────────┐     有缓存     ┌─────────────────┐                      │
│  │ 4. 恢复检查     │─────────────▶│  从缓存恢复     │──┐                   │
│  └────────┬────────┘              └─────────────────┘  │                   │
│           │ 无缓存                        成功          │                   │
│           ▼                                            │                   │
│  ┌─────────────────┐                                  │                   │
│  │ 5. 查询币种     │                                  │                   │
│  └────────┬────────┘                                  │                   │
│           │                                           │                   │
│           ▼                                           │                   │
│  ┌─────────────────┐                                  │                   │
│  │ 6. 查询钱包     │                                  │                   │
│  └────────┬────────┘                                  │                   │
│           │                                           │                   │
│           ▼                                           │                   │
│  ┌─────────────────┐                                  │                   │
│  │7. Bitcoin Core  │                                  │                   │
│  │   链上转账      │                                  │                   │
│  └────────┬────────┘                                  │                   │
│           │ txid                                      │                   │
│           ▼                                           │                   │
│  ┌─────────────────┐                                  │                   │
│  │ 8. 缓存检查点   │                                  │                   │
│  └────────┬────────┘                                  │                   │
│           │                                           │                   │
│           ▼                                           │                   │
│  ┌─────────────────┐                                  │                   │
│  │ 9. 更新数据库   │                                  │                   │
│  └────────┬────────┘                                  │                   │
│           │                                           │                   │
│           ▼                                           ▼                   │
│        处理完成 ◄─────────────────────────────────────┘                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### 3.6.5 恢复机制详解

**为什么需要恢复机制**：

链上交易广播和 MySQL 更新不是原子操作。可能出现以下场景：
1. 交易已广播成功（txid 已获得）
2. Redis 缓存检查点写入成功
3. MySQL 更新失败（网络问题、服务重启等）

如果没有恢复机制，Kafka 重试时会重新广播交易，导致**双重支付**。

**恢复机制工作原理**：

```go
func (s *WithdrawService) finalizeFromCache(ctx context.Context, recordID int64) (bool, error) {
    // 尝试从 Redis 读取缓存检查点
    var entry WithdrawTxCacheEntry
    if err := s.cache.GetCtx(ctx, withdrawTxCacheKey(recordID), &entry); err != nil {
        if errors.Is(err, goredis.Nil) {
            return false, nil  // 缓存不存在，继续正常流程
        }
        return false, fmt.Errorf("load withdraw tx checkpoint: %w", err)
    }

    // 缓存存在，直接更新数据库（不重新广播交易）
    _, err := s.repo.MarkSuccess(ctx, recordID, entry.TxID, entry.DealTime)
    return true, err
}
```

**缓存数据结构**：

```go
type WithdrawTxCacheEntry struct {
    TxID     string `json:"txId"`     // 链上交易哈希
    DealTime int64  `json:"dealTime"` // 处理完成时间戳（毫秒）
}
```

**缓存键格式**：`JOBCENTER::WITHDRAW::TX::{recordID}`

**缓存过期时间**：24 小时

#### 3.6.6 NonRetryableError 详解

```go
type NonRetryableError struct {
    cause error
}

func NewNonRetryableError(err error) error {
    if err == nil {
        return nil
    }
    return &NonRetryableError{cause: err}
}

func IsNonRetryable(err error) bool {
    var target *NonRetryableError
    return errors.As(err, &target)
}
```

**适用场景**：
- 消息格式错误（反序列化失败）
- 不支持的币种（当前仅支持 BTC）
- 记录状态不支持处理（非 Processing 状态）
- 交易已广播但缓存和数据库都写入失败（需要人工介入）

---

### 3.7 domain/service/kline_service.go - K线同步服务

#### 3.7.1 文件职责

`KlineSyncService` 负责从 OKX API 同步 K 线数据到 MongoDB，并可选发布到 Kafka。

#### 3.7.2 核心结构

```go
type KlineSyncService struct {
    marketClient visibleExchangeCoinFinder  // Market RPC 客户端
    okxClient    okxx.Client                // OKX API 客户端
    repo         klineWriter                // K线 Repository
    cache        priceCache                 // Redis 缓存
    publishers   map[string]kafka.Producer  // Kafka 生产者池
}
```

#### 3.7.3 K线同步流程

```go
func (s *KlineSyncService) SyncPeriod(ctx context.Context, period string, publishLatest bool, publishTopic string) error {
    // 1. 从 Market RPC 获取所有可见交易对
    pairs, err := s.marketClient.FindExchangeCoinVisible(ctx, &marketpb.MarketReq{})

    // 2. 遍历每个交易对同步数据
    var joinedErr error
    for _, pair := range pairs.List {
        if err := s.syncSymbol(ctx, pair.Symbol, period, publishLatest, publishTopic); err != nil {
            joinedErr = errors.Join(joinedErr, fmt.Errorf("sync %s %s: %w", pair.Symbol, period, err))
        }
    }
    return joinedErr
}

func (s *KlineSyncService) syncSymbol(ctx context.Context, symbol string, period string, publishLatest bool, publishTopic string) error {
    // 1. 调用 OKX API 获取 K 线数据
    // instID 格式：BTC-USDT（使用 "-" 替代 "/"）
    candles, err := s.okxClient.FetchCandles(ctx, strings.ReplaceAll(symbol, "/", "-"), period)

    // 2. 转换为 model.Kline 格式
    list := make([]*model.Kline, 0, len(candles))
    for _, candle := range candles {
        item := model.NewKlineFromCandle(period, candle)
        if item != nil {
            list = append(list, item)
        }
    }

    // 3. 写入 MongoDB（删除尾部重叠数据后插入）
    if err := s.repo.ReplaceBatch(ctx, symbol, period, list); err != nil {
        return err
    }

    // 4. 如果是 1m 周期，缓存最新价格并发布到 Kafka
    if period == "1m" && len(list) > 0 {
        latest := list[0]

        // 缓存最新价格到 Redis
        cacheKey := strings.ReplaceAll(symbol, "/", "::") + "::RATE"
        s.cache.SetCtx(ctx, cacheKey, latest.ClosePrice)

        // 发布最新 K 线到 Kafka（如果配置）
        if publishLatest && publishTopic != "" {
            publisher := s.publishers[publishTopic]
            payload, _ := json.Marshal(latest)
            publisher.PushWithKey(ctx, symbol, string(payload))
        }
    }

    return nil
}
```

#### 3.7.4 K线同步流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           K线同步流程                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  定时触发                                                                    │
│      │                                                                      │
│      ▼                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  1. Market RPC: FindExchangeCoinVisible()                           │   │
│  │     获取所有可见交易对列表 [BTC/USDT, ETH/USDT, ...]                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│      │                                                                      │
│      ▼                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  2. 遍历每个交易对                                                    │   │
│  │  ┌───────────────────────────────────────────────────────────────┐  │   │
│  │  │  2.1 OKX API: FetchCandles(instID, period)                    │  │   │
│  │  │      例如：FetchCandles("BTC-USDT", "1m")                     │  │   │
│  │  └───────────────────────────────────────────────────────────────┘  │   │
│  │      │                                                               │   │
│  │      ▼                                                               │   │
│  │  ┌───────────────────────────────────────────────────────────────┐  │   │
│  │  │  2.2 转换数据格式                                              │  │   │
│  │  │      Candle → model.Kline                                      │  │   │
│  │  └───────────────────────────────────────────────────────────────┘  │   │
│  │      │                                                               │   │
│  │      ▼                                                               │   │
│  │  ┌───────────────────────────────────────────────────────────────┐  │   │
│  │  │  2.3 MongoDB: ReplaceBatch()                                   │  │   │
│  │  │      删除尾部重叠数据 + 批量插入                                │  │   │
│  │  └───────────────────────────────────────────────────────────────┘  │   │
│  │      │                                                               │   │
│  │      ▼ (仅 1m 周期)                                                  │   │
│  │  ┌───────────────────────────────────────────────────────────────┐  │   │
│  │  │  2.4 Redis: 缓存最新价格                                       │  │   │
│  │  │      key: BTC::USDT::RATE                                      │  │   │
│  │  │      value: closePrice                                         │  │   │
│  │  └───────────────────────────────────────────────────────────────┘  │   │
│  │      │                                                               │   │
│  │      ▼ (如果 publishLatest=true)                                     │   │
│  │  ┌───────────────────────────────────────────────────────────────┐  │   │
│  │  │  2.5 Kafka: 发布最新 K 线                                      │  │   │
│  │  │      topic: kline_1m                                           │  │   │
│  │  │      key: BTC/USDT                                             │  │   │
│  │  │      value: Kline JSON                                         │  │   │
│  │  └───────────────────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

### 3.8 domain/service/exchange_rate_service.go - 汇率同步服务

#### 3.8.1 文件职责

`ExchangeRateSyncService` 负责从 OKX API 获取 USDT/CNY 实时汇率并缓存到 Redis。

#### 3.8.2 核心实现

```go
const usdtCNYRateCacheKey = "USDT::CNY::RATE"

type ExchangeRateSyncService struct {
    cache   exchangeRateCache    // Redis 缓存
    fetcher exchangeRateFetcher  // OKX API 客户端
}

func (s *ExchangeRateSyncService) SyncUSDCNY(ctx context.Context) error {
    // 1. 调用 OKX API 获取汇率
    rate, err := s.fetcher.FetchExchangeRate(ctx)
    if rate == nil || rate.USDCNY <= 0 {
        return fmt.Errorf("exchange-rate payload is invalid")
    }

    // 2. 缓存到 Redis
    if err := s.cache.SetCtx(ctx, usdtCNYRateCacheKey, rate.USDCNY); err != nil {
        return fmt.Errorf("cache usd/cny exchange rate: %w", err)
    }
    return nil
}
```

#### 3.8.3 缓存数据

- **缓存键**：`USDT::CNY::RATE`
- **缓存值**：float64 类型，表示 1 USDT 等于多少 CNY
- **更新频率**：由配置 `Tasks.RateSync.IntervalSeconds` 决定（默认 300 秒）

---

### 3.9 model/withdraw_record.go - 提现记录模型

#### 3.9.1 状态常量

```go
const (
    WithdrawStatusProcessing int32 = iota  // 0: 处理中
    WithdrawStatusWaiting                   // 1: 等待审核
    WithdrawStatusFail                      // 2: 失败
    WithdrawStatusSuccess                   // 3: 成功
)
```

**状态流转图**：

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           提现状态流转                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  用户提交提现申请                                                            │
│      │                                                                      │
│      ▼                                                                      │
│  ┌─────────────────┐                                                        │
│  │   Processing    │ ◄─── jobcenter 处理此状态                              │
│  │    (处理中)     │                                                        │
│  └────────┬────────┘                                                        │
│           │                                                                 │
│     ┌─────┴─────┐                                                           │
│     │           │                                                           │
│     ▼           ▼                                                           │
│  ┌─────────┐ ┌─────────┐                                                    │
│  │ Success │ │  Fail   │                                                    │
│  │ (成功)  │ │ (失败)  │                                                    │
│  └─────────┘ └─────────┘                                                    │
│                                                                             │
│  Waiting (等待审核) ─── 不由 jobcenter 处理                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### 3.9.2 数据结构

```go
// WithdrawRecord - 数据库映射
type WithdrawRecord struct {
    Id                int64   `db:"id"`                  // 主键
    MemberId          int64   `db:"member_id"`           // 用户ID
    CoinId            int64   `db:"coin_id"`             // 币种ID
    TotalAmount       float64 `db:"total_amount"`        // 提现总额
    Fee               float64 `db:"fee"`                 // 手续费
    ArrivedAmount     float64 `db:"arrived_amount"`      // 到账金额
    Address           string  `db:"address"`             // 目标地址
    Remark            string  `db:"remark"`              // 备注
    TransactionNumber string  `db:"transaction_number"`  // 链上交易哈希
    CanAutoWithdraw   int32   `db:"can_auto_withdraw"`   // 是否允许自动提现
    IsAuto            int32   `db:"isAuto"`              // 是否为自动提现
    Status            int32   `db:"status"`              // 状态
    CreateTime        int64   `db:"create_time"`         // 创建时间
    DealTime          int64   `db:"deal_time"`           // 处理完成时间
}

// WithdrawRecordEvent - Kafka 消息结构
type WithdrawRecordEvent struct {
    Id                int64   `json:"Id"`
    MemberId          int64   `json:"MemberId"`
    CoinId            int64   `json:"CoinId"`
    // ... 其他字段
}
```

---

### 3.10 model/kline.go - K线数据模型

#### 3.10.1 数据结构

```go
type Kline struct {
    Period       string  `bson:"period,omitempty" json:"period"`           // K线周期
    OpenPrice    float64 `bson:"openPrice,omitempty" json:"openPrice"`     // 开盘价
    HighestPrice float64 `bson:"highestPrice,omitempty" json:"highestPrice"` // 最高价
    LowestPrice  float64 `bson:"lowestPrice,omitempty" json:"lowestPrice"` // 最低价
    ClosePrice   float64 `bson:"closePrice,omitempty" json:"closePrice"`   // 收盘价
    Time         int64   `bson:"time,omitempty" json:"time"`               // 时间戳（毫秒）
    Count        float64 `bson:"count,omitempty" json:"count"`             // 成交笔数
    Volume       float64 `bson:"volume,omitempty" json:"volume"`           // 成交量
    Turnover     float64 `bson:"turnover,omitempty" json:"turnover"`       // 成交额
    TimeStr      string  `bson:"timeStr,omitempty" json:"timeStr"`         // 格式化时间
}
```

#### 3.10.2 集合命名规则

```go
func (k *Kline) TableName(symbol string, period string) string {
    return "exchange_kline_" + symbol + "_" + period
}
```

例如：`exchange_kline_BTC/USDT_1m`

#### 3.10.3 数据转换

```go
func NewKlineFromCandle(period string, candle *okxx.Candle) *Kline {
    if candle == nil {
        return nil
    }
    return &Kline{
        Period:       period,
        OpenPrice:    candle.OpenPrice,
        HighestPrice: candle.HighestPrice,
        LowestPrice:  candle.LowestPrice,
        ClosePrice:   candle.ClosePrice,
        Time:         candle.Time,
        Count:        candle.Count,
        Volume:       candle.Volume,
        Turnover:     candle.Turnover,
        TimeStr:      time.UnixMilli(candle.Time).Format("2006-01-02 15:04:05"),
    }
}
```

---

### 3.11 repository/withdraw_repository.go - 提现数据访问

#### 3.11.1 文件职责

封装提现记录的 MySQL 数据访问操作。

#### 3.11.2 核心方法

```go
// FindByID - 根据主键查询
func (r *WithdrawRepository) FindByID(ctx context.Context, id int64) (*model.WithdrawRecord, error) {
    var record model.WithdrawRecord
    err := r.db.GetContext(ctx, &record, "SELECT * FROM withdraw_record WHERE id=? LIMIT 1", id)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil  // 记录不存在返回 nil，非错误
    }
    return &record, err
}

// MarkSuccess - 标记成功（乐观锁）
func (r *WithdrawRepository) MarkSuccess(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
    result, err := r.db.ExecContext(
        ctx,
        "UPDATE withdraw_record SET transaction_number=?, status=?, deal_time=? WHERE id=? AND status=?",
        txID,
        model.WithdrawStatusSuccess,
        dealTime,
        id,
        model.WithdrawStatusProcessing,  // 乐观锁条件
    )
    // ...
    rowsAffected, _ := result.RowsAffected()
    return rowsAffected > 0, nil
}
```

**乐观锁说明**：

`WHERE status=?` 条件确保只有当前状态为 `Processing` 时才更新，防止：
- 重复 Kafka 投递覆盖已完成的记录
- 人工后台更正被自动处理覆盖

---

### 3.12 repository/kline_repository.go - K线数据访问

#### 3.12.1 文件职责

封装 K 线数据的 MongoDB 写入操作。

#### 3.12.2 更新策略

采用"删除尾部 + 批量插入"策略：

```go
func (r *KlineRepository) ReplaceBatch(ctx context.Context, symbol string, period string, list []*model.Kline) error {
    collection := r.db.Collection((&model.Kline{}).TableName(symbol, period))

    // 1. 计算分界时间点（最新数据的最早时间）
    cutoff := list[len(list)-1].Time

    // 2. 删除该时间点之后的所有数据
    collection.DeleteMany(ctx, bson.D{{Key: "time", Value: bson.D{{Key: "$gte", Value: cutoff}}}})

    // 3. 批量插入新数据
    documents := make([]interface{}, 0, len(list))
    for _, item := range list {
        if item != nil {
            documents = append(documents, item)
        }
    }
    collection.InsertMany(ctx, documents)

    return nil
}
```

**为什么采用此策略**：

1. OKX 每次请求返回最近时间窗口的数据，尾部重叠是预期行为
2. 批量删除 + 插入保持逻辑紧凑，接近旧服务行为
3. 该集合以追加为主，不属于核心事务状态，重写方式可接受

---

## 4. Kafka 消费流程

### 4.1 消息格式

#### 4.1.1 提现事件消息

**Topic**: `withdraw`

**消息体** (JSON):

```json
{
    "Id": 123,
    "MemberId": 1001,
    "CoinId": 1,
    "TotalAmount": 1.5,
    "Fee": 0.0001,
    "ArrivedAmount": 1.4999,
    "Address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
    "Remark": "用户提现",
    "TransactionNumber": "",
    "CanAutoWithdraw": 1,
    "IsAuto": 1,
    "Status": 0,
    "CreateTime": 1710000000000,
    "DealTime": 0
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| Id | int64 | 提现记录主键 |
| MemberId | int64 | 用户 ID |
| CoinId | int64 | 币种 ID |
| TotalAmount | float64 | 提现总额（含手续费） |
| Fee | float64 | 手续费 |
| ArrivedAmount | float64 | 到账金额 |
| Address | string | 目标提现地址 |
| Status | int32 | 状态（0=处理中） |

### 4.2 消费者配置

```yaml
Kafka:
  Brokers:
    - kafka:9092
  Topic: withdraw
  GroupID: jobcenter-withdraw
  RetryBackoffMs: 2000      # 重试退避时间
  MaxWaitMs: 3000           # 最大等待时间
  DeadLetterTopic: withdraw.dlq  # 死信队列
  AllowAutoTopicCreate: true
```

### 4.3 消息处理流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Kafka 消息处理流程                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Kafka Broker                                                               │
│      │                                                                      │
│      │ 拉取消息                                                             │
│      ▼                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      WithdrawConsumer                               │   │
│  │  ┌───────────────────────────────────────────────────────────────┐  │   │
│  │  │  1. 反序列化消息                                                │  │   │
│  │  │     json.Unmarshal(message.Value, &event)                      │  │   │
│  │  └───────────────────────────────────────────────────────────────┘  │   │
│  │      │                                                               │   │
│  │      │ 成功                              失败                        │   │
│  │      ▼                                  ▼                           │   │
│  │  ┌─────────────────────┐          ┌─────────────────────┐          │   │
│  │  │ 2. 调用领域服务     │          │ NonRetryableError   │          │   │
│  │  │ ProcessApplied()    │          │ → 死信队列          │          │   │
│  │  └──────────┬──────────┘          └─────────────────────┘          │   │
│  │             │                                                        │   │
│  │     ┌───────┴───────┐                                                │   │
│  │     │               │                                                │   │
│  │     ▼               ▼                                                │   │
│  │  ┌─────────┐   ┌─────────────┐                                       │   │
│  │  │  nil    │   │   error     │                                       │   │
│  │  └────┬────┘   └──────┬──────┘                                       │   │
│  │       │               │                                              │   │
│  │       ▼               ▼                                              │   │
│  │  ┌─────────┐   ┌─────────────────────────────────────────────┐      │   │
│  │  │  Ack    │   │  3. 错误分类 classifyWithdrawError()        │      │   │
│  │  └─────────┘   └─────────────────────────────────────────────┘      │   │
│  │                        │                                             │   │
│  │              ┌─────────┼─────────┐                                   │   │
│  │              │         │         │                                   │   │
│  │              ▼         ▼         ▼                                   │   │
│  │         ┌────────┐ ┌────────┐ ┌────────┐                            │   │
│  │         │  Ack   │ │ Retry  │ │DeadLett│                            │   │
│  │         └────────┘ └────────┘ └────────┘                            │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.4 错误处理策略

| 错误类型 | 处理动作 | 说明 |
|---------|---------|------|
| nil | Ack | 消息处理成功，确认消费 |
| 反序列化失败 | Dead Letter | 毒消息，不重试 |
| 不支持的币种 | Dead Letter | 业务限制，不重试 |
| 状态不支持 | Dead Letter | 业务状态已变更 |
| 记录不存在 | Retry | 等待事务提交 |
| RPC 调用失败 | Retry | 临时性错误 |
| Bitcoin Core 失败 | Retry | 临时性错误 |
| 数据库更新失败 | Retry | 有缓存检查点保护 |

### 4.5 死信队列处理

死信队列 Topic: `withdraw.dlq`

死信队列中的消息需要人工介入处理，可能的原因：
- 消息格式错误
- 不支持的币种
- 记录状态已变更
- 交易已广播但无法记录状态

---

## 5. 定时任务机制

### 5.1 任务配置

```yaml
Tasks:
  # 汇率同步任务
  RateSync:
    Enabled: true           # 是否启用
    RunOnStart: true        # 启动时立即执行
    IntervalSeconds: 300    # 执行间隔（秒）

  # K线同步任务（支持多周期）
  Klines:
    - Period: "1m"
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 60
      PublishLatest: true
      PublishTopic: "kline_1m"
    - Period: "5m"
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 300
      PublishLatest: false
      PublishTopic: ""
    # ... 其他周期
```

### 5.2 任务调度原理

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           定时任务调度原理                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  服务启动                                                                    │
│      │                                                                      │
│      ▼                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  task.Service.Start()                                                │   │
│  │                                                                      │   │
│  │  for _, job := range s.jobs {                                        │   │
│  │      if !job.schedule.Enabled {                                      │   │
│  │          continue  // 跳过未启用的任务                                │   │
│  │      }                                                               │   │
│  │      s.waiter.Add(1)                                                 │   │
│  │      go s.runLoop(job)  // 启动独立 goroutine                        │   │
│  │  }                                                                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│      │                                                                      │
│      ▼                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  每个 goroutine 的 runLoop()                                         │   │
│  │                                                                      │   │
│  │  interval := time.Duration(job.schedule.IntervalSeconds) * time.Second│  │
│  │  ticker := time.NewTicker(interval)                                  │   │
│  │                                                                      │   │
│  │  if job.schedule.RunOnStart {                                        │   │
│  │      s.execute(job)  // 立即执行一次                                  │   │
│  │  }                                                                   │   │
│  │                                                                      │   │
│  │  for {                                                                │   │
│  │      select {                                                        │   │
│  │      case <-s.ctx.Done():                                            │   │
│  │          return  // 收到停止信号                                      │   │
│  │      case <-ticker.C:                                                │   │
│  │          s.execute(job)  // 定时执行                                  │   │
│  │      }                                                               │   │
│  │  }                                                                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.3 任务列表

| 任务名称 | 执行内容 | 默认间隔 | 说明 |
|---------|---------|---------|------|
| rate-sync | 同步 USDT/CNY 汇率 | 300s | 从 OKX 获取汇率缓存到 Redis |
| kline-sync-1m | 同步 1 分钟 K 线 | 60s | 发布最新 K 线到 Kafka |
| kline-sync-5m | 同步 5 分钟 K 线 | 300s | - |
| kline-sync-1H | 同步 1 小时 K 线 | 3600s | - |
| kline-sync-1D | 同步 1 天 K 线 | 86400s | - |
| ... | ... | ... | 支持多种周期 |

### 5.4 RunOnStart 机制

`RunOnStart` 配置允许服务启动时立即执行一次任务，适用于：
- 快速预热数据，避免等待首个调度周期
- 服务重启后快速恢复数据

```
服务启动
    │
    ├─── RunOnStart=true ───┐
    │                       │
    │                       ▼
    │                   立即执行一次
    │                       │
    ▼                       │
等待首个调度周期 ◄───────────┘
    │
    ▼
定时执行
```

---

## 6. 与其他服务的调用关系

### 6.1 与 market-rpc 的调用

#### 6.1.1 查询币种信息

**调用场景**：提现处理时查询币种信息

**RPC 方法**：`FindCoinById`

**请求参数**：
```go
&marketpb.MarketReq{Id: record.CoinId}
```

**响应**：
```go
&marketpb.Coin{
    Unit: "BTC",  // 币种单位
    // ... 其他字段
}
```

**用途**：
- 验证币种是否支持（当前仅支持 BTC）
- 获取币种单位用于查询用户钱包

#### 6.1.2 查询可见交易对

**调用场景**：K 线同步时获取所有可见交易对

**RPC 方法**：`FindExchangeCoinVisible`

**请求参数**：
```go
&marketpb.MarketReq{}
```

**响应**：
```go
&marketpb.ExchangeCoinRes{
    List: []*marketpb.ExchangeCoin{
        {Symbol: "BTC/USDT"},
        {Symbol: "ETH/USDT"},
        // ...
    },
}
```

**用途**：
- 获取需要同步 K 线的交易对列表
- 遍历每个交易对调用 OKX API

### 6.2 与 ucenter-rpc 的调用

#### 6.2.1 查询用户钱包地址

**调用场景**：提现处理时获取用户热钱包地址

**RPC 方法**：`FindWalletBySymbol`

**请求参数**：
```go
&assetpb.AssetReq{
    UserId:   record.MemberId,
    CoinName: "BTC",
}
```

**响应**：
```go
&assetpb.MemberWallet{
    Address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",  // 用户热钱包地址
    // ... 其他字段
}
```

**用途**：
- 获取发送方地址（用户热钱包）
- 用于 Bitcoin Core 转账

### 6.3 与 Bitcoin Core 的交互

#### 6.3.1 执行链上转账

**调用场景**：提现处理时执行实际转账

**接口**：`btcx.WithdrawSender.Send`

**参数**：
```go
Send(ctx context.Context,
    fromAddress string,   // 发送方地址（用户热钱包）
    toAddress string,     // 接收方地址（提现目标地址）
    totalAmount float64,  // 总金额
    arrivedAmount float64 // 到账金额
) (string, error)         // 返回交易哈希
```

**Bitcoin Core RPC 调用**：
- 使用配置的 URL、用户名、密码连接 Bitcoin Core
- 调用 `sendfrom` 或类似 RPC 方法
- 返回交易哈希（txid）

**配置**：
```yaml
Bitcoin:
  URL: http://bitcoin:18332/wallet/mscoin
  Username: bitcoin
  Password: "123456"
  MinConfirmations: 0
  MaxConfirmations: 999999
  AddressType: legacy
  TimeoutMs: 20000
```

### 6.4 调用关系图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           服务调用关系图                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│                         ┌─────────────────┐                                │
│                         │   jobcenter     │                                │
│                         └────────┬────────┘                                │
│                                  │                                          │
│        ┌─────────────────────────┼─────────────────────────┐                │
│        │                         │                         │                │
│        │                         │                         │                │
│        ▼                         ▼                         ▼                │
│  ┌───────────┐            ┌───────────┐            ┌───────────┐           │
│  │ market-rpc│            │ucenter-rpc│            │Bitcoin Core│           │
│  └─────┬─────┘            └─────┬─────┘            └───────────┘           │
│        │                        │                                          │
│        │                        │                                          │
│        │  FindCoinById          │  FindWalletBySymbol                      │
│        │  (查询币种)            │  (查询钱包)                               │
│        │                        │                                          │
│        │  FindExchangeCoinVisible│                                         │
│        │  (查询交易对)          │                                          │
│        │                        │                                          │
│        ▼                        ▼                                          │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                         外部 API                                   │    │
│  │  ┌─────────────────┐    ┌─────────────────┐                       │    │
│  │  │    OKX API      │    │    Redis        │                       │    │
│  │  │  (汇率、K线)    │    │  (缓存、检查点) │                       │    │
│  │  └─────────────────┘    └─────────────────┘                       │    │
│  │  ┌─────────────────┐    ┌─────────────────┐                       │    │
│  │  │    MySQL        │    │    MongoDB      │                       │    │
│  │  │  (提现记录)     │    │  (K线数据)      │                       │    │
│  │  └─────────────────┘    └─────────────────┘                       │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 7. 配置说明

### 7.1 完整配置文件

```yaml
# 服务名称
Name: jobcenter

# 运行模式
Mode: dev

# 日志配置
Log:
  ServiceName: jobcenter
  Mode: console

# Kafka 消费者配置
Kafka:
  Brokers:
    - kafka:9092
  Topic: withdraw                    # 消费的 Topic
  GroupID: jobcenter-withdraw        # 消费者组 ID
  RetryBackoffMs: 2000               # 重试退避时间（毫秒）
  MaxWaitMs: 3000                    # 最大等待时间（毫秒）
  DeadLetterTopic: withdraw.dlq      # 死信队列 Topic
  AllowAutoTopicCreate: true         # 允许自动创建 Topic

# MySQL 配置（ucenter 数据库）
UcenterMysql:
  DataSource: root:root@tcp(mysql:3306)/ucenter?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
  MaxOpenConns: 50                   # 最大打开连接数
  MaxIdleConns: 10                   # 最大空闲连接数

# Redis 配置
Redis:
  Addrs:
    - redis:6379
  Password: ""
  DB: 0

# MongoDB 配置
Mongo:
  URI: mongodb://mongo:27017
  Username: root
  Password: root123456
  Database: mscoin

# Market RPC 配置
MarketRPC:
  Etcd:
    Hosts:
      - etcd:2379
    Key: market.rpc
  NonBlock: true                     # 非阻塞模式

# Ucenter RPC 配置
UcenterRPC:
  Etcd:
    Hosts:
      - etcd:2379
    Key: ucenter.rpc
  NonBlock: true

# OKX API 配置
OKX:
  APIKey: ""                         # API Key
  SecretKey: ""                      # Secret Key
  Passphrase: ""                     # Passphrase
  Host: "https://www.okx.com"        # API Host
  Proxy: ""                          # 代理地址
  TimeoutMs: 20000                   # 超时时间（毫秒）

# 定时任务配置
Tasks:
  # 汇率同步任务
  RateSync:
    Enabled: true                    # 是否启用
    RunOnStart: true                 # 启动时立即执行
    IntervalSeconds: 300             # 执行间隔（秒）

  # K线同步任务
  Klines:
    - Period: "1m"                   # 1 分钟 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 60
      PublishLatest: true            # 发布最新 K 线到 Kafka
      PublishTopic: "kline_1m"
    - Period: "3m"                   # 3 分钟 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 180
      PublishLatest: false
      PublishTopic: ""
    - Period: "5m"                   # 5 分钟 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 300
      PublishLatest: false
      PublishTopic: ""
    - Period: "15m"                  # 15 分钟 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 900
      PublishLatest: false
      PublishTopic: ""
    - Period: "30m"                  # 30 分钟 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 1800
      PublishLatest: false
      PublishTopic: ""
    - Period: "1H"                   # 1 小时 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 3600
      PublishLatest: false
      PublishTopic: ""
    - Period: "2H"                   # 2 小时 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 7200
      PublishLatest: false
      PublishTopic: ""
    - Period: "4H"                   # 4 小时 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 14400
      PublishLatest: false
      PublishTopic: ""
    - Period: "1D"                   # 1 天 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 86400
      PublishLatest: false
      PublishTopic: ""
    - Period: "1W"                   # 1 周 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 604800
      PublishLatest: false
      PublishTopic: ""
    - Period: "1M"                   # 1 月 K 线
      Enabled: true
      RunOnStart: true
      IntervalSeconds: 2592000
      PublishLatest: false
      PublishTopic: ""

# Bitcoin Core 配置
Bitcoin:
  URL: http://bitcoin:18332/wallet/mscoin  # Bitcoin Core RPC URL
  Username: bitcoin                        # RPC 用户名
  Password: "123456"                        # RPC 密码
  MinConfirmations: 0                      # 最小确认数
  MaxConfirmations: 999999                 # 最大确认数
  AddressType: legacy                      # 地址类型
  TimeoutMs: 20000                         # 超时时间（毫秒）
```

### 7.2 配置项详解

#### 7.2.1 Kafka 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Brokers | Kafka 集群地址列表 | - |
| Topic | 消费的 Topic 名称 | withdraw |
| GroupID | 消费者组 ID | jobcenter-withdraw |
| RetryBackoffMs | 重试退避时间 | 2000 |
| MaxWaitMs | 最大等待时间 | 3000 |
| DeadLetterTopic | 死信队列 Topic | withdraw.dlq |

#### 7.2.2 定时任务配置

| 配置项 | 说明 |
|--------|------|
| Enabled | 是否启用该任务 |
| RunOnStart | 服务启动时是否立即执行一次 |
| IntervalSeconds | 执行间隔（秒），<= 0 时使用默认值 60 |

#### 7.2.3 K线任务配置

| 配置项 | 说明 |
|--------|------|
| Period | K 线周期（1m、5m、1H、1D 等） |
| PublishLatest | 是否发布最新 K 线到 Kafka |
| PublishTopic | 发布目标 Topic 名称 |

#### 7.2.4 Bitcoin Core 配置

| 配置项 | 说明 |
|--------|------|
| URL | Bitcoin Core RPC URL（包含钱包路径） |
| Username | RPC 用户名 |
| Password | RPC 密码 |
| MinConfirmations | 最小确认数 |
| MaxConfirmations | 最大确认数 |
| AddressType | 地址类型（legacy、segwit 等） |
| TimeoutMs | 超时时间（毫秒） |

---

## 附录

### A. 常见问题排查

#### A.1 提现处理失败

**症状**：提现记录长时间处于 Processing 状态

**排查步骤**：
1. 检查 jobcenter 日志是否有错误
2. 检查 Kafka 消费者是否正常消费
3. 检查 Bitcoin Core 是否可连接
4. 检查死信队列是否有消息

#### A.2 K 线数据不同步

**症状**：前端 K 线数据不更新

**排查步骤**：
1. 检查定时任务是否启用
2. 检查 OKX API 是否可访问
3. 检查 MongoDB 连接是否正常
4. 检查 Market RPC 是否返回交易对列表

#### A.3 汇率不更新

**症状**：汇率显示异常

**排查步骤**：
1. 检查 rate-sync 任务是否启用
2. 检查 OKX API 是否可访问
3. 检查 Redis 连接是否正常

### B. 监控指标

建议监控以下指标：

| 指标 | 说明 |
|------|------|
| Kafka 消费延迟 | 消息积压情况 |
| 提现处理成功率 | 成功/总数 |
| 提现处理延迟 | 处理时间分布 |
| K 线同步成功率 | 成功/总数 |
| 汇率同步成功率 | 成功/总数 |
| Bitcoin Core 调用延迟 | RPC 响应时间 |

### C. 扩展开发

#### C.1 添加新的定时任务

1. 在 `config.go` 中添加任务配置结构
2. 在 `domain/service` 中实现领域服务
3. 在 `task/service.go` 的 `registerJobs` 中注册任务
4. 在配置文件中添加任务配置

#### C.2 添加新的 Kafka 消费者

1. 在 `model` 中定义消息结构
2. 在 `domain/service` 中实现处理逻辑
3. 在 `consumer` 中创建消费者
4. 在 `main.go` 中注册到 ServiceGroup

#### C.3 支持新的币种

1. 在 `withdraw_service.go` 中移除币种检查
2. 实现对应币种的转账逻辑
3. 添加相应的测试用例
