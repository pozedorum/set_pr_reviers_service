package repository

import (
	"errors"
)

var (
	ErrNoTeam = errors.New("no such team")
	ErrNoUser = errors.New("no such user")
	ErrNoPR   = errors.New("no such pull request")
)
