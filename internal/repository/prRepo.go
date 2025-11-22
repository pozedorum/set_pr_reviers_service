package repository

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

type PrRepo struct {
	db     *sqlx.DB
	logger interfaces.Logger
}

func (repo *PrRepo) Create(pr *entity.PullRequest) error {
	return nil
}

func (repo *PrRepo) FindByID(prID string) (*entity.PullRequest, error) {
	return nil, nil
}

func (repo *PrRepo) Update(pr *entity.PullRequest) error {
	return nil
}

func (repo *PrRepo) FindByReviewer(userID string) ([]*entity.PullRequest, error) {
	return nil, nil
}
