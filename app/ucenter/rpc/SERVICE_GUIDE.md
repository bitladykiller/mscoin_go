# UCenter RPC 服务详解

本文档详细描述了 MSCoin 平台的 UCenter RPC 服务，包括服务概述、目录结构、分层架构、核心业务逻辑、gRPC 服务定义、数据库设计、与其他服务的调用关系、Kafka 事件发布及配置说明。

---

## 目录

1. [服务概述](#1-服务概述)
2. [目录结构](#2-目录结构)
3. [分层架构详解](#3-分层架构详解)
4. [每个文件的详细说明](#4-每个文件的详细说明)
5. [gRPC 服务定义](#5-grpc-服务定义)
6. [数据库设计](#6-数据库设计)
7. [与其他服务的调用关系](#7-与其他服务的调用关系)
8. [Kafka 事件发布](#8-kafka-事件发布)
9. [配置说明](#9-配置说明)

---

## 1. 服务概述

### 1.1 功能定位

UCenter RPC（用户中心服务）是 MSCoin 平台的核心微服务之一，负责处理用户相关的所有业务逻辑。该服务采用 go-zero 框架构建，通过 gRPC 协议对外提供服务。

**主要功能模块：**

| 模块 | 功能描述 |
|------|----------|
| **会员注册** | 手机号注册、验证码发送与验证、密码加密存储 |
| **会员登录** | 人机验证、密码验证、JWT Token 生成 |
| **会员信息管理** | 会员信息查询、等级管理、合伙人状态管理 |
| **钱包资产管理** | 钱包查询、充值地址分配、余额管理 |
| **交易记录查询** | 充值/提现/转账/兑换记录查询，支持多条件筛选 |
| **提现申请** | 验证码验证、交易密码验证、余额冻结、提现记录创建 |

### 1.2 在整体架构中的位置

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Gateway                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       UCenter API (HTTP)                         │
│                     用户中心 HTTP 接口层                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      UCenter RPC (gRPC)                          │
│                   ★ 当前文档描述的服务 ★                          │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐     │
│  │  Register   │    Login    │   Member    │    Asset    │     │
│  │   Server    │   Server    │   Server    │   Server    │     │
│  ├─────────────┴─────────────┴─────────────┴─────────────┤     │
│  │                    Withdraw Server                      │     │
│  └───────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   MySQL     │     │   Redis     │     │   Kafka     │
│  会员/钱包   │     │  验证码缓存  │     │  提现事件   │
└─────────────┘     └─────────────┘     └─────────────┘
         │
         ▼
┌─────────────┐     ┌─────────────┐
│ Market RPC  │     │ Bitcoin Core│
│  币种信息    │     │  地址分配    │
└─────────────┘     └─────────────┘
```

### 1.3 技术栈

- **框架**: go-zero (微服务框架)
- **通信协议**: gRPC (基于 HTTP/2 的高性能 RPC 框架)
- **序列化**: Protocol Buffers 3
- **数据库**: MySQL (通过 sqlx 操作)
- **缓存**: Redis (验证码、会话等临时数据)
- **消息队列**: Kafka (提现事件发布)
- **服务注册与发现**: Etcd
- **认证**: JWT (JSON Web Token)

---

## 2. 目录结构

```
/Volumes/移动卷宗/学习/go/mscoin_go/app/ucenter/rpc/
├── main.go                          # 服务入口文件
├── Dockerfile                       # Docker 构建文件
├── etc/                             # 配置文件目录
│   └── ucenter.yaml                 # 服务配置文件
├── internal/                        # 内部实现（不对外暴露）
│   ├── config/                      # 配置结构定义
│   │   └── config.go                # 配置结构体定义
│   ├── domain/                      # 领域层
│   │   └── service/                 # 领域服务
│   │       ├── captcha_service.go   # 验证码服务
│   │       ├── member_service.go    # 会员服务
│   │       ├── transaction_service.go # 交易服务
│   │       ├── wallet_service.go    # 钱包服务
│   │       └── withdraw_service.go  # 提现服务
│   ├── logic/                       # 业务逻辑处理器
│   │   ├── find_address_by_coin_id_logic.go  # 查询提现地址
│   │   ├── find_member_by_id_logic.go        # 查询会员信息
│   │   ├── find_transaction_logic.go         # 查询交易记录
│   │   ├── find_wallet_by_symbol_logic.go    # 按币种查询钱包
│   │   ├── find_wallet_logic.go              # 查询所有钱包
│   │   ├── get_address_logic.go              # 获取充值地址列表
│   │   ├── login_logic.go                    # 登录逻辑
│   │   ├── register_by_phone_logic.go        # 手机号注册
│   │   ├── reset_address_logic.go            # 重置充值地址
│   │   ├── send_code_logic.go                # 发送注册验证码
│   │   ├── send_withdraw_code_logic.go       # 发送提现验证码
│   │   ├── withdraw_code_logic.go            # 提现申请
│   │   └── withdraw_record_logic.go          # 提现记录查询
│   ├── model/                       # 数据模型
│   │   ├── member.go                # 会员模型
│   │   ├── member_address.go        # 会员提现地址模型
│   │   ├── transaction.go           # 交易记录模型
│   │   ├── wallet.go                # 钱包模型
│   │   └── withdraw_record.go       # 提现记录模型
│   ├── repository/                  # 数据仓储层
│   │   ├── member_address_repository.go  # 会员地址仓储
│   │   ├── member_repository.go          # 会员仓储
│   │   ├── transaction_repository.go      # 交易仓储
│   │   ├── wallet_repository.go           # 钱包仓储
│   │   └── withdraw_repository.go         # 提现记录仓储
│   ├── server/                      # gRPC 服务端
│   │   ├── asset_server.go          # 资产服务端
│   │   ├── login_server.go          # 登录服务端
│   │   ├── member_server.go         # 会员服务端
│   │   ├── register_server.go       # 注册服务端
│   │   └── withdraw_server.go       # 提现服务端
│   └── svc/                         # 服务上下文
│       └── service_context.go       # 服务上下文定义
└── pb/                              # 生成的 protobuf 代码
    ├── asset/                       # 资产服务 protobuf
    │   ├── asset.pb.go
    │   └── asset_grpc.pb.go
    ├── login/                       # 登录服务 protobuf
    │   ├── login.pb.go
    │   └── login_grpc.pb.go
    ├── member/                      # 会员服务 protobuf
    │   ├── member.pb.go
    │   └── member_grpc.pb.go
    ├── register/                    # 注册服务 protobuf
    │   ├── register.pb.go
    │   └── register_grpc.pb.go
    └── withdraw/                    # 提现服务 protobuf
        ├── withdraw.pb.go
        └── withdraw_grpc.pb.go
```

---

## 3. 分层架构详解

UCenter RPC 服务采用经典的分层架构，每一层有明确的职责边界。

### 3.1 Server 层 (gRPC 服务端)

**职责**：
- 实现 gRPC 服务接口
- 接收 gRPC 请求并转发给 Logic 层
- 返回 gRPC 响应

**设计原则**：
- 薄层设计：Server 层只做请求转发，不包含业务逻辑
- 单一职责：每个 Server 结构体只实现一个 RPC 服务
- 依赖注入：通过 ServiceContext 获取 Logic 处理器

**代码示例**：
```go
// Server 层只做转发，不包含业务逻辑
func (s *LoginServer) Login(ctx context.Context, in *loginpb.LoginReq) (*loginpb.LoginRes, error) {
    return logic.NewLoginLogic(ctx, s.svcCtx).Login(in)
}
```

### 3.2 Logic 层 (业务逻辑处理器)

**职责**：
- 接收 Server 层转发的请求
- 调用领域服务处理业务逻辑
- 处理跨服务调用（如调用 Market RPC）
- 返回处理结果

**设计原则**：
- 薄层设计：Logic 层主要做请求转发和结果组装
- 单一职责：每个 Logic 结构体只处理一个 RPC 方法
- 依赖注入：通过 ServiceContext 获取领域服务

**代码示例**：
```go
// Logic 层处理跨服务调用和请求组装
func (l *FindWalletLogic) FindWallet(req *assetpb.AssetReq) (*assetpb.MemberWalletList, error) {
    list, err := l.svcCtx.WalletService.FindWallet(l.ctx, req.UserId, func(ctx context.Context, unit string) (*marketpb.Coin, error) {
        // 调用 Market RPC 获取币种信息
        return l.svcCtx.MarketClient.FindCoinInfo(ctx, &marketpb.MarketReq{Unit: unit})
    })
    if err != nil {
        return nil, err
    }
    return &assetpb.MemberWalletList{List: list}, nil
}
```

### 3.3 Domain/Service 层 (领域服务)

**职责**：
- 实现核心业务逻辑
- 协调多个仓储完成复杂业务流程
- 处理事务管理
- 发布领域事件（Kafka 消息）

**设计原则**：
- 单一职责：每个服务只处理一个领域的业务逻辑
- 依赖注入：通过构造函数注入仓储和其他依赖
- 接口隔离：定义所需的最小仓储接口

**核心服务**：
| 服务 | 职责 |
|------|------|
| MemberService | 会员登录、注册、信息查询 |
| WalletService | 钱包查询、地址管理 |
| TransactionService | 交易记录查询 |
| WithdrawService | 提现申请、验证码发送 |
| CaptchaService | 人机验证码验证 |

### 3.4 Repository 层 (数据仓储)

**职责**：
- 封装数据库操作
- 提供 CRUD 接口
- 处理 SQL 查询和结果映射
- 支持事务操作

**设计原则**：
- 单一职责：每个仓储只负责一张主表
- 接口隔离：仓储只暴露必要的数据访问方法
- 错误处理：区分"未找到"和"查询失败"

**核心仓储**：
| 仓储 | 职责 |
|------|------|
| MemberRepository | 会员数据的 CRUD |
| WalletRepository | 钱包数据的 CRUD，支持 FOR UPDATE 锁 |
| TransactionRepository | 交易记录查询 |
| WithdrawRepository | 提现记录的 CRUD |
| MemberAddressRepository | 会员地址查询 |

### 3.5 Model 层 (数据模型)

**职责**：
- 定义数据结构
- 映射数据库表结构
- 提供数据转换方法（如 ToProto）
- 封装简单的业务规则

**设计原则**：
- 贫血模型：模型主要承载数据，复杂业务逻辑在 Service 层
- 结构体标签：使用 `db` 和 `gorm` 标签映射数据库字段
- 业务常量：定义状态码、类型码等业务常量

---

## 4. 每个文件的详细说明

### 4.1 入口文件

#### main.go

**文件职责**：服务的入口点，负责初始化和启动 gRPC 服务。

**关键流程**：
1. 解析命令行参数（配置文件路径）
2. 加载配置文件
3. 创建服务上下文（初始化数据库、缓存、消息队列等）
4. 创建 gRPC 服务端并注册各业务服务
5. 启动服务监听

**关键代码解析**：
```go
// 创建 gRPC 服务端并注册各 RPC 服务
s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
    registerpb.RegisterRegisterServer(grpcServer, server.NewRegisterServer(ctx)) // 注册服务
    loginpb.RegisterLoginServer(grpcServer, server.NewLoginServer(ctx))          // 登录服务
    memberpb.RegisterMemberServer(grpcServer, server.NewMemberServer(ctx))       // 会员服务
    assetpb.RegisterAssetServer(grpcServer, server.NewAssetServer(ctx))          // 资产服务
    withdrawpb.RegisterWithdrawServer(grpcServer, server.NewWithdrawServer(ctx)) // 提现服务
})
```

### 4.2 配置文件

#### internal/config/config.go

**文件职责**：定义服务配置结构。

**关键结构体**：
```go
// Config ucenter RPC 服务配置
type Config struct {
    zrpc.RpcServerConf                      // go-zero RPC 服务端配置（嵌入）
    Mysql     mysqlx.Config                 // MySQL 数据库配置
    Redis     redisx.Config                 // Redis 缓存配置
    Kafka     kafka.Config                  // Kafka 消息队列配置
    JWT       AuthConfig                    // JWT 认证配置
    Captcha   CaptchaConfig                 // 验证码服务配置
    MarketRPC zrpc.RpcClientConf            // Market RPC 客户端配置
    Bitcoin   btcx.NodeConfig               // Bitcoin 节点配置
}

// AuthConfig JWT 认证配置
type AuthConfig struct {
    AccessSecret string // JWT 签名密钥
    AccessExpire int64  // JWT 过期时间（秒）
}

// CaptchaConfig 验证码服务配置
type CaptchaConfig struct {
    Vid string // 验证码服务 ID
    Key string // 验证码服务密钥
}
```

### 4.3 服务上下文

#### internal/svc/service_context.go

**文件职责**：聚合所有业务依赖，实现依赖注入。

**关键结构体**：
```go
type ServiceContext struct {
    Config config.Config         // 服务配置
    DB     *sqlx.DB              // 数据库连接池
    Cache  *redisx.Client        // Redis 缓存客户端
    Queue  kafka.Producer        // Kafka 消息生产者

    MemberService      *service.MemberService      // 会员服务
    WalletService      *service.WalletService      // 钱包服务
    TransactionService *service.TransactionService // 交易服务
    WithdrawService    *service.WithdrawService    // 提现服务
    MarketClient       marketpb.MarketClient       // Market RPC 客户端
    AddressAllocator   btcx.AddressAllocator       // BTC 地址分配器
}
```

**初始化顺序**：
1. 基础设施层：数据库、Redis、Kafka、Bitcoin 节点
2. 仓储层：各数据表的 CRUD 操作封装
3. 服务层：业务逻辑处理

### 4.4 Model 层文件

#### internal/model/member.go

**文件职责**：定义会员模型和会员相关的业务常量。

**关键结构体**：
```go
// Member 会员模型
// 存储会员的基本信息、认证状态、钱包关联等数据
type Member struct {
    Id                         int64   `db:"id" gorm:"column:id"`                     // 会员 ID
    MobilePhone                string  `db:"mobile_phone" gorm:"column:mobile_phone"` // 手机号
    Username                   string  `db:"username" gorm:"column:username"`         // 用户名
    Password                   string  `db:"password" gorm:"column:password"`         // 密码（加密后）
    Salt                       string  `db:"salt" gorm:"column:salt"`                 // 密码盐值
    MemberLevel                int64   `db:"member_level" gorm:"column:member_level"` // 会员等级
    SuperPartner               string  `db:"super_partner" gorm:"column:super_partner"` // 合伙人状态
    Status                     int64   `db:"status" gorm:"column:status"`             // 会员状态
    PromotionCode              string  `db:"promotion_code" gorm:"column:promotion_code"` // 推广码
    // ... 其他字段
}
```

**关键方法**：
```go
// MemberLevelText 返回会员等级文本描述
func (m *Member) MemberLevelText() string {
    switch m.MemberLevel {
    case 0: return "普通会员"
    case 1: return "实名"
    case 2: return "认证商家"
    default: return ""
    }
}

// MemberRate 返回会员费率等级
func (m *Member) MemberRate() int32 {
    switch m.SuperPartner {
    case "1": return 1  // 超级合伙人
    case "2": return 2  // P 级超级合伙人
    default: return 0   // 普通会员
    }
}

// NewMemberForRegister 构建注册所需的会员对象
func NewMemberForRegister(now time.Time, phone string, username string, country string, encodedPassword string, salt string, partner string, promotion string) *Member
```

#### internal/model/wallet.go

**文件职责**：定义会员钱包模型。

**关键结构体**：
```go
// MemberWallet 会员钱包模型
type MemberWallet struct {
    Id                int64   `db:"id"`                    // 钱包 ID
    Address           string  `db:"address"`               // 钱包地址（充值地址）
    Balance           float64 `db:"balance"`               // 可用余额
    FrozenBalance     float64 `db:"frozen_balance"`        // 冻结余额
    ReleaseBalance    float64 `db:"release_balance"`       // 释放余额
    IsLock            int32   `db:"is_lock"`               // 是否锁定
    MemberId          int64   `db:"member_id"`             // 会员 ID
    Version           int32   `db:"version"`               // 版本号（乐观锁）
    CoinId            int64   `db:"coin_id"`               // 币种 ID
    CoinName          string  `db:"coin_name"`             // 币种名称
    ToReleased        float64 `db:"to_released"`           // 待释放金额
    AddressPrivateKey string  `db:"address_private_key"`   // 地址私钥（已废弃）
}
```

**余额字段说明**：
| 字段 | 说明 |
|------|------|
| Balance | 可用余额，会员可用于提现或转账 |
| FrozenBalance | 冻结余额，提现申请时冻结 |
| ReleaseBalance | 释放余额，锁仓释放场景 |
| ToReleased | 待释放金额，锁仓金额逐步释放 |

#### internal/model/transaction.go

**文件职责**：定义会员交易记录模型。

**关键结构体**：
```go
// MemberTransaction 会员交易记录模型
type MemberTransaction struct {
    Id          int64   `db:"id"`           // 交易 ID
    Address     string  `db:"address"`      // 交易地址
    Amount      float64 `db:"amount"`       // 交易金额
    CreateTime  int64   `db:"create_time"`  // 创建时间（毫秒时间戳）
    Fee         float64 `db:"fee"`          // 手续费
    Flag        int32   `db:"flag"`         // 标记
    MemberId    int64   `db:"member_id"`    // 会员 ID
    Symbol      string  `db:"symbol"`       // 币种符号
    Type        int32   `db:"type"`         // 交易类型
    DiscountFee string  `db:"discount_fee"` // 折扣手续费
    RealFee     string  `db:"real_fee"`     // 实际手续费
}
```

**交易类型常量**：
| 常量 | 值 | 说明 |
|------|-----|------|
| transactionRecharge | 0 | 充值：外部转入 |
| transactionWithdraw | 1 | 提现：转出到外部 |
| transactionTransferAccounts | 2 | 转账：会员间内部转账 |
| transactionExchange | 3 | 兑换：币种兑换 |

#### internal/model/withdraw_record.go

**文件职责**：定义提现记录模型。

**关键结构体**：
```go
// WithdrawRecord 提现记录模型
type WithdrawRecord struct {
    Id                int64   `db:"id"`                  // 记录 ID
    MemberId          int64   `db:"member_id"`           // 会员 ID
    CoinId            int64   `db:"coin_id"`             // 币种 ID
    TotalAmount       float64 `db:"total_amount"`        // 提现总额
    Fee               float64 `db:"fee"`                 // 手续费
    ArrivedAmount     float64 `db:"arrived_amount"`      // 到账金额
    Address           string  `db:"address"`             // 提现地址
    TransactionNumber string  `db:"transaction_number"`  // 交易号（链上哈希）
    Status            int32   `db:"status"`              // 提现状态
    CreateTime        int64   `db:"create_time"`         // 创建时间
    DealTime          int64   `db:"deal_time"`           // 处理时间
}
```

**提现状态常量**：
| 常量 | 值 | 说明 |
|------|-----|------|
| WithdrawStatusProcessing | 0 | 处理中：等待链上处理 |
| WithdrawStatusWaiting | 1 | 等待中：保留状态 |
| WithdrawStatusFail | 2 | 失败：提现失败 |
| WithdrawStatusSuccess | 3 | 成功：提现完成 |

#### internal/model/member_address.go

**文件职责**：定义会员提现地址模型。

**关键结构体**：
```go
// MemberAddress 会员提现地址模型
type MemberAddress struct {
    Id         int64  `db:"id"`          // 地址 ID
    MemberId   int64  `db:"member_id"`   // 会员 ID
    CoinId     int64  `db:"coin_id"`     // 币种 ID
    Address    string `db:"address"`     // 提现地址
    Remark     string `db:"remark"`      // 备注说明
    Status     int32  `db:"status"`      // 地址状态
    CreateTime int64  `db:"create_time"` // 创建时间
    DeleteTime int64  `db:"delete_time"` // 删除时间（软删除）
}
```

**与 Wallet 的区别**：
- `MemberWallet`：平台控制的链上钱包，用于充值
- `MemberAddress`：用户配置的提现目标地址，用于提现

### 4.5 Repository 层文件

#### internal/repository/member_repository.go

**文件职责**：封装会员表的数据库操作。

**关键方法**：
```go
// FindByPhone 根据手机号查询会员
func (r *MemberRepository) FindByPhone(ctx context.Context, phone string) (*model.Member, error)

// FindByID 根据会员 ID 查询会员
func (r *MemberRepository) FindByID(ctx context.Context, memberID int64) (*model.Member, error)

// UpdateLoginCount 更新会员登录次数
func (r *MemberRepository) UpdateLoginCount(ctx context.Context, id int64, step int) error

// Save 保存会员记录
func (r *MemberRepository) Save(ctx context.Context, member *model.Member) error
```

**设计要点**：
- 使用 sqlx 进行数据库操作，支持结构体映射
- 区分"未找到"和"查询失败"
- 参数化查询防止 SQL 注入

#### internal/repository/wallet_repository.go

**文件职责**：封装钱包表的数据库操作，支持事务和行锁。

**关键方法**：
```go
// FindByMemberID 根据会员 ID 查询所有钱包
func (r *WalletRepository) FindByMemberID(ctx context.Context, memberID int64) ([]*model.MemberWallet, error)

// FindByMemberIDAndCoinName 根据会员 ID 和币种名称查询钱包
func (r *WalletRepository) FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error)

// FindByMemberIDAndCoinNameForUpdate 使用 FOR UPDATE 锁定钱包行
// 防止并发提现请求同时冻结相同的余额快照
func (r *WalletRepository) FindByMemberIDAndCoinNameForUpdate(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error)

// FreezeBalance 冻结余额（原子操作）
func (r *WalletRepository) FreezeBalance(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error

// UpdateAddress 更新钱包地址
func (r *WalletRepository) UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error
```

**事务安全设计**：
- `FindByMemberIDAndCoinNameForUpdate` 使用 `FOR UPDATE` 锁定行
- `FreezeBalance` 在 SQL 中原子执行余额转移
- 防止并发提现导致余额超扣

#### internal/repository/transaction_repository.go

**文件职责**：封装交易记录表的查询操作。

**关键方法**：
```go
// FindTransaction 查询会员交易记录
// 支持按交易类型、时间范围、币种筛选，支持分页
func (r *TransactionRepository) FindTransaction(
    ctx context.Context,
    memberID int64,
    pageNo int64,
    pageSize int64,
    symbol string,
    startTime string,
    endTime string,
    transactionType string,
) ([]*model.MemberTransaction, int64, error)
```

**时间格式支持**：
- "2006-01-02 15:04:05" (完整时间)
- "2006-01-02 15:04" (省略秒)
- "2006-01-02" (仅日期)
- RFC3339 (标准格式)

#### internal/repository/withdraw_repository.go

**文件职责**：封装提现记录表的数据库操作。

**关键方法**：
```go
// FindByMemberID 根据会员 ID 分页查询提现记录
func (r *WithdrawRepository) FindByMemberID(ctx context.Context, memberID int64, page int64, pageSize int64) ([]*model.WithdrawRecord, int64, error)

// Save 保存提现记录（支持事务）
func (r *WithdrawRepository) Save(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error
```

#### internal/repository/member_address_repository.go

**文件职责**：封装会员地址表的查询操作。

**关键方法**：
```go
// FindByMemberIDAndCoinID 根据会员 ID 和币种 ID 查询提现地址列表
func (r *MemberAddressRepository) FindByMemberIDAndCoinID(ctx context.Context, memberID int64, coinID int64) ([]*model.MemberAddress, error)
```

### 4.6 Domain/Service 层文件

#### internal/domain/service/member_service.go

**文件职责**：实现会员相关的核心业务逻辑。

**关键方法**：

**1. Login - 会员登录**

```go
func (s *MemberService) Login(ctx context.Context, req *loginpb.LoginReq) (*loginpb.LoginRes, error)
```

登录流程：
1. 验证人机验证码（Captcha）
2. 查询会员信息
3. 验证密码（加盐哈希比对）
4. 生成 JWT Token
5. 异步更新登录次数

**2. RegisterByPhone - 手机号注册**

```go
func (s *MemberService) RegisterByPhone(ctx context.Context, req *registerpb.RegReq) (*registerpb.RegRes, error)
```

注册流程：
1. 验证人机验证码
2. 验证短信验证码（Redis 缓存比对）
3. 检查手机号是否已注册
4. 密码加密（加盐哈希）
5. 创建会员记录

**3. SendRegisterCode - 发送注册验证码**

```go
func (s *MemberService) SendRegisterCode(ctx context.Context, req *registerpb.CodeReq) (*registerpb.NoRes, error)
```

验证码规则：
- 长度：4 位数字
- 有效期：15 分钟
- 缓存键：`REGISTER::{phone}`

**4. FindByID - 查询会员信息**

```go
func (s *MemberService) FindByID(ctx context.Context, memberID int64) (*memberpb.MemberInfo, error)
```

#### internal/domain/service/wallet_service.go

**文件职责**：实现钱包相关的业务逻辑。

**关键方法**：

**1. FindWallet - 查询会员所有钱包**

```go
func (s *WalletService) FindWallet(ctx context.Context, memberID int64, findCoin func(context.Context, string) (*marketpb.Coin, error)) ([]*assetpb.MemberWallet, error)
```

**2. FindWalletBySymbol - 根据币种查询钱包**

```go
func (s *WalletService) FindWalletBySymbol(ctx context.Context, memberID int64, coinName string, coin *marketpb.Coin) (*assetpb.MemberWallet, error)
```

**3. EnsureWalletBySymbol - 确保会员拥有指定币种钱包**

```go
func (s *WalletService) EnsureWalletBySymbol(ctx context.Context, memberID int64, coinName string, coin *marketpb.Coin) (*model.MemberWallet, error)
```

设计原则：
- 延迟创建：不预先创建所有币种钱包
- 幂等性：多次调用不会创建重复钱包

**4. UpdateAddress - 更新钱包地址**

```go
func (s *WalletService) UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error
```

#### internal/domain/service/withdraw_service.go

**文件职责**：实现提现相关的核心业务逻辑。

**关键方法**：

**1. Apply - 提现申请（核心方法）**

```go
func (s *WithdrawService) Apply(ctx context.Context, req *withdrawpb.WithdrawReq) error
```

提现流程：
1. 验证请求参数（金额、地址、验证码等）
2. 查询会员信息，获取手机号
3. 验证 Redis 中的验证码
4. 验证交易密码
5. 在事务中执行：
   - 使用 FOR UPDATE 锁定钱包行
   - 检查余额是否充足
   - 冻结余额（Balance -> FrozenBalance）
   - 创建提现记录
6. 发布 Kafka 事件通知下游处理

**事务安全设计**：
- 使用 FOR UPDATE 行锁防止并发提现超扣
- 余额冻结在 SQL 中原子执行
- 提现记录创建和余额冻结在同一事务中

**2. SendCode - 发送提现验证码**

```go
func (s *WithdrawService) SendCode(ctx context.Context, phone string) error
```

验证码规则：
- 长度：6 位数字
- 有效期：5 分钟
- 缓存键：`WITHDRAW::{phone}`

**3. FindRecordList - 查询提现记录**

```go
func (s *WithdrawService) FindRecordList(ctx context.Context, memberID int64, page int64, pageSize int64, findCoin func(context.Context, int64) (*marketpb.Coin, error)) ([]*withdrawpb.WithdrawRecord, int64, error)
```

#### internal/domain/service/captcha_service.go

**文件职责**：实现人机验证码验证。

**关键方法**：

```go
// Verify 验证验证码
func (s *CaptchaService) Verify(server string, vid string, key string, token string, scene int, ip string) bool
```

验证逻辑：
- 验证码服务器地址为空时返回 true（用于开发环境）
- 调用第三方验证码服务验证
- 验证失败返回 false，不暴露具体原因

#### internal/domain/service/transaction_service.go

**文件职责**：实现交易记录查询。

**关键方法**：

```go
// FindTransaction 查询会员交易记录
func (s *TransactionService) FindTransaction(
    ctx context.Context,
    memberID int64,
    pageNo int64,
    pageSize int64,
    symbol string,
    startTime string,
    endTime string,
    transactionType string,
) ([]*assetpb.MemberTransaction, int64, error)
```

### 4.7 Logic 层文件

Logic 层是薄层设计，主要做请求转发和结果组装。以下是各 Logic 文件的职责：

| 文件 | 职责 | 调用的服务 |
|------|------|------------|
| login_logic.go | 处理登录请求 | MemberService.Login |
| register_by_phone_logic.go | 处理注册请求 | MemberService.RegisterByPhone |
| send_code_logic.go | 发送注册验证码 | MemberService.SendRegisterCode |
| find_member_by_id_logic.go | 查询会员信息 | MemberService.FindByID |
| find_wallet_logic.go | 查询所有钱包 | WalletService.FindWallet |
| find_wallet_by_symbol_logic.go | 按币种查询钱包 | WalletService.FindWalletBySymbol |
| find_transaction_logic.go | 查询交易记录 | TransactionService.FindTransaction |
| get_address_logic.go | 获取充值地址列表 | WalletService.GetAllAddress |
| reset_address_logic.go | 重置充值地址 | WalletService.EnsureWalletBySymbol, UpdateAddress |
| find_address_by_coin_id_logic.go | 查询提现地址 | WithdrawService.FindAddressByCoinID |
| send_withdraw_code_logic.go | 发送提现验证码 | WithdrawService.SendCode |
| withdraw_code_logic.go | 处理提现申请 | WithdrawService.Apply |
| withdraw_record_logic.go | 查询提现记录 | WithdrawService.FindRecordList |

### 4.8 Server 层文件

Server 层实现 gRPC 服务接口，以下是各 Server 文件的职责：

| 文件 | 服务名 | 方法 |
|------|--------|------|
| register_server.go | RegisterServer | RegisterByPhone, SendCode |
| login_server.go | LoginServer | Login |
| member_server.go | MemberServer | FindMemberById |
| asset_server.go | AssetServer | FindWalletBySymbol, FindWallet, ResetAddress, FindTransaction, GetAddress |
| withdraw_server.go | WithdrawServer | FindAddressByCoinId, SendCode, WithdrawCode, WithdrawRecord |

---

## 5. gRPC 服务定义

UCenter RPC 服务定义了 5 个独立的 gRPC 服务，每个服务对应一个 proto 文件。

### 5.1 Register 服务（注册服务）

**Proto 文件位置**：`idl/rpc/ucenter/register.proto`

```protobuf
service Register {
  rpc registerByPhone(RegReq) returns(RegRes);  // 手机号注册
  rpc sendCode(CodeReq) returns(NoRes);          // 发送验证码
}

// 注册请求
message RegReq {
  string username = 1;       // 用户名
  string password = 2;       // 密码
  CaptchaReq captcha = 3;    // 人机验证码
  string phone = 4;          // 手机号
  string promotion = 5;      // 推广码
  string code = 6;           // 短信验证码
  string country = 7;        // 国家
  string superPartner = 8;   // 合伙人状态
  string ip = 9;             // 客户端 IP
}

// 验证码请求
message CodeReq {
  string phone = 1;          // 手机号
  string country = 2;        // 国家
}
```

### 5.2 Login 服务（登录服务）

**Proto 文件位置**：`idl/rpc/ucenter/login.proto`

```protobuf
service Login {
  rpc login(LoginReq) returns(LoginRes);  // 会员登录
}

// 登录请求
message LoginReq {
  string username = 1;       // 用户名（手机号）
  string password = 2;       // 密码
  CaptchaReq captcha = 3;    // 人机验证码
  string ip = 4;             // 客户端 IP
}

// 登录响应
message LoginRes {
  string username = 1;       // 用户名
  string token = 2;          // JWT Token
  string memberLevel = 3;    // 会员等级
  string realName = 4;       // 真实姓名
  string country = 5;        // 国家
  string avatar = 6;         // 头像 URL
  string promotionCode = 7;  // 推广码
  int64 id = 8;              // 会员 ID
  int32 loginCount = 9;      // 登录次数
  string superPartner = 10;  // 合伙人状态
  int32 memberRate = 11;     // 会员费率等级
}
```

### 5.3 Member 服务（会员服务）

**Proto 文件位置**：`idl/rpc/ucenter/member.proto`

```protobuf
service Member {
  rpc FindMemberById(MemberReq) returns(MemberInfo);  // 查询会员信息
}

// 会员请求
message MemberReq {
  int64 memberId = 3;        // 会员 ID
}

// 会员信息（完整字段）
message MemberInfo {
  int64 id = 1;
  string mobilePhone = 33;
  string username = 50;
  // ... 其他 60+ 字段
}
```

### 5.4 Asset 服务（资产服务）

**Proto 文件位置**：`idl/rpc/ucenter/asset.proto`

```protobuf
service Asset {
  rpc findWalletBySymbol(AssetReq) returns(MemberWallet);      // 按币种查询钱包
  rpc findWallet(AssetReq) returns(MemberWalletList);          // 查询所有钱包
  rpc ResetAddress(AssetReq) returns(AssetResp);               // 重置充值地址
  rpc FindTransaction(AssetReq) returns(MemberTransactionList); // 查询交易记录
  rpc getAddress(AssetReq) returns(AddressList);               // 获取充值地址列表
}

// 资产请求
message AssetReq {
  string coinName = 1;       // 币种名称
  string ip = 2;             // 客户端 IP
  int64 userId = 3;          // 用户 ID
  string startTime = 4;      // 开始时间
  string endTime = 5;        // 结束时间
  int64 pageNo = 6;          // 页码
  int64 pageSize = 7;        // 每页条数
  string type = 8;           // 交易类型
  string symbol = 9;         // 币种符号
}

// 会员钱包
message MemberWallet {
  int64 id = 1;
  string address = 2;
  double balance = 3;
  double frozenBalance = 4;
  double releaseBalance = 5;
  int32 isLock = 6;
  int64 memberId = 7;
  int32 version = 8;
  Coin coin = 9;
  double toReleased = 10;
}

// 交易记录
message MemberTransaction {
  int64 id = 1;
  string address = 2;
  double amount = 3;
  string createTime = 4;
  double fee = 5;
  int32 flag = 6;
  int64 memberId = 7;
  string symbol = 8;
  string type = 9;
}
```

### 5.5 Withdraw 服务（提现服务）

**Proto 文件位置**：`idl/rpc/ucenter/withdraw.proto`

```protobuf
service Withdraw {
  rpc findAddressByCoinId(WithdrawReq) returns(AddressSimpleList);  // 查询提现地址
  rpc SendCode(WithdrawReq) returns(NoRes);                          // 发送提现验证码
  rpc WithdrawCode(WithdrawReq) returns(NoRes);                      // 提现申请
  rpc WithdrawRecord(WithdrawReq) returns(RecordList);               // 查询提现记录
}

// 提现请求
message WithdrawReq {
  int64 coinId = 1;          // 币种 ID
  string ip = 2;             // 客户端 IP
  int64 userId = 3;          // 用户 ID
  string phone = 4;          // 手机号
  string unit = 5;           // 币种单位
  string address = 6;        // 提现地址
  double amount = 7;         // 提现金额
  double fee = 8;            // 手续费
  string jyPassword = 9;     // 交易密码
  string code = 10;          // 验证码
  int64 page = 11;           // 页码
  int64 pageSize = 12;       // 每页条数
}

// 提现记录
message WithdrawRecord {
  int64 id = 1;
  int64 memberId = 2;
  Coin coin = 3;
  double totalAmount = 4;
  double fee = 5;
  double arrivedAmount = 6;
  string address = 7;
  string remark = 8;
  string transactionNumber = 9;
  int32 status = 12;
  string createTime = 13;
  string dealTime = 14;
}
```

---

## 6. 数据库设计

UCenter 服务使用 MySQL 数据库，主要涉及以下数据表。

### 6.1 member 表（会员表）

**表职责**：存储会员的基本信息、认证状态、钱包关联等数据。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 会员 ID，自增主键 |
| mobile_phone | VARCHAR(20) | 手机号，登录账号 |
| username | VARCHAR(50) | 用户名，显示名称 |
| password | VARCHAR(100) | 密码（加盐哈希后） |
| salt | VARCHAR(32) | 密码盐值 |
| member_level | INT | 会员等级：0-普通，1-实名，2-认证商家 |
| super_partner | VARCHAR(1) | 合伙人状态："0"-普通，"1"-超级合伙人，"2"-P级 |
| status | INT | 会员状态：0-正常，1-违规 |
| promotion_code | VARCHAR(20) | 推广码 |
| registration_time | BIGINT | 注册时间（毫秒时间戳） |
| login_count | INT | 登录次数 |
| last_login_time | BIGINT | 最后登录时间（毫秒时间戳） |
| real_name | VARCHAR(50) | 真实姓名 |
| real_name_status | INT | 实名认证状态：0-未认证，1-已认证 |
| jy_password | VARCHAR(100) | 交易密码（加密后） |
| google_key | VARCHAR(32) | Google 验证器密钥 |
| google_state | INT | Google 验证器状态：0-未绑定，1-已绑定 |
| avatar | VARCHAR(255) | 头像 URL |
| country | VARCHAR(50) | 国家 |
| email | VARCHAR(100) | 邮箱地址 |
| inviter_id | BIGINT | 邀请人 ID |
| kyc_status | INT | KYC 认证状态 |

**索引设计**：
- PRIMARY KEY: `id`
- UNIQUE KEY: `mobile_phone`
- KEY: `promotion_code`

### 6.2 member_wallet 表（钱包表）

**表职责**：存储会员的币种钱包信息，包括余额、冻结余额等。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 钱包 ID，自增主键 |
| member_id | BIGINT | 会员 ID，关联 member 表 |
| coin_id | BIGINT | 币种 ID |
| coin_name | VARCHAR(20) | 币种名称，如 BTC、ETH |
| address | VARCHAR(100) | 钱包地址，用于充值 |
| balance | DECIMAL(20,8) | 可用余额 |
| frozen_balance | DECIMAL(20,8) | 冻结余额 |
| release_balance | DECIMAL(20,8) | 释放余额 |
| to_released | DECIMAL(20,8) | 待释放金额 |
| is_lock | INT | 是否锁定：0-正常，1-锁定 |
| version | INT | 版本号，乐观锁 |
| address_private_key | VARCHAR(255) | 地址私钥（已废弃） |

**索引设计**：
- PRIMARY KEY: `id`
- UNIQUE KEY: `(member_id, coin_name)`
- KEY: `coin_name`

**余额安全说明**：
- `balance`：可用余额，会员可自由支配
- `frozen_balance`：冻结余额，提现申请时冻结
- 余额冻结使用 SQL 原子操作，避免竞态条件

### 6.3 member_transaction 表（交易记录表）

**表职责**：记录会员的所有资产变动，包括充值、提现、转账、兑换。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 交易 ID，自增主键 |
| member_id | BIGINT | 会员 ID |
| symbol | VARCHAR(20) | 币种符号 |
| type | INT | 交易类型：0-充值，1-提现，2-转账，3-兑换 |
| amount | DECIMAL(20,8) | 交易金额 |
| fee | DECIMAL(20,8) | 手续费 |
| address | VARCHAR(100) | 交易地址 |
| flag | INT | 标记 |
| create_time | BIGINT | 创建时间（毫秒时间戳） |
| discount_fee | VARCHAR(50) | 折扣手续费 |
| real_fee | VARCHAR(50) | 实际手续费 |

**索引设计**：
- PRIMARY KEY: `id`
- KEY: `(member_id, create_time)`
- KEY: `(member_id, type)`

### 6.4 withdraw_record 表（提现记录表）

**表职责**：记录会员的每一笔提现申请，跟踪提现全生命周期。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 记录 ID，自增主键 |
| member_id | BIGINT | 会员 ID |
| coin_id | BIGINT | 币种 ID |
| total_amount | DECIMAL(20,8) | 提现总额 |
| fee | DECIMAL(20,8) | 手续费 |
| arrived_amount | DECIMAL(20,8) | 到账金额 |
| address | VARCHAR(100) | 提现地址 |
| remark | VARCHAR(255) | 备注 |
| transaction_number | VARCHAR(100) | 交易号（链上哈希） |
| can_auto_withdraw | INT | 是否可自动提现 |
| isAuto | INT | 是否自动处理 |
| status | INT | 提现状态：0-处理中，1-等待，2-失败，3-成功 |
| create_time | BIGINT | 创建时间（毫秒时间戳） |
| deal_time | BIGINT | 处理时间（毫秒时间戳） |

**索引设计**：
- PRIMARY KEY: `id`
- KEY: `(member_id, create_time)`
- KEY: `status`

**状态机**：
```
Processing (0) ──> Waiting (1) ──> Success (3)
       │
       └──> Fail (2)
```

### 6.5 member_address 表（会员地址表）

**表职责**：存储会员配置的提现目标地址。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 地址 ID，自增主键 |
| member_id | BIGINT | 会员 ID |
| coin_id | BIGINT | 币种 ID |
| address | VARCHAR(100) | 提现地址 |
| remark | VARCHAR(100) | 备注说明 |
| status | INT | 地址状态：0-正常，1-已删除 |
| create_time | BIGINT | 创建时间（毫秒时间戳） |
| delete_time | BIGINT | 删除时间（毫秒时间戳），软删除 |

**索引设计**：
- PRIMARY KEY: `id`
- KEY: `(member_id, coin_id)`

---

## 7. 与其他服务的调用关系

### 7.1 与 Market RPC 的调用

UCenter RPC 调用 Market RPC 获取币种信息，用于丰富钱包和交易记录的展示数据。

**调用场景**：
| 场景 | 调用方法 | 用途 |
|------|----------|------|
| 查询钱包列表 | `FindCoinInfo` | 获取币种汇率、限制等信息 |
| 按币种查询钱包 | `FindCoinInfo` | 获取币种信息 |
| 重置充值地址 | `FindCoinInfo` | 获取币种配置 |
| 查询提现记录 | `FindCoinById` | 获取币种信息 |

**调用代码示例**：
```go
// 在 Logic 层调用 Market RPC
func (l *FindWalletLogic) FindWallet(req *assetpb.AssetReq) (*assetpb.MemberWalletList, error) {
    list, err := l.svcCtx.WalletService.FindWallet(l.ctx, req.UserId, func(ctx context.Context, unit string) (*marketpb.Coin, error) {
        // 调用 Market RPC 获取币种信息
        return l.svcCtx.MarketClient.FindCoinInfo(ctx, &marketpb.MarketReq{Unit: unit})
    })
    // ...
}
```

**服务发现**：
- 通过 Etcd 进行服务发现
- 配置中的 `MarketRPC.Etcd.Key: market.rpc` 用于定位服务

### 7.2 与 Bitcoin Core 的交互

UCenter RPC 连接 Bitcoin Core 节点，为会员分配 BTC 充值地址。

**交互场景**：
| 场景 | Bitcoin RPC 方法 | 用途 |
|------|------------------|------|
| 重置充值地址 | `getnewaddress` | 为会员分配新的 BTC 地址 |

**调用代码示例**：
```go
// 在 ResetAddressLogic 中调用 Bitcoin Core
func (l *ResetAddressLogic) ResetAddress(req *assetpb.AssetReq) (*assetpb.AssetResp, error) {
    // ...
    if req.CoinName == "BTC" && wallet.Address == "" {
        // 从 Bitcoin Core 分配新地址
        address, err := l.svcCtx.AddressAllocator.Allocate(l.ctx, fmt.Sprintf("member-%d-btc", req.UserId))
        if err != nil {
            return nil, err
        }
        wallet.Address = address
        // 更新钱包地址
        if err := l.svcCtx.WalletService.UpdateAddress(l.ctx, wallet); err != nil {
            return nil, err
        }
    }
    // ...
}
```

**Bitcoin 配置**：
```yaml
Bitcoin:
  URL: http://bitcoin:18332/wallet/mscoin  # Bitcoin Core JSON-RPC 地址
  Username: bitcoin                         # RPC 用户名
  Password: "123456"                        # RPC 密码
  AddressType: legacy                       # 地址类型
  TimeoutMs: 20000                          # 超时时间（毫秒）
```

**安全设计**：
- BTC 地址由 Bitcoin Core 管理，私钥不在 MySQL 中存储
- 使用钱包标签（如 `member-{id}-btc`）追踪地址归属

---

## 8. Kafka 事件发布

### 8.1 事件发布场景

UCenter RPC 在提现申请成功后发布 Kafka 事件，通知下游服务（jobcenter）执行实际的链上转账。

**事件发布时机**：
- 提现申请事务提交成功后
- 在 `WithdrawService.Apply` 方法中

### 8.2 Kafka 配置

```yaml
Kafka:
  Brokers:
    - kafka:9092           # Kafka Broker 地址
  Topic: withdraw          # 提现事件主题
  Sync: true               # 同步发送
  AllowAutoTopicCreation: true  # 允许自动创建主题
```

### 8.3 事件消息结构

**消息主题**：`withdraw`

**消息内容**：提现记录 JSON

```json
{
  "id": 123456789,
  "member_id": 10001,
  "coin_id": 1,
  "total_amount": 0.1,
  "fee": 0.0001,
  "arrived_amount": 0.0999,
  "address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
  "remark": "",
  "transaction_number": "",
  "can_auto_withdraw": 0,
  "isAuto": 0,
  "status": 0,
  "create_time": 1672531200000,
  "deal_time": 0
}
```

### 8.4 事件发布代码

```go
// 在 WithdrawService.Apply 中发布事件
func (s *WithdrawService) Apply(ctx context.Context, req *withdrawpb.WithdrawReq) error {
    return s.txManager.WithinTx(ctx, func(exec mysqlx.ExtContext) error {
        // ... 冻结余额、创建提现记录 ...

        // 发布 Kafka 事件
        message, err := json.Marshal(record)
        if err != nil {
            return fmt.Errorf("marshal withdraw event: %w", err)
        }
        // 使用用户 ID 作为分区键，保证同一用户的提现顺序
        if err := s.queue.PushWithKey(ctx, strconv.FormatInt(req.UserId, 10), string(message)); err != nil {
            return fmt.Errorf("publish withdraw event: %w", err)
        }
        return nil
    })
}
```

### 8.5 事件消费流程

```
┌─────────────┐     Kafka      ┌─────────────┐     Bitcoin RPC     ┌─────────────┐
│  UCenter    │ ────────────> │  JobCenter  │ ─────────────────> │ Bitcoin Core│
│    RPC      │   withdraw    │    Job      │    sendtoaddress   │    Node     │
└─────────────┘     Topic     └─────────────┘                    └─────────────┘
      │                               │
      │                               │ 更新提现记录状态
      │                               ▼
      │                        ┌─────────────┐
      └─────────────────────── │   MySQL     │
              查询状态          │withdraw_record│
                               └─────────────┘
```

**消费流程**：
1. JobCenter 消费 `withdraw` 主题的消息
2. 解析提现记录
3. 调用 Bitcoin RPC 执行链上转账
4. 更新提现记录状态为 `Success` 或 `Fail`
5. 如果失败，解冻会员余额

---

## 9. 配置说明

### 9.1 配置文件结构

**文件位置**：`etc/ucenter.yaml`

```yaml
# 服务名称
Name: ucenter.rpc

# 服务监听地址
ListenOn: 0.0.0.0:8081

# Etcd 服务注册
Etcd:
  Hosts:
    - etcd:2379
  Key: ucenter.rpc

# MySQL 数据库配置
Mysql:
  DataSource: root:root@tcp(mysql:3306)/ucenter?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
  MaxOpenConns: 100    # 最大连接数
  MaxIdleConns: 20     # 最大空闲连接数

# Redis 缓存配置
Redis:
  Addrs:
    - redis:6379
  Password: ""
  DB: 0

# Kafka 消息队列配置
Kafka:
  Brokers:
    - kafka:9092
  Topic: withdraw         # 提现事件主题
  Sync: true              # 同步发送
  AllowAutoTopicCreation: true

# JWT 认证配置
JWT:
  AccessSecret: "!@#$mscoin"  # JWT 签名密钥（生产环境应使用环境变量）
  AccessExpire: 604800        # JWT 过期时间（秒），7 天

# 验证码服务配置
Captcha:
  Vid: ""   # 验证码服务 ID（为空时跳过验证，用于开发环境）
  Key: ""   # 验证码服务密钥

# Market RPC 客户端配置
MarketRPC:
  Etcd:
    Hosts:
      - etcd:2379
    Key: market.rpc
  NonBlock: true   # 非阻塞模式，Market RPC 不可用时不影响启动

# Bitcoin 节点配置
Bitcoin:
  URL: http://bitcoin:18332/wallet/mscoin  # Bitcoin Core JSON-RPC 地址
  Username: bitcoin                         # RPC 用户名
  Password: "123456"                        # RPC 密码
  AddressType: legacy                       # 地址类型：legacy（P2PKH）
  TimeoutMs: 20000                          # 请求超时时间（毫秒）
```

### 9.2 配置项详解

#### 服务配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Name | 服务名称，用于日志和监控 | ucenter.rpc |
| ListenOn | 服务监听地址 | 0.0.0.0:8081 |
| Etcd.Hosts | Etcd 集群地址 | etcd:2379 |
| Etcd.Key | 服务注册键 | ucenter.rpc |

#### MySQL 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| DataSource | 数据库连接字符串 | - |
| MaxOpenConns | 最大连接数 | 100 |
| MaxIdleConns | 最大空闲连接数 | 20 |

**DataSource 格式**：
```
{username}:{password}@tcp({host}:{port})/{database}?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
```

#### Redis 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Addrs | Redis 地址列表 | redis:6379 |
| Password | Redis 密码 | "" |
| DB | Redis 数据库编号 | 0 |

#### Kafka 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Brokers | Kafka Broker 地址列表 | kafka:9092 |
| Topic | 提现事件主题 | withdraw |
| Sync | 是否同步发送 | true |
| AllowAutoTopicCreation | 是否允许自动创建主题 | true |

#### JWT 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| AccessSecret | JWT 签名密钥 | "!@#$mscoin" |
| AccessExpire | JWT 过期时间（秒） | 604800（7 天） |

#### Market RPC 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Etcd.Hosts | Etcd 集群地址 | etcd:2379 |
| Etcd.Key | 服务发现键 | market.rpc |
| NonBlock | 非阻塞模式 | true |

#### Bitcoin 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| URL | Bitcoin Core JSON-RPC 地址 | http://bitcoin:18332/wallet/mscoin |
| Username | RPC 用户名 | bitcoin |
| Password | RPC 密码 | 123456 |
| AddressType | 地址类型 | legacy |
| TimeoutMs | 请求超时（毫秒） | 20000 |

### 9.3 环境变量配置

生产环境建议使用环境变量覆盖敏感配置：

```bash
# MySQL
export MYSQL_DATASOURCE="user:pass@tcp(prod-mysql:3306)/ucenter?charset=utf8mb4&parseTime=true"

# Redis
export REDIS_ADDR="prod-redis:6379"
export REDIS_PASSWORD="secure-password"

# Kafka
export KAFKA_BROKERS="kafka1:9092,kafka2:9092,kafka3:9092"

# JWT
export JWT_SECRET="your-secure-secret-key"

# Bitcoin
export BITCOIN_PASSWORD="your-bitcoin-rpc-password"
```

---

## 附录

### A. 错误码说明

| 错误信息 | 说明 | 解决方案 |
|----------|------|----------|
| captcha verification failed | 人机验证码验证失败 | 重新获取验证码 |
| user not registered | 用户未注册 | 先进行注册 |
| wrong password | 密码错误 | 检查密码或重置 |
| phone already registered | 手机号已注册 | 使用其他手机号 |
| verification code unavailable | 验证码不可用 | 重新获取验证码 |
| verification code mismatch | 验证码不匹配 | 检查验证码 |
| member not found | 会员不存在 | 检查会员 ID |
| wallet not found | 钱包不存在 | 先创建钱包 |
| insufficient balance | 余额不足 | 检查余额 |
| wrong transaction password | 交易密码错误 | 检查交易密码 |

### B. API 调用示例

#### 登录示例

```bash
grpcurl -plaintext -d '{
  "username": "13800138000",
  "password": "password123",
  "captcha": {
    "server": "https://captcha.example.com/verify",
    "token": "captcha-token"
  },
  "ip": "127.0.0.1"
}' localhost:8081 login.Login/login
```

#### 注册示例

```bash
grpcurl -plaintext -d '{
  "username": "testuser",
  "password": "password123",
  "phone": "13800138000",
  "code": "1234",
  "country": "CN",
  "captcha": {
    "server": "https://captcha.example.com/verify",
    "token": "captcha-token"
  }
}' localhost:8081 register.Register/registerByPhone
```

#### 提现申请示例

```bash
grpcurl -plaintext -d '{
  "userId": 10001,
  "coinId": 1,
  "unit": "BTC",
  "address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
  "amount": 0.1,
  "fee": 0.0001,
  "jyPassword": "123456",
  "code": "123456"
}' localhost:8081 withdraw.Withdraw/WithdrawCode
```

### C. 常见问题

**Q1: 为什么验证码验证总是失败？**

A: 检查 `Captcha.Vid` 和 `Captcha.Key` 配置是否正确。开发环境可以留空跳过验证。

**Q2: 提现申请失败，提示"insufficient balance"？**

A: 检查会员的可用余额（`balance` 字段），注意不是冻结余额。余额必须大于等于提现金额。

**Q3: 如何调试 gRPC 服务？**

A: 使用 `grpcurl` 工具，确保服务开启了 gRPC 反射（开发模式或测试模式）。

**Q4: Bitcoin 地址分配失败？**

A: 检查 Bitcoin Core 节点是否正常运行，RPC 配置是否正确，钱包是否已创建。

---

*文档版本: 1.0*
*最后更新: 2026-06-08*
