# Indexer

Сервис читает события `BountyEscrow` из RPC и пишет:\n
- нормализованные таблицы (`bounties`, `applications`, `bounty_events`) в Postgres\n
- поток событий в Redis Stream (для воркера уведомлений)\n

## Env\n
Скопируйте `.env.example` → `.env` и установите `ESCROW_ADDRESS` (после деплоя контрактов).\n

## Запуск\n
```bash\n
go run .\n
```\n

