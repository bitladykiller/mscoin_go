package svc

import (
	"mscoin_go/app/exchange/rpc/internal/config"
	"mscoin_go/app/exchange/rpc/internal/domain/service"
	"mscoin_go/app/exchange/rpc/internal/repository"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"

	"github.com/jmoiron/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext wires all runtime dependencies for exchange-rpc.
type ServiceContext struct {
	Config config.Config

	DB    *sqlx.DB
	Cache *redisx.Client

	OrderService *service.OrderService

	MemberClient memberpb.MemberClient
	AssetClient  assetpb.AssetClient
	MarketClient marketpb.MarketClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := mysqlx.New(c.Mysql)
	if err != nil {
		panic(err)
	}

	ucClient := zrpc.MustNewClient(c.UcenterRPC)
	marketClient := zrpc.MustNewClient(c.MarketRPC)

	orderRepo := repository.NewOrderRepository(db)

	return &ServiceContext{
		Config:       c,
		DB:           db,
		Cache:        redisx.New(c.Redis),
		OrderService: service.NewOrderService(orderRepo),
		MemberClient: memberpb.NewMemberClient(ucClient.Conn()),
		AssetClient:  assetpb.NewAssetClient(ucClient.Conn()),
		MarketClient: marketpb.NewMarketClient(marketClient.Conn()),
	}
}

func (s *ServiceContext) Close() {
	if s.DB != nil {
		_ = s.DB.Close()
	}
}
