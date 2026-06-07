// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含会员安全设置相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

// SecuritySettingLogic 是安全设置业务逻辑处理器。
//
// 该结构体负责查询用户的安全认证状态信息。
type SecuritySettingLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewSecuritySettingLogic 创建安全设置业务逻辑处理器实例。
func NewSecuritySettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecuritySettingLogic {
	return &SecuritySettingLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindSecuritySetting 执行查询安全设置业务逻辑。
//
// 查询流程：
//  1. 从 context 获取已认证的用户 ID
//  2. 调用 ucenter-rpc MemberClient 查询会员详情
//  3. 转换会员数据为安全设置响应格式
//
// 返回的安全信息：
//   - 实名认证状态：是否已实名认证、审核状态
//   - 手机认证状态：已认证手机号
//   - 邮箱认证状态：已认证邮箱
//   - 资金密码状态：是否已设置资金密码
//   - 账户验证状态：是否绑定银行卡/支付宝/微信
//
// RPC 调用：
//   - MemberClient.FindMemberById -> ucenter-rpc
//   - ucenter-rpc 负责：查询会员详细信息
//
// 用户身份获取：
//   - 使用 middleware.UserIDFromContext 获取用户 ID
//   - 用户 ID 由 Auth 中间件从 JWT Token 解析并注入
//
// 参数：
//   - _：空请求参数（用户 ID 从 context 获取）
//
// 返回：
//   - *types.MemberSecurity：安全设置信息
//   - error：查询失败时的错误信息
func (l *SecuritySettingLogic) FindSecuritySetting(_ *types.ApproveReq) (*types.MemberSecurity, error) {
	// 从 context 获取已认证的用户 ID
	// 该 ID 由 Auth 中间件从 JWT Token 解析并注入
	userID := middleware.UserIDFromContext(l.ctx)

	// 调用 ucenter-rpc MemberClient 查询会员详情
	memberRes, err := l.svcCtx.MemberClient.FindMemberById(l.ctx, &memberpb.MemberReq{MemberId: userID})
	if err != nil {
		return nil, err
	}

	// 构建 security 响应结构
	// 将 RPC 响应转换为前端所需的格式
	resp := &types.MemberSecurity{
		Username:        memberRes.Username,
		Id:              memberRes.Id,
		CreateTime:      formatMemberTime(memberRes.RegistrationTime),
		LoginVerified:   "true", // 登录验证状态默认为 true（已通过 Auth 中间件）
		RealAuditing:    boolString(memberRes.RealNameStatus == 1), // 实名审核中状态
		Avatar:          memberRes.Avatar,
		AccountVerified: boolString(memberRes.Bank != "" || memberRes.AliNo != "" || memberRes.Wechat != ""), // 账户验证状态
	}

	// 设置邮箱认证状态
	if memberRes.Email != "" {
		resp.EmailVerified = "true"
		resp.Email = memberRes.Email
	} else {
		resp.EmailVerified = "false"
	}

	// 设置资金密码状态
	// 资金密码用于提现等敏感操作的二次验证
	if memberRes.JyPassword != "" {
		resp.FundsVerified = "true"
	} else {
		resp.FundsVerified = "false"
	}

	// 设置手机认证状态
	if memberRes.MobilePhone != "" {
		resp.PhoneVerified = "true"
		resp.MobilePhone = memberRes.MobilePhone
	} else {
		resp.PhoneVerified = "false"
	}

	// 设置实名认证状态
	if memberRes.RealName != "" {
		resp.RealVerified = "true"
		resp.RealName = memberRes.RealName
	} else {
		resp.RealVerified = "false"
	}

	// 设置身份证号（脱敏显示）
	if memberRes.IdNumber != "" {
		resp.IdCard = maskIDCard(memberRes.IdNumber)
	}

	return resp, nil
}

// boolString 将布尔值转换为字符串 "true" 或 "false"。
//
// 为什么使用字符串而非布尔值：
//   - 前端约定使用字符串格式
//   - 保持与其他字段格式一致
//
// 参数：
//   - value：布尔值
//
// 返回：
//   - string："true" 或 "false"
func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// formatMemberTime 将毫秒时间戳格式化为字符串时间。
//
// 参数：
//   - millis：毫秒时间戳
//
// 返回：
//   - string：格式化后的时间字符串（yyyy-MM-dd HH:mm:ss）
func formatMemberTime(millis int64) string {
	if millis <= 0 {
		return ""
	}
	return time.UnixMilli(millis).Format("2006-01-02 15:04:05")
}

// maskIDCard 对身份证号进行脱敏处理。
//
// 脱敏规则：只显示前两位，其余用 * 替换。
// 例如：身份证号 "123456789012345678" 脱敏为 "12********"
//
// 为什么这样脱敏：
//   - 显示前两位可以让用户确认身份证号类型
//   - 隐藏其余部分保护用户隐私
//
// 参数：
//   - idNumber：原始身份证号
//
// 返回：
//   - string：脱敏后的身份证号
func maskIDCard(idNumber string) string {
	if len(idNumber) < 2 {
		return idNumber
	}
	return idNumber[:2] + "********"
}
