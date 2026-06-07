// Package httputil 包含 HTTP 面向服务共享的小型辅助工具。
package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP 从请求中解析调用者 IP。
//
// 为什么需要这个辅助函数：
//   - 传统处理器在下游 RPC 请求中使用调用者 IP
//   - 部署可能位于反向代理之后
//   - 在每个处理器中重复相同的解析代码会很冗余且容易出错
func ClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return normalizeLoopback(ip)
	}

	if ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return normalizeLoopback(strings.TrimSpace(parts[0]))
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return normalizeLoopback(r.RemoteAddr)
	}

	return normalizeLoopback(host)
}

func normalizeLoopback(ip string) string {
	if ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}
