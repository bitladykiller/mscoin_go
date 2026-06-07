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

// Encode hashes a raw password using the same PBKDF2 settings as the legacy
// MSCoin services and returns both the generated salt and the encoded hash.
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
