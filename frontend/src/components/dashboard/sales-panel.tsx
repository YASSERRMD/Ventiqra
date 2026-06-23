"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Deal } from "@/lib/types";

const STAGE_ORDER: Deal["stage"][] = ["lead", "qualified", "proposal", "negotiation"];
const STAGE_TONE: Record<string, string> = {
  lead: "text-muted",
  qualified: "text-sky-400",
  proposal: "text-amber-400",
  negotiation: "text-violet-400",
  closed_won: "text-emerald-400",
  closed_lost: "text-rose-400",
};

export function SalesPanel() {
  const [deals, setDeals] = useState<Deal[] | null>(null);
  const [name, setName] = useState("");
  const [value, setValue] = useState(50_000_00);
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  function refresh() {
    const token = getToken();
    if (!token) return;
    api.get<Deal[]>("/api/v1/companies/me/deals", { token }).then(setDeals).catch(() => setDeals([]));
  }
  useEffect(refresh, []);

  function create(e: React.FormEvent) {
    e.preventDefault();
    const token = getToken();
    if (!token || !name) return;
    setError(null);
    api
      .post<Deal>("/api/v1/companies/me/deals", { name, value_cents: value }, { token })
      .then(() => {
        setName("");
        refresh();
      })
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Create failed."));
  }

  function advance(id: string) {
    const token = getToken();
    if (!token) return;
    setBusy(id);
    setError(null);
    api
      .post(`/api/v1/companies/me/deals/${id}/advance`, undefined, { token })
      .then(() => refresh())
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Advance failed."))
      .finally(() => setBusy(null));
  }

  if (!deals) return null;
  const open = deals.filter((d) => STAGE_ORDER.includes(d.stage));
  const won = deals.filter((d) => d.stage === "closed_won");
  const wonValue = won.reduce((s, d) => s + d.value_cents, 0);

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">
        Sales <span className="text-xs font-normal text-muted">· {won.length} won · {formatCents(wonValue)}</span>
      </h3>

      <form onSubmit={create} className="mt-3 flex flex-wrap items-end gap-2 rounded-xl border border-border bg-surface p-4">
        <label className="flex flex-col gap-1 text-xs text-muted">
          Deal name
          <input className="rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={name} maxLength={80} onChange={(e) => setName(e.target.value)} />
        </label>
        <label className="flex flex-col gap-1 text-xs text-muted">
          Value (cents)
          <input type="number" className="w-32 rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={value} onChange={(e) => setValue(Number(e.target.value))} />
        </label>
        <button type="submit" disabled={!name} className="rounded-md bg-brand px-4 py-1.5 text-xs font-semibold text-surface-muted disabled:opacity-50">
          Add deal
        </button>
      </form>

      {error && <p className="mt-2 text-sm text-rose-400">{error}</p>}

      {deals.length > 0 && (
        <ul className="mt-3 space-y-2">
          {deals.map((d) => {
            const isOpen = STAGE_ORDER.includes(d.stage);
            const stepIdx = STAGE_ORDER.indexOf(d.stage);
            return (
              <li key={d.id} className="rounded-xl border border-border bg-surface px-4 py-3">
                <div className="flex items-center justify-between gap-2">
                  <p className="text-sm font-semibold text-foreground">{d.name}</p>
                  <span className={`text-xs font-medium uppercase ${STAGE_TONE[d.stage]}`}>{d.stage.replace("_", " ")}</span>
                </div>
                <div className="mt-1 flex items-center justify-between text-xs text-muted">
                  <span>{formatCents(d.value_cents)} · {d.probability}% close</span>
                </div>
                {isOpen && (
                  <div className="mt-2 flex gap-1">
                    {STAGE_ORDER.map((st, i) => (
                      <div key={st} className={`h-1.5 flex-1 rounded-full ${i <= stepIdx ? "bg-brand" : "bg-background"}`} />
                    ))}
                  </div>
                )}
                {isOpen && (
                  <button onClick={() => advance(d.id)} disabled={busy === d.id} className="mt-2 rounded-md bg-brand/80 px-3 py-1 text-xs font-semibold text-surface-muted hover:bg-brand disabled:opacity-50">
                    {busy === d.id ? "…" : "Advance"}
                  </button>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
