// Package btcx 提供 Bitcoin 相关的编码工具和钱包功能。
//
// 本包包含 BTC 地址生成所需的 Base58 编解码算法实现，以及本地钱包创建功能。
// 这些工具用于生成与传统 MSCoin 资产服务兼容的 BTC 地址。
//
// 主要功能：
//   - Base58 编码/解码：用于 BTC 地址的字符串表示
//   - 私钥/公钥对生成：基于 ECDSA P256 曲线
//   - 测试网地址生成：生成 Base58 编码的测试网地址
//
// 注意事项：
//   - 本包的 Base58 实现是专门为 BTC 地址设计的，包含前导零的处理
//   - 测试网地址使用 0x6F 版本前缀，主网地址使用 0x00
package btcx

import (
	"bytes"
	"math/big"
)

// base58Alphabet 是 Base58 编码使用的字符集。
// 注意：Base58 去除了容易混淆的字符：0（零）、O（大写O）、I（大写i）、l（小写L）
// 这种设计减少了人工输入时的错误率，特别适合用于加密货币地址。
var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// reverseBytes 原地反转字节切片。
// 这是一个辅助函数，用于在 Base58 编码后修正字节顺序。
// 因为编码过程中产生的结果是反向的，需要翻转得到正确的顺序。
func reverseBytes(data []byte) {
	for left, right := 0, len(data)-1; left < right; left, right = left+1, right-1 {
		data[left], data[right] = data[right], data[left]
	}
}

// encodeBase58 将字节数组编码为 Base58 格式。
//
// 编码算法说明：
//  1. 将输入字节视为一个大整数
//  2. 不断除以 58，取余数作为字符索引
//  3. 处理前导零：每个前导零字节转换为 Base58 字母表的第一个字符 '1'
//
// 这种编码方式确保：
//   - 相同的输入总是产生相同的输出
//   - 前导零被正确保留（对于 BTC 地址很重要）
//
// 参数：
//   - input: 待编码的字节数组
//
// 返回值：
//   - []byte: Base58 编码后的字节数组
func encodeBase58(input []byte) []byte {
	var result []byte

	// 将字节数组转换为大整数，用于后续的除法运算
	value := big.NewInt(0).SetBytes(input)

	// 基数 58
	base := big.NewInt(int64(len(base58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	// 核心编码循环：不断除以 58，取余数
	// 每次余数对应一个 Base58 字符
	for value.Cmp(zero) != 0 {
		value.DivMod(value, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	// 反转结果，因为我们是从低位到高位计算的
	reverseBytes(result)

	// 处理前导零字节
	// 在 BTC 地址中，前导零字节编码为 '1'（Base58 字母表的第一个字符）
	// 这确保了地址的长度和前缀正确
	for _, item := range input {
		if item == 0x00 {
			result = append([]byte{base58Alphabet[0]}, result...)
			continue
		}
		break
	}

	return result
}

// decodeBase58 将 Base58 字符串解码为原始字节数组。
//
// 解码算法说明：
//  1. 遍历每个字符，找到其在字母表中的索引
//  2. 结果 = 结果 * 58 + 索引值
//
// 注意：此实现不处理前导 '1' 字符（对应前导零字节），
// 因为当前使用场景不需要完整解码功能。
//
// 参数：
//   - input: Base58 编码的字节数组
//
// 返回值：
//   - []byte: 解码后的原始字节数组
func decodeBase58(input []byte) []byte {
	result := big.NewInt(0)
	payload := input

	// 核心解码循环：将 Base58 字符转换回原始数值
	// 每次处理一个字符：结果 = 结果 * 58 + 字符索引
	for _, item := range payload {
		index := bytes.IndexByte(base58Alphabet, item)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(index)))
	}

	return result.Bytes()
}
