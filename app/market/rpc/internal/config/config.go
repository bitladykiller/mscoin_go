// Package config defines the market RPC service configuration.
package config

import (
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/store/mongox"

	"github.com/zeromicro/go-zero/zrpc"
)

// Config groups all runtime dependencies required by the market RPC service.
//
// Keeping infrastructure configuration strongly typed is important in this
// refactor because the old project configured each service a little
// differently. The new project intentionally makes those differences explicit.
type Config struct {
	zrpc.RpcServerConf
	Mysql mysqlx.Config
	Mongo mongox.Config
	Redis redisx.Config
}
