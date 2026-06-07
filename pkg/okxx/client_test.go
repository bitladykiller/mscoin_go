package okxx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientValidatesHostAndCredentialShape(t *testing.T) {
	t.Parallel()

	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("NewClient() should fail when host is missing")
	}
	if _, err := NewClient(Config{
		Host:   "https://www.okx.com",
		APIKey: "only-key",
	}); err == nil {
		t.Fatal("NewClient() should fail when credentials are incomplete")
	}
}

func TestFetchExchangeRateReadsUSDCNYValue(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != exchangeRatePath {
			t.Fatalf("path = %q, want %q", r.URL.Path, exchangeRatePath)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "0",
			"msg":  "",
			"data": []map[string]string{
				{"usdCny": "7.1265"},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{Host: server.URL})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	rate, err := client.FetchExchangeRate(context.Background())
	if err != nil {
		t.Fatalf("FetchExchangeRate() error = %v", err)
	}
	if rate.USDCNY != 7.1265 {
		t.Fatalf("FetchExchangeRate().USDCNY = %v, want 7.1265", rate.USDCNY)
	}
}

func TestFetchCandlesParsesRows(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != candlesPath {
			t.Fatalf("path = %q, want %q", r.URL.Path, candlesPath)
		}
		if got := r.URL.Query().Get("instId"); got != "BTC-USDT" {
			t.Fatalf("instId = %q, want BTC-USDT", got)
		}
		if got := r.URL.Query().Get("bar"); got != "1m" {
			t.Fatalf("bar = %q, want 1m", got)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "0",
			"msg":  "",
			"data": [][]string{
				{"1710000000000", "1", "2", "0.5", "1.5", "10", "20", "30"},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{Host: server.URL})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	candles, err := client.FetchCandles(context.Background(), "BTC-USDT", "1m")
	if err != nil {
		t.Fatalf("FetchCandles() error = %v", err)
	}
	if len(candles) != 1 {
		t.Fatalf("FetchCandles() len = %d, want 1", len(candles))
	}
	if candles[0].ClosePrice != 1.5 || candles[0].Turnover != 30 {
		t.Fatalf("FetchCandles() candle = %+v, want close=1.5 turnover=30", candles[0])
	}
}
