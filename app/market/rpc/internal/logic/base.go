// Package logic 包含 gRPC 请求的业务编排逻辑。
//
// Logic 层职责：
//   - 接收来自 Server 层的请求
//   - 编排一个或多个领域服务完成业务用例
//   - 处理超时、错误转换等横切关注点
//   - 将领域模型转换为 protobuf 响应
//
// 调用链路：
//
//	Server -> Logic -> Domain Service -> Repository
//
// 设计说明：
//   - 每个 RPC 方法对应一个 Logic 结构体
//   - marketLogicBase 提供公共依赖和工具方法
//   - 使用 go-zero 的模式生成器风格，便于扩展
package logic

import (
	"context"

	"mscoin_go/app/market/rpc/internal/domain/service"
	"mscoin_go/app/market/rpc/internal/svc"
)

// marketLogicBase 聚合 market RPC logic 文件中使用的领域服务。
//
// 这样既保持每个生成式 logic 文件精简，又使依赖关系显式化。
//
// 包含的依赖：
//   - ctx：请求上下文，用于传播取消和超时
//   - svcCtx：服务上下文，包含配置等
//   - coinService：币种领域服务
//   - exchangeCoinService：交易对领域服务
//   - marketService：市场数据领域服务
//   - rateService：汇率领域服务
type marketLogicBase struct {
	ctx                 context.Context
	svcCtx              *svc.ServiceContext
	coinService         *service.CoinService
	exchangeCoinService *service.ExchangeCoinService
	marketService       *service.MarketService
	rateService         *service.RateService
}

// newMarketLogicBase 创建 marketLogicBase 实例。
//
// 从服务上下文中提取各领域服务，供具体 Logic 使用。
// 这种依赖注入方式便于单元测试时 mock 领域服务。
func newMarketLogicBase(ctx context.Context, svcCtx *svc.ServiceContext) marketLogicBase {
	return marketLogicBase{
		ctx:                 ctx,
		svcCtx:              svcCtx,
		coinService:         svcCtx.CoinService,
		exchangeCoinService: svcCtx.ExchangeCoinService,
		marketService:       svcCtx.MarketService,
		rateService:         svcCtx.RateService,
	}
}
