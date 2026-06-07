// Package svc 定义了 exchange-api 的服务上下文。
// ServiceContext 聚合了所有运行时依赖，包括配置、中间件和 RPC 客户端。
package svc

import (
	"mscoin_go/app/exchange/api/internal/config"
	"mscoin_go/app/exchange/api/internal/middleware"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 聚合 exchange-api 服务的所有运行时依赖。
type ServiceContext struct {
	// Config 是服务配置。
	Config config.Config
	// Auth 是认证中间件，用于验证 JWT 令牌。
	Auth rest.Middleware
	// OrderClient 是 exchange-rpc 服务的客户端，用于订单操作。
	OrderClient orderpb.OrderClient
}

// NewServiceContext 创建 ServiceContext 实例。
// 初始化流程：
// 1. 创建 exchange-rpc 客户端连接
// 2. 创建认证中间件
// 3. 创建订单 RPC 客户端
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建 exchange-rpc 客户端连接
	client := zrpc.MustNewClient(c.ExchangeRPC)
	return &ServiceContext{
		Config:      c,
		Auth:        middleware.NewAuthMiddleware(c.JWT.AccessSecret).Handle,
		OrderClient: orderpb.NewOrderClient(client.Conn()),
	}
}
