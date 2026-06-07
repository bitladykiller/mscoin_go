// Package logic 提供 market-api 服务的业务逻辑实现。
// 每个逻辑结构体负责调用后端 RPC 服务并处理返回数据。
//
// 架构说明：
//   - logic 层是 API 层的核心，负责业务逻辑编排
//   - 通过 ServiceContext 中注入的 RPC 客户端调用后端服务
//   - 处理数据转换和错误封装
//
// 设计模式：
//   - 使用嵌入 marketLogicBase 基类复用通用字段
//   - 每个业务场景一个 Logic 结构体，职责单一
//   - 工厂函数创建实例，便于测试时 mock
//
// 调用链：
//
//	HTTP Request -> Handler -> Logic -> RPC Client -> Backend Service
package logic

import (
	"context"

	"mscoin_go/app/market/api/internal/svc"
)

// marketLogicBase 是所有 market 业务逻辑结构体的基类。
// 提供通用的上下文和服务依赖字段，通过结构体嵌入实现继承。
//
// 设计目的：
//   - 复用通用的 context 和 svcCtx 字段
//   - 避免每个 Logic 结构体重复定义相同字段
//   - 便于统一扩展通用功能（如日志、追踪等）
//
// 使用方式：
//
//	type SomeLogic struct {
//	    marketLogicBase  // 嵌入基类
//	    // 可添加特定字段
//	}
type marketLogicBase struct {
	// ctx 请求上下文，用于传递请求范围的数据
	// 支持超时控制、取消信号、追踪信息等
	ctx context.Context

	// svcCtx 服务上下文，包含 RPC 客户端等依赖
	// 通过它可以访问 MarketClient 和 RateClient
	svcCtx *svc.ServiceContext
}

// newMarketLogicBase 创建 marketLogicBase 基类实例。
// 这是一个包内私有函数，仅供各个 Logic 工厂函数调用。
//
// 参数：
//   - ctx: 请求上下文，从 HTTP handler 传递过来
//   - svcCtx: 服务上下文，包含配置和 RPC 客户端
//
// 返回：
//   - marketLogicBase: 初始化完成的基类实例
func newMarketLogicBase(ctx context.Context, svcCtx *svc.ServiceContext) marketLogicBase {
	return marketLogicBase{ctx: ctx, svcCtx: svcCtx}
}