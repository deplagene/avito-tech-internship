package user

import (
	"context"
	"deplagene/avito-tech-internship/api"   // Замените на ваш путь к модулю
	"deplagene/avito-tech-internship/types" // Замените на ваш путь к модулю
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository реализует интерфейс types.UserRepository для работы с PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository создает новый экземпляр UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Upsert создает или обновляет пользователя.
func (r *UserRepository) Upsert(ctx context.Context, tx pgx.Tx, user api.TeamMember, teamName string) error {
	sql := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active
	`
	_, err := tx.Exec(ctx, sql, user.UserId, user.Username, teamName, user.IsActive)
	return err
}

// GetByID возвращает пользователя по его ID.
func (r *UserRepository) GetByID(ctx context.Context, tx pgx.Tx, id string) (*api.User, error) {
	user := &api.User{}
	sql := "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1"
	err := tx.QueryRow(ctx, sql, id).Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found") // TODO: Define custom error types
	}
	return user, err
}

// SetIsActive устанавливает флаг активности пользователя.
func (r *UserRepository) SetIsActive(ctx context.Context, tx pgx.Tx, id string, isActive bool) error {
	sql := "UPDATE users SET is_active = $1 WHERE user_id = $2"
	_, err := tx.Exec(ctx, sql, isActive, id)
	return err
}

// GetActiveUsersByTeam возвращает активных пользователей из команды, исключая указанного пользователя.
func (r *UserRepository) GetActiveUsersByTeam(ctx context.Context, tx pgx.Tx, teamName string, excludeUserID string, limit int) ([]api.User, error) {
	var users []api.User
	sql := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = TRUE AND user_id != $2
		ORDER BY RANDOM()
		LIMIT $3
	`
	rows, err := tx.Query(ctx, sql, teamName, excludeUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user api.User
		if err := rows.Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// GetTeamByUserID возвращает имя команды, к которой принадлежит пользователь.
func (r *UserRepository) GetTeamByUserID(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	var teamName string
	sql := "SELECT team_name FROM users WHERE user_id = $1"
	err := tx.QueryRow(ctx, sql, userID).Scan(&teamName)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("user not found") // TODO: Define custom error types
	}
	return teamName, err
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.UserRepository = (*UserRepository)(nil)
