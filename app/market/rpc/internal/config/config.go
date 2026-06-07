// Package config 定义 market RPC 服务的配置。
package config

import (
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/store/mongox"

	"github.com/zeromicro/go-zero/zrpc"
)

// Config 聚合 market RPC 服务所需的所有运行时依赖。
//
// 在此次重构中，保持基础设施配置的强类型非常重要，
// 因为旧项目对每个服务的配置略有不同。
// 新项目特意将这些差异显式化，使配置更加清晰。
type Config struct {
	zrpc.RpcServerConf
	Mysql mysqlx.Config
	Mongo mongox.Config
	Redis redisx.Config
}
