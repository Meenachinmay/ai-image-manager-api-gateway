.PHONY: help build run test clean docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application locally"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker-up   - Start Docker containers"
	@echo "  make docker-down - Stop Docker containers"
	@echo "  make dev         - Run in development mode with hot reload"

build:
	go build -o bin/api-gateway cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/ tmp/

docker-up:
	docker compose -f docker-compose.local.yml up -d

docker-down:
	docker compose -f docker-compose.local.yml down

docker-logs:
	docker compose -f docker-compose.local.yml logs -f

dev:
	air -c .air.toml

install-deps:
	go mod download
	go mod tidy

install-air:
	go install github.com/cosmtrek/air@latest

ch:
	curl http://localhost:8080/api/v1/health/

rabbit-health:
	curl http://localhost:8080/api/v1/health/