# CryptoBounty — Web3 Bounty Marketplace (MVP)

An MVP platform where Web3 projects post bounty tasks and deposit payments into an escrow smart contract (ETH + whitelisted ERC20). Hunters apply for and complete tasks, and payment is released upon confirmation of completion or by an arbitration decision (Kleros-like).

## Repository (monorepo)

- `apps/web` — React + TypeScript, wagmi/viem, RainbowKit, TanStack Query, zustand, RHF+zod, Tailwind + shadcn/ui
- `apps/api` — Go (Fiber), GORM, Redis, JWT + SIWE
- `apps/indexer` — Go event indexer for escrow → Postgres/Redis
- `apps/worker` — Go notification worker (email/webhooks/inbox)
- `packages/contracts` — Solidity + Foundry + OpenZeppelin v5
- `packages/shared` — common schemas/types, contract ABIs/addresses

## Quick Start (Local)

### Requirements
- Node.js 20+
- Go 1.22+
- Docker (Postgres + Redis)
- Foundry (`forge`, `cast`, `anvil`)

### Local Infrastructure
Run Postgres + Redis:

```bash
docker compose up -d
```

### Contracts (Anvil)
In a separate terminal:

```bash
anvil
```

Deploy:

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

## Documentation
- Workflow and metadata spec: `docs/spec.md`
