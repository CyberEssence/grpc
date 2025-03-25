package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"call-service/internal/model"
)

// CallRepository определяет интерфейс для работы с заявками в базе данных

type CallRepository interface {
	Create(ctx context.Context, call *model.Call) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Call, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Call, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// callRepository реализует интерфейс CallRepository

type callRepository struct {
	db *bun.DB
}

// NewCallRepository создает новый экземпляр репозитория

func NewCallRepository(db *bun.DB) CallRepository {
	return &callRepository{db: db}
}

// NewCallRepository создает новый экземпляр репозитория

func (r *callRepository) Create(ctx context.Context, call *model.Call) error {
	_, err := r.db.NewInsert().Model(call).Exec(ctx)
	return err
}

// GetByID получает заявку по её ID

func (r *callRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Call, error) {
	call := new(model.Call)
	err := r.db.NewSelect().Model(call).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return call, nil
}

// GetAllByUserID получает все заявки пользователя по его ID

func (r *callRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Call, error) {
	var calls []*model.Call
	err := r.db.NewSelect().Model(&calls).Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return calls, nil
}

// UpdateStatus обновляет статус заявки

func (r *callRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.NewUpdate().Model((*model.Call)(nil)).
		Set("status = ?", status).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// Delete удаляет заявку по её ID

func (r *callRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*model.Call)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
