package service

import (
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"auth-service/internal/model"
	"auth-service/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService определяет интерфейс для аутентификационных операций.
// Предоставляет методы для регистрации, входа в систему и проверки токенов.

type AuthService interface {
	Register(ctx context.Context, username, password string) (string, uuid.UUID, error)
	Login(ctx context.Context, username, password string) (string, uuid.UUID, error)
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)
}

// authService реализует интерфейс AuthService для обработки аутентификационных операций.
// Использует репозиторий для работы с данными пользователей и JWT для аутентификации.

type authService struct {
	userRepo repository.UserRepository
	jwtKey   []byte
}

// NewAuthService создает новый экземпляр сервиса аутентификации.
// Принимает репозиторий пользователей и ключ для подписи JWT-токенов.

func NewAuthService(userRepo repository.UserRepository, jwtKey string) AuthService {
	return &authService{userRepo: userRepo, jwtKey: []byte(jwtKey)}
}

// Register регистрирует нового пользователя в системе.
// Проверяет уникальность имени пользователя, хеширует пароль и создает запись в базе данных.
// Генерирует JWT-токен для успешной регистрации.

func (s *authService) Register(ctx context.Context, username, password string) (string, uuid.UUID, error) {
	existingUser, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return "", uuid.Nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", uuid.Nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", uuid.Nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", uuid.Nil, err
	}

	return token, user.ID, nil
}

// Login аутентифицирует пользователя по имени и паролю.
// Проверяет существование пользователя и корректность пароля.
// Генерирует JWT-токен при успешной аутентификации.

func (s *authService) Login(ctx context.Context, username, password string) (string, uuid.UUID, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", uuid.Nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", uuid.Nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", uuid.Nil, err
	}

	return token, user.ID, nil
}

// ValidateToken проверяет действительность JWT-токена и возвращает ID пользователя.
// Проверяет подпись токена, срок действия и существование пользователя.

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

// generateToken генерирует JWT-токен для указанного ID пользователя.
// Устанавливает срок действия токена на 24 часа.

func (s *authService) generateToken(userID uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID.String()
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
