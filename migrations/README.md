# SakeOfHer migrations v2

Миграции под систему:

```text
Telegram Bot + сайт/backend + PostgreSQL + Remnawave
Оплата: Telegram Stars, Tribute RUB, CryptoBot/Crypto Pay
```

## Логика подписки

Фиксированная бизнес-логика:

```text
срок подписки закончился -> DISABLED в Remnawave
7 дней нет оплаты -> DELETE в Remnawave
```

Локального Telegram-пользователя из таблицы `users` не удаляем: он нужен для истории оплат, повторных покупок, рассылок и админки.

## Файлы

- `001_initial_schema.up.sql` — создаёт все таблицы.
- `001_initial_schema.down.sql` — удаляет все таблицы.
- `002_seed_tariffs_and_prices.up.sql` — добавляет стартовые тарифы и цены.
- `002_seed_tariffs_and_prices.down.sql` — удаляет стартовые тарифы и цены.

## Основные таблицы

- `users` — Telegram-пользователи + связь с Remnawave.
- `tariffs` — сами тарифы: срок, период, лимит трафика.
- `tariff_prices` — цены тарифа для Stars, Tribute, CryptoBot.
- `payments` — платежи всех провайдеров в едином формате.
- `payment_events` — защита от повторной обработки webhook/successful_payment.
- `subscriptions` — активные/истёкшие подписки, трафик и уведомления.
- `admins` — администраторы по Telegram ID.
- `admin_actions` — журнал действий админов.
- `broadcasts` — рассылки.
- `broadcast_recipients` — получатели рассылок.
- `remna_sync_logs` — лог запросов к Remnawave и ошибок синхронизации.

## Применить на чистую базу

```bash
psql "$DATABASE_URL" -f 001_initial_schema.up.sql
psql "$DATABASE_URL" -f 002_seed_tariffs_and_prices.up.sql
```

Или напрямую:

```bash
psql -h localhost -U postgres -d vpn_db -f 001_initial_schema.up.sql
psql -h localhost -U postgres -d vpn_db -f 002_seed_tariffs_and_prices.up.sql
```

## Откат

```bash
psql "$DATABASE_URL" -f 002_seed_tariffs_and_prices.down.sql
psql "$DATABASE_URL" -f 001_initial_schema.down.sql
```

## Стартовые тарифы

В `002_seed_tariffs_and_prices.up.sql` сейчас заданы:

```text
1 месяц, 300 GB:
- Stars: 50 XTR
- Tribute RUB: 70 ₽
- CryptoBot: 65 ₽, accepted_assets: USDT, TON

3 месяца, 300 GB каждые 30 дней:
- Stars: 150 XTR
- Tribute RUB: 210 ₽
- CryptoBot: 195 ₽, accepted_assets: USDT, TON
```

Цены можно поменять прямо в seed-файле или потом через админку.

## Важные поля платежей

`payments.status`:

```text
created
waiting_payment
paid
activation_failed
activated
failed
cancelled
expired
refunded
```

Важно разделять:

```text
paid      = деньги пришли
activated = подписка реально создана/продлена в Remnawave
```

Если Remnawave временно не отвечает, платеж должен остаться `paid` или `activation_failed`, а worker должен повторить активацию позже.

## Idempotency

Все входящие события оплаты сначала пишутся в `payment_events`.

Примеры `event_id`:

```text
Telegram Stars: telegram_payment_charge_id
CryptoBot: update_id или invoice_id + event type
Tribute: order id / webhook id, который приходит от Tribute
```

Перед продлением подписки backend должен проверить, что событие ещё не обработано. Это защищает от повторных webhook и двойного продления.

## Worker-задачи

Минимально нужны такие задачи:

```text
syncRemnaUsageJob       — раз в час получает usage из Remnawave
expireSubscriptionsJob  — раз в час/день отключает истёкшие подписки
resetTrafficPeriodJob   — раз в час/день сбрасывает период трафика
sendNotificationsJob    — раз в день шлёт предупреждения
retryActivationJob      — повторяет Remnawave activation_failed
cleanupDisabledUsersJob — раз в день удаляет из Remnawave после 7 дней без оплаты
```

## Рекомендуемый алгоритм оплаты

```text
1. Пользователь выбирает тариф.
2. Backend создаёт payments.status = created.
3. Backend создаёт invoice/order у провайдера.
4. Backend ставит payments.status = waiting_payment.
5. Пользователь оплачивает.
6. Backend получает webhook / Telegram successful_payment.
7. Backend пишет событие в payment_events.
8. Backend ставит payments.status = paid.
9. Backend создаёт/продлевает Remnawave user.
10. Backend ставит payments.status = activated.
11. Бот отправляет пользователю VPN-ссылку.
```

## Рекомендуемый алгоритм истечения подписки

```text
1. subscriptions.expires_at <= now()
2. Backend отключает пользователя в Remnawave.
3. subscriptions.status = expired
4. subscriptions.period_status = finished
5. users.remna_status = disabled
6. users.disabled_at = now()
7. users.delete_after = now() + interval '7 days'
8. Бот отправляет уведомление.
```

Через 7 дней:

```text
1. users.remna_status = disabled
2. users.delete_after <= now()
3. Backend удаляет пользователя из Remnawave.
4. users.remna_status = deleted
5. users.deleted_at = now()
```

Если пользователь оплатил в течение 7 дней, backend должен не создавать нового пользователя, а включить старого в Remnawave и продлить подписку.
