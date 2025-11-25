package repository

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pozedorum/set_pr_reviers_service/internal/entity"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
	"github.com/pozedorum/set_pr_reviers_service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB     *sql.DB
	testRepo   *PRRepository
	testLogger interfaces.Logger
)

func TestMain(m *testing.M) {
	// Инициализация логгера
	var err error
	testLogger, err = logger.NewLogger("pr-service-test", "logger_for_tests")
	if err != nil {
		panic(fmt.Sprintf("Failed to create test logger: %v", err))
	}
	defer testLogger.Shutdown()

	// Запуск тестового контейнера PostgreSQL
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	// Запускаем контейнер PostgreSQL
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_DB=test_db",
			"POSTGRES_USER=test_user",
			"POSTGRES_PASSWORD=test_password",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		panic(fmt.Sprintf("Could not start resource: %s", err))
	}

	// Конструктор для подключения к БД
	dsn := fmt.Sprintf("postgres://test_user:test_password@localhost:%s/test_db?sslmode=disable", resource.GetPort("5432/tcp"))

	// Ждем пока БД будет готова принимать подключения
	err = pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		return testDB.Ping()
	})
	if err != nil {
		panic(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	// Инициализируем схему БД
	if err := initTestSchema(testDB); err != nil {
		panic(fmt.Sprintf("Could not init test schema: %s", err))
	}

	// Создаем репозиторий
	testRepo, err = NewPRRepository(dsn, testLogger)
	if err != nil {
		panic(fmt.Sprintf("Could not create test repository: %s", err))
	}

	// Запускаем тесты
	code := m.Run()

	// Очистка
	if err := pool.Purge(resource); err != nil {
		panic(fmt.Sprintf("Could not purge resource: %s", err))
	}

	os.Exit(code)
}

func initTestSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS teams (
			team_id SERIAL PRIMARY KEY,
			team_name VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			team_id INTEGER NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (team_id) REFERENCES teams(team_id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS pull_requests (
			pull_request_id VARCHAR(255) PRIMARY KEY,
			pull_request_name VARCHAR(255) NOT NULL,
			author_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			merged_at TIMESTAMP NULL,
			FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS pull_request_reviewers (
			pull_request_id VARCHAR(255) NOT NULL,
			reviewer_id VARCHAR(255) NOT NULL,
			assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (pull_request_id, reviewer_id),
			FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
			FOREIGN KEY (reviewer_id) REFERENCES users(user_id) ON DELETE CASCADE
		);

		-- Индексы для производительности
		CREATE INDEX IF NOT EXISTS idx_users_team_id ON users(team_id);
		CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
		CREATE INDEX IF NOT EXISTS idx_prs_status ON pull_requests(status);
		CREATE INDEX IF NOT EXISTS idx_prs_author_id ON pull_requests(author_id);
		CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer_id ON pull_request_reviewers(reviewer_id);
		CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(team_name);
	`)
	return err
}

func cleanupTestData() {
	if _, err := testDB.Exec("DELETE FROM pull_request_reviewers"); err != nil {
		panic("failed to cleanupTestData 1")
	}
	if _, err := testDB.Exec("DELETE FROM pull_requests"); err != nil {
		panic("failed to cleanupTestData 2")
	}
	if _, err := testDB.Exec("DELETE FROM users"); err != nil {
		panic("failed to cleanupTestData 3")
	}
	if _, err := testDB.Exec("DELETE FROM teams"); err != nil {
		panic("failed to cleanupTestData 4")
	}
}

func TestCreateTeam_Success(t *testing.T) {
	defer cleanupTestData()

	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{
				UserID:   "user1",
				Username: "Alice",
				IsActive: true,
			},
			{
				UserID:   "user2",
				Username: "Bob",
				IsActive: true,
			},
		},
	}

	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	// Проверяем что команда создалась
	foundTeam, err := testRepo.FindTeamByName("backend")
	require.NoError(t, err)
	assert.Equal(t, "backend", foundTeam.TeamName)
	assert.Len(t, foundTeam.Members, 2)
}

func TestCreateTeam_AlreadyExists(t *testing.T) {
	defer cleanupTestData()

	team1 := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
		},
	}

	err := testRepo.CreateTeam(team1)
	require.NoError(t, err)

	// Пытаемся создать команду с тем же именем
	team2 := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "user3", Username: "Charlie", IsActive: true},
		},
	}

	err = testRepo.CreateTeam(team2)
	assert.Error(t, err)
}

func TestFindTeamByName_Success(t *testing.T) {
	defer cleanupTestData()

	team := &entity.Team{
		TeamName: "frontend",
		Members: []entity.TeamMember{
			{
				UserID:   "user3",
				Username: "Charlie",
				IsActive: true,
			},
		},
	}

	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	foundTeam, err := testRepo.FindTeamByName("frontend")
	require.NoError(t, err)
	assert.Equal(t, "frontend", foundTeam.TeamName)
	assert.Len(t, foundTeam.Members, 1)
	assert.Equal(t, "user3", foundTeam.Members[0].UserID)
	assert.Equal(t, "Charlie", foundTeam.Members[0].Username)
	assert.True(t, foundTeam.Members[0].IsActive)
}

func TestFindTeamByName_NotFound(t *testing.T) {
	defer cleanupTestData()

	team, err := testRepo.FindTeamByName("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, team)
}

func TestTeamExists_Success(t *testing.T) {
	defer cleanupTestData()

	team := &entity.Team{
		TeamName: "devops",
		Members: []entity.TeamMember{
			{UserID: "user4", Username: "David", IsActive: true},
		},
	}

	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	exists := testRepo.TeamExists("devops")
	assert.True(t, exists)

	exists = testRepo.TeamExists("nonexistent")
	assert.False(t, exists)
}

func TestCreateUser_Success(t *testing.T) {
	defer cleanupTestData()

	// Сначала создаем команду
	team := &entity.Team{
		TeamName: "backend",
		Members:  []entity.TeamMember{},
	}
	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	// Создаем пользователя
	user := &entity.User{
		UserID:   "user1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	err = testRepo.CreateUser(user)
	require.NoError(t, err)

	// Проверяем что пользователь создался
	foundUser, err := testRepo.FindUserByID("user1")
	require.NoError(t, err)
	assert.Equal(t, "user1", foundUser.UserID)
	assert.Equal(t, "Alice", foundUser.Username)
	assert.Equal(t, "backend", foundUser.TeamName)
	assert.True(t, foundUser.IsActive)
}

func TestFindUserByID_NotFound(t *testing.T) {
	defer cleanupTestData()

	user, err := testRepo.FindUserByID("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUpdateUser_Success(t *testing.T) {
	defer cleanupTestData()

	// Создаем команду и пользователя
	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
		},
	}
	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	user := &entity.User{
		UserID:   "user1",
		Username: "Alice Updated",
		TeamName: "backend",
		IsActive: false,
	}

	err = testRepo.UpdateUser(user)
	require.NoError(t, err)

	// Проверяем обновление
	updatedUser, err := testRepo.FindUserByID("user1")
	require.NoError(t, err)
	assert.Equal(t, "Alice Updated", updatedUser.Username)
	assert.False(t, updatedUser.IsActive)
}

func TestFindUsersByTeam_Success(t *testing.T) {
	defer cleanupTestData()

	// Создаем команду с несколькими пользователями
	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
			{UserID: "user2", Username: "Bob", IsActive: true},
			{UserID: "user3", Username: "Charlie", IsActive: false},
		},
	}
	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	users, err := testRepo.FindUsersByTeam("backend")
	require.NoError(t, err)
	assert.Len(t, users, 3)

	// Проверяем порядок и данные
	assert.Equal(t, "user1", users[0].UserID)
	assert.Equal(t, "user2", users[1].UserID)
	assert.Equal(t, "user3", users[2].UserID)
}

func TestSetActive_Success(t *testing.T) {
	defer cleanupTestData()

	// Создаем команду и пользователя
	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
		},
	}
	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	// Деактивируем пользователя
	err = testRepo.SetActive("user1", false)
	require.NoError(t, err)

	// Проверяем изменение
	user, err := testRepo.FindUserByID("user1")
	require.NoError(t, err)
	assert.False(t, user.IsActive)

	// Активируем обратно
	err = testRepo.SetActive("user1", true)
	require.NoError(t, err)

	user, err = testRepo.FindUserByID("user1")
	require.NoError(t, err)
	assert.True(t, user.IsActive)
}

func TestSetActive_UserNotFound(t *testing.T) {
	defer cleanupTestData()

	err := testRepo.SetActive("nonexistent", true)
	assert.Error(t, err)
}

func setupTestTeamAndUsers() {
	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "author1", Username: "Author", IsActive: true},
			{UserID: "reviewer1", Username: "Reviewer1", IsActive: true},
			{UserID: "reviewer2", Username: "Reviewer2", IsActive: true},
			{UserID: "reviewer3", Username: "Reviewer3", IsActive: false},
		},
	}
	if err := testRepo.CreateTeam(team); err != nil {
		panic("failed to setupTestTeamAndUsers")
	}
}

func TestCreatePR_Success(t *testing.T) {
	defer cleanupTestData()
	setupTestTeamAndUsers()

	pr := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer1", "reviewer2"},
		CreatedAt:         []time.Time{time.Now()}[0],
	}

	err := testRepo.CreatePR(pr)
	require.NoError(t, err)

	// Проверяем что PR создался
	foundPR, err := testRepo.FindPRByID("pr-123")
	require.NoError(t, err)
	assert.Equal(t, "pr-123", foundPR.PullRequestID)
	assert.Equal(t, "Test PR", foundPR.PullRequestName)
	assert.Equal(t, "author1", foundPR.AuthorID)
	assert.Equal(t, entity.PullRequestStatusOpen, foundPR.Status)
	assert.Len(t, foundPR.AssignedReviewers, 2)
	assert.Contains(t, foundPR.AssignedReviewers, "reviewer1")
	assert.Contains(t, foundPR.AssignedReviewers, "reviewer2")
}

func TestFindPRByID_NotFound(t *testing.T) {
	defer cleanupTestData()

	pr, err := testRepo.FindPRByID("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, pr)
}

func TestUpdatePR_Success(t *testing.T) {
	defer cleanupTestData()
	setupTestTeamAndUsers()

	// Создаем PR
	pr := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer1"},
	}
	err := testRepo.CreatePR(pr)
	require.NoError(t, err)

	// Обновляем PR
	updatedPR := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Updated PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusMerged,
		AssignedReviewers: []string{"reviewer1", "reviewer2"},
		MergedAt:          []time.Time{time.Now()}[0],
	}

	err = testRepo.UpdatePR(updatedPR)
	require.NoError(t, err)

	// Проверяем обновление
	foundPR, err := testRepo.FindPRByID("pr-123")
	require.NoError(t, err)
	assert.Equal(t, "Updated PR", foundPR.PullRequestName)
	assert.Equal(t, entity.PullRequestStatusMerged, foundPR.Status)
	assert.Len(t, foundPR.AssignedReviewers, 2)
	assert.NotNil(t, foundPR.MergedAt)
}

// УДАЛЕН: TestClosePR_Success - используем UpdatePR вместо ClosePR

func TestFindPRsByReviewer_Success(t *testing.T) {
	defer cleanupTestData()
	setupTestTeamAndUsers()

	// Создаем несколько PR для одного ревьювера
	pr1 := &entity.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "PR 1",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer1"},
	}
	err := testRepo.CreatePR(pr1)
	require.NoError(t, err)

	pr2 := &entity.PullRequest{
		PullRequestID:     "pr-2",
		PullRequestName:   "PR 2",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusMerged,
		AssignedReviewers: []string{"reviewer1", "reviewer2"},
	}
	err = testRepo.CreatePR(pr2)
	require.NoError(t, err)

	// PR без нашего ревьювера
	pr3 := &entity.PullRequest{
		PullRequestID:     "pr-3",
		PullRequestName:   "PR 3",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer2"},
	}
	err = testRepo.CreatePR(pr3)
	require.NoError(t, err)

	// Ищем PR для reviewer1
	prs, err := testRepo.FindPRsByReviewer("reviewer1")
	require.NoError(t, err)
	assert.Len(t, prs, 2)

	// Проверяем что вернулись правильные PR
	prIDs := make([]string, len(prs))
	for i, pr := range prs {
		prIDs[i] = pr.PullRequestID
	}
	assert.Contains(t, prIDs, "pr-1")
	assert.Contains(t, prIDs, "pr-2")
	assert.NotContains(t, prIDs, "pr-3")

	// Проверяем что у каждого PR есть правильные ревьюверы
	for _, pr := range prs {
		assert.Contains(t, pr.AssignedReviewers, "reviewer1")
	}
}

func TestFindPRsByReviewer_NoPRs(t *testing.T) {
	defer cleanupTestData()
	setupTestTeamAndUsers()

	prs, err := testRepo.FindPRsByReviewer("reviewer1")
	require.NoError(t, err)
	assert.Empty(t, prs)
}

func TestCreateTeam_TransactionRollbackOnDuplicateUser(t *testing.T) {
	defer cleanupTestData()

	// Сначала создаем команду с пользователем
	team1 := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
		},
	}
	err := testRepo.CreateTeam(team1)
	require.NoError(t, err)

	// Пытаемся создать другую команду с тем же пользователем (дубликат user_id)
	team2 := &entity.Team{
		TeamName: "frontend",
		Members: []entity.TeamMember{
			{UserID: "user1", Username: "Alice Duplicate", IsActive: true}, // Тот же user_id!
		},
	}

	err = testRepo.CreateTeam(team2)
	assert.Error(t, err, "Should fail due to duplicate user_id")

	// Проверяем что вторая команда не создалась
	exists := testRepo.TeamExists("frontend")
	assert.False(t, exists, "Team should not be created due to transaction rollback")

	// Проверяем что первая команда осталась нетронутой
	backendTeam, err := testRepo.FindTeamByName("backend")
	require.NoError(t, err)
	assert.Equal(t, "backend", backendTeam.TeamName)
	assert.Len(t, backendTeam.Members, 1)

	// Проверяем что пользователь остался в первой команде
	var userTeamName string
	err = testDB.QueryRow(`
		SELECT t.team_name 
		FROM users u 
		JOIN teams t ON u.team_id = t.team_id 
		WHERE u.user_id = $1
	`, "user1").Scan(&userTeamName)
	require.NoError(t, err)
	assert.Equal(t, "backend", userTeamName)
}

func TestCreatePR_TransactionRollbackOnInvalidReviewer(t *testing.T) {
	defer cleanupTestData()

	// Создаем команду и валидных пользователей
	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "author1", Username: "Author", IsActive: true},
			{UserID: "reviewer1", Username: "Reviewer1", IsActive: true},
		},
	}
	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	// Пытаемся создать PR с несуществующим ревьювером
	pr := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer1", "nonexistent_reviewer"}, // Один ревьювер не существует
	}

	err = testRepo.CreatePR(pr)
	assert.Error(t, err, "Should fail due to foreign key constraint")

	// Проверяем что PR не создался
	foundPR, err := testRepo.FindPRByID("pr-123")
	assert.Error(t, err)
	assert.Nil(t, foundPR)

	// Проверяем что в таблице ревьюверов нет записей для этого PR
	var count int
	err = testDB.QueryRow(
		"SELECT COUNT(*) FROM pull_request_reviewers WHERE pull_request_id = $1",
		"pr-123",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "No reviewers should be created for rolled back PR")
}

func TestUpdatePR_TransactionAtomicity(t *testing.T) {
	defer cleanupTestData()
	setupTestTeamAndUsers()

	// Создаем PR
	pr := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer1"},
	}
	err := testRepo.CreatePR(pr)
	require.NoError(t, err)

	// Пытаемся обновить PR с несуществующим ревьювером
	updatedPR := &entity.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Updated PR",
		AuthorID:          "author1",
		Status:            entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"nonexistent_reviewer"}, // Несуществующий ревьювер
	}

	err = testRepo.UpdatePR(updatedPR)
	assert.Error(t, err, "Should fail due to foreign key constraint")

	// Проверяем что оригинальные данные не изменились
	foundPR, err := testRepo.FindPRByID("pr-123")
	require.NoError(t, err)
	assert.Equal(t, "Test PR", foundPR.PullRequestName, "PR name should not change")
	assert.Equal(t, []string{"reviewer1"}, foundPR.AssignedReviewers, "Reviewers should not change")
}

func TestCreateTeam_TransactionCommitSuccess(t *testing.T) {
	defer cleanupTestData()

	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
			{UserID: "user2", Username: "Bob", IsActive: true},
		},
	}

	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	// Проверяем что команда создалась
	exists := testRepo.TeamExists("backend")
	assert.True(t, exists)

	// Проверяем что все пользователи создались
	user1, err := testRepo.FindUserByID("user1")
	require.NoError(t, err)
	assert.Equal(t, "Alice", user1.Username)
	assert.Equal(t, "backend", user1.TeamName)

	user2, err := testRepo.FindUserByID("user2")
	require.NoError(t, err)
	assert.Equal(t, "Bob", user2.Username)
	assert.Equal(t, "backend", user2.TeamName)

	// Проверяем через raw SQL что данные действительно в базе
	var teamCount, userCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM teams WHERE team_name = $1", "backend").Scan(&teamCount)
	require.NoError(t, err)
	assert.Equal(t, 1, teamCount)

	// ИСПРАВЛЕННЫЙ ЗАПРОС: используем JOIN для подсчета пользователей по team_name
	err = testDB.QueryRow(`
		SELECT COUNT(*) 
		FROM users u 
		JOIN teams t ON u.team_id = t.team_id 
		WHERE t.team_name = $1
	`, "backend").Scan(&userCount)
	require.NoError(t, err)
	assert.Equal(t, 2, userCount)
}

func TestRepository_ConnectionPool(t *testing.T) {
	defer cleanupTestData()

	// Проверяем что репозиторий правильно использует пул соединений
	// путем выполнения нескольких параллельных операций
	team := &entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
		},
	}
	err := testRepo.CreateTeam(team)
	require.NoError(t, err)

	// Параллельные чтения
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			user, err := testRepo.FindUserByID("user1")
			assert.NoError(t, err)
			assert.NotNil(t, user)
			done <- true
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestCreateUser_WithoutTeam(t *testing.T) {
	defer cleanupTestData()

	// Пытаемся создать пользователя без предварительного создания команды
	// Это должно вызвать ошибку внешнего ключа
	user := &entity.User{
		UserID:   "user1",
		Username: "Alice",
		TeamName: "nonexistent_team", // Команды не существует!
		IsActive: true,
	}

	err := testRepo.CreateUser(user)
	assert.Error(t, err, "Should fail due to foreign key constraint")

	// Проверяем что пользователь не создался
	foundUser, err := testRepo.FindUserByID("user1")
	assert.Error(t, err)
	assert.Nil(t, foundUser)
}

func TestSetActive_UserNotExists(t *testing.T) {
	defer cleanupTestData()

	err := testRepo.SetActive("nonexistent_user", true)
	assert.Error(t, err)
	// Исправлено: проверяем на нашу кастомную ошибку, а не sql.ErrNoRows
	assert.ErrorIs(t, err, ErrNoUser)
}
