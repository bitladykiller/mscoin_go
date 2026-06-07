package config

import (
	"mscoin_go/pkg/btcx"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/mq/kafka"

	"github.com/zeromicro/go-zero/zrpc"
)

type AuthConfig struct {
	AccessSecret string
	AccessExpire int64
}

type CaptchaConfig struct {
	Vid string
	Key string
}

type Config struct {
	zrpc.RpcServerConf
	Mysql     mysqlx.Config
	Redis     redisx.Config
	Kafka     kafka.Config
	JWT       AuthConfig
	Captcha   CaptchaConfig
	MarketRPC zrpc.RpcClientConf
	Bitcoin   btcx.NodeConfig
}
