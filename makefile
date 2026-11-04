.PHONY: help dev build run swagger test clean docker-up docker-down

help:
	@echo "Available commands in Biblio:"
	@echo "  make docker-up   - Start background Docker containers (PostgreSQL, MinIO, Jaeger, Grafana)"
	@echo "  make docker-down - Stop and remove Docker containers"
	@echo "  make swagger     - Regenerate OpenAPI / Swagger documentation"
	@echo "  make dev         - Run development server with live Swagger regeneration"
	@echo "  make build       - Compile production binary to bin/biblio.exe"

docker-up:
	docker compose up -d

docker-down:
	docker compose down

swagger:
	swag init -g cmd/api/main.go -o docs

dev: swagger
	go run cmd/api/main.go

build:
	go build -o bin/biblio.exe cmd/api/main.go
