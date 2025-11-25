package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

var (
	ErrNoTeam = errors.New("no such team")
	ErrNoUser = errors.New("no such user")
	ErrNoPR   = errors.New("no such pull request")
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

func (repo *PRRepository) Close() error {
	repo.logger.Info("POSTGRES_REPO", "Closing database connection")
	return repo.db.Close()
}

// Users

func (repo *PRRepository) CreateUser(user *entity.User) error {
	start := time.Now()

	repo.logger.Debug("POSTGRES_CREATE_USER", "Creating user",
		"user_id", user.UserID,
		"username", user.Username,
		"team_name", user.TeamName,
		"is_active", user.IsActive)

	// Получаем team_id по team_name
	teamID, err := repo.getTeamIDByName(user.TeamName)
	if err != nil {
		repo.logger.Error("POSTGRES_CREATE_USER", "Failed to get team ID",
			"user_id", user.UserID,
			"team_name", user.TeamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("get team ID: %w", err)
	}

	query := `
		INSERT INTO users (user_id, username, team_id, is_active)
		VALUES ($1, $2, $3, $4)
	`

	_, err = repo.db.Exec(query, user.UserID, user.Username, teamID, user.IsActive)
	if err != nil {
		repo.logger.Error("POSTGRES_CREATE_USER", "Failed to create user",
			"user_id", user.UserID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("create user: %w", err)
	}

	repo.logger.Info("POSTGRES_CREATE_USER", "User created successfully",
		"user_id", user.UserID,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (repo *PRRepository) FindUserByID(userID string) (*entity.User, error) {
	start := time.Now()

	repo.logger.Debug("POSTGRES_FIND_USER_BY_ID", "Finding user by ID",
		"user_id", userID)

	query := `
		SELECT u.user_id, u.username, t.team_name, u.is_active
		FROM users u
		JOIN teams t ON u.team_id = t.team_id
		WHERE u.user_id = $1
	`

	var user entity.User
	err := repo.db.QueryRow(query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			repo.logger.Debug("POSTGRES_FIND_USER_BY_ID", "User not found",
				"user_id", userID,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("user not found: %w", ErrNoUser)
		}
		repo.logger.Error("POSTGRES_FIND_USER_BY_ID", "Failed to find user",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find user by ID: %w", err)
	}

	repo.logger.Debug("POSTGRES_FIND_USER_BY_ID", "User found successfully",
		"user_id", userID,
		"duration_ms", time.Since(start).Milliseconds())
	return &user, nil
}

func (repo *PRRepository) UpdateUser(user *entity.User) error {
	start := time.Now()

	repo.logger.Debug("POSTGRES_UPDATE_USER", "Updating user",
		"user_id", user.UserID,
		"username", user.Username,
		"team_name", user.TeamName,
		"is_active", user.IsActive)

	// Получаем team_id по team_name
	teamID, err := repo.getTeamIDByName(user.TeamName)
	if err != nil {
		repo.logger.Error("POSTGRES_UPDATE_USER", "Failed to get team ID",
			"user_id", user.UserID,
			"team_name", user.TeamName,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("get team ID: %w", err)
	}

	query := `
		UPDATE users 
		SET username = $1, team_id = $2, is_active = $3, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $4
	`

	result, err := repo.db.Exec(query, user.Username, teamID, user.IsActive, user.UserID)
	if err != nil {
		repo.logger.Error("POSTGRES_UPDATE_USER", "Failed to update user",
			"user_id", user.UserID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("update user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		repo.logger.Warn("POSTGRES_UPDATE_USER", "No user found to update",
			"user_id", user.UserID,
			"duration_ms", time.Since(start).Milliseconds())
		return ErrNoUser
	}

	repo.logger.Info("POSTGRES_UPDATE_USER", "User updated successfully",
		"user_id", user.UserID,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (repo *PRRepository) FindUsersByTeam(teamName string) ([]*entity.User, error) {
	start := time.Now()

	repo.logger.Debug("POSTGRES_FIND_USERS_BY_TEAM", "Finding users by team", "team_name", teamName)

	// Получаем team_id
	teamID, err := repo.getTeamIDByName(teamName)
	if err != nil {
		return nil, fmt.Errorf("find users by team: %w", err)
	}

	query := `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_id = $1
		ORDER BY user_id
	`

	rows, err := repo.db.Query(query, teamID)
	if err != nil {
		return nil, fmt.Errorf("query users by team: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			repo.logger.Error("POSTGRES_FIND_USERS_BY_TEAM", "failed to close sql rows", "error", err)
		}
	}()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.IsActive); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		user.TeamName = teamName // Заполняем team_name для API
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}

	repo.logger.Debug("POSTGRES_FIND_USERS_BY_TEAM", "Users found successfully",
		"team_name", teamName, "team_id", teamID, "users_count", len(users),
		"duration_ms", time.Since(start).Milliseconds())
	return users, nil
}

func (repo *PRRepository) SetActive(userID string, isActive bool) error {
	start := time.Now()

	repo.logger.Debug("POSTGRES_SET_ACTIVE", "Setting user active status",
		"user_id", userID,
		"is_active", isActive)

	query := `
		UPDATE users 
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`

	result, err := repo.db.Exec(query, isActive, userID)
	if err != nil {
		repo.logger.Error("POSTGRES_SET_ACTIVE", "Failed to set user active status",
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
		repo.logger.Warn("POSTGRES_SET_ACTIVE", "No user found to update active status",
			"user_id", userID,
			"duration_ms", time.Since(start).Milliseconds())
		return ErrNoUser
	}

	repo.logger.Info("POSTGRES_SET_ACTIVE", "User active status updated successfully",
		"user_id", userID,
		"is_active", isActive,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

// Teams

func (repo *PRRepository) CreateTeam(team *entity.Team) error {
	repo.logger.Debug("POSTGRES_CREATE_TEAM", "Creating team",
		"team_name", team.TeamName,
		"members_count", len(team.Members))

	// Начинаем транзакцию
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			repo.logger.Error("POSTGRES_CREATE_TEAM", "failed to rollback transaction", "error", err)
		}
	}()

	// Создаем команду и получаем team_id
	var teamID int
	teamQuery := `INSERT INTO teams (team_name) VALUES ($1) RETURNING team_id`
	err = tx.QueryRow(teamQuery, team.TeamName).Scan(&teamID)
	if err != nil {
		repo.logger.Error("POSTGRES_CREATE_TEAM", "Failed to create team",
			"team_name", team.TeamName, "error", err)
		return fmt.Errorf("create team: %w", err)
	}
	// Создаем пользователей
	userQuery := `
		INSERT INTO users (user_id, username, team_id, is_active)
		VALUES ($1, $2, $3, $4)
	`
	for _, member := range team.Members {
		_, err := tx.Exec(userQuery, member.UserID, member.Username, teamID, member.IsActive)
		if err != nil {
			repo.logger.Error("POSTGRES_CREATE_TEAM", "Failed to create team member",
				"team_name", team.TeamName, "user_id", member.UserID, "error", err)
			return fmt.Errorf("create team member %s: %w", member.UserID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		repo.logger.Error("POSTGRES_CREATE_TEAM", "Failed to commit transaction",
			"team_name", team.TeamName, "error", err)
		return fmt.Errorf("commit transaction: %w", err)
	}

	repo.logger.Info("POSTGRES_CREATE_TEAM", "Team created successfully",
		"team_name", team.TeamName, "team_id", teamID)
	return nil
}

func (repo *PRRepository) FindTeamByName(teamName string) (*entity.Team, error) {
	repo.logger.Debug("POSTGRES_FIND_TEAM_BY_NAME", "Finding team by name", "team_name", teamName)

	// Получаем team_id по team_name
	var teamID int
	teamQuery := `SELECT team_id FROM teams WHERE team_name = $1`
	err := repo.db.QueryRow(teamQuery, teamName).Scan(&teamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoTeam
		}
		return nil, fmt.Errorf("find team by name: %w", err)
	}

	// Получаем участников команды по team_id
	membersQuery := `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_id = $1
		ORDER BY user_id
	`

	rows, err := repo.db.Query(membersQuery, teamID)
	if err != nil {
		return nil, fmt.Errorf("query team members: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			repo.logger.Error("POSTGRES_FIND_TEAM_BY_NAME", "failed to close sql rows", "error", err)
		}
	}()

	var members []entity.TeamMember
	for rows.Next() {
		var member entity.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("scan team member row: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team member rows: %w", err)
	}

	team := &entity.Team{
		TeamName: teamName, // Возвращаем только team_name, teamID скрыт
		Members:  members,
	}

	repo.logger.Debug("POSTGRES_FIND_TEAM_BY_NAME", "Team found successfully",
		"team_name", teamName, "team_id", teamID, "members_count", len(members))
	return team, nil
}

func (repo *PRRepository) TeamExists(teamName string) bool {
	repo.logger.Debug("POSTGRES_TEAM_EXISTS", "Checking if team exists", "team_name", teamName)

	query := `SELECT 1 FROM teams WHERE team_name = $1`
	var exists bool
	err := repo.db.QueryRow(query, teamName).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		repo.logger.Error("POSTGRES_TEAM_EXISTS", "Failed to check team existence",
			"team_name", teamName, "error", err)
		return false
	}

	exists = err == nil

	repo.logger.Debug("POSTGRES_TEAM_EXISTS", "Team existence check completed",
		"team_name", teamName, "exists", exists)
	return exists
}

// PRs

func (repo *PRRepository) CreatePR(pr *entity.PullRequest) error {
	start := time.Now()

	repo.logger.Debug("POSTGRES_CREATE_PR", "Creating pull request",
		"pr_id", pr.PullRequestID,
		"pr_name", pr.PullRequestName,
		"author_id", pr.AuthorID,
		"reviewers_count", len(pr.AssignedReviewers))

	// Начинаем транзакцию
	tx, err := repo.db.Begin()
	if err != nil {
		repo.logger.Error("POSTGRES_CREATE_PR", "Failed to begin transaction",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			repo.logger.Error("POSTGRES_CREATE_PR", "failed to rollback transaction", "error", err)
		}
	}()

	// Создаем PR
	prQuery := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
		VALUES ($1, $2, $3, $4)
	`

	_, err = tx.Exec(prQuery,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		string(pr.Status),
	)
	if err != nil {
		repo.logger.Error("POSTGRES_CREATE_PR", "Failed to create pull request",
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
			repo.logger.Error("POSTGRES_CREATE_PR", "Failed to add reviewer to PR",
				"pr_id", pr.PullRequestID,
				"reviewer_id", reviewerID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return fmt.Errorf("add reviewer %s to PR: %w", reviewerID, err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		repo.logger.Error("POSTGRES_CREATE_PR", "Failed to commit transaction",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("commit transaction: %w", err)
	}

	repo.logger.Info("POSTGRES_CREATE_PR", "Pull request created successfully",
		"pr_id", pr.PullRequestID,
		"reviewers_count", len(pr.AssignedReviewers),
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (repo *PRRepository) FindPRByID(prID string) (*entity.PullRequest, error) {
	start := time.Now()

	repo.logger.Debug("POSTGRES_FIND_PR_BY_ID", "Finding pull request by ID",
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

	err := repo.db.QueryRow(prQuery, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&status,
		&pr.CreatedAt,
		&mergedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			repo.logger.Debug("POSTGRES_FIND_PR_BY_ID", "Pull request not found",
				"pr_id", prID,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, ErrNoPR
		}
		repo.logger.Error("POSTGRES_FIND_PR_BY_ID", "Failed to find pull request",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find PR by ID: %w", err)
	}

	pr.Status = entity.PullRequestStatus(status)
	if mergedAt.Valid {
		pr.MergedAt = mergedAt.Time
	}

	// Получаем ревьюверов
	reviewersQuery := `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
		ORDER BY reviewer_id
	`

	rows, err := repo.db.Query(reviewersQuery, prID)
	if err != nil {
		repo.logger.Error("POSTGRES_FIND_PR_BY_ID", "Failed to query PR reviewers",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("query PR reviewers: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			repo.logger.Error("POSTGRES_FIND_PR_BY_ID", "failed to close sql rows", "error", err)
		}
	}()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			repo.logger.Error("POSTGRES_FIND_PR_BY_ID", "Failed to scan reviewer row",
				"pr_id", prID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("scan reviewer row: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		repo.logger.Error("POSTGRES_FIND_PR_BY_ID", "Error iterating reviewer rows",
			"pr_id", prID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("iterate reviewer rows: %w", err)
	}

	pr.AssignedReviewers = reviewers

	repo.logger.Debug("POSTGRES_FIND_PR_BY_ID", "Pull request found successfully",
		"pr_id", prID,
		"reviewers_count", len(reviewers),
		"duration_ms", time.Since(start).Milliseconds())
	return &pr, nil
}

func (repo *PRRepository) UpdatePR(pr *entity.PullRequest) error {
	start := time.Now()

	repo.logger.Debug("POSTGRES_UPDATE_PR", "Updating pull request",
		"pr_id", pr.PullRequestID,
		"status", pr.Status,
		"reviewers_count", len(pr.AssignedReviewers))

	// Начинаем транзакцию
	tx, err := repo.db.Begin()
	if err != nil {
		repo.logger.Error("POSTGRES_UPDATE_PR", "Failed to begin transaction",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			repo.logger.Error("POSTGRES_UPDATE_PR", "failed to rollback transaction", "error", err)
		}
	}()

	// Обновляем основную информацию о PR
	prQuery := `
		UPDATE pull_requests 
		SET pull_request_name = $1, status = $2, merged_at = $3
		WHERE pull_request_id = $4
	`

	result, err := tx.Exec(prQuery, pr.PullRequestName, string(pr.Status), pr.MergedAt, pr.PullRequestID)
	if err != nil {
		repo.logger.Error("POSTGRES_UPDATE_PR", "Failed to update pull request",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("update PR: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		repo.logger.Warn("POSTGRES_UPDATE_PR", "No pull request found to update",
			"pr_id", pr.PullRequestID,
			"duration_ms", time.Since(start).Milliseconds())
		return ErrNoPR
	}

	// Обновляем ревьюверов: удаляем старых и добавляем новых
	deleteReviewersQuery := `DELETE FROM pull_request_reviewers WHERE pull_request_id = $1`
	if _, err := tx.Exec(deleteReviewersQuery, pr.PullRequestID); err != nil {
		repo.logger.Error("POSTGRES_UPDATE_PR", "Failed to delete old reviewers",
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
			repo.logger.Error("POSTGRES_UPDATE_PR", "Failed to add reviewer to PR",
				"pr_id", pr.PullRequestID,
				"reviewer_id", reviewerID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return fmt.Errorf("add reviewer %s to PR: %w", reviewerID, err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		repo.logger.Error("POSTGRES_UPDATE_PR", "Failed to commit transaction",
			"pr_id", pr.PullRequestID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("commit transaction: %w", err)
	}

	repo.logger.Info("POSTGRES_UPDATE_PR", "Pull request updated successfully",
		"pr_id", pr.PullRequestID,
		"reviewers_count", len(pr.AssignedReviewers),
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (repo *PRRepository) FindPRsByReviewer(userID string) ([]*entity.PullRequest, error) {
	start := time.Now()

	repo.logger.Debug("POSTGRES_FIND_PRS_BY_REVIEWER", "Finding PRs by reviewer",
		"user_id", userID)

	// ОДИН запрос вместо N+1
	query := `
		WITH prs_with_reviewers AS (
			SELECT 
				pr.pull_request_id, 
				pr.pull_request_name, 
				pr.author_id, 
				pr.status, 
				pr.created_at, 
				pr.merged_at,
				ARRAY_AGG(prr.reviewer_id) as reviewer_ids
			FROM pull_requests pr
			INNER JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
			WHERE pr.pull_request_id IN (
				SELECT DISTINCT pull_request_id 
				FROM pull_request_reviewers 
				WHERE reviewer_id = $1
			)
			GROUP BY 
				pr.pull_request_id, 
				pr.pull_request_name, 
				pr.author_id, 
				pr.status, 
				pr.created_at, 
				pr.merged_at
		)
		SELECT 
			pull_request_id, 
			pull_request_name, 
			author_id, 
			status, 
			created_at, 
			merged_at,
			reviewer_ids
		FROM prs_with_reviewers
		ORDER BY created_at DESC
	`

	rows, err := repo.db.Query(query, userID)
	if err != nil {
		repo.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "Failed to query PRs by reviewer",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("find PRs by reviewer: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			repo.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "failed to close sql rows", "error", err)
		}
	}()

	var prs []*entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		var status string
		var mergedAt sql.NullTime
		var reviewerIDs []string

		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&status,
			&pr.CreatedAt,
			&mergedAt,
			pq.Array(&reviewerIDs),
		); err != nil {
			repo.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "Failed to scan PR row",
				"user_id", userID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds())
			return nil, fmt.Errorf("scan PR row: %w", err)
		}

		pr.Status = entity.PullRequestStatus(status)
		if mergedAt.Valid {
			pr.MergedAt = mergedAt.Time
		}
		pr.AssignedReviewers = reviewerIDs

		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		repo.logger.Error("POSTGRES_FIND_PRS_BY_REVIEWER", "Error iterating PR rows",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("iterate PR rows: %w", err)
	}

	repo.logger.Debug("POSTGRES_FIND_PRS_BY_REVIEWER", "PRs found successfully",
		"user_id", userID,
		"prs_count", len(prs),
		"duration_ms", time.Since(start).Milliseconds())
	return prs, nil
}

func (repo *PRRepository) getTeamIDByName(teamName string) (int, error) {
	var teamID int
	query := `SELECT team_id FROM teams WHERE team_name = $1`
	err := repo.db.QueryRow(query, teamName).Scan(&teamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNoTeam
		}
		return 0, fmt.Errorf("get team ID by name: %w", err)
	}
	return teamID, nil
}
