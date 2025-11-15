package types

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"

	"github.com/jackc/pgx/v5"
)

// TeamRepository определяет методы для работы с командами.
type TeamRepository interface {
	Create(ctx context.Context, tx pgx.Tx, team api.Team) error
	GetByName(ctx context.Context, tx pgx.Tx, name string) (*api.Team, error)
}

// UserRepository определяет методы для работы с пользователями.
type UserRepository interface {
	Upsert(ctx context.Context, tx pgx.Tx, user api.TeamMember, teamName string) error
	GetByID(ctx context.Context, tx pgx.Tx, id string) (*api.User, error)
	SetIsActive(ctx context.Context, tx pgx.Tx, id string, isActive bool) error
	GetActiveUsersByTeam(ctx context.Context, tx pgx.Tx, teamName string, excludeUserID string, limit int) ([]api.User, error)
	GetTeamByUserID(ctx context.Context, tx pgx.Tx, userID string) (string, error)
}

// PullRequestRepository определяет методы для работы с Pull Request'ами.
type PullRequestRepository interface {
	Create(ctx context.Context, tx pgx.Tx, pr api.PullRequest) error
	GetByID(ctx context.Context, tx pgx.Tx, id string) (*api.PullRequest, error)
	Merge(ctx context.Context, tx pgx.Tx, id string) error
	AddReviewer(ctx context.Context, tx pgx.Tx, prID, userID string) error
	RemoveReviewer(ctx context.Context, tx pgx.Tx, prID, userID string) error
	GetByReviewer(ctx context.Context, tx pgx.Tx, userID string) ([]api.PullRequestShort, error)
}

// TeamService определяет методы бизнес-логики для работы с командами.
type TeamService interface {
	CreateTeam(ctx context.Context, team api.Team) (*api.Team, error)
	GetTeam(ctx context.Context, name string) (*api.Team, error)
}

// UserService определяет методы бизнес-логики для работы с пользователями.
type UserService interface {
	SetUserIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error)
}

// PullRequestService определяет методы бизнес-логики для работы с Pull Request'ами.
type PullRequestService interface {
	CreatePullRequest(ctx context.Context, pr api.PullRequest) (*api.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*api.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*api.PullRequest, string, error)
	GetPullRequestsByReviewer(ctx context.Context, userID string) ([]api.PullRequestShort, error)
}
