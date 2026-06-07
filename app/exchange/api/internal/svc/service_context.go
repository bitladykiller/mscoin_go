package svc

import (
	"mscoin_go/app/exchange/api/internal/config"
	"mscoin_go/app/exchange/api/internal/middleware"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	Auth        rest.Middleware
	OrderClient orderpb.OrderClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.ExchangeRPC)
	return &ServiceContext{
		Config:      c,
		Auth:        middleware.NewAuthMiddleware(c.JWT.AccessSecret).Handle,
		OrderClient: orderpb.NewOrderClient(client.Conn()),
	}
}
