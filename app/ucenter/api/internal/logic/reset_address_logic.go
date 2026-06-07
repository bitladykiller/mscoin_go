// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含重置充值地址相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// ResetAddressLogic 是重置充值地址业务逻辑处理器。
//
// 该结构体负责重置用户指定币种的充值地址。
type ResetAddressLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewResetAddressLogic 创建重置充值地址业务逻辑处理器实例。
func NewResetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetAddressLogic {
	return &ResetAddressLogic{ctx: ctx, svcCtx: svcCtx}
}

// ResetAddress 执行重置充值地址业务逻辑。
//
// 重置流程：
//  1. 从 context 获取已认证的用户 ID
//  2. 获取币种名称（支持 Unit 或 CoinName 参数）
//  3. 调用 ucenter-rpc AssetClient 重置地址
//
// 使用场景：
//   - 用户需要更换充值地址
//   - 原地址出现问题
//   - 安全考虑更换地址
//
// RPC 调用：
//   - AssetClient.ResetAddress -> ucenter-rpc
//   - ucenter-rpc 负责：
//       - 生成新充值地址
//       - 更新钱包记录
//       - 记录操作日志
//
// 安全考虑：
//   - 需要记录操作 IP，用于安全审计
//   - 可能需要限制重置频率
//
// 参数：
//   - req：资产请求，包含币种名称和操作 IP
//
// 返回：
//   - string：空字符串（成功时）
//   - error：重置失败时的错误信息
func (l *ResetAddressLogic) ResetAddress(req *types.AssetReq) (string, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 获取币种名称，支持两种参数格式
	// Unit：表单提交时的参数名
	// CoinName：URL 路径传递时的参数名
	coinName := req.Unit
	if coinName == "" {
		coinName = req.CoinName
	}

	// 从 context 获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)

	// 调用 ucenter-rpc AssetClient 重置充值地址
	_, err := l.svcCtx.AssetClient.ResetAddress(ctx, &assetpb.AssetReq{
		UserId:   userID,
		CoinName: coinName,
		Ip:       req.IP, // 记录操作 IP 用于安全审计
	})
	if err != nil {
		return "", err
	}
	return "", nil
}
