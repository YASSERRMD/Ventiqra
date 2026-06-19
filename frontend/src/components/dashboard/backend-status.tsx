"use client";

import { useEffect, useState } from "react";
import { api, ApiError, type HealthResponse } from "@/lib/api";

type State =
  | { kind: "loading" }
  | { kind: "up"; health: HealthResponse }
  | { kind: "down"; message: string };

export function BackendStatus() {
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    const controller = new AbortController();
    let cancelled = false;

    api
      .health(controller.signal)
      .then((health) => {
        if (!cancelled) setState({ kind: "up", health });
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        const message =
          err instanceof ApiError ? err.message : err instanceof Error ? err.message : "unreachable";
        setState({ kind: "down", message });
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, []);

  const color =
    state.kind === "up"
      ? "text-brand"
      : state.kind === "down"
        ? "text-rose-400"
        : "text-muted";
  const dot =
    state.kind === "up"
      ? "bg-brand"
      : state.kind === "down"
        ? "bg-rose-400"
        : "bg-muted animate-pulse";

  return (
    <div className="flex items-center gap-2 rounded-lg border border-border bg-surface px-4 py-3 text-sm">
      <span className={`h-2.5 w-2.5 rounded-full ${dot}`} />
      <span className="text-muted">Backend:</span>
      <span className={`font-medium ${color}`}>
        {state.kind === "loading" && "checking…"}
        {state.kind === "up" && `up (${state.health.version ?? "v?"})`}
        {state.kind === "down" && "offline"}
      </span>
    </div>
  );
}
