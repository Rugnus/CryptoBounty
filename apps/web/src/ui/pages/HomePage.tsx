import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";

import { api } from "../../lib/api";
import { Input } from "../components/Input";

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

export function HomePage() {
  const [q, setQ] = useState("");
  const queryKey = useMemo(() => ["bounties", q], [q]);

  const { data, isLoading, error } = useQuery({
    queryKey,
    queryFn: async () => {
      const res = await api.get<{ items: Bounty[] }>("/bounties", { params: { q } });
      return res.data.items;
    }
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-xl font-semibold">Bounties</h1>
        <div className="w-full max-w-sm">
          <Input
            placeholder="Search (by metadata URI)"
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </div>
      </div>

      {isLoading ? <div className="text-sm text-neutral-600">Loading…</div> : null}
      {error ? (
        <div className="text-sm text-red-600">Failed to load bounties.</div>
      ) : null}

      <div className="grid gap-3">
        {(data ?? []).map((b) => (
          <Link
            key={`${b.chainId}:${b.bountyId}`}
            to={`/bounties/${b.bountyId}`}
            className="rounded-lg border bg-white p-4 hover:bg-neutral-50"
          >
            <div className="flex items-center justify-between gap-3">
              <div className="font-medium">#{b.bountyId}</div>
              <div className="text-xs rounded bg-neutral-100 px-2 py-1">{b.status}</div>
            </div>
            <div className="mt-2 text-sm text-neutral-700">
              Sponsor: <span className="font-mono">{b.sponsor}</span>
            </div>
            <div className="mt-1 text-sm text-neutral-700">
              Token: <span className="font-mono">{b.token}</span> Amount:{" "}
              <span className="font-mono">{b.amount}</span>
            </div>
            <div className="mt-1 text-xs text-neutral-500 truncate">{b.metadataUri}</div>
          </Link>
        ))}
      </div>
    </div>
  );
}

