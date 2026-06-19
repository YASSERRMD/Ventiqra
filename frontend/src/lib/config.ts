// Centralized environment configuration for the frontend.
//
// Next.js inlines NEXT_PUBLIC_* values at build time, so these are available on
// both the server and the client. Unknown/missing values fall back to sensible
// local development defaults.

type Env = {
  apiBaseUrl: string;
  wsBaseUrl: string;
  appEnv: string;
};

function readEnv(): Env {
  const apiBaseUrl = trim(process.env.NEXT_PUBLIC_API_URL) || "http://localhost:8080";
  const wsBaseUrl = trim(process.env.NEXT_PUBLIC_WS_URL) || "ws://localhost:8080/ws";
  const appEnv = trim(process.env.APP_ENV) || "development";

  if (!/^https?:\/\//.test(apiBaseUrl)) {
    throw new Error(`NEXT_PUBLIC_API_URL must be an http(s) URL, got: ${apiBaseUrl}`);
  }
  if (!/^wss?:\/\//.test(wsBaseUrl)) {
    throw new Error(`NEXT_PUBLIC_WS_URL must be a ws(s) URL, got: ${wsBaseUrl}`);
  }

  return { apiBaseUrl, wsBaseUrl, appEnv };
}

function trim(v: string | undefined): string {
  return (v ?? "").trim();
}

let cached: Env | null = null;

export function env(): Env {
  if (!cached) cached = readEnv();
  return cached;
}

export function apiBaseUrl(): string {
  return env().apiBaseUrl;
}

export function wsBaseUrl(): string {
  return env().wsBaseUrl;
}
