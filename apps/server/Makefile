.PHONY: up down run test

up:
	docker compose -f ../../infra/docker-compose.yml up -d

down:
	docker compose -f ../../infra/docker-compose.yml down

run:
	go run ./cmd/api

test:
	go test ./...