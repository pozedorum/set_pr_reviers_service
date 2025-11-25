.PHONY: all build rebuild clean full_clean generate generate_mocks generate_api

all: run

run:
	docker compose up

build:
	docker compose build

rebuild:
	docker compose down -v
	docker compose up --build


clean:
	docker compose down -v
	rm -rf pr-service	

full_clean: clean_mocks
	rm -rf internal/generated/*
	
clean_mocks:
	rm -rf internal/mocks

lint:
	golangci-lint run ./...
	go vet ./...


test:
	go test ./internal/service/... -v
	go test ./internal/repository/... -v

generate: generate-mocks generate-api
generate_mocks:
	@echo "Generating mocks..."
	mockery --dir internal/interfaces --name Repository --output internal/mocks --outpkg mocks --filename mock_Repository.go --with-expecter
	@echo "Mocks generated successfully!"


generate_api:
	@echo "Generating Go code from OpenAPI spec..."
	@mkdir -p internal/generated
	
	@echo "Generating types..."
	@oapi-codegen \
		-generate types \
		-package generated \
		api/openapi.yml > internal/generated/types.gen.go
	
	@echo "Generating Gin server..."
	@oapi-codegen \
		-generate gin \
		-package generated \
		api/openapi.yml > internal/generated/server.gen.go
	
	@echo "Generating client..."
	@oapi-codegen \
		-generate client \
		-package generated \
		api/openapi.yml > internal/generated/client.gen.go
	
	@echo "API code generation completed!"