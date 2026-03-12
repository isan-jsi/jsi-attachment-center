import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import { setTokenGetter } from "@/api/client";

interface AuthState {
  token: string | null;
  username: string | null;
}

interface AuthContextValue {
  isAuthenticated: boolean;
  username: string | null;
  token: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const STORAGE_KEY = "ibs_doc_engine_auth";

function loadStoredAuth(): AuthState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw);
      return { token: parsed.token ?? null, username: parsed.username ?? null };
    }
  } catch {
    // ignore
  }
  return { token: null, username: null };
}

function saveAuth(state: AuthState) {
  if (state.token) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } else {
    localStorage.removeItem(STORAGE_KEY);
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [auth, setAuth] = useState<AuthState>(loadStoredAuth);

  useEffect(() => {
    setTokenGetter(() => auth.token);
  }, [auth.token]);

  useEffect(() => {
    saveAuth(auth);
  }, [auth]);

  const login = useCallback(async (username: string, password: string) => {
    const res = await fetch("/api/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => null);
      throw new Error(body?.error ?? "Login failed");
    }

    const data = await res.json();
    setAuth({ token: data.token ?? data.data?.token, username });
  }, []);

  const logout = useCallback(() => {
    setAuth({ token: null, username: null });
  }, []);

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated: !!auth.token,
        username: auth.username,
        token: auth.token,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
