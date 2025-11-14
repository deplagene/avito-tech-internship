package team

import (
	"context"
	"deplagene/avito-tech-internship/api"
	"deplagene/avito-tech-internship/services/user" // Import the user repository
	"deplagene/avito-tech-internship/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TeamRepository реализует интерфейс types.TeamRepository для работы с PostgreSQL.
type TeamRepository struct {
	db *pgxpool.Pool
}

// NewTeamRepository создает новый экземпляр TeamRepository.
func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create создает новую команду и ее участников в рамках одной транзакции.
func (r *TeamRepository) Create(ctx context.Context, tx pgx.Tx, team api.Team) error {
	// Создаем команду
	sql := "INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING"
	if _, err := tx.Exec(ctx, sql, team.TeamName); err != nil {
		return err
	}

	// Добавляем/обновляем участников
	userRepo := user.NewUserRepository(r.db) // Create an instance of UserRepository
	for _, member := range team.Members {
		if err := userRepo.Upsert(ctx, tx, member, team.TeamName); err != nil {
			return err // Ошибка при добавлении участника, транзакция будет отменена
		}
	}

	return nil
}

// GetByName возвращает команду и ее участников по имени.
func (r *TeamRepository) GetByName(ctx context.Context, tx pgx.Tx, name string) (*api.Team, error) {
	// Получаем информацию о команде
	team := &api.Team{TeamName: name}

	// Получаем участников команды
	sql := "SELECT user_id, username, is_active FROM users WHERE team_name = $1"
	rows, err := tx.Query(ctx, sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []api.TeamMember
	for rows.Next() {
		var member api.TeamMember
		if err := rows.Scan(&member.UserId, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	team.Members = members
	return team, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.TeamRepository = (*TeamRepository)(nil)
