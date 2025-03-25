package repository

import (
	"auth-service/internal/model"
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// UserRepository определяет интерфейс для работы с данными пользователей.
// Предоставляет методы для создания и получения пользователей из базы данных.

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// userRepository реализует интерфейс UserRepository для работы с базой данных через bun.
// Использует контекст для управления временем выполнения операций.

type userRepository struct {
	db *bun.DB
}

// NewUserRepository создает новый экземпляр репозитория пользователей.
// Принимает подключение к базе данных через bun.DB.

func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{db: db}
}

// Create сохраняет нового пользователя в базу данных.
// Использует контекст для отмены операции при необходимости.

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

// GetByUsername извлекает пользователя из базы данных по его имени.
// Использует контекст для отмены операции при необходимости.

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByID извлекает пользователя из базы данных по его ID.
// Использует контекст для отмены операции при необходимости.

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}
