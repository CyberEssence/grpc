package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"call-service/internal/handler"
	"call-service/internal/middleware"
	"call-service/internal/repository"
	"call-service/internal/service"
	"call-service/pkg/authclient"
)

// Выполняет инициализацию всех компонентов и запускает HTTP-сервер.
func main() {
	// Получение переменных окружения для конфигурации
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "call_service")
	authServiceAddr := getEnv("AUTH_SERVICE_ADDR", "localhost:50051")
	httpPort := getEnv("HTTP_PORT", "8080")

	// Установка подключения к PostgreSQL базе данных
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// Создание клиента для аутентификации
	authClient, err := authclient.NewAuthClient(authServiceAddr)
	if err != nil {
		log.Fatalf("failed to create auth client: %v", err)
	}
	defer authClient.Close()

	// Инициализация репозиториев
	callRepo := repository.NewCallRepository(db)

	// Создание сервисов
	callService := service.NewCallService(callRepo)

	// Создание обработчиков
	authHandler := handler.NewAuthHandler(authClient)
	callHandler := handler.NewCallHandler(callService, authClient)

	// Создание middleware для аутентификации
	authMiddleware := middleware.NewAuthMiddleware(authClient)

	// Создание маршрутизатора
	router := gin.Default()

	// Регистрация маршрутов аутентификации
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	// Группа маршрутов для работы с вызовами
	calls := router.Group("/calls")
	calls.Use(authMiddleware.AuthRequired())
	{
		calls.POST("", callHandler.CreateCall)
		calls.GET("", callHandler.GetAllCalls)
		calls.GET("/:id", callHandler.GetCall)
		calls.PATCH("/:id/status", callHandler.UpdateCallStatus)
		calls.DELETE("/:id", callHandler.DeleteCall)
	}

	// Запуск HTTP-сервера
	log.Printf("Starting HTTP server on port %s", httpPort)
	if err := router.Run(":" + httpPort); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}

// getEnv получает значение переменной окружения с дефолтным значением.
// Если переменная окружения не установлена, возвращается defaultValue.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
