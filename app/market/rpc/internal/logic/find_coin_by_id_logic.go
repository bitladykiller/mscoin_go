package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// FindCoinByIDLogic 处理 FindCoinById 请求。
//
// 业务用例：根据 ID 查询币种信息
//   - 根据币种 ID 查询币种详情
//   - 用于已知币种 ID 的场景（如订单查询后获取币种信息）
//
// 调用链路：
//   MarketServer -> FindCoinByIDLogic -> CoinService.FindCoinByID
type FindCoinByIDLogic struct {
	marketLogicBase
}

// NewFindCoinByIDLogic 创建 FindCoinByIDLogic 实例。
func NewFindCoinByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindCoinByIDLogic {
	return &FindCoinByIDLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// FindByID 执行根据 ID 查询币种详情的业务逻辑。
//
// 请求参数：
//   - req.Id：币种 ID（数据库主键）
//
// 返回数据：与 FindCoinInfo 相同的币种完整信息
//
// 错误情况：
//   - 币种不存在时返回 "not support this coin" 错误
//
// 超时控制：
//   - 设置 5 秒超时
//
// 与 FindCoinInfo 的区别：
//   - 本方法通过 ID 查询，FindCoinInfo 通过 unit 查询
//   - 适用于不同业务场景
func (l *FindCoinByIDLogic) FindByID(req *marketpb.MarketReq) (*marketpb.Coin, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	coin, err := l.coinService.FindCoinByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &marketpb.Coin{}
	if err := copier.Copy(resp, coin); err != nil {
		return nil, err
	}
	return resp, nil
}
