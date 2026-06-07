package btcx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	defaultMinConfirmations = 0
	defaultMaxConfirmations = 999999
	jsonRPCVersion          = "1.0"
)

// NodeConfig 描述一个 Bitcoin Core JSON-RPC 端点的配置。
type NodeConfig struct {
	URL              string
	Username         string
	Password         string
	MinConfirmations int
	MaxConfirmations int
	TimeoutMs        int
	AddressType      string
}

// WithdrawSender 抽象一个链特定的提现广播器。
//
// 为什么 jobcenter 依赖这个抽象而不是直接依赖原始 JSON-RPC 调用：
//   - 领域服务不应该知道传输细节，例如 JSON-RPC 请求格式
//   - 单元测试可以用确定性的 fake 替换 sender
//   - 未来的 ETH/TRON 等 sender 可以遵循相同的编排契约
type WithdrawSender interface {
	Send(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error)
}

// AddressAllocator 抽象钱包管理的地址分配。
//
// 为什么需要专门的抽象：
//   - `ucenter` 必须创建属于同一个 Bitcoin Core 钱包的 BTC 地址，
//     该钱包稍后被 `jobcenter` 用于签署提现交易
//   - 本地生成密钥对会将地址所有权与签名责任分离，
//     导致运行时 `signrawtransactionwithwallet` 失败
//   - 测试可以用确定性的 fake 替换 allocator，
//     而不将业务逻辑耦合到 JSON-RPC 细节
type AddressAllocator interface {
	Allocate(ctx context.Context, label string) (string, error)
}

type rpcClient interface {
	ListUnspent(ctx context.Context, min int, max int, addresses []string) ([]ListUnspentResult, error)
	CreateRawTransaction(ctx context.Context, inputs []Input, outputs []map[string]any) (string, error)
	SignRawTransactionWithWallet(ctx context.Context, hexTx string) (*SignRawTransactionWithWalletResult, error)
	SendRawTransaction(ctx context.Context, signedHex string) (string, error)
}

type addressRPCClient interface {
	GetNewAddress(ctx context.Context, label string, addressType string) (string, error)
}

type sender struct {
	client           rpcClient
	minConfirmations int
	maxConfirmations int
}

type addressAllocator struct {
	client      addressRPCClient
	addressType string
}

type jsonRPCClient struct {
	url      string
	username string
	password string
	client   *http.Client
}

// --- [JSON-RPC 模型] --- //

type rpcRequest struct {
	ID      string `json:"id"`
	Method  string `json:"method"`
	JSONRPC string `json:"jsonrpc"`
	Params  []any  `json:"params"`
}

type rpcResponse[T any] struct {
	ID     string    `json:"id"`
	Error  *rpcError `json:"error"`
	Result T         `json:"result"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ListUnspentResult 描述 Bitcoin Core 返回的一个可花费 UTXO。
type ListUnspentResult struct {
	Txid          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Address       string  `json:"address"`
	Amount        float64 `json:"amount"`
	Confirmations int     `json:"confirmations"`
}

// Input 匹配 `createrawtransaction` 所需的 JSON-RPC 格式。
type Input struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

// SignRawTransactionWithWalletResult 是迁移所需的 Bitcoin Core 签名响应的子集。
type SignRawTransactionWithWalletResult struct {
	Hex      string `json:"hex"`
	Complete bool   `json:"complete"`
}

// NewWithdrawSender 构建 jobcenter 使用的默认 BTC 提现发送器。
func NewWithdrawSender(cfg NodeConfig) (WithdrawSender, error) {
	if err := validateNodeConfig(cfg); err != nil {
		return nil, err
	}

	timeout := 20 * time.Second
	if cfg.TimeoutMs > 0 {
		timeout = time.Duration(cfg.TimeoutMs) * time.Millisecond
	}

	return &sender{
		client: &jsonRPCClient{
			url:      cfg.URL,
			username: cfg.Username,
			password: cfg.Password,
			client:   &http.Client{Timeout: timeout},
		},
		minConfirmations: chooseConfirmations(cfg.MinConfirmations, defaultMinConfirmations),
		maxConfirmations: chooseConfirmations(cfg.MaxConfirmations, defaultMaxConfirmations),
	}, nil
}

// NewAddressAllocator 构建一个基于 Bitcoin Core 的地址分配器。
//
// 当前迁移在 `ucenter` 中使用此功能，以便新重置的 BTC 地址由节点钱包创建，
// 而不是在本地生成。
func NewAddressAllocator(cfg NodeConfig) (AddressAllocator, error) {
	if err := validateNodeConfig(cfg); err != nil {
		return nil, err
	}

	timeout := 20 * time.Second
	if cfg.TimeoutMs > 0 {
		timeout = time.Duration(cfg.TimeoutMs) * time.Millisecond
	}

	return &addressAllocator{
		client: &jsonRPCClient{
			url:      cfg.URL,
			username: cfg.Username,
			password: cfg.Password,
			client:   &http.Client{Timeout: timeout},
		},
		addressType: defaultAddressType(cfg.AddressType),
	}, nil
}

// --- [提现流程] --- //

// Send 选择足够的 UTXO，构造 BTC 交易，用节点钱包签名，然后广播它。
//
// Sender 保持与传统实现相同的值语义：
// `totalAmount` 包含申请时已预留的矿工费，
// `arrivedAmount` 是转移到外部目标地址的金额。
func (s *sender) Send(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error) {
	if strings.TrimSpace(fromAddress) == "" {
		return "", errors.New("source address is required")
	}
	if strings.TrimSpace(toAddress) == "" {
		return "", errors.New("destination address is required")
	}
	if totalAmount <= 0 {
		return "", errors.New("total amount must be greater than zero")
	}
	if arrivedAmount <= 0 {
		return "", errors.New("arrived amount must be greater than zero")
	}
	if arrivedAmount > totalAmount {
		return "", errors.New("arrived amount cannot exceed total amount")
	}

	utxos, err := s.client.ListUnspent(ctx, s.minConfirmations, s.maxConfirmations, []string{fromAddress})
	if err != nil {
		return "", err
	}

	inputs, totalInput, err := collectInputs(utxos, totalAmount)
	if err != nil {
		return "", err
	}

	outputs := []map[string]any{
		{toAddress: arrivedAmount},
	}
	changeAmount := floorFloat(totalInput-totalAmount, 10)
	if changeAmount > 0 {
		outputs = append(outputs, map[string]any{fromAddress: changeAmount})
	}

	rawHex, err := s.client.CreateRawTransaction(ctx, inputs, outputs)
	if err != nil {
		return "", err
	}

	signed, err := s.client.SignRawTransactionWithWallet(ctx, rawHex)
	if err != nil {
		return "", err
	}
	if signed == nil || !signed.Complete || strings.TrimSpace(signed.Hex) == "" {
		return "", errors.New("bitcoin transaction signature is incomplete")
	}

	return s.client.SendRawTransaction(ctx, signed.Hex)
}

func collectInputs(utxos []ListUnspentResult, targetAmount float64) ([]Input, float64, error) {
	var total float64
	inputs := make([]Input, 0, len(utxos))

	for _, utxo := range utxos {
		total += utxo.Amount
		inputs = append(inputs, Input{
			Txid: utxo.Txid,
			Vout: utxo.Vout,
		})
		if total >= targetAmount {
			return inputs, total, nil
		}
	}

	return nil, 0, errors.New("insufficient bitcoin utxo balance")
}

func floorFloat(value float64, precision int) float64 {
	if precision < 0 {
		precision = 0
	}
	ratio := math.Pow10(precision)
	return math.Floor(value*ratio) / ratio
}

func chooseConfirmations(value int, fallback int) int {
	if value > 0 || fallback == 0 {
		return value
	}
	return fallback
}

func validateNodeConfig(cfg NodeConfig) error {
	if strings.TrimSpace(cfg.URL) == "" {
		return errors.New("bitcoin node url is required")
	}
	if strings.TrimSpace(cfg.Username) == "" {
		return errors.New("bitcoin node username is required")
	}
	if strings.TrimSpace(cfg.Password) == "" {
		return errors.New("bitcoin node password is required")
	}
	return nil
}

func defaultAddressType(raw string) string {
	value := strings.TrimSpace(raw)
	if value != "" {
		return value
	}

	// `legacy` 保留旧 MSCoin 代码库使用的传统 Base58 格式 BTC 地址，
	// 同时仍将所有权委托给 Bitcoin Core。
	return "legacy"
}

// --- [JSON-RPC 调用] --- //

func (c *jsonRPCClient) ListUnspent(ctx context.Context, min int, max int, addresses []string) ([]ListUnspentResult, error) {
	var result []ListUnspentResult
	if err := c.call(ctx, "listunspent", []any{min, max, addresses}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *jsonRPCClient) CreateRawTransaction(ctx context.Context, inputs []Input, outputs []map[string]any) (string, error) {
	var result string
	if err := c.call(ctx, "createrawtransaction", []any{inputs, outputs}, &result); err != nil {
		return "", err
	}
	return result, nil
}

func (c *jsonRPCClient) SignRawTransactionWithWallet(ctx context.Context, hexTx string) (*SignRawTransactionWithWalletResult, error) {
	var result SignRawTransactionWithWalletResult
	if err := c.call(ctx, "signrawtransactionwithwallet", []any{hexTx}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *jsonRPCClient) SendRawTransaction(ctx context.Context, signedHex string) (string, error) {
	var result string
	if err := c.call(ctx, "sendrawtransaction", []any{signedHex, 0}, &result); err != nil {
		return "", err
	}
	return result, nil
}

func (c *jsonRPCClient) GetNewAddress(ctx context.Context, label string, addressType string) (string, error) {
	var result string
	if err := c.call(ctx, "getnewaddress", []any{label, addressType}, &result); err != nil {
		return "", err
	}
	return result, nil
}

func (a *addressAllocator) Allocate(ctx context.Context, label string) (string, error) {
	if a == nil || a.client == nil {
		return "", errors.New("bitcoin address allocator is not initialized")
	}
	return a.client.GetNewAddress(ctx, label, a.addressType)
}

func (c *jsonRPCClient) call(ctx context.Context, method string, params []any, result any) error {
	requestBody, err := json.Marshal(rpcRequest{
		ID:      "mscoin_go",
		Method:  method,
		JSONRPC: jsonRPCVersion,
		Params:  params,
	})
	if err != nil {
		return fmt.Errorf("marshal bitcoin rpc request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("build bitcoin rpc request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call bitcoin rpc %s: %w", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read bitcoin rpc %s response: %w", method, err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("bitcoin rpc %s returned status %d", method, resp.StatusCode)
	}

	var envelope rpcResponse[json.RawMessage]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode bitcoin rpc %s response: %w", method, err)
	}
	if envelope.Error != nil {
		return fmt.Errorf("bitcoin rpc %s failed: %s", method, envelope.Error.Message)
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("decode bitcoin rpc %s result: %w", method, err)
	}
	return nil
}
