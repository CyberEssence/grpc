package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"call-service/internal/middleware"
	"call-service/internal/model"
	"call-service/internal/service"
	"call-service/pkg/authclient"
)

// CallHandler представляет обработчик HTTP запросов для работы с заявками

type CallHandler struct {
	callService service.CallService
	authClient  authclient.AuthClient
}

// NewCallHandler создает новый экземпляр CallHandler

func NewCallHandler(callService service.CallService, authClient authclient.AuthClient) *CallHandler {
	return &CallHandler{callService: callService, authClient: authClient}
}

// CreateCall обрабатывает POST запрос на создание новой заявки

func (h *CallHandler) CreateCall(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.CreateCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	call, err := h.callService.CreateCall(c.Request.Context(), &req, userID)
	if err != nil {
		if err == service.ErrInvalidPhoneNumber {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone number format"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create call"})
		return
	}

	c.JSON(http.StatusCreated, call)
}

// GetCall обрабатывает GET запрос на получение информации о заявке

func (h *CallHandler) GetCall(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid call ID"})
		return
	}

	call, err := h.callService.GetCallByID(c.Request.Context(), id, userID)
	if err != nil {
		if err == service.ErrCallNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "call not found"})
			return
		}
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get call"})
		return
	}

	c.JSON(http.StatusOK, call)
}

// GetAllCalls обрабатывает GET запрос на получение списка всех заявок пользователя

func (h *CallHandler) GetAllCalls(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	calls, err := h.callService.GetAllCalls(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get calls"})
		return
	}

	c.JSON(http.StatusOK, calls)
}

// UpdateCallStatus обрабатывает PATCH запрос на обновление статуса заявки

func (h *CallHandler) UpdateCallStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid call ID"})
		return
	}

	var req model.UpdateCallStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.callService.UpdateCallStatus(c.Request.Context(), id, req.Status, userID)
	if err != nil {
		if err == service.ErrCallNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "call not found"})
			return
		}
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if err == service.ErrInvalidStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update call status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

// DeleteCall обрабатывает DELETE запрос на удаление заявки

func (h *CallHandler) DeleteCall(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid call ID"})
		return
	}

	err = h.callService.DeleteCall(c.Request.Context(), id, userID)
	if err != nil {
		if err == service.ErrCallNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "call not found"})
			return
		}
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete call"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "call deleted successfully"})
}
