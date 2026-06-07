# 测试

推荐命令：

```bash
make tidy    # 整理依赖
make fmt     # 格式化代码
make vet     # 静态检查
make test    # 运行测试
make build   # 构建项目
make docker-up    # 启动 Docker 容器
```

当前迁移基线应至少通过以下目录的测试：

- `./app/market/api`
- `./app/market/rpc`
- `./app/exchange/api`
- `./app/exchange/rpc`
- `./app/ucenter/api`
- `./app/ucenter/rpc`
- `./app/jobcenter`

近期 `ucenter` 新增的直接单元测试覆盖：

- 提现支持币种 API 聚合
- 提现验证码 API 映射
- 提现申请领域编排，包括验证码验证、余额冻结、记录持久化和 Kafka 发布失败处理
- 提现历史 API 映射
- 资产重置地址 API 映射
- 资产重置地址 RPC 逻辑
- 资产获取地址 RPC 逻辑
- 提现记录和地址簿模型映射
- 提现记录申请模型构建
- MySQL 事务管理器守卫路径
- Kafka 生产者参数验证
- BTC 地址生成辅助工具

近期 `jobcenter` 新增的直接单元测试覆盖：

- Kafka 消费者参数验证、重试策略、提交行为和死信处理
- BTC 提现发送者验证和交易流程编排
- OKX 市场数据客户端验证、汇率解析和蜡烛图解析
- 基于 goroutine 的任务调度器立即运行行为
- 汇率同步领域编排
- K 线同步领域编排
- 提现异步领域编排，包括：
  - 已提交记录可见性验证
  - Redis txid 恢复路径
  - 仅 BTC 执行路由
  - 最终 MySQL 成功更新路径
- 提现消费者错误分类

近期 `market-rpc` 新增的直接单元测试覆盖：

- Redis 支持的法币汇率查找，带有优雅降级行为

近期 BTC 钱包对齐新增的直接单元测试覆盖：

- `ucenter-rpc` 重置地址逻辑现在验证 Bitcoin Core 支持的地址分配，而非本地私钥生成
- `pkg/btcx` 现已覆盖提现发送者验证和地址分配器引导验证路径

容器验证注意事项：

- `docker-compose.yml` 期望 Bitcoin Core RPC 在 `wallet/mscoin` 下运行
- `deploy/bitcoin/init-wallet.sh` 在 `ucenter-rpc` 和 `jobcenter` 启动前创建或加载该钱包
- `deploy/mysql/init/*.sql` 创建迁移服务查询和运行 K 线同步所需的最小数据库、模式和种子行
- 每个微服务镜像由其自己在 `app/*/{api,rpc,}/Dockerfile` 下的服务级 Dockerfile 构建