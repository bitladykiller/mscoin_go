// Package auth 提供JWT（JSON Web Token）相关的认证功能。
//
// 本包封装了 github.com/golang-jwt/jwt/v4 库，为 MSCoin API 服务提供统一的
// Token 生成和解析能力。采用 HS256 对称加密算法，保持与传统实现兼容。
//
// 主要功能：
//   - GenerateUserToken: 生成用户认证 Token
//   - ParseUserID: 从 Token 中解析用户 ID
//
// 使用场景：
//   - 用户登录成功后生成 Token
//   - API 中间件验证请求中的 Token
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateUserToken 创建 MSCoin API 使用的传统 JWT。
//
// Token 特意保留历史 claim 名称，以便在迁移期间所有 API 中间件和现有客户端保持兼容。
//
// 参数：
//   - secret: JWT 签名密钥，必须保密，建议从环境变量或配置中心获取
//   - issuedAt: Token 签发时间，通常使用 time.Now()
//   - expireSeconds: Token 有效期（秒），建议设置为合理的过期时间（如 3600 秒）
//   - userID: 用户唯一标识，将存储在 Token 的 userId claim 中
//
// 返回值：
//   - string: 生成的 JWT 字符串，格式为 header.payload.signature
//   - error: 签名失败时返回错误
//
// 使用示例：
//
//	token, err := GenerateUserToken("my-secret", time.Now(), 3600, 12345)
//	if err != nil {
//	    // 处理错误
//	}
//	// 将 token 返回给客户端
func GenerateUserToken(secret string, issuedAt time.Time, expireSeconds int64, userID int64) (string, error) {
	// 构建 JWT claims（声明）
	// - exp: 过期时间戳，用于 Token 自动失效
	// - iat: 签发时间戳，用于判断 Token 年龄
	// - userId: 用户 ID，这是 MSCoin 特有的 claim 名称（非标准的 "sub"）
	claims := jwt.MapClaims{
		"exp":    issuedAt.Unix() + expireSeconds,
		"iat":    issuedAt.Unix(),
		"userId": userID,
	}

	// 使用 HS256 算法创建 Token 对象
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	// 使用密钥签名并生成最终字符串
	return token.SignedString([]byte(secret))
}

// ParseUserID 从 JWT token 中提取传统的 `userId` claim。
//
// 为什么需要这个辅助函数：
//   - 旧项目将登录状态存储在 `x-auth-token` 中
//   - 多个 API 服务需要相同的 claim 解析行为
//   - 集中 token 解析可以避免各服务之间出现微妙的偏差
//
// 参数：
//   - tokenString: 客户端提供的 JWT 字符串
//   - secret: 验证签名的密钥，必须与生成时使用的密钥一致
//
// 返回值：
//   - int64: 用户 ID，如果解析成功
//   - error: 以下情况返回错误：
//   - Token 格式无效或签名验证失败
//   - Token 已过期
//   - 缺少必要的 claim（userId 或 exp）
//
// 使用示例：
//
//	userID, err := ParseUserID(tokenFromHeader, "my-secret")
//	if err != nil {
//	    // Token 无效或已过期，返回 401
//	}
//	// 使用 userID 进行业务处理
func ParseUserID(tokenString string, secret string) (int64, error) {
	// 解析并验证 Token
	// keyFunc 回调用于验证签名算法和提供密钥
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 安全检查：确保使用的是预期的 HMAC 签名方法
		// 这防止了算法混淆攻击（如 "none" 算法攻击）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	// 提取 claims 并验证 Token 有效性
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	// 提取 userId claim
	// 注意：JSON 数字解析为 float64，需要类型转换
	userID, ok := claims["userId"].(float64)
	if !ok {
		return 0, errors.New("userId claim missing")
	}

	// 验证 Token 是否已过期
	// 虽然 jwt.Parse 已经验证 exp，但这里显式检查以确保行为一致
	expireAt, ok := claims["exp"].(float64)
	if !ok {
		return 0, errors.New("exp claim missing")
	}
	if int64(expireAt) <= time.Now().Unix() {
		return 0, errors.New("token expired")
	}

	return int64(userID), nil
}
