package user

import (
	"context"
	"deplagene/avito-tech-internship/api"
	"deplagene/avito-tech-internship/types"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service реализует интерфейс types.UserService.
type Service struct {
	userRepo types.UserRepository
	db       *pgxpool.Pool // Для управления транзакциями
}

// NewService создает новый экземпляр UserService.
func NewService(userRepo types.UserRepository, db *pgxpool.Pool) *Service {
	return &Service{
		userRepo: userRepo,
		db:       db,
	}
}

// SetUserIsActive устанавливает флаг активности пользователя.
func (s *Service) SetUserIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	// Проверяем, существует ли пользователь
	existingUser, err := s.userRepo.GetByID(ctx, tx, userID)
	if err != nil {
		if err.Error() == "user not found" { // TODO: Use custom error type
			return nil, fmt.Errorf("user not found") // TODO: Use custom error type (NOT_FOUND)
		}
		return nil, err
	}

	if err := s.userRepo.SetIsActive(ctx, tx, userID, isActive); err != nil {
		return nil, err
	}

	existingUser.IsActive = isActive // Обновляем статус в возвращаемом объекте
	return existingUser, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.UserService = (*Service)(nil)
