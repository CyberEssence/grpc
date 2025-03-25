package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"call-service/pkg/authclient"
)

// AuthHandler обрабатывает запросы аутентификации через HTTP API.
// Использует клиент для взаимодействия с сервисом аутентификации.
type AuthHandler struct {
	authClient authclient.AuthClient
}

// NewAuthHandler создает новый экземпляр обработчика аутентификации.
// Принимает клиент для взаимодействия с сервисом аутентификации.
func NewAuthHandler(authClient authclient.AuthClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

// RegisterRequest содержит данные для регистрации нового пользователя.
// Поля помечены как обязательные для валидации.
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest содержит данные для входа в систему.
// Поля помечены как обязательные для валидации.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse возвращает данные об успешной аутентификации.
type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// Register обрабатывает запрос на регистрацию нового пользователя.
// Принимает JSON с данными пользователя и возвращает токен и ID при успешной регистрации.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, userID, err := h.authClient.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, AuthResponse{
		Token:  token,
		UserID: userID,
	})
}

// Login обрабатывает запрос на вход в систему.
// Принимает JSON с данными пользователя и возвращает токен и ID при успешной аутентификации.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, userID, err := h.authClient.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, AuthResponse{
		Token:  token,
		UserID: userID,
	})
}
