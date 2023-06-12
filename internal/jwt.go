package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"kikitoru/config"
	"net/http"
	"strings"
	"time"
)

var jwtSecret = []byte(config.C.JWTSecret)

type Claims struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Group    string `json:"group"`
	jwt.RegisteredClaims
}

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check token first
		token := c.Query("token")
		if token == "" {
			token = c.Query("auth_token")
			if token == "" {
				token = c.Request.Header.Get("authorization")
			}
		}
		//logs.Debug(token)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No authorization token was found",
			})
			c.Abort()
			return
		}
		// 按空格分割
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "jwt malformed",
			})
			c.Abort()
			return
		}
		mc, err := ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "Invalid Token",
			})
			c.Abort()
			return
		}
		// 将当前请求的username信息保存到请求的上下文c上
		c.Set("username", mc.Username)
		c.Set("usergroup", mc.Group)
		c.Next() // 后续的处理函数可以用过c.Get("username")来获取当前请求的用户信息
	}
}

func GenerateToken(username, group string) (string, error) {
	claims := Claims{
		Username: username,
		Group:    group,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.C.ExpiresIn)), // ExpiresIn in config
			Issuer:    "kikitoru",
			Subject:   "login",
			Audience:  []string{username},
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func ParseToken(token string) (*Claims, error) {
	t := strings.TrimPrefix(token, "Bearer ")
	tokenClaims, err := jwt.ParseWithClaims(t, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
