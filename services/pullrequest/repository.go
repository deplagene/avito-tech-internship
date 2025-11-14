package pullrequest

import (
	"context"
	"deplagene/avito-tech-internship/api"   // Замените на ваш путь к модулю
	"deplagene/avito-tech-internship/types" // Замените на ваш путь к модулю
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PullRequestRepository реализует интерфейс types.PullRequestRepository для работы с PostgreSQL.
type PullRequestRepository struct {
	db *pgxpool.Pool
}

// NewPullRequestRepository создает новый экземпляр PullRequestRepository.
func NewPullRequestRepository(db *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

// Create создает новый Pull Request и назначает ревьюверов.
func (r *PullRequestRepository) Create(ctx context.Context, tx pgx.Tx, pr api.PullRequest) error {
	sql := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.Exec(ctx, sql, pr.PullRequestId, pr.PullRequestName, pr.AuthorId, pr.Status, time.Now())
	if err != nil {
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		if err := r.AddReviewer(ctx, tx, pr.PullRequestId, reviewerID); err != nil {
			return err
		}
	}
	return nil
}

// GetByID возвращает Pull Request по его ID.
func (r *PullRequestRepository) GetByID(ctx context.Context, tx pgx.Tx, id string) (*api.PullRequest, error) {
	pr := &api.PullRequest{}
	var statusStr string
	sql := `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at,
		       ARRAY_AGG(rev.user_id) FILTER (WHERE rev.user_id IS NOT NULL) AS assigned_reviewers
		FROM pull_requests pr
		LEFT JOIN reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE pr.pull_request_id = $1
		GROUP BY pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
	`
	err := tx.QueryRow(ctx, sql, id).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&statusStr,
		&pr.CreatedAt,
		&pr.MergedAt,
		&pr.AssignedReviewers,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("pull request not found") // TODO: Define custom error types
	}
	if err != nil {
		return nil, err
	}
	pr.Status = api.PullRequestStatus(statusStr)
	return pr, nil
}

// Merge помечает Pull Request как MERGED.
func (r *PullRequestRepository) Merge(ctx context.Context, tx pgx.Tx, id string) error {
	sql := "UPDATE pull_requests SET status = $1, merged_at = $2 WHERE pull_request_id = $3 AND status = $4"
	_, err := tx.Exec(ctx, sql, api.PullRequestStatusMERGED, time.Now(), id, api.PullRequestStatusOPEN)
	return err
}

// AddReviewer добавляет ревьювера к Pull Request'у.
func (r *PullRequestRepository) AddReviewer(ctx context.Context, tx pgx.Tx, prID, userID string) error {
	sql := "INSERT INTO reviewers (pull_request_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	_, err := tx.Exec(ctx, sql, prID, userID)
	return err
}

// RemoveReviewer удаляет ревьювера из Pull Request'а.
func (r *PullRequestRepository) RemoveReviewer(ctx context.Context, tx pgx.Tx, prID, userID string) error {
	sql := "DELETE FROM reviewers WHERE pull_request_id = $1 AND user_id = $2"
	_, err := tx.Exec(ctx, sql, prID, userID)
	return err
}

// GetByReviewer возвращает список Pull Request'ов, где пользователь назначен ревьювером.
func (r *PullRequestRepository) GetByReviewer(ctx context.Context, tx pgx.Tx, userID string) ([]api.PullRequestShort, error) {
	var prs []api.PullRequestShort
	sql := `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.user_id = $1
	`
	rows, err := tx.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pr api.PullRequestShort
		var statusStr string
		if err := rows.Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &statusStr); err != nil {
			return nil, err
		}
		pr.Status = api.PullRequestShortStatus(statusStr)
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.PullRequestRepository = (*PullRequestRepository)(nil)
