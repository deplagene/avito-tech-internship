package team

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/internal/user"
	"deplagene/avito-tech-internship/types"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create создает новую команду и ее участников в рамках одной транзакции.
func (r *TeamRepository) Create(ctx context.Context, tx pgx.Tx, team api.Team) error {
	const op = "team.repository.Create"

	if _, err := tx.Exec(ctx, createTeamQuery, team.TeamName); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userRepo := user.NewUserRepository(r.db)
	for _, member := range team.Members {
		if err := userRepo.Upsert(ctx, tx, member, team.TeamName); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

// GetByName возвращает команду и ее участников по имени.
func (r *TeamRepository) GetByName(ctx context.Context, tx pgx.Tx, name string) (*api.Team, error) {
	const op = "team.repository.GetByName"

	team := &api.Team{TeamName: name}

	rows, err := tx.Query(ctx, getByNameTeamQuery, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var members []api.TeamMember
	for rows.Next() {
		var member api.TeamMember
		if err := rows.Scan(&member.UserId, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	team.Members = members
	return team, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.TeamRepository = (*TeamRepository)(nil)
