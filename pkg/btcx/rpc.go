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

// NodeConfig describes one Bitcoin Core JSON-RPC endpoint.
type NodeConfig struct {
	URL              string
	Username         string
	Password         string
	MinConfirmations int
	MaxConfirmations int
	TimeoutMs        int
	AddressType      string
}

// WithdrawSender abstracts one chain-specific withdraw broadcaster.
//
// Why jobcenter depends on this abstraction instead of on raw JSON-RPC calls:
//   - the domain service should not know transport details such as JSON-RPC
//     request shapes
//   - unit tests can replace the sender with deterministic fakes
//   - future ETH/TRON/etc. senders can follow the same orchestration contract
type WithdrawSender interface {
	Send(ctx context.Context, fromAddress string, toAddress string, totalAmount float64, arrivedAmount float64) (string, error)
}

// AddressAllocator abstracts wallet-managed address allocation.
//
// Why a dedicated abstraction is required:
//   - `ucenter` must create BTC addresses that belong to the same Bitcoin Core
//     wallet later used by `jobcenter` to sign withdraw transactions
//   - generating keypairs locally would split address ownership from signing
//     responsibility and make `signrawtransactionwithwallet` fail at runtime
//   - tests can replace the allocator with a deterministic fake without
//     coupling business logic to JSON-RPC details
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

// --- [JSON-RPC Models] --- //

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

// ListUnspentResult describes one spendable UTXO returned by Bitcoin Core.
type ListUnspentResult struct {
	Txid          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Address       string  `json:"address"`
	Amount        float64 `json:"amount"`
	Confirmations int     `json:"confirmations"`
}

// Input matches the JSON-RPC shape required by `createrawtransaction`.
type Input struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

// SignRawTransactionWithWalletResult is the subset the migration needs from
// Bitcoin Core's signing response.
type SignRawTransactionWithWalletResult struct {
	Hex      string `json:"hex"`
	Complete bool   `json:"complete"`
}

// NewWithdrawSender builds the default BTC withdraw sender used by jobcenter.
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

// NewAddressAllocator builds a Bitcoin Core-backed address allocator.
//
// The current migration uses this in `ucenter` so newly reset BTC addresses are
// created by the node wallet instead of being generated locally.
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

// --- [Withdraw Flow] --- //

// Send selects enough UTXOs, constructs the BTC transaction, signs it with the
// node wallet, and broadcasts it.
//
// The sender keeps the same value semantics as the legacy implementation:
// `totalAmount` includes the miner fee already reserved at apply time, while
// `arrivedAmount` is the amount transferred to the external destination.
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

	// `legacy` preserves the historical Base58-style BTC address shape used by
	// the old MSCoin codebase while still delegating ownership to Bitcoin Core.
	return "legacy"
}

// --- [JSON-RPC Calls] --- //

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
