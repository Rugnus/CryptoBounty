import { useQuery } from "@tanstack/react-query";
import { api } from "../../lib/api";
import { useAuthStore } from "../../lib/authStore";

type NotificationItem = {
  id: number;
  kind: string;
  payload: { raw: string };
  createdAt: string;
  readAt?: string | null;
};

export function NotificationsPage() {
  const token = useAuthStore((s) => s.token);
  const { data, isLoading, error } = useQuery({
    queryKey: ["notifications"],
    queryFn: async () => {
      const res = await api.get<{ items: NotificationItem[] }>("/notifications");
      return res.data.items;
    },
    enabled: Boolean(token)
  });

  if (!token) {
    return <div className="text-sm text-neutral-600">Sign in (SIWE) to view inbox.</div>;
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Notifications</h1>
      {isLoading ? <div className="text-sm text-neutral-600">Loading…</div> : null}
      {error ? <div className="text-sm text-red-600">Failed to load.</div> : null}
      <div className="grid gap-2">
        {(data ?? []).map((n) => (
          <div key={n.id} className="rounded-lg border bg-white p-3">
            <div className="flex items-center justify-between">
              <div className="font-medium">{n.kind}</div>
              <div className="text-xs text-neutral-500">{n.createdAt}</div>
            </div>
            <div className="mt-2 text-xs font-mono break-all text-neutral-700">
              {n.payload.raw}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

