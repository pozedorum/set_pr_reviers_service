package repository

import (
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

type PrRepositoryInMemory struct {
	users        map[string]*entity.User
	teams        map[string]*entity.Team
	poolRequests map[string]*entity.PullRequest
}

func NewPrRepositoryInMemory() interfaces.Repository {
	return &PrRepositoryInMemory{
		users:        make(map[string]*entity.User),
		teams:        make(map[string]*entity.Team),
		poolRequests: make(map[string]*entity.PullRequest),
	}
}

func (repo *PrRepositoryInMemory) CreateUser(user *entity.User) error
func (repo *PrRepositoryInMemory) FindUserByID(userID string) (*entity.User, error)
func (repo *PrRepositoryInMemory) UpdateUser(user *entity.User) error
func (repo *PrRepositoryInMemory) FindUsersByTeam(teamName string) ([]*entity.User, error)
func (repo *PrRepositoryInMemory) SetActive(userID string, isActive bool) error
func (repo *PrRepositoryInMemory) CreateTeam(team *entity.Team) error
func (repo *PrRepositoryInMemory) FindTeamByName(teamName string) (*entity.Team, error)
func (repo *PrRepositoryInMemory) TeamExists(teamName string) bool
func (repo *PrRepositoryInMemory) CreatePR(pr *entity.PullRequest) error
func (repo *PrRepositoryInMemory) FindPRByID(prID string) (*entity.PullRequest, error)
func (repo *PrRepositoryInMemory) ClosePR(pr *entity.PullRequest) error
func (repo *PrRepositoryInMemory) UpdatePR(pr *entity.PullRequest) error
func (repo *PrRepositoryInMemory) FindPRsByReviewer(userID string) ([]*entity.PullRequest, error)
