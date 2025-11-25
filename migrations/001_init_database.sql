CREATE TABLE teams (
    team_id SERIAL PRIMARY KEY,           -- Внутренний ID для эффективности БД
    team_name VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_id INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(team_id) ON DELETE CASCADE
);

CREATE TABLE pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP NULL,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE pull_request_reviewers (
    pull_request_id VARCHAR(255) NOT NULL,
    reviewer_id VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (pull_request_id, reviewer_id), -- Составной первичный ключ
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Индексы для производительности
CREATE INDEX idx_users_team_id ON users(team_id);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_prs_status ON pull_requests(status);
CREATE INDEX idx_prs_author_id ON pull_requests(author_id);
CREATE INDEX idx_pr_reviewers_reviewer_id ON pull_request_reviewers(reviewer_id);
CREATE INDEX idx_teams_name ON teams(team_name);