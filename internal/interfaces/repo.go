package interfaces

import "github.com/pozedorum/set_pr_reviers_service/internal/entity"

type UserRepository interface {
	Create(user *entity.User) error
	FindByID(userID string) (*entity.User, error)
	Update(user *entity.User) error
	FindByTeam(teamName string) ([]*entity.User, error)
	SetActive(userID string, isActive bool) error
}

type TeamRepository interface {
	Create(team *entity.Team) error
	FindByName(teamName string) (*entity.Team, error)
	Exists(teamName string) bool
}

type PRRepository interface {
	Create(pr *entity.PullRequest) error
	FindByID(prID string) (*entity.PullRequest, error)
	Update(pr *entity.PullRequest) error
	FindByReviewer(userID string) ([]*entity.PullRequest, error)
}
