package logic

import (
	"context"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// FindSymbolThumbTrendLogic 处理 FindSymbolThumbTrend 请求。
//
// 业务用例：获取首页市场概览数据
//   - 查询所有可见交易对
//   - 为每个交易对计算当日价格摘要和趋势线
//   - 用于前端首页展示
//
// 调用链路：
//   MarketServer -> FindSymbolThumbTrendLogic -> MarketService.SymbolThumbTrend
type FindSymbolThumbTrendLogic struct {
	marketLogicBase
}

// NewFindSymbolThumbTrendLogic 创建 FindSymbolThumbTrendLogic 实例。
func NewFindSymbolThumbTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindSymbolThumbTrendLogic {
	return &FindSymbolThumbTrendLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// FindSymbolThumbTrend 执行获取交易对缩略图趋势的业务逻辑。
//
// 返回数据包括：
//   - Symbol：交易对标识
//   - Open/High/Low/Close：当日开盘、最高、最低、收盘价
//   - Chg：涨跌幅百分比
//   - Change：涨跌额
//   - Trend：价格趋势线数据
//   - Volume/Turnover：成交量和成交额
//
// 注意：当 K 线数据缺失时，返回空的缩略图而不是错误，
// 确保单个交易对数据问题不影响整体列表展示。
func (l *FindSymbolThumbTrendLogic) FindSymbolThumbTrend(*marketpb.MarketReq) (*marketpb.SymbolThumbRes, error) {
	list, err := l.marketService.SymbolThumbTrend(l.ctx)
	if err != nil {
		return nil, err
	}

	return &marketpb.SymbolThumbRes{List: list}, nil
}
