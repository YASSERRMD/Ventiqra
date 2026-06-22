"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { TechDebt } from "@/lib/types";

function debtTone(debt: number): string {
  if (debt < 40) return "bg-emerald-500";
  if (debt < 70) return "bg-amber-500";
  return "bg-rose-500";
}

export function EngineeringPanel() {
  const [td, setTd] = useState<TechDebt | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function refresh() {
    const token = getToken();
    if (!token) return;
    api
      .get<TechDebt>("/api/v1/companies/me/tech-debt", { token })
      .then(setTd)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) setTd(null);
      });
  }
  useEffect(refresh, []);

  function refactor() {
    const token = getToken();
    if (!token) return;
    setBusy(true);
    setError(null);
    api
      .post<TechDebt>("/api/v1/companies/me/tech-debt/refactor", undefined, { token })
      .then(setTd)
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Refactor failed."))
      .finally(() => setBusy(false));
  }

  if (!td) return null;
  const riskPct = Math.round(td.outage_risk * 100);

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Engineering</h3>
      <div className="mt-3 rounded-xl border border-border bg-surface p-4">
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <Metric label="Code quality" value={`${td.quality}/100`} />
          <Metric label="Tech debt" value={`${td.debt}/100`} />
          <Metric label="Outage risk" value={`${riskPct}%`} />
          <Metric label="Refactors" value={String(td.refactors)} />
        </div>

        <div className="mt-4">
          <div className="flex items-center justify-between text-xs text-muted">
            <span>Debt</span>
            <span>{td.debt}/100</span>
          </div>
          <div className="mt-1 h-2 w-full overflow-hidden rounded-full bg-background">
            <div className={`h-full rounded-full ${debtTone(td.debt)}`} style={{ width: `${td.debt}%` }} />
          </div>
        </div>

        {error && <p className="mt-3 text-sm text-rose-400">{error}</p>}

        <button
          onClick={refactor}
          disabled={busy || td.debt === 0}
          className="mt-4 rounded-md bg-brand px-4 py-2 text-xs font-semibold text-surface-muted disabled:opacity-50"
        >
          {busy ? "Refactoring…" : `Refactor (${formatCents(20_000_00)})`}
        </button>
        {td.debt >= 70 && (
          <p className="mt-2 text-xs text-rose-400">
            ⚠ High debt raises outage risk. Consider refactoring before shipping more.
          </p>
        )}
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
