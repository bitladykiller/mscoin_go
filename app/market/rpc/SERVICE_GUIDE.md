# Market RPC 服务详解

## 1. 服务概述

### 1.1 功能定位

Market RPC 服务是 mscoin_go 项目的核心市场数据服务，为整个系统提供市场数据查询能力。该服务采用 gRPC 协议，支持高并发的读密集型查询场景。

**核心功能：**
- 币种信息查询（Coin 相关）：查询系统支持的币种配置、充值提现限制、费率设置等
- 交易对信息查询（ExchangeCoin 相关）：查询交易对配置、交易精度、交易限制等
- K 线历史数据查询：从 MongoDB 查询指定时间范围和周期的 K 线数据
- 法币汇率查询：提供 USD 对各法币的实时汇率查询

### 1.2 在整体架构中的位置

```
┌─────────────────────────────────────────────────────────────────────┐
│                         API Gateway / Load Balancer                  │
└─────────────────────────────────────────────────────────────────────┘
                                      │
              ┌───────────────────────┼───────────────────────┐
              ▼                       ▼                       ▼
    ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
    │  exchange API   │     │   ucenter API   │     │   jobcenter     │
    └─────────────────┘     └─────────────────┘     └─────────────────┘
              │                       │                       │
              ▼                       │                       ▼
    ┌─────────────────┐               │     ┌─────────────────────────┐
    │  exchange RPC   │───────────────┼────▶│    market RPC (本服务)  │
    └─────────────────┘               │     └─────────────────────────┘
              │                       │                   │
              │                       │                   ├──────────┬──────────┬──────────┐
              │                       │                   ▼          ▼          ▼          ▼
              │                       │              ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
              │                       │              │ MySQL  │ │MongoDB │ │ Redis  │ │ Etcd   │
              │                       │              │(币种/  │ │(K线)   │ │(汇率   │ │(服务   │
              │                       │              │交易对) │ │        │ │ 缓存)  │ │ 发现)  │
              │                       │              └────────┘ └────────┘ └────────┘ └────────┘
              ▼                       ▼
    ┌─────────────────┐     ┌─────────────────┐
    │   ucenter RPC   │◀────│   market RPC    │
    └─────────────────┘     └─────────────────┘
```

### 1.3 技术栈

- **框架**: go-zero (zrpc)
- **通信协议**: gRPC + Protobuf
- **服务发现**: Etcd
- **数据库**:
  - MySQL: 存储币种配置、交易对配置（元数据）
  - MongoDB: 存储 K 线历史数据（时序数据）
  - Redis: 缓存汇率数据

### 1.4 设计特点

1. **读密集型优化**: 所有方法均为查询方法，无写操作，便于水平扩展和缓存优化
2. **混合存储策略**: 
   - 元数据（币种、交易对）存储在 MySQL，便于事务和关系查询
   - K 线历史数据存储在 MongoDB，天然适合时序数据的追加写入和范围查询
   - 汇率数据缓存在 Redis，支持高频访问
3. **优雅降级**: 汇率查询在缓存不可用时使用硬编码回退值，确保服务始终可用
4. **分层架构**: 采用清晰的分层设计，便于测试和维护

---

## 2. 目录结构

```
/Volumes/移动卷宗/学习/go/mscoin_go/app/market/rpc/
├── main.go                      # 服务入口点，启动 gRPC 服务
├── etc/
│   └── market.yaml              # 配置文件（数据库连接、服务监听等）
├── pb/                          # Protobuf 生成的代码
│   ├── market/
│   │   ├── market.pb.go         # Market 消息定义
│   │   └── market_grpc.pb.go    # Market gRPC 服务端/客户端代码
│   └── rate/
│       ├── rate.pb.go           # Rate 消息定义
│       └── rate_grpc.pb.go      # Rate gRPC 服务端/客户端代码
└── internal/                    # 内部实现（不对外暴露）
    ├── config/
    │   └── config.go            # 配置结构定义
    ├── svc/
    │   └── service_context.go   # 服务上下文（依赖注入容器）
    ├── server/                  # Server 层：gRPC 服务端实现
    │   ├── market_server.go     # MarketServer 实现
    │   └── exchange_rate_server.go # ExchangeRateServer 实现
    ├── logic/                   # Logic 层：业务编排
    │   ├── base.go              # 公共基础结构
    │   ├── find_symbol_thumb_trend_logic.go
    │   ├── find_symbol_info_logic.go
    │   ├── find_coin_info_logic.go
    │   ├── find_all_coin_logic.go
    │   ├── find_coin_by_id_logic.go
    │   ├── find_exchange_coin_visible_logic.go
    │   ├── history_kline_logic.go
    │   └── usd_rate_logic.go
    ├── domain/                  # 领域层
    │   └── service/             # Domain Service：业务规则
    │       ├── coin_service.go
    │       ├── exchange_coin_service.go
    │       ├── market_service.go
    │       ├── rate_service.go
    │       └── rate_service_test.go
    ├── repository/              # Repository 层：数据访问
    │   ├── coin_repository.go
    │   ├── exchange_coin_repository.go
    │   └── kline_repository.go
    └── model/                   # Model 层：数据模型
        ├── coin.go
        ├── exchange_coin.go
        └── kline.go
```

---

## 3. 分层架构详解

### 3.1 架构总览

```
┌─────────────────────────────────────────────────────────────────────┐
│                           gRPC Request                              │
└─────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Server 层                                                           │
│  ┌─────────────────────┐    ┌─────────────────────────┐              │
│  │   MarketServer      │    │  ExchangeRateServer     │              │
│  │  (market_server.go) │    │(exchange_rate_server.go)│              │
│  └─────────────────────┘    └─────────────────────────┘              │
│  职责：实现 gRPC 接口，路由请求到 Logic 层                              │
└─────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Logic 层                                                            │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │  marketLogicBase (base.go)                                    │   │
│  │  ├── coinService                                              │   │
│  │  ├── exchangeCoinService                                      │   │
│  │  ├── marketService                                            │   │
│  │  └── rateService                                              │   │
│  └───────────────────────────────────────────────────────────────┘   │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐     │
│  │FindSymbol   │ │FindSymbol   │ │FindCoinInfo │ │HistoryKline │ ... │
│  │ThumbTrend   │ │Info         │ │             │ │             │     │
│  │Logic        │ │Logic        │ │Logic        │ │Logic        │     │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘     │
│  职责：业务编排，协调领域服务，处理超时和错误转换                          │
└─────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Domain Service 层                                                   │
│  ┌─────────────┐ ┌─────────────────┐ ┌───────────────┐ ┌───────────┐  │
│  │CoinService  │ │ExchangeCoin     │ │MarketService  │ │RateService│  │
│  │             │ │Service          │ │               │ │           │  │
│  └─────────────┘ └─────────────────┘ └───────────────┘ └───────────┘  │
│  职责：封装核心业务规则，协调 Repository 完成业务用例                     │
└─────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Repository 层                                                       │
│  ┌─────────────────┐ ┌─────────────────────┐ ┌───────────────────┐  │
│  │CoinRepository   │ │ExchangeCoin         │ │KlineRepository    │  │
│  │(MySQL)          │ │Repository (MySQL)   │ │(MongoDB)          │  │
│  └─────────────────┘ └─────────────────────┘ └─────────────────────┘  │
│  职责：封装数据库访问，隔离 SQL/NoSQL 细节                               │
└─────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Model 层                                                            │
│  ┌─────────┐ ┌──────────────┐ ┌─────────┐                            │
│  │  Coin   │ │ExchangeCoin  │ │ Kline   │                            │
│  └─────────┘ └──────────────┘ └─────────┘                            │
│  职责：定义数据实体结构，映射数据库表/集合                                 │
└─────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        数据库 (MySQL / MongoDB / Redis)               │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Server 层

Server 层是 gRPC 服务的入口，负责接收请求并路由到对应的 Logic 处理器。

#### 3.2.1 MarketServer (`market_server.go`)

**文件职责**: 实现市场数据查询的 gRPC 服务端，处理币种、交易对、K 线等市场数据查询请求。

**关键结构体**:

```go
// MarketServer 是市场领域请求的 RPC 门面
type MarketServer struct {
    svcCtx *svc.ServiceContext
    marketpb.UnimplementedMarketServer
}
```

**实现的方法**:

| 方法名 | 功能说明 | 调用的 Logic |
|--------|----------|--------------|
| `FindSymbolThumbTrend` | 获取所有可见交易对的缩略图和趋势数据 | `FindSymbolThumbTrendLogic` |
| `FindSymbolInfo` | 根据 symbol 获取交易对详情 | `FindSymbolInfoLogic` |
| `FindCoinInfo` | 根据 unit 获取币种详情 | `FindCoinInfoLogic` |
| `FindAllCoin` | 获取所有币种列表 | `FindAllCoinLogic` |
| `HistoryKline` | 获取 K 线历史数据 | `HistoryKlineLogic` |
| `FindExchangeCoinVisible` | 获取所有可见的交易对 | `FindExchangeCoinVisibleLogic` |
| `FindCoinById` | 根据 ID 获取币种详情 | `FindCoinByIDLogic` |

**设计说明**:
- Server 层不包含任何业务逻辑，仅作为传输层适配器
- 每个方法都是一行代码，将请求委托给对应的 Logic 处理器
- 这种设计使得 Server 层极其简洁，便于测试和维护

#### 3.2.2 ExchangeRateServer (`exchange_rate_server.go`)

**文件职责**: 实现法币汇率查询的 gRPC 服务端。

**关键结构体**:

```go
// ExchangeRateServer 是法币汇率查询的 RPC 门面
type ExchangeRateServer struct {
    svcCtx *svc.ServiceContext
    ratepb.UnimplementedExchangeRateServer
}
```

**实现的方法**:

| 方法名 | 功能说明 | 调用的 Logic |
|--------|----------|--------------|
| `UsdRate` | 获取 USD 对目标法币的汇率 | `UsdRateLogic` |

**设计说明**:
- 汇率数据优先从 Redis 缓存读取
- 缓存未命中或读取失败时使用硬编码的回退值
- 确保服务启动不依赖外部汇率数据

### 3.3 Logic 层

Logic 层负责业务编排，协调一个或多个领域服务完成业务用例，处理超时、错误转换等横切关注点。

#### 3.3.1 base.go - 公共基础结构

**文件职责**: 提供 Logic 层的公共依赖和工具方法。

**关键结构体**:

```go
// marketLogicBase 聚合 market RPC logic 文件中使用的领域服务
type marketLogicBase struct {
    ctx                 context.Context      // 请求上下文
    svcCtx              *svc.ServiceContext  // 服务上下文
    coinService         *service.CoinService         // 币种领域服务
    exchangeCoinService *service.ExchangeCoinService // 交易对领域服务
    marketService       *service.MarketService       // 市场数据领域服务
    rateService         *service.RateService         // 汇率领域服务
}
```

**设计说明**:
- 使用组合模式，每个具体 Logic 嵌入 `marketLogicBase`
- 从服务上下文中提取各领域服务，供具体 Logic 使用
- 这种依赖注入方式便于单元测试时 mock 领域服务

#### 3.3.2 find_symbol_thumb_trend_logic.go

**文件职责**: 处理 `FindSymbolThumbTrend` 请求，获取首页市场概览数据。

**业务用例**:
- 查询所有可见交易对
- 为每个交易对计算当日价格摘要和趋势线
- 用于前端首页展示

**核心逻辑**:

```go
func (l *FindSymbolThumbTrendLogic) FindSymbolThumbTrend(*marketpb.MarketReq) (*marketpb.SymbolThumbRes, error) {
    list, err := l.marketService.SymbolThumbTrend(l.ctx)
    if err != nil {
        return nil, err
    }
    return &marketpb.SymbolThumbRes{List: list}, nil
}
```

**返回数据包括**:
- Symbol: 交易对标识
- Open/High/Low/Close: 当日开盘、最高、最低、收盘价
- Chg: 涨跌幅百分比
- Change: 涨跌额
- Trend: 价格趋势线数据
- Volume/Turnover: 成交量和成交额

**注意事项**:
- 当 K 线数据缺失时，返回空的缩略图而不是错误
- 确保单个交易对数据问题不影响整体列表展示

#### 3.3.3 find_symbol_info_logic.go

**文件职责**: 处理 `FindSymbolInfo` 请求，根据 symbol 查询交易对详情。

**请求参数**:
- `req.Symbol`: 交易对标识，格式为 "BASEQUOTE"，如 "BTCUSDT"、"ETHUSDT"

**返回数据包括**:
- Symbol: 交易对标识
- BaseSymbol/CoinSymbol: 基础币种和计价币种
- CoinScale/BaseCoinScale: 价格和数量精度
- Fee: 交易手续费率
- Enable/Visible: 启用和可见状态
- MinTurnover/MinVolume: 最小成交额和最小交易量

**错误情况**:
- 交易对不存在时返回 "trading pair not found" 错误

#### 3.3.4 find_coin_info_logic.go

**文件职责**: 处理 `FindCoinInfo` 请求，根据 unit 查询币种详情。

**请求参数**:
- `req.Unit`: 币种单位标识，如 "BTC"、"ETH"、"USDT"

**返回数据包括**:
- Name/NameCN: 币种英文名和中文名
- Unit: 币种单位
- CanWithdraw/CanRecharge/CanTransfer: 提现、充值、转账开关
- MaxWithdrawAmount/MinWithdrawAmount: 提现限额
- MaxTxFee/MinTxFee: 手续费范围
- Status: 币种状态

**超时控制**:
- 设置 5 秒超时，防止数据库查询阻塞

**错误情况**:
- 币种不存在时返回 "not support this coin: {unit}" 错误

#### 3.3.5 find_all_coin_logic.go

**文件职责**: 处理 `FindAllCoin` 请求，获取系统所有币种列表。

**返回数据**:
- 所有币种的完整信息列表

**注意事项**:
- 返回所有币种，包括已禁用的
- 如果只需要启用的币种，调用方需要自行过滤

#### 3.3.6 find_coin_by_id_logic.go

**文件职责**: 处理 `FindCoinById` 请求，根据 ID 查询币种信息。

**请求参数**:
- `req.Id`: 币种 ID（数据库主键）

**与 FindCoinInfo 的区别**:
- 本方法通过 ID 查询，FindCoinInfo 通过 unit 查询
- 适用于不同业务场景（如订单查询后获取币种信息）

#### 3.3.7 find_exchange_coin_visible_logic.go

**文件职责**: 处理 `FindExchangeCoinVisible` 请求，获取所有可见的交易对。

**返回数据**:
- 只返回 visible=1 的交易对列表

**与 FindSymbolInfo 的区别**:
- 本方法返回列表，FindSymbolInfo 返回单个
- 本方法只返回 visible=1 的交易对

#### 3.3.8 history_kline_logic.go

**文件职责**: 处理 `HistoryKline` 请求，获取 K 线历史数据。

**请求参数**:
- `req.Symbol`: 交易对标识，如 "BTCUSDT"
- `req.From`: 开始时间（毫秒时间戳）
- `req.To`: 结束时间（毫秒时间戳）
- `req.Resolution`: K 线周期

**Resolution 支持的值**:
| 值 | 含义 |
|----|------|
| "1" | 1 分钟 |
| "5" | 5 分钟 |
| "15" | 15 分钟 |
| "30" | 30 分钟 |
| "1H" (默认) | 1 小时 |
| "1D" | 1 天 |
| "1W" | 1 周 |
| "1M" | 1 月 |

**超时控制**:
- 设置 10 秒超时，K 线数据可能较大

#### 3.3.9 usd_rate_logic.go

**文件职责**: 处理 `UsdRate` 请求，获取 USD 对目标法币的汇率。

**请求参数**:
- `req.Unit`: 目标法币代码，如 "CNY"、"JPY"

**数据来源优先级**:
1. Redis 缓存（由外部同步任务更新）
2. 硬编码的回退值（服务可用性保障）

**设计说明**:
- 本方法不会返回错误，即使缓存不可用也会返回回退值
- 确保汇率查询始终可用

### 3.4 Domain Service 层

Domain Service 层封装核心业务规则和业务逻辑，协调一个或多个 Repository 完成业务用例。

#### 3.4.1 coin_service.go

**文件职责**: 封装币种相关的业务规则。

**关键结构体**:

```go
type CoinService struct {
    repo *repository.CoinRepository
}
```

**提供的方法**:

| 方法名 | 功能说明 |
|--------|----------|
| `FindCoinInfo(ctx, unit string)` | 根据 unit 查询币种详情 |
| `FindCoinByID(ctx, id int64)` | 根据 ID 查询币种详情 |
| `FindAll(ctx)` | 获取所有币种列表 |

**业务规则**:
- 如果币种不存在，返回明确的业务错误
- unit 为币种单位标识，如 "BTC"、"ETH"、"USDT"

#### 3.4.2 exchange_coin_service.go

**文件职责**: 封装可见交易对相关的业务规则。

**关键结构体**:

```go
type ExchangeCoinService struct {
    repo *repository.ExchangeCoinRepository
}
```

**提供的方法**:

| 方法名 | 功能说明 |
|--------|----------|
| `FindVisible(ctx)` | 获取所有可见的交易对 |
| `FindBySymbol(ctx, symbol string)` | 根据 symbol 查询交易对详情 |

**业务规则**:
- 只返回 visible=1 的交易对
- symbol 格式为 "BASEQUOTE"，如 "BTCUSDT"、"ETHUSDT"

**调用关系**:
- MarketService 依赖本服务获取可见交易对列表

#### 3.4.3 market_service.go

**文件职责**: 协调读密集型的市场业务流程，结合交易对元数据和历史 K 线数据。

**关键结构体**:

```go
type MarketService struct {
    klineRepo           *repository.KlineRepository
    exchangeCoinService *ExchangeCoinService
}
```

**提供的方法**:

| 方法名 | 功能说明 |
|--------|----------|
| `SymbolThumbTrend(ctx)` | 计算每个可见交易对的当前快照列表 |
| `HistoryKline(ctx, symbol, from, to, resolution)` | 返回 K 线历史数据 |

**SymbolThumbTrend 核心逻辑**:

```go
func (s *MarketService) SymbolThumbTrend(ctx context.Context) ([]*marketpb.CoinThumb, error) {
    // 1. 获取所有可见交易对
    coins, err := s.exchangeCoinService.FindVisible(ctx)
    
    // 2. 为每个交易对查询当日 K 线数据
    for i, coin := range coins {
        klines, err := s.klineRepo.FindBySymbolTime(ctx, coin.Symbol, "1H", from, to, "")
        if err != nil || len(klines) == 0 {
            // K 线数据缺失时，回退为空的缩略图
            thumbs[i] = model.DefaultCoinThumb(coin.Symbol)
            continue
        }
        thumbs[i] = buildThumb(coin.Symbol, klines)
    }
    return thumbs, nil
}
```

**buildThumb 计算逻辑**:
- 取最新一根 K 线作为基准
- 计算涨跌：最新收盘价 - 当日第一根收盘价
- 计算涨跌幅：涨跌 / 第一根收盘价 * 100
- 遍历所有 K 线计算当日最高、最低、总成交量、总成交额
- 收集收盘价作为趋势线数据

**Resolution 到 Period 的映射**:

| 前端 Resolution | MongoDB Period |
|-----------------|----------------|
| "1" | "1m" |
| "5" | "5m" |
| "15" | "15m" |
| "30" | "30m" |
| "1D" | "1D" |
| "1W" | "1W" |
| "1M" | "1M" |
| 其他 | "1H" (默认) |

#### 3.4.4 rate_service.go

**文件职责**: 封装汇率查询规则。

**关键结构体**:

```go
type RateService struct {
    cache     rateCache          // Redis 缓存客户端
    fallbacks map[string]float64 // 硬编码的回退汇率表
}
```

**常量**:

```go
const usdtCNYRateCacheKey = "USDT::CNY::RATE"  // Redis 缓存键
```

**初始化的回退汇率**:
- CNY: 6.95
- JPY: 136.23

**USDRate 方法核心逻辑**:

```go
func (s *RateService) USDRate(ctx context.Context, unit string) float64 {
    normalized := strings.ToUpper(strings.TrimSpace(unit))
    if normalized == "CNY" && s.cache != nil {
        // 从 Redis 读取实时汇率
        var raw string
        if err := s.cache.GetCtx(ctx, usdtCNYRateCacheKey, &raw); err == nil {
            if value, parseErr := strconv.ParseFloat(strings.TrimSpace(raw), 64); parseErr == nil && value > 0 {
                return value
            }
        }
        // 缓存错误有意回退，确保服务可用
    }
    return s.fallbacks[normalized]
}
```

**为什么保留回退值**:
- 公开汇率接口应在异步同步任务暂时滞后时优雅降级
- 启动顺序不应导致 market-rpc 不可用
- 目前只有动态同步的货币需要读取缓存

#### 3.4.5 rate_service_test.go

**文件职责**: 测试 RateService 的各种场景。

**测试用例**:

| 测试名 | 场景说明 |
|--------|----------|
| `TestRateServiceReturnsRedisCNYRate` | Redis 返回有效汇率时使用该值 |
| `TestRateServiceFallsBackWhenRedisMisses` | Redis 缓存未命中时使用回退值 |
| `TestRateServiceFallsBackWhenRedisFails` | Redis 连接失败时使用回退值 |

**测试辅助结构**:

```go
type fakeRateCache struct {
    getFn func(ctx context.Context, key string, value any) error
}

func (f *fakeRateCache) GetCtx(ctx context.Context, key string, value any) error {
    return f.getFn(ctx, key, value)
}
```

### 3.5 Repository 层

Repository 层封装所有数据库访问逻辑，隔离 SQL/NoSQL 细节，对上层透明。

#### 3.5.1 coin_repository.go

**文件职责**: 封装对 `coin` 表的直接 SQL 访问。

**数据源**: MySQL
**表名**: coin

**关键结构体**:

```go
type CoinRepository struct {
    db *sqlx.DB
}
```

**提供的方法**:

| 方法名 | SQL 说明 |
|--------|----------|
| `FindByUnit(ctx, unit string)` | `SELECT * FROM coin WHERE unit=? LIMIT 1` |
| `FindByID(ctx, id int64)` | `SELECT * FROM coin WHERE id=? LIMIT 1` |
| `FindAll(ctx)` | `SELECT * FROM coin` |

**设计说明**:
- 记录不存在时返回 nil, nil（而非错误）
- 便于上层判断"不存在"和"数据库错误"

#### 3.5.2 exchange_coin_repository.go

**文件职责**: 封装对 `exchange_coin` 表的 SQL 访问。

**数据源**: MySQL
**表名**: exchange_coin

**关键结构体**:

```go
type ExchangeCoinRepository struct {
    db *sqlx.DB
}
```

**提供的方法**:

| 方法名 | SQL 说明 |
|--------|----------|
| `FindVisible(ctx)` | `SELECT * FROM exchange_coin WHERE visible=1` |
| `FindBySymbol(ctx, symbol string)` | `SELECT * FROM exchange_coin WHERE symbol=? LIMIT 1` |

#### 3.5.3 kline_repository.go

**文件职责**: 封装对 MongoDB K 线历史数据的访问。

**数据源**: MongoDB
**集合命名规则**: `exchange_kline_{symbol}_{period}`
- symbol: 交易对标识，如 "BTCUSDT"
- period: K 线周期，如 "1H"、"1D"、"15m"

**为什么使用 MongoDB**:
- K 线历史是追加密集型数据（只写入，极少更新）
- 按币种/时间范围查询导向的访问模式
- 不属于必须留在 MySQL 中的核心事务状态
- 天然适合时序数据存储

**关键结构体**:

```go
type KlineRepository struct {
    db *mongo.Database
}
```

**提供的方法**:

| 方法名 | 功能说明 |
|--------|----------|
| `FindBySymbolTime(ctx, symbol, period, from, to, sortOrder)` | 加载指定时间范围内的 K 线数据 |

**查询规则**:
- 集合名由 symbol 和 period 动态计算
- 时间范围查询：`time >= from AND time <= to`
- 支持升序或降序排序
  - "asc": 按时间升序（用于 K 线图表展示）
  - 其他: 按时间降序（用于计算最新数据）

### 3.6 Model 层

Model 层定义数据实体结构，映射数据库表/集合结构。

#### 3.6.1 coin.go

**文件职责**: 定义 `coin` 表的实体结构。

**关键结构体**:

```go
type Coin struct {
    ID                int     `db:"id" gorm:"column:id"`
    Name              string  `db:"name" gorm:"column:name"`
    CanAutoWithdraw   int     `db:"can_auto_withdraw" gorm:"column:can_auto_withdraw"`
    CanRecharge       int     `db:"can_recharge" gorm:"column:can_recharge"`
    CanTransfer       int     `db:"can_transfer" gorm:"column:can_transfer"`
    CanWithdraw       int     `db:"can_withdraw" gorm:"column:can_withdraw"`
    CNYRate           float64 `db:"cny_rate" gorm:"column:cny_rate"`
    EnableRPC         int     `db:"enable_rpc" gorm:"column:enable_rpc"`
    IsPlatformCoin    int     `db:"is_platform_coin" gorm:"column:is_platform_coin"`
    MaxTxFee          float64 `db:"max_tx_fee" gorm:"column:max_tx_fee"`
    MaxWithdrawAmount float64 `db:"max_withdraw_amount" gorm:"column:max_withdraw_amount"`
    MinTxFee          float64 `db:"min_tx_fee" gorm:"column:min_tx_fee"`
    MinWithdrawAmount float64 `db:"min_withdraw_amount" gorm:"column:min_withdraw_amount"`
    NameCN            string  `db:"name_cn" gorm:"column:name_cn"`
    Sort              int     `db:"sort" gorm:"column:sort"`
    Status            int     `db:"status" gorm:"column:status"`
    Unit              string  `db:"unit" gorm:"column:unit"`
    USDTRate          float64 `db:"usd_rate" gorm:"column:usd_rate"`
    WithdrawThreshold float64 `db:"withdraw_threshold" gorm:"column:withdraw_threshold"`
    HasLegal          int     `db:"has_legal" gorm:"column:has_legal"`
    ColdWalletAddress string  `db:"cold_wallet_address" gorm:"column:cold_wallet_address"`
    MinerFee          float64 `db:"miner_fee" gorm:"column:miner_fee"`
    WithdrawScale     int     `db:"withdraw_scale" gorm:"column:withdraw_scale"`
    AccountType       int     `db:"account_type" gorm:"column:account_type"`
    DepositAddress    string  `db:"deposit_address" gorm:"column:deposit_address"`
    InfoLink          string  `db:"infolink" gorm:"column:infolink"`
    Information       string  `db:"information" gorm:"column:information"`
    MinRechargeAmount float64 `db:"min_recharge_amount" gorm:"column:min_recharge_amount"`
}
```

**字段说明**:

| 字段名 | 类型 | 业务含义 |
|--------|------|----------|
| ID | int | 币种唯一标识（数据库主键） |
| Name | string | 币种英文名 |
| NameCN | string | 币种中文名 |
| Unit | string | 币种单位标识（如 "BTC"、"ETH"） |
| CanWithdraw | int | 提现开关（1=启用） |
| CanRecharge | int | 充值开关（1=启用） |
| CanTransfer | int | 转账开关（1=启用） |
| CanAutoWithdraw | int | 自动提现开关（1=启用） |
| MaxWithdrawAmount | float64 | 单次提现金额上限 |
| MinWithdrawAmount | float64 | 单次提现金额下限 |
| MaxTxFee | float64 | 手续费上限 |
| MinTxFee | float64 | 手续费下限 |
| WithdrawThreshold | float64 | 提现阈值（触发审核） |
| MinerFee | float64 | 矿工费 |
| WithdrawScale | int | 提现精度（小数位数） |
| CNYRate | float64 | 对 CNY 的汇率 |
| USDTRate | float64 | 对 USDT 的汇率 |
| Status | int | 币种状态 |
| Sort | int | 排序权重 |
| EnableRPC | int | RPC 调用开关 |
| IsPlatformCoin | int | 是否为平台币 |
| HasLegal | int | 是否有法律合规要求 |
| AccountType | int | 账户类型 |
| ColdWalletAddress | string | 冷钱包地址 |
| DepositAddress | string | 充值地址 |
| InfoLink | string | 币种介绍链接 |
| Information | string | 币种描述 |
| MinRechargeAmount | float64 | 最小充值金额 |

#### 3.6.2 exchange_coin.go

**文件职责**: 定义 `exchange_coin` 表的实体结构。

**关键结构体**:

```go
type ExchangeCoin struct {
    ID               int64   `db:"id" gorm:"column:id"`
    Symbol           string  `db:"symbol" gorm:"column:symbol"`
    BaseCoinScale    int64   `db:"base_coin_scale" gorm:"column:base_coin_scale"`
    BaseSymbol       string  `db:"base_symbol" gorm:"column:base_symbol"`
    CoinScale        int64   `db:"coin_scale" gorm:"column:coin_scale"`
    CoinSymbol       string  `db:"coin_symbol" gorm:"column:coin_symbol"`
    Enable           int64   `db:"enable" gorm:"column:enable"`
    Fee              float64 `db:"fee" gorm:"column:fee"`
    Sort             int64   `db:"sort" gorm:"column:sort"`
    EnableMarketBuy  int64   `db:"enable_market_buy" gorm:"column:enable_market_buy"`
    EnableMarketSell int64   `db:"enable_market_sell" gorm:"column:enable_market_sell"`
    MinSellPrice     float64 `db:"min_sell_price" gorm:"column:min_sell_price"`
    Flag             int64   `db:"flag" gorm:"column:flag"`
    MaxTradingOrder  int64   `db:"max_trading_order" gorm:"column:max_trading_order"`
    MaxTradingTime   int64   `db:"max_trading_time" gorm:"column:max_trading_time"`
    MinTurnover      float64 `db:"min_turnover" gorm:"column:min_turnover"`
    ClearTime        int64   `db:"clear_time" gorm:"column:clear_time"`
    EndTime          int64   `db:"end_time" gorm:"column:end_time"`
    Exchangeable     int64   `db:"exchangeable" gorm:"column:exchangeable"`
    MaxBuyPrice      float64 `db:"max_buy_price" gorm:"column:max_buy_price"`
    MaxVolume        float64 `db:"max_volume" gorm:"column:max_volume"`
    MinVolume        float64 `db:"min_volume" gorm:"column:min_volume"`
    PublishAmount    float64 `db:"publish_amount" gorm:"column:publish_amount"`
    PublishPrice     float64 `db:"publish_price" gorm:"column:publish_price"`
    PublishType      int64   `db:"publish_type" gorm:"column:publish_type"`
    RobotType        int64   `db:"robot_type" gorm:"column:robot_type"`
    StartTime        int64   `db:"start_time" gorm:"column:start_time"`
    Visible          int64   `db:"visible" gorm:"column:visible"`
    Zone             int64   `db:"zone" gorm:"column:zone"`
}
```

**字段说明**:

| 字段名 | 类型 | 业务含义 |
|--------|------|----------|
| ID | int64 | 交易对唯一标识 |
| Symbol | string | 交易对标识（如 "BTCUSDT"） |
| BaseSymbol | string | 基础币种单位（如 "BTC"） |
| CoinSymbol | string | 计价币种单位（如 "USDT"） |
| BaseCoinScale | int64 | 基础币种精度（价格精度） |
| CoinScale | int64 | 计价币种精度（数量精度） |
| Enable | int64 | 是否启用交易 |
| Visible | int64 | 是否对用户可见 |
| Fee | float64 | 交易手续费率 |
| Sort | int64 | 排序权重 |
| EnableMarketBuy | int64 | 是否支持市价买入 |
| EnableMarketSell | int64 | 是否支持市价卖出 |
| MinSellPrice | float64 | 卖出最低价限制 |
| MaxBuyPrice | float64 | 买入最高价限制 |
| MinVolume | float64 | 最小交易量 |
| MaxVolume | float64 | 最大交易量 |
| MinTurnover | float64 | 最小成交额 |
| MaxTradingOrder | int64 | 最大挂单数量 |
| MaxTradingTime | int64 | 最大挂单时长（秒） |
| Flag | int64 | 特殊标记位 |
| Zone | int64 | 分区标识 |
| PublishType | int64 | 上币类型 |
| PublishAmount | float64 | 发行量 |
| PublishPrice | float64 | 发行价 |
| StartTime | int64 | 上币活动开始时间 |
| EndTime | int64 | 上币活动结束时间 |
| ClearTime | int64 | 清算时间 |
| Exchangeable | int64 | 是否可交易 |
| RobotType | int64 | 机器人类型 |

#### 3.6.3 kline.go

**文件职责**: 定义 MongoDB K 线文档的结构。

**关键结构体**:

```go
type Kline struct {
    Period       string  `bson:"period,omitempty"`
    OpenPrice    float64 `bson:"openPrice,omitempty"`
    HighestPrice float64 `bson:"highestPrice,omitempty"`
    LowestPrice  float64 `bson:"lowestPrice,omitempty"`
    ClosePrice   float64 `bson:"closePrice,omitempty"`
    Time         int64   `bson:"time,omitempty"`
    Count        float64 `bson:"count,omitempty"`
    Volume       float64 `bson:"volume,omitempty"`
    Turnover     float64 `bson:"turnover,omitempty"`
}
```

**字段说明**:

| 字段名 | 类型 | 业务含义 |
|--------|------|----------|
| Period | string | K 线周期标识（如 "1H"、"1D"） |
| OpenPrice | float64 | 开盘价 |
| HighestPrice | float64 | 最高价 |
| LowestPrice | float64 | 最低价 |
| ClosePrice | float64 | 收盘价 |
| Time | int64 | K 线时间（毫秒时间戳） |
| Count | float64 | 成交笔数 |
| Volume | float64 | 成交量（以基础币种计） |
| Turnover | float64 | 成交额（以计价币种计） |

**方法**:

| 方法名 | 功能说明 |
|--------|----------|
| `TableName(symbol, period string)` | 推导 MongoDB 集合名称 |
| `ToCoinThumb(symbol string, oldest *Kline)` | 转换为缩略图数据 |
| `DefaultCoinThumb(symbol string)` | 返回空的缩略图数据 |

**TableName 计算逻辑**:

```go
func (k *Kline) TableName(symbol, period string) string {
    return "exchange_kline_" + symbol + "_" + period
}
// 例如：exchange_kline_BTCUSDT_1H
```

**ToCoinThumb 计算逻辑**:

```go
func (k *Kline) ToCoinThumb(symbol string, oldest *Kline) *marketpb.CoinThumb {
    change := k.ClosePrice - oldest.ClosePrice
    chg := 0.0
    if oldest.ClosePrice != 0 {
        chg = change / oldest.ClosePrice * 100
    }
    return &marketpb.CoinThumb{
        Symbol:   symbol,
        Open:     k.OpenPrice,
        High:     k.HighestPrice,
        Low:      k.LowestPrice,
        Close:    k.ClosePrice,
        Chg:      chg,
        Change:   change,
        UsdRate:  k.ClosePrice,
        BaseUsdRate: 1,
        DateTime: k.Time,
    }
}
```

---

## 4. gRPC 服务定义

### 4.1 Market 服务 (`market.proto`)

**服务定义**:

```protobuf
service Market {
  rpc FindSymbolThumbTrend(MarketReq) returns(SymbolThumbRes);
  rpc FindSymbolInfo(MarketReq) returns(ExchangeCoin);
  rpc FindCoinInfo(MarketReq) returns(Coin);
  rpc FindAllCoin(MarketReq) returns(CoinList);
  rpc HistoryKline(MarketReq) returns(HistoryRes);
  rpc FindExchangeCoinVisible(MarketReq) returns(ExchangeCoinRes);
  rpc FindCoinById(MarketReq) returns(Coin);
}
```

**请求消息 MarketReq**:

```protobuf
message MarketReq {
  string ip = 1;         // 客户端 IP（可选）
  string symbol = 2;     // 交易对标识（如 "BTCUSDT"）
  string unit = 3;       // 币种单位（如 "BTC"）
  int64 from = 4;        // 开始时间（毫秒时间戳）
  int64 to = 5;          // 结束时间（毫秒时间戳）
  string resolution = 6; // K 线周期
  int64 id = 7;          // 币种 ID
}
```

**响应消息**:

| 消息名 | 字段 | 说明 |
|--------|------|------|
| SymbolThumbRes | list: CoinThumb[] | 交易对缩略图列表 |
| ExchangeCoin | 见模型层 | 交易对详情 |
| Coin | 见模型层 | 币种详情 |
| CoinList | list: Coin[], total: int64 | 币种列表 |
| HistoryRes | list: History[] | K 线历史数据 |
| ExchangeCoinRes | list: ExchangeCoin[] | 交易对列表 |

**CoinThumb 消息**:

```protobuf
message CoinThumb {
  string symbol = 1;           // 交易对标识
  double open = 2;             // 开盘价
  double high = 3;             // 最高价
  double low = 4;              // 最低价
  double close = 5;            // 收盘价
  double chg = 6;              // 涨跌幅百分比
  double change = 7;           // 涨跌额
  double volume = 8;           // 成交量
  double turnover = 9;         // 成交额
  double lastDayClose = 10;    // 昨日收盘价
  double usdRate = 11;         // USD 汇率
  double baseUsdRate = 12;     // 基础币 USD 汇率
  double zone = 13;            // 分区
  int64 dateTime = 14;         // 时间戳
  repeated double trend = 15;  // 趋势线数据
}
```

**History 消息**:

```protobuf
message History {
  int64 time = 1;    // 时间戳
  double open = 2;   // 开盘价
  double close = 3;  // 收盘价
  double high = 4;   // 最高价
  double low = 5;    // 最低价
  double volume = 6; // 成交量
}
```

### 4.2 ExchangeRate 服务 (`rate.proto`)

**服务定义**:

```protobuf
service ExchangeRate {
  rpc usdRate(RateReq) returns(RateRes);
}
```

**请求消息 RateReq**:

```protobuf
message RateReq {
  string unit = 1; // 目标法币代码（如 "CNY"、"JPY"）
  string ip = 2;  // 客户端 IP（可选）
}
```

**响应消息 RateRes**:

```protobuf
message RateRes {
  double rate = 1; // USD 对目标法币的汇率
}
```

---

## 5. 数据库设计

### 5.1 MySQL 表

#### 5.1.1 coin 表

**用途**: 存储币种配置信息。

**主要字段**:

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | INT | PRIMARY KEY, AUTO_INCREMENT | 币种唯一标识 |
| name | VARCHAR(50) | NOT NULL | 币种英文名 |
| name_cn | VARCHAR(50) | | 币种中文名 |
| unit | VARCHAR(20) | UNIQUE, NOT NULL | 币种单位标识 |
| can_withdraw | TINYINT | DEFAULT 0 | 提现开关 |
| can_recharge | TINYINT | DEFAULT 0 | 充值开关 |
| can_transfer | TINYINT | DEFAULT 0 | 转账开关 |
| can_auto_withdraw | TINYINT | DEFAULT 0 | 自动提现开关 |
| max_withdraw_amount | DECIMAL(20,8) | | 单次提现上限 |
| min_withdraw_amount | DECIMAL(20,8) | | 单次提现下限 |
| max_tx_fee | DECIMAL(20,8) | | 手续费上限 |
| min_tx_fee | DECIMAL(20,8) | | 手续费下限 |
| withdraw_threshold | DECIMAL(20,8) | | 提现阈值 |
| miner_fee | DECIMAL(20,8) | | 矿工费 |
| withdraw_scale | INT | | 提现精度 |
| cny_rate | DECIMAL(20,8) | | CNY 汇率 |
| usd_rate | DECIMAL(20,8) | | USDT 汇率 |
| status | TINYINT | DEFAULT 1 | 币种状态 |
| sort | INT | DEFAULT 0 | 排序权重 |
| enable_rpc | TINYINT | DEFAULT 0 | RPC 开关 |
| is_platform_coin | TINYINT | DEFAULT 0 | 是否平台币 |
| has_legal | TINYINT | DEFAULT 0 | 是否有法律要求 |
| account_type | INT | | 账户类型 |
| cold_wallet_address | VARCHAR(200) | | 冷钱包地址 |
| deposit_address | VARCHAR(200) | | 充值地址 |
| infolink | VARCHAR(500) | | 介绍链接 |
| information | TEXT | | 币种描述 |
| min_recharge_amount | DECIMAL(20,8) | | 最小充值金额 |

#### 5.1.2 exchange_coin 表

**用途**: 存储交易对配置信息。

**主要字段**:

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 交易对唯一标识 |
| symbol | VARCHAR(20) | UNIQUE, NOT NULL | 交易对标识 |
| base_symbol | VARCHAR(20) | NOT NULL | 基础币种单位 |
| coin_symbol | VARCHAR(20) | NOT NULL | 计价币种单位 |
| base_coin_scale | INT | | 基础币种精度 |
| coin_scale | INT | | 计价币种精度 |
| enable | TINYINT | DEFAULT 1 | 是否启用 |
| visible | TINYINT | DEFAULT 1 | 是否可见 |
| fee | DECIMAL(10,4) | | 交易手续费率 |
| sort | INT | DEFAULT 0 | 排序权重 |
| enable_market_buy | TINYINT | DEFAULT 0 | 支持市价买入 |
| enable_market_sell | TINYINT | DEFAULT 0 | 支持市价卖出 |
| min_sell_price | DECIMAL(20,8) | | 卖出最低价 |
| max_buy_price | DECIMAL(20,8) | | 买入最高价 |
| min_volume | DECIMAL(20,8) | | 最小交易量 |
| max_volume | DECIMAL(20,8) | | 最大交易量 |
| min_turnover | DECIMAL(20,8) | | 最小成交额 |
| max_trading_order | INT | | 最大挂单数量 |
| max_trading_time | INT | | 最大挂单时长 |
| flag | INT | | 特殊标记 |
| zone | INT | | 分区标识 |
| publish_type | INT | | 上币类型 |
| publish_amount | DECIMAL(20,8) | | 发行量 |
| publish_price | DECIMAL(20,8) | | 发行价 |
| start_time | BIGINT | | 活动开始时间 |
| end_time | BIGINT | | 活动结束时间 |
| clear_time | BIGINT | | 清算时间 |
| exchangeable | TINYINT | | 是否可交易 |
| robot_type | INT | | 机器人类型 |

### 5.2 MongoDB 集合

#### 5.2.1 exchange_kline_{symbol}_{period} 集合

**命名规则**: `exchange_kline_{symbol}_{period}`
- symbol: 交易对标识，如 "BTCUSDT"
- period: K 线周期，如 "1H"、"1D"、"15m"

**示例集合名**: `exchange_kline_BTCUSDT_1H`

**文档结构**:

| 字段名 | 类型 | 索引 | 说明 |
|--------|------|------|------|
| time | int64 | 主键/索引 | K 线时间（毫秒时间戳） |
| period | string | | K 线周期 |
| openPrice | double | | 开盘价 |
| highestPrice | double | | 最高价 |
| lowestPrice | double | | 最低价 |
| closePrice | double | | 收盘价 |
| count | double | | 成交笔数 |
| volume | double | | 成交量 |
| turnover | double | | 成交额 |

**数据特点**:
- 追加密集型数据（只写入，极少更新）
- 按时间范围查询为主
- 天然适合时序数据存储

### 5.3 Redis 缓存

#### 5.3.1 汇率缓存

**缓存键**: `USDT::CNY::RATE`

**数据类型**: String

**数据内容**: USDT 对 CNY 的汇率值（字符串形式）

**更新方式**: 由外部同步任务定期更新

**有效期**: 无固定有效期，依赖外部同步任务

---

## 6. 与其他服务的调用关系

### 6.1 服务依赖图

```
                    ┌─────────────────────┐
                    │   exchange API      │
                    └─────────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │   exchange RPC      │◀──────────────────────┐
                    └─────────────────────┘                       │
                              │                                   │
                              │ FindSymbolInfo()                  │
                              ▼                                   │
                    ┌─────────────────────┐                       │
                    │   market RPC        │───────────────────────┘
                    │   (本服务)           │
                    └─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
  ┌───────────┐         ┌───────────┐         ┌───────────┐
  │  MySQL    │         │  MongoDB  │         │  Redis    │
  │(币种/交易对)│         │  (K线)    │         │ (汇率缓存) │
  └───────────┘         └───────────┘         └───────────┘
```

### 6.2 被调用关系

**exchange RPC 对 market RPC 的调用**:

| 调用方 | 调用方法 | 用途说明 |
|--------|----------|----------|
| exchange RPC | `MarketClient.FindSymbolInfo()` | 下单时查询交易对配置，验证交易对是否可交易 |

**调用场景详解**:

1. **下单验证（exchange RPC -> market RPC）**:
   - 调用 `FindSymbolInfo()` 查询交易对配置
   - 验证交易对是否存在且启用
   - 获取交易精度、手续费率、价格限制等配置
   - 验证订单价格是否在允许范围内（MinSellPrice ~ MaxBuyPrice）
   - 验证订单数量是否在允许范围内（MinVolume ~ MaxVolume）

### 6.3 不调用其他服务

Market RPC 是一个纯查询服务，不调用其他 RPC 服务。所有数据都来自本地数据库（MySQL、MongoDB、Redis）。

---

## 7. 配置说明

### 7.1 配置文件 (`etc/market.yaml`)

```yaml
# 服务名称
Name: market.rpc

# 监听地址（0.0.0.0 表示监听所有网卡）
ListenOn: 0.0.0.0:8082

# 服务发现配置（Etcd）
Etcd:
  Hosts:
    - etcd:2379      # Etcd 地址
  Key: market.rpc    # 服务注册键

# MySQL 配置
Mysql:
  DataSource: root:root@tcp(mysql:3306)/market?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
  MaxOpenConns: 100  # 最大连接数
  MaxIdleConns: 20   # 最大空闲连接数

# MongoDB 配置
Mongo:
  URI: mongodb://mongo:27017
  Username: root
  Password: root123456
  Database: mscoin

# Redis 配置
Redis:
  Addrs:
    - redis:6379
  Password: ""
  DB: 0
```

### 7.2 配置结构体 (`internal/config/config.go`)

```go
type Config struct {
    zrpc.RpcServerConf           // go-zero RPC 服务基础配置
    Mysql mysqlx.Config           // MySQL 配置
    Mongo mongox.Config           // MongoDB 配置
    Redis redisx.Config           // Redis 配置
}
```

### 7.3 配置项详解

#### 7.3.1 服务基础配置 (RpcServerConf)

| 配置项 | 说明 |
|--------|------|
| Name | 服务名称，用于日志和监控 |
| ListenOn | 监听地址，格式为 `ip:port` |
| Etcd.Hosts | Etcd 集群地址列表 |
| Etcd.Key | 服务注册键，用于服务发现 |

#### 7.3.2 MySQL 配置

| 配置项 | 说明 |
|--------|------|
| DataSource | 数据源连接字符串 |
| MaxOpenConns | 最大打开连接数 |
| MaxIdleConns | 最大空闲连接数 |

**DataSource 格式**: `{username}:{password}@tcp({host}:{port})/{database}?charset={charset}&parseTime=true&loc={timezone}`

#### 7.3.3 MongoDB 配置

| 配置项 | 说明 |
|--------|------|
| URI | MongoDB 连接 URI |
| Username | 用户名 |
| Password | 密码 |
| Database | 数据库名 |

#### 7.3.4 Redis 配置

| 配置项 | 说明 |
|--------|------|
| Addrs | Redis 地址列表（支持集群） |
| Password | 密码（如有） |
| DB | 数据库编号 |

---

## 8. 服务启动流程

### 8.1 启动入口 (`main.go`)

```go
func main() {
    // 1. 解析命令行参数
    flag.Parse()

    // 2. 加载配置文件
    var c config.Config
    conf.MustLoad(*configFile, &c)

    // 3. 初始化服务上下文（数据库连接、领域服务等）
    ctx := svc.NewServiceContext(c)
    defer ctx.Close()

    // 4. 创建 gRPC 服务器
    s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
        // 注册 MarketServer
        marketpb.RegisterMarketServer(grpcServer, server.NewMarketServer(ctx))
        // 注册 ExchangeRateServer
        ratepb.RegisterExchangeRateServer(grpcServer, server.NewExchangeRateServer(ctx))

        // 开发/测试模式下启用 gRPC 反射
        if c.Mode == service.DevMode || c.Mode == service.TestMode {
            reflection.Register(grpcServer)
        }
    })
    defer s.Stop()

    // 5. 启动服务监听
    fmt.Printf("Starting market rpc server at %s...\n", c.ListenOn)
    s.Start()
}
```

### 8.2 服务上下文初始化 (`internal/svc/service_context.go`)

```go
func NewServiceContext(c config.Config) *ServiceContext {
    // 1. 连接 MySQL
    db, err := mysqlx.New(c.Mysql)
    if err != nil {
        log.Fatalf("init mysql: %v", err)
    }

    // 2. 连接 MongoDB
    mongoClient, err := mongox.New(c.Mongo)
    if err != nil {
        log.Fatalf("init mongo: %v", err)
    }

    // 3. 创建 Repository 实例
    coinRepo := repository.NewCoinRepository(db)
    exchangeCoinRepo := repository.NewExchangeCoinRepository(db)
    klineRepo := repository.NewKlineRepository(mongoClient.Database())

    // 4. 创建 Redis 缓存客户端
    cache := redisx.New(c.Redis)

    // 5. 创建 Domain Service 实例
    exchangeCoinService := service.NewExchangeCoinService(exchangeCoinRepo)

    return &ServiceContext{
        Config: c,
        DB:     db,
        Mongo:  mongoClient,
        Cache:  cache,

        CoinService:         service.NewCoinService(coinRepo),
        ExchangeCoinService: exchangeCoinService,
        MarketService:       service.NewMarketService(klineRepo, exchangeCoinService),
        RateService:         service.NewRateService(cache),
    }
}
```

### 8.3 依赖初始化顺序

```
1. 基础设施层（MySQL、MongoDB、Redis）
        │
        ▼
2. Repository 层（CoinRepository、ExchangeCoinRepository、KlineRepository）
        │
        ▼
3. Domain Service 层（CoinService、ExchangeCoinService、MarketService、RateService）
        │
        ▼
4. ServiceContext（聚合所有依赖）
```

---

## 9. 常见问题与最佳实践

### 9.1 如何调试 gRPC 服务

使用 grpcurl 工具：

```bash
# 列出所有服务
grpcurl -plaintext localhost:8082 list

# 列出 Market 服务的所有方法
grpcurl -plaintext localhost:8082 list market.Market

# 调用 FindSymbolInfo 方法
grpcurl -plaintext -d '{"symbol":"BTCUSDT"}' localhost:8082 market.Market/FindSymbolInfo

# 调用 FindCoinInfo 方法
grpcurl -plaintext -d '{"unit":"BTC"}' localhost:8082 market.Market/FindCoinInfo

# 调用 HistoryKline 方法
grpcurl -plaintext -d '{"symbol":"BTCUSDT","from":1704067200000,"to":1704153600000,"resolution":"1H"}' localhost:8082 market.Market/HistoryKline
```

### 9.2 性能优化建议

1. **K 线查询优化**:
   - 合理设置查询时间范围，避免一次性查询过多数据
   - 前端应做分页或增量加载

2. **汇率查询优化**:
   - 汇率数据已有 Redis 缓存，无需额外优化
   - 客户端可缓存汇率，避免频繁请求

3. **交易对列表优化**:
   - FindSymbolThumbTrend 会查询所有可见交易对并计算趋势
   - 如需更高性能，可考虑增加缓存层

### 9.3 扩展指南

**添加新的查询方法**:

1. 在 `idl/rpc/market/market.proto` 中定义新的 RPC 方法
2. 执行 `protoc` 生成代码
3. 在 `internal/logic/` 下创建对应的 Logic 文件
4. 在 `internal/server/market_server.go` 中添加路由

**添加新的数据源**:

1. 在 `internal/model/` 下定义数据模型
2. 在 `internal/repository/` 下实现 Repository
3. 在 `internal/domain/service/` 下实现 Domain Service
4. 在 `internal/svc/service_context.go` 中注册依赖

---

## 10. 总结

Market RPC 服务是一个设计良好的读密集型市场数据服务，具有以下特点：

1. **清晰的分层架构**: Server -> Logic -> Domain Service -> Repository -> Model，每层职责明确
2. **混合存储策略**: MySQL 存储元数据，MongoDB 存储时序数据，Redis 缓存热点数据
3. **优雅降级设计**: 汇率查询在缓存不可用时使用回退值，确保服务可用性
4. **详细的代码注释**: 每个文件都有详细的注释说明，便于理解和维护
5. **完善的测试**: RateService 有完整的单元测试覆盖

该服务作为系统的市场数据核心，为交易系统提供必要的币种、交易对和市场行情数据支撑。
