// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含检查登录状态相关的业务逻辑。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/pkg/auth"
)

// CheckLoginLogic 是检查登录状态业务逻辑处理器。
//
// 该结构体负责验证 JWT Token 是否有效。
type CheckLoginLogic struct {
	// ctx 是请求上下文。
	ctx    context.Context

	// svcCtx 是服务上下文，提供配置访问能力。
	svcCtx *svc.ServiceContext
}

// NewCheckLoginLogic 创建检查登录状态业务逻辑处理器实例。
func NewCheckLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckLoginLogic {
	return &CheckLoginLogic{ctx: ctx, svcCtx: svcCtx}
}

// CheckLogin 执行检查登录状态业务逻辑。
//
// 该方法验证 JWT Token 的有效性，用于前端判断用户登录状态。
//
// 验证流程：
//  1. 使用 JWT 密钥解析 Token
//  2. 解析成功则 Token 有效
//  3. 解析失败则 Token 无效或已过期
//
// 为什么不使用 Auth 中间件：
//   - 该接口的目的是返回 Token 是否有效的状态
//   - 使用中间件会在 Token 无效时直接返回错误
//   - 直接处理可以返回更友好的状态信息（true/false）
//
// 参数：
//   - token：JWT Token 字符串
//
// 返回：
//   - bool：Token 是否有效
//   - error：处理过程中的错误（通常不返回，Token 无效时返回 false）
func (l *CheckLoginLogic) CheckLogin(token string) (bool, error) {
	// 使用 auth 包解析 Token 获取用户 ID
	// ParseUserID 内部会验证：
	//   - 签名是否正确
	//   - Token 是否过期
	//   - 格式是否有效
	_, err := auth.ParseUserID(token, l.svcCtx.Config.JWT.AccessSecret)
	if err != nil {
		// Token 无效或已过期，返回 false
		return false, nil
	}
	// Token 有效
	return true, nil
}
