package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role"` // 用户角色，例如 "admin"、"user"
	jwt.RegisteredClaims
}

var secretKey = []byte("bXlTZXJldEtleU5hY2tTZXh2cXlNT1BQRXZ6YmVhdHppYm9m5vbT28yZXB2U3Y0==")

// 生成jwt
func GenerateToken(userID, userRole string) (string, error) {
	// 设置 token 的过期时间
	expirationTime := time.Now().Add(24 * time.Hour) // 24 小时有效期

	// 创建 claims
	claims := &CustomClaims{
		UserID:   userID,
		UserRole: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "yunshu_blog", // 签发人
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	fmt.Println("claims:", claims, "结束")

	// 使用 HMAC SHA256 签名算法生成 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并返回 JWT 字符串
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// 解析和验证 JWT token
func ParseToken(tokenString string) (*CustomClaims, error) {
	fmt.Println("tokenString:", tokenString)
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}


	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization header required"})
			c.Abort()
			return
		}

		// 去掉 "Bearer " 前缀
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// 解析和验证 token
		claims, err := ParseToken(tokenString)
		if err != nil {
			fmt.Println("解析和验证 token:", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 将解析出的用户信息存储到上下文中，后续的处理函数可以通过 c.Get("user") 获取
		c.Set("user", claims)

		// 继续执行
		c.Next()
	}
}

// 权限检查中间件（示例：只有管理员可以访问某些 API）
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
			c.Abort()
			return
		}

		claims, ok := user.(*CustomClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid claims"})
			c.Abort()
			return
		}

		// 判断用户角色是否为 admin
		if claims.UserRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"message": "Access forbidden: Admins only"})
			c.Abort()
			return
		}

		// 继续执行
		c.Next()
	}
}
