package config

import (
	"github.com/zeromicro/go-zero/rest"
	marketconf "github.com/zeromicro/go-zero/zrpc"
)

type AuthConfig struct {
	AccessSecret string
	AccessExpire int64
}

type Config struct {
	rest.RestConf
	UcenterRPC marketconf.RpcClientConf
	MarketRPC  marketconf.RpcClientConf
	JWT        AuthConfig
}
