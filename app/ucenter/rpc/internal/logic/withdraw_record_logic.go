// Package logic 定义提现记录查询业务逻辑处理器。
package logic

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// WithdrawRecordLogic 提现记录查询逻辑处理器
// 处理会员提现记录查询 RPC 请求
type WithdrawRecordLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewWithdrawRecordLogic 创建逻辑处理器实例
func NewWithdrawRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawRecordLogic {
	return &WithdrawRecordLogic{ctx: ctx, svcCtx: svcCtx}
}

// WithdrawRecord 查询会员提现记录
// 调用 WithdrawService.FindRecordList 处理查询逻辑
//
// 参数：
//   - req: 提现请求，包含用户 ID、分页参数
//
// 返回：
//   - RecordList: 提现记录列表和总数
//   - error: 错误信息
//
// 排序规则：按创建时间倒序，最新申请在前
func (l *WithdrawRecordLogic) WithdrawRecord(req *withdrawpb.WithdrawReq) (*withdrawpb.RecordList, error) {
	// 调用 WithdrawService 查询提现记录
	// 传入获取币种信息的回调函数
	list, total, err := l.svcCtx.WithdrawService.FindRecordList(l.ctx, req.UserId, req.Page, req.PageSize, func(ctx context.Context, coinID int64) (*marketpb.Coin, error) {
		// 从 Market RPC 获取币种信息
		return l.svcCtx.MarketClient.FindCoinById(ctx, &marketpb.MarketReq{Id: coinID})
	})
	if err != nil {
		return nil, err
	}

	return &withdrawpb.RecordList{
		List:  list,
		Total: total,
	}, nil
}
