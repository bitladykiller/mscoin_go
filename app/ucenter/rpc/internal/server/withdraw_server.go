// Package server 定义提现 RPC 服务端。
package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// WithdrawServer 提现 RPC 服务端
// 处理提现相关的 RPC 请求
//
// 实现 withdrawpb.WithdrawServer 接口
// 提供以下方法：
//   - FindAddressByCoinId: 根据币种 ID 查询提现地址
//   - SendCode: 发送提现验证码
//   - WithdrawCode: 处理提现申请
//   - WithdrawRecord: 查询提现记录
type WithdrawServer struct {
	svcCtx *svc.ServiceContext          // 服务上下文
	withdrawpb.UnimplementedWithdrawServer // 未实现方法的默认实现
}

// NewWithdrawServer 创建提现服务端实例
func NewWithdrawServer(svcCtx *svc.ServiceContext) *WithdrawServer {
	return &WithdrawServer{svcCtx: svcCtx}
}

// FindAddressByCoinId 根据币种 ID 查询提现地址
// 接收 gRPC 请求并转发给 FindAddressByCoinIDLogic 处理
// 用于提现页面展示会员保存的提现地址列表
func (s *WithdrawServer) FindAddressByCoinId(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.AddressSimpleList, error) {
	return logic.NewFindAddressByCoinIDLogic(ctx, s.svcCtx).FindAddressByCoinID(in)
}

// SendCode 发送提现验证码
// 接收 gRPC 请求并转发给 SendWithdrawCodeLogic 处理
// 验证码有效期 5 分钟
func (s *WithdrawServer) SendCode(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	return logic.NewSendWithdrawCodeLogic(ctx, s.svcCtx).SendCode(in)
}

// WithdrawCode 处理提现申请
// 接收 gRPC 请求并转发给 WithdrawCodeLogic 处理
// 在事务中冻结余额、创建提现记录、发布 Kafka 事件
func (s *WithdrawServer) WithdrawCode(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	return logic.NewWithdrawCodeLogic(ctx, s.svcCtx).WithdrawCode(in)
}

// WithdrawRecord 查询提现记录
// 接收 gRPC 请求并转发给 WithdrawRecordLogic 处理
// 支持分页查询
func (s *WithdrawServer) WithdrawRecord(ctx context.Context, in *withdrawpb.WithdrawReq) (*withdrawpb.RecordList, error) {
	return logic.NewWithdrawRecordLogic(ctx, s.svcCtx).WithdrawRecord(in)
}
