package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

type WithdrawServer struct {
	svcCtx *svc.ServiceContext
	withdrawpb.UnimplementedWithdrawServer
}

func NewWithdrawServer(svcCtx *svc.ServiceContext) *WithdrawServer {
	return &WithdrawServer{svcCtx: svcCtx}
}

func (s *WithdrawServer) FindAddressByCoinId(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.AddressSimpleList, error) {
	return logic.NewFindAddressByCoinIDLogic(ctx, s.svcCtx).FindAddressByCoinID(in)
}

func (s *WithdrawServer) SendCode(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	return logic.NewSendWithdrawCodeLogic(ctx, s.svcCtx).SendCode(in)
}

func (s *WithdrawServer) WithdrawCode(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	return logic.NewWithdrawCodeLogic(ctx, s.svcCtx).WithdrawCode(in)
}

func (s *WithdrawServer) WithdrawRecord(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.RecordList, error) {
	return logic.NewWithdrawRecordLogic(ctx, s.svcCtx).WithdrawRecord(in)
}
