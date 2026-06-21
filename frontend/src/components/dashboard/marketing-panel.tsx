"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Marketing } from "@/lib/types";

export function MarketingPanel() {
  const [m, setM] = useState<Marketing | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<Marketing>("/api/v1/companies/me/marketing", { token })
      .then(setM)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) {
          setM(null);
        }
      });
  }, []);

  if (!m) return null;

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Marketing</h3>
      <p className="mt-1 text-xs text-muted">
        Set a budget in Finance. Spend drives conversions and CAC with diminishing returns.
      </p>
      <div className="mt-3 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs uppercase tracking-wide text-muted">Monthly budget</p>
          <p className="mt-2 text-xl font-semibold text-foreground">{formatCents(m.monthly_budget_cents)}</p>
        </div>
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs uppercase tracking-wide text-muted">Daily conversions</p>
          <p className="mt-2 text-xl font-semibold text-emerald-400">{m.daily_conversions}</p>
        </div>
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs uppercase tracking-wide text-muted">CAC (monthly)</p>
          <p className="mt-2 text-xl font-semibold text-foreground">
            {m.cac_cents > 0 ? formatCents(m.cac_cents) : "—"}
          </p>
        </div>
      </div>
      <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {m.channels.map((c) => {
          const w = Math.max(0, Math.min(100, c.weight * 100));
          return (
            <div key={c.name} className="rounded-xl border border-border bg-surface p-4">
              <p className="text-sm font-medium capitalize text-foreground">{c.name}</p>
              <p className="text-xs text-muted">Conv {(c.conversion * 100).toFixed(0)}%</p>
              <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-background">
                <div className="h-full rounded-full bg-brand" style={{ width: `${w}%` }} />
              </div>
              <p className="mt-1 text-xs text-muted">{w.toFixed(0)}% of mix</p>
            </div>
          );
        })}
      </div>
    </section>
  );
}
