package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// FindExchangeCoinVisibleLogic 处理 FindExchangeCoinVisible 请求。
//
// 业务用例：获取所有可见的交易对
//   - 查询 visible=1 的交易对
//   - 用于交易对选择器、行情列表等场景
//
// 调用链路：
//   MarketServer -> FindExchangeCoinVisibleLogic -> ExchangeCoinService.FindVisible
type FindExchangeCoinVisibleLogic struct {
	marketLogicBase
}

// NewFindExchangeCoinVisibleLogic 创建 FindExchangeCoinVisibleLogic 实例。
func NewFindExchangeCoinVisibleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindExchangeCoinVisibleLogic {
	return &FindExchangeCoinVisibleLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// FindExchangeCoinVisible 执行获取可见交易对列表的业务逻辑。
//
// 返回数据：
//   - List：可见交易对数组，每个交易对包含
//     - Symbol：交易对标识
//     - BaseSymbol/coinSymbol：基础币种和计价币种
//     - CoinScale/BaseCoinScale：精度配置
//     - Fee：交易手续费率
//     - Enable：是否启用交易
//     - 各种交易限制配置
//
// 与 FindSymbolInfo 的区别：
//   - 本方法返回列表，FindSymbolInfo 返回单个
//   - 本方法只返回 visible=1 的交易对
func (l *FindExchangeCoinVisibleLogic) FindExchangeCoinVisible(*marketpb.MarketReq) (*marketpb.ExchangeCoinRes, error) {
	list, err := l.exchangeCoinService.FindVisible(l.ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]*marketpb.ExchangeCoin, len(list))
	if err := copier.Copy(&resp, list); err != nil {
		return nil, err
	}

	return &marketpb.ExchangeCoinRes{List: resp}, nil
}
