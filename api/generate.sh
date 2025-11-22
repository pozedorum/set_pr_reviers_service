#!/bin/bash

set -e

echo "🔨 Generating Go code from OpenAPI spec..."

# Создаем директорию если нет
mkdir -p ../internal/generated

# Генерация моделей
echo "📦 Generating types..."
oapi-codegen \
    -generate types \
    -package generated \
    openapi.yml > ../internal/generated/types.gen.go

# Генерация Gin сервера
echo "🚀 Generating Gin server..."
oapi-codegen \
    -generate gin \
    -package generated \
    openapi.yml > ../internal/generated/server.gen.go

# Генерация клиента (опционально)
echo "🔌 Generating client..."
oapi-codegen \
    -generate client \
    -package generated \
    openapi.yml > ../internal/generated/client.gen.go

echo "✅ All code generation completed!"