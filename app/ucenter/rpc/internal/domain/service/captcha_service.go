// Package service 定义验证码领域服务。
//
// CaptchaService 是人机验证码的服务，负责：
//   - 验证码验证：调用第三方验证码服务验证用户提交的验证码
//
// 使用场景：
//   - 登录：防止机器人暴力破解密码
//   - 注册：防止机器人批量注册
//   - 提现：防止机器人批量提现
//
// 设计原则：
//   - 简单服务：不依赖仓储，只有验证逻辑
//   - 可配置：验证码服务器地址可配置
//   - 安全性：验证码验证失败返回 false，不暴露具体原因
package service

import (
	"encoding/json"

	"mscoin_go/pkg/httpxutil"
)

// captchaRequest 验证码验证请求结构
// 用于构造发送给验证码服务器的请求
type captchaRequest struct {
	Id        string `json:"id"`        // 验证码服务 ID，由验证码服务提供商分配
	Secretkey string `json:"secretkey"` // 验证码服务密钥，用于验证请求的合法性
	Scene     int    `json:"scene"`     // 验证场景，如登录、注册、提现等
	Token     string `json:"token"`     // 验证码 Token，客户端提交的验证凭证
	Ip        string `json:"ip"`        // 客户端 IP 地址，用于风控
}

// captchaResponse 验证码验证响应结构
// 用于解析验证码服务器的响应
type captchaResponse struct {
	Success int `json:"success"` // 验证结果，1 表示成功，0 表示失败
}

// CaptchaService 验证码服务
// 负责调用第三方验证码服务验证用户提交的验证码
type CaptchaService struct{}

// NewCaptchaService 创建验证码服务实例
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{}
}

// Verify 验证验证码
// 调用第三方验证码服务验证用户提交的验证码
//
// 参数：
//   - server: 验证码服务器地址，如果为空则跳过验证（开发环境）
//   - vid: 验证码服务 ID
//   - key: 验证码服务密钥
//   - token: 客户端提交的验证码 Token
//   - scene: 验证场景（如登录、注册等）
//   - ip: 客户端 IP 地址
//
// 返回：
//   - true: 验证通过
//   - false: 验证失败
//
// 安全说明：
//   - 验证码服务器地址为空时返回 true（用于开发环境）
//   - 验证失败不暴露具体原因，防止攻击者获取信息
func (s *CaptchaService) Verify(server string, vid string, key string, token string, scene int, ip string) bool {
	// 验证码服务器地址为空时跳过验证
	// 用于开发环境或未配置验证码服务的场景
	if server == "" {
		return true
	}

	// 调用验证码服务器
	// 使用 POST 请求发送 JSON 格式的验证请求
	body, err := httpxutil.PostJSON(server, &captchaRequest{
		Id:        vid,
		Secretkey: key,
		Scene:     scene,
		Token:     token,
		Ip:        ip,
	})
	if err != nil {
		// 请求失败返回 false
		// 不暴露具体原因
		return false
	}

	// 解析响应
	// Success 字段为 1 表示验证成功
	var resp captchaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return false
	}
	return resp.Success == 1
}
