package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/generated"
	"github.com/pozedorum/set_pr_reviers_service/internal/service"
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

		switch err {
		case service.ErrTeamAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{
				"code":    "TEAM_EXISTS",
				"message": err.Error(),
			}})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := entityTeamToGenerated(team)
	c.JSON(http.StatusCreated, gin.H{"team": response})
}

func (s *PRServer) handleGetTeam(c *gin.Context) {
	teamName := getTeamNameFromContext(c)
	if teamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_name parameter is required"})
		return
	}

	team, err := s.serv.GetTeam(teamName)
	if err != nil {
		s.logger.Error("GET_TEAM_ERROR", "Failed to get team", "error", err, "team_name", teamName)

		switch err {
		case service.ErrNoTeam:
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			}})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := entityTeamToGenerated(*team)
	c.JSON(http.StatusOK, response)
}

func (s *PRServer) handleSetUserActive(c *gin.Context) {
	var request struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	usr, err := s.serv.SetUserActive(request.UserID, request.IsActive)
	if err != nil {
		s.logger.Error("SET_USER_ACTIVE_ERROR", "Failed to set user active",
			"error", err, "user_id", request.UserID, "is_active", request.IsActive)

		switch err {
		case service.ErrNoUser:
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			}})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := entityUserToGenerated(*usr)
	c.JSON(http.StatusOK, gin.H{"user": response})
}

func (s *PRServer) handleGetUserReviews(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id parameter is required"})
		return
	}

	prs, err := s.serv.GetUserReviews(userID)
	if err != nil {
		s.logger.Error("GET_USER_REVIEWS_ERROR", "Failed to get user reviews",
			"error", err, "user_id", userID)

		switch err {
		case service.ErrNoUser:
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			}})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := make([]generated.PullRequestShort, len(prs))
	for i, pr := range prs {
		response[i] = entityPRToShortGenerated(*pr)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": response,
	})
}

func (s *PRServer) handleCreatePR(c *gin.Context) {
	var request struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	pr := entity.PullRequest{
		PullRequestID:   request.PullRequestID,
		PullRequestName: request.PullRequestName,
		AuthorID:        request.AuthorID,
	}

	err := s.serv.CreatePR(&pr)
	if err != nil {
		s.logger.Error("CREATE_PR_ERROR", "Failed to create PR",
			"error", err, "pr_id", pr.PullRequestID, "author_id", pr.AuthorID)

		switch err {
		case service.ErrPRAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{
				"code":    "PR_EXISTS",
				"message": err.Error(),
			}})
		case service.ErrNoUser:
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			}})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := entityPRToGenerated(pr)
	c.JSON(http.StatusCreated, gin.H{"pr": response})
}

func (s *PRServer) handleMergePR(c *gin.Context) {
	var request struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	pr, err := s.serv.MergePR(request.PullRequestID)
	if err != nil {
		s.logger.Error("MERGE_PR_ERROR", "Failed to merge PR",
			"error", err, "pr_id", request.PullRequestID)

		switch err {
		case service.ErrNoPR:
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			}})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := entityPRToGenerated(*pr)
	c.JSON(http.StatusOK, gin.H{"pr": response})
}

func (s *PRServer) handleReassignReviewer(c *gin.Context) {
	var request struct {
		PullRequestID string `json:"pull_request_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedPR, newReviewerID, err := s.serv.ReassignReviewer(request.PullRequestID, request.OldReviewerID)
	if err != nil {
		s.logger.Error("REASSIGN_REVIEWER_ERROR", "Failed to reassign reviewer",
			"error", err, "pr_id", request.PullRequestID, "old_reviewer", request.OldReviewerID)

		switch err {
		case service.ErrNoPR, service.ErrNoUser:
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			}})
		case service.ErrCannotReassingOnMergedPR:
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{
				"code":    "PR_MERGED",
				"message": err.Error(),
			}})
		case service.ErrNoReplacementCandidate:
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{
				"code":    "NO_CANDIDATE",
				"message": err.Error(),
			}})
		default:
			// Проверяем текст ошибки для более специфичных случаев
			if err.Error() == fmt.Sprintf("reviewer %s not assigned to this PR", request.OldReviewerID) {
				c.JSON(http.StatusConflict, gin.H{"error": gin.H{
					"code":    "NOT_ASSIGNED",
					"message": err.Error(),
				}})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}
		return
	}

	response := entityPRToGenerated(*updatedPR)
	c.JSON(http.StatusOK, gin.H{
		"pr":          response,
		"replaced_by": newReviewerID,
	})
}

// Вспомогательные функции для извлечения параметров из контекста
func getTeamNameFromContext(c *gin.Context) string {
	if teamName, exists := c.Get("team_name"); exists {
		return teamName.(string)
	}
	return ""
}

func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}
