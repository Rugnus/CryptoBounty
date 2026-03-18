import { Link, Route, Routes } from "react-router-dom";
import { ConnectButton } from "@rainbow-me/rainbowkit";
import { useAccount } from "wagmi";

import { useAuthStore } from "../lib/authStore";
import { Button } from "./components/Button";
import { HomePage } from "./pages/HomePage";
import { CreateBountyPage } from "./pages/CreateBountyPage";
import { BountyPage } from "./pages/BountyPage";
import { NotificationsPage } from "./pages/NotificationsPage";
import { useSIWE } from "./hooks/useSIWE";

export function App() {
  const { isConnected } = useAccount();
  const { token, address, logout } = useAuthStore();
  const { login, isLoading } = useSIWE();

  return (
    <div className="min-h-full bg-neutral-50 text-neutral-900">
      <header className="sticky top-0 z-10 border-b bg-white/80 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between gap-3 px-4 py-3">
          <div className="flex items-center gap-4">
            <Link to="/" className="font-semibold">
              CryptoBounty
            </Link>
            <Link to="/create" className="text-sm text-neutral-700 hover:text-black">
              Create
            </Link>
            <Link
              to="/notifications"
              className="text-sm text-neutral-700 hover:text-black"
            >
              Notifications
            </Link>
          </div>

          <div className="flex items-center gap-2">
            <ConnectButton />
            {isConnected && !token ? (
              <Button onClick={login} disabled={isLoading}>
                Sign-In (SIWE)
              </Button>
            ) : null}
            {token ? (
              <Button variant="secondary" onClick={logout} title={address ?? undefined}>
                Logout
              </Button>
            ) : null}
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-4 py-6">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/create" element={<CreateBountyPage />} />
          <Route path="/bounties/:id" element={<BountyPage />} />
          <Route path="/notifications" element={<NotificationsPage />} />
        </Routes>
      </main>
    </div>
  );
}

