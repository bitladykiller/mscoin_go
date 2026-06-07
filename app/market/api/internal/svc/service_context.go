// Package svc 提供 market-api 服务的依赖注入容器（ServiceContext）。
// ServiceContext 负责初始化和管理服务运行时所需的所有依赖项，
// 包括配置对象和 RPC 客户端连接。
//
// 该包是连接 HTTP handler 层和 RPC 服务层的桥梁，
// handler 通过 ServiceContext 获取 RPC 客户端来调用后端服务。
package svc

import (
	"mscoin_go/app/market/api/internal/config"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	ratepb "mscoin_go/app/market/rpc/pb/rate"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 是 market-api 服务的依赖注入容器。
// 它持有服务运行时所需的所有依赖项，包括配置和 RPC 客户端。
//
// 设计模式：依赖注入（Dependency Injection）
// 通过将依赖集中管理，实现了：
//   - 配置与业务逻辑解耦
//   - 便于单元测试时 mock 依赖
//   - 统一的依赖生命周期管理
//
// 使用流程:
//  1. 在 main.go 中通过 NewServiceContext 创建实例
//  2. 将实例传递给 handler 注册函数
//  3. handler 通过 ServiceContext 访问 RPC 客户端
type ServiceContext struct {
	// Config 服务配置，包含 HTTP 服务器配置和 RPC 连接配置
	Config config.Config

	// MarketClient 是 market-rpc 服务的 gRPC 客户端。
	// 提供以下 RPC 方法：
	//   - FindCoinInfo: 查询币种信息
	//   - FindSymbolInfo: 查询交易对信息
	//   - FindSymbolThumbTrend: 查询行情缩略图和趋势数据
	//   - HistoryKline: 查询 K 线历史数据
	MarketClient marketpb.MarketClient

	// RateClient 是汇率服务的 gRPC 客户端。
	// 提供以下 RPC 方法：
	//   - UsdRate: 查询指定法币对 USD 的汇率
	RateClient ratepb.ExchangeRateClient
}

// NewServiceContext 创建并初始化 ServiceContext 实例。
// 这是 ServiceContext 的工厂函数，负责建立 RPC 连接并注入依赖。
//
// 参数:
//   - c: 服务配置，包含 RPC 连接信息
//
// 返回:
//   - *ServiceContext: 初始化完成的服务上下文
//
// 工作流程:
//  1. 使用配置中的 MarketRPC 信息创建 gRPC 客户端连接
//  2. 基于连接创建 market-rpc 的客户端桩（stub）
//  3. 基于连接创建 rate-rpc 的客户端桩（两个 RPC 共用同一个连接）
//  4. 返回包含所有依赖的 ServiceContext 实例
//
// 注意事项:
//   - 使用 MustNewClient，连接失败时会 panic
//   - MarketRPC 配置必须正确，否则服务无法启动
//   - MarketClient 和 RateClient 共用同一个 gRPC 连接，减少连接开销
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建 gRPC 客户端连接
	// MustNewClient 在连接失败时会 panic，确保服务启动时连接可用
	client := zrpc.MustNewClient(c.MarketRPC)

	// 获取底层 gRPC 连接
	// 该连接可被多个 client stub 复用
	conn := client.Conn()

	// 返回初始化完成的 ServiceContext
	// 使用同一连接创建两个 RPC 客户端，提高资源利用率
	return &ServiceContext{
		Config:       c,
		MarketClient: marketpb.NewMarketClient(conn),
		RateClient:   ratepb.NewExchangeRateClient(conn),
	}
}