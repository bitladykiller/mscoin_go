# MSCoin-Go 学习路线

## 目标

通过本项目学习以下技术栈：
- **go-zero**：Go 语言微服务框架
- **Protocol Buffers (Proto)**：接口定义语言
- **gRPC**：高性能 RPC 框架
- **Kafka**：消息队列
- **Docker Compose**：容器编排

---

## 项目架构概览

```
mscoin_go/
├── app/                    # 业务应用目录
│   ├── market/            # 市场服务（币种、K线、汇率）
│   │   ├── api/           # HTTP API 层
│   │   └── rpc/           # gRPC RPC 层
│   ├── exchange/          # 交易服务（订单）
│   │   ├── api/
│   │   └── rpc/
│   ├── ucenter/           # 用户中心服务（用户、钱包、提现）
│   │   ├── api/
│   │   └── rpc/
│   └── jobcenter/         # 异步任务中心（消费者、定时任务）
├── idl/                    # 接口定义（契约）
│   ├── api/               # HTTP API 定义（.api 文件）
│   └── rpc/               # RPC 定义（.proto 文件）
├── pkg/                    # 公共基础设施包
│   ├── auth/              # JWT 认证
│   ├── btcx/              # Bitcoin RPC 封装
│   ├── cache/redisx/      # Redis 封装
│   ├── db/mysqlx/         # MySQL 封装
│   ├── mq/kafka/          # Kafka 生产者/消费者封装
│   ├── okxx/              # OKX 交易所 API
│   └── store/mongox/      # MongoDB 封装
└── docker-compose.yml      # 容器编排
```

---

## 第一阶段：环境搭建与项目运行（预计 1-2 天）

### 1.1 前置知识准备

| 知识点 | 学习资源 |
|--------|---------|
| Go 基础语法 | [Go 官方教程](https://go.dev/tour/) |
| Go 模块管理 | 了解 `go.mod`、`go.work` |
| Docker 基础 | 了解镜像、容器、网络 |

### 1.2 启动项目

```bash
# 进入项目目录
cd /Volumes/移动卷宗/学习/go/mscoin_go

# 拉取所需镜像（仅首次）
docker pull mongo:7.0
docker pull bitcoin/bitcoin:27.0
docker pull obsidiandynamics/kafdrop:4.0.2

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f market-api
```

### 1.3 验证服务

```bash
# 测试市场 API
curl -X POST http://localhost:8889/market/exchange-rate/usd/CNY

# 测试用户中心 API
curl -X POST http://localhost:8888/uc/send/code -d '{"phone":"13800138000"}'
```

### 1.4 学习任务

- [ ] 理解 `docker-compose.yml` 中的服务依赖关系
- [ ] 查看 Kafdrop 管理界面：http://localhost:19000
- [ ] 理解各服务如何通过服务名通信（如 `mysql:3306`）

---

## 第二阶段：go-zero 基础概念（预计 3-5 天）

### 2.1 go-zero 核心组件

go-zero 提供两种服务类型：

| 类型 | 说明 | 项目示例 |
|------|------|---------|
| **API 服务** | HTTP REST 服务 | `app/*/api/` |
| **RPC 服务** | gRPC 服务 | `app/*/rpc/` |

### 2.2 学习文件：`app/market/api/main.go`

```go
// 入口文件解析
var configFile = flag.String("f", "etc/market-api.yaml", "配置文件路径")

func main() {
    flag.Parse()

    // 1. 加载配置
    var c config.Config
    conf.MustLoad(*configFile, &c)

    // 2. 创建 HTTP 服务器
    server := rest.MustNewServer(c.RestConf)

    // 3. 创建服务上下文（依赖注入）
    ctx := svc.NewServiceContext(c)

    // 4. 注册路由
    handler.RegisterHandlers(server, ctx)

    // 5. 启动服务
    server.Start()
}
```

### 2.3 go-zero 目录结构约定

每个服务遵循以下结构：

```
app/market/api/
├── main.go                    # 入口文件
├── etc/                       # 配置文件
│   └── market-api.yaml
└── internal/                   # 内部实现（不对外暴露）
    ├── config/                # 配置结构定义
    ├── handler/               # HTTP 处理器（路由绑定）
    ├── logic/                 # 业务逻辑
    ├── middleware/            # 中间件
    ├── svc/                   # 服务上下文（依赖容器）
    └── types/                 # 请求/响应类型
```

### 2.4 学习任务

- [ ] 阅读 `app/market/api/main.go`，理解启动流程
- [ ] 阅读 `app/market/api/internal/config/config.go`，理解配置结构
- [ ] 阅读 `app/market/api/internal/handler/routes.go`，理解路由注册
- [ ] 对比 `app/market/api` 和 `app/market/rpc` 的区别

---

## 第三阶段：API 服务详解（预计 3-5 天）

### 3.1 API 定义文件：`idl/api/market.api`

go-zero 使用 `.api` 文件定义 HTTP 接口：

```go
// 类型定义
type RateRequest {
    Unit string `path:"unit" json:"unit"`
    Ip   string `json:"ip,optional"`
}

type RateResponse {
    Rate float64 `json:"rate"`
}

// 路由定义
@server (
    prefix: /market
)
service market {
    @handler UsdRate
    post /exchange-rate/usd/:unit (RateRequest) returns (RateResult)
}
```

**关键语法：**
- `path:"unit"` → URL 路径参数 `/market/exchange-rate/usd/:unit`
- `form:"symbol,optional"` → 表单参数
- `json:"symbol,optional"` → JSON body 参数

### 3.2 请求处理流程

```
HTTP 请求 → Handler → Logic → RPC Client → 响应
```

**文件阅读顺序：**

1. **Handler**：`app/market/api/internal/handler/coin_info_handler.go`
   ```go
   func CoinInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           // 1. 解析请求参数
           var req types.MarketReq
           httpx.ParseForm(r, &req)

           // 2. 调用 Logic 处理
           resp, err := logic.NewCoinInfoLogic(r.Context(), svcCtx).CoinInfo(&req)

           // 3. 返回响应
           httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
       }
   }
   ```

2. **Logic**：`app/market/api/internal/logic/coin_info_logic.go`
   ```go
   func (l *CoinInfoLogic) CoinInfo(req *types.MarketReq) (*types.Coin, error) {
       // 调用 RPC 服务
       coin, err := l.svcCtx.MarketClient.FindCoinInfo(l.ctx, &marketpb.MarketReq{Unit: req.Unit})
       // 转换响应
       return resp, nil
   }
   ```

3. **ServiceContext**：`app/market/api/internal/svc/service_context.go`
   ```go
   type ServiceContext struct {
       MarketClient marketpb.MarketClient  // RPC 客户端
   }
   ```

### 3.3 学习任务

- [ ] 阅读 `idl/api/market.api`，理解 API 定义语法
- [ ] 跟踪一个完整请求：`/market/coin-info`
- [ ] 理解 `types/types.go` 中的类型定义
- [ ] 修改一个 API 返回字段，重新编译测试

---

## 第四阶段：RPC 服务与 gRPC 详解（预计 5-7 天）

### 4.1 Proto 文件：`idl/rpc/market/market.proto`

Protocol Buffers 是 gRPC 的接口定义语言：

```protobuf
syntax = "proto3";

package market;

// 生成 Go 代码的包路径
option go_package = "grpc-common/market/types/market";

// 消息定义
message MarketReq {
    string ip = 1;
    string symbol = 2;
    string unit = 3;
}

message Coin {
    int32 id = 1;
    string name = 2;
    string unit = 3;
    double usd_rate = 4;
}

// 服务定义
service Market {
    rpc FindCoinInfo(MarketReq) returns(Coin);
    rpc FindAllCoin(MarketReq) returns(CoinList);
}
```

**Proto 语法要点：**
- `syntax = "proto3"` → 使用 proto3 版本
- `message` → 定义数据结构（请求/响应）
- `service` → 定义 RPC 服务接口
- 字段编号（`= 1`, `= 2`）用于序列化，不能重复使用

### 4.2 Proto 编译

```bash
# 安装 protoc 编译器
brew install protobuf

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 编译 proto 文件
protoc --go_out=. --go-grpc_out=. idl/rpc/market/market.proto
```

生成的文件：
- `market.pb.go` → 消息序列化代码
- `market_grpc.pb.go` → gRPC 服务端/客户端代码

### 4.3 RPC 服务入口：`app/market/rpc/main.go`

```go
func main() {
    // 1. 加载配置
    var c config.Config
    conf.MustLoad(*configFile, &c)

    // 2. 创建服务上下文
    ctx := svc.NewServiceContext(c)

    // 3. 创建 gRPC 服务器
    s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
        // 注册服务实现
        marketpb.RegisterMarketServer(grpcServer, server.NewMarketServer(ctx))
    })

    // 4. 启动服务
    s.Start()
}
```

### 4.4 RPC 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                    Server 层                             │
│  server/market_server.go - 接收 gRPC 请求               │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                    Logic 层                              │
│  logic/find_coin_info_logic.go - 业务编排               │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                 Domain Service 层                        │
│  domain/service/coin_service.go - 业务规则              │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  Repository 层                           │
│  repository/coin_repository.go - 数据访问               │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   Model 层                               │
│  model/coin.go - 数据模型                               │
└─────────────────────────────────────────────────────────┘
```

### 4.5 各层代码示例

**Server 层**：`app/market/rpc/internal/server/market_server.go`
```go
// RPC 门面，将请求分发到 Logic 层
func (s *MarketServer) FindCoinInfo(ctx context.Context, req *marketpb.MarketReq) (*marketpb.Coin, error) {
    return logic.NewFindCoinInfoLogic(ctx, s.svcCtx).FindCoinInfo(req)
}
```

**Logic 层**：`app/market/rpc/internal/logic/find_coin_info_logic.go`
```go
// 业务编排，调用 Domain Service
func (l *FindCoinInfoLogic) FindCoinInfo(req *marketpb.MarketReq) (*marketpb.Coin, error) {
    coin, err := l.coinService.FindCoinInfo(l.ctx, req.Unit)
    // 转换为 Proto 消息
    return resp, nil
}
```

**Domain Service 层**：`app/market/rpc/internal/domain/service/coin_service.go`
```go
// 业务规则，调用 Repository
func (s *CoinService) FindCoinInfo(ctx context.Context, unit string) (*model.Coin, error) {
    coin, err := s.repo.FindByUnit(ctx, unit)
    if coin == nil {
        return nil, errors.New("not support this coin")
    }
    return coin, nil
}
```

**Repository 层**：`app/market/rpc/internal/repository/coin_repository.go`
```go
// 数据访问，使用 sqlx
func (r *CoinRepository) FindByUnit(ctx context.Context, unit string) (*model.Coin, error) {
    var coin model.Coin
    err := r.db.GetContext(ctx, &coin, "SELECT * FROM coin WHERE unit=? LIMIT 1", unit)
    return &coin, nil
}
```

### 4.6 服务发现：Etcd

go-zero 使用 Etcd 进行服务注册与发现：

**配置文件** `app/market/rpc/etc/market.yaml`：
```yaml
Name: market.rpc
ListenOn: 0.0.0.0:8082

Etcd:
  Hosts:
    - etcd:2379
  Key: market.rpc
```

**客户端配置** `app/market/api/etc/market-api.yaml`：
```yaml
MarketRPC:
  Etcd:
    Hosts:
      - etcd:2379
    Key: market.rpc
```

### 4.7 学习任务

- [ ] 阅读 `idl/rpc/market/market.proto`，理解 Proto 语法
- [ ] 查看 `app/market/rpc/pb/market/market.pb.go`，理解生成代码
- [ ] 跟踪一个完整 RPC 调用：`FindCoinInfo`
- [ ] 理解 Etcd 服务发现的配置方式
- [ ] 尝试添加一个新的 RPC 方法

---

## 第五阶段：服务间通信（预计 3-5 天）

### 5.1 API 调用 RPC

**API 服务通过 RPC 客户端调用 RPC 服务**：

`app/market/api/internal/svc/service_context.go`：
```go
type ServiceContext struct {
    MarketClient marketpb.MarketClient
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        MarketClient: marketpb.NewMarketClient(zrpc.MustNewClient(c.MarketRPC).Connection()),
    }
}
```

### 5.2 RPC 调用 RPC

**一个 RPC 服务可以调用另一个 RPC 服务**：

`app/ucenter/rpc/internal/logic/reset_address_logic.go`：
```go
func (l *ResetAddressLogic) ResetAddress(req *assetpb.AssetReq) (*assetpb.AssetResp, error) {
    // ucenter-rpc 调用 market-rpc
    coin, err := l.svcCtx.MarketClient.FindCoinInfo(l.ctx, &marketpb.MarketReq{Unit: req.CoinName})

    // 本地业务处理
    wallet, err := l.svcCtx.WalletService.EnsureWalletBySymbol(l.ctx, req.UserId, req.CoinName, coin)

    return &assetpb.AssetResp{}, nil
}
```

### 5.3 调用链示例

```
用户请求：POST /uc/asset/reset-address

┌─────────────┐     HTTP      ┌─────────────┐
│  ucenter    │ ─────────────►│  ucenter    │
│    api      │               │    rpc      │
└─────────────┘               └──────┬──────┘
                                     │ gRPC
                              ┌──────▼──────┐
                              │   market    │
                              │    rpc      │
                              └──────┬──────┘
                                     │ SQL
                              ┌──────▼──────┐
                              │    MySQL    │
                              │  (coin表)   │
                              └─────────────┘
```

### 5.4 学习任务

- [ ] 阅读 `app/ucenter/rpc/internal/svc/service_context.go`，理解多 RPC 客户端配置
- [ ] 跟踪跨服务调用：`ucenter-api → ucenter-rpc → market-rpc`
- [ ] 理解 `NonBlock: true` 配置的作用（非阻塞调用）

---

## 第六阶段：异步任务与消息队列（预计 5-7 天）

### 6.1 架构设计

```
┌─────────────┐    Kafka     ┌─────────────┐
│  ucenter    │ ──────────► │  jobcenter  │
│    rpc      │  withdraw   │  consumer   │
└─────────────┘    topic    └──────┬──────┘
                                 │
                          ┌──────▼──────┐
                          │  Bitcoin    │
                          │    RPC      │
                          └─────────────┘
```

### 6.2 生产者：发布事件

`app/ucenter/rpc/internal/domain/service/withdraw_service.go`：
```go
func (s *WithdrawService) Apply(ctx context.Context, req *ApplyRequest) error {
    // 1. 冻结余额（数据库事务）
    tx, _ := s.txManager.Begin(ctx)
    s.walletRepo.FreezeBalance(tx, req.MemberID, req.Amount)

    // 2. 创建提现记录
    record := s.buildRecord(req)
    s.withdrawRepo.Save(tx, record)

    // 3. 提交事务
    tx.Commit()

    // 4. 发布 Kafka 事件（异步处理）
    event := json.Marshal(record)
    s.producer.PushWithKey(ctx, strconv.FormatInt(req.MemberID, 10), string(event))

    return nil
}
```

### 6.3 消费者：处理事件

`app/jobcenter/internal/consumer/withdraw_consumer.go`：
```go
func NewWithdrawConsumer(svcCtx *svc.ServiceContext) (coreservice.Service, error) {
    return kafka.NewConsumerService(
        svcCtx.Config.Kafka,
        func(ctx context.Context, message kafka.Message) error {
            // 解析消息
            var event model.WithdrawRecordEvent
            json.Unmarshal(message.Value, &event)

            // 调用领域服务处理
            return svcCtx.WithdrawService.ProcessApplied(ctx, &event)
        },
        classifyWithdrawError,  // 错误分类
    )
}
```

### 6.4 Kafka 封装

`pkg/mq/kafka/consumer.go` 提供了消费者抽象：

```go
// 消费动作
type ConsumeAction int

const (
    ConsumeAck        // 确认消息
    ConsumeRetry      // 重试
    ConsumeDeadLetter // 死信队列
)

// 错误分类器
type ErrorClassifier func(err error) ConsumeAction
```

### 6.5 定时任务

`app/jobcenter/internal/task/service.go`：
```go
func (s *Service) Start() {
    // 使用 goroutine + time.Ticker 实现定时任务
    for _, task := range s.tasks {
        go s.runPeriodic(task)
    }
}

func (s *Service) runPeriodic(task TaskFunc) {
    ticker := time.NewTicker(s.interval)
    for {
        select {
        case <-ticker.C:
            task(s.ctx)
        }
    }
}
```

### 6.6 学习任务

- [ ] 阅读 `pkg/mq/kafka/kafka.go`，理解生产者封装
- [ ] 阅读 `pkg/mq/kafka/consumer.go`，理解消费者封装
- [ ] 跟踪提现流程：`ucenter-rpc → Kafka → jobcenter → Bitcoin`
- [ ] 理解死信队列（Dead Letter）的作用
- [ ] 查看 Kafdrop 观察消息

---

## 第七阶段：基础设施封装（预计 3-5 天）

### 7.1 MySQL 封装

`pkg/db/mysqlx/mysql.go`：
```go
// 使用 sqlx 封装 MySQL 连接
func New(cfg Config) (*sqlx.DB, error) {
    db, err := sqlx.Open("mysql", cfg.DataSource)
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    return db, nil
}
```

`pkg/db/mysqlx/tx.go`：
```go
// 事务管理器接口
type TxManager interface {
    Begin(ctx context.Context) (*sqlx.Tx, error)
}

// 在事务中执行函数
func (m *txManager) RunInTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
    tx, _ := m.Begin(ctx)
    defer tx.Rollback()

    if err := fn(tx); err != nil {
        return err
    }
    return tx.Commit()
}
```

### 7.2 Redis 封装

`pkg/cache/redisx/redis.go`：
```go
type Client struct {
    rdb *redis.Client
}

// 缓存汇率
func (c *Client) GetRate(ctx context.Context, key string) (float64, error) {
    val, err := c.rdb.Get(ctx, key).Float64()
    if errors.Is(err, redis.Nil) {
        return 0, nil  // 缓存未命中
    }
    return val, nil
}
```

### 7.3 MongoDB 封装

`pkg/store/mongox/mongo.go`：
```go
// K线数据存储
type Client struct {
    client *mongo.Client
}

func (c *Client) Database() *mongo.Database {
    return c.client.Database("mscoin")
}
```

### 7.4 学习任务

- [ ] 阅读 `pkg/db/mysqlx/`，理解数据库连接池配置
- [ ] 阅读 `pkg/cache/redisx/`，理解 Redis 封装
- [ ] 阅读 `pkg/store/mongox/`，理解 MongoDB 封装
- [ ] 理解为什么要在 `pkg/` 中封装基础设施

---

## 第八阶段：综合实践（预计 5-7 天）

### 8.1 实践任务：添加新功能

**任务：添加「获取币种列表」功能**

1. **定义 Proto**（`idl/rpc/market/market.proto`）：
   ```protobuf
   message CoinListRequest {}

   service Market {
       rpc ListCoins(CoinListRequest) returns(CoinList);
   }
   ```

2. **实现 RPC Logic**（`app/market/rpc/internal/logic/list_coins_logic.go`）

3. **定义 API**（`idl/api/market.api`）：
   ```
   @handler ListCoins
   get /coins returns (CoinListResult)
   ```

4. **实现 API Logic**（`app/market/api/internal/logic/list_coins_logic.go`）

5. **测试验证**

### 8.2 实践任务：添加定时任务

**任务：每分钟同步 BTC 价格**

1. **配置任务**（`app/jobcenter/etc/jobcenter.yaml`）：
   ```yaml
   Tasks:
     BtcPrice:
       Enabled: true
       IntervalSeconds: 60
   ```

2. **实现任务逻辑**

### 8.3 实践任务：添加 Kafka 消费者

**任务：消费订单事件**

1. **定义消费者配置**
2. **实现消息处理逻辑**
3. **注册到 ServiceGroup**

---

## 第九阶段：进阶主题（预计 7-10 天）

### 9.1 go-zero 代码生成

```bash
# 安装 goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 生成 API 代码
goctl api go -api idl/api/market.api -dir app/market/api

# 生成 RPC 代码
goctl rpc protoc idl/rpc/market/market.proto --go_out=. --go-grpc_out=.
```

### 9.2 中间件开发

`app/ucenter/api/internal/middleware/auth_middleware.go`：
```go
// JWT 认证中间件
func AuthMiddleware(svcCtx *svc.ServiceContext) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("x-auth-token")
            userID, err := jwt.ParseUserID(token, svcCtx.Config.JWT.AccessSecret)
            if err != nil {
                http.Error(w, "unauthorized", 401)
                return
            }
            // 将 userID 注入上下文
            ctx := context.WithValue(r.Context(), "userId", userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 9.3 错误处理

`pkg/result/result.go`：
```go
// 统一响应格式
type Result struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

func New() *Result {
    return &Result{Code: 0, Message: "success"}
}

func (r *Result) Deal(data interface{}, err error) *Result {
    if err != nil {
        r.Code = -1
        r.Message = err.Error()
        return r
    }
    r.Data = data
    return r
}
```

### 9.4 学习任务

- [ ] 使用 goctl 生成新服务
- [ ] 编写自定义中间件
- [ ] 理解 go-zero 的错误处理机制
- [ ] 阅读 go-zero 官方文档

---

## 学习资源

### 官方文档

| 资源 | 链接 |
|------|------|
| go-zero 官方文档 | https://go-zero.dev/ |
| gRPC 官方文档 | https://grpc.io/docs/ |
| Protocol Buffers | https://protobuf.dev/ |
| Kafka 官方文档 | https://kafka.apache.org/documentation/ |

### 推荐阅读顺序

1. **go-zero 入门**：https://go-zero.dev/docs/concept
2. **gRPC 基础**：https://grpc.io/docs/languages/go/basics/
3. **Proto3 语法**：https://protobuf.dev/programming-guides/proto3/
4. **Kafka 概念**：https://kafka.apache.org/documentation/#gettingStarted

### 项目关键文件速查

| 功能 | 文件路径 |
|------|---------|
| API 定义 | `idl/api/*.api` |
| Proto 定义 | `idl/rpc/**/*.proto` |
| API 入口 | `app/*/api/main.go` |
| RPC 入口 | `app/*/rpc/main.go` |
| 服务配置 | `app/*/etc/*.yaml` |
| 业务逻辑 | `app/*/internal/logic/*.go` |
| 数据访问 | `app/*/internal/repository/*.go` |
| Kafka 生产者 | `pkg/mq/kafka/kafka.go` |
| Kafka 消费者 | `pkg/mq/kafka/consumer.go` |

---

## 学习建议

### 学习方法

1. **先运行，后理解**：先启动项目，看到效果，再深入代码
2. **跟踪请求**：从一个 HTTP 请求开始，跟踪完整调用链
3. **修改验证**：修改代码，观察变化，加深理解
4. **画架构图**：画出服务间的调用关系，理清架构

### 避免陷阱

1. **不要跳跃**：按顺序学习，基础不牢后面会很吃力
2. **不要只看不写**：一定要动手修改代码
3. **不要忽略配置**：配置文件是理解服务的关键
4. **不要死磕细节**：先理解整体，再深入细节

### 检验标准

每阶段结束后，问自己：
- [ ] 我能画出这个模块的架构图吗？
- [ ] 我能解释清楚数据流向吗？
- [ ] 我能独立添加类似功能吗？

---

祝你学习顺利！🚀
