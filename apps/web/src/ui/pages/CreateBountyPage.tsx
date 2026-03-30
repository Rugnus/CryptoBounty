import { zodResolver } from "@hookform/resolvers/zod";
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { isAddress, keccak256, parseEther, toBytes } from "viem";
import { useChainId, usePublicClient, useWriteContract } from "wagmi";
import { z } from "zod";

import { bountyEscrowAbi, contracts } from "@cryptobounty/shared";
import { Button } from "../components/Button";
import { Input } from "../components/Input";

const schema = z.object({
  title: z.string().min(3).max(140),
  description: z.string().min(1).max(20000),
  metadataUri: z.string().min(3),
  token: z.string().optional(), // blank => ETH
  amount: z.string().min(1),
});

type FormValues = z.infer<typeof schema>;

export function CreateBountyPage() {
  const chainId = useChainId();
  const escrow = (contracts as any)[chainId]?.bountyEscrow as
    | `0x${string}`
    | undefined;
  const publicClient = usePublicClient();
  const { writeContractAsync } = useWriteContract();
  const [txHash, setTxHash] = useState<`0x${string}` | null>(null);
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { isSubmitting, errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      metadataUri: "ipfs://todo",
      token: "",
      amount: "0.01",
    },
  });

  const onSubmit = useMemo(
    () =>
      handleSubmit(async (v) => {
        setError(null);
        setTxHash(null);

        if (
          !escrow ||
          escrow === "0x0000000000000000000000000000000000000000"
        ) {
          setError(
            "ESCROW address not configured for this chain (update shared addresses).",
          );
          return;
        }
        if (!publicClient) {
          setError("No public client.");
          return;
        }

        const metadata = {
          title: v.title,
          description: v.description,
          category: "other",
          tags: [],
          difficulty: "medium",
          payout: { tokenSymbol: v.token ? "ERC20" : "ETH", amount: v.amount },
          chainId,
          createdAt: new Date().toISOString(),
        };
        const metadataHash = keccak256(toBytes(JSON.stringify(metadata)));

        if (!v.token) {
          const hash = await writeContractAsync({
            abi: bountyEscrowAbi,
            address: escrow,
            functionName: "createBountyETH",
            args: [v.metadataUri, metadataHash],
            value: parseEther(v.amount),
          });
          await publicClient.waitForTransactionReceipt({ hash });
          setTxHash(hash);
          return;
        }

        if (!isAddress(v.token)) {
          setError("Invalid ERC20 token address.");
          return;
        }
        // ERC20 flow requires approve first; in MVP user does it manually (or we add approve UI later).
        setError("ERC20 create requires token approve() UI; use ETH for now.");
      }),
    [chainId, escrow, handleSubmit, publicClient, writeContractAsync],
  );

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Create bounty</h1>

      <form
        onSubmit={onSubmit}
        className="space-y-3 rounded-lg border bg-white p-4"
      >
        <div className="space-y-1">
          <div className="text-sm font-medium">Title</div>
          <Input {...register("title")} placeholder="Fix critical bug in…" />
        </div>
        <div className="space-y-1">
          <div className="text-sm font-medium">Description</div>
          <textarea
            className="w-full rounded-md border border-neutral-300 bg-white px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-black"
            rows={6}
            {...register("description")}
            placeholder="What needs to be done…"
          />
        </div>
        <div className="grid gap-3 md:grid-cols-2">
          <div className="space-y-1">
            <div className="text-sm font-medium">Metadata URI</div>
            <Input {...register("metadataUri")} />
          </div>
          <div className="space-y-1">
            <div className="text-sm font-medium">Amount (ETH)</div>
            <Input {...register("amount")} />
          </div>
        </div>
        <div className="text-xs text-neutral-500">
          ERC20 (USDC/USDT) требует approve — добавим в следующей итерации.
        </div>

        {error ? <div className="text-sm text-red-600">{error}</div> : null}
        {txHash ? (
          <div className="text-sm text-neutral-700">
            Created. Tx: <span className="font-mono">{txHash}</span>
          </div>
        ) : null}

        {Object.keys(errors).length > 0 && (
          <div className="text-sm text-red-600">
            {Object.entries(errors).map(([key, val]) => (
              <div key={key}>
                {key}: {val?.message}
              </div>
            ))}
          </div>
        )}

        <Button type="submit" disabled={isSubmitting}>
          Create & deposit
        </Button>
      </form>
    </div>
  );
}
