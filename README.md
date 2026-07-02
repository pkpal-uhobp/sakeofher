# SakeOfHer

Backend + Telegram bot + worker + frontend для продажи VPN-подписок через Remnawave.

## Backend stack

- Go
- `net/http` вместо стороннего router
- `pgx/v5` + `pgxpool` для PostgreSQL
- `zap` для логирования
- `envconfig` для конфигурации

## Архитектура

```text
transport -> service -> repository
               ↓
            gateway
```

Основные приложения:

```text
cmd/api    — HTTP API и webhooks
cmd/bot    — Telegram bot
cmd/worker — фоновые задачи
```

## Запуск

```bash
cp .env.example .env
go mod tidy
docker compose up -d postgres
make run-api
```

Проверка:

```bash
curl http://localhost:8080/api/v1/health
```
