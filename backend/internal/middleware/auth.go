package middleware

import (
	"net/http"
	"strings"

	"github.com/atakhanov/sensory-navigator/backend/internal/auth"
	"github.com/atakhanov/sensory-navigator/backend/internal/config"
	"github.com/gin-gonic/gin"
)

const (
	UserContextKey = "current_user_id"
)

// RequireAuth требует наличия JWT-токена в заголовке Authorization.
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "требуется авторизация"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "некорректный формат токена"})
			return
		}
		claims, err := auth.ParseToken(parts[1], cfg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set(UserContextKey, claims.UserID)
		c.Next()
	}
}

// OptionalAuth подкладывает id пользователя при наличии валидного токена,
// но не прерывает запрос при его отсутствии.
func OptionalAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			if claims, err := auth.ParseToken(parts[1], cfg); err == nil {
				c.Set(UserContextKey, claims.UserID)
			}
		}
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(UserContextKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}