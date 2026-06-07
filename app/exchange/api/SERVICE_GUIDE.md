# Exchange API 服务详细文档

## 目录

1. [服务概述](#1-服务概述)
2. [目录结构](#2-目录结构)
3. [每个文件的详细说明](#3-每个文件的详细说明)
   - [main.go - 服务入口](#31-maingo---服务入口)
   - [config/config.go - 配置定义](#32-configconfiggo---配置定义)
   - [types/types.go - 请求响应类型](#33-typestypesgo---请求响应类型)
   - [svc/service_context.go - 服务上下文](#34-svcservice_contextgo---服务上下文)
   - [middleware/auth_middleware.go - 认证中间件](#35-middlewareauth_middlewarego---认证中间件)
   - [handler/routes.go - 路由注册](#36-handlerroutesgo---路由注册)
   - [handler/add_handler.go - 新增订单处理器](#37-handleradd_handlergo---新增订单处理器)
   - [handler/current_handler.go - 当前订单处理器](#38-handlercurrent_handlergo---当前订单处理器)
   - [handler/history_handler.go - 历史订单处理器](#39-handlerhistory_handlergo---历史订单处理器)
   - [logic/add_logic.go - 新增订单业务逻辑](#310-logicadd_logicgo---新增订单业务逻辑)
   - [logic/current_logic.go - 当前订单业务逻辑](#311-logiccurrent_logicgo---当前订单业务逻辑)
   - [logic/history_logic.go - 历史订单业务逻辑](#312-logichistory_logicgo---历史订单业务逻辑)
4. [请求处理流程](#4-请求处理流程)
5. [与其他服务的调用关系](#5-与其他服务的调用关系)
6. [配置说明](#6-配置说明)
7. [API 接口列表](#7-api-接口列表)
8. [依赖的公共包](#8-依赖的公共包)

---

## 1. 服务概述

### 1.1 功能定位

`exchange-api` 是 MSCoin 交易系统的**交易订单 HTTP API 服务**，为前端提供交易相关的 RESTful 接口。它是用户进行交易操作的入口，负责接收用户下单请求、查询当前委托和历史订单等功能。

### 1.2 核心职责

| 职责 | 说明 |
|------|------|
| **下单功能** | 接收用户下单请求（市价单/限价单），验证参数并转发到 exchange-rpc 创建订单 |
| **当前委托查询** | 查询用户当前正在委托中的订单（状态为 TRADING） |
| **历史订单查询** | 查询用户已完成或已取消的历史订单 |
| **认证验证** | 通过 JWT 中间件验证用户身份，确保只有登录用户才能访问 |

### 1.3 在整体架构中的位置

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端 (Web/App)                         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      exchange-api (本服务)                       │
│                    HTTP REST 服务 (端口 8890)                    │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    认证中间件 (JWT)                        │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                │                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Handler 层 (HTTP 处理)                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                │                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Logic 层 (业务逻辑)                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼ gRPC 调用
┌─────────────────────────────────────────────────────────────────┐
│                       exchange-rpc 服务                          │
│                    订单核心业务逻辑服务                           │
└─────────────────────────────────────────────────────────────────┘
                                │
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
              ucenter-rpc  market-rpc   数据库
```

### 1.4 技术栈

- **框架**: go-zero REST 框架
- **通信协议**: HTTP/1.1 + JSON
- **服务发现**: Etcd (通过 go-zero 的 RPC 客户端)
- **认证方式**: JWT (HS256 算法)
- **容器化**: Docker + Dockerfile 多阶段构建

---

## 2. 目录结构

```
app/exchange/api/
├── Dockerfile                 # Docker 构建文件，用于容器化部署
├── main.go                    # 服务入口文件，负责启动 HTTP 服务器
├── etc/                       # 配置文件目录
│   └── exchange-api.yaml      # 服务配置文件
└── internal/                  # 内部实现（不对外暴露）
    ├── config/                # 配置结构定义
    │   └── config.go          # 配置结构体定义
    ├── handler/               # HTTP 请求处理器
    │   ├── routes.go          # 路由注册
    │   ├── add_handler.go     # 新增订单处理器
    │   ├── current_handler.go # 当前订单查询处理器
    │   └── history_handler.go # 历史订单查询处理器
    ├── logic/                 # 业务逻辑层
    │   ├── add_logic.go       # 新增订单业务逻辑
    │   ├── current_logic.go   # 当前订单查询业务逻辑
    │   └── history_logic.go   # 历史订单查询业务逻辑
    ├── middleware/            # 中间件
    │   └── auth_middleware.go # JWT 认证中间件
    ├── svc/                   # 服务上下文
    │   └── service_context.go # 服务依赖聚合
    └── types/                 # 类型定义
        └── types.go           # 请求/响应类型定义
```

### 2.1 目录职责说明

| 目录/文件 | 职责 |
|-----------|------|
| `main.go` | 服务启动入口，初始化服务器、注册路由、启动监听 |
| `etc/` | 存放配置文件，支持不同环境的配置 |
| `internal/config/` | 配置结构体定义，将 YAML 配置映射到 Go 结构体 |
| `internal/handler/` | HTTP 请求处理器，负责解析请求、调用 Logic、返回响应 |
| `internal/logic/` | 业务逻辑层，负责具体业务处理，调用 RPC 服务 |
| `internal/middleware/` | 中间件实现，如认证、日志等 |
| `internal/svc/` | 服务上下文，聚合所有运行时依赖 |
| `internal/types/` | 请求和响应类型定义，用于 HTTP 参数绑定 |

---

## 3. 每个文件的详细说明

### 3.1 main.go - 服务入口

**文件职责**:
- 作为 exchange-api 服务的入口点，负责启动整个服务
- 加载配置、创建服务器、注册路由、启动监听

**关键代码分析**:

```go
// configFile 指定配置文件路径，默认为 "etc/exchange-api.yaml"
var configFile = flag.String("f", "etc/exchange-api.yaml", "配置文件路径")
```

`configFile` 是一个命令行参数，允许在启动时指定不同的配置文件，便于多环境部署。

```go
func main() {
    flag.Parse()

    var c config.Config
    conf.MustLoad(*configFile, &c)  // 加载 YAML 配置到结构体
```

`conf.MustLoad` 会读取配置文件并解析到 `config.Config` 结构体中。如果解析失败，程序会 panic。

```go
    server := rest.MustNewServer(
        c.RestConf,
        rest.WithCustomCors(func(header http.Header) {
            header.Set("Access-Control-Allow-Headers",
                "DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization,token,x-auth-token")
        }, nil, "*"),
    )
```

创建 REST 服务器时配置了 CORS（跨域资源共享）:
- `Access-Control-Allow-Headers`: 允许的请求头
- `"*"`: 允许所有来源访问

```go
    ctx := svc.NewServiceContext(c)
    handler.RegisterHandlers(server, ctx)

    fmt.Printf("Starting exchange api server at %s:%d...\n", c.Host, c.Port)
    server.Start()
}
```

服务启动流程:
1. 创建服务上下文 `ServiceContext`（初始化 RPC 客户端、中间件等）
2. 注册路由处理器
3. 启动服务器监听请求

---

### 3.2 config/config.go - 配置定义

**文件职责**:
- 定义服务配置的结构体
- 包含 REST 服务器配置、RPC 客户端配置、JWT 认证配置

**关键结构体**:

```go
// AuthConfig 定义 JWT 认证相关配置
type AuthConfig struct {
    AccessSecret string  // JWT 签名使用的密钥
    AccessExpire int64   // JWT 令牌的过期时间（秒）
}
```

`AuthConfig` 用于 JWT 认证:
- `AccessSecret`: 签名密钥，必须保密，建议使用环境变量或配置中心管理
- `AccessExpire`: 令牌有效期，默认 604800 秒（7 天）

```go
// Config 是 exchange-api 服务的完整配置结构
type Config struct {
    rest.RestConf                    // 嵌入 go-zero 的 REST 配置
    ExchangeRPC zrpc.RpcClientConf   // exchange-rpc 服务客户端配置
    JWT AuthConfig                   // JWT 认证配置
}
```

`Config` 结构体说明:
- `rest.RestConf`: go-zero 内置的 REST 服务器配置，包含 Name、Host、Port 等
- `ExchangeRPC`: RPC 客户端配置，包含 Etcd 地址和服务名称
- `JWT`: 认证配置，用于验证用户令牌

**配置示例**:
```yaml
Name: exchange-api
Host: 0.0.0.0
Port: 8890
JWT:
  AccessSecret: "!@#$mscoin"
  AccessExpire: 604800
ExchangeRPC:
  Etcd:
    Hosts:
      - etcd:2379
    Key: exchange.rpc
```

---

### 3.3 types/types.go - 请求响应类型

**文件职责**:
- 定义 HTTP 请求参数的结构体
- 提供参数验证方法

**关键结构体**:

```go
// ExchangeReq 是交易相关的通用请求结构体
type ExchangeReq struct {
    IP          string  `json:"ip,optional" form:"ip,optional"`           // 客户端 IP
    Symbol      string  `json:"symbol,optional" form:"symbol,optional"`   // 交易对符号
    PageNo      int64   `json:"pageNo,optional" form:"pageNo,optional"`   // 分页页码
    PageSize    int64   `json:"pageSize,optional" form:"pageSize,optional"` // 每页记录数
    Price       float64 `json:"price,optional" form:"price,optional"`     // 订单价格
    Amount      float64 `json:"amount,optional" form:"amount,optional"`   // 订单数量
    Direction   string  `json:"direction,optional" form:"direction,optional"` // 订单方向
    Type        string  `json:"type,optional" form:"type,optional"`       // 订单类型
    UseDiscount float64 `json:"useDiscount,optional" form:"useDiscount,optional"` // 折扣金额
}
```

**字段详细说明**:

| 字段 | 类型 | 说明 | 使用场景 |
|------|------|------|----------|
| `IP` | string | 客户端 IP 地址 | 用于风控和日志记录，由服务端自动获取 |
| `Symbol` | string | 交易对符号，如 "BTCUSDT" | 下单和查询时指定交易对 |
| `PageNo` | int64 | 分页页码，从 0 开始 | 查询订单列表时使用 |
| `PageSize` | int64 | 每页记录数 | 查询订单列表时使用 |
| `Price` | float64 | 订单价格 | 限价单必填，市价单不需要 |
| `Amount` | float64 | 订单数量 | 下单时必填 |
| `Direction` | string | 订单方向 | "BUY" 买入或 "SELL" 卖出 |
| `Type` | string | 订单类型 | "MARKET_PRICE" 市价或 "LIMIT_PRICE" 限价 |
| `UseDiscount` | float64 | 使用的折扣金额 | 可选字段 |

```go
// OrderValid 验证订单请求参数的有效性
func (r *ExchangeReq) OrderValid() bool {
    return r.Direction != "" && r.Type != ""
}
```

`OrderValid` 方法用于验证下单请求是否有效:
- 检查 `Direction` 和 `Type` 字段是否非空
- 在 `AddLogic` 中调用此方法进行前置验证

---

### 3.4 svc/service_context.go - 服务上下文

**文件职责**:
- 聚合服务的所有运行时依赖
- 初始化 RPC 客户端、中间件等
- 提供统一的依赖注入入口

**关键结构体**:

```go
// ServiceContext 聚合 exchange-api 服务的所有运行时依赖
type ServiceContext struct {
    Config      config.Config        // 服务配置
    Auth        rest.Middleware      // 认证中间件
    OrderClient orderpb.OrderClient  // exchange-rpc 客户端
}
```

**字段说明**:
- `Config`: 保存服务配置，供各模块访问
- `Auth`: JWT 认证中间件，所有需要认证的路由都会经过此中间件
- `OrderClient`: gRPC 客户端，用于调用 exchange-rpc 服务

```go
func NewServiceContext(c config.Config) *ServiceContext {
    // 创建 exchange-rpc 客户端连接
    client := zrpc.MustNewClient(c.ExchangeRPC)
    return &ServiceContext{
        Config:      c,
        Auth:        middleware.NewAuthMiddleware(c.JWT.AccessSecret).Handle,
        OrderClient: orderpb.NewOrderClient(client.Conn()),
    }
}
```

**初始化流程**:
1. 使用 `zrpc.MustNewClient` 创建 RPC 客户端连接
   - 通过 Etcd 发现 exchange-rpc 服务
   - 如果连接失败，程序会 panic
2. 创建认证中间件，传入 JWT 密钥
3. 创建 OrderClient，用于调用 RPC 方法

---

### 3.5 middleware/auth_middleware.go - 认证中间件

**文件职责**:
- 实现 JWT 认证中间件
- 验证请求头中的令牌并提取用户 ID
- 将用户 ID 存入请求上下文供后续使用

**关键结构体和方法**:

```go
// contextKey 定义上下文键的类型，确保类型安全
type contextKey string

// userIDKey 是存储用户 ID 的上下文键
const userIDKey contextKey = "userId"
```

使用自定义类型 `contextKey` 可以避免上下文键冲突，这是 Go 的最佳实践。

```go
// AuthMiddleware 是 JWT 认证中间件
type AuthMiddleware struct {
    secret string  // JWT 签名验证密钥
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
    return &AuthMiddleware{secret: secret}
}
```

```go
// Handle 实现中间件处理函数
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. 从请求头获取认证令牌
        token := r.Header.Get("x-auth-token")
        if token == "" {
            failed := result.New()
            failed.Fail(4000, "no login")
            httpx.WriteJson(w, http.StatusOK, failed)
            return
        }

        // 2. 解析令牌获取用户 ID
        userID, err := auth.ParseUserID(token, m.secret)
        if err != nil {
            failed := result.New()
            failed.Fail(4000, "no login")
            httpx.WriteJson(w, http.StatusOK, failed)
            return
        }

        // 3. 将用户 ID 存入上下文，供后续处理器使用
        ctx := context.WithValue(r.Context(), userIDKey, userID)
        next(w, r.WithContext(ctx))
    }
}
```

**认证流程**:
1. 从请求头 `x-auth-token` 获取 JWT 令牌
2. 如果令牌缺失，返回错误码 4000（未登录）
3. 使用 `auth.ParseUserID` 解析令牌，验证签名和有效期
4. 如果验证失败，返回错误码 4000（未登录）
5. 将用户 ID 存入请求上下文
6. 调用下一个处理器

```go
// UserIDFromContext 从上下文中提取用户 ID
func UserIDFromContext(ctx context.Context) int64 {
    value, _ := ctx.Value(userIDKey).(int64)
    return value
}
```

`UserIDFromContext` 供 Logic 层调用，从上下文中获取已认证的用户 ID。

---

### 3.6 handler/routes.go - 路由注册

**文件职责**:
- 注册所有 HTTP 路由
- 配置认证中间件
- 定义路由前缀

**关键函数**:

```go
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
    server.AddRoutes(
        rest.WithMiddlewares(
            []rest.Middleware{serverCtx.Auth},  // 认证中间件
            []rest.Route{
                {Method: http.MethodPost, Path: "/order/add", Handler: AddHandler(serverCtx)},
                {Method: http.MethodPost, Path: "/order/current", Handler: CurrentHandler(serverCtx)},
                {Method: http.MethodPost, Path: "/order/history", Handler: HistoryHandler(serverCtx)},
            }...,
        ),
        rest.WithPrefix("/exchange"),  // 路由前缀
    )
}
```

**路由配置说明**:

| 完整路径 | 方法 | 处理器 | 功能 |
|----------|------|--------|------|
| `/exchange/order/add` | POST | AddHandler | 新增订单 |
| `/exchange/order/current` | POST | CurrentHandler | 查询当前订单 |
| `/exchange/order/history` | POST | HistoryHandler | 查询历史订单 |

**中间件链**:
```
请求 -> Auth 中间件 -> Handler -> Logic -> RPC 调用 -> 响应
```

---

### 3.7 handler/add_handler.go - 新增订单处理器

**文件职责**:
- 处理新增订单的 HTTP 请求
- 解析请求参数，获取客户端 IP
- 调用业务逻辑层处理请求

**关键函数**:

```go
func AddHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.ExchangeReq
        if err := httpx.ParseForm(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        // 获取客户端 IP 并设置到请求中
        req.IP = httputil.ClientIP(r)

        // 调用业务逻辑层处理请求
        resp, err := logic.NewAddLogic(r.Context(), svcCtx).Add(&req)
        httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
    }
}
```

**处理流程**:
1. 解析请求参数到 `ExchangeReq` 结构体
2. 调用 `httputil.ClientIP` 获取客户端真实 IP
3. 创建 `AddLogic` 实例并调用 `Add` 方法
4. 使用 `result.New().Deal()` 封装响应结果

**响应格式**:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "orderId": "ORD123456"
    }
}
```

---

### 3.8 handler/current_handler.go - 当前订单处理器

**文件职责**:
- 处理查询当前订单的 HTTP 请求
- 返回用户当前正在委托中的订单

**关键函数**:

```go
func CurrentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.ExchangeReq
        if err := httpx.ParseForm(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        req.IP = httputil.ClientIP(r)
        resp, err := logic.NewCurrentLogic(r.Context(), svcCtx).Current(&req)
        httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
    }
}
```

**响应格式**:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "content": [
            {
                "orderId": "ORD123456",
                "symbol": "BTCUSDT",
                "direction": "BUY",
                "type": "LIMIT_PRICE",
                "price": 50000.0,
                "amount": 0.1,
                "tradedAmount": 0.05,
                "status": "TRADING"
            }
        ],
        "totalElements": 1,
        "number": 0,
        "totalPages": 1,
        "hasNext": false,
        "isLast": true
    }
}
```

---

### 3.9 handler/history_handler.go - 历史订单处理器

**文件职责**:
- 处理查询历史订单的 HTTP 请求
- 返回用户已完成或已取消的订单

**关键函数**:

```go
func HistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.ExchangeReq
        if err := httpx.ParseForm(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        req.IP = httputil.ClientIP(r)
        resp, err := logic.NewHistoryLogic(r.Context(), svcCtx).History(&req)
        httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
    }
}
```

**返回的订单状态**:
- `COMPLETED`: 已完成（订单全部成交）
- `CANCELED`: 已取消（用户主动取消）
- `OVERTIMED`: 已超时（订单超过有效期被系统取消）

---

### 3.10 logic/add_logic.go - 新增订单业务逻辑

**文件职责**:
- 实现新增订单的业务逻辑
- 验证请求参数
- 调用 exchange-rpc 服务创建订单

**关键结构体和方法**:

```go
type AddLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
    return &AddLogic{ctx: ctx, svcCtx: svcCtx}
}
```

```go
func (l *AddLogic) Add(req *types.ExchangeReq) (string, error) {
    // 1. 验证订单请求参数
    if !req.OrderValid() {
        return "", errors.New("invalid request")
    }

    // 2. 从上下文获取已认证的用户 ID
    userID := middleware.UserIDFromContext(l.ctx)

    // 3. 调用 RPC 服务创建订单
    resp, err := l.svcCtx.OrderClient.Add(l.ctx, &orderpb.OrderReq{
        Symbol:    req.Symbol,
        UserId:    userID,
        Direction: req.Direction,
        Type:      req.Type,
        Price:     req.Price,
        Amount:    req.Amount,
    })
    if err != nil {
        return "", err
    }

    return resp.OrderId, nil
}
```

**业务流程**:
1. 调用 `req.OrderValid()` 验证参数（Direction 和 Type 必须非空）
2. 从上下文中获取用户 ID（由认证中间件存入）
3. 构建 gRPC 请求参数
4. 调用 `OrderClient.Add()` 创建订单
5. 返回订单 ID

**订单类型说明**:
| 类型 | 说明 |
|------|------|
| `MARKET_PRICE` | 市价单，按市场当前最优价格立即成交 |
| `LIMIT_PRICE` | 限价单，按指定价格挂单等待成交 |

**订单方向说明**:
| 方向 | 说明 |
|------|------|
| `BUY` | 买入方向，用基础币种购买交易币种 |
| `SELL` | 卖出方向，用交易币种换取基础币种 |

---

### 3.11 logic/current_logic.go - 当前订单业务逻辑

**文件职责**:
- 实现查询当前订单的业务逻辑
- 设置请求超时
- 封装分页结果

**关键方法**:

```go
func (l *CurrentLogic) Current(req *types.ExchangeReq) (*page.Result, error) {
    // 1. 设置 10 秒超时，防止长时间阻塞
    ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
    defer cancel()

    // 2. 从上下文获取已认证的用户 ID
    userID := middleware.UserIDFromContext(l.ctx)

    // 3. 调用 RPC 服务查询当前订单
    resp, err := l.svcCtx.OrderClient.FindOrderCurrent(ctx, &orderpb.OrderReq{
        Symbol:   req.Symbol,
        Page:     req.PageNo,
        PageSize: req.PageSize,
        UserId:   userID,
    })
    if err != nil {
        return nil, err
    }

    // 4. 将 RPC 响应转换为通用分页结果
    items := make([]any, len(resp.List))
    for i := range resp.List {
        items[i] = resp.List[i]
    }
    return page.New(items, req.PageNo, req.PageSize, resp.Total), nil
}
```

**超时设置**:
- 使用 `context.WithTimeout` 设置 10 秒超时
- 防止因 RPC 服务慢响应导致 HTTP 请求长时间阻塞
- 使用 `defer cancel()` 确保资源释放

**分页结果封装**:
- 使用 `page.New` 将 RPC 响应转换为统一的分页格式
- 自动计算总页数、是否有下一页等信息

---

### 3.12 logic/history_logic.go - 历史订单业务逻辑

**文件职责**:
- 实现查询历史订单的业务逻辑
- 与当前订单查询逻辑类似，但查询的是已完成/已取消的订单

**关键方法**:

```go
func (l *HistoryLogic) History(req *types.ExchangeReq) (*page.Result, error) {
    ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
    defer cancel()

    userID := middleware.UserIDFromContext(l.ctx)

    resp, err := l.svcCtx.OrderClient.FindOrderHistory(ctx, &orderpb.OrderReq{
        Symbol:   req.Symbol,
        Page:     req.PageNo,
        PageSize: req.PageSize,
        UserId:   userID,
    })
    if err != nil {
        return nil, err
    }

    items := make([]any, len(resp.List))
    for i := range resp.List {
        items[i] = resp.List[i]
    }
    return page.New(items, req.PageNo, req.PageSize, resp.Total), nil
}
```

**与 CurrentLogic 的区别**:
- 调用的 RPC 方法不同：`FindOrderHistory` vs `FindOrderCurrent`
- 查询的订单状态不同：已完成/已取消 vs 交易中

---

## 4. 请求处理流程

### 4.1 完整请求流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              客户端发起请求                                  │
│                   POST /exchange/order/add                                  │
│                   Header: x-auth-token: <JWT>                               │
│                   Body: {"symbol":"BTCUSDT","direction":"BUY",...}          │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           go-zero REST Server                                │
│                          (监听 0.0.0.0:8890)                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           AuthMiddleware (认证中间件)                        │
│                                                                             │
│  1. 从 Header 获取 x-auth-token                                            │
│  2. 验证 JWT 签名和有效期                                                   │
│  3. 解析用户 ID                                                            │
│  4. 将用户 ID 存入 context                                                 │
│                                                                             │
│  失败: 返回 {"code":4000,"message":"no login"}                             │
│  成功: 继续到 Handler                                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AddHandler                                      │
│                                                                             │
│  1. 解析请求参数到 ExchangeReq                                              │
│  2. 获取客户端 IP (httputil.ClientIP)                                      │
│  3. 创建 AddLogic 实例                                                      │
│  4. 调用 AddLogic.Add()                                                    │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AddLogic                                        │
│                                                                             │
│  1. 验证请求参数 (req.OrderValid)                                           │
│  2. 从 context 获取用户 ID                                                  │
│  3. 构建 gRPC 请求 (OrderReq)                                              │
│  4. 调用 OrderClient.Add()                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         gRPC 调用 exchange-rpc                               │
│                                                                             │
│  OrderClient.Add(ctx, &OrderReq{...})                                       │
│                                                                             │
│  通过 Etcd 服务发现，连接到 exchange-rpc 服务                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           exchange-rpc 服务                                  │
│                                                                             │
│  1. 验证用户交易状态 (调用 ucenter-rpc)                                     │
│  2. 验证交易对配置 (调用 market-rpc)                                        │
│  3. 验证钱包状态                                                            │
│  4. 创建订单记录                                                            │
│  5. 返回订单 ID                                                            │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              响应返回                                        │
│                                                                             │
│  HTTP 200 OK                                                                │
│  {                                                                          │
│    "code": 0,                                                               │
│    "message": "success",                                                    │
│    "data": {"orderId": "ORD123456"}                                         │
│  }                                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 时序图

```
客户端          AuthMiddleware      AddHandler       AddLogic       exchange-rpc
  │                  │                 │               │               │
  │ POST /order/add  │                 │               │               │
  │─────────────────>│                 │               │               │
  │                  │ 验证 JWT        │               │               │
  │                  │─────────┐       │               │               │
  │                  │         │       │               │               │
  │                  │<────────┘       │               │               │
  │                  │ 存入 userID     │               │               │
  │                  │────────────────>│               │               │
  │                  │                 │ 解析参数      │               │
  │                  │                 │ 获取 IP       │               │
  │                  │                 │──────────────>│               │
  │                  │                 │               │ 验证参数      │
  │                  │                 │               │─────────┐     │
  │                  │                 │               │         │     │
  │                  │                 │               │<────────┘     │
  │                  │                 │               │ gRPC 调用    │
  │                  │                 │               │──────────────>│
  │                  │                 │               │               │
  │                  │                 │               │               │ 创建订单
  │                  │                 │               │               │───────┐
  │                  │                 │               │               │       │
  │                  │                 │               │               │<──────┘
  │                  │                 │               │<──────────────│
  │                  │                 │<──────────────│               │
  │                  │<────────────────│               │               │
  │<─────────────────│                 │               │               │
  │   HTTP 200       │                 │               │               │
  │   {"code":0,...} │                 │               │               │
```

---

## 5. 与其他服务的调用关系

### 5.1 直接调用的服务

| 服务 | 调用方式 | 调用方法 | 用途 |
|------|----------|----------|------|
| exchange-rpc | gRPC | `Add` | 创建订单 |
| exchange-rpc | gRPC | `FindOrderCurrent` | 查询当前委托订单 |
| exchange-rpc | gRPC | `FindOrderHistory` | 查询历史订单 |

### 5.2 gRPC 接口定义

```protobuf
// exchange/order.proto
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

### 5.3 exchange-rpc 内部的服务调用

当 exchange-api 调用 exchange-rpc 创建订单时，exchange-rpc 会进一步调用其他服务:

```
exchange-rpc
     │
     ├──> ucenter-rpc.MemberClient.FindMemberById()
     │    查询会员信息，验证用户交易状态
     │
     ├──> market-rpc.MarketClient.FindSymbolInfo()
     │    查询交易对配置，验证是否可交易
     │
     └──> ucenter-rpc.AssetClient.FindWalletBySymbol()
          查询钱包信息，验证钱包是否被锁定
```

### 5.4 为什么需要调用这些服务

| 调用 | 原因 |
|------|------|
| **exchange-rpc** | 订单核心业务逻辑在 exchange-rpc 中实现，包括订单创建、查询、取消等。API 层只负责 HTTP 协议处理，业务逻辑由 RPC 服务处理，实现关注点分离。 |
| **ucenter-rpc**（exchange-rpc 调用） | 验证用户状态（是否被禁用交易）、查询用户钱包信息（是否被锁定）、验证用户余额是否充足。 |
| **market-rpc**（exchange-rpc 调用） | 验证交易对是否存在、是否开启交易、获取交易对配置（最小交易量、价格精度等）。 |

---

## 6. 配置说明

### 6.1 配置文件 (etc/exchange-api.yaml)

```yaml
# 服务名称
Name: exchange-api

# 监听地址，0.0.0.0 表示监听所有网卡
Host: 0.0.0.0

# 监听端口
Port: 8890

# JWT 认证配置
JWT:
  # JWT 签名密钥（生产环境应使用环境变量或配置中心）
  AccessSecret: "!@#$mscoin"
  # 令牌有效期（秒），604800 秒 = 7 天
  AccessExpire: 604800

# exchange-rpc 服务配置
ExchangeRPC:
  # Etcd 服务发现配置
  Etcd:
    Hosts:
      - etcd:2379    # Etcd 服务地址
    Key: exchange.rpc  # 服务名称键
```

### 6.2 配置项详解

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `Name` | string | exchange-api | 服务名称，用于日志和监控标识 |
| `Host` | string | 0.0.0.0 | 监听地址，0.0.0.0 监听所有网卡 |
| `Port` | int | 8890 | HTTP 服务端口 |
| `JWT.AccessSecret` | string | - | JWT 签名密钥，必须保密 |
| `JWT.AccessExpire` | int64 | 604800 | 令牌有效期（秒） |
| `ExchangeRPC.Etcd.Hosts` | []string | - | Etcd 服务地址列表 |
| `ExchangeRPC.Etcd.Key` | string | - | 服务发现键 |

### 6.3 环境变量配置（推荐）

生产环境建议使用环境变量覆盖敏感配置:

```bash
# 设置 JWT 密钥
export EXCHANGE_API_JWT_ACCESSSECRET="your-secret-key"

# 设置 Etcd 地址
export EXCHANGE_API_EXCHANGERPC_ETCD_HOSTS="etcd1:2379,etcd2:2379"
```

---

## 7. API 接口列表

### 7.1 新增订单

**请求**:
```
POST /exchange/order/add
Content-Type: application/json
x-auth-token: <JWT>

{
    "symbol": "BTCUSDT",
    "direction": "BUY",
    "type": "LIMIT_PRICE",
    "price": 50000.0,
    "amount": 0.1
}
```

**响应**:
```json
{
    "code": 0,
    "message": "success",
    "data": "ORD123456789"
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| symbol | string | 是 | 交易对符号，如 "BTCUSDT" |
| direction | string | 是 | 订单方向："BUY" 或 "SELL" |
| type | string | 是 | 订单类型："MARKET_PRICE" 或 "LIMIT_PRICE" |
| price | float64 | 限价单必填 | 订单价格 |
| amount | float64 | 是 | 订单数量 |
| useDiscount | float64 | 否 | 使用的折扣金额 |

### 7.2 查询当前订单

**请求**:
```
POST /exchange/order/current
Content-Type: application/json
x-auth-token: <JWT>

{
    "symbol": "BTCUSDT",
    "pageNo": 0,
    "pageSize": 10
}
```

**响应**:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "content": [
            {
                "id": 1,
                "orderId": "ORD123456",
                "symbol": "BTCUSDT",
                "baseSymbol": "USDT",
                "coinSymbol": "BTC",
                "direction": "BUY",
                "type": "LIMIT_PRICE",
                "price": 50000.0,
                "amount": 0.1,
                "tradedAmount": 0.05,
                "turnover": 2500.0,
                "status": "TRADING",
                "time": 1672531200
            }
        ],
        "totalElements": 1,
        "number": 0,
        "totalPages": 1,
        "hasNext": false,
        "isLast": true
    }
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| symbol | string | 否 | 交易对符号，为空时查询所有交易对 |
| pageNo | int64 | 否 | 页码，从 0 开始，默认 0 |
| pageSize | int64 | 否 | 每页记录数，默认 10 |

### 7.3 查询历史订单

**请求**:
```
POST /exchange/order/history
Content-Type: application/json
x-auth-token: <JWT>

{
    "symbol": "BTCUSDT",
    "pageNo": 0,
    "pageSize": 10
}
```

**响应**:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "content": [
            {
                "id": 1,
                "orderId": "ORD123456",
                "symbol": "BTCUSDT",
                "direction": "BUY",
                "type": "LIMIT_PRICE",
                "price": 50000.0,
                "amount": 0.1,
                "tradedAmount": 0.1,
                "turnover": 5000.0,
                "status": "COMPLETED",
                "time": 1672531200,
                "completedTime": 1672534800
            }
        ],
        "totalElements": 1,
        "number": 0,
        "totalPages": 1,
        "hasNext": false,
        "isLast": true
    }
}
```

### 7.4 错误响应

**认证失败**:
```json
{
    "code": 4000,
    "message": "no login",
    "data": null
}
```

**业务错误**:
```json
{
    "code": 500,
    "message": "余额不足",
    "data": null
}
```

### 7.5 订单状态说明

| 状态 | 说明 |
|------|------|
| TRADING | 交易中，订单正在撮合队列等待成交 |
| COMPLETED | 已完成，订单全部成交 |
| CANCELED | 已取消，用户主动取消 |
| OVERTIMED | 已超时，订单超过有效期被系统取消 |

---

## 8. 依赖的公共包

### 8.1 pkg/auth - JWT 认证

**文件**: `/pkg/auth/jwt.go`

**主要函数**:

| 函数 | 说明 |
|------|------|
| `GenerateUserToken(secret, issuedAt, expireSeconds, userID)` | 生成用户 JWT 令牌 |
| `ParseUserID(tokenString, secret)` | 解析令牌获取用户 ID |

**令牌 Claims**:
```json
{
    "exp": 1673136000,
    "iat": 1672531200,
    "userId": 12345
}
```

### 8.2 pkg/result - 统一响应格式

**文件**: `/pkg/result/result.go`

**主要结构体**:
```go
type Result struct {
    Code    int    `json:"code"`    // 状态码：0 成功，500 失败
    Message string `json:"message"` // 状态描述
    Data    any    `json:"data"`    // 业务数据
}
```

**主要方法**:
| 方法 | 说明 |
|------|------|
| `New()` | 创建空的结果对象 |
| `Success(data)` | 设置成功响应 |
| `Fail(code, message)` | 设置失败响应 |
| `Deal(data, error)` | 处理 `(data, error)` 模式 |

### 8.3 pkg/page - 分页结果

**文件**: `/pkg/page/page.go`

**主要结构体**:
```go
type Result struct {
    Content       []any `json:"content"`       // 当前页数据
    TotalElements int64 `json:"totalElements"` // 总记录数
    Number        int64 `json:"number"`        // 当前页码（从 0 开始）
    TotalPages    int64 `json:"totalPages"`    // 总页数
    HasNext       bool  `json:"hasNext"`       // 是否有下一页
    IsLast        bool  `json:"isLast"`        // 是否最后一页
}
```

### 8.4 pkg/httputil - HTTP 工具

**文件**: `/pkg/httputil/client_ip.go`

**主要函数**:
| 函数 | 说明 |
|------|------|
| `ClientIP(r *http.Request) string` | 从请求中提取客户端真实 IP |

**IP 获取优先级**:
1. `X-Real-IP` 请求头
2. `X-Forwarded-For` 请求头的第一个 IP
3. `RemoteAddr`

---

## 9. Dockerfile 说明

```dockerfile
# 构建阶段
ARG GO_IMAGE=golang:1.26.3
FROM ${GO_IMAGE} AS builder

WORKDIR /workspace

# 复制依赖文件
COPY go.mod go.sum go.work go.work.sum ./
RUN go mod download

# 复制源代码
COPY app ./app
COPY idl ./idl
COPY pkg ./pkg

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/server ./app/exchange/api

# 运行阶段
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# 复制二进制文件和配置文件
COPY --from=builder /out/server /app/server
COPY --from=builder /workspace/app/exchange/api/etc /app/etc

ENTRYPOINT ["/app/server"]
CMD ["-f", "/app/etc/exchange-api.yaml"]
```

**多阶段构建优势**:
1. 构建镜像大（包含 Go 工具链），但运行镜像小
2. 使用 `distroless` 基础镜像，减少攻击面
3. 静态编译，无需 CGO 依赖

---

## 10. 开发指南

### 10.1 本地运行

```bash
# 进入项目目录
cd /Volumes/移动卷宗/学习/go/mscoin_go/app/exchange/api

# 启动服务（需要先启动 Etcd 和 exchange-rpc）
go run . -f etc/exchange-api.yaml
```

### 10.2 构建 Docker 镜像

```bash
# 构建镜像
docker build -t exchange-api:latest -f Dockerfile ../..

# 运行容器
docker run -p 8890:8890 exchange-api:latest
```

### 10.3 添加新接口

1. 在 `internal/types/types.go` 添加请求/响应结构体
2. 在 `internal/handler/` 添加 Handler 文件
3. 在 `internal/logic/` 添加 Logic 文件
4. 在 `internal/handler/routes.go` 注册路由
5. 更新 `exchange-rpc` 的 proto 文件（如需新的 RPC 方法）

---

## 11. 注意事项

1. **JWT 密钥安全**: 生产环境切勿使用默认密钥，应使用环境变量或配置中心管理
2. **超时设置**: 查询类接口设置了 10 秒超时，创建订单接口未设置超时（依赖 RPC 层超时）
3. **CORS 配置**: 当前允许所有来源访问，生产环境应限制允许的域名
4. **IP 获取**: `X-Forwarded-For` 可被客户端伪造，生产环境应配合可信代理列表使用
5. **错误处理**: 所有错误都通过 `result.Deal()` 统一处理，确保响应格式一致
