package config

import (
	"mscoin_go/pkg/btcx"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/mq/kafka"

	"github.com/zeromicro/go-zero/zrpc"
)

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

// Config ucenter RPC 服务配置
type Config struct {
	zrpc.RpcServerConf
	Mysql     mysqlx.Config      // MySQL 数据库配置
	Redis     redisx.Config      // Redis 缓存配置
	Kafka     kafka.Config       // Kafka 消息队列配置
	JWT       AuthConfig         // JWT 认证配置
	Captcha   CaptchaConfig      // 验证码服务配置
	MarketRPC zrpc.RpcClientConf // Market RPC 客户端配置
	Bitcoin   btcx.NodeConfig    // Bitcoin 节点配置
}
