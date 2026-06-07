package svc

import (
	"mscoin_go/app/market/api/internal/config"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	ratepb "mscoin_go/app/market/rpc/pb/rate"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 将 HTTP handler 与其依赖的 RPC 客户端进行连接和装配。
type ServiceContext struct {
	Config       config.Config
	MarketClient marketpb.MarketClient
	RateClient   ratepb.ExchangeRateClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.MarketRPC)
	conn := client.Conn()
	return &ServiceContext{
		Config:       c,
		MarketClient: marketpb.NewMarketClient(conn),
		RateClient:   ratepb.NewExchangeRateClient(conn),
	}
}
