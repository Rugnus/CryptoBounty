import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";

// Reads Foundry broadcast JSON and writes addresses into packages/shared/src/contracts/addresses.ts
// Usage: node scripts/sync-contract-addresses.mjs <chainId> <broadcastJsonPath>

const [chainIdArg, broadcastPath] = process.argv.slice(2);
if (!chainIdArg || !broadcastPath) {
  console.error(
    "Usage: node scripts/sync-contract-addresses.mjs <chainId> <broadcastJsonPath>"
  );
  process.exit(1);
}

const chainId = Number(chainIdArg);
const json = JSON.parse(await readFile(broadcastPath, "utf8"));

// Foundry script broadcast structure: transactions[] contain contractName + contractAddress
const txs = json.transactions ?? [];
const byName = new Map();
for (const t of txs) {
  if (t.contractName && t.contractAddress) byName.set(t.contractName, t.contractAddress);
}

const bountyEscrow = byName.get("BountyEscrow");
if (!bountyEscrow) {
  console.error("Could not find BountyEscrow in broadcast file.");
  process.exit(1);
}

const usdc = byName.get("MockERC20"); // there are two; Foundry uses same name twice
const mockArbitrator = byName.get("MockArbitrator");

const addressesTsPath = path.join(
  process.cwd(),
  "packages",
  "shared",
  "src",
  "contracts",
  "addresses.ts"
);
const current = await readFile(addressesTsPath, "utf8");

function replaceChainBlock(text) {
  const re = new RegExp(`${chainId}: \\{[\\s\\S]*?\\n\\s*\\},`, "m");
  const block =
    chainId === 31337
      ? `${chainId}: {\n    chainId: ${chainId},\n    name: "anvil",\n    bountyEscrow: "${bountyEscrow}",\n    usdc: "${usdc ?? "0x0000000000000000000000000000000000000000"}",\n    usdt: "0x0000000000000000000000000000000000000000",\n    mockArbitrator: "${mockArbitrator ?? "0x0000000000000000000000000000000000000000"}"\n  },`
      : `${chainId}: {\n    chainId: ${chainId},\n    name: "sepolia",\n    bountyEscrow: "${bountyEscrow}"\n  },`;

  if (!re.test(text)) throw new Error(`chainId block ${chainId} not found in addresses.ts`);
  return text.replace(re, block);
}

await writeFile(addressesTsPath, replaceChainBlock(current), "utf8");
console.log(`Updated ${addressesTsPath} for chainId=${chainId}`);

