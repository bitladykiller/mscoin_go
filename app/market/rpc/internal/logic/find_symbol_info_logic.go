package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// FindSymbolInfoLogic 处理 FindSymbolInfo 请求。
//
// 业务用例：查询交易对配置信息
//   - 根据 symbol（如 "BTCUSDT"）查询交易对详情
//   - 返回交易精度、费率、交易限制等配置
//
// 调用链路：
//   MarketServer -> FindSymbolInfoLogic -> ExchangeCoinService.FindBySymbol
type FindSymbolInfoLogic struct {
	marketLogicBase
}

// NewFindSymbolInfoLogic 创建 FindSymbolInfoLogic 实例。
func NewFindSymbolInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindSymbolInfoLogic {
	return &FindSymbolInfoLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// FindSymbolInfo 执行查询交易对详情的业务逻辑。
//
// 请求参数：
//   - req.Symbol：交易对标识，格式为 "BASEQUOTE"，如 "BTCUSDT"、"ETHUSDT"
//
// 返回数据包括：
//   - Symbol：交易对标识
//   - BaseSymbol/coinSymbol：基础币种和计价币种
//   - CoinScale/BaseCoinScale：价格和数量精度
//   - Fee：交易手续费率
//   - Enable/Visible：启用和可见状态
//   - MinTurnover/MinVolume：最小成交额和最小交易量
//
// 错误情况：
//   - 交易对不存在时返回 "trading pair not found" 错误
func (l *FindSymbolInfoLogic) FindSymbolInfo(req *marketpb.MarketReq) (*marketpb.ExchangeCoin, error) {
	item, err := l.exchangeCoinService.FindBySymbol(l.ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	resp := &marketpb.ExchangeCoin{}
	if err := copier.Copy(resp, item); err != nil {
		return nil, err
	}
	return resp, nil
}
