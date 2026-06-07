# UCenter API 服务详细文档

## 目录

1. [服务概述](#1-服务概述)
2. [目录结构](#2-目录结构)
3. [配置说明](#3-配置说明)
4. [核心组件详解](#4-核心组件详解)
5. [API 接口列表](#5-api-接口列表)
6. [JWT 认证流程](#6-jwt-认证流程)
7. [请求处理流程](#7-请求处理流程)
8. [与其他服务的调用关系](#8-与其他服务的调用关系)
9. [公共包依赖](#9-公共包依赖)
10. [部署说明](#10-部署说明)

---

## 1. 服务概述

### 1.1 服务定位

`ucenter-api` 是 MSCoin 项目的**用户中心 HTTP API 服务**，作为用户与系统交互的前端入口，提供用户注册、登录、钱包管理、资产查询、提现等核心功能的 RESTful 接口。

### 1.2 核心职责

| 职责领域 | 具体功能 |
|---------|---------|
| 用户认证 | 用户注册、用户登录、登录状态检查 |
| 验证码服务 | 发送短信验证码（注册用、提现用） |
| 安全设置 | 查询用户安全认证状态（实名、手机、邮箱、资金密码等） |
| 资产管理 | 钱包余额查询、交易记录查询、充值地址管理 |
| 提现服务 | 提现申请、提现记录查询、提现币种信息查询 |

### 1.3 在整体架构中的位置

```
┌─────────────────────────────────────────────────────────────────┐
│                         前端应用层                                │
│                    (Web / iOS / Android)                        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      ucenter-api (本服务)                        │
│                         HTTP:8888                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Handler    │  │   Logic      │  │  Middleware  │          │
│  │   (HTTP层)   │──│   (业务层)   │──│   (认证层)   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
           │                    │                    │
           ▼                    ▼                    ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│   ucenter-rpc    │  │   market-rpc     │  │   其他微服务      │
│  (用户中心RPC)    │  │   (市场RPC)      │  │                  │
│                  │  │                  │  │                  │
│ - register       │  │ - coin info      │  │ - exchange       │
│ - login          │  │ - market data    │  │ - order          │
│ - member         │  │                  │  │ - ...            │
│ - asset          │  │                  │  │                  │
│ - withdraw       │  │                  │  │                  │
└──────────────────┘  └──────────────────┘  └──────────────────┘
           │                    │                    │
           ▼                    ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                      基础设施层                                   │
│              MySQL / Redis / Etcd / Kafka                        │
└─────────────────────────────────────────────────────────────────┘
```

### 1.4 技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| Web框架 | go-zero/rest | 高性能微服务框架 |
| RPC通信 | gRPC + Protobuf | 服务间通信 |
| 服务发现 | Etcd | 服务注册与发现 |
| 认证方案 | JWT (HS256) | 无状态认证 |
| 配置格式 | YAML | 声明式配置 |

---

## 2. 目录结构

```
/Volumes/移动卷宗/学习/go/mscoin_go/app/ucenter/api/
├── Dockerfile                    # Docker 构建文件
├── main.go                       # 服务入口文件
├── etc/                          # 配置文件目录
│   └── ucenter-api.yaml          # 服务配置文件
└── internal/                     # 内部实现（不对外暴露）
    ├── config/                   # 配置结构定义
    │   └── config.go             # 配置结构体定义
    ├── handler/                  # HTTP 处理器层
    │   ├── routes.go             # 路由注册
    │   ├── register_handler.go   # 注册处理器
    │   ├── login_handler.go      # 登录处理器
    │   ├── send_code_handler.go  # 发送验证码处理器
    │   ├── check_login_handler.go# 检查登录状态处理器
    │   ├── security_setting_handler.go    # 安全设置处理器
    │   ├── find_wallet_handler.go         # 查询钱包列表处理器
    │   ├── find_wallet_by_symbol_handler.go # 按币种查询钱包处理器
    │   ├── find_transaction_handler.go    # 查询交易记录处理器
    │   ├── reset_address_handler.go       # 重置充值地址处理器
    │   ├── send_withdraw_code_handler.go  # 发送提现验证码处理器
    │   ├── withdraw_code_handler.go       # 申请提现处理器
    │   ├── withdraw_record_handler.go     # 查询提现记录处理器
    │   └── query_withdraw_coin_handler.go # 查询可提现币种处理器
    ├── logic/                    # 业务逻辑层
    │   ├── register_logic.go     # 注册业务逻辑
    │   ├── login_logic.go        # 登录业务逻辑
    │   ├── send_code_logic.go    # 发送验证码业务逻辑
    │   ├── check_login_logic.go  # 检查登录状态业务逻辑
    │   ├── security_setting_logic.go       # 安全设置业务逻辑
    │   ├── find_wallet_logic.go            # 查询钱包列表业务逻辑
    │   ├── find_wallet_by_symbol_logic.go  # 按币种查询钱包业务逻辑
    │   ├── find_transaction_logic.go       # 查询交易记录业务逻辑
    │   ├── reset_address_logic.go          # 重置充值地址业务逻辑
    │   ├── send_withdraw_code_logic.go     # 发送提现验证码业务逻辑
    │   ├── withdraw_code_logic.go          # 申请提现业务逻辑
    │   ├── withdraw_record_logic.go        # 查询提现记录业务逻辑
    │   ├── query_withdraw_coin_logic.go    # 查询可提现币种业务逻辑
    │   └── *_test.go             # 单元测试文件
    ├── middleware/               # HTTP 中间件
    │   └── auth_middleware.go    # JWT 认证中间件
    ├── svc/                      # 服务上下文
    │   └── service_context.go    # 服务依赖容器
    └── types/                    # 类型定义
        └── types.go              # 请求/响应数据结构
```

### 2.1 目录职责说明

| 目录/文件 | 职责描述 |
|----------|---------|
| `main.go` | 服务启动入口，初始化配置、创建服务器、注册路由 |
| `etc/` | 存放服务配置文件，支持不同环境配置 |
| `internal/config/` | 定义配置结构体，与配置文件映射 |
| `internal/handler/` | HTTP 请求处理器，负责请求解析和响应格式化 |
| `internal/logic/` | 业务逻辑实现，调用 RPC 服务完成具体业务 |
| `internal/middleware/` | HTTP 中间件，提供认证等横切关注点 |
| `internal/svc/` | 服务上下文，管理所有依赖（RPC 客户端等） |
| `internal/types/` | API 层数据结构定义（请求体、响应体） |

### 2.2 分层架构说明

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Request                            │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Middleware Layer                          │
│              (auth_middleware.go)                            │
│                                                              │
│  职责：JWT Token 验证、用户身份注入 Context                    │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Handler Layer                            │
│              (*_handler.go)                                  │
│                                                              │
│  职责：                                                       │
│  - 解析 HTTP 请求参数（JSON Body / Form / Path）              │
│  - 参数验证                                                   │
│  - 调用 Logic 层                                              │
│  - 格式化响应（使用 result 包）                                │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Logic Layer                             │
│              (*_logic.go)                                    │
│                                                              │
│  职责：                                                       │
│  - 实现具体业务逻辑                                            │
│  - 调用 RPC 服务                                              │
│  - 数据转换和组装                                              │
│  - 返回处理结果                                                │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     RPC Services                             │
│           (ucenter-rpc / market-rpc)                         │
│                                                              │
│  职责：数据持久化、业务规则验证、跨服务协调                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. 配置说明

### 3.1 配置文件详解

配置文件路径：`etc/ucenter-api.yaml`

```yaml
# 服务基础配置
Name: ucenter-api          # 服务名称，用于日志和监控标识
Host: 0.0.0.0             # 监听地址，0.0.0.0 表示监听所有网卡
Port: 8888                # 监听端口

# JWT 认证配置
JWT:
  AccessSecret: "!@#$mscoin"   # JWT 签名密钥（生产环境应使用环境变量）
  AccessExpire: 604800         # Token 过期时间（秒），604800 = 7天

# ucenter-rpc 服务配置
UcenterRPC:
  Etcd:
    Hosts:
      - etcd:2379            # Etcd 服务地址
    Key: ucenter.rpc         # 服务发现 Key

# market-rpc 服务配置
MarketRPC:
  Etcd:
    Hosts:
      - etcd:2379            # Etcd 服务地址
    Key: market.rpc          # 服务发现 Key
```

### 3.2 配置结构体

配置结构体定义在 `internal/config/config.go`：

```go
// Config 是 ucenter-api 服务的完整配置结构
type Config struct {
    rest.RestConf                    // go-zero REST 服务器基础配置

    UcenterRPC marketconf.RpcClientConf  // 用户中心 RPC 服务配置
    MarketRPC marketconf.RpcClientConf   // 市场 RPC 服务配置
    JWT AuthConfig                        // JWT 认证配置
}

// AuthConfig 定义 JWT 认证相关的配置
type AuthConfig struct {
    AccessSecret string    // JWT 签名密钥
    AccessExpire int64     // Token 过期时间（秒）
}
```

### 3.3 配置项说明

| 配置项 | 类型 | 说明 | 默认值 |
|-------|------|------|--------|
| `Name` | string | 服务名称 | ucenter-api |
| `Host` | string | 监听地址 | 0.0.0.0 |
| `Port` | int | 监听端口 | 8888 |
| `JWT.AccessSecret` | string | JWT 签名密钥 | 必填 |
| `JWT.AccessExpire` | int64 | Token 有效期（秒） | 604800 (7天) |
| `UcenterRPC.Etcd.Hosts` | []string | Etcd 集群地址 | 必填 |
| `UcenterRPC.Etcd.Key` | string | 服务发现 Key | ucenter.rpc |
| `MarketRPC.Etcd.Hosts` | []string | Etcd 集群地址 | 必填 |
| `MarketRPC.Etcd.Key` | string | 服务发现 Key | market.rpc |

---

## 4. 核心组件详解

### 4.1 服务入口 (main.go)

#### 文件职责
`main.go` 是整个服务的启动入口，负责：
1. 解析命令行参数获取配置文件路径
2. 加载配置文件初始化配置结构
3. 创建 REST 服务器并配置 CORS
4. 初始化服务上下文（建立 RPC 连接）
5. 注册路由处理器
6. 启动 HTTP 服务器

#### 关键代码解析

```go
package main

import (
    "flag"
    "fmt"
    "net/http"

    "mscoin_go/app/ucenter/api/internal/config"
    "mscoin_go/app/ucenter/api/internal/handler"
    "mscoin_go/app/ucenter/api/internal/svc"

    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/rest"
)

// 命令行参数：指定配置文件路径
var configFile = flag.String("f", "etc/ucenter-api.yaml", "the config file")

func main() {
    flag.Parse()

    // 1. 加载配置文件
    var c config.Config
    conf.MustLoad(*configFile, &c)

    // 2. 创建 REST 服务器，配置 CORS
    server := rest.MustNewServer(
        c.RestConf,
        rest.WithCustomCors(func(header http.Header) {
            // 设置允许的请求头
            header.Set("Access-Control-Allow-Headers",
                "DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,"+
                "If-Modified-Since,Cache-Control,Content-Type,Authorization,"+
                "token,x-auth-token")
        }, nil, "*"),  // 允许所有来源
    )
    defer server.Stop()

    // 3. 初始化服务上下文（建立 RPC 连接）
    ctx := svc.NewServiceContext(c)

    // 4. 注册路由处理器
    handler.RegisterHandlers(server, ctx)

    // 5. 启动服务器
    fmt.Printf("Starting ucenter api server at %s:%d...\n", c.Host, c.Port)
    server.Start()
}
```

#### CORS 配置说明
服务配置了 CORS 跨域支持，允许前端应用从不同域名访问 API：
- `Access-Control-Allow-Headers`: 允许的请求头包括认证相关的 `Authorization`、`token`、`x-auth-token`
- 允许所有来源 (`*`)，生产环境建议配置具体域名

---

### 4.2 服务上下文 (service_context.go)

#### 文件职责
`service_context.go` 定义了服务运行时的依赖容器，负责：
1. 建立与 RPC 服务的连接
2. 初始化 JWT 认证中间件
3. 管理所有 RPC 客户端存根

#### ServiceContext 结构体

```go
// ServiceContext 是 ucenter-api 服务的运行时上下文容器
type ServiceContext struct {
    Config config.Config       // 服务配置
    Auth rest.Middleware       // JWT 认证中间件

    // RPC 客户端
    RegisterClient registerpb.RegisterClient   // 注册服务
    LoginClient loginpb.LoginClient             // 登录服务
    MemberClient memberpb.MemberClient          // 会员服务
    AssetClient assetpb.AssetClient             // 资产服务
    WithdrawClient withdrawpb.WithdrawClient    // 提现服务
    MarketClient marketpb.MarketClient          // 市场服务
}
```

#### 初始化流程

```go
func NewServiceContext(c config.Config) *ServiceContext {
    // 1. 建立 ucenter-rpc 客户端连接
    ucenterClient := zrpc.MustNewClient(c.UcenterRPC)
    ucenterConn := ucenterClient.Conn()

    // 2. 建立 market-rpc 客户端连接
    marketClient := zrpc.MustNewClient(c.MarketRPC)
    marketConn := marketClient.Conn()

    // 3. 返回初始化完成的服务上下文
    return &ServiceContext{
        Config:         c,
        Auth:           middleware.NewAuthMiddleware(c.JWT.AccessSecret).Handle,
        RegisterClient: registerpb.NewRegisterClient(ucenterConn),
        LoginClient:    loginpb.NewLoginClient(ucenterConn),
        MemberClient:   memberpb.NewMemberClient(ucenterConn),
        AssetClient:    assetpb.NewAssetClient(ucenterConn),
        WithdrawClient: withdrawpb.NewWithdrawClient(ucenterConn),
        MarketClient:   marketpb.NewMarketClient(marketConn),
    }
}
```

---

### 4.3 认证中间件 (auth_middleware.go)

#### 文件职责
`auth_middleware.go` 实现 JWT 认证中间件，是保护需要登录才能访问的接口的第一道防线。

#### 认证流程

```
┌─────────────────────────────────────────────────────────────────┐
│                      HTTP Request                               │
│                 Header: x-auth-token: xxx                       │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Auth Middleware                               │
│                                                                 │
│  1. 从请求头获取 x-auth-token                                    │
│  2. Token 为空？──> 返回错误码 4000 (未登录)                      │
│  3. 解析 Token 获取用户 ID                                       │
│  4. 解析失败？──> 返回错误码 4000 (未登录)                        │
│  5. 将用户 ID 注入 Context                                       │
│  6. 调用下一个处理器                                              │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Handler / Logic                              │
│                                                                 │
│  通过 middleware.UserIDFromContext(ctx) 获取用户 ID              │
└─────────────────────────────────────────────────────────────────┘
```

#### 核心代码

```go
package middleware

import (
    "context"
    "net/http"

    "mscoin_go/pkg/auth"
    "mscoin_go/pkg/result"

    "github.com/zeromicro/go-zero/rest/httpx"
)

// contextKey 是 context.Value 的键类型
type contextKey string

// userIDKey 是存储用户 ID 的 context 键
const userIDKey contextKey = "userId"

// AuthMiddleware 是 JWT 认证中间件结构
type AuthMiddleware struct {
    secret string  // JWT 签名密钥
}

// Handle 是中间件的处理函数
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. 从请求头获取 Token
        token := r.Header.Get("x-auth-token")
        if token == "" {
            failed := result.New()
            failed.Fail(4000, "no login")
            httpx.WriteJson(w, http.StatusOK, failed)
            return
        }

        // 2. 解析 Token 获取用户 ID
        userID, err := auth.ParseUserID(token, m.secret)
        if err != nil {
            failed := result.New()
            failed.Fail(4000, "no login")
            httpx.WriteJson(w, http.StatusOK, failed)
            return
        }

        // 3. 将用户 ID 注入 context
        ctx := context.WithValue(r.Context(), userIDKey, userID)
        next(w, r.WithContext(ctx))
    }
}

// UserIDFromContext 从 context 中获取已认证的用户 ID
func UserIDFromContext(ctx context.Context) int64 {
    value, _ := ctx.Value(userIDKey).(int64)
    return value
}

// WithUserID 将用户 ID 注入到 context 中（用于测试）
func WithUserID(ctx context.Context, userID int64) context.Context {
    return context.WithValue(ctx, userIDKey, userID)
}
```

#### 为什么选择错误码 4000
- 前端约定使用 4000 表示"未登录"状态
- 区别于 HTTP 401 状态码，保持响应体格式一致
- 不暴露 Token 具体问题（过期、签名错误等），统一处理为"需要重新登录"

---

### 4.4 类型定义 (types.go)

#### 文件职责
`types.go` 定义了所有 HTTP 接口的请求和响应数据结构，是 API 层与 Logic 层之间的数据契约。

#### 主要数据结构

##### 1. 验证码相关

```go
// CaptchaReq 是验证码验证请求数据
type CaptchaReq struct {
    Server string `json:"server"`  // 验证码服务服务器地址
    Token  string `json:"token"`   // 验证码响应令牌
}

// CodeRequest 是发送短信验证码的请求结构
type CodeRequest struct {
    Phone   string `json:"phone,optional"`   // 手机号码
    Country string `json:"country,optional"` // 国家代码
}

// CodeResponse 是发送验证码的响应结构
type CodeResponse struct{}
```

##### 2. 用户认证相关

```go
// Request 是通用的用户请求结构（主要用于注册）
type Request struct {
    Username     string      `json:"username,optional"`     // 用户名
    Password     string      `json:"password,optional"`     // 密码
    Captcha      *CaptchaReq `json:"captcha,optional"`      // 验证码
    Phone        string      `json:"phone,optional"`        // 手机号
    Promotion    string      `json:"promotion,optional"`    // 邀请码
    Code         string      `json:"code,optional"`         // 短信验证码
    Country      string      `json:"country,optional"`      // 国家代码
    SuperPartner string      `json:"superPartner,optional"` // 超级合伙人标识
    IP           string      `json:"ip,optional"`           // 客户端 IP
}

// LoginReq 是用户登录请求结构
type LoginReq struct {
    Username string      `json:"username"`           // 用户名
    Password string      `json:"password"`           // 密码
    Captcha  *CaptchaReq `json:"captcha,optional"`   // 验证码
    IP       string      `json:"ip,optional"`        // 客户端 IP
}

// LoginRes 是用户登录响应结构
type LoginRes struct {
    Username      string `json:"username"`       // 用户名
    Token         string `json:"token"`          // JWT Token
    MemberLevel   string `json:"memberLevel"`    // 会员等级
    RealName      string `json:"realName"`       // 真实姓名
    Country       string `json:"country"`        // 国家
    Avatar        string `json:"avatar"`         // 头像 URL
    PromotionCode string `json:"promotionCode"`  // 邀请码
    Id            int64  `json:"id"`             // 用户 ID
    LoginCount    int    `json:"loginCount"`     // 登录次数
    SuperPartner  string `json:"superPartner"`   // 超级合伙人标识
    MemberRate    int    `json:"memberRate"`     // 会员费率等级
}
```

##### 3. 资产相关

```go
// AssetReq 是资产查询请求结构
type AssetReq struct {
    CoinName  string `json:"coinName,optional" path:"coinName,optional"` // 币种名称
    IP        string `json:"ip,optional"`                                // 客户端 IP
    Unit      string `json:"unit,optional" form:"unit,optional"`         // 币种单位
    PageNo    int    `json:"pageNo,optional" form:"pageNo,optional"`     // 页码
    PageSize  int    `json:"pageSize,optional" form:"pageSize,optional"` // 每页数量
    StartTime string `json:"startTime,optional" form:"startTime,optional"` // 起始时间
    EndTime   string `json:"endTime,optional" form:"endTime,optional"`   // 结束时间
    Symbol    string `json:"symbol,optional" form:"symbol,optional"`     // 交易对
    Type      string `json:"type,optional" form:"type,optional"`         // 交易类型
}

// Coin 是币种信息结构
type Coin struct {
    Id                int32   `json:"id"`
    Name              string  `json:"name"`              // 币种名称
    CanAutoWithdraw   int32   `json:"canAutoWithdraw"`   // 是否支持自动提现
    CanRecharge       int32   `json:"canRecharge"`       // 是否开放充值
    CanTransfer       int32   `json:"canTransfer"`       // 是否支持转账
    CanWithdraw       int32   `json:"canWithdraw"`       // 是否开放提现
    CnyRate           float64 `json:"cnyRate"`           // 人民币汇率
    MaxWithdrawAmount float64 `json:"maxWithdrawAmount"` // 最大提现金额
    MinWithdrawAmount float64 `json:"minWithdrawAmount"` // 最小提现金额
    MinTxFee          float64 `json:"minTxFee"`          // 最小矿工费
    MaxTxFee          float64 `json:"maxTxFee"`          // 最大矿工费
    Unit              string  `json:"unit"`              // 币种单位
    UsdRate           float64 `json:"usdRate"`           // 美元汇率
    WithdrawThreshold float64 `json:"withdrawThreshold"` // 提现阈值
    // ... 其他字段
}

// MemberWallet 是会员钱包结构
type MemberWallet struct {
    Id             int64   `json:"id"`             // 钱包 ID
    Address        string  `json:"address"`        // 充值地址
    Balance        float64 `json:"balance"`        // 可用余额
    FrozenBalance  float64 `json:"frozenBalance"`  // 冻结余额
    ReleaseBalance float64 `json:"releaseBalance"` // 已释放余额
    IsLock         int32   `json:"isLock"`         // 锁定状态
    MemberId       int64   `json:"memberId"`       // 会员 ID
    Coin           Coin    `json:"coin"`           // 币种信息
    ToReleased     float64 `json:"toReleased"`     // 待释放余额
}

// MemberTransaction 是会员交易记录结构
type MemberTransaction struct {
    Id          int64   `json:"id"`          // 交易 ID
    Address     string  `json:"address"`     // 交易对方地址
    Amount      float64 `json:"amount"`      // 交易金额
    CreateTime  string  `json:"createTime"`  // 交易时间
    Fee         float64 `json:"fee"`         // 手续费
    Symbol      string  `json:"symbol"`      // 币种符号
    Type        string  `json:"type"`        // 交易类型
    // ... 其他字段
}
```

##### 4. 提现相关

```go
// WithdrawReq 是提现相关请求结构
type WithdrawReq struct {
    Unit       string  `json:"unit,optional" form:"unit,optional"`       // 币种单位
    Address    string  `json:"address,optional" form:"address,optional"` // 提现地址
    Amount     float64 `json:"amount,optional" form:"amount,optional"`   // 提现金额
    Fee        float64 `json:"fee,optional" form:"fee,optional"`         // 矿工费
    JyPassword string  `json:"jyPassword,optional" form:"jyPassword,optional"` // 资金密码
    Code       string  `json:"code,optional" form:"code,optional"`       // 短信验证码
    Page       int     `json:"page,optional" form:"page,optional"`       // 页码
    PageSize   int     `json:"pageSize,optional" form:"pageSize,optional"` // 每页数量
}

// WithdrawWalletInfo 是提现钱包信息聚合结构
type WithdrawWalletInfo struct {
    Unit            string          `json:"unit"`            // 币种单位
    Threshold       float64         `json:"threshold"`       // 提现阈值
    MinAmount       float64         `json:"minAmount"`       // 最小提现金额
    MaxAmount       float64         `json:"maxAmount"`       // 最大提现金额
    MinTxFee        float64         `json:"minTxFee"`        // 最小矿工费
    MaxTxFee        float64         `json:"maxTxFee"`        // 最大矿工费
    NameCn          string          `json:"nameCn"`          // 中文名称
    Name            string          `json:"name"`            // 英文名称
    Balance         float64         `json:"balance"`         // 用户余额
    CanAutoWithdraw string          `json:"canAutoWithdraw"` // 是否支持自动提现
    WithdrawScale   int32           `json:"withdrawScale"`   // 提现精度
    AccountType     int32           `json:"accountType"`     // 账户类型
    Addresses       []AddressSimple `json:"addresses"`       // 已保存地址列表
}

// WithdrawRecord 是提现记录结构
type WithdrawRecord struct {
    Id                int64   `json:"id"`                // 提现记录 ID
    MemberId          int64   `json:"memberId"`          // 会员 ID
    Coin              Coin    `json:"coin"`              // 币种信息
    TotalAmount       float64 `json:"totalAmount"`       // 提现总金额
    Fee               float64 `json:"fee"`               // 手续费
    ArrivedAmount     float64 `json:"arrivedAmount"`     // 到账金额
    Address           string  `json:"address"`           // 提现地址
    TransactionNumber string  `json:"transactionNumber"` // 区块链交易哈希
    Status            int32   `json:"status"`            // 提现状态
    CreateTime        string  `json:"createTime"`        // 创建时间
    // ... 其他字段
}
```

##### 5. 安全设置相关

```go
// MemberSecurity 是会员安全设置信息
type MemberSecurity struct {
    Username             string `json:"username"`             // 用户名
    Id                   int64  `json:"id"`                   // 会员 ID
    CreateTime           string `json:"createTime"`           // 注册时间
    RealVerified         string `json:"realVerified"`         // 实名认证状态
    EmailVerified        string `json:"emailVerified"`        // 邮箱认证状态
    PhoneVerified        string `json:"phoneVerified"`        // 手机认证状态
    FundsVerified        string `json:"fundsVerified"`        // 资金密码设置状态
    MobilePhone          string `json:"mobilePhone"`          // 已认证手机号
    Email                string `json:"email"`                // 已认证邮箱
    RealName             string `json:"realName"`             // 已认证真实姓名
    IdCard               string `json:"idCard"`               // 身份证号（脱敏）
    Avatar               string `json:"avatar"`               // 头像 URL
    AccountVerified      string `json:"accountVerified"`      // 账户验证状态
    RealNameRejectReason string `json:"realNameRejectReason"` // 实名拒绝原因
    RealAuditing         string `json:"realAuditing"`         // 实名审核中状态
    LoginVerified        string `json:"loginVerified"`        // 登录验证状态
}
```

---

## 5. API 接口列表

### 5.1 公开接口（无需认证）

| 接口路径 | 方法 | 功能描述 | Handler | Logic |
|---------|------|---------|---------|-------|
| `/uc/mobile/code` | POST | 发送短信验证码 | SendCodeHandler | SendCodeLogic |
| `/uc/register/phone` | POST | 手机号注册 | RegisterHandler | RegisterLogic |
| `/uc/check/login` | POST | 检查登录状态 | CheckLoginHandler | CheckLoginLogic |
| `/uc/login` | POST | 用户登录 | LoginHandler | LoginLogic |

### 5.2 认证接口（需要 JWT Token）

| 接口路径 | 方法 | 功能描述 | Handler | Logic |
|---------|------|---------|---------|-------|
| `/uc/approve/security/setting` | POST | 查询安全设置 | SecuritySettingHandler | SecuritySettingLogic |
| `/uc/asset/transaction/all` | POST | 查询交易记录 | FindTransactionHandler | FindTransactionLogic |
| `/uc/asset/wallet` | POST | 查询所有钱包 | FindWalletHandler | FindWalletLogic |
| `/uc/asset/wallet/reset-address` | POST | 重置充值地址 | ResetAddressHandler | ResetAddressLogic |
| `/uc/asset/wallet/:coinName` | POST | 查询单个钱包 | FindWalletBySymbolHandler | FindWalletBySymbolLogic |
| `/uc/mobile/withdraw/code` | POST | 发送提现验证码 | SendWithdrawCodeHandler | SendWithdrawCodeLogic |
| `/uc/withdraw/apply/code` | POST | 申请提现 | WithdrawCodeHandler | WithdrawCodeLogic |
| `/uc/withdraw/record` | POST | 查询提现记录 | WithdrawRecordHandler | WithdrawRecordLogic |
| `/uc/withdraw/support/coin/info` | POST | 查询可提现币种 | QueryWithdrawCoinHandler | QueryWithdrawCoinLogic |

### 5.3 接口详细说明

#### 5.3.1 发送短信验证码

**请求**
```
POST /uc/mobile/code
Content-Type: application/json

{
    "phone": "13800138000",
    "country": "CN"
}
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

**处理流程**
1. Handler 解析请求参数（手机号、国家代码）
2. 调用 Logic 层的 SendCode 方法
3. Logic 调用 ucenter-rpc RegisterClient.SendCode
4. RPC 服务验证手机号格式、调用短信服务发送验证码

---

#### 5.3.2 手机号注册

**请求**
```
POST /uc/register/phone
Content-Type: application/json

{
    "username": "testuser",
    "password": "password123",
    "phone": "13800138000",
    "code": "123456",
    "country": "CN",
    "captcha": {
        "server": "https://captcha.example.com",
        "token": "captcha-token"
    },
    "promotion": "INVITE123"
}
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

**处理流程**
1. Handler 验证验证码存在
2. 获取客户端 IP
3. 调用 Logic 层的 Register 方法
4. Logic 调用 ucenter-rpc RegisterClient.RegisterByPhone
5. RPC 服务验证短信验证码、创建会员账号、建立邀请关系

---

#### 5.3.3 用户登录

**请求**
```
POST /uc/login
Content-Type: application/json

{
    "username": "testuser",
    "password": "password123",
    "captcha": {
        "server": "https://captcha.example.com",
        "token": "captcha-token"
    }
}
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "username": "testuser",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "memberLevel": "普通会员",
        "realName": "张三",
        "country": "CN",
        "avatar": "https://example.com/avatar.png",
        "promotionCode": "PROMO123",
        "id": 10001,
        "loginCount": 5,
        "superPartner": "0",
        "memberRate": 1
    }
}
```

**处理流程**
1. Handler 验证验证码存在
2. 获取客户端 IP
3. 调用 Logic 层的 Login 方法
4. Logic 调用 ucenter-rpc LoginClient.Login
5. RPC 服务验证用户名密码、生成 JWT Token、记录登录日志
6. Logic 转换 RPC 响应为 API 响应格式

---

#### 5.3.4 检查登录状态

**请求**
```
POST /uc/check/login
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": true
}
```

**处理流程**
1. Handler 从请求头获取 JWT Token
2. 调用 Logic 层的 CheckLogin 方法
3. Logic 使用 auth.ParseUserID 解析 Token
4. 返回 Token 是否有效（true/false）

---

#### 5.3.5 查询安全设置

**请求**
```
POST /uc/approve/security/setting
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "username": "testuser",
        "id": 10001,
        "createTime": "2024-01-01 12:00:00",
        "realVerified": "true",
        "emailVerified": "false",
        "phoneVerified": "true",
        "fundsVerified": "true",
        "mobilePhone": "138****8000",
        "email": "",
        "realName": "张三",
        "idCard": "11********",
        "avatar": "https://example.com/avatar.png",
        "accountVerified": "true",
        "loginVerified": "true",
        "realAuditing": "false",
        "realNameRejectReason": ""
    }
}
```

**处理流程**
1. Auth 中间件验证 JWT Token，注入用户 ID 到 Context
2. Handler 调用 Logic 层的 FindSecuritySetting 方法
3. Logic 从 Context 获取用户 ID
4. 调用 ucenter-rpc MemberClient.FindMemberById
5. 转换响应格式，设置各认证状态
6. 对敏感信息（如身份证号）进行脱敏处理

---

#### 5.3.6 查询钱包列表

**请求**
```
POST /uc/asset/wallet
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "id": 1,
            "address": "0x1234...abcd",
            "balance": 1.5,
            "frozenBalance": 0.1,
            "releaseBalance": 0,
            "isLock": 0,
            "memberId": 10001,
            "coin": {
                "id": 1,
                "name": "Bitcoin",
                "unit": "BTC",
                "canWithdraw": 0,
                "canRecharge": 0,
                "maxWithdrawAmount": 10,
                "minWithdrawAmount": 0.001
            },
            "toReleased": 0
        }
    ]
}
```

**处理流程**
1. Auth 中间件验证 JWT Token
2. Handler 调用 Logic 层的 FindWallet 方法
3. Logic 从 Context 获取用户 ID
4. 调用 ucenter-rpc AssetClient.FindWallet
5. 使用 copier 转换响应格式

---

#### 5.3.7 按币种查询钱包

**请求**
```
POST /uc/asset/wallet/BTC
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "address": "0x1234...abcd",
        "balance": 1.5,
        "frozenBalance": 0.1,
        "coin": {
            "id": 1,
            "name": "Bitcoin",
            "unit": "BTC"
        }
    }
}
```

**处理流程**
1. Handler 从 URL 路径解析币种名称
2. 调用 Logic 层的 FindWalletBySymbol 方法
3. Logic 调用 ucenter-rpc AssetClient.FindWalletBySymbol
4. 转换响应格式

---

#### 5.3.8 查询交易记录

**请求**
```
POST /uc/asset/transaction/all
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/x-www-form-urlencoded

pageNo=1&pageSize=10&startTime=2024-01-01&endTime=2024-12-31&symbol=BTC&type=DEPOSIT
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "content": [
            {
                "id": 1,
                "address": "0x1234...abcd",
                "amount": 0.5,
                "createTime": "2024-06-01 12:00:00",
                "fee": 0.0001,
                "symbol": "BTC",
                "type": "DEPOSIT"
            }
        ],
        "totalElements": 100,
        "number": 0,
        "totalPages": 10,
        "hasNext": true,
        "isLast": false
    }
}
```

**处理流程**
1. Handler 从表单解析查询参数
2. 调用 Logic 层的 FindTransaction 方法
3. Logic 处理分页参数默认值
4. 调用 ucenter-rpc AssetClient.FindTransaction
5. 使用 page.New 构建分页响应

---

#### 5.3.9 重置充值地址

**请求**
```
POST /uc/asset/wallet/reset-address
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/x-www-form-urlencoded

unit=BTC
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": ""
}
```

**处理流程**
1. Handler 从表单解析参数，获取客户端 IP
2. 调用 Logic 层的 ResetAddress 方法
3. Logic 调用 ucenter-rpc AssetClient.ResetAddress
4. RPC 服务生成新充值地址、更新钱包记录

---

#### 5.3.10 发送提现验证码

**请求**
```
POST /uc/mobile/withdraw/code
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": "success"
}
```

**处理流程**
1. Auth 中间件验证 JWT Token
2. Handler 调用 Logic 层的 SendCode 方法
3. Logic 先调用 MemberClient.FindMemberById 获取用户手机号
4. 然后调用 WithdrawClient.SendCode 发送验证码到用户手机

---

#### 5.3.11 申请提现

**请求**
```
POST /uc/withdraw/apply/code
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/x-www-form-urlencoded

unit=USDT&address=0x1234...abcd&amount=100&fee=1&jyPassword=123456&code=123456
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": "success"
}
```

**处理流程**
1. Handler 从表单解析提现参数
2. 调用 Logic 层的 WithdrawCode 方法
3. Logic 调用 ucenter-rpc WithdrawClient.WithdrawCode
4. RPC 服务执行多重验证：
   - 资金密码验证
   - 短信验证码验证
   - 余额检查
   - 提现限额检查
   - 钱包状态检查
5. 创建提现工单，冻结提现金额

---

#### 5.3.12 查询提现记录

**请求**
```
POST /uc/withdraw/record
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/x-www-form-urlencoded

page=1&pageSize=10
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "content": [
            {
                "id": 1,
                "memberId": 10001,
                "coin": {...},
                "totalAmount": 100,
                "fee": 1,
                "arrivedAmount": 99,
                "address": "0x1234...abcd",
                "transactionNumber": "0xabcd...",
                "status": 2,
                "createTime": "2024-06-01 12:00:00"
            }
        ],
        "totalElements": 50,
        "number": 0,
        "totalPages": 5,
        "hasNext": true,
        "isLast": false
    }
}
```

**处理流程**
1. Handler 从表单解析分页参数
2. 调用 Logic 层的 Record 方法
3. Logic 调用 ucenter-rpc WithdrawClient.WithdrawRecord
4. 使用 page.New 构建分页响应

---

#### 5.3.13 查询可提现币种信息

**请求**
```
POST /uc/withdraw/support/coin/info
x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "unit": "USDT",
            "threshold": 0,
            "minAmount": 10,
            "maxAmount": 10000,
            "minTxFee": 1,
            "maxTxFee": 10,
            "nameCn": "泰达币",
            "name": "Tether",
            "balance": 1000,
            "canAutoWithdraw": "true",
            "withdrawScale": 6,
            "accountType": 1,
            "addresses": [
                {"remark": "主钱包", "address": "0x1234...abcd"}
            ]
        }
    ]
}
```

**处理流程**
1. Auth 中间件验证 JWT Token
2. Handler 调用 Logic 层的 QueryWithdrawCoin 方法
3. Logic 执行数据聚合：
   - 调用 market-rpc 获取币种列表
   - 调用 ucenter-rpc AssetClient 获取钱包余额
   - 遍历钱包，调用 WithdrawClient.FindAddressByCoinId 获取地址簿
4. 组装聚合响应

---

## 6. JWT 认证流程

### 6.1 Token 生成

Token 在用户登录成功后由 `ucenter-rpc` 服务生成：

```go
// Token 生成逻辑（在 pkg/auth/jwt.go）
func GenerateUserToken(secret string, issuedAt time.Time, expireSeconds int64, userID int64) (string, error) {
    claims := jwt.MapClaims{
        "exp":    issuedAt.Unix() + expireSeconds,  // 过期时间
        "iat":    issuedAt.Unix(),                   // 签发时间
        "userId": userID,                            // 用户 ID
    }

    token := jwt.New(jwt.SigningMethodHS256)
    token.Claims = claims

    return token.SignedString([]byte(secret))
}
```

### 6.2 Token 验证

Token 验证在 API 层的认证中间件中进行：

```go
// Token 解析逻辑（在 pkg/auth/jwt.go）
func ParseUserID(tokenString string, secret string) (int64, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // 验证签名算法
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil
    })
    if err != nil {
        return 0, err
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return 0, errors.New("invalid token")
    }

    // 提取 userId
    userID, ok := claims["userId"].(float64)
    if !ok {
        return 0, errors.New("userId claim missing")
    }

    // 验证过期时间
    expireAt, ok := claims["exp"].(float64)
    if !ok {
        return 0, errors.New("exp claim missing")
    }
    if int64(expireAt) <= time.Now().Unix() {
        return 0, errors.New("token expired")
    }

    return int64(userID), nil
}
```

### 6.3 认证流程图

```
┌─────────────────────────────────────────────────────────────────────┐
│                          用户登录                                    │
│                                                                     │
│  1. 用户提交用户名密码                                               │
│  2. ucenter-rpc 验证成功                                            │
│  3. 生成 JWT Token:                                                 │
│     {                                                               │
│       "exp": 1717987200,    // 过期时间                              │
│       "iat": 1717382400,    // 签发时间                              │
│       "userId": 10001       // 用户 ID                               │
│     }                                                               │
│  4. 返回 Token 给前端                                               │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       前端存储 Token                                 │
│                                                                     │
│  - localStorage / sessionStorage                                    │
│  - 后续请求在 x-auth-token 头部携带 Token                            │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      访问受保护接口                                   │
│                                                                     │
│  POST /uc/asset/wallet                                              │
│  x-auth-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...              │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Auth 中间件验证                                   │
│                                                                     │
│  1. 从请求头获取 x-auth-token                                        │
│  2. Token 为空？──> 返回 4000 (未登录)                               │
│  3. 解析 Token 验证签名                                              │
│  4. 检查是否过期                                                     │
│  5. 提取 userId                                                      │
│  6. 将 userId 注入 Context                                           │
│  7. 继续处理请求                                                     │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Logic 层处理                                    │
│                                                                     │
│  userID := middleware.UserIDFromContext(ctx)                        │
│  // 使用 userID 进行业务处理                                         │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.4 安全考虑

| 安全措施 | 说明 |
|---------|------|
| HS256 签名 | 使用 HMAC-SHA256 对称加密算法 |
| 过期时间 | Token 默认有效期 7 天，过期后需重新登录 |
| 密钥保护 | 生产环境密钥应从环境变量或密钥管理服务获取 |
| 不暴露细节 | Token 无效时统一返回"未登录"，不暴露具体原因 |
| Context 隔离 | 用户 ID 通过 Context 传递，避免全局状态 |

---

## 7. 请求处理流程

### 7.1 标准请求处理流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                        HTTP 请求到达                                 │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      路由匹配 (routes.go)                            │
│                                                                     │
│  根据请求路径和方法匹配对应的 Handler                                  │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    中间件处理 (如需要)                                │
│                                                                     │
│  - Auth Middleware: JWT Token 验证                                  │
│  - 用户 ID 注入 Context                                              │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Handler 处理                                    │
│                                                                     │
│  1. 解析请求参数:                                                    │
│     - httpx.ParseJsonBody(r, &req)  // JSON Body                   │
│     - httpx.ParseForm(r, &req)      // 表单数据                    │
│     - httpx.ParsePath(r, &req)      // URL 路径参数                │
│                                                                     │
│  2. 参数验证（如验证码存在性检查）                                     │
│                                                                     │
│  3. 获取附加信息（如客户端 IP）                                       │
│                                                                     │
│  4. 创建 Logic 实例并调用业务方法                                     │
│     resp, err := logic.NewXxxLogic(ctx, svcCtx).Xxx(&req)          │
│                                                                     │
│  5. 格式化响应                                                       │
│     httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))  │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Logic 处理                                      │
│                                                                     │
│  1. 从 Context 获取用户 ID（如需要）                                  │
│     userID := middleware.UserIDFromContext(ctx)                     │
│                                                                     │
│  2. 设置 RPC 调用超时                                                 │
│     ctx, cancel := context.WithTimeout(ctx, 5*time.Second)         │
│                                                                     │
│  3. 调用 RPC 服务                                                     │
│     resp, err := l.svcCtx.XxxClient.Xxx(ctx, &pb.Req{...})         │
│                                                                     │
│  4. 数据转换和组装                                                    │
│     - 使用 copier.Copy 进行类型转换                                  │
│     - 使用 page.New 构建分页响应                                     │
│                                                                     │
│  5. 返回结果                                                         │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      响应返回                                        │
│                                                                     │
│  {                                                                  │
│    "code": 0,           // 0 成功, 500 失败, 4000 未登录            │
│    "message": "success",                                            │
│    "data": {...}        // 业务数据                                  │
│  }                                                                  │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.2 登录流程详解

```
用户                  前端                  API                  RPC
 │                    │                    │                    │
 │  输入用户名密码     │                    │                    │
 │ ─────────────────> │                    │                    │
 │                    │                    │                    │
 │                    │  POST /uc/login    │                    │
 │                    │  {username,pwd,    │                    │
 │                    │   captcha}         │                    │
 │                    │ ─────────────────> │                    │
 │                    │                    │                    │
 │                    │                    │ 验证验证码          │
 │                    │                    │ 获取客户端 IP       │
 │                    │                    │                    │
 │                    │                    │  Login RPC         │
 │                    │                    │ ─────────────────> │
 │                    │                    │                    │
 │                    │                    │                    │ 验证验证码
 │                    │                    │                    │ 验证用户名密码
 │                    │                    │                    │ 生成 JWT Token
 │                    │                    │                    │ 记录登录日志
 │                    │                    │                    │
 │                    │                    │  {token, userInfo} │
 │                    │                    │ <───────────────── │
 │                    │                    │                    │
 │                    │                    │ 转换响应格式        │
 │                    │                    │                    │
 │                    │  {code:0,          │                    │
 │                    │   data:{token,...}}│                    │
 │                    │ <───────────────── │                    │
 │                    │                    │                    │
 │  存储 Token        │                    │                    │
 │ <───────────────── │                    │                    │
 │                    │                    │                    │
```

### 7.3 提现流程详解

```
用户                  前端                  API                  RPC
 │                    │                    │                    │
 │  点击"获取验证码"   │                    │                    │
 │ ─────────────────> │                    │                    │
 │                    │                    │                    │
 │                    │ POST /uc/mobile/   │                    │
 │                    │ withdraw/code      │                    │
 │                    │ x-auth-token: xxx  │                    │
 │                    │ ─────────────────> │                    │
 │                    │                    │                    │
 │                    │                    │ Auth 中间件验证     │
 │                    │                    │ 获取用户手机号      │
 │                    │                    │                    │
 │                    │                    │  SendCode RPC      │
 │                    │                    │ ─────────────────> │
 │                    │                    │                    │
 │                    │                    │                    │ 发送短信验证码
 │                    │                    │                    │
 │                    │  {code:0}          │                    │
 │                    │ <───────────────── │                    │
 │                    │                    │                    │
 │  收到短信验证码     │                    │                    │
 │ <───────────────── │                    │                    │
 │                    │                    │                    │
 │  填写提现信息       │                    │                    │
 │  提交提现申请       │                    │                    │
 │ ─────────────────> │                    │                    │
 │                    │                    │                    │
 │                    │ POST /uc/withdraw/ │                    │
 │                    │ apply/code         │                    │
 │                    │ {unit,amount,addr, │                    │
 │                    │  jyPassword,code}  │                    │
 │                    │ ─────────────────> │                    │
 │                    │                    │                    │
 │                    │                    │  WithdrawCode RPC  │
 │                    │                    │ ─────────────────> │
 │                    │                    │                    │
 │                    │                    │                    │ 验证资金密码
 │                    │                    │                    │ 验证短信验证码
 │                    │                    │                    │ 检查余额
 │                    │                    │                    │ 检查限额
 │                    │                    │                    │ 创建提现工单
 │                    │                    │                    │ 冻结金额
 │                    │                    │                    │
 │                    │  {code:0}          │                    │
 │                    │ <───────────────── │                    │
 │                    │                    │                    │
 │  提现申请成功       │                    │                    │
 │ <───────────────── │                    │                    │
 │                    │                    │                    │
```

---

## 8. 与其他服务的调用关系

### 8.1 服务依赖图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         ucenter-api                                  │
│                           (本服务)                                    │
└─────────────────────────────────────────────────────────────────────┘
          │                    │                    │
          │ gRPC               │ gRPC               │
          ▼                    ▼                    ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│   ucenter-rpc    │  │   market-rpc     │  │    Etcd          │
│                  │  │                  │  │                  │
│ 注册服务         │  │ 币种信息         │  │ 服务发现         │
│ ├─ Register      │  │ ├─ FindAllCoin   │  │                  │
│ └─ SendCode      │  │ └─ FindCoin      │  │                  │
│                  │  │                  │  │                  │
│ 登录服务         │  │                  │  │                  │
│ └─ Login         │  │                  │  │                  │
│                  │  │                  │  │                  │
│ 会员服务         │  │                  │  │                  │
│ └─ FindMemberById│  │                  │  │                  │
│                  │  │                  │  │                  │
│ 资产服务         │  │                  │  │                  │
│ ├─ FindWallet    │  │                  │  │                  │
│ ├─ FindWalletBy  │  │                  │  │                  │
│ │   Symbol       │  │                  │  │                  │
│ ├─ FindTransaction│  │                  │  │                  │
│ └─ ResetAddress  │  │                  │  │                  │
│                  │  │                  │  │                  │
│ 提现服务         │  │                  │  │                  │
│ ├─ SendCode      │  │                  │  │                  │
│ ├─ WithdrawCode  │  │                  │  │                  │
│ ├─ WithdrawRecord│  │                  │  │                  │
│ └─ FindAddressBy │  │                  │  │                  │
│    CoinId        │  │                  │  │                  │
└──────────────────┘  └──────────────────┘  └──────────────────┘
```

### 8.2 RPC 调用关系表

| API 接口 | 调用的 RPC 服务 | RPC 方法 |
|---------|----------------|---------|
| POST /uc/mobile/code | ucenter-rpc (register) | RegisterClient.SendCode |
| POST /uc/register/phone | ucenter-rpc (register) | RegisterClient.RegisterByPhone |
| POST /uc/login | ucenter-rpc (login) | LoginClient.Login |
| POST /uc/check/login | 无（本地验证） | - |
| POST /uc/approve/security/setting | ucenter-rpc (member) | MemberClient.FindMemberById |
| POST /uc/asset/wallet | ucenter-rpc (asset) | AssetClient.FindWallet |
| POST /uc/asset/wallet/:coinName | ucenter-rpc (asset) | AssetClient.FindWalletBySymbol |
| POST /uc/asset/transaction/all | ucenter-rpc (asset) | AssetClient.FindTransaction |
| POST /uc/asset/wallet/reset-address | ucenter-rpc (asset) | AssetClient.ResetAddress |
| POST /uc/mobile/withdraw/code | ucenter-rpc (member) | MemberClient.FindMemberById |
| | ucenter-rpc (withdraw) | WithdrawClient.SendCode |
| POST /uc/withdraw/apply/code | ucenter-rpc (withdraw) | WithdrawClient.WithdrawCode |
| POST /uc/withdraw/record | ucenter-rpc (withdraw) | WithdrawClient.WithdrawRecord |
| POST /uc/withdraw/support/coin/info | market-rpc | MarketClient.FindAllCoin |
| | ucenter-rpc (asset) | AssetClient.FindWallet |
| | ucenter-rpc (withdraw) | WithdrawClient.FindAddressByCoinId |

### 8.3 数据聚合示例

`QueryWithdrawCoin` 接口是典型的数据聚合场景，需要从多个 RPC 服务获取数据：

```go
func (l *QueryWithdrawCoinLogic) QueryWithdrawCoin() ([]*types.WithdrawWalletInfo, error) {
    // 1. 从 market-rpc 获取币种配置
    coinList, err := l.svcCtx.MarketClient.FindAllCoin(ctx, &marketpb.MarketReq{})

    // 2. 从 ucenter-rpc 获取用户钱包
    walletList, err := l.svcCtx.AssetClient.FindWallet(ctx, &assetpb.AssetReq{UserId: userID})

    // 3. 遍历钱包，获取每个币种的提现地址
    for _, wallet := range walletList.GetList() {
        addressList, err := l.svcCtx.WithdrawClient.FindAddressByCoinId(ctx, &withdrawpb.WithdrawReq{
            UserId: userID,
            CoinId: int64(coin.Id),
        })
        // 聚合数据...
    }

    return resp, nil
}
```

---

## 9. 公共包依赖

### 9.1 依赖的公共包

| 包路径 | 功能说明 | 主要导出 |
|-------|---------|---------|
| `mscoin_go/pkg/auth` | JWT 认证 | GenerateUserToken, ParseUserID |
| `mscoin_go/pkg/result` | 统一响应格式 | Result, New, Success, Fail, Deal |
| `mscoin_go/pkg/page` | 分页响应 | Result, New |
| `mscoin_go/pkg/httputil` | HTTP 工具 | ClientIP |
| `mscoin_go/pkg/passwordx` | 密码加密 | - |

### 9.2 统一响应格式 (result 包)

```go
// 所有 API 响应使用统一的信封格式
type Result struct {
    Code    int    `json:"code"`    // 0 成功, 500 失败, 4000 未登录
    Message string `json:"message"` // 状态描述
    Data    any    `json:"data"`    // 业务数据
}

// 使用示例
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := service.GetUser(userId)
    httpx.OkJsonCtx(r.Context(), w, result.New().Deal(user, err))
}
```

### 9.3 分页响应格式 (page 包)

```go
// 分页响应结构
type Result struct {
    Content       []any `json:"content"`       // 数据列表
    TotalElements int64 `json:"totalElements"` // 总记录数
    Number        int64 `json:"number"`        // 当前页码（从0开始）
    TotalPages    int64 `json:"totalPages"`    // 总页数
    HasNext       bool  `json:"hasNext"`       // 是否有下一页
    IsLast        bool  `json:"isLast"`        // 是否最后一页
}

// 使用示例
func (l *Logic) ListTransactions(req *types.AssetReq) (*page.Result, error) {
    records, total := rpc.FindTransactions(...)
    items := make([]any, len(records))
    for i, item := range records {
        items[i] = item
    }
    return page.New(items, pageNo, pageSize, total), nil
}
```

---

## 10. 部署说明

### 10.1 Docker 部署

服务提供了 Dockerfile，支持多阶段构建：

```dockerfile
# 构建阶段
FROM golang:1.26.3 AS builder
WORKDIR /workspace

# 复制依赖文件
COPY go.mod go.sum go.work go.work.sum ./
RUN go mod download

# 复制源码
COPY app ./app
COPY idl ./idl
COPY pkg ./pkg

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./app/ucenter/api

# 运行阶段
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY --from=builder /workspace/app/ucenter/api/etc /app/etc

ENTRYPOINT ["/app/server"]
CMD ["-f", "/app/etc/ucenter-api.yaml"]
```

### 10.2 Docker Compose 示例

```yaml
version: '3.8'
services:
  ucenter-api:
    build:
      context: .
      dockerfile: app/ucenter/api/Dockerfile
    ports:
      - "8888:8888"
    environment:
      - ETCD_HOSTS=etcd:2379
    depends_on:
      - etcd
      - ucenter-rpc
      - market-rpc
    networks:
      - mscoin-network

  etcd:
    image: bitnami/etcd:latest
    ports:
      - "2379:2379"
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
    networks:
      - mscoin-network

networks:
  mscoin-network:
    driver: bridge
```

### 10.3 启动命令

```bash
# 本地开发启动
go run main.go -f etc/ucenter-api.yaml

# Docker 启动
docker run -p 8888:8888 ucenter-api:latest -f /app/etc/ucenter-api.yaml

# 指定配置文件启动
./server -f /path/to/config.yaml
```

### 10.4 环境变量配置

生产环境建议通过环境变量覆盖敏感配置：

```bash
# JWT 密钥
export JWT_ACCESS_SECRET="your-production-secret"

# Etcd 地址
export UCENTER_RPC_ETCD_HOSTS="etcd1:2379,etcd2:2379"
export MARKET_RPC_ETCD_HOSTS="etcd1:2379,etcd2:2379"
```

---

## 附录：常见问题

### Q1: Token 过期后如何处理？
前端收到 4000 错误码时，应清除本地存储的 Token，跳转到登录页面让用户重新登录。

### Q2: 为什么 CheckLogin 接口不使用 Auth 中间件？
该接口的目的是返回 Token 是否有效的状态，如果使用中间件，Token 无效时会被拦截返回错误，无法返回友好的 true/false 状态。

### Q3: 提现为什么需要双重验证？
- 资金密码：防止账号被盗后资产被盗
- 短信验证码：防止密码泄露后资产被盗

### Q4: 如何添加新的 API 接口？
1. 在 `types.go` 中定义请求和响应结构
2. 在 `logic/` 目录创建业务逻辑文件
3. 在 `handler/` 目录创建 HTTP 处理器
4. 在 `routes.go` 中注册路由

### Q5: 如何进行单元测试？
参考 `login_logic_test.go`，使用 fake 客户端模拟 RPC 响应：

```go
type fakeLoginClient struct {
    loginFn func(ctx context.Context, in *loginpb.LoginReq, opts ...grpc.CallOption) (*loginpb.LoginRes, error)
}

func TestLoginLogic(t *testing.T) {
    logic := NewLoginLogic(context.Background(), &svc.ServiceContext{
        LoginClient: &fakeLoginClient{
            loginFn: func(...) (*loginpb.LoginRes, error) {
                return &loginpb.LoginRes{...}, nil
            },
        },
    })
    // 测试逻辑...
}
```
