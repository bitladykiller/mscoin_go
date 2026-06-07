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

// Config describes the first migrated jobcenter runtime.
//
// The initial refactor phase focuses on Kafka withdraw consumption, but the
// structure already reserves Mongo and Redis dependencies because later
// jobcenter responsibilities will reuse the same process skeleton.
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
