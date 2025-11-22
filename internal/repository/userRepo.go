package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

type UserRepo struct {
	db     *sqlx.DB
	logger interfaces.Logger
}

func (repo *UserRepo) Create(user *entity.User) error {
	return nil
}

func (repo *UserRepo) FindByID(userID string) (*entity.User, error) {
	return nil, nil
}

func (repo *UserRepo) Update(user *entity.User) error {
	return nil
}

func (repo *UserRepo) FindByTeam(teamName string) ([]*entity.User, error) {
	return nil, nil
}

func (repo *UserRepo) SetActive(userID string, isActive bool) error {
	return nil
}
