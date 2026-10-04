package jwttool

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//var jwtSecret = []byte("your-secret-key") // 替换成你的秘钥

// Claims 定义JWT的Payload
type Claims struct {
	UserID   uint64 `json:"id"`
	UserName string `json:"userName"`
	DeptName string `json:"deptName"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT
func GenerateToken(userID uint64, userName string, deptName string, minute int64, secret string) (string, error) {
	expireTime := time.Now().Add(time.Duration(minute) * time.Minute)
	claims := &Claims{
		UserID:   userID,
		UserName: userName,
		DeptName: deptName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ParseToken 解析JWT
func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
