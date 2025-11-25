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

func (servs *PrService) CreatePR(pr *entity.PullRequest) error {
	start := time.Now()

	if err := checkPRCorrectness(pr); err != nil {
		servs.logger.Warn("SERVICE_CREATE_PR", "PR data validation failed",
			"pr_id", pr.PullRequestID,
			"pr_name", pr.PullRequestName,
			"author_id", pr.AuthorID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return err
	}

	servs.logger.Debug("SERVICE_CREATE_PR", "Starting PR creation",
		"pr_id", pr.PullRequestID,
		"pr_name", pr.PullRequestName,
		"author_id", pr.AuthorID)

	// Проверяем что автор существует
	author, err := servs.repo.FindUserByID(pr.AuthorID)
	if err != nil {
		servs.logger.Error("SERVICE_CREATE_PR", "Author not found",
			"author_id", pr.AuthorID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("author not found: %w", err)
	}

	// Проверяем что PR не существует
	if existingPR, err := servs.repo.FindPRByID(pr.PullRequestID); err == nil && existingPR != nil {
		servs.logger.Warn("SERVICE_CREATE_PR", "PR already exists",
			"pr_id", pr.PullRequestID,
			"duration_ms", time.Since(start).Milliseconds())
		return ErrPRAlreadyExists
	}

	// Назначаем ревьюверов
	candidates, err := servs.findReviewCandidates(author.TeamName, pr.AuthorID)
	if err != nil {
		servs.logger.Error("SERVICE_CREATE_PR", "Failed to find review candidates",
			"team_name", author.TeamName,
			"author_id", pr.AuthorID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("find review candidates: %w", err)
	}

	reviewers := servs.selectReviewers(candidates, 2)

	servs.logger.Debug("SERVICE_CREATE_PR", "Reviewers selected",
		"pr_id", pr.PullRequestID,
		"candidates_count", len(candidates),
		"reviewers_selected", reviewers)

	pr.Status = entity.PullRequestStatusOpen
	pr.AssignedReviewers = reviewers
	pr.CreatedAt = time.Now() // Добавляем timestamp

	if err := servs.repo.CreatePR(pr); err != nil {
		servs.logger.Error("SERVICE_CREATE_PR", "Failed to create PR in repository",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return err
	}

	servs.logger.Info("SERVICE_CREATE_PR", "PR created successfully",
		"pr_id", pr.PullRequestID,
		"pr_name", pr.PullRequestName,
		"author_id", pr.AuthorID,
		"reviewers_count", len(reviewers),
		"team_name", author.TeamName,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
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
	pr.MergedAt = start
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

	// Проверяем что старый ревьювер существует и получаем его команду
	oldUser, err := servs.repo.FindUserByID(oldUserID)
	if err != nil {
		servs.logger.Error("SERVICE_REASSIGN_REVIEWER", "Failed to find old user",
			"old_user_id", oldUserID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", fmt.Errorf("find old user: %w", err)
	}

	// Проверяем что старый ревьювер действительно назначен на этот PR
	if !servs.containsReviewer(pr.AssignedReviewers, oldUserID) {
		servs.logger.Warn("SERVICE_REASSIGN_REVIEWER", "Reviewer not assigned to this PR",
			"pr_id", prID,
			"old_user_id", oldUserID,
			"assigned_reviewers", pr.AssignedReviewers,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", fmt.Errorf("reviewer %s not assigned to this PR", oldUserID)
	}

	// Ищем кандидатов для замены из команды старого ревьювера
	excludeUsers := []string{pr.AuthorID}
	excludeUsers = append(excludeUsers, pr.AssignedReviewers...) // исключаем уже назначенных
	var flag bool = false
	for _, usersId := range excludeUsers {
		if usersId == oldUserID {
			flag = true
			break
		}
	}
	if !flag {
		excludeUsers = append(excludeUsers, oldUserID)
	}

	candidates, err := servs.findReviewCandidates(oldUser.TeamName, excludeUsers...)
	if err != nil {
		servs.logger.Error("SERVICE_REASSIGN_REVIEWER", "Failed to find replacement candidates",
			"team_name", oldUser.TeamName,
			"excluded_users", excludeUsers,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", fmt.Errorf("find replacement candidates: %w", err)
	}

	// Выбираем одного случайного кандидата
	newReviewers := servs.selectReviewers(candidates, 1)
	if len(newReviewers) == 0 {
		servs.logger.Warn("SERVICE_REASSIGN_REVIEWER", "No replacement candidates available",
			"team_name", oldUser.TeamName,
			"candidates_count", len(candidates),
			"duration_ms", time.Since(start).Milliseconds())
		return nil, "", ErrNoReplacementCandidate
	}

	newReviewerID := newReviewers[0]

	// Заменяем ревьювера в списке
	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			pr.AssignedReviewers[i] = newReviewerID
			break
		}
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
		"new_user_id", newReviewerID,
		"team_name", oldUser.TeamName,
		"duration_ms", time.Since(start).Milliseconds())
	return pr, newReviewerID, nil
}

// Вспомогательная функция для проверки наличия ревьювера
func (servs *PrService) containsReviewer(reviewers []string, userID string) bool {
	for _, reviewer := range reviewers {
		if reviewer == userID {
			return true
		}
	}
	return false
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

func checkPRCorrectness(pr *entity.PullRequest) error {
	switch {
	case pr == nil:
		return ErrNilPR
	case pr.PullRequestID == "":
		return ErrEmptyPRID
	case pr.PullRequestName == "":
		return ErrEmptyPRName
	case pr.AuthorID == "":
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
