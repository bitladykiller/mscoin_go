// Package httpxutil 提供简化的 HTTP 客户端工具函数。
//
// 本包封装了 Go 标准库的 HTTP 客户端，提供：
//   - 简洁的 POST JSON 请求方法
//   - 自动超时控制
//   - 标准化的错误处理
//
// 使用场景：
//   - 调用第三方 API
//   - 服务间 HTTP 通信
//   - Webhook 回调
//
// 注意事项：
//   - 当前实现使用默认的 http.DefaultClient
//   - 在高并发场景下，建议使用自定义的 HTTP 客户端以控制连接池
package httpxutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// PostJSON 发送一个 POST 请求，请求体为 JSON 格式。
//
// 该方法会：
//   - 自动将 payload 编码为 JSON
//   - 设置 Content-Type: application/json 头
//   - 使用 20 秒超时
//   - 读取并返回完整响应体
//
// 参数：
//   - url: 目标 URL
//   - payload: 请求体，可以是任意可 JSON 序列化的值
//
// 返回值：
//   - []byte: 响应体字节
//   - error: 请求失败时返回错误
//
// 使用示例：
//
//	resp, err := PostJSON("https://api.example.com/users", map[string]any{
//	    "name": "test",
//	})
//	if err != nil {
//	    // 处理错误
//	}
//	// 处理响应
//
// 注意事项：
//   - 不区分成功和失败的 HTTP 状态码，都会返回响应体
//   - 如果需要根据状态码处理，请检查 resp 或使用自定义客户端
func PostJSON(url string, payload any) ([]byte, error) {
	// 创建 20 秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 编码请求体为 JSON
	// 忽略错误，因为 json.Marshal 对基本类型不会失败
	raw, _ := json.Marshal(payload)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	// 设置 Content-Type 头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	return io.ReadAll(resp.Body)
}