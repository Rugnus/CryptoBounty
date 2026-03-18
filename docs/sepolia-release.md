# Sepolia release checklist (MVP)

## 0) Preconditions
- Foundry установлен (`foundryup`).
- RPC провайдер для Sepolia (Alchemy/Infura) и достаточно testnet ETH на деплойер-адресе.
- Для фронта: WalletConnect Project ID.

## 1) Deploy contracts to Sepolia
В `packages/contracts` создайте `.env`:

```env
DEPLOYER_PRIVATE_KEY=0x...
ARBITRATOR_ADDRESS=0x... # опционально: адрес реального Kleros-like arbitrator
```

Деплой:

```bash
cd packages/contracts
npm run install:deps
forge build
forge script script/Deploy.s.sol:Deploy --rpc-url $SEPOLIA_RPC_URL --broadcast
```

После деплоя синхронизируйте адреса в `packages/shared`:

```bash
node ../../scripts/sync-contract-addresses.mjs 11155111 broadcast/Deploy.s.sol/11155111/run-latest.json
```

## 2) Configure services (Sepolia)
### API (`apps/api/.env`)
- `CHAIN_ID=11155111`
- `POSTGRES_URL=...`
- `REDIS_URL=...`
- `JWT_SECRET=...`
- `SIWE_DOMAIN=<ваш домен>`
- `SIWE_URI=<https://ваш-frontend>`
- `CORS_ORIGIN=<https://ваш-frontend>`

### Indexer (`apps/indexer/.env`)
- `CHAIN_ID=11155111`
- `RPC_URL=$SEPOLIA_RPC_URL`
- `ESCROW_ADDRESS=<адрес из shared>`
- `CONFIRMATIONS=4` (рекомендовано для testnet)

### Worker (`apps/worker/.env`)
- Postgres/Redis как обычно
- (опционально) SMTP параметры + verified emails в `user_emails`

## 3) Configure web (`apps/web/.env`)
- `VITE_API_URL=https://<api>/api`
- `VITE_RPC_SEPOLIA=$SEPOLIA_RPC_URL`
- `VITE_WC_PROJECT_ID=...`

## 4) E2E test plan (минимум)
### Scenario A: approve payout
- Sponsor создает bounty (ETH) через UI → ждете подтверждение.
- Hunter делает `Apply` → Sponsor делает `Assign`.
- Hunter `Submit work` → Sponsor `Approve` → Sponsor `Payout`.
- Проверить:
  - событие `PaidOut` в explorer,
  - `GET /bounties/:id` → `status=PaidOut`,
  - `GET /notifications` у hunter и sponsor содержит события.

### Scenario B: dispute → ruling → payout/refund
- Sponsor создает bounty → assign → submit.
- Sponsor нажимает `Reject & dispute` с fee.
- Арбитражный протокол должен вызвать `rule()` (в MVP на Sepolia можно подключить реальный Kleros-like arbitrator; пока используется mock arbitrator из Deploy скрипта).
- После `ruling`:
  - если hunter wins → `Payout`
  - если sponsor wins → `Refund`

## 5) Observability (минимум)
- Логи сервисов в stdout + ротация (Docker/PM2/Systemd).
- Метрика: текущий head/last_finalized_block (можно добавить позже).

