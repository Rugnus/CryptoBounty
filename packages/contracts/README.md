# Contracts

## Install deps

```bash
npm run install:deps
```

## Local (Anvil)
Run anvil:

```bash
anvil
```

Use Anvil default dev key 0:
- Address: `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266`
- Private key: `0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80`

Create `.env` from `.env.example`:

```bash
cp .env.example .env
```

and set:

```env
DEPLOYER_PRIVATE_KEY=0xac0974...
```

Deploy:

```bash
forge script script/Deploy.s.sol:Deploy --rpc-url http://127.0.0.1:8545 --broadcast
```

Sync addresses into shared package:

```bash
node ../../scripts/sync-contract-addresses.mjs 31337 broadcast/Deploy.s.sol/31337/run-latest.json
```

