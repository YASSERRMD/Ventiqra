"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Infrastructure } from "@/lib/types";

export function InfrastructurePanel() {
  const [inf, setInf] = useState<Infrastructure | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function refresh() {
    const token = getToken();
    if (!token) return;
    api
      .get<Infrastructure>("/api/v1/companies/me/infrastructure", { token })
      .then(setInf)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) setInf(null);
      });
  }
  useEffect(refresh, []);

  function scale() {
    const token = getToken();
    if (!token) return;
    setBusy(true);
    setError(null);
    api
      .post<Infrastructure>("/api/v1/companies/me/infrastructure/scale", undefined, { token })
      .then(setInf)
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Scale-up failed."))
      .finally(() => setBusy(false));
  }

  if (!inf) return null;
  const loadPct = Math.round(inf.load_ratio * 100);
  const riskPct = Math.round(inf.outage_risk * 100);
  const loadTone = loadPct < 70 ? "bg-emerald-500" : loadPct < 90 ? "bg-amber-500" : "bg-rose-500";

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Infrastructure</h3>
      <div className="mt-3 rounded-xl border border-border bg-surface p-4">
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <Metric label="Tier" value={`${inf.tier}/10`} />
          <Metric label="Capacity" value={inf.capacity.toLocaleString("en-US")} />
          <Metric label="Customers" value={inf.customers.toLocaleString("en-US")} />
          <Metric label="Hosting/mo" value={formatCents(inf.hosting_cost_cents)} />
        </div>

        <div className="mt-4">
          <div className="flex items-center justify-between text-xs text-muted">
            <span>Load</span>
            <span>{loadPct}% · outage risk {riskPct}%</span>
          </div>
          <div className="mt-1 h-2 w-full overflow-hidden rounded-full bg-background">
            <div className={`h-full rounded-full ${loadTone}`} style={{ width: `${Math.min(loadPct, 100)}%` }} />
          </div>
        </div>

        {error && <p className="mt-3 text-sm text-rose-400">{error}</p>}
        {loadPct >= 90 && (
          <p className="mt-3 text-xs text-rose-400">⚠ Near capacity — outages likely. Scale up to add headroom.</p>
        )}

        <button
          onClick={scale}
          disabled={busy || inf.tier >= 10}
          className="mt-4 rounded-md bg-brand px-4 py-2 text-xs font-semibold text-surface-muted disabled:opacity-50"
        >
          {busy ? "Scaling…" : `Scale up (${formatCents(inf.scale_up_cost_cents)})`}
        </button>
      </div>
    </section>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-muted">{label}</p>
      <p className="text-sm font-semibold text-foreground">{value}</p>
    </div>
  );
}
