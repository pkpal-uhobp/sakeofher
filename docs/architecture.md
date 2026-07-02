# SakeOfHer Architecture

Проект использует слоистую архитектуру:

```text
transport -> service -> repository
               ↓
            gateway
```

## Стек backend

- HTTP: стандартный `net/http` и `http.ServeMux`.
- PostgreSQL: `pgx/v5` и `pgxpool`.
- Logger: `zap`.
- Config: `envconfig`.

## Слои

### transport

Принимает входящие запросы:

- HTTP API;
- webhooks;
- Telegram bot updates.

Transport не работает напрямую с PostgreSQL и Remnawave. Он вызывает только service-слой.

### service

Содержит бизнес-логику:

- создание платежей;
- обработка webhook;
- активация подписок;
- отключение пользователя после окончания срока;
- удаление пользователя через 7 дней;
- синхронизация трафика;
- рассылки и уведомления.

Транзакции открываются именно здесь.

### repository

Отвечает только за PostgreSQL.

Внутри repository есть инфраструктурные подпакеты:

```text
repository/pool — создание pgxpool.Pool
repository/tx   — transaction manager и общий Querier
```

Все repository-методы принимают `context.Context` и выполняют запросы через `tx.Manager.Querier(ctx)`.
Если в context есть транзакция, запрос идёт через неё. Если нет — через pgxpool.

### gateway

Отвечает за внешние API:

- Remnawave;
- Tribute;
- CryptoBot;
- Telegram notifier/login/stars helpers.

Gateway не знает про PostgreSQL и транзакции.

## Правило транзакций

Внешние API-вызовы нельзя держать внутри долгой SQL-транзакции.

Правильно:

```text
1. Короткая транзакция: payment_event + payment = paid
2. ВНЕ транзакции: вызов Remnawave API
3. Короткая транзакция: сохранить remna_uuid + subscription + payment = activated
```
