// Package config 定义 market RPC 服务的配置结构。
//
// 配置层职责：
//   - 定义服务运行所需的配置项
//   - 继承 go-zero 的 RpcServerConf（包含 ListenOn、Etcd 等配置）
//   - 声明基础设施连接参数（MySQL、MongoDB、Redis）
//
// 配置文件位于 etc/market.yaml，由 main.go 加载解析。
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
//
// 字段说明：
//   - RpcServerConf：go-zero RPC 服务基础配置（监听地址、服务发现等）
//   - Mysql：MySQL 数据库配置，存储币种和交易对元数据
//   - Mongo：MongoDB 配置，存储 K 线历史数据
//   - Redis：Redis 缓存配置，缓存汇率等热点数据
type Config struct {
	zrpc.RpcServerConf
	Mysql mysqlx.Config
	Mongo mongox.Config
	Redis redisx.Config
}
