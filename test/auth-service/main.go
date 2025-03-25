package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"auth-service/internal/handler"
	pb "auth-service/internal/proto"
	"auth-service/internal/repository"
	"auth-service/internal/service"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Основная функция программы, которая запускает gRPC-сервер аутентификации.
// Устанавливает соединение с базой данных PostgreSQL, создает сервисы и запускает сервер.
func main() {
	// Загружаем конфигурационные параметры из переменных окружения
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "auth_service")
	jwtKey := getEnv("JWT_KEY", "59347add01aacae058d36f0593c39412cec5630e66fbc290ecb933024514189d60da0cbc9b3184b721373415cee4eccf4aeff3e6c1518d97cd38c8e83dd58a17896841a6e8f36e999cff36bb56b8bf91844082a64c0ff92c618cdb484e7fb54773731d41d73d78eb72056a1c5411781b928018a5ae930cdd07253b061edfbaf437054d6c76d5b105318fe5d6ff56b868de0da03be72332ae752cf0e05e757718e9404ac4d1fc69c301f316602658ae242e19025da4ea8f96ab5b7910597e25fc02b5a9660729b888d66f0e0bf93a685172e91a0d0029c75610421bb51b8a5c436090208119e327fe5235e4d5d3ce34d09de562eb887c23257514ca65a3b759f1")
	grpcPort := getEnv("GRPC_PORT", "50051")

	// Формируем строку подключения к PostgreSQL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Создаем подключение к базе данных
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// Проверяем соединение с базой данных
	if err := checkDatabaseConnection(db); err != nil {
		log.Fatalf("Cannot proceed due to database connection failure: %v", err)
	}

	// Создаем репозиторий и сервис для работы с пользователями
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtKey)

	// Создаем TCP-соединение для gRPC-сервера
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Создаем gRPC-сервер с обработчиком контекста
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				return handler(ctx, req)
			},
		),
	)

	// Регистрируем рефлексию для gRPC
	reflection.Register(grpcServer)

	// Создаем и регистрируем обработчик аутентификации
	authHandler := handler.NewAuthHandler(authService)
	pb.RegisterAuthServiceServer(grpcServer, authHandler)

	// Запускаем сервер
	log.Printf("Starting gRPC server on port %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Проверяет соединение с базой данных.
// Возвращает ошибку, если соединение невозможно установить.
func checkDatabaseConnection(db *bun.DB) error {
	err := db.Ping()
	if err != nil {
		log.Printf("Database connection test failed: %v", err)
		return err
	}
	log.Println("Database connection successful!")
	return nil
}

// Получает значение переменной окружения с указанным именем.
// Если переменная не установлена, возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
