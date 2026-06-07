package passwordx

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"
)

const (
	defaultIterations = 10000
	defaultKeyLen     = 128
	defaultSaltLen    = 64
)

const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Encode 使用与传统 MSCoin 服务相同的 PBKDF2 设置对原始密码进行哈希，
// 并返回生成的盐和编码后的哈希值。
func Encode(rawPwd string) (string, string) {
	saltBytes := make([]byte, defaultSaltLen)
	if _, err := rand.Read(saltBytes); err != nil {
		panic(err)
	}
	for index, value := range saltBytes {
		saltBytes[index] = alphanum[value%byte(len(alphanum))]
	}

	salt := string(saltBytes)
	hash := pbkdf2.Key([]byte(rawPwd), []byte(salt), defaultIterations, defaultKeyLen, sha512.New)
	return salt, hex.EncodeToString(hash)
}

func Verify(rawPwd string, salt string, encodedPwd string) bool {
	hash := pbkdf2.Key([]byte(rawPwd), []byte(salt), defaultIterations, defaultKeyLen, sha512.New)
	return encodedPwd == hex.EncodeToString(hash)
}
