package service

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"

	"call-service/internal/model"
	"call-service/internal/repository"
)

// Константы ошибок для сервисного слоя

var (
	ErrInvalidPhoneNumber = errors.New("invalid phone number format")
	ErrCallNotFound       = errors.New("call not found")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidStatus      = errors.New("invalid status")
)

// Регулярное выражение для валидации номера телефона

var validPhoneRegex = regexp.MustCompile(`^[0-9+\-]+$`)

// CallService определяет интерфейс сервиса для работы с заявками

type CallService interface {
	CreateCall(ctx context.Context, req *model.CreateCallRequest, userID uuid.UUID) (*model.Call, error)
	GetCallByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Call, error)
	GetAllCalls(ctx context.Context, userID uuid.UUID) ([]*model.Call, error)
	UpdateCallStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID) error
	DeleteCall(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// callService реализует интерфейс CallService

type callService struct {
	callRepo repository.CallRepository
}

// NewCallService создает новый экземпляр сервиса

func NewCallService(callRepo repository.CallRepository) CallService {
	return &callService{callRepo: callRepo}
}

// CreateCall создает новую заявку

func (s *callService) CreateCall(ctx context.Context, req *model.CreateCallRequest, userID uuid.UUID) (*model.Call, error) {
	if !validPhoneRegex.MatchString(req.PhoneNumber) {
		return nil, ErrInvalidPhoneNumber
	}

	call := &model.Call{
		ClientName:  req.ClientName,
		PhoneNumber: req.PhoneNumber,
		Description: req.Description,
		Status:      "открыта",
		UserID:      userID,
	}

	if err := s.callRepo.Create(ctx, call); err != nil {
		return nil, err
	}

	return call, nil
}

// GetCallByID получает информацию о заявке по её ID

func (s *callService) GetCallByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Call, error) {
	call, err := s.callRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCallNotFound
	}

	if call.UserID != userID {
		return nil, ErrForbidden
	}

	return call, nil
}

// GetAllCalls получает список всех заявок пользователя

func (s *callService) GetAllCalls(ctx context.Context, userID uuid.UUID) ([]*model.Call, error) {
	return s.callRepo.GetAllByUserID(ctx, userID)
}

// UpdateCallStatus обновляет статус заявки

func (s *callService) UpdateCallStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID) error {
	if status != "открыта" && status != "закрыта" {
		return ErrInvalidStatus
	}

	call, err := s.callRepo.GetByID(ctx, id)
	if err != nil {
		return ErrCallNotFound
	}

	if call.UserID != userID {
		return ErrForbidden
	}

	return s.callRepo.UpdateStatus(ctx, id, status)
}

// DeleteCall удаляет заявку

func (s *callService) DeleteCall(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	call, err := s.callRepo.GetByID(ctx, id)
	if err != nil {
		return ErrCallNotFound
	}

	if call.UserID != userID {
		return ErrForbidden
	}

	return s.callRepo.Delete(ctx, id)
}
