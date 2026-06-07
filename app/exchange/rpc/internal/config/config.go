package config

import (
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"

	"github.com/zeromicro/go-zero/zrpc"
)

// Config defines the runtime dependencies for exchange-rpc.
type Config struct {
	zrpc.RpcServerConf
	Mysql      mysqlx.Config
	Redis      redisx.Config
	UcenterRPC zrpc.RpcClientConf
	MarketRPC  zrpc.RpcClientConf
}
