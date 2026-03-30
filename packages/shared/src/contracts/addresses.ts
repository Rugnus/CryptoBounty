export type ChainId = 31337 | 11155111;

export const contracts = {
  31337: {
    chainId: 31337,
    name: "anvil",
    bountyEscrow: "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512",
    usdc: "0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0",
    usdt: "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9",
    mockArbitrator: "0x5FbDB2315678afecb367f032d93F642f64180aa3"
  },
  11155111: {
    chainId: 11155111,
    name: "sepolia",
    bountyEscrow: "0x0000000000000000000000000000000000000000"
  }
} as const satisfies Record<number, unknown>;

