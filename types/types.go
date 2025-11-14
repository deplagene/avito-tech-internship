package types

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	Id       uuid.UUID `json:"user_id"`
	Name     string    `json:"username"`
	TeamName string    `json:"team_name"`
	IsActive bool      `json:"is_active"`
}

type Team struct {
	Name string `json:"team_name"` // Уникальное имя
}

type PullRequest struct {
	Id       uuid.UUID   `json:"pull_request_id"`
	AuthorId uuid.UUID   `json:"author_id"`
	Status   string      `json:"status"`
	Users    []uuid.UUID `json:"users"`
}

type PullRequestStore interface {
	Create(ctx context.Context, pr *PullRequest) error
}
