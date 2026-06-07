// Package httputil 包含 HTTP 面向服务共享的小型辅助工具。
//
// 本包提供 HTTP 请求处理相关的实用函数，包括：
//   - ClientIP: 从 HTTP 请求中提取客户端真实 IP 地址
//
// 使用场景：
//   - API 网关获取客户端 IP 用于限流
//   - 记录用户访问日志
//   - 实现基于 IP 的安全策略
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
//
// IP 获取优先级：
//  1. X-Real-IP 头：通常由 Nginx 等反向代理设置
//  2. X-Forwarded-For 头的第一个 IP：可能由多级代理设置
//  3. RemoteAddr：直接连接的客户端地址
//
// 参数：
//   - r: HTTP 请求对象
//
// 返回值：
//   - string: 客户端 IP 地址（IPv4 格式）
//
// 使用示例：
//
//	func Handler(w http.ResponseWriter, r *http.Request) {
//	    ip := ClientIP(r)
//	    log.Printf("Request from %s", ip)
//	}
//
// 注意事项：
//   - X-Forwarded-For 可被客户端伪造，在生产环境中应配合可信代理列表使用
//   - IPv6 本地地址 "::1" 会被转换为 "127.0.0.1"
func ClientIP(r *http.Request) string {
	// 优先检查 X-Real-IP 头
	// 这是最可靠的来源，由反向代理直接设置
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return normalizeLoopback(ip)
	}

	// 其次检查 X-Forwarded-For 头
	// 格式：client, proxy1, proxy2
	// 取第一个 IP 作为客户端 IP
	if ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return normalizeLoopback(strings.TrimSpace(parts[0]))
		}
	}

	// 最后使用 RemoteAddr
	// 格式：IP:Port 或 [IPv6]:Port
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// 如果解析失败，直接返回原始地址
		return normalizeLoopback(r.RemoteAddr)
	}

	return normalizeLoopback(host)
}

// normalizeLoopback 将 IPv6 本地地址转换为 IPv4 格式。
//
// 为什么需要这个转换：
//   - Go 的 HTTP 服务器在本地访问时使用 IPv6 格式 "::1"
//   - 但许多日志系统和安全策略期望看到 "127.0.0.1"
//   - 统一格式便于日志分析和 IP 比较
//
// 参数：
//   - ip: IP 地址字符串
//
// 返回值：
//   - string: 规范化后的 IP 地址
func normalizeLoopback(ip string) string {
	// IPv6 本地地址转换为 IPv4
	if ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}