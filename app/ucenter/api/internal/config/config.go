// Package config 提供 ucenter-api 服务的配置结构定义。
//
// 该包定义了服务运行所需的配置项，包括：
//   - HTTP 服务器配置（端口、主机等）
//   - RPC 客户端配置（ucenter-rpc、market-rpc 连接信息）
//   - JWT 认证配置（密钥、过期时间）
//
// 配置文件通常使用 YAML 格式，通过 go-zero 的 conf 包加载。
package config

import (
	"github.com/zeromicro/go-zero/rest"
	marketconf "github.com/zeromicro/go-zero/zrpc"
)

// AuthConfig 定义 JWT 认证相关的配置。
//
// 该配置用于生成和验证用户登录后的 JWT Token，确保 API 请求的安全性。
type AuthConfig struct {
	// AccessSecret 是 JWT 签名密钥，用于生成和验证 Token。
	// 必须与 ucenter-rpc 服务使用相同的密钥，否则 Token 无法验证。
	AccessSecret string

	// AccessExpire 是 Token 过期时间（秒）。
	// 过期后用户需要重新登录获取新 Token。
	AccessExpire int64
}

// Config 是 ucenter-api 服务的完整配置结构。
//
// 配置层次：
//   - RestConf：go-zero REST 服务器基础配置（Host、Port、日志等）
//   - UcenterRPC：用户中心 RPC 服务连接配置
//   - MarketRPC：市场 RPC 服务连接配置
//   - JWT：认证相关配置
//
// 调用关系说明：
//   - UcenterRPC 连接提供：注册、登录、会员信息、资产、提现等核心功能
//   - MarketRPC 连接提供：币种信息、市场行情等查询功能
type Config struct {
	rest.RestConf

	// UcenterRPC 是用户中心 RPC 服务的客户端配置。
	// 该服务提供用户注册、登录验证、会员管理、钱包资产、提现处理等功能。
	UcenterRPC marketconf.RpcClientConf

	// MarketRPC 是市场 RPC 服务的客户端配置。
	// 该服务提供币种列表、币种详情、市场行情等查询功能，
	// 用于提现时获取币种的提现限额、手续费等信息。
	MarketRPC marketconf.RpcClientConf

	// JWT 是认证配置，包含 Token 签名密钥和过期时间。
	// API 层使用此配置验证请求中的 x-auth-token 头部。
	JWT AuthConfig
}
