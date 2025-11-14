package team

import (
	"context"
	"deplagene/avito-tech-internship/api"
	"deplagene/avito-tech-internship/types"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service реализует интерфейс types.TeamService.
type Service struct {
	teamRepo types.TeamRepository
	userRepo types.UserRepository
	db       *pgxpool.Pool // Для управления транзакциями
}

// NewService создает новый экземпляр TeamService.
func NewService(teamRepo types.TeamRepository, userRepo types.UserRepository, db *pgxpool.Pool) *Service {
	return &Service{
		teamRepo: teamRepo,
		userRepo: userRepo,
		db:       db,
	}
}

// CreateTeam создает новую команду и ее участников.
func (s *Service) CreateTeam(ctx context.Context, team api.Team) (*api.Team, error) {
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

	// Проверяем, существует ли команда
	existingTeam, err := s.teamRepo.GetByName(ctx, tx, team.TeamName)
	if err != nil && err.Error() != "team not found" { // TODO: Use custom error type
		return nil, err
	}
	if existingTeam != nil && existingTeam.TeamName == team.TeamName {
		return nil, fmt.Errorf("team already exists") // TODO: Use custom error type (TEAM_EXISTS)
	}

	if err = s.teamRepo.Create(ctx, tx, team); err != nil {
		return nil, err
	}

	return &team, nil
}

// GetTeam возвращает команду по имени.
func (s *Service) GetTeam(ctx context.Context, name string) (*api.Team, error) {
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

	team, err := s.teamRepo.GetByName(ctx, tx, name)
	if err != nil {
		if err.Error() == "team not found" { // TODO: Use custom error type
			return nil, fmt.Errorf("team not found") // TODO: Use custom error type (NOT_FOUND)
		}
		return nil, err
	}
	return team, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.TeamService = (*Service)(nil)
