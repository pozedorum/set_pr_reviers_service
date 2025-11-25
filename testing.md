# Тестирование PR Reviewer Assignment Service

Это руководство поможет вам протестировать все эндпоинты микросервиса согласно OpenAPI спецификации.

## Быстрый старт

### 1. Запуск сервиса
```bash
docker-compose up -d
```

### 2. Проверка здоровья
```bash
curl http://localhost:8080/health
```
**Ожидаемый ответ:**
```json
{"status": "ok"}
```

## Полное тестирование всех эндпоинтов

### Тест 1: Создание команды
```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {
        "user_id": "user1",
        "username": "Alice",
        "is_active": true
      },
      {
        "user_id": "user2", 
        "username": "Bob",
        "is_active": true
      },
      {
        "user_id": "user3",
        "username": "Charlie",
        "is_active": true
      },
      {
        "user_id": "user4",
        "username": "Dave",
        "is_active": false
      },
      {
        "user_id": "user5",
        "username": "Vladimir",
        "is_active": true
      }
    ]
  }'
```

**Ожидаемый ответ (201 Created):**
```json
{
  "team": {
    "team_name": "backend",
    "members": [
      {"user_id": "user1", "username": "Alice", "is_active": true},
      {"user_id": "user2", "username": "Bob", "is_active": true},
      {"user_id": "user3", "username": "Charlie", "is_active": true},
      {"user_id": "user4", "username": "Dave", "is_active": false}
    ]
  }
}
```

### Тест 2: Получение информации о команде
```bash
curl "http://localhost:8080/team/get?team_name=backend"
```

**Ожидаемый ответ (200 OK):**
```json
{
  "team_name": "backend",
  "members": [
    {"user_id": "user1", "username": "Alice", "is_active": true},
    {"user_id": "user2", "username": "Bob", "is_active": true},
    {"user_id": "user3", "username": "Charlie", "is_active": true},
    {"user_id": "user4", "username": "Dave", "is_active": false}
  ]
}
```

### Тест 3: Изменение активности пользователя
```bash
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user2",
    "is_active": false
  }'
```

**Ожидаемый ответ (200 OK):**
```json
{
  "user": {
    "user_id": "user2",
    "username": "Bob", 
    "team_name": "backend",
    "is_active": false
  }
}
```

### Тест 4: Создание Pull Request
```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-001",
    "pull_request_name": "Add user authentication",
    "author_id": "user1"
  }'
```

**Ожидаемый ответ (201 Created):**
```json
{
  "pr": {
    "pull_request_id": "pr-001",
    "pull_request_name": "Add user authentication",
    "author_id": "user1",
    "status": "OPEN",
    "assigned_reviewers": ["user3"],
    "createdAt": "2024-01-15T10:30:00Z",
    "mergedAt": null
  }
}
```

**Примечание:** Автоматически назначаются до 2 активных ревьюверов из команды автора, исключая самого автора и неактивных пользователей.

### Тест 5: Получение PR пользователя как ревьювера
```bash
curl "http://localhost:8080/users/getReview?user_id=user3"
```

**Ожидаемый ответ (200 OK):**
```json
{
  "user_id": "user3",
  "pull_requests": [
    {
      "pull_request_id": "pr-001",
      "pull_request_name": "Add user authentication", 
      "author_id": "user1",
      "status": "OPEN"
    }
  ]
}
```

### Тест 6: Переназначение ревьювера
```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-001",
    "old_reviewer_id": "user3"
  }'
```

**Ожидаемый ответ (200 OK):**
```json
{
  "pr": {
    "pull_request_id": "pr-001",
    "pull_request_name": "Add user authentication",
    "author_id": "user1", 
    "status": "OPEN",
    "assigned_reviewers": ["user5"],
    "createdAt": "2024-01-15T10:30:00Z",
    "mergedAt": null
  },
  "replaced_by": "user2"
}
```
### Тест 7: Merge Pull Request
```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-001"
  }'
```

**Ожидаемый ответ (200 OK):**
```json
{
  "pr": {
    "pull_request_id": "pr-001",
    "pull_request_name": "Add user authentication",
    "author_id": "user1",
    "status": "MERGED", 
    "assigned_reviewers": ["user2"],
    "createdAt": "2024-01-15T10:30:00Z",
    "mergedAt": "2024-01-15T11:00:00Z"
  }
}
```

### Тест 8: Идемпотентность merge операции
```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-001"
  }'
```

**Ожидаемый ответ (200 OK):** Тот же самый ответ без ошибки.

## Тестирование ошибок

### Ошибка 1: Создание существующей команды
```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "user5", "username": "Eve", "is_active": true}
    ]
  }'
```

**Ожидаемый ответ (409 Conflict):**
```json
{
  "error": {
    "code": "TEAM_EXISTS",
    "message": "team already exists"
  }
}
```

### Ошибка 2: Создание существующего PR
```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-001",
    "pull_request_name": "Another PR",
    "author_id": "user1"
  }'
```

**Ожидаемый ответ (409 Conflict):**
```json
{
  "error": {
    "code": "PR_EXISTS", 
    "message": "pull request already exists"
  }
}
```

### Ошибка 3: Переназначение на мерженом PR
```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-001", 
    "old_reviewer_id": "user2"
  }'
```

**Ожидаемый ответ (409 Conflict):**
```json
{
  "error": {
    "code": "PR_MERGED",
    "message": "cannot reassign on merged PR"
  }
}
```

### Ошибка 4: Несуществующая команда
```bash
curl "http://localhost:8080/team/get?team_name=nonexistent"
```

**Ожидаемый ответ (404 Not Found):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "no such team"
  }
}
```

