// Package okxx 集中管理重构后 MSCoin 服务使用的最小 OKX 市场数据集成。
//
// 本包提供与 OKX 交易所 API 的交互功能，包括：
//   - 汇率查询：获取 USD/CNY 实时汇率
//   - K 线数据：获取加密货币价格历史
//   - HMAC 签名认证：支持私有 API 调用
//
// API 端点：
//   - exchangeRatePath: /api/v5/market/exchange-rate 汇率查询
//   - candlesPath: /api/v5/market/candles K 线数据查询
//
// 认证说明：
//   - 公开 API（汇率、K 线）无需认证即可使用
//   - 如果提供了凭据，会自动添加签名头
//   - 凭据必须同时配置 APIKey、SecretKey 和 Passphrase
//
// 使用场景：
//   - jobcenter 服务查询汇率用于资产计价
//   - 价格预警和趋势分析
package okxx

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// API 端点路径常量
const (
	// exchangeRatePath 是 OKX 汇率查询 API 端点
	exchangeRatePath = "/api/v5/market/exchange-rate"

	// candlesPath 是 OKX K 线数据查询 API 端点
	candlesPath = "/api/v5/market/candles"
)

// Config 包含 OKX 端点和可选的凭据。
//
// 为什么凭据在这个包装器中是可选的：
//   - 当前迁移使用的市场数据端点是公开的
//   - 本地开发应该能够在不配置交易所密钥的情况下运行
//   - 如果部署确实提供了凭据，我们可以保留传统的签名头行为，而无需更改任务层代码
//
// 字段说明：
//   - APIKey: OKX API 密钥
//   - SecretKey: OKX API 密钥对应的密码
//   - Passphrase: OKX API 的口令
//   - Host: OKX API 主机地址，如 "https://www.okx.com"
//   - Proxy: 代理地址，用于网络受限环境
//   - TimeoutMs: 请求超时时间（毫秒）
type Config struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	Host       string
	Proxy      string
	TimeoutMs  int
}

// ExchangeRate 包含面向市场的汇率 API 所需的 USD/CNY 报价。
//
// 字段说明：
//   - USDCNY: 美元兑人民币的汇率值
type ExchangeRate struct {
	USDCNY float64
}

// Candle 以重构后服务内部使用的规范化格式捕获一条 OKX K 线数据。
//
// K 线（Candlestick）是金融市场的价格数据图表，
// 每条 K 线包含一段时间内的开盘价、最高价、最低价、收盘价等信息。
//
// 字段说明：
//   - Time: 时间戳（毫秒）
//   - OpenPrice: 开盘价
//   - HighestPrice: 最高价
//   - LowestPrice: 最低价
//   - ClosePrice: 收盘价
//   - Count: 交易笔数
//   - Volume: 成交量（以币计）
//   - Turnover: 成交额（以计价货币计）
type Candle struct {
	Time         int64
	OpenPrice    float64
	HighestPrice float64
	LowestPrice  float64
	ClosePrice   float64
	Count        float64
	Volume       float64
	Turnover     float64
}

// Client 是 jobcenter 依赖的外部市场数据接口。
//
// 设计为接口便于：
//   - 单元测试使用 mock 替换真实 API
//   - 未来切换到其他交易所而不影响业务代码
type Client interface {
	// FetchExchangeRate 获取 USD/CNY 汇率
	FetchExchangeRate(ctx context.Context) (*ExchangeRate, error)

	// FetchCandles 获取 K 线数据
	//
	// 参数：
	//   - instID: 产品 ID，如 "BTC-USDT"
	//   - bar: K 线周期，如 "1m"、"5m"、"1H"、"1D"
	FetchCandles(ctx context.Context, instID string, bar string) ([]*Candle, error)
}

// client 是 Client 接口的默认实现。
type client struct {
	cfg        Config
	httpClient *http.Client
}

// exchangeRateResponse 是汇率 API 的响应结构。
type exchangeRateResponse struct {
	Code string              `json:"code"` // OKX 状态码，"0" 表示成功
	Msg  string              `json:"msg"`  // 错误信息
	Data []exchangeRateEntry `json:"data"` // 汇率数据列表
}

// exchangeRateEntry 是汇率数据的单项结构。
type exchangeRateEntry struct {
	USDCNY string `json:"usdCny"` // USD/CNY 汇率，字符串格式
}

// candlesResponse 是 K 线 API 的响应结构。
type candlesResponse struct {
	Code string     `json:"code"` // OKX 状态码
	Msg  string     `json:"msg"`  // 错误信息
	Data [][]string `json:"data"` // K 线数据列表，每个元素是一条 K 线的字符串数组
}

// NewClient 构建一个可复用的 OKX 客户端。
//
// 参数：
//   - cfg: OKX 配置
//
// 返回值：
//   - Client: OKX 客户端实例
//   - error: 配置验证失败时返回错误
//
// 使用示例：
//
//	client, err := NewClient(Config{
//	    Host:      "https://www.okx.com",
//	    TimeoutMs: 30000,
//	})
//	if err != nil {
//	    // 处理错误
//	}
//	rate, err := client.FetchExchangeRate(ctx)
func NewClient(cfg Config) (Client, error) {
	// 主机地址必须配置
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, errors.New("okx host is required")
	}

	// 设置超时时间，默认 20 秒
	timeout := 20 * time.Second
	if cfg.TimeoutMs > 0 {
		timeout = time.Duration(cfg.TimeoutMs) * time.Millisecond
	}

	// 创建 HTTP 客户端
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// 如果配置了代理，设置代理传输层
	if strings.TrimSpace(cfg.Proxy) != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("parse okx proxy: %w", err)
		}
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	// 验证凭据配置完整性
	// 如果提供了部分凭据，返回错误
	// 凭据必须全部配置或全部不配置
	if providedCredentialsCount(cfg) > 0 && providedCredentialsCount(cfg) < 3 {
		return nil, errors.New("okx credentials must be configured together")
	}

	return &client{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

// FetchExchangeRate 获取 USD/CNY 实时汇率。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//
// 返回值：
//   - *ExchangeRate: 汇率数据
//   - error: API 调用失败或数据无效时返回错误
func (c *client) FetchExchangeRate(ctx context.Context) (*ExchangeRate, error) {
	var payload exchangeRateResponse
	if err := c.get(ctx, exchangeRatePath, nil, &payload); err != nil {
		return nil, err
	}

	// 检查 OKX API 状态码
	if payload.Code != "0" {
		return nil, fmt.Errorf("okx exchange-rate api failed: %s", payload.Msg)
	}

	// 验证数据存在
	if len(payload.Data) == 0 {
		return nil, errors.New("okx exchange-rate data is empty")
	}

	// 解析汇率值
	value, err := strconv.ParseFloat(strings.TrimSpace(payload.Data[0].USDCNY), 64)
	if err != nil {
		return nil, fmt.Errorf("parse okx usdCny: %w", err)
	}

	// 验证汇率值有效性
	if value <= 0 {
		return nil, fmt.Errorf("okx usdCny is invalid: %v", value)
	}

	return &ExchangeRate{USDCNY: value}, nil
}

// FetchCandles 获取指定产品的 K 线数据。
//
// 参数：
//   - ctx: 上下文
//   - instID: 产品 ID，如 "BTC-USDT"、"ETH-USDT"
//   - bar: K 线周期，支持 "1m"、"5m"、"15m"、"30m"、"1H"、"4H"、"1D"、"1W"、"1M"
//
// 返回值：
//   - []*Candle: K 线数据列表
//   - error: API 调用失败或数据无效时返回错误
//
// 使用示例：
//
//	candles, err := client.FetchCandles(ctx, "BTC-USDT", "1H")
func (c *client) FetchCandles(ctx context.Context, instID string, bar string) ([]*Candle, error) {
	// 参数验证和清理
	instID = strings.TrimSpace(instID)
	bar = strings.TrimSpace(bar)
	if instID == "" {
		return nil, errors.New("okx instId is required")
	}
	if bar == "" {
		return nil, errors.New("okx bar is required")
	}

	var payload candlesResponse
	if err := c.get(ctx, candlesPath, map[string]string{
		"instId": instID,
		"bar":    bar,
	}, &payload); err != nil {
		return nil, err
	}

	// 检查 OKX API 状态码
	if payload.Code != "0" {
		return nil, fmt.Errorf("okx candles api failed: %s", payload.Msg)
	}

	// 解析 K 线数据
	result := make([]*Candle, 0, len(payload.Data))
	for _, row := range payload.Data {
		// 每条 K 线包含 8 个字段
		if len(row) < 8 {
			return nil, fmt.Errorf("okx candle row is incomplete: %v", row)
		}

		// 解析各个字段
		timestamp, err := strconv.ParseInt(strings.TrimSpace(row[0]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse okx candle time: %w", err)
		}
		openPrice, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse okx candle open price: %w", err)
		}
		highestPrice, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse okx candle highest price: %w", err)
		}
		lowestPrice, err := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse okx candle lowest price: %w", err)
		}
		closePrice, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse okx candle close price: %w", err)
		}
		count, err := strconv.ParseFloat(strings.TrimSpace(row[5]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse okx candle count: %w", err)
		}
		volume, err := strconv.ParseFloat(strings.TrimSpace(row[6]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse okx candle volume: %w", err)
		}
		turnover, err := strconv.ParseFloat(strings.TrimSpace(row[7]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse okx candle turnover: %w", err)
		}

		result = append(result, &Candle{
			Time:         timestamp,
			OpenPrice:    openPrice,
			HighestPrice: highestPrice,
			LowestPrice:  lowestPrice,
			ClosePrice:   closePrice,
			Count:        count,
			Volume:       volume,
			Turnover:     turnover,
		})
	}

	return result, nil
}

// get 是通用的 HTTP GET 请求方法。
//
// 参数：
//   - ctx: 上下文
//   - path: API 路径
//   - query: 查询参数
//   - target: 响应解析目标
//
// 返回值：
//   - error: 请求失败或解析失败时返回错误
func (c *client) get(ctx context.Context, path string, query map[string]string, target any) error {
	// 构造完整 URL
	baseURL := strings.TrimRight(c.cfg.Host, "/") + path
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("parse okx url: %w", err)
	}

	// 添加查询参数
	if len(query) > 0 {
		values := reqURL.Query()
		for key, value := range query {
			values.Set(key, value)
		}
		reqURL.RawQuery = values.Encode()
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return fmt.Errorf("build okx request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 添加签名头（如果配置了凭据）
	c.setSignedHeaders(req, pathWithQuery(path, query))

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call okx api: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read okx response: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("okx api returned status %d", resp.StatusCode)
	}

	// 解析 JSON 响应
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode okx response: %w", err)
	}
	return nil
}

// setSignedHeaders 设置 OKX API 签名认证头。
//
// OKX API 签名算法：
//   1. 生成 UTC 时间戳
//   2. 构造签名字串：timestamp + method + requestPath
//   3. 使用 HMAC-SHA256 算法签名
//   4. Base64 编码签名结果
//
// 必须设置的头：
//   - OK-ACCESS-KEY: API 密钥
//   - OK-ACCESS-SIGN: 签名
//   - OK-ACCESS-TIMESTAMP: 时间戳
//   - OK-ACCESS-PASSPHRASE: 口令
func (c *client) setSignedHeaders(req *http.Request, requestPath string) {
	// 只有完整配置凭据时才添加签名头
	if providedCredentialsCount(c.cfg) != 3 {
		return
	}

	// 生成 UTC 时间戳（ISO 8601 格式）
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// 构造签名字串并计算签名
	signature := computeSignature(timestamp+"GET"+requestPath, c.cfg.SecretKey)

	// 设置认证头
	req.Header.Set("OK-ACCESS-KEY", c.cfg.APIKey)
	req.Header.Set("OK-ACCESS-SIGN", signature)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", c.cfg.Passphrase)
}

// pathWithQuery 构造带查询参数的完整路径。
// 用于签名计算，签名必须包含完整的请求路径（含查询参数）。
func pathWithQuery(path string, query map[string]string) string {
	if len(query) == 0 {
		return path
	}
	values := url.Values{}
	for key, value := range query {
		values.Set(key, value)
	}
	return path + "?" + values.Encode()
}

// computeSignature 计算 HMAC-SHA256 签名并 Base64 编码。
//
// 参数：
//   - message: 待签名的消息
//   - secret: 密钥
//
// 返回值：
//   - string: Base64 编码的签名
func computeSignature(message string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// providedCredentialsCount 统计已配置的凭据数量。
// 用于验证凭据配置的完整性。
func providedCredentialsCount(cfg Config) int {
	count := 0
	if strings.TrimSpace(cfg.APIKey) != "" {
		count++
	}
	if strings.TrimSpace(cfg.SecretKey) != "" {
		count++
	}
	if strings.TrimSpace(cfg.Passphrase) != "" {
		count++
	}
	return count
}