# CryptoBounty — Web3 Bounty Marketplace (MVP)

MVP-платформа, где Web3-проекты публикуют bounty-задачи и депонируют оплату в escrow-смарт-контракте (ETH + whitelisted ERC20), хантеры откликаются/выполняют работу, а выплата происходит после подтверждения выполнения или по решению арбитража (Kleros-like).

## Репозиторий (monorepo)

- `apps/web` — React + TypeScript, wagmi/viem, RainbowKit, TanStack Query, zustand, RHF+zod, Tailwind + shadcn/ui
- `apps/api` — Go (Fiber), GORM, Redis, JWT + SIWE
- `apps/indexer` — Go индексер событий escrow → Postgres/Redis
- `apps/worker` — Go воркер уведомлений (email/webhooks/inbox)
- `packages/contracts` — Solidity + Foundry + OpenZeppelin v5
- `packages/shared` — общие схемы/типы, контрактные ABI/адреса

## Быстрый старт (локально)

### Требования
- Node.js 20+
- Go 1.22+
- Docker (Postgres + Redis)
- Foundry (`forge`, `cast`, `anvil`)

### Локальная инфраструктура
Запустите Postgres + Redis:

```bash
docker compose up -d
```

### Контракты (Anvil)
В отдельном терминале:

```bash
anvil
```

Деплой:

```bash
cd packages/contracts
forge install
forge build
forge script script/Deploy.s.sol:Deploy --rpc-url http://127.0.0.1:8545 --broadcast
```

### Web + API + Indexer + Worker

```bash
npm install
npm run dev
```

## Документация
- Спека workflow и метаданных: `docs/spec.md`

