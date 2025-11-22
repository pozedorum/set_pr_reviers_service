package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

type TeamRepo struct {
	db     *sqlx.DB
	logger interfaces.Logger
}

func (repo *TeamRepo) Create(team *entity.Team) error {
	return nil
}

func (repo *TeamRepo) FindByName(teamName string) (*entity.Team, error) {
	return nil, nil
}

func (repo *TeamRepo) Exists(teamName string) bool {
	return false
}
