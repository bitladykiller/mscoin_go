package main

import (
	"flag"
	"fmt"

	"mscoin_go/app/market/rpc/internal/config"
	"mscoin_go/app/market/rpc/internal/server"
	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	ratepb "mscoin_go/app/market/rpc/pb/rate"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/market.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		marketpb.RegisterMarketServer(grpcServer, server.NewMarketServer(ctx))
		ratepb.RegisterExchangeRateServer(grpcServer, server.NewExchangeRateServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting market rpc server at %s...\n", c.ListenOn)
	s.Start()
}
