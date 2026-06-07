# MSCoin Go 公共包文档

## 概述

`pkg` 目录是 MSCoin Go 项目的公共基础设施包集合，提供各微服务共享的通用功能。这些包遵循以下设计原则：

1. **接口抽象**：每个包都定义接口，便于测试和扩展
2. **配置驱动**：所有组件通过配置结构体初始化，避免硬编码
3. **错误处理**：统一的错误处理策略，提供清晰的错误信息
4. **上下文支持**：所有 I/O 操作支持 `context.Context`，便于超时和取消控制

---

## 目录

- [pkg/auth - JWT 认证](#pkgauth---jwt-认证)
- [pkg/btcx - Bitcoin RPC 封装](#pkgbtcx---bitcoin-rpc-封装)
- [pkg/cache/redisx - Redis 封装](#pkgcacheredisx---redis-封装)
- [pkg/db/mysqlx - MySQL 封装](#pkgdbmysqlx---mysql-封装)
- [pkg/httputil - HTTP 工具](#pkghttputil---http-工具)
- [pkg/httpxutil - HTTP 客户端](#pkghttpxutil---http-客户端)
- [pkg/mq/kafka - Kafka 生产者和消费者](#pkgmqkafka---kafka-生产者和消费者)
- [pkg/okxx - OKX 交易所 API](#pkgokxx---okx-交易所-api)
- [pkg/page - 分页工具](#pkgpage---分页工具)
- [pkg/passwordx - 密码加密](#pkgpasswordx---密码加密)
- [pkg/result - 统一响应格式](#pkgresult---统一响应格式)
- [pkg/store/mongox - MongoDB 封装](#pkgstoremongox---mongodb-封装)

---

## pkg/auth - JWT 认证

### 功能说明

`auth` 包提供 JWT（JSON Web Token）相关的认证功能，封装了 `github.com/golang-jwt/jwt/v4` 库。采用 HS256 对称加密算法，与传统 MSCoin API 服务保持兼容。

### 主要函数

#### GenerateUserToken

```go
func GenerateUserToken(secret string, issuedAt time.Time, expireSeconds int64, userID int64) (string, error)
```

创建 MSCoin API 使用的 JWT Token。

**参数说明：**
- `secret`: JWT 签名密钥，必须保密，建议从环境变量或配置中心获取
- `issuedAt`: Token 签发时间，通常使用 `time.Now()`
- `expireSeconds`: Token 有效期（秒），建议设置为合理的过期时间（如 3600 秒）
- `userID`: 用户唯一标识，将存储在 Token 的 `userId` claim 中

**返回值：**
- `string`: 生成的 JWT 字符串，格式为 `header.payload.signature`
- `error`: 签名失败时返回错误

#### ParseUserID

```go
func ParseUserID(tokenString string, secret string) (int64, error)
```

从 JWT Token 中解析用户 ID。

**参数说明：**
- `tokenString`: 客户端提供的 JWT 字符串
- `secret`: 验证签名的密钥，必须与生成时使用的密钥一致

**返回值：**
- `int64`: 用户 ID
- `error`: Token 无效、已过期或缺少必要 claim 时返回错误

### 使用示例

```go
package main

import (
    "time"
    "fmt"
    "mscoin_go/pkg/auth"
)

func main() {
    secret := "your-jwt-secret-key"
    
    // 生成 Token（用户登录成功后）
    token, err := auth.GenerateUserToken(secret, time.Now(), 3600, 12345)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Generated Token: %s\n", token)
    
    // 解析 Token（API 中间件验证）
    userID, err := auth.ParseUserID(token, secret)
    if err != nil {
        // Token 无效或已过期，返回 401
        fmt.Printf("Token invalid: %v\n", err)
        return
    }
    fmt.Printf("User ID: %d\n", userID)
}
```

### 设计思路

1. **保留传统 Claim 名称**：使用 `userId` 而非标准的 `sub`，保持与传统 API 的兼容性
2. **安全检查**：验证签名算法，防止算法混淆攻击
3. **显式过期检查**：虽然 `jwt.Parse` 已验证 `exp`，但显式检查确保行为一致
4. **类型安全**：JSON 数字解析为 `float64`，需要转换为 `int64`

---

## pkg/btcx - Bitcoin RPC 封装

### 功能说明

`btcx` 包提供 Bitcoin Core 节点的 JSON-RPC 客户端功能，以及本地 BTC 钱包创建功能。

### 主要结构体和接口

#### NodeConfig - 节点配置

```go
type NodeConfig struct {
    URL              string  // Bitcoin Core 节点的 RPC 地址
    Username         string  // RPC 认证用户名
    Password         string  // RPC 认证密码
    MinConfirmations int     // UTXO 最小确认数
    MaxConfirmations int     // UTXO 最大确认数
    TimeoutMs        int     // RPC 请求超时时间（毫秒）
    AddressType      string  // 地址类型：legacy、bech32
}
```

#### WithdrawSender - 提现发送接口

```go
type WithdrawSender interface {
    Send(ctx context.Context, fromAddress string, toAddress string, 
         totalAmount float64, arrivedAmount float64) (string, error)
}
```

提现发送器接口，用于执行 BTC 提现交易。

#### AddressAllocator - 地址分配接口

```go
type AddressAllocator interface {
    Allocate(ctx context.Context, label string) (string, error)
}
```

地址分配器接口，用于在节点钱包中创建新地址。

#### Wallet - 本地钱包

```go
type Wallet struct {
    // 包含私钥和公钥
}

func NewWallet() (*Wallet, error)                    // 创建新钱包
func (w *Wallet) TestnetAddress() string            // 获取测试网地址
func (w *Wallet) EncodedPrivateKey() (string, error) // 获取编码后的私钥
```

### 使用示例

#### 提现发送

```go
package main

import (
    "context"
    "fmt"
    "mscoin_go/pkg/btcx"
)

func main() {
    // 创建提现发送器
    sender, err := btcx.NewWithdrawSender(btcx.NodeConfig{
        URL:      "http://127.0.0.1:8332",
        Username: "bitcoin",
        Password: "password",
        TimeoutMs: 30000,
    })
    if err != nil {
        panic(err)
    }
    
    // 执行提现
    ctx := context.Background()
    txid, err := sender.Send(ctx, 
        "sourceAddress",    // 平台热钱包地址
        "targetAddress",    // 用户提现地址
        0.001,              // 总金额（包含矿工费）
        0.0009,             // 用户到账金额
    )
    if err != nil {
        panic(err)
    }
    fmt.Printf("Transaction ID: %s\n", txid)
}
```

#### 地址分配

```go
// 创建地址分配器
allocator, err := btcx.NewAddressAllocator(btcx.NodeConfig{
    URL:         "http://127.0.0.1:8332",
    Username:    "bitcoin",
    Password:    "password",
    AddressType: "legacy",
})

// 为用户生成新地址
address, err := allocator.Allocate(ctx, "member-123-btc")
fmt.Printf("New Address: %s\n", address)
```

#### 本地钱包创建

```go
// 创建本地钱包
wallet, err := btcx.NewWallet()
if err != nil {
    panic(err)
}

// 获取测试网地址
address := wallet.TestnetAddress()
fmt.Printf("Testnet Address: %s\n", address)

// 获取编码后的私钥（用于存储）
privateKey, err := wallet.EncodedPrivateKey()
fmt.Printf("Private Key: %s\n", privateKey)
```

### 设计思路

1. **接口抽象**：`WithdrawSender` 和 `AddressAllocator` 接口解耦业务与传输层
2. **贪心 UTXO 选择**：简单按顺序选择 UTXO，满足当前业务场景
3. **传统兼容**：保留 Base58 编码的测试网地址格式
4. **安全设计**：
   - 私钥使用 ECDSA P256 曲线
   - 私钥存储为 PEM + Base58 编码
   - 地址生成遵循比特币规范

### 交易流程

```
Send() 方法执行流程：
1. 验证参数（地址、金额）
2. 查询源地址的 UTXO
3. 选择足够的 UTXO 作为输入
4. 构造交易输出（目标地址 + 找零地址）
5. 创建原始交易（createrawtransaction）
6. 使用钱包签名（signrawtransactionwithwallet）
7. 广播到网络（sendrawtransaction）
```

---

## pkg/cache/redisx - Redis 封装

### 功能说明

`redisx` 包封装了 `github.com/go-redis/redis/v8` 库，提供统一的 Redis 客户端创建和常用缓存操作。

### 主要结构体和方法

#### Config - 配置

```go
type Config struct {
    Addrs    []string  // Redis 服务器地址列表
    Password string    // Redis 认证密码
    DB       int       // 数据库编号（0-15）
}
```

#### Client - 客户端

```go
type Client struct {
    // 内部包装 go-redis 客户端
}

func New(cfg Config) *Client                                      // 创建客户端
func (c *Client) Raw() goredis.UniversalClient                    // 获取原始客户端
func (c *Client) Get(key string, value any) error                 // 获取值
func (c *Client) GetCtx(ctx context.Context, key string, value any) error
func (c *Client) Set(key string, value any) error                 // 设置值
func (c *Client) SetCtx(ctx context.Context, key string, value any) error
func (c *Client) SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error
func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error
```

### 使用示例

#### 基本操作

```go
package main

import (
    "context"
    "time"
    "mscoin_go/pkg/cache/redisx"
)

func main() {
    // 创建 Redis 客户端
    client := redisx.New(redisx.Config{
        Addrs:    []string{"127.0.0.1:6379"},
        Password: "",
        DB:       0,
    })
    
    ctx := context.Background()
    
    // 存储字符串
    err := client.SetCtx(ctx, "user:123:token", "abc123")
    
    // 存储带过期时间的数据
    err = client.SetWithExpireCtx(ctx, "verify:13800138000", "123456", 5*time.Minute)
    
    // 读取字符串
    var token string
    err = client.GetCtx(ctx, "user:123:token", &token)
    
    // 存储结构体（JSON 格式）
    user := map[string]any{
        "id":   123,
        "name": "test",
    }
    err = client.SetJSON(ctx, "user:123", user, 1*time.Hour)
    
    // 使用原始客户端进行高级操作
    raw := client.Raw()
    result, err := raw.Incr(ctx, "counter").Result()
}
```

### 设计思路

1. **自动类型编码**：
   - `string`: 直接存储
   - `[]byte`: 转换为字符串
   - `fmt.Stringer`: 调用 `String()` 方法
   - 基本数值类型: 使用 `fmt.Sprint` 格式化
   - 其他类型: JSON 编码

2. **自动类型解码**：
   - `*string`: 直接转换
   - `*[]byte`: 复制字节
   - 其他类型: JSON 解码

3. **UniversalClient**：自动适应单机或集群模式

---

## pkg/db/mysqlx - MySQL 封装

### 功能说明

`mysqlx` 包封装了 `github.com/jmoiron/sqlx` 库，提供统一的 MySQL 数据库连接管理和事务编排。

### 主要结构体和接口

#### Config - 配置

```go
type Config struct {
    DataSource      string  // DSN 连接字符串
    MaxOpenConns    int     // 最大打开连接数
    MaxIdleConns    int     // 最大空闲连接数
    ConnMaxLifetime int64   // 连接最大存活时间（秒）
    ConnMaxIdleTime int64   // 空闲连接最大存活时间（秒）
}
```

#### ExtContext - 执行上下文

```go
type ExtContext = sqlx.ExtContext
```

统一 `*sqlx.DB` 和 `*sqlx.Tx` 的执行上下文，使仓库方法可以同时支持事务和非事务场景。

#### TxManager - 事务管理接口

```go
type TxManager interface {
    WithinTx(ctx context.Context, fn func(exec ExtContext) error) error
}
```

### 使用示例

#### 数据库连接

```go
package main

import (
    "mscoin_go/pkg/db/mysqlx"
)

func main() {
    db, err := mysqlx.New(mysqlx.Config{
        DataSource:      "root:password@tcp(127.0.0.1:3306)/mscoin?parseTime=true",
        MaxOpenConns:    100,
        MaxIdleConns:    25,
        ConnMaxLifetime: 1800,
    })
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // 执行查询
    var users []User
    err = db.SelectContext(ctx, &users, "SELECT * FROM users WHERE status = ?", 1)
}
```

#### 事务管理

```go
// 创建事务管理器
txManager := mysqlx.NewTxManager(db)

// 在事务中执行多个操作
err := txManager.WithinTx(ctx, func(exec mysqlx.ExtContext) error {
    // 创建用户
    if _, err := exec.ExecContext(ctx, 
        "INSERT INTO users (name, email) VALUES (?, ?)", 
        "test", "test@example.com"); err != nil {
        return err
    }
    
    // 创建钱包
    if _, err := exec.ExecContext(ctx,
        "INSERT INTO wallets (user_id, balance) VALUES (LAST_INSERT_ID(), 0)"); err != nil {
        return err
    }
    
    return nil
})
```

#### 仓库方法设计

```go
// 仓库方法接收 ExtContext，同时支持 DB 和 Tx
type UserRepository struct {
    db *sqlx.DB
}

func (r *UserRepository) Create(ctx context.Context, exec mysqlx.ExtContext, user *User) error {
    _, err := exec.ExecContext(ctx,
        "INSERT INTO users (name, email) VALUES (?, ?)",
        user.Name, user.Email,
    )
    return err
}

// 非事务调用
err := repo.Create(ctx, db, user)

// 事务调用
err := txManager.WithinTx(ctx, func(exec mysqlx.ExtContext) error {
    return repo.Create(ctx, exec, user)
})
```

### 设计思路

1. **驼峰转蛇形命名**：自动将 `UserName` 映射到 `user_name` 列
2. **GORM 标签兼容**：支持 `gorm:"column:custom_name"` 自定义列名
3. **快速失败**：启动时 Ping 数据库，尽早发现连接问题
4. **事务自动管理**：成功自动提交，失败自动回滚

---

## pkg/httputil - HTTP 工具

### 功能说明

`httputil` 包提供 HTTP 请求处理相关的实用函数，主要用于从 HTTP 请求中提取客户端真实 IP 地址。

### 主要函数

#### ClientIP

```go
func ClientIP(r *http.Request) string
```

从 HTTP 请求中解析客户端真实 IP 地址。

**IP 获取优先级：**
1. `X-Real-IP` 头：通常由 Nginx 等反向代理设置
2. `X-Forwarded-For` 头的第一个 IP：可能由多级代理设置
3. `RemoteAddr`：直接连接的客户端地址

### 使用示例

```go
package main

import (
    "net/http"
    "fmt"
    "mscoin_go/pkg/httputil"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    ip := httputil.ClientIP(r)
    fmt.Printf("Request from IP: %s\n", ip)
    
    // 用于限流或日志记录
    if isRateLimited(ip) {
        http.Error(w, "Too many requests", 429)
        return
    }
}

func main() {
    http.HandleFunc("/api", Handler)
    http.ListenAndServe(":8080", nil)
}
```

### 设计思路

1. **优先级设计**：`X-Real-IP` 最可靠，优先使用
2. **IPv6 兼容**：将 `::1` 转换为 `127.0.0.1`，统一格式
3. **安全警告**：`X-Forwarded-For` 可被客户端伪造，生产环境应配合可信代理列表使用

---

## pkg/httpxutil - HTTP 客户端

### 功能说明

`httpxutil` 包提供简化的 HTTP 客户端工具函数，主要用于发送 JSON 格式的 POST 请求。

### 主要函数

#### PostJSON

```go
func PostJSON(url string, payload any) ([]byte, error)
```

发送 POST 请求，请求体为 JSON 格式。

**特点：**
- 自动将 payload 编码为 JSON
- 设置 `Content-Type: application/json` 头
- 使用 20 秒超时
- 返回完整响应体

### 使用示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "mscoin_go/pkg/httpxutil"
)

func main() {
    // 调用第三方 API
    resp, err := httpxutil.PostJSON("https://api.example.com/users", map[string]any{
        "name":  "test",
        "email": "test@example.com",
    })
    if err != nil {
        panic(err)
    }
    
    // 解析响应
    var result struct {
        ID    int64  `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    json.Unmarshal(resp, &result)
    fmt.Printf("Created user: %d\n", result.ID)
}
```

### 设计思路

1. **简洁 API**：一行代码完成 JSON POST 请求
2. **默认超时**：20 秒超时防止请求阻塞
3. **注意事项**：不区分成功和失败的 HTTP 状态码，都会返回响应体

---

## pkg/mq/kafka - Kafka 生产者和消费者

### 功能说明

`kafka` 包提供 Kafka 消息队列的生产者和消费者实现，基于 `github.com/segmentio/kafka-go` 库。

### 生产者

#### Config - 生产者配置

```go
type Config struct {
    Brokers                []string  // Kafka 集群地址
    Topic                  string    // 目标主题
    Sync                   bool      // 是否同步发送
    AllowAutoTopicCreation bool      // 是否允许自动创建主题
}
```

#### Producer 接口

```go
type Producer interface {
    PushWithKey(ctx context.Context, key string, value string) error
    Close() error
}
```

#### 使用示例

```go
package main

import (
    "context"
    "encoding/json"
    "mscoin_go/pkg/mq/kafka"
)

func main() {
    // 创建生产者
    producer, err := kafka.NewProducer(kafka.Config{
        Brokers: []string{"localhost:9092"},
        Topic:   "withdraw",
        Sync:    true, // 同步发送，确保消息可靠
    })
    if err != nil {
        panic(err)
    }
    defer producer.Close()
    
    // 发送消息
    message := map[string]any{
        "userId": 123,
        "amount": 100.0,
    }
    data, _ := json.Marshal(message)
    
    ctx := context.Background()
    err = producer.PushWithKey(ctx, "user:123", string(data))
}
```

### 消费者

#### ConsumerConfig - 消费者配置

```go
type ConsumerConfig struct {
    Brokers              []string  // Kafka 集群地址
    Topic                string    // 订阅主题
    GroupID              string    // 消费者组 ID
    MinBytes             int       // 每次拉取最小字节数
    MaxBytes             int       // 每次拉取最大字节数
    MaxWaitMs            int       // 等待消息最长时间
    RetryBackoffMs       int       // 重试退避时间
    StartOffset          int64     // 起始偏移量（-1 最新，-2 最早）
    DeadLetterTopic      string    // 死信队列主题
    AllowAutoTopicCreate bool      // 是否允许自动创建主题
}
```

#### ConsumeAction - 消费动作

```go
const (
    ConsumeAck       ConsumeAction = iota  // 提交 offset
    ConsumeRetry                           // 重试消息
    ConsumeDeadLetter                      // 发送到死信队列
)
```

#### 使用示例

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "mscoin_go/pkg/mq/kafka"
)

func main() {
    // 定义消息处理器
    handler := func(ctx context.Context, msg kafka.Message) error {
        var withdraw struct {
            UserID int64   `json:"userId"`
            Amount float64 `json:"amount"`
        }
        if err := json.Unmarshal(msg.Value, &withdraw); err != nil {
            return err // 格式错误，将进入死信队列
        }
        
        // 处理提现逻辑
        log.Printf("Processing withdraw for user %d: %.2f", withdraw.UserID, withdraw.Amount)
        return nil
    }
    
    // 定义错误分类器
    classifier := func(err error) kafka.ConsumeAction {
        if err == nil {
            return kafka.ConsumeAck
        }
        // 根据错误类型决定动作
        if isTemporaryError(err) {
            return kafka.ConsumeRetry // 临时错误，重试
        }
        return kafka.ConsumeDeadLetter // 永久错误，进入死信队列
    }
    
    // 创建消费者服务
    service, err := kafka.NewConsumerService(
        kafka.ConsumerConfig{
            Brokers:         []string{"localhost:9092"},
            Topic:           "withdraw",
            GroupID:         "withdraw-processor",
            DeadLetterTopic: "withdraw.dlq",
        },
        handler,
        classifier,
    )
    if err != nil {
        panic(err)
    }
    
    // 启动消费者（阻塞运行）
    service.Start()
}
```

### 设计思路

1. **接口抽象**：`Producer` 接口便于测试和替换实现
2. **消息键支持**：相同键的消息发送到同一分区，保证顺序
3. **错误分类**：区分临时错误和永久错误，采用不同处理策略
4. **死信队列**：避免毒消息阻塞正常处理
5. **go-zero 集成**：消费者服务实现 `service.Service` 接口

---

## pkg/okxx - OKX 交易所 API

### 功能说明

`okxx` 包提供与 OKX 交易所 API 的交互功能，主要用于获取汇率和 K 线数据。

### 主要结构体和接口

#### Config - 配置

```go
type Config struct {
    APIKey     string  // API 密钥
    SecretKey  string  // API 密钥
    Passphrase string  // API 口令
    Host       string  // API 主机地址
    Proxy      string  // 代理地址
    TimeoutMs  int     // 请求超时时间
}
```

#### ExchangeRate - 汇率

```go
type ExchangeRate struct {
    USDCNY float64  // 美元兑人民币汇率
}
```

#### Candle - K 线数据

```go
type Candle struct {
    Time         int64   // 时间戳（毫秒）
    OpenPrice    float64 // 开盘价
    HighestPrice float64 // 最高价
    LowestPrice  float64 // 最低价
    ClosePrice   float64 // 收盘价
    Count        float64 // 交易笔数
    Volume       float64 // 成交量
    Turnover     float64 // 成交额
}
```

#### Client 接口

```go
type Client interface {
    FetchExchangeRate(ctx context.Context) (*ExchangeRate, error)
    FetchCandles(ctx context.Context, instID string, bar string) ([]*Candle, error)
}
```

### 使用示例

```go
package main

import (
    "context"
    "fmt"
    "mscoin_go/pkg/okxx"
)

func main() {
    // 创建 OKX 客户端
    client, err := okxx.NewClient(okxx.Config{
        Host:      "https://www.okx.com",
        TimeoutMs: 30000,
    })
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // 获取 USD/CNY 汇率
    rate, err := client.FetchExchangeRate(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("USD/CNY: %.4f\n", rate.USDCNY)
    
    // 获取 BTC-USDT 的 1 小时 K 线
    candles, err := client.FetchCandles(ctx, "BTC-USDT", "1H")
    if err != nil {
        panic(err)
    }
    for _, c := range candles {
        fmt.Printf("Time: %d, Close: %.2f\n", c.Time, c.ClosePrice)
    }
}
```

### 设计思路

1. **凭据可选**：市场数据端点公开，无需认证即可使用
2. **HMAC 签名**：如果提供了凭据，自动添加签名头
3. **代理支持**：支持配置代理，用于网络受限环境
4. **错误处理**：统一的 OKX API 状态码检查

---

## pkg/page - 分页工具

### 功能说明

`page` 包提供统一的分页响应格式，用于列表类 API 的返回，与传统 MSCoin 前端契约保持兼容。

### 主要结构体

#### Result - 分页结果

```go
type Result struct {
    Content       []any `json:"content"`       // 当前页数据列表
    TotalElements int64 `json:"totalElements"` // 总记录数
    Number        int64 `json:"number"`        // 当前页码（从 0 开始）
    TotalPages    int64 `json:"totalPages"`    // 总页数
    HasNext       bool  `json:"hasNext"`       // 是否有下一页
    IsLast        bool  `json:"isLast"`        // 是否最后一页
}
```

### 使用示例

```go
package main

import (
    "mscoin_go/pkg/page"
)

func GetUserList(pageNum int64, pageSize int64) *page.Result {
    // 查询数据
    users := db.QueryUsers(pageNum, pageSize)
    total := db.CountUsers()
    
    // 构造分页结果
    // 转换为 []any
    content := make([]any, len(users))
    for i, u := range users {
        content[i] = u
    }
    
    return page.New(content, pageNum, pageSize, total)
}

// 使用示例
result := GetUserList(1, 10) // 第 2 页，每页 10 条
// JSON 输出：
// {
//     "content": [...],
//     "totalElements": 100,
//     "number": 1,
//     "totalPages": 10,
//     "hasNext": true,
//     "isLast": false
// }
```

### 设计思路

1. **前端兼容**：字段名称与传统 MSCoin 前端契约一致
2. **自动计算**：自动计算总页数、是否有下一页等字段
3. **从 0 开始**：页码从 0 开始，前端显示时需 +1

---

## pkg/passwordx - 密码加密

### 功能说明

`passwordx` 包使用 PBKDF2 算法进行密码哈希和验证，是 OWASP 推荐的密码存储方案之一。

### 算法参数

- **迭代次数**：10000 次
- **密钥长度**：128 字节
- **盐长度**：64 字节
- **哈希函数**：SHA512

### 主要函数

#### Encode

```go
func Encode(rawPwd string) (string, string)
```

加密密码，返回盐和哈希值。

#### Verify

```go
func Verify(rawPwd string, salt string, encodedPwd string) bool
```

验证密码是否匹配。

### 使用示例

```go
package main

import (
    "mscoin_go/pkg/passwordx"
)

func main() {
    // 用户注册：加密密码
    salt, hash := passwordx.Encode("user_password")
    
    // 存储到数据库
    db.Exec("INSERT INTO users (username, salt, password_hash) VALUES (?, ?, ?)",
        "test", salt, hash)
    
    // 用户登录：验证密码
    storedSalt, storedHash := db.GetUserCredentials("test")
    
    if passwordx.Verify("user_password", storedSalt, storedHash) {
        // 密码正确，允许登录
        println("Login successful")
    } else {
        // 密码错误，拒绝登录
        println("Invalid password")
    }
}
```

### 设计思路

1. **随机盐**：每次加密生成随机盐，防止彩虹表攻击
2. **高迭代次数**：10000 次迭代增加计算成本，抵御暴力破解
3. **字母数字盐**：盐使用字母数字字符，便于存储和传输
4. **相同密码不同结果**：每次加密结果不同，增强安全性

---

## pkg/result - 统一响应格式

### 功能说明

`result` 包定义 API 服务的统一 HTTP 响应格式，所有 HTTP API 都使用此格式返回数据。

### 响应格式

```json
{
    "code": 0,
    "message": "success",
    "data": {...}
}
```

- `code`: 状态码，0 表示成功，500 表示失败
- `message`: 状态描述信息
- `data`: 业务数据

### 主要结构体和方法

```go
type Result struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"`
}

func New() *Result                           // 创建空结果
func (r *Result) Success(data any)          // 标记成功
func (r *Result) Fail(code int, message string) // 标记失败
func (r *Result) Deal(data any, err error) *Result // 处理 (data, err) 模式
```

### 使用示例

```go
package main

import (
    "encoding/json"
    "net/http"
    "mscoin_go/pkg/result"
)

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    userId := parseUserId(r)
    
    user, err := userService.GetUser(userId)
    
    // 方式一：使用 Deal 方法
    resp := result.New().Deal(user, err)
    json.NewEncoder(w).Encode(resp)
    
    // 方式二：手动构建
    resp := result.New()
    if err != nil {
        resp.Fail(500, "用户不存在")
    } else {
        resp.Success(user)
    }
    json.NewEncoder(w).Encode(resp)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := parseUser(r)
    if err != nil {
        json.NewEncoder(w).Encode(result.New().Fail(400, "参数错误"))
        return
    }
    
    err = userService.CreateUser(user)
    resp := result.New().Deal(user, err)
    json.NewEncoder(w).Encode(resp)
}
```

### 设计思路

1. **统一契约**：所有 API 返回相同格式，前端统一处理
2. **简化处理器**：`Deal` 方法简化常见的 `(data, err)` 处理模式
3. **可扩展状态码**：除 0/500 外，可定义其他状态码（如 401 未授权）

---

## pkg/store/mongox - MongoDB 封装

### 功能说明

`mongox` 包封装了 MongoDB Go Driver，提供统一的客户端创建和生命周期管理。

### 主要结构体

#### Config - 配置

```go
type Config struct {
    URI      string  // MongoDB 连接 URI
    Username string  // 认证用户名（可选）
    Password string  // 认证密码（可选）
    Database string  // 默认数据库名称
}
```

#### Client - 客户端

```go
type Client struct {
    // 内部包装 MongoDB 客户端
}

func New(cfg Config) (*Client, error)           // 创建客户端
func (c *Client) Database() *mongo.Database     // 获取默认数据库
func (c *Client) Disconnect(ctx context.Context) error // 关闭连接
```

### 使用示例

```go
package main

import (
    "context"
    "mscoin_go/pkg/store/mongox"
)

func main() {
    // 创建 MongoDB 客户端
    client, err := mongox.New(mongox.Config{
        URI:      "mongodb://localhost:27017",
        Username: "admin",
        Password: "password",
        Database: "mscoin",
    })
    if err != nil {
        panic(err)
    }
    defer client.Disconnect(context.Background())
    
    ctx := context.Background()
    
    // 获取集合
    collection := client.Database().Collection("users")
    
    // 插入文档
    user := map[string]any{
        "name":  "test",
        "email": "test@example.com",
    }
    _, err = collection.InsertOne(ctx, user)
    
    // 查询文档
    var result map[string]any
    err = collection.FindOne(ctx, map[string]any{"name": "test"}).Decode(&result)
}
```

### 设计思路

1. **快速失败**：启动时 Ping 数据库，尽早发现连接问题
2. **默认数据库**：通过 `Database()` 方法快速访问默认数据库
3. **优雅关闭**：提供 `Disconnect` 方法释放资源
4. **认证支持**：支持 MongoDB 用户名密码认证

---

## 最佳实践

### 1. 依赖注入

推荐通过依赖注入使用这些包，而不是在函数内部创建实例：

```go
// 推荐
type UserService struct {
    db    *sqlx.DB
    cache *redisx.Client
}

func NewUserService(db *sqlx.DB, cache *redisx.Client) *UserService {
    return &UserService{db: db, cache: cache}
}

// 不推荐
func GetUser(id int64) (*User, error) {
    db, _ := mysqlx.New(...) // 每次调用都创建新连接
    // ...
}
```

### 2. 配置管理

将配置集中管理，从环境变量或配置文件读取：

```go
type Config struct {
    MySQL mysqlx.Config
    Redis redisx.Config
    Kafka kafka.Config
}

func LoadConfig() *Config {
    return &Config{
        MySQL: mysqlx.Config{
            DataSource: os.Getenv("MYSQL_DSN"),
        },
        // ...
    }
}
```

### 3. 上下文传递

始终使用 `context.Context` 进行超时控制：

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := userService.GetUser(ctx, userId)
```

### 4. 错误处理

不要忽略错误，提供有意义的错误信息：

```go
user, err := userService.GetUser(ctx, userId)
if err != nil {
    return fmt.Errorf("get user %d: %w", userId, err)
}
```

---

## 版本兼容性

- Go 1.21+
- github.com/golang-jwt/jwt/v4
- github.com/go-redis/redis/v8
- github.com/jmoiron/sqlx
- github.com/segmentio/kafka-go
- go.mongodb.org/mongo-driver
- golang.org/x/crypto
