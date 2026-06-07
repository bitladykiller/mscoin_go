// Package logic 定义交易记录查询业务逻辑处理器。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// FindTransactionLogic 查询交易记录的逻辑处理器
// 处理会员交易记录查询 RPC 请求
type FindTransactionLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewFindTransactionLogic 创建逻辑处理器实例
func NewFindTransactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTransactionLogic {
	return &FindTransactionLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindTransaction 查询会员交易记录
// 调用 TransactionService.FindTransaction 处理查询逻辑
//
// 参数：
//   - req: 资产请求，包含用户 ID、分页参数、筛选条件
//
// 返回：
//   - MemberTransactionList: 交易记录列表和总数
//   - error: 错误信息
//
// 筛选条件：
//   - symbol: 币种符号
//   - startTime/endTime: 时间范围
//   - type: 交易类型（RECHARGE/WITHDRAW/TRANSFER_ACCOUNTS/EXCHANGE）
func (l *FindTransactionLogic) FindTransaction(req *assetpb.AssetReq) (*assetpb.MemberTransactionList, error) {
	// 调用 TransactionService 查询交易记录
	// 支持多条件筛选和分页
	list, total, err := l.svcCtx.TransactionService.FindTransaction(
		l.ctx,
		req.UserId,
		req.PageNo,
		req.PageSize,
		req.Symbol,
		req.StartTime,
		req.EndTime,
		req.Type,
	)
	if err != nil {
		return nil, err
	}

	return &assetpb.MemberTransactionList{
		List:  list,
		Total: total,
	}, nil
}
