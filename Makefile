ifneq (,$(wildcard .env))
include .env
export
endif

export PROJECT_ROOT=${CURDIR}

.PHONY: env-up env-down env-cleanup env-reset env-port-forward env-port-close \
        migrate-create migrate-up migrate-down migrate-force migrate-drop migrate-action \
        run-api run-bot run-worker api-up api-down bot-up bot-down worker-up worker-down \
        frontend-install frontend-run frontend-dev-up frontend-dev-recreate frontend-down \
        prod-up prod-down logs logs-db logs-api logs-bot logs-worker logs-frontend \
        db-shell db-status test coverage fmt tidy seed

env-up:
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "New-Item -ItemType Directory -Force -Path 'out/logs' | Out-Null"
	docker compose --profile dev up -d sakeofher-postgres port-forwarder
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "for ($$i = 0; $$i -lt 60; $$i++) { docker exec sakeofher-env-postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB) -h 127.0.0.1; if ($$LASTEXITCODE -eq 0) { Write-Host 'postgres is ready on 127.0.0.1:5433'; exit 0 }; Start-Sleep -Seconds 1 }; Write-Host 'postgres is not ready'; docker logs sakeofher-env-postgres --tail 80; exit 1"

env-down:
	docker compose down

env-cleanup:
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "$$ans = Read-Host 'Clean containers and DB volume? Warning: data loss. [y/N]'; if ($$ans -eq 'y') { docker compose down -v --remove-orphans; docker rm -f sakeofher-env-postgres sakeofher-postgres-port-forwarder 2>$$null; docker volume rm sakeofher_sakeofher-postgres-data 2>$$null; docker volume rm sakeofher-postgres-data 2>$$null; if (Test-Path 'out/pgdata') { Remove-Item -Recurse -Force 'out/pgdata' }; Write-Host 'done' } else { Write-Host 'cancel operation' }"

env-reset:
	make env-cleanup
	make env-up

env-port-forward:
	docker compose --profile dev up -d port-forwarder

env-port-close:
	docker compose stop port-forwarder

db-status:
	docker ps -a --filter "name=sakeofher"
	docker compose ps

db-shell:
	docker exec -it sakeofher-env-postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

migrate-create:
	@if "$(seq)"=="" (echo not seq && exit /b 1)
	docker compose --profile tools run --rm sakeofher-migrate create -ext sql -dir /migrations -seq "$(seq)"

migrate-up:
	make migrate-action action=up

migrate-down:
	make migrate-action action=down

migrate-force:
	@if "$(version)"=="" (echo not version && exit /b 1)
	make migrate-action action="force $(version)"

migrate-drop:
	make migrate-action action=drop

migrate-action:
	@if "$(action)"=="" (echo not action && exit /b 1)
	docker compose --profile tools run --rm sakeofher-migrate \
		-path /migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@sakeofher-postgres:5432/$(POSTGRES_DB)?sslmode=disable" \
		$(action)

run-api:
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "New-Item -ItemType Directory -Force -Path 'out/logs' | Out-Null"
	set "LOGGER_FOLDER=%CD%\out\logs" && set "POSTGRES_HOST=127.0.0.1" && set "POSTGRES_PORT=5433" && go run .\cmd\api

run-bot:
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "New-Item -ItemType Directory -Force -Path 'out/logs' | Out-Null"
	set "LOGGER_FOLDER=%CD%\out\logs" && set "POSTGRES_HOST=127.0.0.1" && set "POSTGRES_PORT=5433" && go run .\cmd\bot

run-worker:
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "New-Item -ItemType Directory -Force -Path 'out/logs' | Out-Null"
	set "LOGGER_FOLDER=%CD%\out\logs" && set "POSTGRES_HOST=127.0.0.1" && set "POSTGRES_PORT=5433" && go run .\cmd\worker

api-up:
	docker compose --profile app up -d --build sakeofher-api

api-down:
	docker compose stop sakeofher-api

bot-up:
	docker compose --profile app up -d --build sakeofher-bot

bot-down:
	docker compose stop sakeofher-bot

worker-up:
	docker compose --profile app up -d --build sakeofher-worker

worker-down:
	docker compose stop sakeofher-worker

frontend-install:
	cd frontend && npm install

frontend-run:
	cd frontend && npm run dev

frontend-dev-up:
	docker compose --profile frontend-dev up -d --force-recreate sakeofher-frontend-dev

frontend-dev-recreate:
	docker compose --profile frontend-dev up -d --force-recreate sakeofher-frontend-dev

frontend-down:
	docker compose stop sakeofher-frontend-dev sakeofher-frontend

prod-up:
	docker compose --profile prod up -d --build

prod-down:
	docker compose --profile prod down

logs:
	docker compose logs -f

logs-db:
	docker logs -f sakeofher-env-postgres

logs-api:
	docker compose logs -f sakeofher-api

logs-bot:
	docker compose logs -f sakeofher-bot

logs-worker:
	docker compose logs -f sakeofher-worker

logs-frontend:
	docker compose logs -f sakeofher-frontend-dev sakeofher-frontend

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

test:
	go test ./... -v -count=1

coverage:
	go test ./... -v -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

seed:
	@if not exist scripts\seed_test_data.sql (echo scripts\seed_test_data.sql not found && exit /b 1)
	docker cp .\scripts\seed_test_data.sql sakeofher-env-postgres:/tmp/seed_test_data.sql
	docker exec -i sakeofher-env-postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -f /tmp/seed_test_data.sql
