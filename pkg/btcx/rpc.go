// Package btcx 提供 Bitcoin Core 节点的 JSON-RPC 客户端功能。
//
// 本包实现了与 Bitcoin Core 节点通信的核心功能，包括：
//   - UTXO 查询：获取未花费的交易输出
//   - 原始交易创建：构造 BTC 转账交易
//   - 交易签名：使用节点钱包签名
//   - 交易广播：将签名后的交易发送到网络
//   - 地址生成：在节点钱包中创建新地址
//
// 架构设计：
//   - WithdrawSender 接口：抽象提现发送行为，便于测试和扩展
//   - AddressAllocator 接口：抽象地址分配，解耦业务与传输层
//
// 使用场景：
//   - jobcenter 服务使用 WithdrawSender 发起 BTC 提现
//   - ucenter 服务使用 AddressAllocator 为用户生成 BTC 充值地址
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

// 默认确认数配置常量
// 这些值用于筛选可花费的 UTXO
const (
	// defaultMinConfirmations 是筛选 UTXO 时的最小确认数下限
	// 设置为 0 表示包含未确认的交易
	defaultMinConfirmations = 0

	// defaultMaxConfirmations 是筛选 UTXO 时的最大确认数上限
	// 设置为 999999 实际上表示无上限
	defaultMaxConfirmations = 999999

	// jsonRPCVersion 是 Bitcoin Core JSON-RPC 协议版本
	jsonRPCVersion = "1.0"
)

// NodeConfig 描述一个 Bitcoin Core JSON-RPC 端点的配置。
//
// 字段说明：
//   - URL: Bitcoin Core 节点的 RPC 地址，如 "http://127.0.0.1:8332"
//   - Username: RPC 认证用户名，在 bitcoin.conf 中配置
//   - Password: RPC 认证密码，在 bitcoin.conf 中配置
//   - MinConfirmations: UTXO 最小确认数，用于提现时筛选可用输出
//   - MaxConfirmations: UTXO 最大确认数，通常无需限制
//   - TimeoutMs: RPC 请求超时时间（毫秒），默认 20 秒
//   - AddressType: 地址类型，如 "legacy"（传统 Base58）、"bech32"（原生 SegWit）
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
	// Send 执行提现交易
	//
	// 参数：
	//   - ctx: 上下文，用于取消和超时控制
	//   - fromAddress: 源地址（平台热钱包地址）
	//   - toAddress: 目标地址（用户提现地址）
	//   - totalAmount: 总金额（BTC），包含矿工费
	//   - arrivedAmount: 到账金额（BTC），用户实际收到的金额
	//
	// 返回值：
	//   - string: 交易 ID（txid），用于后续追踪
	//   - error: 交易失败时返回错误
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
	// Allocate 在节点钱包中创建一个新地址
	//
	// 参数：
	//   - ctx: 上下文，用于取消和超时控制
	//   - label: 地址标签，用于在钱包中标识地址用途
	//
	// 返回值：
	//   - string: 新生成的 BTC 地址
	//   - error: 创建失败时返回错误
	Allocate(ctx context.Context, label string) (string, error)
}

// rpcClient 定义提现所需的 JSON-RPC 方法集合。
// 这是一个内部接口，用于解耦 sender 实现与具体的 RPC 客户端。
type rpcClient interface {
	ListUnspent(ctx context.Context, min int, max int, addresses []string) ([]ListUnspentResult, error)
	CreateRawTransaction(ctx context.Context, inputs []Input, outputs []map[string]any) (string, error)
	SignRawTransactionWithWallet(ctx context.Context, hexTx string) (*SignRawTransactionWithWalletResult, error)
	SendRawTransaction(ctx context.Context, signedHex string) (string, error)
}

// addressRPCClient 定义地址分配所需的 JSON-RPC 方法集合。
type addressRPCClient interface {
	GetNewAddress(ctx context.Context, label string, addressType string) (string, error)
}

// sender 是 WithdrawSender 的默认实现，通过 JSON-RPC 与 Bitcoin Core 通信。
type sender struct {
	client           rpcClient
	minConfirmations int
	maxConfirmations int
}

// addressAllocator 是 AddressAllocator 的默认实现。
type addressAllocator struct {
	client      addressRPCClient
	addressType string
}

// jsonRPCClient 封装了与 Bitcoin Core JSON-RPC 端点的 HTTP 通信。
type jsonRPCClient struct {
	url      string
	username string
	password string
	client   *http.Client
}

// --- [JSON-RPC 模型] --- //

// rpcRequest 表示一个标准的 JSON-RPC 请求结构。
type rpcRequest struct {
	ID      string `json:"id"`
	Method  string `json:"method"`
	JSONRPC string `json:"jsonrpc"`
	Params  []any  `json:"params"`
}

// rpcResponse 表示一个标准的 JSON-RPC 响应结构。
// 使用泛型类型 T 以支持不同的返回值类型。
type rpcResponse[T any] struct {
	ID     string    `json:"id"`
	Error  *rpcError `json:"error"`
	Result T         `json:"result"`
}

// rpcError 表示 JSON-RPC 调用返回的错误信息。
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ListUnspentResult 描述 Bitcoin Core 返回的一个可花费 UTXO。
//
// 字段说明：
//   - Txid: 交易 ID，唯一标识这笔交易
//   - Vout: 输出索引，标识交易中的第几个输出
//   - Address: 接收地址
//   - Amount: 金额（BTC）
//   - Confirmations: 确认数，越大越安全
type ListUnspentResult struct {
	Txid          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Address       string  `json:"address"`
	Amount        float64 `json:"amount"`
	Confirmations int     `json:"confirmations"`
}

// Input 匹配 `createrawtransaction` 所需的 JSON-RPC 格式。
// 它标识了一个要花费的 UTXO。
type Input struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

// SignRawTransactionWithWalletResult 是迁移所需的 Bitcoin Core 签名响应的子集。
//
// 字段说明：
//   - Hex: 签名后的原始交易十六进制字符串
//   - Complete: 是否签名完成，false 表示需要更多签名（多签场景）
type SignRawTransactionWithWalletResult struct {
	Hex      string `json:"hex"`
	Complete bool   `json:"complete"`
}

// NewWithdrawSender 构建 jobcenter 使用的默认 BTC 提现发送器。
//
// 参数：
//   - cfg: Bitcoin Core 节点配置
//
// 返回值：
//   - WithdrawSender: 提现发送器实例
//   - error: 配置验证失败时返回错误
//
// 使用示例：
//
//	sender, err := NewWithdrawSender(NodeConfig{
//	    URL:      "http://127.0.0.1:8332",
//	    Username: "bitcoin",
//	    Password: "password",
//	})
//	if err != nil {
//	    // 处理错误
//	}
//	txid, err := sender.Send(ctx, "fromAddr", "toAddr", 0.001, 0.0009)
func NewWithdrawSender(cfg NodeConfig) (WithdrawSender, error) {
	if err := validateNodeConfig(cfg); err != nil {
		return nil, err
	}

	// 设置超时，默认 20 秒
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
//
// 参数：
//   - cfg: Bitcoin Core 节点配置
//
// 返回值：
//   - AddressAllocator: 地址分配器实例
//   - error: 配置验证失败时返回错误
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
//
// 交易流程：
//  1. 验证参数
//  2. 查询源地址的 UTXO
//  3. 选择足够的 UTXO 作为输入
//  4. 构造交易输出（目标地址 + 找零地址）
//  5. 创建原始交易
//  6. 使用钱包签名
//  7. 广播到网络
func (s *sender) Send(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error) {
	// 参数验证
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

	// 查询源地址的可花费 UTXO
	utxos, err := s.client.ListUnspent(ctx, s.minConfirmations, s.maxConfirmations, []string{fromAddress})
	if err != nil {
		return "", err
	}

	// 收集足够的 UTXO 作为交易输入
	inputs, totalInput, err := collectInputs(utxos, totalAmount)
	if err != nil {
		return "", err
	}

	// 构造交易输出
	// 第一个输出是目标地址的转账金额
	outputs := []map[string]any{
		{toAddress: arrivedAmount},
	}

	// 计算找零金额（输入总额 - 总花费）
	// 矿工费 = totalAmount - arrivedAmount，隐含在找零中
	changeAmount := floorFloat(totalInput-totalAmount, 10)
	if changeAmount > 0 {
		outputs = append(outputs, map[string]any{fromAddress: changeAmount})
	}

	// 创建原始交易（未签名）
	rawHex, err := s.client.CreateRawTransaction(ctx, inputs, outputs)
	if err != nil {
		return "", err
	}

	// 使用钱包私钥签名交易
	signed, err := s.client.SignRawTransactionWithWallet(ctx, rawHex)
	if err != nil {
		return "", err
	}
	if signed == nil || !signed.Complete || strings.TrimSpace(signed.Hex) == "" {
		return "", errors.New("bitcoin transaction signature is incomplete")
	}

	// 广播签名后的交易
	return s.client.SendRawTransaction(ctx, signed.Hex)
}

// collectInputs 从 UTXO 列表中选择足够的输入以满足目标金额。
//
// 策略：简单贪心算法，按顺序选择 UTXO 直到满足金额要求。
// 这不是最优的 UTXO 选择策略，但对于当前业务场景足够。
//
// 参数：
//   - utxos: 可用 UTXO 列表
//   - targetAmount: 需要达到的目标金额
//
// 返回值：
//   - []Input: 选中的交易输入
//   - float64: 输入总额
//   - error: 余额不足时返回错误
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

// floorFloat 将浮点数向下取整到指定小数位数。
// 用于计算找零金额时避免精度问题。
//
// 参数：
//   - value: 原始浮点数
//   - precision: 保留的小数位数
//
// 返回值：
//   - float64: 向下取整后的值
func floorFloat(value float64, precision int) float64 {
	if precision < 0 {
		precision = 0
	}
	ratio := math.Pow10(precision)
	return math.Floor(value*ratio) / ratio
}

// chooseConfirmations 选择有效的确认数配置。
// 如果用户配置了正值，使用用户配置；否则使用默认值。
func chooseConfirmations(value int, fallback int) int {
	if value > 0 || fallback == 0 {
		return value
	}
	return fallback
}

// validateNodeConfig 验证 Bitcoin Core 节点配置的完整性。
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

// defaultAddressType 返回默认的地址类型。
// `legacy` 保留旧 MSCoin 代码库使用的传统 Base58 格式 BTC 地址，
// 同时仍将所有权委托给 Bitcoin Core。
func defaultAddressType(raw string) string {
	value := strings.TrimSpace(raw)
	if value != "" {
		return value
	}
	return "legacy"
}

// --- [JSON-RPC 调用] --- //

// ListUnspent 调用 Bitcoin Core 的 listunspent RPC 方法。
func (c *jsonRPCClient) ListUnspent(ctx context.Context, min int, max int, addresses []string) ([]ListUnspentResult, error) {
	var result []ListUnspentResult
	if err := c.call(ctx, "listunspent", []any{min, max, addresses}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateRawTransaction 调用 Bitcoin Core 的 createrawtransaction RPC 方法。
func (c *jsonRPCClient) CreateRawTransaction(ctx context.Context, inputs []Input, outputs []map[string]any) (string, error) {
	var result string
	if err := c.call(ctx, "createrawtransaction", []any{inputs, outputs}, &result); err != nil {
		return "", err
	}
	return result, nil
}

// SignRawTransactionWithWallet 调用 Bitcoin Core 的 signrawtransactionwithwallet RPC 方法。
func (c *jsonRPCClient) SignRawTransactionWithWallet(ctx context.Context, hexTx string) (*SignRawTransactionWithWalletResult, error) {
	var result SignRawTransactionWithWalletResult
	if err := c.call(ctx, "signrawtransactionwithwallet", []any{hexTx}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SendRawTransaction 调用 Bitcoin Core 的 sendrawtransaction RPC 方法。
// 第二个参数 0 表示不广播到其他节点（仅本地提交）。
func (c *jsonRPCClient) SendRawTransaction(ctx context.Context, signedHex string) (string, error) {
	var result string
	if err := c.call(ctx, "sendrawtransaction", []any{signedHex, 0}, &result); err != nil {
		return "", err
	}
	return result, nil
}

// GetNewAddress 调用 Bitcoin Core 的 getnewaddress RPC 方法。
func (c *jsonRPCClient) GetNewAddress(ctx context.Context, label string, addressType string) (string, error) {
	var result string
	if err := c.call(ctx, "getnewaddress", []any{label, addressType}, &result); err != nil {
		return "", err
	}
	return result, nil
}

// Allocate 在节点钱包中创建一个新的 BTC 地址。
func (a *addressAllocator) Allocate(ctx context.Context, label string) (string, error) {
	if a == nil || a.client == nil {
		return "", errors.New("bitcoin address allocator is not initialized")
	}
	return a.client.GetNewAddress(ctx, label, a.addressType)
}

// call 是通用的 JSON-RPC 调用方法。
//
// 参数：
//   - ctx: 上下文
//   - method: RPC 方法名
//   - params: 方法参数
//   - result: 结果接收指针
//
// 返回值：
//   - error: 调用或解析失败时返回错误
func (c *jsonRPCClient) call(ctx context.Context, method string, params []any, result any) error {
	// 构造 JSON-RPC 请求体
	requestBody, err := json.Marshal(rpcRequest{
		ID:      "mscoin_go",
		Method:  method,
		JSONRPC: jsonRPCVersion,
		Params:  params,
	})
	if err != nil {
		return fmt.Errorf("marshal bitcoin rpc request: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("build bitcoin rpc request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 设置 Basic Auth 认证
	req.SetBasicAuth(c.username, c.password)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call bitcoin rpc %s: %w", method, err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read bitcoin rpc %s response: %w", method, err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("bitcoin rpc %s returned status %d", method, resp.StatusCode)
	}

	// 解析 JSON-RPC 响应
	var envelope rpcResponse[json.RawMessage]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode bitcoin rpc %s response: %w", method, err)
	}

	// 检查 JSON-RPC 错误
	if envelope.Error != nil {
		return fmt.Errorf("bitcoin rpc %s failed: %s", method, envelope.Error.Message)
	}

	// 如果不需要结果，直接返回
	if result == nil {
		return nil
	}

	// 解析结果
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("decode bitcoin rpc %s result: %w", method, err)
	}
	return nil
}