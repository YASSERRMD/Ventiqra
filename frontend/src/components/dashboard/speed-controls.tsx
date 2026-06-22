"use client";

import { useEffect, useRef, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import type { SimControl } from "@/lib/types";

const SPEEDS = [1, 5, 30] as const;

/**
 * SpeedControls renders the simulation transport: pause/play, speed selection
 * (1x/5x/30x), and a manual step button. In auto mode it polls the tick
 * endpoint at the chosen speed, stopping on unmount, pause, or error.
 */
export function SpeedControls() {
  const [control, setControl] = useState<SimControl | null>(null);
  const [error, setError] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  async function refresh() {
    const token = getToken();
    if (!token) return;
    try {
      const c = await api.get<SimControl>("/api/v1/companies/me/sim/control", { token });
      setControl(c);
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.status === 404)) setControl(null);
    }
  }

  useEffect(() => {
    // Initial load: fetch the current control state once on mount.
    let cancelled = false;
    const token = getToken();
    if (!token) return;
    api
      .get<SimControl>("/api/v1/companies/me/sim/control", { token })
      .then((c) => {
        if (!cancelled) setControl(c);
      })
      .catch((err: unknown) => {
        if (!cancelled && err instanceof ApiError && (err.status === 401 || err.status === 404)) {
          setControl(null);
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);;

  // Drive auto-mode polling. When mode === 'auto', schedule ticks at the
  // interval implied by the speed. Each tick dispatches a window event so other
  // panels refresh; we avoid calling setState here to keep the effect clean.
  useEffect(() => {
    if (!control || control.mode !== "auto") {
      if (pollRef.current) clearTimeout(pollRef.current);
      return;
    }
    const token = getToken();
    if (!token) return;
    const intervalMs = 1000 / control.speed;

    const loop = () => {
      api
        .post("/api/v1/companies/me/sim/tick", undefined, { token })
        .then(() => {
          window.dispatchEvent(new CustomEvent("ventiqra:tick"));
        })
        .catch((err: unknown) => {
          if (err instanceof ApiError) setError(err.message);
        })
        .finally(() => {
          pollRef.current = setTimeout(loop, intervalMs);
        });
    };
    pollRef.current = setTimeout(loop, intervalMs);

    return () => {
      if (pollRef.current) clearTimeout(pollRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [control?.mode, control?.speed]);

  async function act(path: string, body?: Record<string, unknown>) {
    const token = getToken();
    if (!token) return;
    setError(null);
    try {
      await api.post(`/api/v1/companies/me/sim/${path}`, body, { token });
      await refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Action failed.");
    }
  }

  async function step() {
    const token = getToken();
    if (!token) return;
    setError(null);
    try {
      await api.post("/api/v1/companies/me/sim/tick", undefined, { token });
      window.dispatchEvent(new CustomEvent("ventiqra:tick"));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Step failed.");
    }
  }

  if (!control) return null;
  const running = control.mode === "auto";

  return (
    <div className="flex flex-wrap items-center gap-2 rounded-xl border border-border bg-surface px-3 py-2">
      <button
        onClick={() => act(running ? "pause" : "resume")}
        className={`rounded-md px-3 py-1.5 text-xs font-semibold ${
          running ? "bg-amber-500/20 text-amber-400" : "bg-emerald-500/20 text-emerald-400"
        }`}
      >
        {running ? "⏸ Pause" : "▶ Play"}
      </button>

      <div className="flex items-center gap-1">
        {SPEEDS.map((sp) => (
          <button
            key={sp}
            onClick={() => act("speed", { speed: sp })}
            className={`rounded-md px-2.5 py-1 text-xs ${
              control.speed === sp
                ? "bg-brand text-surface-muted"
                : "border border-border text-muted hover:bg-background"
            }`}
          >
            {sp}x
          </button>
        ))}
      </div>

      <button
        onClick={step}
        disabled={running}
        className="rounded-md border border-border px-3 py-1.5 text-xs text-muted hover:bg-background disabled:opacity-40"
      >
        ⏭ Step
      </button>

      {error && <span className="text-xs text-rose-400">{error}</span>}
    </div>
  );
}
