"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { EnterpriseContract } from "@/lib/types";

const STATUS_TONE: Record<EnterpriseContract["status"], string> = {
  active: "text-emerald-400",
  renewed: "text-sky-400",
  churned: "text-rose-400",
};

export function ContractsPanel() {
  const [contracts, setContracts] = useState<EnterpriseContract[] | null>(null);
  const [customer, setCustomer] = useState("");
  const [annual, setAnnual] = useState(1_200_000_00);
  const [discount, setDiscount] = useState(0);
  const [years, setYears] = useState(1);
  const [error, setError] = useState<string | null>(null);

  function refresh() {
    const token = getToken();
    if (!token) return;
    api.get<EnterpriseContract[]>("/api/v1/companies/me/contracts", { token }).then(setContracts).catch(() => setContracts([]));
  }
  useEffect(refresh, []);

  function sign(e: React.FormEvent) {
    e.preventDefault();
    const token = getToken();
    if (!token || !customer) return;
    setError(null);
    api
      .post<EnterpriseContract>("/api/v1/companies/me/contracts", {
        customer_name: customer, annual_value: annual, discount_pct: discount, term_years: years,
      }, { token })
      .then(() => {
        setCustomer("");
        refresh();
      })
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Sign failed."));
  }

  if (!contracts) return null;
  const active = contracts.filter((c) => c.status === "active");
  const arr = active.reduce((s, c) => s + c.annual_value, 0);

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">
        Enterprise contracts <span className="text-xs font-normal text-muted">· {formatCents(arr)}/yr active</span>
      </h3>

      <form onSubmit={sign} className="mt-3 flex flex-wrap items-end gap-2 rounded-xl border border-border bg-surface p-4">
        <label className="flex flex-col gap-1 text-xs text-muted">
          Customer
          <input className="rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={customer} maxLength={80} onChange={(e) => setCustomer(e.target.value)} />
        </label>
        <label className="flex flex-col gap-1 text-xs text-muted">
          Annual value (cents)
          <input type="number" className="w-32 rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={annual} onChange={(e) => setAnnual(Number(e.target.value))} />
        </label>
        <label className="flex flex-col gap-1 text-xs text-muted">
          Discount %
          <input type="number" min={0} max={50} className="w-20 rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={discount} onChange={(e) => setDiscount(Number(e.target.value))} />
        </label>
        <label className="flex flex-col gap-1 text-xs text-muted">
          Term (years)
          <input type="number" min={1} max={5} className="w-20 rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={years} onChange={(e) => setYears(Number(e.target.value))} />
        </label>
        <button type="submit" disabled={!customer} className="rounded-md bg-brand px-4 py-1.5 text-xs font-semibold text-surface-muted disabled:opacity-50">
          Sign
        </button>
      </form>

      {error && <p className="mt-2 text-sm text-rose-400">{error}</p>}

      {contracts.length > 0 && (
        <ul className="mt-3 space-y-2">
          {contracts.map((c) => (
            <li key={c.id} className="rounded-xl border border-border bg-surface px-4 py-3">
              <div className="flex items-center justify-between gap-2">
                <p className="text-sm font-semibold text-foreground">{c.customer_name}</p>
                <span className={`text-xs font-medium uppercase ${STATUS_TONE[c.status]}`}>{c.status}</span>
              </div>
              <p className="mt-1 text-xs text-muted">
                {formatCents(c.annual_value)}/yr · {c.discount_pct}% discount · {c.remaining_days}/{c.term_days} days left
              </p>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
