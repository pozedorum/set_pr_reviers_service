package interfaces

import (
	"context"

	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
)

type Repository interface {
	Close() error
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

type Service interface {
	// Teams
	CreateTeam(team *entity.Team) error
	GetTeam(teamName string) (*entity.Team, error)

	// Users
	SetUserActive(userID string, isActive bool) (*entity.User, error)
	GetUserReviews(userID string) ([]*entity.PullRequest, error)

	// PRs
	CreatePR(prID, prName, authorID string) (*entity.PullRequest, error)
	MergePR(prID string) (*entity.PullRequest, error)
	ReassignReviewer(prID, oldUserID string) (*entity.PullRequest, string, error)
}

type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type Logger interface {
	Debug(operation, message string, keyvals ...interface{})
	Info(operation, message string, keyvals ...interface{})
	Warn(operation, message string, keyvals ...interface{})
	Error(operation, message string, keyvals ...interface{})
	Shutdown()
}
