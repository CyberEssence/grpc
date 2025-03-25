package authclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "call-service/proto"
)

// AuthClient представляет интерфейс клиента аутентификации.
// Предоставляет методы для регистрации пользователя, входа в систему и проверки токенов.

type AuthClient interface {
	Register(ctx context.Context, username, password string) (string, string, error)
	Login(ctx context.Context, username, password string) (string, string, error)
	ValidateToken(ctx context.Context, token string) (bool, string, error)
	Close() error
}

// authClient реализует интерфейс AuthClient для взаимодействия с gRPC-сервисом аутентификации.

type authClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

// NewAuthClient создает новый экземпляр клиента аутентификации.

func NewAuthClient(addr string) (AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthServiceClient(conn)
	return &authClient{client: client, conn: conn}, nil
}

// Register регистрирует нового пользователя в системе.
//
// Параметры:
// ctx - контекст выполнения запроса
// username - имя пользователя для регистрации
// password - пароль пользователя
//
// Возвращает:
// token - токен аутентификации нового пользователя
// userId - ID зарегистрированного пользователя
// error - ошибка регистрации, если произошла

func (c *authClient) Register(ctx context.Context, username, password string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	resp, err := c.client.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		return "", "", err
	}

	return resp.Token, resp.UserId, nil
}

// Login выполняет вход пользователя в систему.
//
// Параметры:
// ctx - контекст выполнения запроса
// username - имя пользователя
// password - пароль пользователя
//
// Возвращает:
// token - токен аутентификации
// userId - ID пользователя
// error - ошибка входа, если произошла

func (c *authClient) Login(ctx context.Context, username, password string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	resp, err := c.client.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		return "", "", err
	}

	return resp.Token, resp.UserId, nil
}

// ValidateToken проверяет валидность токена аутентификации.
//
// Параметры:
// ctx - контекст выполнения запроса
// token - токен для проверки
//
// Возвращает:
// valid - true если токен валиден, false если нет
// userId - ID пользователя, если токен валиден
// error - ошибка проверки токена, если произошла

func (c *authClient) ValidateToken(ctx context.Context, token string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Token: token,
	})

	if err != nil {
		return false, "", err
	}

	return resp.Valid, resp.UserId, nil
}

// Close закрывает gRPC подключение к сервису аутентификации.

func (c *authClient) Close() error {
	return c.conn.Close()
}
