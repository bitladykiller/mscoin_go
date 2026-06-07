// Package btcx 提供 Bitcoin 钱包相关的密钥管理和地址生成功能。
//
// 本包实现了本地 BTC 钱包创建，包括：
//   - ECDSA P256 密钥对生成
//   - 测试网地址生成（Base58 编码）
//   - 私钥的 PEM + Base58 序列化存储
//
// 安全说明：
//   - 私钥使用 ECDSA P256 曲线，提供足够的安全性
//   - 私钥存储格式为 PEM 编码后再 Base58 编码，便于文本传输
//   - 地址生成遵循比特币测试网规范
//
// 使用场景：
//   - 为新用户生成 BTC 钱包
//   - 重置用户钱包地址
package btcx

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

// 比特币地址生成相关常量
const (
	// testnetVersion 是比特币测试网地址的版本前缀字节。
	// 主网使用 0x00，测试网使用 0x6F。
	// 这里保留测试网是因为历史重置地址流程为 BTC 钱包创建了测试网地址。
	testnetVersion = byte(0x6F)

	// addressChecksumSize 是地址校验和的字节数。
	// 比特币地址使用 SHA256(SHA256(payload)) 的前 4 字节作为校验和。
	addressChecksumSize = 4
)

// Wallet 生成传统 MSCoin 资产服务使用的确定性 BTC 兼容地址材料。
//
// 重构将此辅助工具保留在 `pkg/` 中，以便地址生成保持可复用且独立于传输逻辑。
//
// 字段说明：
//   - privateKey: ECDSA 私钥，用于签名交易
//   - publicKey: 公钥字节，用于生成地址
type Wallet struct {
	privateKey ecdsa.PrivateKey
	publicKey  []byte
}

// NewWallet 创建一个新的 BTC 钱包实例。
//
// 该函数会：
//  1. 使用 ECDSA P256 曲线生成新的密钥对
//  2. 提取公钥字节用于地址生成
//
// 返回值：
//   - *Wallet: 钱包实例，包含私钥和公钥
//   - error: 密钥生成失败时返回错误（通常因系统熵不足）
//
// 使用示例：
//
//	wallet, err := NewWallet()
//	if err != nil {
//	    // 处理错误
//	}
//	address := wallet.TestnetAddress()
func NewWallet() (*Wallet, error) {
	privateKey, publicKey, err := newKeyPair()
	if err != nil {
		return nil, err
	}

	return &Wallet{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// newKeyPair 生成 ECDSA P256 密钥对。
//
// 为什么选择 P256 曲线：
//   - 广泛支持的 NIST 标准曲线
//   - 性能与安全性的良好平衡
//   - 与传统实现兼容
//
// 返回值：
//   - ecdsa.PrivateKey: 私钥，包含公钥信息
//   - []byte: 公钥字节（X || Y 拼接）
//   - error: 生成失败时返回错误
func newKeyPair() (ecdsa.PrivateKey, []byte, error) {
	// 使用 P256 曲线（也称为 secp256r1 或 prime256v1）
	curve := elliptic.P256()

	// 从系统随机源生成私钥
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return ecdsa.PrivateKey{}, nil, fmt.Errorf("generate ecdsa key pair: %w", err)
	}

	// 提取公钥：X 和 Y 坐标拼接
	// 这是一种非压缩公钥格式（不含 0x04 前缀）
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey, nil
}

// TestnetAddress 返回与传统实现兼容的 Base58 编码 BTC 测试网地址。
//
// 地址生成算法：
//  1. SHA256(公钥) -> RIPEMD160 得到公钥哈希
//  2. 添加版本前缀（0x6F 为测试网）
//  3. 计算双重 SHA256 校验和
//  4. Base58 编码
//
// 这里保留测试网是因为历史重置地址流程为 BTC 钱包创建了测试网地址。
//
// 返回值：
//   - string: Base58 编码的测试网地址
func (w *Wallet) TestnetAddress() string {
	// 步骤 1: SHA256 -> RIPEMD160 公钥哈希
	publicHash := ripemd160Hash(w.publicKey)

	// 步骤 2: 添加版本前缀
	versioned := append([]byte{testnetVersion}, publicHash...)

	// 步骤 3: 计算校验和
	checksum := checksum(versioned)

	// 步骤 4: 拼接有效载荷和校验和
	payload := append(versioned, checksum...)

	// 步骤 5: Base58 编码
	return string(encodeBase58(payload))
}

// EncodedPrivateKey 将私钥序列化为 PEM，然后 Base58 编码 PEM 字节，
// 镜像 MSCoin 使用的历史存储格式。
//
// 序列化步骤：
//  1. 使用 x509 将 ECDSA 私钥编码为 DER 格式
//  2. 将 DER 字节包装为 PEM 格式
//  3. 对 PEM 字节进行 Base58 编码
//
// 为什么使用这种格式：
//   - PEM 格式是标准的私钥存储格式
//   - Base58 编码便于在文本环境中存储和传输
//   - 与传统系统保持兼容
//
// 返回值：
//   - string: Base58 编码的 PEM 格式私钥
//   - error: 序列化失败时返回错误
func (w *Wallet) EncodedPrivateKey() (string, error) {
	// 将 ECDSA 私钥编码为 DER 格式
	keyBytes, err := x509.MarshalECPrivateKey(&w.privateKey)
	if err != nil {
		return "", fmt.Errorf("marshal private key: %w", err)
	}

	// 创建 PEM 块
	block := pem.Block{
		Type:  "ECD PRIVATE KEY",
		Bytes: keyBytes,
	}

	// 将 PEM 块写入缓冲区
	buffer := bytes.NewBuffer(nil)
	writer := bufio.NewWriter(buffer)
	if err := pem.Encode(writer, &block); err != nil {
		return "", fmt.Errorf("encode pem private key: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("flush private key buffer: %w", err)
	}

	// Base58 编码 PEM 字节
	return string(encodeBase58(buffer.Bytes())), nil
}

// ripemd160Hash 计算公钥的 RIPEMD160 哈希。
//
// 这是比特币地址生成的标准算法：
//   - 先对公钥进行 SHA256 哈希
//   - 再对 SHA256 结果进行 RIPEMD160 哈希
//
// 参数：
//   - publicKey: 原始公钥字节
//
// 返回值：
//   - []byte: 20 字节的公钥哈希
func ripemd160Hash(publicKey []byte) []byte {
	// 第一阶段：SHA256
	stageOne := sha256.Sum256(publicKey)

	// 第二阶段：RIPEMD160
	stageTwo := ripemd160.New()
	_, _ = stageTwo.Write(stageOne[:])
	return stageTwo.Sum(nil)
}

// checksum 计算比特币地址的校验和。
//
// 算法：SHA256(SHA256(payload)) 的前 4 字节
// 这是比特币地址验证的一部分，用于防止地址输入错误。
//
// 参数：
//   - payload: 需要计算校验和的数据（版本前缀 + 公钥哈希）
//
// 返回值：
//   - []byte: 4 字节校验和
func checksum(payload []byte) []byte {
	// 双重 SHA256
	stageOne := sha256.Sum256(payload)
	stageTwo := sha256.Sum256(stageOne[:])

	// 取前 4 字节作为校验和
	return stageTwo[:addressChecksumSize]
}