import { create } from "zustand";

type AuthState = {
  token: string | null;
  address: `0x${string}` | null;
  setAuth: (token: string, address: `0x${string}`) => void;
  logout: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem("cb_jwt"),
  address: (localStorage.getItem("cb_addr") as `0x${string}` | null) ?? null,
  setAuth: (token, address) => {
    localStorage.setItem("cb_jwt", token);
    localStorage.setItem("cb_addr", address);
    set({ token, address });
  },
  logout: () => {
    localStorage.removeItem("cb_jwt");
    localStorage.removeItem("cb_addr");
    set({ token: null, address: null });
  }
}));

