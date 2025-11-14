package pullrequest

import (
	"context"
	"deplagene/avito-tech-internship/api"
	"deplagene/avito-tech-internship/types"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service реализует интерфейс types.PullRequestService.
type Service struct {
	prRepo   types.PullRequestRepository
	userRepo types.UserRepository
	teamRepo types.TeamRepository
	db       *pgxpool.Pool // Для управления транзакциями
}

// NewService создает новый экземпляр PullRequestService.
func NewService(
	prRepo types.PullRequestRepository,
	userRepo types.UserRepository,
	teamRepo types.TeamRepository,
	db *pgxpool.Pool,
) *Service {
	return &Service{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		db:       db,
	}
}

// CreatePullRequest создает PR и автоматически назначает до 2 ревьюверов из команды автора.
func (s *Service) CreatePullRequest(ctx context.Context, pr api.PullRequest) (*api.PullRequest, error) {
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

	// Проверяем, существует ли PR
	existingPR, err := s.prRepo.GetByID(ctx, tx, pr.PullRequestId)
	if err != nil && err.Error() != "pull request not found" { // TODO: Use custom error type
		return nil, err
	}
	if existingPR != nil && existingPR.PullRequestId == pr.PullRequestId {
		return nil, fmt.Errorf("pull request already exists") // TODO: Use custom error type (PR_EXISTS)
	}

	// Получаем автора PR
	author, err := s.userRepo.GetByID(ctx, tx, pr.AuthorId)
	if err != nil {
		if err.Error() == "user not found" { // TODO: Use custom error type
			return nil, fmt.Errorf("author not found") // TODO: Use custom error type (NOT_FOUND)
		}
		return nil, err
	}

	// Получаем активных пользователей из команды автора, исключая самого автора
	candidates, err := s.userRepo.GetActiveUsersByTeam(ctx, tx, author.TeamName, author.UserId, 2)
	if err != nil {
		return nil, err
	}

	// Назначаем ревьюверов
	pr.AssignedReviewers = make([]string, 0, 2)
	for _, candidate := range candidates {
		pr.AssignedReviewers = append(pr.AssignedReviewers, candidate.UserId)
	}

	pr.Status = api.PullRequestStatusOPEN
	pr.CreatedAt = api.Ptr(time.Now())

	if err := s.prRepo.Create(ctx, tx, pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

// MergePullRequest помечает PR как MERGED (идемпотентная операция).
func (s *Service) MergePullRequest(ctx context.Context, prID string) (*api.PullRequest, error) {
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

	pr, err := s.prRepo.GetByID(ctx, tx, prID)
	if err != nil {
		if err.Error() == "pull request not found" { // TODO: Use custom error type
			return nil, fmt.Errorf("pull request not found") // TODO: Use custom error type (NOT_FOUND)
		}
		return nil, err
	}

	if pr.Status == api.PullRequestStatusMERGED {
		return pr, nil // Идемпотентность: если уже MERGED, просто возвращаем текущее состояние
	}

	if err := s.prRepo.Merge(ctx, tx, prID); err != nil {
		return nil, err
	}

	pr.Status = api.PullRequestStatusMERGED
	pr.MergedAt = api.Ptr(time.Now())
	return pr, nil
}

// ReassignReviewer переназначает конкретного ревьювера на другого из его команды.
func (s *Service) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*api.PullRequest, string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to begin transaction: %w", err)
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

	pr, err := s.prRepo.GetByID(ctx, tx, prID)
	if err != nil {
		if err.Error() == "pull request not found" { // TODO: Use custom error type
			return nil, "", fmt.Errorf("pull request not found") // TODO: Use custom error type (NOT_FOUND)
		}
		return nil, "", err
	}

	if pr.Status == api.PullRequestStatusMERGED {
		return nil, "", fmt.Errorf("cannot reassign on merged PR") // TODO: Use custom error type (PR_MERGED)
	}

	// Проверяем, был ли старый ревьювер назначен
	isAssigned := false
	for _, rID := range pr.AssignedReviewers {
		if rID == oldReviewerID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, "", fmt.Errorf("reviewer is not assigned to this PR") // TODO: Use custom error type (NOT_ASSIGNED)
	}

	// Получаем команду старого ревьювера
	oldReviewerTeam, err := s.userRepo.GetTeamByUserID(ctx, tx, oldReviewerID)
	if err != nil {
		if err.Error() == "user not found" { // TODO: Use custom error type
			return nil, "", fmt.Errorf("old reviewer not found") // TODO: Use custom error type (NOT_FOUND)
		}
		return nil, "", err
	}

	// Ищем кандидатов на замену (активных пользователей из той же команды, исключая текущих ревьюверов)
	var currentReviewers []string
	for _, rID := range pr.AssignedReviewers {
		currentReviewers = append(currentReviewers, rID)
	}

	allTeamMembers, err := s.userRepo.GetActiveUsersByTeam(ctx, tx, oldReviewerTeam, "", 0) // Get all active members
	if err != nil {
		return nil, "", err
	}

	var replacementCandidates []api.User
	for _, member := range allTeamMembers {
		isCurrentReviewer := false
		for _, cr := range currentReviewers {
			if member.UserId == cr {
				isCurrentReviewer = true
				break
			}
		}
		if member.UserId != oldReviewerID && !isCurrentReviewer {
			replacementCandidates = append(replacementCandidates, member)
		}
	}

	if len(replacementCandidates) == 0 {
		return nil, "", fmt.Errorf("no active replacement candidate in team") // TODO: Use custom error type (NO_CANDIDATE)
	}

	// Выбираем случайного кандидата
	rand.Seed(time.Now().UnixNano())
	newReviewer := replacementCandidates[rand.Intn(len(replacementCandidates))]

	// Удаляем старого и добавляем нового ревьювера
	if err := s.prRepo.RemoveReviewer(ctx, tx, prID, oldReviewerID); err != nil {
		return nil, "", err
	}
	if err := s.prRepo.AddReviewer(ctx, tx, prID, newReviewer.UserId); err != nil {
		return nil, "", err
	}

	// Обновляем список ревьюверов в PR объекте
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

	prs, err := s.prRepo.GetByReviewer(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

// Проверка соответствия интерфейсу во время компиляции
var _ types.PullRequestService = (*Service)(nil)
