package jwt

import (
	"fmt"
	"time"

	"bluebell/settings"
	"github.com/golang-jwt/jwt/v5"
)

// Myclaims 自定义的 JWT 声明结构体
// - 将来可以根据业务需要在这里扩展字段，例如角色、权限等
// - 嵌入 jwt.RegisteredClaims 用于支持标准声明（exp、iat、nbf、iss、aud、sub 等）
type Myclaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// loadConfig 从全局配置中读取 JWT 密钥与过期时间
// 注意：生产环境请通过配置文件或环境变量注入密钥，切勿把密钥硬编码在源码中。
func loadConfig() ([]byte, time.Duration, error) {
	if settings.GlobalConfig == nil {
		return nil, 0, fmt.Errorf("jwt config is not initialized")
	}
	secret := settings.GlobalConfig.JWT.Secret
	if secret == "" {
		return nil, 0, fmt.Errorf("jwt secret is empty")
	}

	expireSeconds := settings.GlobalConfig.JWT.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = int64((24 * time.Hour).Seconds())
	}

	return []byte(secret), time.Duration(expireSeconds) * time.Second, nil
}

// GenToken 根据传入的自定义声明生成签名后的 JWT 字符串
// 使用 HS256(HMAC + SHA256) 算法签名。
// 生成时会自动为 claims 注册标准的过期时间（ExpiresAt）和签发时间（IssuedAt），
// 如果调用方已设置这些字段则不会覆盖。
func GenToken(claims *Myclaims) (string, error) {
	secretKey, expireDuration, err := loadConfig()
	if err != nil {
		return "", err
	}

	now := time.Now()
	// 如果调用方没有设置 IssuedAt / ExpiresAt，则使用默认值
	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(now)
	}
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(expireDuration))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// SignedString 内部会根据 SigningMethod 使用 secretKey 进行签名
	return token.SignedString(secretKey)
}

// ParseToken 解析并验证一个 JWT 字符串，返回解析出的自定义声明
// - 如果 token 合法且未过期，返回 *Myclaims
// - 否则返回对应的错误（例如过期、签名不匹配等）
func ParseToken(tokenString string) (*Myclaims, error) {
	secretKey, _, err := loadConfig()
	if err != nil {
		return nil, err
	}

	claims := &Myclaims{}
	// ParseWithClaims 会解析 token 并把 claim 解入 claims 中，同时验证签名（key func）
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// 这里只支持 HMAC 签名方法（HS256）；可以根据需要扩展其他方法的支持
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	// 额外可以在此处执行自定义校验（例如 iss、aud）
	return claims, nil
}
