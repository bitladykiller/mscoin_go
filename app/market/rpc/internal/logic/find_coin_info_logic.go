package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// FindCoinInfoLogic 处理 FindCoinInfo 请求。
//
// 业务用例：根据 unit 查询币种配置信息
//   - 根据 unit（如 "BTC"、"ETH"）查询币种详情
//   - 返回充值/提现开关、费率、限制等配置
//
// 调用链路：
//   MarketServer -> FindCoinInfoLogic -> CoinService.FindCoinInfo
type FindCoinInfoLogic struct {
	marketLogicBase
}

// NewFindCoinInfoLogic 创建 FindCoinInfoLogic 实例。
func NewFindCoinInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindCoinInfoLogic {
	return &FindCoinInfoLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// FindCoinInfo 执行查询币种详情的业务逻辑。
//
// 请求参数：
//   - req.Unit：币种单位标识，如 "BTC"、"ETH"、"USDT"
//
// 返回数据包括：
//   - Name/NameCN：币种英文名和中文名
//   - Unit：币种单位
//   - CanWithdraw/CanRecharge/CanTransfer：提现、充值、转账开关
//   - MaxWithdrawAmount/MinWithdrawAmount：提现限额
//   - MaxTxFee/MinTxFee：手续费范围
//   - Status：币种状态
//
// 错误情况：
//   - 币种不存在时返回 "not support this coin: {unit}" 错误
//
// 超时控制：
//   - 设置 5 秒超时，防止数据库查询阻塞
func (l *FindCoinInfoLogic) FindCoinInfo(req *marketpb.MarketReq) (*marketpb.Coin, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	coin, err := l.coinService.FindCoinInfo(ctx, req.Unit)
	if err != nil {
		return nil, err
	}

	resp := &marketpb.Coin{}
	if err := copier.Copy(resp, coin); err != nil {
		return nil, err
	}
	return resp, nil
}
