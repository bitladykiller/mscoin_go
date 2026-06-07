package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config defines the runtime configuration for the HTTP-facing market API.
type Config struct {
	rest.RestConf
	MarketRPC zrpc.RpcClientConf
}
