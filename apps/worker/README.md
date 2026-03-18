# Worker

Читает Redis Stream событий (`cb_events`), пишет **in-app inbox** (`notifications`) и отправляет **webhooks** (и опционально email).

## Webhook signature
- Header `X-CB-Signature`: `sha256=<hex(hmac_sha256(secret, body))>`
- Header `X-CB-Event`: имя события (например, `BountyCreated`)

## Запуск
Скопируйте `.env.example` → `.env` и выполните:

```bash
go run .
```

