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

// Wallet generates deterministic BTC-compatible address material used by the
// legacy MSCoin asset service. The refactor keeps this helper in `pkg/` so
// address generation remains reusable and independent from transport logic.
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

// TestnetAddress returns a Base58-encoded BTC testnet address compatible with
// the legacy implementation. Testnet is preserved here because the historical
// reset-address flow created testnet addresses for BTC wallets.
func (w *Wallet) TestnetAddress() string {
	publicHash := ripemd160Hash(w.publicKey)
	versioned := append([]byte{testnetVersion}, publicHash...)
	checksum := checksum(versioned)
	payload := append(versioned, checksum...)
	return string(encodeBase58(payload))
}

// EncodedPrivateKey serializes the private key into PEM and then Base58-encodes
// the PEM bytes, mirroring the historical storage format used by MSCoin.
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
