COMPOSE := docker compose --env-file backend/.env
DB_URL ?= $(DATABASE_URL)

.PHONY: up down rebuild logs migrate test test-integration lint fmt

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down -v

rebuild:
	$(COMPOSE) up -d --build --force-recreate backend

logs:
	$(COMPOSE) logs -f backend

migrate:
	cd backend && goose -dir ./migrations postgres "$(DB_URL)" up

test:
	cd backend && go test ./internal/...

test-integration:
	cd backend && go test ./tests/integration/...

lint:
	cd backend && golangci-lint run

fmt:
	cd backend && go fmt ./...
