package logic

import (
	"context"
	"time"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// HistoryKlineLogic 处理 HistoryKline 请求。
//
// 业务用例：获取 K 线历史数据
//   - 根据交易对、时间范围、周期查询 K 线数据
//   - 用于 K 线图表展示、技术分析等场景
//
// 调用链路：
//   MarketServer -> HistoryKlineLogic -> MarketService.HistoryKline
type HistoryKlineLogic struct {
	marketLogicBase
}

// NewHistoryKlineLogic 创建 HistoryKlineLogic 实例。
func NewHistoryKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryKlineLogic {
	return &HistoryKlineLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// HistoryKline 执行获取 K 线历史数据的业务逻辑。
//
// 请求参数：
//   - req.Symbol：交易对标识，如 "BTCUSDT"
//   - req.From：开始时间（毫秒时间戳）
//   - req.To：结束时间（毫秒时间戳）
//   - req.Resolution：K 线周期
//     - "1"：1 分钟
//     - "5"：5 分钟
//     - "15"：15 分钟
//     - "30"：30 分钟
//     - "1H"（默认）：1 小时
//     - "1D"：1 天
//     - "1W"：1 周
//     - "1M"：1 月
//
// 返回数据：
//   - List：K 线数据数组，每项包含
//     - Time：时间戳
//     - Open/Close/High/Low：开高低收价格
//     - Volume：成交量
//
// 超时控制：
//   - 设置 10 秒超时，K 线数据可能较大
func (l *HistoryKlineLogic) HistoryKline(req *marketpb.MarketReq) (*marketpb.HistoryRes, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	list, err := l.marketService.HistoryKline(ctx, req.Symbol, req.From, req.To, req.Resolution)
	if err != nil {
		return nil, err
	}

	return &marketpb.HistoryRes{List: list}, nil
}
