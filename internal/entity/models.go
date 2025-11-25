package entity

import "time"

type User struct {
	UserID   string
	Username string
	TeamName string
	IsActive bool
}

type Team struct {
	TeamName string
	Members  []TeamMember
}

type TeamMember struct {
	UserID   string
	Username string
	IsActive bool
}

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            PullRequestStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          time.Time
}

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)
