// Package config 定义 ucenter RPC 服务的配置结构。
//
// 配置采用 go-zero 的声明式配置风格，通过 YAML 文件加载。
// 配置项包括：
//   - RPC 服务端配置（端口、超时等）
//   - MySQL 数据库连接配置
//   - Redis 缓存配置
//   - Kafka 消息队列配置
//   - JWT 认证配置
//   - 验证码服务配置
//   - Market RPC 客户端配置
//   - Bitcoin 节点配置
package config

import (
	"mscoin_go/pkg/btcx"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/mq/kafka"

	"github.com/zeromicro/go-zero/zrpc"
)

// AuthConfig JWT 认证配置
// 用于生成和验证用户登录 Token
type AuthConfig struct {
	AccessSecret string // JWT 签名密钥，用于生成和验证 Token 的签名
	AccessExpire int64  // JWT 过期时间（秒），控制 Token 的有效期
}

// CaptchaConfig 验证码服务配置
// 用于对接第三方验证码服务（如阿里云验证码）
type CaptchaConfig struct {
	Vid string // 验证码服务 ID，由验证码服务提供商分配
	Key string // 验证码服务密钥，用于验证请求的合法性
}

// Config ucenter RPC 服务配置
// 聚合了服务运行所需的所有配置项
type Config struct {
	zrpc.RpcServerConf                      // go-zero RPC 服务端配置（嵌入）
	Mysql     mysqlx.Config                 // MySQL 数据库配置，用于持久化会员、钱包等数据
	Redis     redisx.Config                 // Redis 缓存配置，用于缓存验证码、会话等临时数据
	Kafka     kafka.Config                  // Kafka 消息队列配置，用于发布提现等异步事件
	JWT       AuthConfig                    // JWT 认证配置，用于生成和验证登录 Token
	Captcha   CaptchaConfig                 // 验证码服务配置，用于人机验证
	MarketRPC zrpc.RpcClientConf            // Market RPC 客户端配置，用于获取币种市场信息
	Bitcoin   btcx.NodeConfig               // Bitcoin 节点配置，用于 BTC 地址分配
}
