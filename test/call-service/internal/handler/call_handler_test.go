package handler

import (
	"bytes"
	"call-service/internal/middleware"
	"call-service/internal/model"
	"call-service/internal/service"
	"call-service/pkg/authclient"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthClient реализует интерфейс AuthClient для тестирования.
// Использует библиотеку testify/mock для создания мок-объекта.

type MockAuthClient struct {
	mock.Mock
}

// Register имитирует регистрацию пользователя.
// Возвращает токен, ID пользователя и ошибку.

func (m *MockAuthClient) Register(ctx context.Context, username, password string) (string, string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.String(1), args.Error(2)
}

// Login имитирует вход пользователя в систему.
// Возвращает токен, ID пользователя и ошибку.

func (m *MockAuthClient) Login(ctx context.Context, username, password string) (string, string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.String(1), args.Error(2)
}

// ValidateToken имитирует проверку валидности токена.
// Возвращает флаг валидности, ID пользователя и ошибку.

func (m *MockAuthClient) ValidateToken(ctx context.Context, token string) (bool, string, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.String(1), args.Error(2)
}

// Close имитирует закрытие соединения.
// Возвращает ошибку при неудачном закрытии.

func (m *MockAuthClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockCallService реализует интерфейс CallService для тестирования.
// Использует библиотеку testify/mock для создания мок-объекта.

type MockCallService struct {
	mock.Mock
}

// CreateCall имитирует создание новой заявки.
// Возвращает созданную заявку и ошибку.

func (m *MockCallService) CreateCall(ctx context.Context, req *model.CreateCallRequest, userID uuid.UUID) (*model.Call, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Call), args.Error(1)
}

// GetCallByID имитирует получение заявки по ID.
// Возвращает заявку и ошибку.

func (m *MockCallService) GetCallByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Call, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Call), args.Error(1)
}

// GetAllCalls имитирует получение всех заявок пользователя.
// Возвращает список заявок и ошибку.

func (m *MockCallService) GetAllCalls(ctx context.Context, userID uuid.UUID) ([]*model.Call, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Call), args.Error(1)
}

// UpdateCallStatus имитирует обновление статуса заявки.
// Возвращает ошибку при неудачном обновлении.

func (m *MockCallService) UpdateCallStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID) error {
	args := m.Called(ctx, id, status, userID)
	return args.Error(0)
}

// DeleteCall имитирует удаление заявки.
// Возвращает ошибку при неудачном удалении.

func (m *MockCallService) DeleteCall(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

// printRequestResponse выводит детали тестового запроса и ответа для отладки.
// Показывает метод, URL, заголовки и тело запроса, а также статус и тело ответа.

func printRequestResponse(t *testing.T, req *http.Request, w *httptest.ResponseRecorder) {
	t.Logf("\n=== Тестовый запрос ===")
	t.Logf("Метод: %s", req.Method)
	t.Logf("URL: %s", req.URL)
	t.Logf("Заголовки:")
	for k, v := range req.Header {
		t.Logf("  %s: %s", k, v)
	}

	if req.Body != nil {
		// Если это не GET/HEAD запрос, то пытаемся прочитать тело
		if req.Method != "GET" && req.Method != "HEAD" {
			// Копируем оригинальное тело, чтобы не потерять его
			bodyBytes, err := io.ReadAll(req.Body)
			if err == nil {
				t.Logf("Тело запроса: %s", string(bodyBytes))
				// Восстанавливаем тело запроса для последующей обработки
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else {
				t.Logf("Ошибка чтения тела запроса: %v", err)
			}
		}
	}

	t.Logf("\n=== Ответ ===")
	t.Logf("Статус: %d", w.Code)
	t.Logf("Тело ответа: %s", w.Body.String())
	t.Logf("==================")

	// Также вывести в стандартный вывод для наглядности при тестировании
	fmt.Printf("\n=== Тестовый запрос (%s %s) ===\n", req.Method, req.URL)
	fmt.Printf("Статус ответа: %d\n", w.Code)
	fmt.Printf("Тело ответа: %s\n", w.Body.String())
	fmt.Printf("==================\n")
}

// setupRouter настраивает тестовый маршрутизатор с mock-сервисами.
// Возвращает экземпляр gin.Engine с установленными маршрутами и middleware.

func setupRouter(callService service.CallService, authClient authclient.AuthClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	callHandler := NewCallHandler(callService, authClient)
	authMiddleware := middleware.NewAuthMiddleware(authClient)
	calls := router.Group("/calls")
	calls.Use(authMiddleware.AuthRequired())
	{
		calls.POST("", callHandler.CreateCall)
		calls.GET("", callHandler.GetAllCalls)
		calls.GET("/:id", callHandler.GetCall)
		calls.PATCH("/:id/status", callHandler.UpdateCallStatus)
		calls.DELETE("/:id", callHandler.DeleteCall)
	}
	return router
}

// TestCreateCall проверяет создание новой заявки.
// Тестирует успешное создание с валидными данными и проверку ответа.

func TestCreateCall(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	testCall := &model.Call{
		ID:          uuid.New(),
		ClientName:  "Test Client",
		PhoneNumber: "+1234567890",
		Description: "Test Description",
		Status:      "открыта",
		UserID:      testUserID,
	}
	testReq := &model.CreateCallRequest{
		ClientName:  "Test Client",
		PhoneNumber: "+1234567890",
		Description: "Test Description",
	}
	mockCallService.On("CreateCall", mock.Anything, mock.MatchedBy(func(req *model.CreateCallRequest) bool {
		return req.ClientName == testReq.ClientName &&
			req.PhoneNumber == testReq.PhoneNumber &&
			req.Description == testReq.Description
	}), testUserID).Return(testCall, nil)

	// Создаем запрос
	reqBody, _ := json.Marshal(testReq)
	req, _ := http.NewRequest("POST", "/calls", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+testToken)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.Bytes()
	t.Logf("Ответ сервера: %s", body)

	// Проверяем результат
	assert.Equal(t, http.StatusCreated, w.Code)
	var response model.Call
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, testCall.ID, response.ID)
	assert.Equal(t, testCall.ClientName, response.ClientName)
	assert.Equal(t, testCall.PhoneNumber, response.PhoneNumber)
	assert.Equal(t, testCall.Description, response.Description)
	assert.Equal(t, testCall.Status, response.Status)
	assert.Equal(t, testCall.UserID, response.UserID)

	// Проверяем, что все ожидаемые вызовы были выполнены
	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestGetCall проверяет получение заявки по ID.
// Тестирует успешное получение заявки с валидным ID.

func TestGetCall(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"
	testCallID := uuid.New()

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	testCall := &model.Call{
		ID:          testCallID,
		ClientName:  "Test Client",
		PhoneNumber: "+1234567890",
		Description: "Test Description",
		Status:      "открыта",
		UserID:      testUserID,
	}
	mockCallService.On("GetCallByID", mock.Anything, testCallID, testUserID).Return(testCall, nil)

	// Создаем запрос
	req, _ := http.NewRequest("GET", "/calls/"+testCallID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	var response model.Call
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, testCall.ID, response.ID)
	assert.Equal(t, testCall.ClientName, response.ClientName)
	assert.Equal(t, testCall.PhoneNumber, response.PhoneNumber)
	assert.Equal(t, testCall.Description, response.Description)
	assert.Equal(t, testCall.Status, response.Status)
	assert.Equal(t, testCall.UserID, response.UserID)

	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestGetAllCalls проверяет получение всех заявок пользователя.
// Тестирует успешное получение списка заявок.

func TestGetAllCalls(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	testCalls := []*model.Call{
		{
			ID:          uuid.New(),
			ClientName:  "Test Client 1",
			PhoneNumber: "+1234567890",
			Description: "Test Description 1",
			Status:      "открыта",
			UserID:      testUserID,
		},
		{
			ID:          uuid.New(),
			ClientName:  "Test Client 2",
			PhoneNumber: "+0987654321",
			Description: "Test Description 2",
			Status:      "закрыта",
			UserID:      testUserID,
		},
	}
	mockCallService.On("GetAllCalls", mock.Anything, testUserID).Return(testCalls, nil)

	// Создаем запрос
	req, _ := http.NewRequest("GET", "/calls", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	var response []*model.Call
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, len(testCalls), len(response))
	for i, call := range response {
		assert.Equal(t, testCalls[i].ID, call.ID)
		assert.Equal(t, testCalls[i].ClientName, call.ClientName)
		assert.Equal(t, testCalls[i].PhoneNumber, call.PhoneNumber)
		assert.Equal(t, testCalls[i].Description, call.Description)
		assert.Equal(t, testCalls[i].Status, call.Status)
		assert.Equal(t, testCalls[i].UserID, call.UserID)
	}

	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestUpdateCallStatus проверяет обновление статуса заявки.
// Тестирует успешное обновление статуса.

func TestUpdateCallStatus(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"
	testCallID := uuid.New()

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	mockCallService.On("UpdateCallStatus", mock.Anything, testCallID, "закрыта", testUserID).Return(nil)

	// Создаем запрос
	reqBody, _ := json.Marshal(map[string]string{"status": "закрыта"})
	req, _ := http.NewRequest("PATCH", "/calls/"+testCallID.String()+"/status", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+testToken)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "status updated successfully", response["message"])

	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestDeleteCall проверяет удаление заявки.
// Тестирует успешное удаление заявки.
func TestDeleteCall(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"
	testCallID := uuid.New()

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	mockCallService.On("DeleteCall", mock.Anything, testCallID, testUserID).Return(nil)

	// Создаем запрос
	req, _ := http.NewRequest("DELETE", "/calls/"+testCallID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "call deleted successfully", response["message"])

	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestCreateCall_InvalidPhone проверяет обработку неправильно переданного номера телефона.
// Тестирует успешную обработку неправильно переданного номера телефона.
func TestCreateCall_InvalidPhone(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	testReq := &model.CreateCallRequest{
		ClientName:  "Test Client",
		PhoneNumber: "invalid phone",
		Description: "Test Description",
	}
	mockCallService.On("CreateCall", mock.Anything, mock.MatchedBy(func(req *model.CreateCallRequest) bool {
		return req.ClientName == testReq.ClientName &&
			req.PhoneNumber == testReq.PhoneNumber &&
			req.Description == testReq.Description
	}), testUserID).Return(nil, service.ErrInvalidPhoneNumber)

	// Создаем запрос
	reqBody, _ := json.Marshal(testReq)
	req, _ := http.NewRequest("POST", "/calls", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+testToken)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestGetCall_Forbidden проверяет обработку 403 статуса ошибки.
// Тестирует успешную обработку 403 статуса ошибки.

func TestGetCall_Forbidden(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"
	testCallID := uuid.New()

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	mockCallService.On("GetCallByID", mock.Anything, testCallID, testUserID).Return(nil, service.ErrForbidden)

	// Создаем запрос
	req, _ := http.NewRequest("GET", "/calls/"+testCallID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusForbidden, w.Code)

	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestGetCall_NotFound проверяет обработку 404 статуса ошибки.
// Тестирует успешную обработку 404 статуса ошибки.

func TestGetCall_NotFound(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"
	testCallID := uuid.New()

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	mockCallService.On("GetCallByID", mock.Anything, testCallID, testUserID).Return(nil, service.ErrCallNotFound)

	// Создаем запрос
	req, _ := http.NewRequest("GET", "/calls/"+testCallID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestGetCall_NotFound проверяет обработку неправильно переданного статуса заявки.
// Тестирует успешную обработку неправильно переданного статуса заявки.

func TestUpdateCallStatus_InvalidStatus(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testUserID := uuid.New()
	testToken := "test-token"
	testCallID := uuid.New()

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(true, testUserID.String(), nil)
	mockCallService.On("UpdateCallStatus", mock.Anything, testCallID, "неверный статус", testUserID).Return(service.ErrInvalidStatus)

	// Создаем запрос
	reqBody, _ := json.Marshal(map[string]string{"status": "неверный статус"})
	req, _ := http.NewRequest("PATCH", "/calls/"+testCallID.String()+"/status", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+testToken)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockCallService.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
}

// TestInvalidAuth проверяет обработку невалидной аутентификации.
// Тестирует успешную обработку невалидной аутентификации.

func TestInvalidAuth(t *testing.T) {
	mockCallService := new(MockCallService)
	mockAuthClient := new(MockAuthClient)
	router := setupRouter(mockCallService, mockAuthClient)
	testToken := "invalid-token"

	// Настройка поведения mock-объектов
	mockAuthClient.On("ValidateToken", mock.Anything, testToken).Return(false, "", nil)

	// Создаем запрос
	req, _ := http.NewRequest("GET", "/calls", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Выводим детали запроса и ответа
	printRequestResponse(t, req, w)

	// Проверяем результат
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mockAuthClient.AssertExpectations(t)
}
