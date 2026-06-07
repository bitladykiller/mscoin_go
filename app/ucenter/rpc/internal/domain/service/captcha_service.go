package service

import (
	"encoding/json"

	"mscoin_go/pkg/httpxutil"
)

// captchaRequest 验证码验证请求结构
type captchaRequest struct {
	Id        string `json:"id"`        // 验证码服务 ID
	Secretkey string `json:"secretkey"` // 验证码服务密钥
	Scene     int    `json:"scene"`     // 验证场景
	Token     string `json:"token"`     // 验证码 Token
	Ip        string `json:"ip"`        // 客户端 IP 地址
}

// captchaResponse 验证码验证响应结构
type captchaResponse struct {
	Success int `json:"success"` // 验证结果，1 表示成功
}

// CaptchaService 验证码服务
type CaptchaService struct{}

// NewCaptchaService 创建验证码服务实例
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{}
}

// Verify 验证验证码
// 参数：
//   - server: 验证码服务器地址
//   - vid: 验证码服务 ID
//   - key: 验证码服务密钥
//   - token: 客户端提交的验证码 Token
//   - scene: 验证场景
//   - ip: 客户端 IP 地址
// 返回：
//   - true: 验证通过
//   - false: 验证失败
func (s *CaptchaService) Verify(server string, vid string, key string, token string, scene int, ip string) bool {
	if server == "" {
		return true
	}
	body, err := httpxutil.PostJSON(server, &captchaRequest{
		Id:        vid,
		Secretkey: key,
		Scene:     scene,
		Token:     token,
		Ip:        ip,
	})
	if err != nil {
		return false
	}

	var resp captchaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return false
	}
	return resp.Success == 1
}
