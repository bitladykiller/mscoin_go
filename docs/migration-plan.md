# 迁移计划

已完成：

- 仓库根骨架
- `idl/` 下的集中契约
- `market-api` 和 `market-rpc` 迁移示例
- `exchange-api` 和 `exchange-rpc` 迁移示例
- `ucenter-api` 和 `ucenter-rpc` 最小可行迁移

当前已迁移仓库中的 `ucenter` 范围：

- 发送注册验证码
- 手机号注册
- 登录
- 检查登录令牌
- 按 ID 查询会员
- 安全设置
- 钱包列表
- 按符号查询钱包
- 钱包重置地址
- 交易历史
- 提现支持币种信息
- 发送提现验证码
- 提现申请
- 提现记录

已完成的异步后续处理：

- `jobcenter` 标准 go-zero 工作器引导
- 共享 Kafka 消费者抽象
- 提现事件消费和 BTC 提现执行
- `jobcenter` 中基于 goroutine 的间隔任务调度器
- OKX 驱动的 USD/CNY 汇率同步
- OKX 驱动的 K 线同步
- Redis 支持的 `market-rpc` 法币汇率查找
- `ucenter-rpc` 中 Bitcoin Core 钱包支持的 BTC 地址分配
- 服务级 Dockerfile、Compose 栈、MySQL 初始化脚本和 Bitcoin 钱包引导

下一步：

1. 添加剩余的 Kafka 消费者/补偿重构
2. 扩展跨服务链路的集成测试
3. 在标准骨架上继续填充剩余业务模块