package webJwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/utils"
	"time"
)

// CustomClaims 自定义的 JWT Claims，可以添加需要保存的信息
type CustomClaims struct {
	UserName    string `json:"user_name"`    // 用户名
	UserEmail   string `json:"user_email"`   // 用户邮箱
	RandomToken string `json:"random_token"` // 随机字符串
	jwt.RegisteredClaims
}

// GenerateJWTToken 生成一个带有自定义声明的 JWT Token。
// 返回值：
//   - string: 生成的 JWT Token 字符串。
//   - error: 如果在生成随机字符串或签名 Token 时发生错误，则返回相应的错误信息。
func GenerateJWTToken() (string, error) {
	// 生成随机字符串，用于增强 Token 的安全性。
	randomStr, err := utils.HashWithLength(config.User.UserEmail, 128)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %v", err)
	}

	// 设置自定义声明（Claims），包括用户信息、随机字符串以及标准的 JWT 注册声明。
	claims := CustomClaims{
		UserName:    config.User.Username,
		UserEmail:   config.User.UserEmail,
		RandomToken: randomStr,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Server.TokenExpireDuration) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    config.User.Username,
		},
	}

	// 使用 HMAC SHA-256 签名方法创建 JWT Token，并附加自定义声明。
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 对 Token 进行签名，生成最终的 JWT Token 字符串。
	tokenString, err := token.SignedString([]byte(config.Server.TokenKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	// 返回生成的 JWT Token 字符串。
	return tokenString, nil
}

// ParseJWTToken 解析并验证JWT令牌，返回自定义声明（CustomClaims）和可能的错误。
// 参数:
//   - tokenString: 待解析的JWT令牌字符串。
//   - secretKey: 用于验证JWT签名的密钥。
//
// 返回值:
//   - *CustomClaims: 如果解析成功且令牌有效，则返回包含自定义声明的结构体指针。
//   - error: 如果解析失败、签名验证失败或令牌无效，则返回错误信息。
func ParseJWTToken(tokenString string, secretKey string) (*CustomClaims, error) {
	// 使用jwt.ParseWithClaims解析令牌，并提供一个回调函数用于验证签名方法。
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法是否为HMAC算法，如果不是则返回错误。
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// 返回密钥以供签名验证使用。
		return []byte(secretKey), nil
	})
	if err != nil {
		// 如果解析过程中发生错误，直接返回错误信息。
		return nil, err
	}

	// 检查解析后的声明是否为CustomClaims类型，并验证令牌是否有效。
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	// 如果声明类型不匹配或令牌无效，返回错误。
	return nil, errors.New("invalid token")
}
