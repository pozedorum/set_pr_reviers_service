package interfaces

import "github.com/pozedorum/set_pr_reviers_service/internal/entity"

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
