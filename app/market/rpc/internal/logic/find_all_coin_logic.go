package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// FindAllCoinLogic 处理 FindAllCoin 请求。
//
// 业务用例：获取系统所有币种列表
//   - 查询所有配置的币种（不区分状态）
//   - 用于币种选择器、资产管理等场景
//
// 调用链路：
//   MarketServer -> FindAllCoinLogic -> CoinService.FindAll
type FindAllCoinLogic struct {
	marketLogicBase
}

// NewFindAllCoinLogic 创建 FindAllCoinLogic 实例。
func NewFindAllCoinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindAllCoinLogic {
	return &FindAllCoinLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// FindAllCoin 执行获取所有币种列表的业务逻辑。
//
// 返回所有币种的完整信息列表，包括：
//   - 币种基本信息（名称、单位等）
//   - 充值/提现/转账配置
//   - 费率和限额配置
//
// 注意：返回所有币种，包括已禁用的。
// 如果只需要启用的币种，调用方需要自行过滤。
func (l *FindAllCoinLogic) FindAllCoin(*marketpb.MarketReq) (*marketpb.CoinList, error) {
	list, err := l.coinService.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]*marketpb.Coin, len(list))
	if err := copier.Copy(&resp, list); err != nil {
		return nil, err
	}

	return &marketpb.CoinList{List: resp}, nil
}
