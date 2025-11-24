package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

type PrService struct {
	repo   interfaces.Repository
	logger interfaces.Logger
	random *rand.Rand
}

func NewPRService(repo interfaces.Repository, logger interfaces.Logger) interfaces.Service {
	return &PrService{
		repo:   repo,
		logger: logger,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewPRServiceWithSeed создает сервис с константным сидом (для тестов)
func NewPRServiceWithSeed(repo interfaces.Repository, logger interfaces.Logger, seed int64) interfaces.Service {
	return &PrService{
		repo:   repo,
		logger: logger,
		random: rand.New(rand.NewSource(seed)),
	}
}

func (servs *PrService) SetSeed(seed int64) {
	servs.random = rand.New(rand.NewSource(seed))
}

func (servs *PrService) CreateTeam(team *entity.Team) error {
	start := time.Now()

	if team == nil {
		servs.logger.Warn("SERVICE_CREATE_TEAM", "Attempt to create empty team",
			"duration_ms", time.Since(start).Milliseconds())
		return ErrCreateEmptyTeam
	}

	servs.logger.Debug("SERVICE_CREATE_TEAM", "Starting team creation",
		"team_name", team.TeamName,
		"members_count", len(team.Members))

	// Проверка корректности данных
	if err := checkTeamCorrectness(team); err != nil {
		servs.logger.Warn("SERVICE_CREATE_TEAM", "Team data validation failed",
			"team_name", team.TeamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return err
	}

	// Проверка валидности создания
	if err := servs.validateTeamCreation(team); err != nil {
		servs.logger.Warn("SERVICE_CREATE_TEAM", "Team creation validation failed",
			"team_name", team.TeamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return err
	}

	if err := servs.repo.CreateTeam(team); err != nil {
		servs.logger.Error("SERVICE_CREATE_TEAM", "Failed to create team in repository",
			"team_name", team.TeamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return err
	}

	servs.logger.Info("SERVICE_CREATE_TEAM", "Team created successfully",
		"team_name", team.TeamName,
		"members_count", len(team.Members),
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (servs *PrService) GetTeam(teamName string) (*entity.Team, error) {
	start := time.Now()

	servs.logger.Debug("SERVICE_GET_TEAM", "Getting team",
		"team_name", teamName)

	if teamName == "" {
		servs.logger.Warn("SERVICE_GET_TEAM", "Empty team name provided",
			"duration_ms", time.Since(start).Milliseconds())
		return nil, ErrEmptyTeamName
	}

	team, err := servs.repo.FindTeamByName(teamName)
	if err != nil {
		servs.logger.Error("SERVICE_GET_TEAM", "Failed to find team",
			"team_name", teamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, err
	}

	servs.logger.Info("SERVICE_GET_TEAM", "Team retrieved successfully",
		"team_name", teamName,
		"members_count", len(team.Members),
		"duration_ms", time.Since(start).Milliseconds())
	return team, nil
}

func (servs *PrService) SetUserActive(userID string, isActive bool) (*entity.User, error) {
	start := time.Now()

	servs.logger.Debug("SERVICE_SET_USER_ACTIVE", "Setting user active status",
		"user_id", userID,
		"is_active", isActive)

	if userID == "" {
		servs.logger.Warn("SERVICE_SET_USER_ACTIVE", "Empty user ID provided",
			"duration_ms", time.Since(start).Milliseconds())
		return nil, ErrEmptyUserID
	}

	user, err := servs.repo.FindUserByID(userID)
	if err != nil {
		servs.logger.Error("SERVICE_SET_USER_ACTIVE", "Failed to find user",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, err
	}

	if user.IsActive == isActive {
		servs.logger.Debug("SERVICE_SET_USER_ACTIVE", "User already has desired active status",
			"user_id", userID,
			"is_active", isActive,
			"duration_ms", time.Since(start).Milliseconds())
		return user, nil
	}

	err = servs.repo.SetActive(userID, isActive)
	if err != nil {
		servs.logger.Error("SERVICE_SET_USER_ACTIVE", "Failed to set user active status in repository",
			"user_id", userID,
			"is_active", isActive,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, err
	}

	user.IsActive = isActive

	servs.logger.Info("SERVICE_SET_USER_ACTIVE", "User active status updated successfully",
		"user_id", userID,
		"is_active", isActive,
		"team_name", user.TeamName,
		"duration_ms", time.Since(start).Milliseconds())
	return user, nil
}

func (servs *PrService) GetUserReviews(userID string) ([]*entity.PullRequest, error) {
	start := time.Now()

	servs.logger.Debug("SERVICE_GET_USER_REVIEWS", "Getting user reviews",
		"user_id", userID)

	if userID == "" {
		servs.logger.Warn("SERVICE_GET_USER_REVIEWS", "Empty user ID provided",
			"duration_ms", time.Since(start).Milliseconds())
		return nil, ErrEmptyUserID
	}

	if _, err := servs.repo.FindUserByID(userID); err != nil {
		servs.logger.Error("SERVICE_GET_USER_REVIEWS", "Failed to find user",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, err
	}

	prs, err := servs.repo.FindPRsByReviewer(userID)
	if err != nil {
		servs.logger.Error("SERVICE_GET_USER_REVIEWS", "Failed to get user reviews from repository",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, err
	}

	servs.logger.Info("SERVICE_GET_USER_REVIEWS", "User reviews retrieved successfully",
		"user_id", userID,
		"reviews_count", len(prs),
		"duration_ms", time.Since(start).Milliseconds())
	return prs, nil
}

func (servs *PrService) CreatePR(prID, prName, authorID string) (*entity.PullRequest, error) {
	start := time.Now()

	servs.logger.Debug("SERVICE_CREATE_PR", "Starting PR creation",
		"pr_id", prID,
		"pr_name", prName,
		"author_id", authorID)

	if err := checkPRCorrectness(prID, prName, authorID); err != nil {
		servs.logger.Warn("SERVICE_CREATE_PR", "PR data validation failed",
			"pr_id", prID,
			"pr_name", prName,
			"author_id", authorID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, err
	}

	// Проверяем что автор существует
	author, err := servs.repo.FindUserByID(authorID)
	if err != nil {
		servs.logger.Error("SERVICE_CREATE_PR", "Author not found",
			"author_id", authorID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("author not found: %w", err)
	}

	// Проверяем что PR не существует
	if _, err = servs.repo.FindPRByID(prID); err == nil {
		servs.logger.Warn("SERVICE_CREATE_PR", "PR already exists",
			"pr_id", prID,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, ErrPRAlreadyExists
	}

	// Назначаем ревьюверов
	candidates, err := servs.findReviewCandidates(author.TeamName, authorID)
	if err != nil {
		servs.logger.Error("SERVICE_CREATE_PR", "Failed to find review candidates",
			"team_name", author.TeamName,
			"author_id", authorID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find review candidates: %w", err)
	}

	reviewers := servs.selectReviewers(candidates, 2)

	servs.logger.Debug("SERVICE_CREATE_PR", "Reviewers selected",
		"pr_id", prID,
		"candidates_count", len(candidates),
		"reviewers_selected", reviewers)

	pr := entity.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := servs.repo.CreatePR(&pr); err != nil {
		servs.logger.Error("SERVICE_CREATE_PR", "Failed to create PR in repository",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, err
	}

	servs.logger.Info("SERVICE_CREATE_PR", "PR created successfully",
		"pr_id", prID,
		"pr_name", prName,
		"author_id", authorID,
		"reviewers_count", len(reviewers),
		"team_name", author.TeamName,
		"duration_ms", time.Since(start).Milliseconds())
	return &pr, nil
}

func (servs *PrService) MergePR(prID string) (*entity.PullRequest, error) {
	start := time.Now()

	servs.logger.Debug("SERVICE_MERGE_PR", "Starting PR merge",
		"pr_id", prID)

	if prID == "" {
		servs.logger.Warn("SERVICE_MERGE_PR", "Empty PR ID provided",
			"duration_ms", time.Since(start).Milliseconds())
		return nil, ErrEmptyPRID
	}

	pr, err := servs.repo.FindPRByID(prID)
	if err != nil {
		servs.logger.Error("SERVICE_MERGE_PR", "Failed to find PR",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find PR: %w", err)
	}

	if pr.Status == entity.PullRequestStatusMerged {
		servs.logger.Debug("SERVICE_MERGE_PR", "PR already merged",
			"pr_id", prID,
			"duration_ms", time.Since(start).Milliseconds())
		return pr, nil
	}

	pr.Status = entity.PullRequestStatusMerged
	if err := servs.repo.UpdatePR(pr); err != nil {
		servs.logger.Error("SERVICE_MERGE_PR", "Failed to update PR status in repository",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("update PR: %w", err)
	}

	servs.logger.Info("SERVICE_MERGE_PR", "PR merged successfully",
		"pr_id", prID,
		"pr_name", pr.PullRequestName,
		"author_id", pr.AuthorID,
		"reviewers_count", len(pr.AssignedReviewers),
		"duration_ms", time.Since(start).Milliseconds())
	return pr, nil
}

func (servs *PrService) ReassignReviewer(prID, oldUserID string) (*entity.PullRequest, string, error) {
	start := time.Now()

	servs.logger.Debug("SERVICE_REASSIGN_REVIEWER", "Starting reviewer reassignment",
		"pr_id", prID,
		"old_user_id", oldUserID)

	if prID == "" {
		servs.logger.Warn("SERVICE_REASSIGN_REVIEWER", "Empty PR ID provided",
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", ErrEmptyPRID
	}
	if oldUserID == "" {
		servs.logger.Warn("SERVICE_REASSIGN_REVIEWER", "Empty user ID provided",
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", ErrEmptyUserID
	}

	// Находим PR
	pr, err := servs.repo.FindPRByID(prID)
	if err != nil {
		servs.logger.Error("SERVICE_REASSIGN_REVIEWER", "Failed to find PR",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", fmt.Errorf("find PR: %w", err)
	}

	// Проверяем что PR не мерджен
	if pr.Status == entity.PullRequestStatusMerged {
		servs.logger.Warn("SERVICE_REASSIGN_REVIEWER", "Attempt to reassign on merged PR",
			"pr_id", prID,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", ErrCannotReassingOnMergedPR
	}

	user, err := servs.repo.FindUserByID(oldUserID)
	if err != nil {
		servs.logger.Error("SERVICE_REASSIGN_REVIEWER", "Failed to find user",
			"old_user_id", oldUserID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", fmt.Errorf("find user: %w", err)
	}

	candidates, err := servs.findReviewCandidates(user.TeamName, pr.AuthorID, oldUserID)
	if err != nil {
		servs.logger.Error("SERVICE_REASSIGN_REVIEWER", "Failed to find replacement candidates",
			"team_name", user.TeamName,
			"author_id", pr.AuthorID,
			"old_user_id", oldUserID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", fmt.Errorf("find replacement candidates: %w", err)
	}

	newReviewer := servs.selectReviewers(candidates, 1)
	if len(newReviewer) == 0 {
		servs.logger.Warn("SERVICE_REASSIGN_REVIEWER", "No replacement candidates available",
			"team_name", user.TeamName,
			"candidates_count", len(candidates),
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", ErrNoReplacementCandidate
	}

	found := false
	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			pr.AssignedReviewers[i] = newReviewer[0]
			found = true
			break
		}
	}
	if !found {
		servs.logger.Warn("SERVICE_REASSIGN_REVIEWER", "Reviewer not assigned to this PR",
			"pr_id", prID,
			"old_user_id", oldUserID,
			"assigned_reviewers", pr.AssignedReviewers,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", fmt.Errorf("reviewer %s not assigned to this PR", oldUserID)
	}

	// Сохраняем изменения
	if err := servs.repo.UpdatePR(pr); err != nil {
		servs.logger.Error("SERVICE_REASSIGN_REVIEWER", "Failed to update PR in repository",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", fmt.Errorf("update PR: %w", err)
	}

	servs.logger.Info("SERVICE_REASSIGN_REVIEWER", "Reviewer reassigned successfully",
		"pr_id", prID,
		"old_user_id", oldUserID,
		"new_user_id", newReviewer[0],
		"team_name", user.TeamName,
		"duration_ms", time.Since(start).Milliseconds())
	return pr, newReviewer[0], nil
}

// Вспомогательные функции (остаются без логирования)
func checkTeamCorrectness(team *entity.Team) error {
	if team.TeamName == "" {
		return ErrEmptyTeamName
	}
	if len(team.Members) == 0 {
		return ErrEmptyTeam
	}
	for _, member := range team.Members {
		if err := checkTeamMemberCorrectness(member); err != nil {
			return err
		}
	}
	return nil
}

func checkTeamMemberCorrectness(member entity.TeamMember) error {
	if member.UserID == "" {
		return ErrEmptyUserID
	}
	if member.Username == "" {
		return ErrEmptyUserUsername
	}
	return nil
}

func checkPRCorrectness(prID, prName, authorID string) error {
	switch {
	case prID == "":
		return ErrEmptyPRID
	case prName == "":
		return ErrEmptyPRName
	case authorID == "":
		return ErrEmptyPRAuthorID
	}
	return nil
}

func (servs *PrService) validateTeamCreation(team *entity.Team) error {
	if exists := servs.repo.TeamExists(team.TeamName); exists {
		return ErrTeamAlreadyExists
	}

	for _, member := range team.Members {
		if _, err := servs.repo.FindUserByID(member.UserID); err == nil {
			return ErrUserAlreadyExists(member.UserID)
		}
	}

	return nil
}

func (servs *PrService) findReviewCandidates(teamName string, excludeUserIDs ...string) ([]*entity.User, error) {
	if teamName == "" {
		return nil, ErrEmptyTeamName
	}

	// Получаем всех пользователей команды
	teamUsers, err := servs.repo.FindUsersByTeam(teamName)
	if err != nil {
		return nil, fmt.Errorf("find team users: %w", err)
	}

	// Создаём множество для быстрого исключения
	excludeSet := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeSet[id] = true
	}

	// Фильтруем кандидатов
	var candidates []*entity.User
	for _, user := range teamUsers {
		// Исключаем неактивных и пользователей из exclude списка
		if user.IsActive && !excludeSet[user.UserID] {
			candidates = append(candidates, user)
		}
	}

	return candidates, nil
}

func (servs *PrService) selectReviewers(candidates []*entity.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	// Если кандидатов меньше или равно maxCount - возвращаем всех
	if len(candidates) <= maxCount {
		userIDs := make([]string, len(candidates))
		for i, user := range candidates {
			userIDs[i] = user.UserID
		}
		return userIDs
	}

	// Используем random сервиса
	selected := make([]string, 0, maxCount)
	used := make(map[int]bool)

	for len(selected) < maxCount && len(used) < len(candidates) {
		idx := servs.random.Intn(len(candidates))
		if !used[idx] {
			selected = append(selected, candidates[idx].UserID)
			used[idx] = true
		}
	}

	return selected
}
