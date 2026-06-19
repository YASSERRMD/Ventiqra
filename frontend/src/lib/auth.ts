import { api, type HealthResponse } from "@/lib/api";

// Browser localStorage-based token storage. The token is a JWT issued by the
// backend and sent as `Authorization: Bearer <token>` via the API client.

const TOKEN_KEY = "ventiqra.token";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(TOKEN_KEY);
}

export type CurrentUser = {
  id: string;
  email: string;
  name: string;
};

export type AuthResult = {
  user: CurrentUser;
  token: string;
  expires_at: string;
};

export async function login(email: string, password: string): Promise<AuthResult> {
  const res = await api.post<AuthResult>("/api/v1/auth/login", { email, password });
  setToken(res.token);
  return res;
}

export async function register(
  email: string,
  password: string,
  name: string,
): Promise<AuthResult> {
  const res = await api.post<AuthResult>("/api/v1/auth/register", {
    email,
    password,
    name,
  });
  setToken(res.token);
  return res;
}

export async function logout(): Promise<void> {
  clearToken();
}

export async function fetchCurrentUser(): Promise<CurrentUser> {
  const token = getToken();
  if (!token) throw new Error("not authenticated");
  const res = await api.get<{ user: CurrentUser }>("/api/v1/me", { token });
  return res.user;
}

// Re-export for convenience.
export type { HealthResponse };
