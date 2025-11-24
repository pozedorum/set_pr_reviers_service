package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

type PRRepository struct {
	db     *sql.DB
	logger interfaces.Logger
}

func NewPRRepository(dataSourceName string, logger interfaces.Logger) (*PRRepository, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("POSTGRES_REPO", "PostgreSQL repository initialized successfully")

	return &PRRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (r *PRRepository) Close() error {
	r.logger.Info("POSTGRES_REPO", "Closing database connection")
	return r.db.Close()
}

// Users

func (r *PRRepository) CreateUser(user *entity.User) error {
	start := time.Now()

	r.logger.Debug("POSTGRES_CREATE_USER", "Creating user",
		"user_id", user.UserID,
		"username", user.Username,
		"team_name", user.TeamName,
		"is_active", user.IsActive)

	query := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(query, user.UserID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		r.logger.Error("POSTGRES_CREATE_USER", "Failed to create user",
			"user_id", user.UserID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("create user: %w", err)
	}

	r.logger.Info("POSTGRES_CREATE_USER", "User created successfully",
		"user_id", user.UserID,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *PRRepository) FindUserByID(userID string) (*entity.User, error) {
	start := time.Now()

	r.logger.Debug("POSTGRES_FIND_USER_BY_ID", "Finding user by ID",
		"user_id", userID)

	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	var user entity.User
	err := r.db.QueryRow(query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("POSTGRES_FIND_USER_BY_ID", "User not found",
				"user_id", userID,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("user not found: %w", ErrNoUser)
		}
		r.logger.Error("POSTGRES_FIND_USER_BY_ID", "Failed to find user",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find user by ID: %w", err)
	}

	r.logger.Debug("POSTGRES_FIND_USER_BY_ID", "User found successfully",
		"user_id", userID,
		"duration_ms", time.Since(start).Milliseconds())
	return &user, nil
}

func (r *PRRepository) UpdateUser(user *entity.User) error {
	start := time.Now()

	r.logger.Debug("POSTGRES_UPDATE_USER", "Updating user",
		"user_id", user.UserID,
		"username", user.Username,
		"team_name", user.TeamName,
		"is_active", user.IsActive)

	query := `
		UPDATE users 
		SET username = $1, team_name = $2, is_active = $3, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $4
	`

	result, err := r.db.Exec(query, user.Username, user.TeamName, user.IsActive, user.UserID)
	if err != nil {
		r.logger.Error("POSTGRES_UPDATE_USER", "Failed to update user",
			"user_id", user.UserID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("update user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("POSTGRES_UPDATE_USER", "No user found to update",
			"user_id", user.UserID,
			"duration_ms", time.Since(start).Milliseconds())
		return ErrNoUser
	}

	r.logger.Info("POSTGRES_UPDATE_USER", "User updated successfully",
		"user_id", user.UserID,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *PRRepository) FindUsersByTeam(teamName string) ([]*entity.User, error) {
	start := time.Now()

	r.logger.Debug("POSTGRES_FIND_USERS_BY_TEAM", "Finding users by team",
		"team_name", teamName)

	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`

	rows, err := r.db.Query(query, teamName)
	if err != nil {
		r.logger.Error("POSTGRES_FIND_USERS_BY_TEAM", "Failed to query users by team",
			"team_name", teamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find users by team: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
		); err != nil {
			r.logger.Error("POSTGRES_FIND_USERS_BY_TEAM", "Failed to scan user row",
				"team_name", teamName,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("POSTGRES_FIND_USERS_BY_TEAM", "Error iterating user rows",
			"team_name", teamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}

	r.logger.Debug("POSTGRES_FIND_USERS_BY_TEAM", "Users found successfully",
		"team_name", teamName,
		"users_count", len(users),
		"duration_ms", time.Since(start).Milliseconds())
	return users, nil
}

func (r *PRRepository) SetActive(userID string, isActive bool) error {
	start := time.Now()

	r.logger.Debug("POSTGRES_SET_ACTIVE", "Setting user active status",
		"user_id", userID,
		"is_active", isActive)

	query := `
		UPDATE users 
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`

	result, err := r.db.Exec(query, isActive, userID)
	if err != nil {
		r.logger.Error("POSTGRES_SET_ACTIVE", "Failed to set user active status",
			"user_id", userID,
			"is_active", isActive,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("set user active: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("POSTGRES_SET_ACTIVE", "No user found to update active status",
			"user_id", userID,
			"duration_ms", time.Since(start).Milliseconds())
		return sql.ErrNoRows // Возвращаем стандартную ошибку
	}

	r.logger.Info("POSTGRES_SET_ACTIVE", "User active status updated successfully",
		"user_id", userID,
		"is_active", isActive,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

// Teams

func (r *PRRepository) CreateTeam(team *entity.Team) error {
	start := time.Now()

	r.logger.Debug("POSTGRES_CREATE_TEAM", "Creating team",
		"team_name", team.TeamName,
		"members_count", len(team.Members))

	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("POSTGRES_CREATE_TEAM", "Failed to begin transaction",
			"team_name", team.TeamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создаем команду
	teamQuery := `INSERT INTO teams (team_name) VALUES ($1)`
	if _, err := tx.Exec(teamQuery, team.TeamName); err != nil {
		r.logger.Error("POSTGRES_CREATE_TEAM", "Failed to create team",
			"team_name", team.TeamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("create team: %w", err)
	}

	// Создаем пользователей
	userQuery := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`
	for _, member := range team.Members {
		if _, err := tx.Exec(userQuery, member.UserID, member.Username, team.TeamName, member.IsActive); err != nil {
			r.logger.Error("POSTGRES_CREATE_TEAM", "Failed to create team member",
				"team_name", team.TeamName,
				"user_id", member.UserID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return fmt.Errorf("create team member %s: %w", member.UserID, err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		r.logger.Error("POSTGRES_CREATE_TEAM", "Failed to commit transaction",
			"team_name", team.TeamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("commit transaction: %w", err)
	}

	r.logger.Info("POSTGRES_CREATE_TEAM", "Team created successfully",
		"team_name", team.TeamName,
		"members_count", len(team.Members),
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *PRRepository) FindTeamByName(teamName string) (*entity.Team, error) {
	start := time.Now()

	r.logger.Debug("POSTGRES_FIND_TEAM_BY_NAME", "Finding team by name",
		"team_name", teamName)

	// Получаем основную информацию о команде
	teamQuery := `SELECT team_name FROM teams WHERE team_name = $1`
	var team entity.Team
	err := r.db.QueryRow(teamQuery, teamName).Scan(&team.TeamName)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("POSTGRES_FIND_TEAM_BY_NAME", "Team not found",
				"team_name", teamName,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, ErrNoTeam
		}
		r.logger.Error("POSTGRES_FIND_TEAM_BY_NAME", "Failed to find team",
			"team_name", teamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find team by name: %w", err)
	}

	// Получаем участников команды
	membersQuery := `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`

	rows, err := r.db.Query(membersQuery, teamName)
	if err != nil {
		r.logger.Error("POSTGRES_FIND_TEAM_BY_NAME", "Failed to query team members",
			"team_name", teamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("query team members: %w", err)
	}
	defer rows.Close()

	var members []entity.TeamMember
	for rows.Next() {
		var member entity.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			r.logger.Error("POSTGRES_FIND_TEAM_BY_NAME", "Failed to scan team member row",
				"team_name", teamName,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("scan team member row: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("POSTGRES_FIND_TEAM_BY_NAME", "Error iterating team member rows",
			"team_name", teamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("iterate team member rows: %w", err)
	}

	team.Members = members

	r.logger.Debug("POSTGRES_FIND_TEAM_BY_NAME", "Team found successfully",
		"team_name", teamName,
		"members_count", len(members),
		"duration_ms", time.Since(start).Milliseconds())
	return &team, nil
}

func (r *PRRepository) TeamExists(teamName string) bool {
	start := time.Now()

	r.logger.Debug("POSTGRES_TEAM_EXISTS", "Checking if team exists",
		"team_name", teamName)

	query := `SELECT 1 FROM teams WHERE team_name = $1`
	var exists bool
	err := r.db.QueryRow(query, teamName).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		r.logger.Error("POSTGRES_TEAM_EXISTS", "Failed to check team existence",
			"team_name", teamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return false
	}

	exists = err == nil

	r.logger.Debug("POSTGRES_TEAM_EXISTS", "Team existence check completed",
		"team_name", teamName,
		"exists", exists,
		"duration_ms", time.Since(start).Milliseconds())
	return exists
}

// PRs

func (r *PRRepository) CreatePR(pr *entity.PullRequest) error {
	start := time.Now()

	r.logger.Debug("POSTGRES_CREATE_PR", "Creating pull request",
		"pr_id", pr.PullRequestID,
		"pr_name", pr.PullRequestName,
		"author_id", pr.AuthorID,
		"reviewers_count", len(pr.AssignedReviewers))

	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("POSTGRES_CREATE_PR", "Failed to begin transaction",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создаем PR
	prQuery := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	var createdAt time.Time
	if pr.CreatedAt != nil {
		createdAt = *pr.CreatedAt
	} else {
		createdAt = time.Now()
	}

	_, err = tx.Exec(prQuery,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		string(pr.Status),
		createdAt,
	)
	if err != nil {
		r.logger.Error("POSTGRES_CREATE_PR", "Failed to create pull request",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("create pull request: %w", err)
	}

	// Добавляем ревьюверов
	reviewerQuery := `
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2)
	`

	for _, reviewerID := range pr.AssignedReviewers {
		_, err := tx.Exec(reviewerQuery, pr.PullRequestID, reviewerID)
		if err != nil {
			r.logger.Error("POSTGRES_CREATE_PR", "Failed to add reviewer to PR",
				"pr_id", pr.PullRequestID,
				"reviewer_id", reviewerID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return fmt.Errorf("add reviewer %s to PR: %w", reviewerID, err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		r.logger.Error("POSTGRES_CREATE_PR", "Failed to commit transaction",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("commit transaction: %w", err)
	}

	r.logger.Info("POSTGRES_CREATE_PR", "Pull request created successfully",
		"pr_id", pr.PullRequestID,
		"reviewers_count", len(pr.AssignedReviewers),
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *PRRepository) FindPRByID(prID string) (*entity.PullRequest, error) {
	start := time.Now()

	r.logger.Debug("POSTGRES_FIND_PR_BY_ID", "Finding pull request by ID",
		"pr_id", prID)

	// Получаем основную информацию о PR
	prQuery := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	var pr entity.PullRequest
	var status string
	var mergedAt sql.NullTime

	err := r.db.QueryRow(prQuery, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&status,
		&pr.CreatedAt,
		&mergedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("POSTGRES_FIND_PR_BY_ID", "Pull request not found",
				"pr_id", prID,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, ErrNoPR
		}
		r.logger.Error("POSTGRES_FIND_PR_BY_ID", "Failed to find pull request",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find PR by ID: %w", err)
	}

	pr.Status = entity.PullRequestStatus(status)
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	// Получаем ревьюверов
	reviewersQuery := `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
		ORDER BY reviewer_id
	`

	rows, err := r.db.Query(reviewersQuery, prID)
	if err != nil {
		r.logger.Error("POSTGRES_FIND_PR_BY_ID", "Failed to query PR reviewers",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("query PR reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			r.logger.Error("POSTGRES_FIND_PR_BY_ID", "Failed to scan reviewer row",
				"pr_id", prID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("scan reviewer row: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("POSTGRES_FIND_PR_BY_ID", "Error iterating reviewer rows",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("iterate reviewer rows: %w", err)
	}

	pr.AssignedReviewers = reviewers

	r.logger.Debug("POSTGRES_FIND_PR_BY_ID", "Pull request found successfully",
		"pr_id", prID,
		"reviewers_count", len(reviewers),
		"duration_ms", time.Since(start).Milliseconds())
	return &pr, nil
}

func (r *PRRepository) ClosePR(pr *entity.PullRequest) error {
	start := time.Now()

	r.logger.Debug("POSTGRES_CLOSE_PR", "Closing pull request",
		"pr_id", pr.PullRequestID)

	query := `
		UPDATE pull_requests 
		SET status = $1, merged_at = $2
		WHERE pull_request_id = $3
	`

	var mergedAt interface{}
	if pr.MergedAt != nil {
		mergedAt = *pr.MergedAt
	} else {
		mergedAt = nil
	}

	result, err := r.db.Exec(query, string(entity.PullRequestStatusMerged), mergedAt, pr.PullRequestID)
	if err != nil {
		r.logger.Error("POSTGRES_CLOSE_PR", "Failed to close pull request",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("close PR: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("POSTGRES_CLOSE_PR", "No pull request found to close",
			"pr_id", pr.PullRequestID,
			"duration_ms", time.Since(start).Milliseconds())
		return ErrNoPR
	}

	r.logger.Info("POSTGRES_CLOSE_PR", "Pull request closed successfully",
		"pr_id", pr.PullRequestID,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *PRRepository) UpdatePR(pr *entity.PullRequest) error {
	start := time.Now()

	r.logger.Debug("POSTGRES_UPDATE_PR", "Updating pull request",
		"pr_id", pr.PullRequestID,
		"status", pr.Status,
		"reviewers_count", len(pr.AssignedReviewers))

	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("POSTGRES_UPDATE_PR", "Failed to begin transaction",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Обновляем основную информацию о PR
	prQuery := `
		UPDATE pull_requests 
		SET pull_request_name = $1, status = $2, merged_at = $3
		WHERE pull_request_id = $4
	`

	var mergedAt interface{}
	if pr.MergedAt != nil {
		mergedAt = *pr.MergedAt
	} else {
		mergedAt = nil
	}

	result, err := tx.Exec(prQuery, pr.PullRequestName, string(pr.Status), mergedAt, pr.PullRequestID)
	if err != nil {
		r.logger.Error("POSTGRES_UPDATE_PR", "Failed to update pull request",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("update PR: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("POSTGRES_UPDATE_PR", "No pull request found to update",
			"pr_id", pr.PullRequestID,
			"duration_ms", time.Since(start).Milliseconds())
		return ErrNoPR
	}

	// Обновляем ревьюверов: удаляем старых и добавляем новых
	deleteReviewersQuery := `DELETE FROM pull_request_reviewers WHERE pull_request_id = $1`
	if _, err := tx.Exec(deleteReviewersQuery, pr.PullRequestID); err != nil {
		r.logger.Error("POSTGRES_UPDATE_PR", "Failed to delete old reviewers",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("delete old reviewers: %w", err)
	}

	insertReviewerQuery := `
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2)
	`

	for _, reviewerID := range pr.AssignedReviewers {
		if _, err := tx.Exec(insertReviewerQuery, pr.PullRequestID, reviewerID); err != nil {
			r.logger.Error("POSTGRES_UPDATE_PR", "Failed to add reviewer to PR",
				"pr_id", pr.PullRequestID,
				"reviewer_id", reviewerID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return fmt.Errorf("add reviewer %s to PR: %w", reviewerID, err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		r.logger.Error("POSTGRES_UPDATE_PR", "Failed to commit transaction",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("commit transaction: %w", err)
	}

	r.logger.Info("POSTGRES_UPDATE_PR", "Pull request updated successfully",
		"pr_id", pr.PullRequestID,
		"reviewers_count", len(pr.AssignedReviewers),
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *PRRepository) FindPRsByReviewer(userID string) ([]*entity.PullRequest, error) {
	start := time.Now()

	r.logger.Debug("POSTGRES_FIND_PRS_BY_REVIEWER", "Finding PRs by reviewer",
		"user_id", userID)

	query := `
		SELECT 
			pr.pull_request_id, 
			pr.pull_request_name, 
			pr.author_id, 
			pr.status, 
			pr.created_at, 
			pr.merged_at
		FROM pull_requests pr
		INNER JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		r.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "Failed to query PRs by reviewer",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		var status string
		var mergedAt sql.NullTime

		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&status,
			&pr.CreatedAt,
			&mergedAt,
		); err != nil {
			r.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "Failed to scan PR row",
				"user_id", userID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("scan PR row: %w", err)
		}

		pr.Status = entity.PullRequestStatus(status)
		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}

		// Получаем ревьюверов для каждого PR
		reviewersQuery := `
			SELECT reviewer_id
			FROM pull_request_reviewers
			WHERE pull_request_id = $1
		`
		reviewerRows, err := r.db.Query(reviewersQuery, pr.PullRequestID)
		if err != nil {
			r.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "Failed to query reviewers for PR",
				"pr_id", pr.PullRequestID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("query reviewers for PR %s: %w", pr.PullRequestID, err)
		}

		var reviewers []string
		for reviewerRows.Next() {
			var reviewerID string
			if err := reviewerRows.Scan(&reviewerID); err != nil {
				reviewerRows.Close()
				r.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "Failed to scan reviewer for PR",
					"pr_id", pr.PullRequestID,
					"error", err,
					"duration_ms", time.Since(start).Milliseconds())
				return nil, fmt.Errorf("scan reviewer for PR %s: %w", pr.PullRequestID, err)
			}
			reviewers = append(reviewers, reviewerID)
		}
		reviewerRows.Close()

		pr.AssignedReviewers = reviewers
		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "Error iterating PR rows",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("iterate PR rows: %w", err)
	}

	r.logger.Debug("POSTGRES_FIND_PRS_BY_REVIEWER", "PRs found successfully",
		"user_id", userID,
		"prs_count", len(prs),
		"duration_ms", time.Since(start).Milliseconds())
	return prs, nil
}
