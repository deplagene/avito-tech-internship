package pullrequest

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/internal/team"
	"deplagene/avito-tech-internship/internal/user"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullRequestService_CreatePullRequest(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPool.Close()

	mockPRRepo := new(MockPullRequestRepository)
	mockUserRepo := new(user.MockUserRepository)
	mockTeamRepo := new(team.MockTeamRepository)

	prService := NewService(mockPRRepo, mockUserRepo, mockTeamRepo, mockPool, nil)

	ctx := context.Background()
	pr := api.PullRequest{
		PullRequestId:   "pr1",
		PullRequestName: "Test PR",
		AuthorId:        "user1",
	}

	author := &api.User{
		UserId:   "user1",
		Username: "Author",
		TeamName: "team1",
		IsActive: true,
	}

	reviewers := []api.User{
		{UserId: "user2", Username: "Reviewer1", TeamName: "team1", IsActive: true},
		{UserId: "user3", Username: "Reviewer2", TeamName: "team1", IsActive: true},
	}

	mockPool.On("Begin", ctx).Return(mockPool, nil)
	mockPool.On("Commit", ctx).Return(nil)
	mockUserRepo.On("GetByID", ctx, mock.Anything, pr.AuthorId).Return(author, nil)
	mockUserRepo.On("GetActiveUsersByTeam", ctx, mock.Anything, author.TeamName, pr.AuthorId, 2).Return(reviewers, nil)
	mockPRRepo.On("GetByID", ctx, mock.Anything, pr.PullRequestId).Return(nil, nil) // Not found
	mockPRRepo.On("Create", ctx, mock.Anything, mock.AnythingOfType("api.PullRequest")).Return(nil)

	createdPR, err := prService.CreatePullRequest(ctx, pr)

	assert.NoError(t, err)
	assert.NotNil(t, createdPR)
	assert.Equal(t, pr.PullRequestId, createdPR.PullRequestId)
	assert.Equal(t, pr.PullRequestName, createdPR.PullRequestName)
	assert.Equal(t, pr.AuthorId, createdPR.AuthorId)
	assert.Equal(t, api.PullRequestStatusOPEN, createdPR.Status)
	assert.Len(t, createdPR.AssignedReviewers, 2)
	assert.Contains(t, createdPR.AssignedReviewers, "user2")
	assert.Contains(t, createdPR.AssignedReviewers, "user3")

	mockPool.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}