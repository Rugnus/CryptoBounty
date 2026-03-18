import { useCallback, useState } from "react";
import { SiweMessage } from "siwe";
import { useAccount, useChainId, useSignMessage } from "wagmi";
import { api } from "../../lib/api";
import { useAuthStore } from "../../lib/authStore";

export function useSIWE() {
  const { address } = useAccount();
  const chainId = useChainId();
  const { signMessageAsync } = useSignMessage();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [isLoading, setIsLoading] = useState(false);

  const login = useCallback(async () => {
    if (!address) return;
    setIsLoading(true);
    try {
      const { data } = await api.get<{ nonce: string }>("/auth/siwe/nonce");
      const msg = new SiweMessage({
        domain: window.location.hostname,
        address,
        statement: "Sign in to CryptoBounty.",
        uri: window.location.origin,
        version: "1",
        chainId,
        nonce: data.nonce
      });
      const message = msg.prepareMessage();
      const signature = await signMessageAsync({ message });
      const verified = await api.post<{ token: string; address: string }>(
        "/auth/siwe/verify",
        { message, signature }
      );
      setAuth(verified.data.token, verified.data.address as `0x${string}`);
    } finally {
      setIsLoading(false);
    }
  }, [address, chainId, setAuth, signMessageAsync]);

  return { login, isLoading };
}

