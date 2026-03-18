import { getDefaultConfig } from "@rainbow-me/rainbowkit";
import { http } from "wagmi";
import { anvil, sepolia } from "wagmi/chains";

export const chains = [anvil, sepolia] as const;

export const wagmiConfig = getDefaultConfig({
  appName: "CryptoBounty",
  projectId: import.meta.env.VITE_WC_PROJECT_ID ?? "demo",
  chains,
  transports: {
    [anvil.id]: http(import.meta.env.VITE_RPC_ANVIL ?? "http://127.0.0.1:8545"),
    [sepolia.id]: http(import.meta.env.VITE_RPC_SEPOLIA)
  },
  ssr: false
});

