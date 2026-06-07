// Package main 是 market RPC 服务的入口点。
//
// market RPC 服务提供市场数据查询能力，包括：
//   - 币种信息查询（Coin 相关）
//   - 交易对信息查询（ExchangeCoin 相关）
//   - K 线历史数据查询
//   - 法币汇率查询
//
// 服务架构采用分层设计：
//   - server 层：gRPC 服务端，接收请求并路由到 logic 层
//   - logic 层：业务编排，协调领域服务完成用例
//   - domain/service 层：业务规则，封装核心业务逻辑
//   - repository 层：数据访问，封装数据库操作
//   - model 层：数据模型，定义实体结构
//
// 服务依赖：
//   - MySQL：存储币种配置、交易对配置
//   - MongoDB：存储 K 线历史数据
//   - Redis：缓存汇率数据
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

// configFile 指定配置文件路径，默认为 etc/market.yaml。
// 可通过命令行参数 -f 覆盖。
var configFile = flag.String("f", "etc/market.yaml", "the config file")

// main 是服务的启动入口。
//
// 启动流程：
//  1. 解析命令行参数
//  2. 加载配置文件
//  3. 初始化服务上下文（数据库连接、领域服务等）
//  4. 创建 gRPC 服务器
//  5. 注册 MarketServer 和 ExchangeRateServer
//  6. 在开发/测试模式下启用 gRPC 反射
//  7. 启动服务监听
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册 MarketServer：处理币种、交易对、K 线等市场数据查询
		marketpb.RegisterMarketServer(grpcServer, server.NewMarketServer(ctx))
		// 注册 ExchangeRateServer：处理法币汇率查询
		ratepb.RegisterExchangeRateServer(grpcServer, server.NewExchangeRateServer(ctx))

		// 在开发和测试模式下启用 gRPC 反射，支持 grpcurl 等工具调试
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting market rpc server at %s...\n", c.ListenOn)
	s.Start()
}
