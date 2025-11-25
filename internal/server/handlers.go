package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/generated"
)

func (s *PRServer) handleCreateTeam(c *gin.Context) {
	var request generated.Team
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	team := generatedTeamToEntity(request)

	if err := s.serv.CreateTeam(&team); err != nil {
		s.logger.Error("CREATE_TEAM_ERROR", "Failed to create team", "error", err, "team_name", team.TeamName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "team created"})
}

func (s *PRServer) handleGetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_name parameter is required"})
		return
	}

	team, err := s.serv.GetTeam(teamName)
	if err != nil {
		s.logger.Error("GET_TEAM_ERROR", "Failed to get team", "error", err, "team_name", teamName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if team == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	response := entityTeamToGenerated(*team)
	c.JSON(http.StatusOK, response)
}

func (s *PRServer) handleSetUserActive(c *gin.Context) {
	var (
		usr     *entity.User
		err     error
		request struct {
			UserID   string `json:"user_id"`
			IsActive bool   `json:"is_active"`
		}
	)

	if err = c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if usr, err = s.serv.SetUserActive(request.UserID, request.IsActive); err != nil {
		s.logger.Error("SET_USER_ACTIVE_ERROR", "Failed to set user active",
			"error", err, "user_id", request.UserID, "is_active", request.IsActive)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user activity updated", "status": usr.IsActive})
}

func (s *PRServer) handleGetUserReviews(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id parameter is required"})
		return
	}

	prs, err := s.serv.GetUserReviews(userID)
	if err != nil {
		s.logger.Error("GET_USER_REVIEWS_ERROR", "Failed to get user reviews",
			"error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]generated.PullRequestShort, len(prs))
	for i, pr := range prs {
		response[i] = entityPRToShortGenerated(*pr)
	}

	c.JSON(http.StatusOK, response)
}

func (s *PRServer) handleCreatePR(c *gin.Context) {
	var request generated.PullRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	pr := generatedPRToEntity(request)

	err := s.serv.CreatePR(&pr)
	if err != nil {
		s.logger.Error("CREATE_PR_ERROR", "Failed to create PR",
			"error", err, "pr_id", pr.PullRequestID, "author_id", pr.AuthorID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := entityPRToGenerated(pr)
	c.JSON(http.StatusCreated, response)
}

func (s *PRServer) handleMergePR(c *gin.Context) {
	var request struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if _, err := s.serv.MergePR(request.PullRequestID); err != nil {
		s.logger.Error("MERGE_PR_ERROR", "Failed to merge PR",
			"error", err, "pr_id", request.PullRequestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "PR merged"})
}

func (s *PRServer) handleReassignReviewer(c *gin.Context) {
	//		newUserID string
	var request struct {
		PullRequestID string `json:"pull_request_id"`
		OldReviewerID string `json:"old_reviewer_id"`
		NewReviewerID string `json:"new_reviewer_id,omitempty"` // опционально - если не указан, выбирается автоматически
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedPR, _, err := s.serv.ReassignReviewer(request.PullRequestID, request.OldReviewerID)
	if err != nil {
		s.logger.Error("REASSIGN_REVIEWER_ERROR", "Failed to reassign reviewer",
			"error", err, "pr_id", request.PullRequestID, "old_reviewer", request.OldReviewerID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := entityPRToGenerated(*updatedPR)
	c.JSON(http.StatusOK, response)
}
