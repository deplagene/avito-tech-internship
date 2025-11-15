package pullrequest

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/types"
	"deplagene/avito-tech-internship/utils"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	prRepo   types.PullRequestRepository
	userRepo types.UserRepository
	teamRepo types.TeamRepository
	db       *pgxpool.Pool
	logger   *slog.Logger
}

func NewService(
	prRepo types.PullRequestRepository,
	userRepo types.UserRepository,
	teamRepo types.TeamRepository,
	db *pgxpool.Pool,
	logger *slog.Logger,
) *Service {
	return &Service{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		db:       db,
		logger:   logger,
	}
}

// CreatePullRequest создает PR и автоматически назначает до 2 ревьюверов из команды автора.
func (s *Service) CreatePullRequest(ctx context.Context, pr api.PullRequest) (*api.PullRequest, error) {
	const op = "pullrequest.service.CreatePullRequest"

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

	existingPR, err := s.prRepo.GetByID(ctx, tx, pr.PullRequestId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if existingPR != nil && existingPR.PullRequestId == pr.PullRequestId {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	author, err := s.userRepo.GetByID(ctx, tx, pr.AuthorId)
	if err != nil {
		return nil, err
	}

	candidates, err := s.userRepo.GetActiveUsersByTeam(ctx, tx, author.TeamName, author.UserId, 2)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pr.AssignedReviewers = make([]string, 0, 2)
	for _, candidate := range candidates {
		pr.AssignedReviewers = append(pr.AssignedReviewers, candidate.UserId)
	}

	pr.Status = api.PullRequestStatusOPEN
	pr.CreatedAt = api.Ptr(time.Now())

	if err := s.prRepo.Create(ctx, tx, pr); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &pr, nil
}

// MergePullRequest помечает PR как MERGED.
func (s *Service) MergePullRequest(ctx context.Context, prID string) (*api.PullRequest, error) {
	const op = "pullrequest.service.MergePullRequest"

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

	pr, err := s.prRepo.GetByID(ctx, tx, prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if pr.Status == api.PullRequestStatusMERGED {
		return pr, nil // ! если уже MERGED, просто возвращаем текущее состояние
	}

	if err := s.prRepo.Merge(ctx, tx, prID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pr.Status = api.PullRequestStatusMERGED
	pr.MergedAt = api.Ptr(time.Now())
	return pr, nil
}

// ReassignReviewer переназначает конкретного ревьювера на другого из его команды.
func (s *Service) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*api.PullRequest, string, error) {
	const op = "pullrequest.service.ReassignReviewer"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
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

	pr, err := s.prRepo.GetByID(ctx, tx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	if pr.Status == api.PullRequestStatusMERGED {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	isAssigned := slices.Contains(pr.AssignedReviewers, oldReviewerID)
	if !isAssigned {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	oldReviewerTeam, err := s.userRepo.GetTeamByUserID(ctx, tx, oldReviewerID)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	var currentReviewers []string
	currentReviewers = append(currentReviewers, pr.AssignedReviewers...)

	allTeamMembers, err := s.userRepo.GetActiveUsersByTeam(ctx, tx, oldReviewerTeam, "", 0)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	var replacementCandidates []api.User
	for _, member := range allTeamMembers {
		isCurrentReviewer := slices.Contains(currentReviewers, member.UserId)
		if member.UserId != oldReviewerID && !isCurrentReviewer {
			replacementCandidates = append(replacementCandidates, member)
		}
	}

	if len(replacementCandidates) == 0 {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	newReviewer := replacementCandidates[r.Intn(len(replacementCandidates))]

	if err := s.prRepo.RemoveReviewer(ctx, tx, prID, oldReviewerID); err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}
	if err := s.prRepo.AddReviewer(ctx, tx, prID, newReviewer.UserId); err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	for i, rID := range pr.AssignedReviewers {
		if rID == oldReviewerID {
			pr.AssignedReviewers[i] = newReviewer.UserId
			break
		}
	}

	return pr, newReviewer.UserId, nil
}

// GetPullRequestsByReviewer возвращает PR'ы, где пользователь назначен ревьювером.
func (s *Service) GetPullRequestsByReviewer(ctx context.Context, userID string) ([]api.PullRequestShort, error) {
	const op = "pullrequest.service.GetPullRequestsByReviewer"

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

	prs, err := s.prRepo.GetByReviewer(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return prs, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.PullRequestService = (*Service)(nil)
