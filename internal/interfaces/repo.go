package interfaces

import "github.com/pozedorum/set_pr_reviers_service/internal/entity"

type Repository interface {
	// Users
	CreateUser(user *entity.User) error
	FindUserByID(userID string) (*entity.User, error)
	UpdateUser(user *entity.User) error
	FindUsersByTeam(teamName string) ([]*entity.User, error)
	SetActive(userID string, isActive bool) error

	// Teams
	CreateTeam(team *entity.Team) error
	FindTeamByName(teamName string) (*entity.Team, error)
	TeamExists(teamName string) bool

	// PRs
	CreatePR(pr *entity.PullRequest) error
	FindPRByID(prID string) (*entity.PullRequest, error)
	ClosePR(pr *entity.PullRequest) error
	UpdatePR(pr *entity.PullRequest) error
	FindPRsByReviewer(userID string) ([]*entity.PullRequest, error)
}
