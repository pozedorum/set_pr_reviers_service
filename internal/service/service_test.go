package service

import (
	"errors"
	"testing"
	"time"

	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/mocks"
	"github.com/pozedorum/set_pr_reviers_service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTeam_Success(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)
	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{
				UserID:   "user1",
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	mockRepo.On("TeamExists", "backend").Return(false)
	mockRepo.On("FindUserByID", "user1").Return(nil, errors.New("not found"))
	mockRepo.On("CreateTeam", team).Return(nil)

	service := NewPRServiceWithSeed(mockRepo, logger, 42)

	err = service.CreateTeam(team)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateTeam_AlreadyExists(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)
	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{
				UserID:   "user1",
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	mockRepo.On("TeamExists", "backend").Return(true)

	service := NewPRService(mockRepo, logger)

	err = service.CreateTeam(team)

	assert.Error(t, err)
	assert.Equal(t, ErrTeamAlreadyExists, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateTeam_EmptyTeam(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)
	service := NewPRService(mockRepo, logger)

	err = service.CreateTeam(nil)

	assert.Error(t, err)
	assert.Equal(t, ErrCreateEmptyTeam, err)
}

func TestCreatePR_Success(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	author := &entity.User{
		UserID:   "author1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	candidates := []*entity.User{
		{UserID: "user1", Username: "Bob", TeamName: "backend", IsActive: true},
		{UserID: "user2", Username: "Charlie", TeamName: "backend", IsActive: true},
		{UserID: "user3", Username: "David", TeamName: "backend", IsActive: true},
	}

	pr := &entity.PullRequest{
		PullRequestID:   "pr-123",
		PullRequestName: "Test PR",
		AuthorID:        "author1",
		Status:          entity.PullRequestStatusOpen,
	}

	mockRepo.On("FindUserByID", "author1").Return(author, nil)
	mockRepo.On("FindPRByID", "pr-123").Return(nil, errors.New("not found"))
	mockRepo.On("FindUsersByTeam", "backend").Return(candidates, nil)
	mockRepo.On("CreatePR", mock.MatchedBy(func(pr *entity.PullRequest) bool {
		return pr.PullRequestID == "pr-123" &&
			pr.AuthorID == "author1" &&
			len(pr.AssignedReviewers) == 2 &&
			pr.Status == entity.PullRequestStatusOpen
	})).Return(nil)

	service := NewPRServiceWithSeed(mockRepo, logger, 42)

	err = service.CreatePR(pr)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreatePR_AuthorNotFound(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	pr := &entity.PullRequest{
		PullRequestID:   "pr-123",
		PullRequestName: "Test PR",
		AuthorID:        "author1",
	}

	mockRepo.On("FindUserByID", "author1").Return(nil, errors.New("not found"))

	service := NewPRService(mockRepo, logger)

	err = service.CreatePR(pr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "author not found")
	mockRepo.AssertExpectations(t)
}

func TestCreatePR_AlreadyExists(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	author := &entity.User{
		UserID:   "author1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	existingPR := &entity.PullRequest{
		PullRequestID: "pr-123",
	}

	pr := &entity.PullRequest{
		PullRequestID:   "pr-123",
		PullRequestName: "Test PR",
		AuthorID:        "author1",
	}

	mockRepo.On("FindUserByID", "author1").Return(author, nil)
	mockRepo.On("FindPRByID", "pr-123").Return(existingPR, nil)

	service := NewPRService(mockRepo, logger)

	err = service.CreatePR(pr)

	assert.Error(t, err)
	assert.Equal(t, ErrPRAlreadyExists, err)
	mockRepo.AssertExpectations(t)
}

func TestMergePR_Success(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	existingPR := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"user1", "user2"},
		CreatedAt:         time.Now().Add(-time.Hour),
	}

	mockRepo.On("FindPRByID", "pr-123").Return(existingPR, nil)
	mockRepo.On("UpdatePR", mock.MatchedBy(func(pr *entity.PullRequest) bool {
		return pr.Status == entity.PullRequestStatusMerged &&
			pr.MergedAt.After(pr.CreatedAt)
	})).Return(nil)

	service := NewPRService(mockRepo, logger)

	mergedPR, err := service.MergePR("pr-123")

	assert.NoError(t, err)
	assert.Equal(t, entity.PullRequestStatusMerged, mergedPR.Status)
	assert.True(t, mergedPR.MergedAt.After(mergedPR.CreatedAt))
	mockRepo.AssertExpectations(t)
}

func TestMergePR_AlreadyMerged(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	existingPR := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusMerged,
		AssignedReviewers: []string{"user1", "user2"},
		CreatedAt:         time.Now().Add(-time.Hour),
		MergedAt:          time.Now().Add(-30 * time.Minute),
	}

	mockRepo.On("FindPRByID", "pr-123").Return(existingPR, nil)

	service := NewPRService(mockRepo, logger)

	mergedPR, err := service.MergePR("pr-123")

	assert.NoError(t, err)
	assert.Equal(t, entity.PullRequestStatusMerged, mergedPR.Status)
	mockRepo.AssertExpectations(t)
}

func TestMergePR_NotFound(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	mockRepo.On("FindPRByID", "pr-123").Return(nil, errors.New("not found"))

	service := NewPRService(mockRepo, logger)

	mergedPR, err := service.MergePR("pr-123")

	assert.Error(t, err)
	assert.Nil(t, mergedPR)
	assert.Contains(t, err.Error(), "find PR")
	mockRepo.AssertExpectations(t)
}

func TestReassignReviewer_Success(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	pr := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"user1", "user2"},
	}

	oldUser := &entity.User{
		UserID:   "user1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	candidates := []*entity.User{
		{UserID: "user3", Username: "Charlie", TeamName: "backend", IsActive: true},
		{UserID: "user4", Username: "David", TeamName: "backend", IsActive: true},
	}

	mockRepo.On("FindPRByID", "pr-123").Return(pr, nil)
	mockRepo.On("FindUserByID", "user1").Return(oldUser, nil)
	mockRepo.On("FindUsersByTeam", "backend").Return(candidates, nil)

	// Исправляем матчер - проверяем что user1 заменён, но не проверяем конкретно на кого
	mockRepo.On("UpdatePR", mock.MatchedBy(func(pr *entity.PullRequest) bool {
		return !contains(pr.AssignedReviewers, "user1") &&
			len(pr.AssignedReviewers) == 2 // Должно остаться 2 ревьювера
	})).Return(nil)

	service := NewPRServiceWithSeed(mockRepo, logger, 42)

	updatedPR, newReviewer, err := service.ReassignReviewer("pr-123", "user1")

	assert.NoError(t, err)
	assert.NotEqual(t, "user1", newReviewer)
	assert.True(t, contains([]string{"user3", "user4"}, newReviewer)) // Может быть user3 или user4
	assert.NotNil(t, updatedPR)
	assert.False(t, contains(updatedPR.AssignedReviewers, "user1"))
	assert.Len(t, updatedPR.AssignedReviewers, 2)
	mockRepo.AssertExpectations(t)
}

func TestReassignReviewer_MergedPR(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	pr := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusMerged,
		AssignedReviewers: []string{"user1", "user2"},
	}

	mockRepo.On("FindPRByID", "pr-123").Return(pr, nil)

	service := NewPRService(mockRepo, logger)

	updatedPR, newReviewer, err := service.ReassignReviewer("pr-123", "user1")

	assert.Error(t, err)
	assert.Nil(t, updatedPR)
	assert.Empty(t, newReviewer)
	assert.Equal(t, ErrCannotReassingOnMergedPR, err)
	mockRepo.AssertExpectations(t)
}

func TestReassignReviewer_ReviewerNotFound(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	pr := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"user2", "user3"}, // user1 нет в списке
	}

	oldUser := &entity.User{
		UserID:   "user1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	mockRepo.On("FindPRByID", "pr-123").Return(pr, nil)
	mockRepo.On("FindUserByID", "user1").Return(oldUser, nil)

	service := NewPRService(mockRepo, logger)

	updatedPR, newReviewer, err := service.ReassignReviewer("pr-123", "user1")

	assert.Error(t, err)
	assert.Nil(t, updatedPR)
	assert.Empty(t, newReviewer)
	assert.Contains(t, err.Error(), "reviewer user1 not assigned to this PR")
	mockRepo.AssertExpectations(t)
}

func TestSetUserActive_Success(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	user := &entity.User{
		UserID:   "user1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: false,
	}

	mockRepo.On("FindUserByID", "user1").Return(user, nil)
	mockRepo.On("SetActive", "user1", true).Return(nil)

	service := NewPRService(mockRepo, logger)

	updatedUser, err := service.SetUserActive("user1", true)

	assert.NoError(t, err)
	assert.True(t, updatedUser.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestSetUserActive_AlreadyActive(t *testing.T) {
	mockRepo := &mocks.Repository{}
	logger, err := logger.NewLogger("pr-service", "logger_for_tests")
	assert.NoError(t, err)

	user := &entity.User{
		UserID:   "user1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	mockRepo.On("FindUserByID", "user1").Return(user, nil)
	// SetActive не должен вызываться!

	service := NewPRService(mockRepo, logger)

	updatedUser, err := service.SetUserActive("user1", true)

	assert.NoError(t, err)
	assert.True(t, updatedUser.IsActive)
	mockRepo.AssertExpectations(t)
}

// Вспомогательная функция
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
