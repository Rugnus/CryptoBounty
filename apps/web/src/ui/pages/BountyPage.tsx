import { useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { keccak256, toBytes, parseEther } from "viem";
import { useAccount, useChainId, usePublicClient, useWriteContract } from "wagmi";

import { api } from "../../lib/api";
import { Button } from "../components/Button";
import { Input } from "../components/Input";
import { bountyEscrowAbi, contracts } from "@cryptobounty/shared";

type Bounty = {
  chainId: number;
  bountyId: string;
  sponsor: string;
  token: string;
  amount: string;
  metadataUri: string;
  metadataHash: string;
  status: string;
  hunter?: string;
};

export function BountyPage() {
  const { id } = useParams();
  const { address } = useAccount();
  const chainId = useChainId();
  const escrow = (contracts as any)[chainId]?.bountyEscrow as `0x${string}` | undefined;
  const publicClient = usePublicClient();
  const { writeContractAsync } = useWriteContract();

  const [messageUri, setMessageUri] = useState("ipfs://apply");
  const [assignAddr, setAssignAddr] = useState("");
  const [workUri, setWorkUri] = useState("ipfs://work");
  const [disputeFeeEth, setDisputeFeeEth] = useState("0.01");
  const [tx, setTx] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  const { data } = useQuery({
    queryKey: ["bounty", id],
    queryFn: async () => {
      const res = await api.get<Bounty>(`/bounties/${id}`);
      return res.data;
    },
    enabled: Boolean(id)
  });

  const canWrite = useMemo(() => {
    return Boolean(escrow && escrow !== "0x0000000000000000000000000000000000000000");
  }, [escrow]);

  async function runTx(fn: () => Promise<`0x${string}`>) {
    setErr(null);
    setTx(null);
    if (!publicClient) {
      setErr("No public client");
      return;
    }
    try {
      const h = await fn();
      await publicClient.waitForTransactionReceipt({ hash: h });
      setTx(h);
    } catch (e: any) {
      setErr(e?.shortMessage ?? e?.message ?? "Tx failed");
    }
  }

  if (!data) return <div className="text-sm text-neutral-600">Loading…</div>;

  return (
    <div className="space-y-4">
      <div className="rounded-lg border bg-white p-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-xl font-semibold">Bounty #{data.bountyId}</h1>
          <div className="text-xs rounded bg-neutral-100 px-2 py-1">{data.status}</div>
        </div>
        <div className="mt-2 text-sm text-neutral-700">
          Sponsor: <span className="font-mono">{data.sponsor}</span>
        </div>
        <div className="mt-1 text-sm text-neutral-700">
          Token: <span className="font-mono">{data.token}</span> Amount:{" "}
          <span className="font-mono">{data.amount}</span>
        </div>
        <div className="mt-1 text-xs text-neutral-500 truncate">{data.metadataUri}</div>

        {!canWrite ? (
          <div className="mt-3 text-sm text-red-600">
            Escrow address not configured for this chain.
          </div>
        ) : null}

        {err ? <div className="mt-3 text-sm text-red-600">{err}</div> : null}
        {tx ? (
          <div className="mt-2 text-sm text-neutral-700">
            Tx: <span className="font-mono">{tx}</span>
          </div>
        ) : null}
      </div>

      <div className="grid gap-3 md:grid-cols-2">
        <div className="rounded-lg border bg-white p-4 space-y-2">
          <div className="font-medium">Hunter actions</div>
          <div className="text-xs text-neutral-500">
            Connected: <span className="font-mono">{address ?? "—"}</span>
          </div>

          <div className="space-y-1">
            <div className="text-sm">Apply message URI</div>
            <Input value={messageUri} onChange={(e) => setMessageUri(e.target.value)} />
          </div>
          <Button
            disabled={!canWrite}
            onClick={() =>
              runTx(() =>
                writeContractAsync({
                  abi: bountyEscrowAbi,
                  address: escrow!,
                  functionName: "apply",
                  args: [BigInt(data.bountyId), messageUri]
                })
              )
            }
          >
            Apply (on-chain)
          </Button>

          <div className="space-y-1 pt-2">
            <div className="text-sm">Work URI</div>
            <Input value={workUri} onChange={(e) => setWorkUri(e.target.value)} />
          </div>
          <Button
            disabled={!canWrite}
            onClick={() =>
              runTx(() =>
                writeContractAsync({
                  abi: bountyEscrowAbi,
                  address: escrow!,
                  functionName: "submitWork",
                  args: [BigInt(data.bountyId), workUri, keccak256(toBytes(workUri))]
                })
              )
            }
          >
            Submit work
          </Button>
        </div>

        <div className="rounded-lg border bg-white p-4 space-y-2">
          <div className="font-medium">Sponsor actions</div>
          <div className="space-y-1">
            <div className="text-sm">Assign hunter address</div>
            <Input value={assignAddr} onChange={(e) => setAssignAddr(e.target.value)} />
          </div>
          <Button
            disabled={!canWrite}
            onClick={() =>
              runTx(() =>
                writeContractAsync({
                  abi: bountyEscrowAbi,
                  address: escrow!,
                  functionName: "assignHunter",
                  args: [BigInt(data.bountyId), assignAddr as `0x${string}`]
                })
              )
            }
          >
            Assign hunter
          </Button>

          <Button
            variant="secondary"
            disabled={!canWrite}
            onClick={() =>
              runTx(() =>
                writeContractAsync({
                  abi: bountyEscrowAbi,
                  address: escrow!,
                  functionName: "approve",
                  args: [BigInt(data.bountyId)]
                })
              )
            }
          >
            Approve
          </Button>

          <Button
            disabled={!canWrite}
            onClick={() =>
              runTx(() =>
                writeContractAsync({
                  abi: bountyEscrowAbi,
                  address: escrow!,
                  functionName: "payout",
                  args: [BigInt(data.bountyId)]
                })
              )
            }
          >
            Payout
          </Button>

          <div className="pt-2 border-t" />
          <Button
            variant="secondary"
            disabled={!canWrite}
            onClick={() =>
              runTx(() =>
                writeContractAsync({
                  abi: bountyEscrowAbi,
                  address: escrow!,
                  functionName: "cancelBySponsor",
                  args: [BigInt(data.bountyId)]
                })
              )
            }
          >
            Cancel (Created only)
          </Button>

          <div className="space-y-1">
            <div className="text-sm">Dispute fee (ETH)</div>
            <Input value={disputeFeeEth} onChange={(e) => setDisputeFeeEth(e.target.value)} />
          </div>
          <Button
            variant="secondary"
            disabled={!canWrite}
            onClick={() =>
              runTx(() =>
                writeContractAsync({
                  abi: bountyEscrowAbi,
                  address: escrow!,
                  functionName: "rejectAndDispute",
                  args: [BigInt(data.bountyId)],
                  value: parseEther(disputeFeeEth)
                })
              )
            }
          >
            Reject & dispute
          </Button>

          <Button
            variant="secondary"
            disabled={!canWrite}
            onClick={() =>
              runTx(() =>
                writeContractAsync({
                  abi: bountyEscrowAbi,
                  address: escrow!,
                  functionName: "refund",
                  args: [BigInt(data.bountyId)]
                })
              )
            }
          >
            Refund (after ruling)
          </Button>
        </div>
      </div>
    </div>
  );
}

