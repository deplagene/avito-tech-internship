package user

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/types"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository создает новый экземпляр UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Upsert создает или обновляет пользователя.
func (r *UserRepository) Upsert(ctx context.Context, tx pgx.Tx, user api.TeamMember, teamName string) error {
	const op = "user.repository.Upsert"

	_, err := tx.Exec(ctx, upsertUserQuery, user.UserId, user.Username, teamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetByID возвращает пользователя по его ID.
func (r *UserRepository) GetByID(ctx context.Context, tx pgx.Tx, id string) (*api.User, error) {
	const op = "user.repository.GetByID"

	user := &api.User{}

	err := tx.QueryRow(ctx, getBydIdUserQuery, id).Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// SetIsActive устанавливает флаг активности пользователя.
func (r *UserRepository) SetIsActive(ctx context.Context, tx pgx.Tx, id string, isActive bool) error {
	const op = "user.repository.SetIsActive"

	_, err := tx.Exec(ctx, setIsActiveUserQuery, isActive, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetActiveUsersByTeam возвращает активных пользователей из команды, исключая указанного пользователя.
func (r *UserRepository) GetActiveUsersByTeam(ctx context.Context, tx pgx.Tx, teamName string, excludeUserID string, limit int) ([]api.User, error) {
	const op = "user.repository.GetActiveUsersByTeam"

	var users []api.User

	rows, err := tx.Query(ctx, getActiveUsersByTeamQuery, teamName, excludeUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var user api.User
		if err := rows.Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, user)
	}
	return users, fmt.Errorf("%s: %w", op, rows.Err())
}

// GetTeamByUserID возвращает имя команды, к которой принадлежит пользователь.
func (r *UserRepository) GetTeamByUserID(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	const op = "user.repository.GetTeamByUserID"

	var teamName string

	err := tx.QueryRow(ctx, getTeamByUserIdQuery, userID).Scan(&teamName)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return teamName, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.UserRepository = (*UserRepository)(nil)
