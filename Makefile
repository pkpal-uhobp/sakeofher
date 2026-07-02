.PHONY: run-api run-bot run-worker test fmt tidy docker-up docker-down

run-api:
	go run ./cmd/api

run-bot:
	go run ./cmd/bot

run-worker:
	go run ./cmd/worker

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

test:
	go test ./...

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
