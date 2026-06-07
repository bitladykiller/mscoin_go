# Market API 服务详细文档

## 目录

1. [服务概述](#1-服务概述)
2. [目录结构](#2-目录结构)
3. [每个文件的详细说明](#3-每个文件的详细说明)
   - [main.go - 服务入口](#31-maingo---服务入口)
   - [config/config.go - 配置定义](#32-configconfiggo---配置定义)
   - [svc/service_context.go - 服务上下文](#33-svcservice_contextgo---服务上下文)
   - [types/types.go - 类型定义](#34-typestypesgo---类型定义)
   - [handler 层](#35-handler-层)
   - [logic 层](#36-logic-层)
4. [请求处理流程](#4-请求处理流程)
5. [与其他服务的调用关系](#5-与其他服务的调用关系)
6. [配置说明](#6-配置说明)
7. [API 接口列表](#7-api-接口列表)
8. [公共包说明](#8-公共包说明)
9. [部署说明](#9-部署说明)

---

## 1. 服务概述

### 1.1 服务定位

`market-api` 是 MSCoin 交易系统的**市场数据 API 网关服务**，基于 go-zero 框架构建的 HTTP REST API 服务。它作为系统的前端入口，接收来自 Web 端、移动端的 HTTP 请求，并通过 RPC 调用后端 `market-rpc` 服务获取数据。

### 1.2 核心功能

| 功能模块 | 描述 |
|---------|------|
| 币种信息查询 | 获取币种的详细信息，包括名称、充提配置、钱包地址等 |
| 交易对信息查询 | 获取交易对的配置信息，包括精度、手续费、交易限制等 |
| 行情缩略图 | 获取所有交易对的实时行情快照，用于首页展示 |
| 行情趋势数据 | 获取带有趋势图的行情数据，用于绘制迷你趋势图 |
| K线历史数据 | 获取指定交易对的历史K线数据，用于图表展示 |
| 法币汇率查询 | 获取法币对USD的实时汇率，用于资产换算 |

### 1.3 在整体架构中的位置

```
┌─────────────────────────────────────────────────────────────────┐
│                         前端应用层                               │
│          (Web / Mobile App / 第三方接入)                        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │ HTTP REST API
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      market-api (本服务)                         │
│                    HTTP API 网关层                               │
│    - 接收 HTTP 请求                                              │
│    - 参数解析与校验                                              │
│    - 调用 RPC 服务                                               │
│    - 响应封装                                                    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │ gRPC
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      market-rpc 服务                             │
│                    业务逻辑层                                    │
│    - 币种/交易对数据管理                                         │
│    - 行情数据聚合                                                │
│    - K线数据计算                                                 │
│    - 汇率服务                                                    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      数据存储层                                  │
│    - MySQL (币种/交易对配置)                                     │
│    - Redis (行情缓存)                                            │
│    - MongoDB (K线数据)                                           │
└─────────────────────────────────────────────────────────────────┘
```

### 1.4 技术栈

- **框架**: go-zero (REST + RPC)
- **协议**: HTTP REST + gRPC
- **服务发现**: Etcd
- **配置格式**: YAML
- **容器化**: Docker

---

## 2. 目录结构

```
/Volumes/移动卷宗/学习/go/mscoin_go/app/market/api/
├── Dockerfile                 # Docker 构建文件
├── main.go                    # 服务入口文件
├── etc/                       # 配置文件目录
│   └── market-api.yaml        # 服务配置文件
└── internal/                  # 内部实现（不对外暴露）
    ├── config/                # 配置结构定义
    │   └── config.go          # Config 结构体定义
    ├── handler/               # HTTP 请求处理器
    │   ├── routes.go          # 路由注册
    │   ├── coin_info_handler.go        # 币种信息处理器
    │   ├── symbol_info_handler.go      # 交易对信息处理器
    │   ├── symbol_thumb_handler.go     # 行情缩略图处理器
    │   ├── symbol_thumb_trend_handler.go # 行情趋势处理器
    │   ├── history_handler.go          # K线历史处理器
    │   └── usd_rate_handler.go         # 汇率查询处理器
    ├── logic/                 # 业务逻辑层
    │   ├── base.go            # Logic 基类定义
    │   ├── coin_info_logic.go         # 币种信息逻辑
    │   ├── symbol_info_logic.go       # 交易对信息逻辑
    │   ├── symbol_thumb_logic.go      # 行情缩略图逻辑
    │   ├── symbol_thumb_trend_logic.go # 行情趋势逻辑
    │   ├── history_logic.go           # K线历史逻辑
    │   └── usd_rate_logic.go          # 汇率查询逻辑
    ├── svc/                   # 服务上下文
    │   └── service_context.go # 依赖注入容器
    └── types/                 # 类型定义
        └── types.go           # 请求/响应结构体
```

### 2.1 目录职责说明

| 目录/文件 | 职责 |
|----------|------|
| `main.go` | 服务启动入口，初始化服务器和依赖 |
| `etc/` | 存放配置文件 |
| `internal/config/` | 定义配置结构体，与 YAML 配置文件对应 |
| `internal/handler/` | HTTP 处理层，负责请求解析和响应封装 |
| `internal/logic/` | 业务逻辑层，调用 RPC 服务处理数据 |
| `internal/svc/` | 服务上下文，管理依赖注入 |
| `internal/types/` | 定义 API 的请求和响应数据结构 |

---

## 3. 每个文件的详细说明

### 3.1 main.go - 服务入口

**文件职责**：作为 market-api 服务的入口点，负责服务的启动流程。

**关键代码结构**：

```go
// 命令行参数定义
var configFile = flag.String("f", "etc/market-api.yaml", "配置文件路径")

// 主函数
func main() {
    // 1. 解析命令行参数
    flag.Parse()

    // 2. 加载配置文件
    var c config.Config
    conf.MustLoad(*configFile, &c)

    // 3. 创建 HTTP 服务器，配置 CORS
    server := rest.MustNewServer(
        c.RestConf,
        rest.WithCustomCors(func(header http.Header) {
            header.Set("Access-Control-Allow-Headers", "...")
        }, nil, "*"),
    )
    defer server.Stop()

    // 4. 初始化服务上下文
    ctx := svc.NewServiceContext(c)

    // 5. 注册路由处理器
    handler.RegisterHandlers(server, ctx)

    // 6. 启动服务器
    server.Start()
}
```

**启动流程详解**：

1. **参数解析**：通过 `-f` 参数指定配置文件路径，默认为 `etc/market-api.yaml`
2. **配置加载**：使用 go-zero 的 `conf.MustLoad` 加载 YAML 配置
3. **服务器创建**：
   - 使用配置中的 `RestConf` 创建 REST 服务器
   - 配置自定义 CORS 策略，允许跨域访问
   - 设置允许的请求头，包括认证相关的头（Authorization, token, x-auth-token）
4. **依赖注入**：创建 `ServiceContext`，初始化 RPC 客户端连接
5. **路由注册**：将所有 HTTP 路由绑定到对应的 handler
6. **服务启动**：阻塞式启动，监听 HTTP 请求

**CORS 配置说明**：

当前配置允许任意来源（`*`）访问，适用于开发环境。生产环境应考虑收紧 CORS 策略，限制允许的来源。

### 3.2 config/config.go - 配置定义

**文件职责**：定义服务配置的结构体，与 YAML 配置文件一一对应。

**Config 结构体详解**：

```go
type Config struct {
    rest.RestConf           // HTTP 服务器配置（嵌入）
    MarketRPC zrpc.RpcClientConf  // market-rpc 服务配置
}
```

**RestConf 字段说明**（继承自 go-zero）：

| 字段 | 类型 | 说明 | 示例值 |
|-----|------|------|--------|
| Name | string | 服务名称，用于日志标识 | "market-api" |
| Host | string | 监听地址 | "0.0.0.0" |
| Port | int | 监听端口 | 8889 |
| Timeout | int64 | 请求超时时间（毫秒） | 可选 |
| MaxConns | int | 最大连接数 | 可选 |
| MaxBytes | int64 | 请求体最大字节数 | 可选 |

**MarketRPC 字段说明**：

| 字段 | 类型 | 说明 |
|-----|------|------|
| Etcd.Hosts | []string | Etcd 注册中心地址列表 |
| Etcd.Key | string | 服务注册的 key |

**服务发现模式**：

1. **直连模式**：直接指定 `Endpoints` 地址列表
2. **服务发现模式**（推荐）：通过 Etcd 注册中心动态发现服务实例

### 3.3 svc/service_context.go - 服务上下文

**文件职责**：作为依赖注入容器，管理服务运行时的所有依赖项。

**设计模式**：依赖注入（Dependency Injection）

**ServiceContext 结构体**：

```go
type ServiceContext struct {
    Config       config.Config              // 服务配置
    MarketClient marketpb.MarketClient      // market-rpc 客户端
    RateClient   ratepb.ExchangeRateClient  // 汇率服务客户端
}
```

**NewServiceContext 函数**：

```go
func NewServiceContext(c config.Config) *ServiceContext {
    // 1. 创建 gRPC 客户端连接
    client := zrpc.MustNewClient(c.MarketRPC)

    // 2. 获取底层 gRPC 连接
    conn := client.Conn()

    // 3. 返回包含所有依赖的 ServiceContext
    return &ServiceContext{
        Config:       c,
        MarketClient: marketpb.NewMarketClient(conn),
        RateClient:   ratepb.NewExchangeRateClient(conn),
    }
}
```

**设计优点**：

1. **配置与业务解耦**：配置集中管理，业务逻辑不直接依赖配置文件
2. **便于测试**：可以轻松 mock RPC 客户端进行单元测试
3. **统一生命周期管理**：所有依赖在服务启动时初始化

**注意事项**：

- `MarketClient` 和 `RateClient` 共用同一个 gRPC 连接，减少连接开销
- `MustNewClient` 在连接失败时会 panic，确保服务启动时连接可用

### 3.4 types/types.go - 类型定义

**文件职责**：定义 HTTP 接口的输入输出数据结构，确保 API 层与前端之间的数据交互格式一致。

#### 3.4.1 RateRequest - 汇率请求

```go
type RateRequest struct {
    Unit string `path:"unit" json:"unit"`     // 法币单位代码
    IP   string `json:"ip,optional"`          // 客户端 IP（可选）
}
```

**使用场景**：查询指定法币对 USD 的汇率

#### 3.4.2 RateResponse - 汇率响应

```go
type RateResponse struct {
    Rate float64 `json:"rate"`  // 1 USD 可兑换的法币数量
}
```

#### 3.4.3 MarketReq - 市场请求（通用）

```go
type MarketReq struct {
    IP         string `json:"ip,optional" form:"ip,optional"`         // 客户端 IP
    Symbol     string `json:"symbol,optional" form:"symbol,optional"` // 交易对代码
    Unit       string `json:"unit,optional" form:"unit,optional"`     // 币种单位
    From       int64  `json:"from,optional" form:"from,optional"`     // 起始时间戳
    To         int64  `json:"to,optional" form:"to,optional"`         // 结束时间戳
    Resolution string `json:"resolution,optional" form:"resolution,optional"` // K线周期
}
```

**设计说明**：该结构体被多个接口复用，字段名与旧 API 保持一致以确保前端兼容。

#### 3.4.4 CoinThumbResp - 行情快照响应

```go
type CoinThumbResp struct {
    Symbol       string    `json:"symbol"`       // 交易对代码
    Open         float64   `json:"open"`         // 开盘价
    High         float64   `json:"high"`         // 最高价
    Low          float64   `json:"low"`          // 最低价
    Close        float64   `json:"close"`        // 最新价
    Chg          float64   `json:"chg"`          // 涨跌幅百分比
    Change       float64   `json:"change"`       // 涨跌额
    Volume       float64   `json:"volume"`       // 成交量
    Turnover     float64   `json:"turnover"`     // 成交额
    LastDayClose float64   `json:"lastDayClose"` // 昨日收盘价
    USDTRate     float64   `json:"usdRate"`      // 对 USDT 的汇率
    BaseUSDTRate float64   `json:"baseUsdRate"`  // 基础货币对 USDT 的汇率
    Zone         int       `json:"zone"`         // 交易区域
    Trend        []float64 `json:"trend,optional"` // 价格趋势数据
}
```

**计算公式**：
- 涨跌幅 = (close - lastDayClose) / lastDayClose * 100
- 涨跌额 = close - lastDayClose

#### 3.4.5 ExchangeCoinResp - 交易对信息响应

包含交易对的完整配置信息，字段众多，主要包括：

- 基本信息：ID, Symbol, BaseSymbol, CoinSymbol
- 精度配置：BaseCoinScale, CoinScale
- 交易限制：MinVolume, MaxVolume, MinTurnover
- 状态信息：Enable, Visible, Zone

#### 3.4.6 Coin - 币种信息响应

包含币种的完整信息，字段包括：

- 基本信息：ID, Name, NameCN, Unit
- 充提配置：CanRecharge, CanWithdraw, CanTransfer
- 限制设置：MinWithdrawAmount, MaxWithdrawAmount, WithdrawThreshold
- 汇率信息：CNYRate, USDTRate

#### 3.4.7 HistoryKline - K线历史响应

```go
type HistoryKline struct {
    List [][]any `json:"list"`  // K线数据列表
}
```

**数据格式**：每个元素为 `[时间戳, 开盘价, 最高价, 最低价, 收盘价, 成交量]`

**设计说明**：使用 `[][]any` 而非结构体数组，是为了保持与旧 API 的格式一致，前端可以直接使用此格式渲染 K 线图表。

### 3.5 handler 层

handler 层负责 HTTP 协议相关处理，包括参数解析、响应封装。

#### 3.5.1 routes.go - 路由注册

**RegisterHandlers 函数**：注册所有 HTTP 路由

```go
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
    server.AddRoutes(
        []rest.Route{
            {Method: http.MethodPost, Path: "/coin-info", Handler: CoinInfoHandler(serverCtx)},
            {Method: http.MethodPost, Path: "/exchange-rate/usd/:unit", Handler: UsdRateHandler(serverCtx)},
            {Method: http.MethodGet, Path: "/history", Handler: HistoryHandler(serverCtx)},
            {Method: http.MethodPost, Path: "/symbol-info", Handler: SymbolInfoHandler(serverCtx)},
            {Method: http.MethodPost, Path: "/symbol-thumb", Handler: SymbolThumbHandler(serverCtx)},
            {Method: http.MethodPost, Path: "/symbol-thumb-trend", Handler: SymbolThumbTrendHandler(serverCtx)},
        },
        rest.WithPrefix("/market"),  // 路由前缀
    )
}
```

**路由表**：

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | /market/coin-info | CoinInfoHandler | 币种信息查询 |
| POST | /market/exchange-rate/usd/:unit | UsdRateHandler | 汇率查询 |
| GET | /market/history | HistoryHandler | K线历史查询 |
| POST | /market/symbol-info | SymbolInfoHandler | 交易对信息查询 |
| POST | /market/symbol-thumb | SymbolThumbHandler | 行情缩略图 |
| POST | /market/symbol-thumb-trend | SymbolThumbTrendHandler | 行情趋势 |

#### 3.5.2 各 Handler 详细说明

**共同的处理流程**：

1. 解析请求参数（路径参数、表单参数或 JSON）
2. 提取客户端真实 IP 地址
3. 调用对应的 Logic 层方法
4. 使用统一响应封装器返回结果

**CoinInfoHandler**：

```go
func CoinInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.MarketReq
        if err := httpx.ParseForm(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        req.IP = httputil.ClientIP(r)
        resp, err := logic.NewCoinInfoLogic(r.Context(), svcCtx).CoinInfo(&req)
        httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
    }
}
```

**UsdRateHandler**：解析路径参数 `:unit`

**HistoryHandler**：使用 GET 方法，解析查询参数，响应只返回 `resp.List`

**SymbolThumbHandler / SymbolThumbTrendHandler**：无需额外参数，自动填充 IP

### 3.6 logic 层

logic 层负责业务逻辑处理，通过 RPC 调用后端服务。

#### 3.6.1 base.go - Logic 基类

```go
type marketLogicBase struct {
    ctx    context.Context        // 请求上下文
    svcCtx *svc.ServiceContext    // 服务上下文
}
```

**设计目的**：复用通用字段，避免每个 Logic 结构体重复定义。

#### 3.6.2 各 Logic 详细说明

**CoinInfoLogic**：

```go
func (l *CoinInfoLogic) CoinInfo(req *types.MarketReq) (*types.Coin, error) {
    ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
    defer cancel()

    coin, err := l.svcCtx.MarketClient.FindCoinInfo(ctx, &marketpb.MarketReq{Unit: req.Unit})
    if err != nil {
        return nil, err
    }

    resp := &types.Coin{}
    if err := copier.Copy(resp, coin); err != nil {
        return nil, errors.New("market coin payload copy failed")
    }
    return resp, nil
}
```

**处理步骤**：
1. 创建带超时的子上下文（5秒）
2. 调用 `MarketClient.FindCoinInfo` RPC 方法
3. 使用 `copier` 进行数据转换
4. 返回结果

**SymbolThumbLogic / SymbolThumbTrendLogic**：

两者调用同一个 RPC 方法 `FindSymbolThumbTrend`，区别在于：
- `SymbolThumbLogic`：不使用 Trend 字段
- `SymbolThumbTrendLogic`：使用完整的 Trend 数据

**HistoryLogic**：

特殊处理：将 RPC 响应的结构体数组转换为前端兼容的二维数组格式。

```go
list := make([][]any, len(payload.List))
for i, item := range payload.List {
    list[i] = []any{item.Time, item.Open, item.High, item.Low, item.Close, item.Volume}
}
```

**UsdRateLogic**：

调用 `RateClient.UsdRate` 获取汇率数据，超时时间较短（5秒）。

---

## 4. 请求处理流程

### 4.1 完整请求流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                      客户端发起请求                              │
│            POST /market/coin-info                               │
│            {"unit": "BTC"}                                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    go-zero REST Server                          │
│    - 路由匹配                                                    │
│    - CORS 处理                                                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    CoinInfoHandler                              │
│    1. httpx.ParseForm(r, &req) 解析参数                         │
│    2. httputil.ClientIP(r) 获取客户端 IP                        │
│    3. 调用 Logic 层                                             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    CoinInfoLogic                                │
│    1. 创建带超时的 context                                       │
│    2. 调用 MarketClient.FindCoinInfo RPC                        │
│    3. copier.Copy 数据转换                                       │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │ gRPC
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    market-rpc 服务                              │
│    - 查询数据库                                                  │
│    - 返回 Coin 数据                                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    响应封装                                      │
│    result.New().Deal(resp, err)                                 │
│    {                                                            │
│        "code": 0,                                               │
│        "message": "success",                                    │
│        "data": {...}                                            │
│    }                                                            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    返回给客户端                                  │
│    HTTP 200 OK                                                  │
│    Content-Type: application/json                               │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 请求处理关键点

1. **参数解析**：根据请求方法选择不同的解析方式
   - POST 表单：`httpx.ParseForm`
   - GET 查询参数：`httpx.ParseForm`
   - 路径参数：`httpx.ParsePath`

2. **客户端 IP 获取**：支持代理场景
   - 优先级：X-Real-IP > X-Forwarded-For > RemoteAddr

3. **超时控制**：所有 RPC 调用都设置超时
   - 轻量操作：5秒
   - 重量操作（K线、行情）：10秒

4. **统一响应格式**：
   ```json
   {
       "code": 0,        // 0 成功，500 失败
       "message": "success",
       "data": {...}
   }
   ```

---

## 5. 与其他服务的调用关系

### 5.1 调用的 RPC 服务

#### 5.1.1 market-rpc 服务

**服务地址**：通过 Etcd 服务发现，key 为 `market.rpc`

**调用的 RPC 方法**：

| 方法 | 用途 | 请求参数 | 响应类型 |
|-----|------|---------|---------|
| FindCoinInfo | 查询币种信息 | Unit | Coin |
| FindSymbolInfo | 查询交易对信息 | Symbol | ExchangeCoin |
| FindSymbolThumbTrend | 查询行情数据 | IP | SymbolThumbRes |
| HistoryKline | 查询K线历史 | Symbol, From, To, Resolution | HistoryRes |

**protobuf 定义**：

```protobuf
service Market {
    rpc FindSymbolThumbTrend(MarketReq) returns (SymbolThumbRes);
    rpc FindSymbolInfo(MarketReq) returns (ExchangeCoin);
    rpc FindCoinInfo(MarketReq) returns (Coin);
    rpc HistoryKline(MarketReq) returns (HistoryRes);
}
```

#### 5.1.2 rate-rpc 服务（汇率服务）

**服务地址**：与 market-rpc 共用连接

**调用的 RPC 方法**：

| 方法 | 用途 | 请求参数 | 响应类型 |
|-----|------|---------|---------|
| UsdRate | 查询汇率 | Unit, IP | RateRes |

**protobuf 定义**：

```protobuf
service ExchangeRate {
    rpc UsdRate(RateReq) returns (RateRes);
}

message RateReq {
    string unit = 1;  // 法币单位
    string ip = 2;    // 客户端 IP
}

message RateRes {
    double rate = 1;  // 汇率值
}
```

### 5.2 为什么需要调用这些服务

#### 5.2.1 架构设计原因

采用 **API 网关 + RPC 微服务** 架构模式，原因如下：

1. **职责分离**：
   - API 层：处理 HTTP 协议，参数校验，响应封装
   - RPC 层：处理业务逻辑，数据聚合，缓存管理

2. **复用性**：
   - 多个 API 服务可以复用同一个 RPC 服务
   - 例如：Web API、管理后台 API、第三方开放 API

3. **可扩展性**：
   - API 层可以独立扩缩容
   - RPC 层可以独立扩缩容

4. **性能**：
   - gRPC 使用 Protocol Buffers，序列化效率高
   - HTTP/2 支持多路复用

#### 5.2.2 业务需求原因

**调用 market-rpc 的原因**：

- 币种信息、交易对信息存储在数据库，需要 RPC 服务查询
- 行情数据需要聚合多个数据源（交易所行情、K线数据），计算复杂
- K线数据可能存储在 MongoDB 或时序数据库，需要专门服务处理

**调用 rate-rpc 的原因**：

- 汇率数据需要实时更新
- 可能对接第三方汇率 API
- 支持基于 IP 地理位置的汇率选择

---

## 6. 配置说明

### 6.1 配置文件示例

```yaml
# etc/market-api.yaml

# 服务名称，用于日志标识和监控
Name: market-api

# 监听地址，0.0.0.0 表示监听所有网卡
Host: 0.0.0.0

# 监听端口号
Port: 8889

# market-rpc 服务配置
MarketRPC:
  # Etcd 服务发现配置
  Etcd:
    # Etcd 集群地址列表
    Hosts:
      - etcd:2379
    # 服务注册的 key
    Key: market.rpc
```

### 6.2 配置字段详解

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|-----|------|-----|--------|------|
| Name | string | 是 | - | 服务名称，用于日志标识 |
| Host | string | 是 | - | 监听地址，0.0.0.0 监听所有网卡 |
| Port | int | 是 | - | 监听端口号 |
| Timeout | int64 | 否 | 3000 | 请求超时时间（毫秒） |
| MaxConns | int | 否 | 10000 | 最大并发连接数 |
| MaxBytes | int64 | 否 | 1048576 | 请求体最大字节数（1MB） |
| MarketRPC.Etcd.Hosts | []string | 是 | - | Etcd 集群地址 |
| MarketRPC.Etcd.Key | string | 是 | - | 服务注册的 key |
| MarketRPC.Endpoints | []string | 否 | - | 直连模式的服务地址 |

### 6.3 环境变量支持

go-zero 支持通过环境变量覆盖配置：

```bash
# 设置环境变量
export MARKET_API_HOST=0.0.0.0
export MARKET_API_PORT=8889
```

### 6.4 多环境配置

建议为不同环境创建不同的配置文件：

```
etc/
├── market-api.yaml           # 默认配置
├── market-api-dev.yaml       # 开发环境
├── market-api-test.yaml      # 测试环境
└── market-api-prod.yaml      # 生产环境
```

启动时指定配置文件：

```bash
./market-api -f etc/market-api-prod.yaml
```

---

## 7. API 接口列表

### 7.1 币种信息查询

**接口路径**：`POST /market/coin-info`

**功能描述**：查询指定币种的详细信息

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| unit | string | 是 | 币种单位，如 "BTC", "ETH", "USDT" |

**请求示例**：

```bash
curl -X POST http://localhost:8889/market/coin-info \
  -d "unit=BTC"
```

**响应示例**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "Bitcoin",
    "nameCn": "比特币",
    "unit": "BTC",
    "status": 1,
    "canRecharge": 1,
    "canWithdraw": 1,
    "canTransfer": 1,
    "canAutoWithdraw": 1,
    "minWithdrawAmount": 0.001,
    "maxWithdrawAmount": 100,
    "withdrawThreshold": 10,
    "usdRate": 42000.00,
    "cnyRate": 304000.00
  }
}
```

**响应字段说明**：

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | int | 币种 ID |
| name | string | 币种英文名 |
| nameCn | string | 币种中文名 |
| unit | string | 币种单位代码 |
| status | int | 币种状态，1 启用 |
| canRecharge | int | 是否允许充值，1 允许 |
| canWithdraw | int | 是否允许提现，1 允许 |
| minWithdrawAmount | float64 | 最小提现金额 |
| maxWithdrawAmount | float64 | 最大提现金额 |
| usdRate | float64 | 对 USD 汇率 |

---

### 7.2 交易对信息查询

**接口路径**：`POST /market/symbol-info`

**功能描述**：查询指定交易对的配置信息

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| symbol | string | 是 | 交易对代码，如 "BTCUSDT" |

**请求示例**：

```bash
curl -X POST http://localhost:8889/market/symbol-info \
  -d "symbol=BTCUSDT"
```

**响应示例**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "symbol": "BTCUSDT",
    "baseSymbol": "BTC",
    "coinSymbol": "USDT",
    "baseCoinScale": 8,
    "coinScale": 2,
    "enable": 1,
    "fee": 0.001,
    "minVolume": 0.0001,
    "maxVolume": 1000,
    "minTurnover": 10
  }
}
```

---

### 7.3 行情缩略图

**接口路径**：`POST /market/symbol-thumb`

**功能描述**：获取所有交易对的行情快照

**请求参数**：无需参数，IP 自动获取

**请求示例**：

```bash
curl -X POST http://localhost:8889/market/symbol-thumb
```

**响应示例**：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "symbol": "BTCUSDT",
      "open": 42000.00,
      "high": 43000.00,
      "low": 41500.00,
      "close": 42500.00,
      "chg": 1.19,
      "change": 500.00,
      "volume": 12345.67,
      "turnover": 520000000.00,
      "usdRate": 1.0,
      "zone": 0
    },
    {
      "symbol": "ETHUSDT",
      "open": 2200.00,
      "high": 2300.00,
      "low": 2150.00,
      "close": 2280.00,
      "chg": 3.64,
      "change": 80.00,
      "volume": 50000.00,
      "turnover": 112000000.00,
      "usdRate": 1.0,
      "zone": 0
    }
  ]
}
```

---

### 7.4 行情趋势数据

**接口路径**：`POST /market/symbol-thumb-trend`

**功能描述**：获取带有趋势图的行情数据

**请求参数**：无需参数，IP 自动获取

**请求示例**：

```bash
curl -X POST http://localhost:8889/market/symbol-thumb-trend
```

**响应示例**：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "symbol": "BTCUSDT",
      "open": 42000.00,
      "close": 42500.00,
      "chg": 1.19,
      "trend": [42000, 42100, 42300, 42400, 42500]
    }
  ]
}
```

**与 /symbol-thumb 的区别**：响应中包含 `trend` 数组，用于绘制迷你趋势图。

---

### 7.5 K线历史数据

**接口路径**：`GET /market/history`

**功能描述**：获取指定交易对的历史K线数据

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| symbol | string | 是 | 交易对代码，如 "BTCUSDT" |
| from | int64 | 是 | 起始时间戳（毫秒） |
| to | int64 | 是 | 结束时间戳（毫秒） |
| resolution | string | 是 | K线周期 |

**K线周期值**：

| 值 | 说明 |
|---|------|
| 1 | 1分钟 |
| 5 | 5分钟 |
| 15 | 15分钟 |
| 30 | 30分钟 |
| 60 | 1小时 |
| 1D | 1天 |
| 1W | 1周 |
| 1M | 1月 |

**请求示例**：

```bash
curl -X GET "http://localhost:8889/market/history?symbol=BTCUSDT&from=1704067200000&to=1704153600000&resolution=1D"
```

**响应示例**：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    [1704067200000, 42000.00, 43000.00, 41500.00, 42500.00, 12345.67],
    [1704153600000, 42500.00, 43500.00, 42000.00, 43000.00, 15678.90]
  ]
}
```

**数据格式**：每项为 `[时间戳, 开盘价, 最高价, 最低价, 收盘价, 成交量]`

---

### 7.6 法币汇率查询

**接口路径**：`POST /market/exchange-rate/usd/:unit`

**功能描述**：查询指定法币对 USD 的汇率

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| unit | string | 是 | 法币单位（路径参数），如 "CNY", "EUR", "JPY" |

**请求示例**：

```bash
curl -X POST http://localhost:8889/market/exchange-rate/usd/CNY
```

**响应示例**：

```json
{
  "code": 0,
  "message": "success",
  "data": 7.24
}
```

**说明**：`data` 字段直接返回汇率值，表示 1 USD = 7.24 CNY

---

## 8. 公共包说明

### 8.1 pkg/result - 统一响应封装

**位置**：`/pkg/result/result.go`

**Result 结构体**：

```go
type Result struct {
    Code    int    `json:"code"`    // 状态码：0 成功，500 失败
    Message string `json:"message"` // 状态描述
    Data    any    `json:"data"`    // 业务数据
}
```

**使用方法**：

```go
// 成功响应
resp, err := logic.SomeMethod()
httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))

// 等价于：
// 成功时: {"code": 0, "message": "success", "data": resp}
// 失败时: {"code": 500, "message": "错误信息", "data": null}
```

### 8.2 pkg/httputil - HTTP 工具函数

**位置**：`/pkg/httputil/client_ip.go`

**ClientIP 函数**：从 HTTP 请求中提取客户端真实 IP

**IP 获取优先级**：

1. `X-Real-IP` 请求头
2. `X-Forwarded-For` 请求头的第一个 IP
3. `RemoteAddr`

**使用方法**：

```go
ip := httputil.ClientIP(r)
```

**注意事项**：

- `X-Forwarded-For` 可被客户端伪造
- 生产环境应配置可信代理列表

---

## 9. 部署说明

### 9.1 Docker 构建

**Dockerfile 分析**：

```dockerfile
# 第一阶段：构建
FROM golang:1.26.3 AS builder
WORKDIR /workspace
# 复制依赖文件
COPY go.mod go.sum go.work go.work.sum ./
# 下载依赖（带缓存）
RUN go mod download
# 复制源代码
COPY app ./app
COPY idl ./idl
COPY pkg ./pkg
# 构建
RUN CGO_ENABLED=0 go build -o /out/server ./app/market/api

# 第二阶段：运行
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY --from=builder /workspace/app/market/api/etc /app/etc
ENTRYPOINT ["/app/server"]
CMD ["-f", "/app/etc/market-api.yaml"]
```

**构建命令**：

```bash
docker build -t market-api:latest -f app/market/api/Dockerfile .
```

### 9.2 启动命令

**直接运行**：

```bash
./market-api -f etc/market-api.yaml
```

**Docker 运行**：

```bash
docker run -d \
  --name market-api \
  -p 8889:8889 \
  -v /path/to/market-api.yaml:/app/etc/market-api.yaml \
  market-api:latest
```

### 9.3 Docker Compose 示例

```yaml
version: '3.8'
services:
  market-api:
    build:
      context: .
      dockerfile: app/market/api/Dockerfile
    ports:
      - "8889:8889"
    environment:
      - TZ=Asia/Shanghai
    depends_on:
      - etcd
      - market-rpc
    networks:
      - mscoin-network

  etcd:
    image: bitnami/etcd:latest
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
    ports:
      - "2379:2379"
    networks:
      - mscoin-network

  market-rpc:
    build:
      context: .
      dockerfile: app/market/rpc/Dockerfile
    networks:
      - mscoin-network

networks:
  mscoin-network:
    driver: bridge
```

### 9.4 健康检查

go-zero 默认提供 `/health` 端点用于健康检查：

```bash
curl http://localhost:8889/health
```

### 9.5 监控指标

go-zero 默认在 `/metrics` 端点暴露 Prometheus 指标：

```bash
curl http://localhost:8889/metrics
```

---

## 附录

### A. 错误码说明

| Code | 说明 |
|------|------|
| 0 | 成功 |
| 500 | 通用业务失败 |
| 400 | 参数错误 |
| 401 | 未授权 |
| 404 | 资源不存在 |

### B. 常见问题

**Q1: 服务启动失败，提示连接 Etcd 失败？**

A: 检查 Etcd 服务是否正常运行，配置的 Hosts 地址是否正确。

**Q2: RPC 调用超时？**

A: 检查 market-rpc 服务是否正常运行，网络是否通畅。

**Q3: CORS 跨域问题？**

A: 当前配置允许任意来源访问，如需限制，修改 main.go 中的 CORS 配置。

### C. 相关文档

- [go-zero 官方文档](https://go-zero.dev/)
- [gRPC 官方文档](https://grpc.io/docs/)
- [Protocol Buffers 指南](https://protobuf.dev/programming-guides/)
