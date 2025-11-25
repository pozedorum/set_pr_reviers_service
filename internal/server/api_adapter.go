package server

import (
	"github.com/gin-gonic/gin"
	"github.com/pozedorum/set_pr_reviers_service/internal/generated"
)

type APIAdapter struct {
	server *PRServer
}

func NewAPIAdapter(server *PRServer) generated.ServerInterface {
	return &APIAdapter{server: server}
}

func (a *APIAdapter) PostTeamAdd(c *gin.Context) {
	a.server.handleCreateTeam(c)
}

func (a *APIAdapter) GetTeamGet(c *gin.Context, params generated.GetTeamGetParams) {
	c.Set("team_name", params.TeamName)
	a.server.handleGetTeam(c)
}

func (a *APIAdapter) PostUsersSetIsActive(c *gin.Context) {
	a.server.handleSetUserActive(c)
}

func (a *APIAdapter) GetUsersGetReview(c *gin.Context, params generated.GetUsersGetReviewParams) {
	c.Set("user_id", params.UserId)
	a.server.handleGetUserReviews(c)
}

func (a *APIAdapter) PostPullRequestCreate(c *gin.Context) {
	a.server.handleCreatePR(c)
}

func (a *APIAdapter) PostPullRequestMerge(c *gin.Context) {
	a.server.handleMergePR(c)
}

func (a *APIAdapter) PostPullRequestReassign(c *gin.Context) {
	a.server.handleReassignReviewer(c)
}
