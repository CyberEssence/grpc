package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"call-service/pkg/authclient"
)

// AuthMiddleware представляет middleware для проверки аутентификации в HTTP запросах

type AuthMiddleware struct {
	authClient authclient.AuthClient
}

// NewAuthMiddleware создает новый экземпляр middleware для аутентификации

func NewAuthMiddleware(authClient authclient.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient}
}

// AuthRequired возвращает обработчик middleware, который проверяет наличие и валидность токена аутентификации

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		token := parts[1]

		valid, userID, err := m.authClient.ValidateToken(c.Request.Context(), token)
		if err != nil || !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		uuidObj, err := uuid.Parse(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			return
		}

		c.Set("userID", uuidObj)
		c.Next()
	}
}

// GetUserID извлекает ID пользователя из контекста запроса

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}

	return userID.(uuid.UUID), true
}
