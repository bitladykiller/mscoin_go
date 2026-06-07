// Package config 定义了 jobcenter 服务的运行时配置结构。
//
// 该包采用 go-zero 的配置加载机制，支持从 YAML 文件加载配置。
// 配置涵盖了数据库连接、消息队列、RPC 客户端、外部 API 等所有依赖。
package config

import (
	"mscoin_go/pkg/btcx"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/mq/kafka"
	"mscoin_go/pkg/okxx"
	"mscoin_go/pkg/store/mongox"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

// ScheduleConfig 定义定时任务的基础调度配置。
// 该配置可控制任务的启用状态、启动行为和执行间隔。
type ScheduleConfig struct {
	// Enabled 是否启用该定时任务。
	// 当为 false 时，任务不会启动，仅记录日志。
	Enabled bool

	// RunOnStart 服务启动时是否立即执行一次。
	// 适用于需要快速预热数据的场景，避免等待首个调度周期。
	RunOnStart bool

	// IntervalSeconds 任务执行间隔，单位为秒。
	// 如果值 <= 0，将使用默认值 60 秒。
	IntervalSeconds int
}

// KlineTaskConfig 定义 K 线同步任务的配置。
// 继承 ScheduleConfig 的调度能力，并增加 K 线特有的配置项。
type KlineTaskConfig struct {
	ScheduleConfig

	// Period K 线周期，如 "1m"、"5m"、"1h"、"1d" 等。
	// 该值与 OKX API 的 bar 参数对应。
	Period string

	// PublishLatest 是否将最新 K 线数据发布到 Kafka。
	// 启用后，每次同步完成会将最新一根 K 线推送到消息队列，
	// 供前端 WebSocket 订阅使用。
	PublishLatest bool

	// PublishTopic 发布最新 K 线的 Kafka Topic 名称。
	// 仅当 PublishLatest 为 true 时有效。
	PublishTopic string
}

// TasksConfig 聚合所有定时任务的配置。
type TasksConfig struct {
	// RateSync 汇率同步任务配置。
	// 从 OKX 获取 USDT/CNY 实时汇率并缓存到 Redis。
	RateSync ScheduleConfig

	// Klines K 线同步任务配置列表。
	// 支持配置多个不同周期的 K 线同步任务。
	Klines []KlineTaskConfig
}

// Config 描述首次迁移后的 jobcenter 运行时配置。
//
// 初始重构阶段主要关注 Kafka 提现消费，但结构已预留 Mongo 和 Redis 依赖，
// 因为后续 jobcenter 的其他职责将复用相同的进程骨架。
type Config struct {
	service.ServiceConf
	Kafka        kafka.ConsumerConfig
	UcenterMysql mysqlx.Config
	Redis        redisx.Config
	Mongo        mongox.Config
	MarketRPC    zrpc.RpcClientConf
	UcenterRPC   zrpc.RpcClientConf
	OKX          okxx.Config
	Tasks        TasksConfig
	Bitcoin      btcx.NodeConfig
}
