package server

import (
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/generated"
)

// Перевод из openapi в entity
func generatedTeamToEntity(gTeam generated.Team) entity.Team {
	members := make([]entity.TeamMember, len(gTeam.Members))
	for i, m := range gTeam.Members {
		members[i] = entity.TeamMember{
			UserID:   m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	return entity.Team{
		TeamName: gTeam.TeamName,
		Members:  members,
	}
}

func generatedPRToEntity(gPR generated.PullRequest) entity.PullRequest {
	return entity.PullRequest{
		PullRequestID:     gPR.PullRequestId,
		PullRequestName:   gPR.PullRequestName,
		AuthorID:          gPR.AuthorId,
		Status:            entity.PullRequestStatus(gPR.Status),
		AssignedReviewers: gPR.AssignedReviewers,
		CreatedAt:         gPR.CreatedAt,
		MergedAt:          gPR.MergedAt,
	}
}

func generatedUserToEntity(gUser generated.User) entity.User {
	return entity.User{
		UserID:   gUser.UserId,
		Username: gUser.Username,
		TeamName: gUser.TeamName,
		IsActive: gUser.IsActive,
	}
}

// Перевод из entity в openapi
func entityTeamToGenerated(eTeam entity.Team) generated.Team {
	members := make([]generated.TeamMember, len(eTeam.Members))
	for i, m := range eTeam.Members {
		members[i] = generated.TeamMember{
			UserId:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	return generated.Team{
		TeamName: eTeam.TeamName,
		Members:  members,
	}
}

func entityUserToGenerated(eUser entity.User) generated.User {
	return generated.User{
		UserId:   eUser.UserID,
		Username: eUser.Username,
		TeamName: eUser.TeamName,
		IsActive: eUser.IsActive,
	}
}

func entityPRToGenerated(ePR entity.PullRequest) generated.PullRequest {
	return generated.PullRequest{
		PullRequestId:     ePR.PullRequestID,
		PullRequestName:   ePR.PullRequestName,
		AuthorId:          ePR.AuthorID,
		Status:            generated.PullRequestStatus(ePR.Status),
		AssignedReviewers: ePR.AssignedReviewers,
		CreatedAt:         ePR.CreatedAt,
		MergedAt:          ePR.MergedAt,
	}
}

func entityPRToShortGenerated(ePR entity.PullRequest) generated.PullRequestShort {
	return generated.PullRequestShort{
		PullRequestId:   ePR.PullRequestID,
		PullRequestName: ePR.PullRequestName,
		AuthorId:        ePR.AuthorID,
		Status:          generated.PullRequestShortStatus(ePR.Status),
	}
}
