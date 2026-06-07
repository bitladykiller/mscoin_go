// Package okxx centralizes the minimal OKX market-data integration used by the
// refactored MSCoin services.
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

const (
	exchangeRatePath = "/api/v5/market/exchange-rate"
	candlesPath      = "/api/v5/market/candles"
)

// Config contains the OKX endpoint and optional credentials.
//
// Why credentials are optional in this wrapper:
//   - the market-data endpoints used by the current migration are public
//   - local development should still run without provisioning exchange secrets
//   - if a deployment does provide credentials, we can preserve the legacy
//     signed-header behavior without changing task-layer code
type Config struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	Host       string
	Proxy      string
	TimeoutMs  int
}

// ExchangeRate contains the USD/CNY quote needed by market-facing rate APIs.
type ExchangeRate struct {
	USDCNY float64
}

// Candle captures one OKX K-line row in the normalized format the refactored
// services use internally.
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

// Client is the interface jobcenter depends on for external market data.
type Client interface {
	FetchExchangeRate(ctx context.Context) (*ExchangeRate, error)
	FetchCandles(ctx context.Context, instID string, bar string) ([]*Candle, error)
}

type client struct {
	cfg        Config
	httpClient *http.Client
}

type exchangeRateResponse struct {
	Code string              `json:"code"`
	Msg  string              `json:"msg"`
	Data []exchangeRateEntry `json:"data"`
}

type exchangeRateEntry struct {
	USDCNY string `json:"usdCny"`
}

type candlesResponse struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

// NewClient builds one reusable OKX client.
func NewClient(cfg Config) (Client, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, errors.New("okx host is required")
	}

	timeout := 20 * time.Second
	if cfg.TimeoutMs > 0 {
		timeout = time.Duration(cfg.TimeoutMs) * time.Millisecond
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}
	if strings.TrimSpace(cfg.Proxy) != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("parse okx proxy: %w", err)
		}
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	if providedCredentialsCount(cfg) > 0 && providedCredentialsCount(cfg) < 3 {
		return nil, errors.New("okx credentials must be configured together")
	}

	return &client{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

func (c *client) FetchExchangeRate(ctx context.Context) (*ExchangeRate, error) {
	var payload exchangeRateResponse
	if err := c.get(ctx, exchangeRatePath, nil, &payload); err != nil {
		return nil, err
	}
	if payload.Code != "0" {
		return nil, fmt.Errorf("okx exchange-rate api failed: %s", payload.Msg)
	}
	if len(payload.Data) == 0 {
		return nil, errors.New("okx exchange-rate data is empty")
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(payload.Data[0].USDCNY), 64)
	if err != nil {
		return nil, fmt.Errorf("parse okx usdCny: %w", err)
	}
	if value <= 0 {
		return nil, fmt.Errorf("okx usdCny is invalid: %v", value)
	}

	return &ExchangeRate{USDCNY: value}, nil
}

func (c *client) FetchCandles(ctx context.Context, instID string, bar string) ([]*Candle, error) {
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
	if payload.Code != "0" {
		return nil, fmt.Errorf("okx candles api failed: %s", payload.Msg)
	}

	result := make([]*Candle, 0, len(payload.Data))
	for _, row := range payload.Data {
		if len(row) < 8 {
			return nil, fmt.Errorf("okx candle row is incomplete: %v", row)
		}

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

func (c *client) get(ctx context.Context, path string, query map[string]string, target any) error {
	baseURL := strings.TrimRight(c.cfg.Host, "/") + path
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("parse okx url: %w", err)
	}

	if len(query) > 0 {
		values := reqURL.Query()
		for key, value := range query {
			values.Set(key, value)
		}
		reqURL.RawQuery = values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return fmt.Errorf("build okx request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setSignedHeaders(req, pathWithQuery(path, query))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call okx api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read okx response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("okx api returned status %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode okx response: %w", err)
	}
	return nil
}

func (c *client) setSignedHeaders(req *http.Request, requestPath string) {
	if providedCredentialsCount(c.cfg) != 3 {
		return
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	signature := computeSignature(timestamp+"GET"+requestPath, c.cfg.SecretKey)

	req.Header.Set("OK-ACCESS-KEY", c.cfg.APIKey)
	req.Header.Set("OK-ACCESS-SIGN", signature)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", c.cfg.Passphrase)
}

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

func computeSignature(message string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

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
