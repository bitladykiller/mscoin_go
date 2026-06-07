package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 定义 market API 的 HTTP 服务运行时配置。
type Config struct {
	rest.RestConf
	MarketRPC zrpc.RpcClientConf
}
