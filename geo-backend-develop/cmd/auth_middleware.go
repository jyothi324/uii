package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type authClaims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

func initJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "Tulip@123"
	}
	jwtSecret = []byte(secret)
}

func authDisabled() bool {
	return strings.EqualFold(os.Getenv("AUTH_DISABLED"), "true")
}

func parseToken(tokenStr string) (*authClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*authClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func tokenFromRequest(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), true
	}
	cookie, err := r.Cookie("token")
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func authenticateRequest(r *http.Request) (*authClaims, error) {
	tokenStr, ok := tokenFromRequest(r)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}
	return parseToken(tokenStr)
}

func setAuthContext(c *gin.Context, claims *authClaims) {
	c.Set("user_id", claims.UserID)
	c.Set("user_name", claims.Name)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		if authDisabled() {
			c.Next()
			return
		}

		claims, err := authenticateRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		setAuthContext(c, claims)
		c.Next()
	}
}

func getRequestUserID(c *gin.Context) string {
	if authDisabled() {
		return c.Query("agentId")
	}
	if val, ok := c.Get("user_id"); ok {
		if userID, ok := val.(string); ok {
			return userID
		}
	}
	return ""
}
