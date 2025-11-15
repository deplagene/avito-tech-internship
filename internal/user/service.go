package user

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/types"
	"deplagene/avito-tech-internship/utils"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	userRepo types.UserRepository
	db       *pgxpool.Pool
	logger   *slog.Logger
}

func NewService(userRepo types.UserRepository, db *pgxpool.Pool, logger *slog.Logger) *Service {
	return &Service{
		userRepo: userRepo,
		db:       db,
		logger:   logger,
	}
}

// SetUserIsActive устанавливает флаг активности пользователя.
func (s *Service) SetUserIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	const op = "user.service.SetUserIsActive"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if r := recover(); r != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.logger.Error("failed to rollback transaction after panic", utils.Err(err))
			}
			panic(r)
		} else if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.logger.Error("failed to rollback transaction", utils.Err(err))
			}
		} else {
			err = tx.Commit(ctx)
		}
	}()

	existingUser, err := s.userRepo.GetByID(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if existingUser == nil {
		return nil, types.ErrNotFound
	}

	if err := s.userRepo.SetIsActive(ctx, tx, userID, isActive); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	existingUser.IsActive = isActive
	return existingUser, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.UserService = (*Service)(nil)
