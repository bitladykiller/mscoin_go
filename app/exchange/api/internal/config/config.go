// Package config 定义了 exchange-api 服务的配置结构。
// 包含 REST 服务器配置、RPC 客户端配置和 JWT 认证配置。
package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// AuthConfig 定义 JWT 认证相关配置。
type AuthConfig struct {
	// AccessSecret 是 JWT 签名使用的密钥。
	AccessSecret string
	// AccessExpire 是 JWT 令牌的过期时间（秒）。
	AccessExpire int64
}

// Config 是 exchange-api 服务的完整配置结构。
// 嵌入了 go-zero 的 REST 配置，并添加了 RPC 客户端和 JWT 配置。
type Config struct {
	rest.RestConf
	// ExchangeRPC 是 exchange-rpc 服务的客户端配置。
	ExchangeRPC zrpc.RpcClientConf
	// JWT 是认证中间件使用的 JWT 配置。
	JWT AuthConfig
}
