package service

import (
	"encoding/json"

	"mscoin_go/pkg/httpxutil"
)

type captchaRequest struct {
	Id        string `json:"id"`
	Secretkey string `json:"secretkey"`
	Scene     int    `json:"scene"`
	Token     string `json:"token"`
	Ip        string `json:"ip"`
}

type captchaResponse struct {
	Success int `json:"success"`
}

type CaptchaService struct{}

func NewCaptchaService() *CaptchaService {
	return &CaptchaService{}
}

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
