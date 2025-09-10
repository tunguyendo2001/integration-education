.PHONY: build run test clean docker-build docker-up docker-down migrate

# Build all binaries
build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/api ./cmd/api
	CGO_ENABLED=0 GOOS=linux go build -o bin/ingestion-worker ./cmd/ingestion-worker
	CGO_ENABLED=0 GOOS=linux go build -o bin/sync-worker ./cmd/sync-worker

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v

# Database migration (if using migrate tool)
migrate-up:
	migrate -path internal/db/migrations -database "mysql://root:rootpassword@tcp(localhost:3306)/grades_db" up

migrate-down:
	migrate -path internal/db/migrations -database "mysql://root:rootpassword@tcp(localhost:3306)/grades_db" down

# Development commands
dev-api:
	go run ./cmd/api

dev-ingestion:
	go run ./cmd/ingestion-worker

dev-sync:
	go run ./cmd/sync-worker

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# MySQL specific commands
mysql-client:
	docker-compose exec mysql mysql -u root -prootpassword grades_db

mysql-logs:
	docker-compose logs mysql
