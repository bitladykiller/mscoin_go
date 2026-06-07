# Exchange-RPC 服务详细文档

## 目录

1. [服务概述](#1-服务概述)
2. [目录结构](#2-目录结构)
3. [分层架构详解](#3-分层架构详解)
4. [每个文件的详细说明](#4-每个文件的详细说明)
5. [gRPC 服务定义](#5-grpc-服务定义)
6. [数据库设计](#6-数据库设计)
7. [与其他服务的调用关系](#7-与其他服务的调用关系)
8. [配置说明](#8-配置说明)

---

## 1. 服务概述

### 1.1 服务定位

`exchange-rpc` 是 mscoin 交易系统的核心服务之一，负责处理所有与交易订单相关的业务逻辑。该服务采用 gRPC 协议对外提供接口，支持高性能的订单创建、查询、取消等操作。

### 1.2 核心功能

| 功能 | 描述 |
|------|------|
| 订单创建 | 支持限价单和市价单的创建，包含完整的业务规则验证 |
| 订单查询 | 支持历史订单查询和当前委托订单查询 |
| 订单详情 | 根据订单 ID 查询订单详细信息 |
| 订单取消 | 取消正在委托中的订单 |

### 1.3 在整体架构中的位置

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   exchange-api (HTTP)                        │
│              (对外提供 RESTful 接口)                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  exchange-rpc (gRPC)  <-- 本文档             │
│              (订单核心业务逻辑处理)                            │
└─────────────────────────────────────────────────────────────┘
            │                           │                   │
            ▼                           ▼                   ▼
┌───────────────────┐    ┌───────────────────┐    ┌───────────────────┐
│    ucenter-rpc    │    │    market-rpc     │    │      MySQL        │
│  (用户中心服务)    │    │   (行情服务)       │    │   (订单数据库)     │
└───────────────────┘    └───────────────────┘    └───────────────────┘
```

### 1.4 技术栈

- **框架**: go-zero (微服务框架)
- **通信协议**: gRPC
- **数据库**: MySQL (通过 sqlx 访问)
- **缓存**: Redis
- **服务发现**: Etcd
- **序列化**: Protocol Buffers

---

## 2. 目录结构

```
/Volumes/移动卷宗/学习/go/mscoin_go/app/exchange/rpc/
├── main.go                    # 服务入口文件
├── Dockerfile                 # Docker 构建文件
├── etc/                       # 配置文件目录
│   └── exchange.yaml          # 服务配置文件
├── internal/                  # 内部代码目录（不对外暴露）
│   ├── config/               # 配置结构定义
│   │   └── config.go         # Config 结构体定义
│   ├── model/                # 数据模型层
│   │   └── order.go          # 订单实体和状态常量定义
│   ├── repository/           # 数据仓库层
│   │   └── order_repository.go # 订单数据库操作
│   ├── domain/               # 领域层
│   │   └── service/          # 领域服务
│   │       └── order_service.go # 订单业务逻辑
│   ├── logic/                # 业务逻辑层
│   │   ├── base.go           # Logic 基类
│   │   ├── add_logic.go      # 新增订单逻辑
│   │   ├── find_order_history_logic.go    # 历史订单查询逻辑
│   │   ├── find_order_current_logic.go    # 当前订单查询逻辑
│   │   ├── find_by_order_id_logic.go      # 订单详情查询逻辑
│   │   └── cancel_order_logic.go          # 取消订单逻辑
│   ├── server/               # gRPC 服务器层
│   │   └── order_server.go   # Order gRPC 服务实现
│   └── svc/                  # 服务上下文
│       └── service_context.go # ServiceContext 依赖容器
└── pb/                       # Protocol Buffers 定义
    └── order/
        ├── order.pb.go       # Protobuf 消息定义
        └── order_grpc.pb.go  # gRPC 服务接口定义
```

### 2.1 目录职责说明

| 目录/文件 | 职责 |
|-----------|------|
| `main.go` | 服务启动入口，初始化 gRPC 服务器 |
| `etc/` | 存放配置文件，支持不同环境配置 |
| `internal/config/` | 定义配置结构体，映射配置文件 |
| `internal/model/` | 定义数据模型、实体结构和常量 |
| `internal/repository/` | 数据访问层，封装数据库操作 |
| `internal/domain/service/` | 领域服务层，封装核心业务规则 |
| `internal/logic/` | RPC 方法对应的业务逻辑处理 |
| `internal/server/` | gRPC 服务器实现，路由请求到 Logic |
| `internal/svc/` | 服务上下文，管理依赖注入 |
| `pb/` | Protocol Buffers 生成的代码 |

---

## 3. 分层架构详解

### 3.1 架构分层图

```
┌─────────────────────────────────────────────────────────────┐
│                    Server Layer                              │
│              (order_server.go)                               │
│        gRPC 服务入口，路由请求到 Logic                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Logic Layer                               │
│    (add_logic.go, find_order_*.go, cancel_order_logic.go)   │
│        处理 RPC 请求，协调外部服务和领域服务                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Domain Service Layer                         │
│               (order_service.go)                             │
│        封装核心业务规则，可被多个 Logic 共享                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                           │
│             (order_repository.go)                            │
│        数据访问层，封装 SQL 操作                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Model Layer                              │
│                   (order.go)                                 │
│        数据实体定义、状态/方向/类型常量                         │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 各层职责详解

#### 3.2.1 Server 层

**文件**: `internal/server/order_server.go`

**职责**:
- 实现 gRPC 服务接口 (`orderpb.OrderServer`)
- 接收 RPC 请求，创建对应的 Logic 实例
- 将请求路由到 Logic 层处理
- 返回处理结果

**设计特点**:
- 每个 RPC 方法对应一个 Logic 结构体
- 通过 ServiceContext 访问所有依赖
- 在开发/测试模式启用 gRPC 反射服务

#### 3.2.2 Logic 层

**文件**: `internal/logic/*.go`

**职责**:
- 处理具体的 RPC 请求
- 调用外部 RPC 服务（ucenter-rpc、market-rpc）
- 调用领域服务进行业务规则验证
- 组装响应数据

**设计特点**:
- 每个 RPC 方法一个 Logic 结构体
- 继承 `exchangeLogicBase` 基类
- 通过 ServiceContext 访问 RPC 客户端和领域服务

#### 3.2.3 Domain Service 层

**文件**: `internal/domain/service/order_service.go`

**职责**:
- 封装核心业务规则
- 不依赖外部 RPC 服务，便于测试
- 可被多个 Logic 共享复用

**设计特点**:
- 纯业务逻辑，不涉及基础设施
- 接收外部服务查询结果作为参数
- 返回业务规则验证结果

#### 3.2.4 Repository 层

**文件**: `internal/repository/order_repository.go`

**职责**:
- 封装所有数据库操作
- 使用 sqlx 进行类型安全的 SQL 查询
- 不包含业务逻辑

**设计特点**:
- 一个 Repository 对应一个数据库表
- 所有方法接收 context.Context 参数
- 错误使用 fmt.Errorf 包装

#### 3.2.5 Model 层

**文件**: `internal/model/order.go`

**职责**:
- 定义数据库实体结构体
- 定义视图对象（VO）
- 定义状态、方向、类型常量

**设计特点**:
- 使用 `db` tag 映射数据库字段
- 提供 `ToView()` 方法转换展示格式
- 提供标签到代码的转换函数

---

## 4. 每个文件的详细说明

### 4.1 main.go - 服务入口

**文件路径**: `/app/exchange/rpc/main.go`

**职责**: 服务的启动入口，负责初始化和启动 gRPC 服务器。

**核心流程**:

```go
func main() {
    // 1. 解析命令行参数
    flag.Parse()

    // 2. 加载配置文件
    var c config.Config
    conf.MustLoad(*configFile, &c)

    // 3. 初始化服务上下文
    ctx := svc.NewServiceContext(c)
    defer ctx.Close()

    // 4. 创建 gRPC 服务器
    s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
        orderpb.RegisterOrderServer(grpcServer, server.NewOrderServer(ctx))
        // 开发/测试模式启用反射
        if c.Mode == service.DevMode || c.Mode == service.TestMode {
            reflection.Register(grpcServer)
        }
    })
    defer s.Stop()

    // 5. 启动服务
    fmt.Printf("Starting exchange rpc server at %s...\n", c.ListenOn)
    s.Start()
}
```

**关键点**:
- 使用 `flag.String` 定义配置文件路径参数
- `ServiceContext` 作为依赖容器管理所有依赖
- gRPC 反射服务便于使用 `grpcurl` 等工具调试

---

### 4.2 config/config.go - 配置定义

**文件路径**: `/app/exchange/rpc/internal/config/config.go`

**职责**: 定义服务配置结构体。

**Config 结构体**:

```go
type Config struct {
    zrpc.RpcServerConf          // RPC 服务器配置（ListenOn, Etcd 等）

    Mysql      mysqlx.Config    // MySQL 数据库配置
    Redis      redisx.Config    // Redis 缓存配置
    UcenterRPC zrpc.RpcClientConf // ucenter-rpc 客户端配置
    MarketRPC  zrpc.RpcClientConf // market-rpc 客户端配置
}
```

**配置项说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `ListenOn` | string | 服务监听地址 |
| `Etcd` | struct | 服务发现配置 |
| `Mysql.DataSource` | string | MySQL 连接字符串 |
| `Mysql.MaxOpenConns` | int | 最大打开连接数 |
| `Mysql.MaxIdleConns` | int | 最大空闲连接数 |
| `UcenterRPC` | struct | ucenter-rpc 服务发现配置 |
| `MarketRPC` | struct | market-rpc 服务发现配置 |
| `Redis` | struct | Redis 连接配置 |

---

### 4.3 svc/service_context.go - 服务上下文

**文件路径**: `/app/exchange/rpc/internal/svc/service_context.go`

**职责**: 作为依赖注入容器，管理所有运行时依赖。

**ServiceContext 结构体**:

```go
type ServiceContext struct {
    Config       config.Config              // 服务配置
    DB           *sqlx.DB                   // MySQL 连接
    Cache        *redisx.Client             // Redis 缓存
    OrderService *service.OrderService      // 订单领域服务

    // RPC 客户端
    MemberClient memberpb.MemberClient      // ucenter-rpc 会员客户端
    AssetClient  assetpb.AssetClient        // ucenter-rpc 资产客户端
    MarketClient marketpb.MarketClient      // market-rpc 市场客户端
}
```

**初始化流程**:

```go
func NewServiceContext(c config.Config) *ServiceContext {
    // 1. 创建 MySQL 连接
    db, err := mysqlx.New(c.Mysql)
    if err != nil {
        panic(err)
    }

    // 2. 创建 RPC 客户端
    ucClient := zrpc.MustNewClient(c.UcenterRPC)
    marketClient := zrpc.MustNewClient(c.MarketRPC)

    // 3. 创建订单仓库
    orderRepo := repository.NewOrderRepository(db)

    // 4. 创建领域服务
    orderService := service.NewOrderService(orderRepo)

    return &ServiceContext{...}
}
```

**设计特点**:
- 集中管理所有依赖，便于测试时 Mock
- 使用依赖注入，降低组件耦合度

---

### 4.4 model/order.go - 数据模型

**文件路径**: `/app/exchange/rpc/internal/model/order.go`

**职责**: 定义订单实体、视图对象和常量。

#### 4.4.1 订单状态常量

```go
const (
    OrderTrading   = iota  // 0: 交易中，订单正在撮合队列中等待成交
    OrderCompleted         // 1: 已完成，订单已全部成交
    OrderCanceled          // 2: 已取消，订单被用户取消
    OrderOverTimed         // 3: 已超时，订单超过有效期被系统取消
    OrderInit              // 4: 初始化，订单刚创建尚未进入撮合队列
)

var StatusLabels = map[int]string{
    OrderTrading:   "TRADING",
    OrderCompleted: "COMPLETED",
    OrderCanceled:  "CANCELED",
    OrderOverTimed: "OVERTIMED",
}
```

#### 4.4.2 订单方向常量

```go
const (
    DirectionBuy  = iota  // 0: 买入
    DirectionSell         // 1: 卖出
)

var DirectionLabels = map[int]string{
    DirectionBuy:  "BUY",
    DirectionSell: "SELL",
}
```

#### 4.4.3 订单类型常量

```go
const (
    TypeMarketPrice = iota  // 0: 市价单，按市场当前最优价格立即成交
    TypeLimitPrice          // 1: 限价单，按指定价格挂单等待成交
)

var TypeLabels = map[int]string{
    TypeMarketPrice: "MARKET_PRICE",
    TypeLimitPrice:  "LIMIT_PRICE",
}
```

#### 4.4.4 ExchangeOrder 实体

```go
type ExchangeOrder struct {
    ID            int64   `db:"id"`              // 数据库自增主键
    OrderId       string  `db:"order_id"`        // 业务订单号，格式为 "E" + 时间戳纳秒
    Amount        float64 `db:"amount"`          // 订单总数量
    BaseSymbol    string  `db:"base_symbol"`     // 基础币种符号，如 USDT
    CanceledTime  int64   `db:"canceled_time"`   // 订单取消时间戳（毫秒）
    CoinSymbol    string  `db:"coin_symbol"`     // 交易币种符号，如 BTC
    CompletedTime int64   `db:"completed_time"`  // 订单完成时间戳（毫秒）
    Direction     int     `db:"direction"`       // 订单方向（0:买入, 1:卖出）
    MemberId      int64   `db:"member_id"`       // 会员 ID
    Price         float64 `db:"price"`           // 订单价格，市价单为 0
    Status        int     `db:"status"`          // 订单状态
    Symbol        string  `db:"symbol"`          // 交易对符号，如 "BTCUSDT"
    Time          int64   `db:"time"`            // 订单创建时间戳（毫秒）
    TradedAmount  float64 `db:"traded_amount"`   // 已成交数量
    Turnover      float64 `db:"turnover"`        // 已成交金额
    Type          int     `db:"type"`            // 订单类型（0:市价, 1:限价）
    UseDiscount   string  `db:"use_discount"`    // 使用的折扣金额
}
```

#### 4.4.5 OrderView 视图对象

```go
type OrderView struct {
    OrderId       string  `json:"orderId"`       // 业务订单号
    Amount        float64 `json:"amount"`         // 订单总数量
    BaseSymbol    string  `json:"baseSymbol"`     // 基础币种符号
    CanceledTime  int64   `json:"canceledTime"`   // 订单取消时间戳
    CoinSymbol    string  `json:"coinSymbol"`     // 交易币种符号
    CompletedTime int64   `json:"completedTime"`  // 订单完成时间戳
    Direction     string  `json:"direction"`      // 订单方向标签（"BUY" 或 "SELL"）
    MemberId      int64   `json:"memberId"`       // 会员 ID
    Price         float64 `json:"price"`          // 订单价格
    Status        string  `json:"status"`         // 订单状态标签
    Symbol        string  `json:"symbol"`         // 交易对符号
    Time          int64   `json:"time"`           // 订单创建时间戳
    TradedAmount  float64 `json:"tradedAmount"`   // 已成交数量
    Turnover      float64 `json:"turnover"`       // 已成交金额
    Type          string  `json:"type"`           // 订单类型标签
    UseDiscount   string  `json:"useDiscount"`    // 使用的折扣金额
}
```

**ExchangeOrder 与 OrderView 的区别**:
- `ExchangeOrder`: 使用数值型状态/方向/类型，便于数据库存储和内部处理
- `OrderView`: 使用字符串标签，便于前端直接展示，无需再做转换

---

### 4.5 repository/order_repository.go - 订单仓库

**文件路径**: `/app/exchange/rpc/internal/repository/order_repository.go`

**职责**: 封装所有订单相关的数据库操作。

#### 4.5.1 Save - 保存订单

```go
func (r *OrderRepository) Save(ctx context.Context, order *model.ExchangeOrder) error
```

**功能**: 将新订单插入 `exchange_order` 表。

**插入字段**: order_id, amount, base_symbol, coin_symbol, direction, member_id, price, status, symbol, time, type, traded_amount, turnover, use_discount, canceled_time, completed_time

**使用场景**: 创建新订单时调用。

#### 4.5.2 FindOrderHistory - 查询历史订单

```go
func (r *OrderRepository) FindOrderHistory(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.ExchangeOrder, int64, error)
```

**功能**: 查询用户的历史订单列表。

**查询条件**:
- symbol: 交易对符号
- member_id: 会员 ID

**分页参数**:
- page: 页码，从 1 开始
- size: 每页记录数

**返回值**:
- list: 订单实体列表
- total: 符合条件的订单总数

#### 4.5.3 FindOrderCurrent - 查询当前委托订单

```go
func (r *OrderRepository) FindOrderCurrent(ctx context.Context, symbol string, page int64, size int64, memberID int64) ([]*model.ExchangeOrder, int64, error)
```

**功能**: 查询用户当前正在委托中的订单列表。

**查询条件**:
- symbol: 交易对符号
- member_id: 会员 ID
- status: 固定为 OrderTrading（交易中）

**返回值**: 仅返回 TRADING 状态的订单。

#### 4.5.4 FindCurrentTradingCount - 查询委托订单数量

```go
func (r *OrderRepository) FindCurrentTradingCount(ctx context.Context, memberID int64, symbol string, direction int) (int64, error)
```

**功能**: 查询用户当前正在交易中的订单数量。

**使用场景**: 下单前检查用户是否超过最大委托订单数限制。

**查询条件**:
- member_id: 会员 ID
- symbol: 交易对符号
- direction: 订单方向
- status: OrderTrading

#### 4.5.5 FindByOrderID - 根据订单 ID 查询

```go
func (r *OrderRepository) FindByOrderID(ctx context.Context, orderID string) (*model.ExchangeOrder, error)
```

**功能**: 根据业务订单号查询订单。

**返回值**:
- 订单存在：返回订单实体
- 订单不存在：返回 nil（不返回错误）

#### 4.5.6 UpdateStatus - 更新订单状态

```go
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status int) error
```

**功能**: 更新订单状态。

**使用场景**:
- 取消订单：将状态从 TRADING 更新为 CANCELED
- 订单完成：将状态从 TRADING 更新为 COMPLETED
- 订单超时：将状态从 TRADING 更新为 OVERTIMED

---

### 4.6 domain/service/order_service.go - 订单领域服务

**文件路径**: `/app/exchange/rpc/internal/domain/service/order_service.go`

**职责**: 封装交易订单的核心业务规则。

#### 4.6.1 ValidateAddRequest - 验证下单请求

```go
func (s *OrderService) ValidateAddRequest(
    reqDirection string,           // 订单方向标签
    reqType string,                // 订单类型标签
    price float64,                 // 订单价格
    amount float64,                // 订单数量
    memberInfo *memberpb.MemberInfo,     // 会员信息
    exchangeCoin *marketpb.ExchangeCoin, // 交易对配置
    baseWallet *assetpb.MemberWallet,    // 基础币种钱包
    coinWallet *assetpb.MemberWallet,    // 交易币种钱包
) error
```

**验证规则**:

| 序号 | 验证项 | 验证规则 | 错误信息 |
|------|--------|----------|----------|
| 1 | 用户交易状态 | memberInfo.transactionStatus == 1 | "this user is forbidden to trade" |
| 2 | 限价单价格 | 限价单必须 price > 0 | "limit price mode requires price > 0" |
| 3 | 订单数量 | amount > 0 | "amount must be > 0" |
| 4 | 交易对状态 | exchangeable == 1 && enable == 1 | "coin forbidden" |
| 5 | 钱包状态 | baseWallet.isLock == 0 && coinWallet.isLock == 0 | "wallet locked" |
| 6 | 卖出价格下限 | price >= exchangeCoin.minSellPrice | "price must be >= {minSellPrice}" |
| 7 | 买入价格上限 | price <= exchangeCoin.maxBuyPrice | "price must be <= {maxBuyPrice}" |
| 8 | 市价买入支持 | enableMarketBuy == 1 | "market buy is not supported" |
| 9 | 市价卖出支持 | enableMarketSell == 1 | "market sell is not supported" |

#### 4.6.2 BuildNewOrder - 构建新订单实体

```go
func (s *OrderService) BuildNewOrder(
    memberID int64,      // 会员 ID
    symbol string,       // 交易对符号
    baseSymbol string,   // 基础币种符号
    coinSymbol string,   // 交易币种符号
    reqType string,      // 订单类型标签
    reqDirection string, // 订单方向标签
    price float64,       // 订单价格
    amount float64,      // 订单数量
) *model.ExchangeOrder
```

**订单构建规则**:

| 字段 | 构建规则 |
|------|----------|
| OrderId | "E" + time.Now().UnixNano() |
| Status | OrderInit（初始化状态） |
| Time | time.Now().UnixMilli() |
| TradedAmount | 0 |
| Turnover | 0 |
| UseDiscount | "0" |
| Price | 市价单为 0，限价单为传入价格 |

---

### 4.7 logic/add_logic.go - 新增订单逻辑

**文件路径**: `/app/exchange/rpc/internal/logic/add_logic.go`

**职责**: 处理新增订单的 RPC 请求。

#### 订单创建完整流程

```
┌────────────────────────────────────────────────────────────────────────────┐
│                           Add 方法处理流程                                   │
└────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 1. 查询会员信息                                                               │
│    调用: MemberClient.FindMemberById(memberId)                               │
│    目的: 验证用户是否存在、交易状态是否正常                                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 2. 查询交易对信息                                                              │
│    调用: MarketClient.FindSymbolInfo(symbol)                                 │
│    目的: 验证交易对是否可交易、获取价格限制和配置信息                              │
│    返回: baseSymbol, coinSymbol, minSellPrice, maxBuyPrice, maxTradingOrder  │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 3. 查询用户钱包信息                                                            │
│    调用: AssetClient.FindWalletBySymbol(userId, baseSymbol)                  │
│    调用: AssetClient.FindWalletBySymbol(userId, coinSymbol)                  │
│    目的: 验证用户是否有足够的资产、钱包是否被锁定                                  │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 4. 验证订单请求参数                                                            │
│    调用: OrderService.ValidateAddRequest(...)                                │
│    验证: 用户状态、价格、数量、交易对状态、钱包状态、价格范围、市价支持             │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 5. 检查委托订单数量                                                            │
│    调用: OrderService.FindCurrentTradingCount(...)                           │
│    验证: 用户当前委托订单数量是否超过 maxTradingOrder 限制                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 6. 构建新订单实体                                                              │
│    调用: OrderService.BuildNewOrder(...)                                     │
│    生成: 订单 ID、设置订单属性、初始化状态                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 7. 保存订单到数据库                                                            │
│    调用: OrderService.AddOrder(order)                                        │
│    操作: 将订单实体持久化到 exchange_order 表                                   │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 8. 返回订单 ID                                                                │
│    响应: AddOrderRes{OrderId: order.OrderId}                                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

### 4.8 logic/find_order_history_logic.go - 历史订单查询逻辑

**文件路径**: `/app/exchange/rpc/internal/logic/find_order_history_logic.go`

**职责**: 处理查询历史订单的 RPC 请求。

**处理流程**:
1. 调用 `OrderService.FindOrderHistory` 查询数据库
2. 将订单实体列表转换为 RPC 响应格式
3. 返回分页结果

**返回的订单状态**:
- COMPLETED: 已完成（订单全部成交）
- CANCELED: 已取消（用户主动取消）
- OVERTIMED: 已超时（订单超过有效期被系统取消）

---

### 4.9 logic/find_order_current_logic.go - 当前订单查询逻辑

**文件路径**: `/app/exchange/rpc/internal/logic/find_order_current_logic.go`

**职责**: 处理查询当前委托订单的 RPC 请求。

**处理流程**:
1. 调用 `OrderService.FindOrderCurrent` 查询数据库
2. 将订单实体列表转换为 RPC 响应格式
3. 返回分页结果

**返回的订单状态**:
- TRADING: 交易中（订单正在撮合队列中等待成交）

---

### 4.10 logic/find_by_order_id_logic.go - 订单详情查询逻辑

**文件路径**: `/app/exchange/rpc/internal/logic/find_by_order_id_logic.go`

**职责**: 处理根据订单 ID 查询订单的 RPC 请求。

**处理流程**:
1. 调用 `OrderService.FindByOrderID` 查询数据库
2. 将订单实体转换为 `ExchangeOrderOrigin` 格式
3. 返回订单原始信息

**与其他查询方法的区别**:

| 方法 | 返回类型 | 状态/方向/类型 |
|------|----------|----------------|
| FindOrderHistory | ExchangeOrder | 字符串标签 |
| FindOrderCurrent | ExchangeOrder | 字符串标签 |
| FindByOrderId | ExchangeOrderOrigin | 数值代码 |

---

### 4.11 logic/cancel_order_logic.go - 取消订单逻辑

**文件路径**: `/app/exchange/rpc/internal/logic/cancel_order_logic.go`

**职责**: 处理取消订单的 RPC 请求。

**处理流程**:
1. 调用 `OrderService.CancelOrder` 更新订单状态
2. 返回取消结果（订单 ID）

**取消订单说明**:
- 将订单状态从 TRADING 更新为 CANCELED
- 取消后订单不再参与撮合
- 取消后的订单会出现在历史订单列表中

**注意事项**:
- 当前实现不调用 ucenter-rpc 释放冻结资金
- 实际应用中需要调用资产服务解冻资金

---

### 4.12 server/order_server.go - gRPC 服务器

**文件路径**: `/app/exchange/rpc/internal/server/order_server.go`

**职责**: 实现 Order 服务的 gRPC 接口。

**OrderServer 结构体**:

```go
type OrderServer struct {
    svcCtx *svc.ServiceContext
    orderpb.UnimplementedOrderServer
}
```

**RPC 方法映射**:

| RPC 方法 | Logic 结构体 | 处理方法 |
|----------|--------------|----------|
| FindOrderHistory | FindOrderHistoryLogic | FindOrderHistory(req) |
| FindOrderCurrent | FindOrderCurrentLogic | FindOrderCurrent(req) |
| Add | AddLogic | Add(req) |
| FindByOrderId | FindByOrderIDLogic | FindByOrderID(req) |
| CancelOrder | CancelOrderLogic | CancelOrder(req) |

---

## 5. gRPC 服务定义

### 5.1 Proto 文件结构

**消息定义**:

```protobuf
// 订单请求消息
message OrderReq {
    string ip = 1;              // 客户端 IP
    string symbol = 2;          // 交易对符号
    int64 page = 4;             // 页码
    int64 pageSize = 5;         // 每页大小
    int64 userId = 6;           // 用户 ID
    double price = 7;           // 订单价格
    double amount = 8;          // 订单数量
    string direction = 9;       // 订单方向（BUY/SELL）
    string type = 10;           // 订单类型（MARKET_PRICE/LIMIT_PRICE）
    int32 useDiscount = 11;     // 是否使用折扣
    string orderId = 12;        // 订单 ID
    int32 updateStatus = 13;    // 更新状态
}

// 订单列表响应
message OrderRes {
    repeated ExchangeOrder list = 1;  // 订单列表
    int64 total = 2;                 // 总数
}

// 订单信息（字符串标签）
message ExchangeOrder {
    int64 id = 1;
    string orderId = 2;
    double amount = 3;
    string baseSymbol = 4;
    int64 canceledTime = 5;
    string coinSymbol = 6;
    int64 completedTime = 7;
    string direction = 8;        // 字符串标签
    int64 memberId = 11;
    double price = 12;
    string status = 13;          // 字符串标签
    string symbol = 14;
    int64 time = 15;
    double tradedAmount = 16;
    double turnover = 17;
    string type = 18;            // 字符串标签
    string useDiscount = 21;
}

// 新增订单响应
message AddOrderRes {
    string orderId = 1;          // 新创建的订单 ID
}

// 取消订单响应
message CancelOrderRes {
    string orderId = 1;          // 被取消的订单 ID
}

// 订单原始信息（数值代码）
message ExchangeOrderOrigin {
    int64 id = 1;
    string orderId = 2;
    double amount = 3;
    string baseSymbol = 4;
    int64 canceledTime = 5;
    string coinSymbol = 6;
    int64 completedTime = 7;
    int32 direction = 8;         // 数值代码
    int64 memberId = 11;
    double price = 12;
    int32 status = 13;           // 数值代码
    string symbol = 14;
    int64 time = 15;
    double tradedAmount = 16;
    double turnover = 17;
    int32 type = 18;             // 数值代码
    string useDiscount = 21;
}
```

**服务定义**:

```protobuf
service Order {
    // 查询历史订单
    rpc FindOrderHistory(OrderReq) returns (OrderRes);

    // 查询当前委托订单
    rpc FindOrderCurrent(OrderReq) returns (OrderRes);

    // 新增订单
    rpc Add(OrderReq) returns (AddOrderRes);

    // 根据订单 ID 查询
    rpc FindByOrderId(OrderReq) returns (ExchangeOrderOrigin);

    // 取消订单
    rpc CancelOrder(OrderReq) returns (CancelOrderRes);
}
```

### 5.2 客户端接口

```go
type OrderClient interface {
    FindOrderHistory(ctx context.Context, in *OrderReq, opts ...grpc.CallOption) (*OrderRes, error)
    FindOrderCurrent(ctx context.Context, in *OrderReq, opts ...grpc.CallOption) (*OrderRes, error)
    Add(ctx context.Context, in *OrderReq, opts ...grpc.CallOption) (*AddOrderRes, error)
    FindByOrderId(ctx context.Context, in *OrderReq, opts ...grpc.CallOption) (*ExchangeOrderOrigin, error)
    CancelOrder(ctx context.Context, in *OrderReq, opts ...grpc.CallOption) (*CancelOrderRes, error)
}
```

---

## 6. 数据库设计

### 6.1 exchange_order 表结构

```sql
CREATE TABLE `exchange_order` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '数据库自增主键',
    `order_id` VARCHAR(64) NOT NULL COMMENT '业务订单号，格式为 E+时间戳纳秒',
    `amount` DECIMAL(20,8) NOT NULL COMMENT '订单总数量',
    `base_symbol` VARCHAR(20) NOT NULL COMMENT '基础币种符号，如 USDT',
    `canceled_time` BIGINT DEFAULT 0 COMMENT '订单取消时间戳（毫秒）',
    `coin_symbol` VARCHAR(20) NOT NULL COMMENT '交易币种符号，如 BTC',
    `completed_time` BIGINT DEFAULT 0 COMMENT '订单完成时间戳（毫秒）',
    `direction` TINYINT NOT NULL COMMENT '订单方向（0:买入, 1:卖出）',
    `member_id` BIGINT NOT NULL COMMENT '会员 ID',
    `price` DECIMAL(20,8) DEFAULT 0 COMMENT '订单价格，市价单为 0',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '订单状态（0:交易中, 1:已完成, 2:已取消, 3:已超时, 4:初始化）',
    `symbol` VARCHAR(30) NOT NULL COMMENT '交易对符号，如 BTCUSDT',
    `time` BIGINT NOT NULL COMMENT '订单创建时间戳（毫秒）',
    `traded_amount` DECIMAL(20,8) DEFAULT 0 COMMENT '已成交数量',
    `turnover` DECIMAL(20,8) DEFAULT 0 COMMENT '已成交金额',
    `type` TINYINT NOT NULL COMMENT '订单类型（0:市价, 1:限价）',
    `use_discount` VARCHAR(50) DEFAULT '0' COMMENT '使用的折扣金额',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_id` (`order_id`),
    KEY `idx_member_symbol` (`member_id`, `symbol`),
    KEY `idx_member_status` (`member_id`, `status`),
    KEY `idx_time` (`time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易订单表';
```

### 6.2 字段详细说明

| 字段 | 类型 | 说明 | 取值范围 |
|------|------|------|----------|
| id | BIGINT | 自增主键 | - |
| order_id | VARCHAR(64) | 业务订单号 | E + 纳秒时间戳 |
| amount | DECIMAL(20,8) | 订单总数量 | > 0 |
| base_symbol | VARCHAR(20) | 基础币种符号 | 如 USDT |
| canceled_time | BIGINT | 取消时间戳 | 毫秒时间戳 |
| coin_symbol | VARCHAR(20) | 交易币种符号 | 如 BTC |
| completed_time | BIGINT | 完成时间戳 | 毫秒时间戳 |
| direction | TINYINT | 订单方向 | 0-买入, 1-卖出 |
| member_id | BIGINT | 会员 ID | - |
| price | DECIMAL(20,8) | 订单价格 | >= 0 |
| status | TINYINT | 订单状态 | 0-交易中, 1-已完成, 2-已取消, 3-已超时, 4-初始化 |
| symbol | VARCHAR(30) | 交易对符号 | 如 BTCUSDT |
| time | BIGINT | 创建时间戳 | 毫秒时间戳 |
| traded_amount | DECIMAL(20,8) | 已成交数量 | >= 0 |
| turnover | DECIMAL(20,8) | 已成交金额 | >= 0 |
| type | TINYINT | 订单类型 | 0-市价, 1-限价 |
| use_discount | VARCHAR(50) | 折扣使用 | - |

### 6.3 索引说明

| 索引名 | 类型 | 字段 | 用途 |
|--------|------|------|------|
| PRIMARY | 主键 | id | 唯一标识 |
| uk_order_id | 唯一索引 | order_id | 订单号唯一性 |
| idx_member_symbol | 普通索引 | member_id, symbol | 查询用户交易对订单 |
| idx_member_status | 普通索引 | member_id, status | 查询用户特定状态订单 |
| idx_time | 普通索引 | time | 按时间范围查询 |

---

## 7. 与其他服务的调用关系

### 7.1 与 ucenter-rpc 的调用

#### 7.1.1 会员信息查询

**调用位置**: `logic/add_logic.go`

**调用方法**:
```go
memberInfo, err := l.svcCtx.MemberClient.FindMemberById(l.ctx, &memberpb.MemberReq{MemberId: req.UserId})
```

**调用目的**:
- 验证用户是否存在
- 验证用户交易状态是否正常（transactionStatus == 1）

**MemberInfo 关键字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 会员 ID |
| transactionStatus | int32 | 交易状态（0-禁止, 1-正常） |
| email | string | 邮箱 |
| mobile | string | 手机号 |
| googleState | int32 | Google 验证状态 |

#### 7.1.2 钱包信息查询

**调用位置**: `logic/add_logic.go`

**调用方法**:
```go
// 查询基础币种钱包
baseWallet, err := l.svcCtx.AssetClient.FindWalletBySymbol(l.ctx, &assetpb.AssetReq{
    UserId:   req.UserId,
    CoinName: baseSymbol,
})

// 查询交易币种钱包
coinWallet, err := l.svcCtx.AssetClient.FindWalletBySymbol(l.ctx, &assetpb.AssetReq{
    UserId:   req.UserId,
    CoinName: coinSymbol,
})
```

**调用目的**:
- 验证用户钱包是否存在
- 验证钱包是否被锁定（isLock == 0）
- 后续可扩展验证余额是否充足

**MemberWallet 关键字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 钱包 ID |
| memberId | int64 | 会员 ID |
| balance | float64 | 可用余额 |
| frozenBalance | float64 | 冻结余额 |
| isLock | int32 | 锁定状态（0-未锁定, 1-已锁定） |

### 7.2 与 market-rpc 的调用

#### 7.2.1 交易对信息查询

**调用位置**: `logic/add_logic.go`

**调用方法**:
```go
exchangeCoin, err := l.svcCtx.MarketClient.FindSymbolInfo(l.ctx, &marketpb.MarketReq{Symbol: req.Symbol})
```

**调用目的**:
- 验证交易对是否存在
- 验证交易对是否可交易（exchangeable == 1 && enable == 1）
- 获取价格限制（minSellPrice, maxBuyPrice）
- 获取市价交易支持配置（enableMarketBuy, enableMarketSell）
- 获取最大委托订单数限制（maxTradingOrder）

**ExchangeCoin 关键字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| symbol | string | 交易对符号 |
| baseSymbol | string | 基础币种符号 |
| coinSymbol | string | 交易币种符号 |
| exchangeable | int64 | 是否可交易（0-否, 1-是） |
| enable | int64 | 是否启用（0-否, 1-是） |
| minSellPrice | float64 | 最低卖出价格 |
| maxBuyPrice | float64 | 最高买入价格 |
| enableMarketBuy | int64 | 是否支持市价买入 |
| enableMarketSell | int64 | 是否支持市价卖出 |
| maxTradingOrder | int64 | 最大委托订单数 |
| fee | float64 | 交易手续费率 |

### 7.3 调用关系图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            exchange-rpc                                      │
│                                                                              │
│  ┌─────────────────┐                                                        │
│  │    AddLogic     │                                                        │
│  │                 │                                                        │
│  │  1. ────────────────────────► ucenter-rpc.MemberClient.FindMemberById  │
│  │     查询会员信息             │                                            │
│  │                              │                                            │
│  │  2. ────────────────────────► market-rpc.MarketClient.FindSymbolInfo    │
│  │     查询交易对信息           │                                            │
│  │                              │                                            │
│  │  3. ────────────────────────► ucenter-rpc.AssetClient.FindWalletBySymbol│
│  │     查询基础币种钱包         │                                            │
│  │                              │                                            │
│  │  4. ────────────────────────► ucenter-rpc.AssetClient.FindWalletBySymbol│
│  │     查询交易币种钱包         │                                            │
│  │                              │                                            │
│  │  5. ────────────────────────► OrderService.ValidateAddRequest           │
│  │     验证业务规则             │                                            │
│  │                              │                                            │
│  │  6. ────────────────────────► OrderService.FindCurrentTradingCount      │
│  │     检查委托订单数量         │                                            │
│  │                              │                                            │
│  │  7. ────────────────────────► OrderService.BuildNewOrder                │
│  │     构建新订单实体           │                                            │
│  │                              │                                            │
│  │  8. ────────────────────────► OrderService.AddOrder                     │
│  │     保存订单到数据库         │                                            │
│  │                              │                                            │
│  └──────────────────────────────┘                                            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. 配置说明

### 8.1 配置文件示例

**文件路径**: `/app/exchange/rpc/etc/exchange.yaml`

```yaml
# 服务名称，用于服务发现
Name: exchange.rpc

# 服务监听地址
ListenOn: 0.0.0.0:8083

# Etcd 服务发现配置
Etcd:
  Hosts:
    - etcd:2379
  Key: exchange.rpc

# MySQL 数据库配置
Mysql:
  DataSource: root:root@tcp(mysql:3306)/exchange?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
  MaxOpenConns: 100    # 最大打开连接数
  MaxIdleConns: 20     # 最大空闲连接数

# ucenter-rpc 客户端配置
UcenterRPC:
  Etcd:
    Hosts:
      - etcd:2379
    Key: ucenter.rpc
  NonBlock: true       # 非阻塞模式

# market-rpc 客户端配置
MarketRPC:
  Etcd:
    Hosts:
      - etcd:2379
    Key: market.rpc
  NonBlock: true       # 非阻塞模式

# Redis 缓存配置
Redis:
  Addrs:
    - redis:6379
  Password: ""
  DB: 0
```

### 8.2 配置项详细说明

#### 8.2.1 服务基础配置

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| Name | string | 服务名称，用于服务发现 | exchange.rpc |
| ListenOn | string | 服务监听地址 | 0.0.0.0:8083 |

#### 8.2.2 Etcd 配置

| 配置项 | 类型 | 说明 |
|--------|------|------|
| Etcd.Hosts | []string | Etcd 服务地址列表 |
| Etcd.Key | string | 服务发现键名 |

#### 8.2.3 MySQL 配置

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| Mysql.DataSource | string | 数据库连接字符串 | - |
| Mysql.MaxOpenConns | int | 最大打开连接数 | 100 |
| Mysql.MaxIdleConns | int | 最大空闲连接数 | 20 |

**DataSource 格式**:
```
{username}:{password}@tcp({host}:{port})/{database}?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
```

#### 8.2.4 RPC 客户端配置

| 配置项 | 类型 | 说明 |
|--------|------|------|
| {Service}RPC.Etcd.Hosts | []string | Etcd 服务地址列表 |
| {Service}RPC.Etcd.Key | string | 服务发现键名 |
| {Service}RPC.NonBlock | bool | 是否非阻塞模式 |

#### 8.2.5 Redis 配置

| 配置项 | 类型 | 说明 |
|--------|------|------|
| Redis.Addrs | []string | Redis 服务地址列表 |
| Redis.Password | string | Redis 密码 |
| Redis.DB | int | 数据库编号 |

### 8.3 环境变量覆盖

支持通过环境变量覆盖配置：

| 环境变量 | 对应配置 |
|----------|----------|
| EXCHANGE_RPC_LISTEN_ON | ListenOn |
| MYSQL_DATA_SOURCE | Mysql.DataSource |
| ETCD_HOSTS | Etcd.Hosts |
| REDIS_ADDRS | Redis.Addrs |

---

## 附录

### A. 错误码说明

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| "this user is forbidden to trade" | 用户被禁止交易 | 联系管理员解除限制 |
| "limit price mode requires price > 0" | 限价单价格无效 | 设置有效价格 |
| "amount must be > 0" | 订单数量无效 | 设置有效数量 |
| "coin forbidden" | 交易对不可交易 | 选择其他交易对 |
| "wallet locked" | 钱包被锁定 | 联系管理员解锁 |
| "price must be >= {minSellPrice}" | 卖出价格过低 | 提高卖出价格 |
| "price must be <= {maxBuyPrice}" | 买入价格过高 | 降低买入价格 |
| "market buy is not supported" | 不支持市价买入 | 使用限价单 |
| "market sell is not supported" | 不支持市价卖出 | 使用限价单 |
| "too many trading orders" | 委托订单过多 | 取消部分订单 |
| "nonsupport coin" | 不支持的交易对 | 选择支持的交易对 |
| "orderId not found" | 订单不存在 | 检查订单 ID |

### B. 性能优化建议

1. **数据库索引优化**
   - 确保 `order_id` 唯一索引
   - 为常用查询条件建立联合索引

2. **缓存策略**
   - 热门交易对配置缓存到 Redis
   - 用户钱包信息短期缓存

3. **连接池配置**
   - 根据并发量调整 MySQL 连接池大小
   - RPC 客户端使用连接复用

4. **批量操作**
   - 批量查询订单时使用 IN 查询
   - 避免循环单条查询

### C. 安全注意事项

1. **输入验证**
   - 验证所有用户输入参数
   - 防止 SQL 注入

2. **权限控制**
   - 验证订单归属用户
   - 验证用户交易权限

3. **敏感信息**
   - 不在日志中输出敏感信息
   - 使用加密通道传输

---

**文档版本**: 1.0
**最后更新**: 2026-06-08
**作者**: Claude AI
