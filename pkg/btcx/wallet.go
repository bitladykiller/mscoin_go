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

const (
	testnetVersion      = byte(0x6F)
	addressChecksumSize = 4
)

// Wallet 生成传统 MSCoin 资产服务使用的确定性 BTC 兼容地址材料。
// 重构将此辅助工具保留在 `pkg/` 中，以便地址生成保持可复用且独立于传输逻辑。
type Wallet struct {
	privateKey ecdsa.PrivateKey
	publicKey  []byte
}

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

func newKeyPair() (ecdsa.PrivateKey, []byte, error) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return ecdsa.PrivateKey{}, nil, fmt.Errorf("generate ecdsa key pair: %w", err)
	}

	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey, nil
}

// TestnetAddress 返回与传统实现兼容的 Base58 编码 BTC 测试网地址。
// 这里保留测试网是因为历史重置地址流程为 BTC 钱包创建了测试网地址。
func (w *Wallet) TestnetAddress() string {
	publicHash := ripemd160Hash(w.publicKey)
	versioned := append([]byte{testnetVersion}, publicHash...)
	checksum := checksum(versioned)
	payload := append(versioned, checksum...)
	return string(encodeBase58(payload))
}

// EncodedPrivateKey 将私钥序列化为 PEM，然后 Base58 编码 PEM 字节，
// 镜像 MSCoin 使用的历史存储格式。
func (w *Wallet) EncodedPrivateKey() (string, error) {
	keyBytes, err := x509.MarshalECPrivateKey(&w.privateKey)
	if err != nil {
		return "", fmt.Errorf("marshal private key: %w", err)
	}

	block := pem.Block{
		Type:  "ECD PRIVATE KEY",
		Bytes: keyBytes,
	}

	buffer := bytes.NewBuffer(nil)
	writer := bufio.NewWriter(buffer)
	if err := pem.Encode(writer, &block); err != nil {
		return "", fmt.Errorf("encode pem private key: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("flush private key buffer: %w", err)
	}

	return string(encodeBase58(buffer.Bytes())), nil
}

func ripemd160Hash(publicKey []byte) []byte {
	stageOne := sha256.Sum256(publicKey)
	stageTwo := ripemd160.New()
	_, _ = stageTwo.Write(stageOne[:])
	return stageTwo.Sum(nil)
}

func checksum(payload []byte) []byte {
	stageOne := sha256.Sum256(payload)
	stageTwo := sha256.Sum256(stageOne[:])
	return stageTwo[:addressChecksumSize]
}
