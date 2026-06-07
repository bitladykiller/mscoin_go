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

type ScheduleConfig struct {
	Enabled         bool
	RunOnStart      bool
	IntervalSeconds int
}

type KlineTaskConfig struct {
	ScheduleConfig
	Period        string
	PublishLatest bool
	PublishTopic  string
}

type TasksConfig struct {
	RateSync ScheduleConfig
	Klines   []KlineTaskConfig
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
