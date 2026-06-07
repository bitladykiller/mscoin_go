// Package auth contains authentication helpers shared across API services.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateUserToken creates the legacy JWT used by MSCoin APIs.
//
// The token intentionally keeps the historical claim names so all API
// middleware and existing clients remain compatible during the migration.
func GenerateUserToken(secret string, issuedAt time.Time, expireSeconds int64, userID int64) (string, error) {
	claims := jwt.MapClaims{
		"exp":    issuedAt.Unix() + expireSeconds,
		"iat":    issuedAt.Unix(),
		"userId": userID,
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secret))
}

// ParseUserID extracts the legacy `userId` claim from a JWT token.
//
// Why this helper exists:
// - the old project stores login state in `x-auth-token`
// - multiple API services need the same claim-parsing behavior
// - centralizing token parsing avoids subtle drift across services
func ParseUserID(tokenString string, secret string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	userID, ok := claims["userId"].(float64)
	if !ok {
		return 0, errors.New("userId claim missing")
	}

	expireAt, ok := claims["exp"].(float64)
	if !ok {
		return 0, errors.New("exp claim missing")
	}
	if int64(expireAt) <= time.Now().Unix() {
		return 0, errors.New("token expired")
	}

	return int64(userID), nil
}
