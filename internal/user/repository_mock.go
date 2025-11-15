package user

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Upsert(ctx context.Context, tx pgx.Tx, user api.TeamMember, teamName string) error {
	args := m.Called(ctx, tx, user, teamName)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, tx pgx.Tx, id string) (*api.User, error) {
	args := m.Called(ctx, tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.User), args.Error(1)
}

func (m *MockUserRepository) SetIsActive(ctx context.Context, tx pgx.Tx, id string, isActive bool) error {
	args := m.Called(ctx, tx, id, isActive)
	return args.Error(0)
}

func (m *MockUserRepository) GetActiveUsersByTeam(ctx context.Context, tx pgx.Tx, teamName string, excludeUserID string, limit int) ([]api.User, error) {
	args := m.Called(ctx, tx, teamName, excludeUserID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]api.User), args.Error(1)
}

func (m *MockUserRepository) GetTeamByUserID(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	args := m.Called(ctx, tx, userID)
	return args.String(0), args.Error(1)
}
