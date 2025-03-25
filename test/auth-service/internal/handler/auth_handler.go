package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "auth-service/internal/proto"
	"auth-service/internal/service"
)

// AuthHandler реализует интерфейс AuthServiceServer для обработки аутентификационных запросов.
// Структура содержит сервис аутентификации и реализует все необходимые методы для регистрации,
// входа в систему и проверки токенов.

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewAuthHandler создает новый экземпляр AuthHandler с переданным сервисом аутентификации.

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register обрабатывает запрос на регистрацию нового пользователя.
// Проверяет корректность входных данных и вызывает соответствующий метод сервиса.
//
// Args:
//   ctx: контекст выполнения операции
//   req: структура с данными для регистрации (username и password)
//
// Returns:
//   *pb.RegisterResponse: токен и ID пользователя при успешной регистрации
//   error: ошибка с соответствующим кодом gRPC если:
//     - отсутствуют обязательные поля (codes.InvalidArgument)
//     - пользователь уже существует (codes.AlreadyExists)
//     - произошла внутренняя ошибка (codes.Internal)

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	token, userID, err := h.authService.Register(ctx, req.Username, req.Password)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &pb.RegisterResponse{
		Token:  token,
		UserId: userID.String(),
	}, nil
}

// Login обрабатывает запрос на вход в систему существующего пользователя.
// Проверяет корректность учетных данных и выдает токен доступа при успешной аутентификации.
//
// Args:
//   ctx: контекст выполнения операции
//   req: структура с данными для входа (username и password)
//
// Returns:
//   *pb.LoginResponse: токен и ID пользователя при успешном входе
//   error: ошибка с соответствующим кодом gRPC если:
//     - отсутствуют обязательные поля (codes.InvalidArgument)
//     - неверные учетные данные (codes.Unauthenticated)
//     - произошла внутренняя ошибка (codes.Internal)

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	token, userID, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to login user")
	}

	return &pb.LoginResponse{
		Token:  token,
		UserId: userID.String(),
	}, nil
}

// ValidateToken проверяет действительность токена аутентификации.
//
// Args:
//
//	ctx: контекст выполнения операции
//	req: структура с токеном для проверки
//
// Returns:
//
//	*pb.ValidateTokenResponse: структура содержит поле Valid и UserId при успешной проверке
//	error: ошибка с соответствующим кодом gRPC если:
//	  - отсутствует токен (codes.InvalidArgument)

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	userID, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID.String(),
	}, nil
}
