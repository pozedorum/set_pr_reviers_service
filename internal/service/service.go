package service

import (
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

type PrService struct {
	userRepo interfaces.UserRepository
	teamRepo interfaces.TeamRepository
	prRepo   interfaces.PRRepository
	logger   interfaces.Logger
}

func NewPRService(
	userRepo interfaces.UserRepository,
	teamRepo interfaces.TeamRepository,
	prRepo interfaces.PRRepository,
) interfaces.Service {
	return &PrService{
		userRepo: userRepo,
		teamRepo: teamRepo,
		prRepo:   prRepo,
	}
}

func (servs *PrService) CreateTeam(team *entity.Team) error {
	return nil
}

func (servs *PrService) GetTeam(teamName string) (*entity.Team, error) {
	return nil, nil
}

func (servs *PrService) SetUserActive(userID string, isActive bool) (*entity.User, error) {
	return nil, nil
}

func (servs *PrService) GetUserReviews(userID string) ([]*entity.PullRequest, error) {
	return nil, nil
}

func (servs *PrService) CreatePR(prID, prName, authorID string) (*entity.PullRequest, error) {
	return nil, nil
}

func (servs *PrService) MergePR(prID string) (*entity.PullRequest, error) {
	return nil, nil
}

func (servs *PrService) ReassignReviewer(prID, oldUserID string) (*entity.PullRequest, string, error) {
	return nil, "", nil
}
