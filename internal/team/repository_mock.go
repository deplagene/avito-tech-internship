package team

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
)

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(ctx context.Context, tx pgx.Tx, team api.Team) error {
	args := m.Called(ctx, tx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByName(ctx context.Context, tx pgx.Tx, name string) (*api.Team, error) {
	args := m.Called(ctx, tx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.Team), args.Error(1)
}
