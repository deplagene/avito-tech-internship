package team

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
	teamRepo types.TeamRepository
	userRepo types.UserRepository
	db       *pgxpool.Pool
	logger   *slog.Logger
}

func NewService(teamRepo types.TeamRepository, userRepo types.UserRepository, db *pgxpool.Pool, logger *slog.Logger) *Service {
	return &Service{
		teamRepo: teamRepo,
		userRepo: userRepo,
		db:       db,
		logger:   logger,
	}
}

// CreateTeam создает новую команду и ее участников.
func (s *Service) CreateTeam(ctx context.Context, team api.Team) (createdTeam *api.Team, err error) {
	const op = "team.service.CreateTeam"

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

	existingTeam, err := s.teamRepo.GetByName(ctx, tx, team.TeamName)
	if err != nil {
		return createdTeam, fmt.Errorf("%s: %w", op, err)
	}
	if existingTeam == nil {
		if err = s.teamRepo.Create(ctx, tx, team); err != nil {
			return createdTeam, fmt.Errorf("%s: %w", op, err)
		}
	}

	for _, member := range team.Members {
		if err = s.userRepo.Upsert(ctx, tx, member, team.TeamName); err != nil {
			return createdTeam, fmt.Errorf("%s: %w", op, err)
		}
	}

	return &team, nil
}

// GetTeam возвращает команду по имени.
func (s *Service) GetTeam(ctx context.Context, name string) (*api.Team, error) {
	const op = "team.service.GetTeam"

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

	team, err := s.teamRepo.GetByName(ctx, tx, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if team == nil {
		return nil, types.ErrNotFound
	}
	return team, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.TeamService = (*Service)(nil)
