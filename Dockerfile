FROM golang:1.25.1-alpine

WORKDIR /app

# Установка зависимостей
RUN apk add --no-cache git postgresql-client

# Копируем сначала только mod файлы для кэширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем остальные файлы проекта
COPY . .
# COPY internal/frontend ./internal/frontend
# Собираем приложение
RUN go build -o pr-service  ./cmd/main.go

EXPOSE 8080

CMD ["./pr-service"]