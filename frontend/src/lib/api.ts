import { apiBaseUrl } from "@/lib/config";

// ApiError wraps a non-2xx API response with status, message, and body.
export class ApiError extends Error {
  readonly status: number;
  readonly body: unknown;

  constructor(status: number, message: string, body: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

export type QueryValue = string | number | boolean | undefined;

export type RequestOptions = {
  body?: unknown;
  query?: Record<string, QueryValue>;
  headers?: Record<string, string>;
  signal?: AbortSignal;
  token?: string;
  // Next.js fetch caching options.
  cache?: RequestCache;
  revalidate?: number;
};

export type HealthResponse = {
  status: string;
  service: string;
  env: string;
  version?: string;
  timestamp: string;
  checks?: Record<string, unknown>;
};

function joinPath(base: string, path: string): string {
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const cleanBase = base.replace(/\/+$/, "");
  const cleanPath = path.startsWith("/") ? path : `/${path}`;
  return `${cleanBase}${cleanPath}`;
}

function buildUrl(base: string, path: string, query?: RequestOptions["query"]): string {
  const url = new URL(joinPath(base, path));
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined) url.searchParams.set(key, String(value));
    }
  }
  return url.toString();
}

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null;
}

async function safeJson(res: Response): Promise<unknown> {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return { raw: text };
  }
}

// ApiClient is a thin, typed wrapper around fetch for the Ventiqra backend.
export class ApiClient {
  constructor(private readonly baseUrl: string = apiBaseUrl()) {}

  async request<T>(
    method: string,
    path: string,
    options: RequestOptions = {},
  ): Promise<T> {
    const headers: Record<string, string> = {
      Accept: "application/json",
      ...options.headers,
    };
    if (options.body !== undefined) headers["Content-Type"] = "application/json";
    if (options.token) headers.Authorization = `Bearer ${options.token}`;

    const init: RequestInit & { cache?: RequestCache; next?: { revalidate?: number } } = {
      method,
      headers,
      body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
      signal: options.signal,
    };
    if (options.cache !== undefined) init.cache = options.cache;
    if (options.revalidate !== undefined) init.next = { revalidate: options.revalidate };

    const res = await fetch(buildUrl(this.baseUrl, path, options.query), init);
    const raw = await safeJson(res);

    if (!res.ok) {
      const message =
        (isRecord(raw) && typeof raw.error === "string" && raw.error) ||
        (isRecord(raw) && typeof raw.message === "string" && raw.message) ||
        res.statusText ||
        `request failed with ${res.status}`;
      throw new ApiError(res.status, message, raw);
    }

    return raw as T;
  }

  get<T>(path: string, options?: RequestOptions) {
    return this.request<T>("GET", path, options);
  }
  post<T>(path: string, body?: unknown, options?: RequestOptions) {
    return this.request<T>("POST", path, { ...options, body });
  }
  put<T>(path: string, body?: unknown, options?: RequestOptions) {
    return this.request<T>("PUT", path, { ...options, body });
  }
  patch<T>(path: string, body?: unknown, options?: RequestOptions) {
    return this.request<T>("PATCH", path, { ...options, body });
  }
  del<T>(path: string, options?: RequestOptions) {
    return this.request<T>("DELETE", path, options);
  }

  health(signal?: AbortSignal) {
    return this.get<HealthResponse>("/healthz", { signal });
  }
}

// Singleton instance used across the app.
export const api = new ApiClient();
