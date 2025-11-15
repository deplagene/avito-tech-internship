package pullrequest

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
)

type MockPullRequestRepository struct {
	mock.Mock
}

func (m *MockPullRequestRepository) Create(ctx context.Context, tx pgx.Tx, pr api.PullRequest) error {
	args := m.Called(ctx, tx, pr)
	return args.Error(0)
}

func (m *MockPullRequestRepository) GetByID(ctx context.Context, tx pgx.Tx, id string) (*api.PullRequest, error) {
	args := m.Called(ctx, tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) Merge(ctx context.Context, tx pgx.Tx, id string) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockPullRequestRepository) AddReviewer(ctx context.Context, tx pgx.Tx, prID, userID string) error {
	args := m.Called(ctx, tx, prID, userID)
	return args.Error(0)
}

func (m *MockPullRequestRepository) RemoveReviewer(ctx context.Context, tx pgx.Tx, prID, userID string) error {
	args := m.Called(ctx, tx, prID, userID)
	return args.Error(0)
}

func (m *MockPullRequestRepository) GetByReviewer(ctx context.Context, tx pgx.Tx, userID string) ([]api.PullRequestShort, error) {
	args := m.Called(ctx, tx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]api.PullRequestShort), args.Error(1)
}
