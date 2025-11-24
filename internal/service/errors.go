package service

import (
	"errors"
	"fmt"
)

var (
	// service errors
	ErrCreateEmptyTeam   = errors.New("cannot create empty team")
	ErrNoTeam            = errors.New("no such team")
	ErrEmptyTeamName     = errors.New("empty team name")
	ErrEmptyTeam         = errors.New("team has no members")
	ErrTeamAlreadyExists = errors.New("team already exists")

	ErrNoUser            = errors.New("no such user")
	ErrEmptyUserID       = errors.New("empty team member user ID")
	ErrEmptyUserUsername = errors.New("empty team member username")

	ErrNoPR            = errors.New("no such pull request")
	ErrEmptyPRID       = errors.New("empty pull request ID")
	ErrEmptyPRName     = errors.New("empty pull request name")
	ErrEmptyPRAuthorID = errors.New("empty author ID of pull request")
	ErrPRAlreadyExists = errors.New("pull request already exists")

	ErrCannotReassingOnMergedPR = errors.New("cannot reasing reviewer on closed pull request")
	ErrWrongReassignReviewer    = errors.New("reassigned reviewer not in a team")
	ErrNoReplacementCandidate   = errors.New("no available candidates for replacement")
)

func ErrUserAlreadyExists(UserID string) error {
	return fmt.Errorf("user %s already exists", UserID)
}
