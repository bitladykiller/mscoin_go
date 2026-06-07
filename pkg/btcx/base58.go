package btcx

import (
	"bytes"
	"math/big"
)

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func reverseBytes(data []byte) {
	for left, right := 0, len(data)-1; left < right; left, right = left+1, right-1 {
		data[left], data[right] = data[right], data[left]
	}
}

func encodeBase58(input []byte) []byte {
	var result []byte
	value := big.NewInt(0).SetBytes(input)
	base := big.NewInt(int64(len(base58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for value.Cmp(zero) != 0 {
		value.DivMod(value, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}
	reverseBytes(result)

	for _, item := range input {
		if item == 0x00 {
			result = append([]byte{base58Alphabet[0]}, result...)
			continue
		}
		break
	}
	return result
}

func decodeBase58(input []byte) []byte {
	result := big.NewInt(0)
	payload := input

	for _, item := range payload {
		index := bytes.IndexByte(base58Alphabet, item)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(index)))
	}
	return result.Bytes()
}
