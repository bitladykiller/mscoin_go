// Package passwordx 提供密码哈希和验证功能。
//
// 本包使用 PBKDF2 算法进行密码哈希，这是 OWASP 推荐的密码存储方案之一。
// PBKDF2 通过多次迭代增加计算成本，有效抵御暴力破解攻击。
//
// 算法参数说明：
//   - 迭代次数：10000 次，提供足够的计算成本
//   - 密钥长度：128 字节，产生足够长的哈希值
//   - 盐长度：64 字节，使用字母数字字符
//   - 哈希函数：SHA512，提供强密码学安全性
//
// 安全说明：
//   - 每次加密生成随机盐，防止彩虹表攻击
//   - 盐使用字母数字字符，便于存储和传输
//   - 相同密码每次加密结果不同
//
// 使用场景：
//   - 用户注册时加密密码
//   - 用户登录时验证密码
//   - 密码重置时重新加密
package passwordx

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"
)

// PBKDF2 算法参数常量
const (
	// defaultIterations 是 PBKDF2 的迭代次数。
	// 10000 次迭代在现代硬件上需要约 100ms，
	// 对于用户登录场景可接受，同时对攻击者足够昂贵。
	defaultIterations = 10000

	// defaultKeyLen 是生成的哈希密钥长度（字节）。
	// 128 字节（1024 位）提供足够的输出长度。
	defaultKeyLen = 128

	// defaultSaltLen 是随机盐的长度（字节）。
	// 64 字节（512 位）的盐足以防止彩虹表攻击。
	defaultSaltLen = 64
)

// alphanum 是用于生成盐的字母数字字符集。
// 包含大小写字母和数字，便于在文本格式中存储和传输。
const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Encode 使用与传统 MSCoin 服务相同的 PBKDF2 设置对原始密码进行哈希，
// 并返回生成的盐和编码后的哈希值。
//
// 编码流程：
//  1. 生成随机盐（字母数字字符）
//  2. 使用 PBKDF2-SHA512 算法计算哈希
//  3. 将哈希值转换为十六进制字符串
//
// 参数：
//   - rawPwd: 用户输入的原始密码
//
// 返回值：
//   - string: 随机盐（需要存储到数据库）
//   - string: 十六进制编码的哈希值（需要存储到数据库）
//
// 使用示例：
//
//	salt, hash := passwordx.Encode("user_password")
//	// 存储到数据库：salt 和 hash
//	db.SaveUserSaltAndHash(userID, salt, hash)
//
// 安全说明：
//   - 盐和哈希都需要存储到数据库
//   - 验证时需要使用相同的盐
//   - 系统熵不足时会 panic（极少发生）
func Encode(rawPwd string) (string, string) {
	// 生成随机盐字节
	saltBytes := make([]byte, defaultSaltLen)
	if _, err := rand.Read(saltBytes); err != nil {
		// 系统熵不足，这是一个严重错误
		// 在正常环境下不应发生
		panic(err)
	}

	// 将随机字节转换为字母数字字符
	// 使用模运算确保字符在 alphanum 范围内
	for index, value := range saltBytes {
		saltBytes[index] = alphanum[value%byte(len(alphanum))]
	}

	// 生成盐字符串
	salt := string(saltBytes)

	// 使用 PBKDF2 算法计算哈希
	// 参数：密码、盐、迭代次数、密钥长度、哈希函数
	hash := pbkdf2.Key([]byte(rawPwd), []byte(salt), defaultIterations, defaultKeyLen, sha512.New)

	// 返回盐和十六进制编码的哈希
	return salt, hex.EncodeToString(hash)
}

// Verify 验证用户输入的密码是否与存储的哈希匹配。
//
// 验证流程：
//  1. 使用相同的盐和参数计算输入密码的哈希
//  2. 将计算的哈希与存储的哈希进行比较
//
// 参数：
//   - rawPwd: 用户输入的原始密码
//   - salt: 存储的盐值
//   - encodedPwd: 存储的十六进制编码哈希值
//
// 返回值：
//   - bool: true 表示密码匹配，false 表示不匹配
//
// 使用示例：
//
//	salt, hash := db.GetUserSaltAndHash(userID)
//	if passwordx.Verify(inputPassword, salt, hash) {
//	    // 密码正确，允许登录
//	} else {
//	    // 密码错误，拒绝登录
//	}
//
// 安全说明：
//   - 即使密码错误也不要透露具体原因
//   - 考虑添加登录失败次数限制防止暴力破解
func Verify(rawPwd string, salt string, encodedPwd string) bool {
	// 使用相同的参数计算输入密码的哈希
	hash := pbkdf2.Key([]byte(rawPwd), []byte(salt), defaultIterations, defaultKeyLen, sha512.New)

	// 比较十六进制编码的哈希值
	return encodedPwd == hex.EncodeToString(hash)
}