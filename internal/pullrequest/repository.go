package pullrequest

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/types"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestRepository struct {
	db *pgxpool.Pool
}

func NewPullRequestRepository(db *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

// Create создает новый Pull Request и назначает ревьюверов.
func (r *PullRequestRepository) Create(ctx context.Context, tx pgx.Tx, pr api.PullRequest) error {
	const op = "pullrequest.repository.Create"

	_, err := tx.Exec(ctx, createPullRequestQuery, pr.PullRequestId, pr.PullRequestName, pr.AuthorId, pr.Status, time.Now())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, reviewerID := range pr.AssignedReviewers {
		if err := r.AddReviewer(ctx, tx, pr.PullRequestId, reviewerID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

// GetByID возвращает Pull Request по его ID.
func (r *PullRequestRepository) GetByID(ctx context.Context, tx pgx.Tx, id string) (*api.PullRequest, error) {
	const op = "pullrequest.repository.GetByID"

	pr := &api.PullRequest{}
	var statusStr string

	err := tx.QueryRow(ctx, getPullRequestByIdQuery, id).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&statusStr,
		&pr.CreatedAt,
		&pr.MergedAt,
		&pr.AssignedReviewers,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	
	pr.Status = api.PullRequestStatus(statusStr)
	return pr, nil
}

// Merge помечает Pull Request как MERGED.
func (r *PullRequestRepository) Merge(ctx context.Context, tx pgx.Tx, id string) error {
	const op = "pullrequest.repository.Merge"

	_, err := tx.Exec(ctx, setMergeStatusQuery, api.PullRequestStatusMERGED, time.Now(), id, api.PullRequestStatusOPEN)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// AddReviewer добавляет ревьювера к Pull Request'у.
func (r *PullRequestRepository) AddReviewer(ctx context.Context, tx pgx.Tx, prID, userID string) error {
	const op = "pullrequest.repository.AddReviewer"

	_, err := tx.Exec(ctx, addReviewerQuery, prID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// RemoveReviewer удаляет ревьювера из Pull Request'а.
func (r *PullRequestRepository) RemoveReviewer(ctx context.Context, tx pgx.Tx, prID, userID string) error {
	const op = "pullrequest.repository.RemoveReviewer"

	_, err := tx.Exec(ctx, deleteReviewerQuery, prID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetByReviewer возвращает список Pull Request'ов, где пользователь назначен ревьювером.
func (r *PullRequestRepository) GetByReviewer(ctx context.Context, tx pgx.Tx, userID string) ([]api.PullRequestShort, error) {
	const op = "pullrequest.repository.GetByReviewer"

	var prs []api.PullRequestShort

	rows, err := tx.Query(ctx, getPullRequestsByReviewerQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var pr api.PullRequestShort
		var statusStr string
		if err := rows.Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &statusStr); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		pr.Status = api.PullRequestShortStatus(statusStr)
		prs = append(prs, pr)
	}
	return prs, fmt.Errorf("%s: %w", op, rows.Err())
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.PullRequestRepository = (*PullRequestRepository)(nil)
