// Package logic 定义提现地址查询业务逻辑处理器。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// FindAddressByCoinIDLogic 根据币种 ID 查询提现地址的逻辑处理器
// 处理查询会员保存的提现地址 RPC 请求
type FindAddressByCoinIDLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewFindAddressByCoinIDLogic 创建逻辑处理器实例
func NewFindAddressByCoinIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindAddressByCoinIDLogic {
	return &FindAddressByCoinIDLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindAddressByCoinID 根据币种 ID 查询会员提现地址
// 调用 WithdrawService.FindAddressByCoinID 处理查询逻辑
//
// 参数：
//   - req: 提现请求，包含用户 ID 和币种 ID
//
// 返回：
//   - AddressSimpleList: 提现地址列表
//   - error: 错误信息
//
// 使用场景：提现页面展示会员保存的提现地址，便于快速选择
func (l *FindAddressByCoinIDLogic) FindAddressByCoinID(req *withdrawpb.WithdrawReq) (*withdrawpb.AddressSimpleList, error) {
	// 调用 WithdrawService 查询提现地址列表
	list, err := l.svcCtx.WithdrawService.FindAddressByCoinID(l.ctx, req.UserId, req.CoinId)
	if err != nil {
		return nil, err
	}
	return &withdrawpb.AddressSimpleList{List: list}, nil
}
