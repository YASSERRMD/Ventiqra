"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import type { SupportState } from "@/lib/types";

export function SupportPanel() {
  const [sup, setSup] = useState<SupportState | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<SupportState>("/api/v1/companies/me/support", { token })
      .then(setSup)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) setSup(null);
      });
  }, []);

  if (!sup) return null;
  const overloaded = sup.open_tickets > 100 && sup.support_agents === 0;

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Customer support</h3>
      <div className="mt-3 rounded-xl border border-border bg-surface p-4">
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <Metric label="Open tickets" value={String(sup.open_tickets)} />
          <Metric label="Resolved total" value={String(sup.resolved_total)} />
          <Metric label="Support agents" value={String(sup.support_agents)} />
          <Metric label="Sat. penalty" value={`-${sup.satisfaction_penalty}`} />
        </div>
        {overloaded && (
          <p className="mt-3 text-xs text-rose-400">
            ⚠ Ticket backlog is large and there are no support agents. Hire support staff to clear tickets and protect satisfaction.
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
