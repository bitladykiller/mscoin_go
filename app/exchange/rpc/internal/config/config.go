// Package config 定义了 exchange-rpc 服务的配置结构。
// 包含 RPC 服务器配置、MySQL、Redis 以及其他 RPC 客户端配置。
package config

import (
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"

	"github.com/zeromicro/go-zero/zrpc"
)

// Config 定义 exchange-rpc 服务的运行时依赖配置。
type Config struct {
	zrpc.RpcServerConf
	// Mysql 是 MySQL 数据库连接配置。
	Mysql mysqlx.Config
	// Redis 是 Redis 缓存连接配置。
	Redis redisx.Config
	// UcenterRPC 是用户中心 RPC 服务配置，用于查询会员信息和钱包。
	UcenterRPC zrpc.RpcClientConf
	// MarketRPC 是行情 RPC 服务配置，用于查询交易对信息。
	MarketRPC zrpc.RpcClientConf
}
