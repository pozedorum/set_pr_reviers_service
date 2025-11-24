.PHONY: all build rebuild clean full_clean generate generate-mocks generate-api

all:

build:
	go build -o bin/app cmd/main.go

rebuild:



clean:


full_clean: clean_mocks
	rm -rf internal/generated/*
	
clean_mocks:
	rm -rf internal/mocks


test:
	go test ./internal/service/... -v

generate: generate-mocks generate-api
generate-mocks:
	@echo "Generating mocks..."
	mockery --dir internal/interfaces --name Repository --output internal/mocks --outpkg mocks --filename mock_Repository.go --with-expecter
	@echo "Mocks generated successfully!"


generate-api:
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