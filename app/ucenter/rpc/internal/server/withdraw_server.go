package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// WithdrawServer 提现 RPC 服务端
// 处理提现相关的 RPC 请求
type WithdrawServer struct {
	svcCtx *svc.ServiceContext
	withdrawpb.UnimplementedWithdrawServer
}

// NewWithdrawServer 创建提现服务端实例
func NewWithdrawServer(svcCtx *svc.ServiceContext) *WithdrawServer {
	return &WithdrawServer{svcCtx: svcCtx}
}

// FindAddressByCoinId 根据币种 ID 查询提现地址
func (s *WithdrawServer) FindAddressByCoinId(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.AddressSimpleList, error) {
	return logic.NewFindAddressByCoinIDLogic(ctx, s.svcCtx).FindAddressByCoinID(in)
}

// SendCode 发送提现验证码
func (s *WithdrawServer) SendCode(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	return logic.NewSendWithdrawCodeLogic(ctx, s.svcCtx).SendCode(in)
}

// WithdrawCode 处理提现申请
func (s *WithdrawServer) WithdrawCode(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	return logic.NewWithdrawCodeLogic(ctx, s.svcCtx).WithdrawCode(in)
}

// WithdrawRecord 查询提现记录
func (s *WithdrawServer) WithdrawRecord(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.RecordList, error) {
	return logic.NewWithdrawRecordLogic(ctx, s.svcCtx).WithdrawRecord(in)
}
