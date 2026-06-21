"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Finance, FundingSummary } from "@/lib/types";
import { PageHeader } from "@/components/layout/page-header";

type State =
  | { kind: "loading" }
  | { kind: "unauth" }
  | { kind: "no-company" }
  | { kind: "ready"; finance: Finance }
  | { kind: "error"; message: string };

export default function FinancePage() {
  const [state, setState] = useState<State>(() =>
    getToken() ? { kind: "loading" } : { kind: "unauth" },
  );

  useEffect(() => {
    if (state.kind !== "loading") return;
    void load(setState);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (state.kind === "loading") return <p className="text-sm text-muted">Loading…</p>;

  if (state.kind === "unauth") {
    return (
      <div>
        <PageHeader title="Finance" subtitle="Track cash, burn, runway, and funding." />
        <div className="rounded-xl border border-border bg-surface p-8 text-center text-sm text-muted">
          <Link href="/login" className="text-brand hover:underline">Log in</Link> to view finances.
        </div>
      </div>
    );
  }
  if (state.kind === "no-company") {
    return (
      <div>
        <PageHeader title="Finance" />
        <div className="rounded-xl border border-border bg-surface p-8 text-center text-sm text-muted">
          Create a <Link href="/company" className="text-brand hover:underline">company</Link> first.
        </div>
      </div>
    );
  }
  if (state.kind === "error") {
    return (
      <div>
        <PageHeader title="Finance" />
        <div className="rounded-xl border border-border bg-surface p-6 text-sm text-rose-400">{state.message}</div>
      </div>
    );
  }

  const f = state.finance;
  const profitable = f.profit_loss_cents >= 0;
  const rows: [string, number][] = [
    ["Base overhead", f.burn.base_cents],
    ["Salaries", f.burn.salary_cents],
    ["Infrastructure", f.burn.infra_cents],
    ["Marketing", f.burn.marketing_cents],
  ];

  return (
    <div>
      <PageHeader
        title="Finance"
        subtitle="Monthly profit & loss, burn breakdown, and marketing budget."
        action={<MarketingBudget key={f.marketing_budget_cents} current={f.marketing_budget_cents} onChanged={() => void load(setState)} />}
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Stat label="Monthly revenue" value={formatCents(f.monthly_revenue_cents)} tone="text-emerald-400" />
        <Stat label="Monthly burn" value={formatCents(f.burn.total_burn_cents)} tone="text-amber-400" />
        <Stat
          label={profitable ? "Monthly profit" : "Monthly loss"}
          value={formatCents(Math.abs(f.profit_loss_cents))}
          tone={profitable ? "text-emerald-400" : "text-rose-400"}
        />
      </div>

      <h3 className="mt-8 text-base font-semibold text-foreground">Burn breakdown</h3>
      <ul className="mt-3 divide-y divide-border rounded-xl border border-border bg-surface">
        {rows.map(([label, cents]) => (
          <li key={label} className="flex items-center justify-between px-4 py-3 text-sm">
            <span className="text-muted">{label}</span>
            <span className="text-foreground">{formatCents(cents)}</span>
          </li>
        ))}
        <li className="flex items-center justify-between border-t border-border bg-background/40 px-4 py-3 text-sm font-semibold">
          <span className="text-foreground">Total burn</span>
          <span className="text-foreground">{formatCents(f.burn.total_burn_cents)}</span>
        </li>
      </ul>
      <p className="mt-3 text-xs text-muted">Simulated day {f.day}.</p>

      <FundingSection />
    </div>
  );
}

function FundingSection() {
  const [summary, setSummary] = useState<FundingSummary | null>(null);
  const [amount, setAmount] = useState("1000000");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<FundingSummary>("/api/v1/companies/me/funding", { token })
      .then(setSummary)
      .catch(() => setSummary(null));
  }, []);

  async function reload() {
    const token = getToken();
    if (!token) return;
    try {
      const s = await api.get<FundingSummary>("/api/v1/companies/me/funding", { token });
      setSummary(s);
    } catch {
      setSummary(null);
    }
  }

  async function raise(e: React.FormEvent) {
    e.preventDefault();
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    const cents = Math.max(1, Math.round(Number(amount) || 0));
    setBusy(true);
    setError(null);
    try {
      await api.post("/api/v1/companies/me/funding/raise", { amount_cents: cents }, { token });
      await reload();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not raise funding.");
    } finally {
      setBusy(false);
    }
  }

  if (!summary) return null;

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Funding</h3>
      <div className="mt-3 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Stat label="Pre-money valuation" value={formatCents(summary.pre_money_cents)} tone="text-foreground" />
        <Stat label="Founder equity" value={`${summary.founder_equity_percent.toFixed(1)}%`} tone="text-foreground" />
        <Stat
          label="Investor interest"
          value={`${(summary.investor_interest * 100).toFixed(0)}%`}
          tone="text-emerald-400"
        />
      </div>

      <form
        onSubmit={raise}
        className="mt-4 flex max-w-xl items-end gap-3 rounded-xl border border-border bg-surface p-4"
      >
        <label className="flex flex-1 flex-col gap-1 text-sm">
          <span className="text-muted">Raise amount (cents)</span>
          <input
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            inputMode="numeric"
          />
        </label>
        <button
          type="submit"
          disabled={busy}
          className="rounded-md bg-brand px-4 py-2 text-sm font-semibold text-surface-muted disabled:opacity-60"
        >
          {busy ? "Raising…" : "Raise round"}
        </button>
        {error && <p className="basis-full text-sm text-rose-400">{error}</p>}
      </form>

      {summary.rounds.length > 0 && (
        <ul className="mt-4 divide-y divide-border rounded-xl border border-border bg-surface">
          {summary.rounds.map((r) => (
            <li key={r.id} className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-sm">
              <div>
                <p className="font-medium capitalize text-foreground">{r.round_name}</p>
                <p className="text-xs text-muted">Day {r.sim_day}</p>
              </div>
              <div className="text-right text-xs text-muted">
                <p>
                  <span className="text-foreground">{formatCents(r.amount_cents)}</span> for{" "}
                  <span className="text-foreground">{r.equity_percent.toFixed(1)}%</span>
                </p>
                <p>pre {formatCents(r.pre_money_cents)}</p>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

async function load(setState: (s: State) => void) {
  const token = getToken();
  if (!token) {
    setState({ kind: "unauth" });
    return;
  }
  try {
    const finance = await api.get<Finance>("/api/v1/companies/me/finance", { token });
    setState({ kind: "ready", finance });
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      clearToken();
      setState({ kind: "unauth" });
    } else if (err instanceof ApiError && err.status === 404) {
      setState({ kind: "no-company" });
    } else {
      setState({ kind: "error", message: err instanceof Error ? err.message : "Could not load finance." });
    }
  }
}

function Stat({ label, value, tone }: { label: string; value: string; tone: string }) {
  return (
    <div className="rounded-xl border border-border bg-surface p-5">
      <p className="text-xs text-muted">{label}</p>
      <p className={`mt-1 text-lg font-semibold ${tone}`}>{value}</p>
    </div>
  );
}

function MarketingBudget({ current, onChanged }: { current: number; onChanged: () => void }) {
  const [budget, setBudget] = useState(String(current));
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function save() {
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    const cents = Math.max(0, Math.round(Number(budget) || 0));
    setBusy(true);
    setError(null);
    try {
      await api.patch("/api/v1/companies/me/finance", { marketing_budget_cents: cents }, { token });
      onChanged();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not update budget.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex items-end gap-2 rounded-lg border border-border bg-surface px-3 py-2">
      <label className="flex flex-col gap-1 text-xs text-muted">
        <span>Marketing budget (cents/mo)</span>
        <input
          className="rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground outline-none focus:border-brand"
          value={budget}
          onChange={(e) => setBudget(e.target.value)}
          inputMode="numeric"
        />
      </label>
      <button
        type="button"
        onClick={save}
        disabled={busy}
        className="rounded-md bg-brand px-3 py-1.5 text-xs font-semibold text-surface-muted disabled:opacity-60"
      >
        {busy ? "Saving…" : "Set"}
      </button>
      {error && <p className="text-xs text-rose-400">{error}</p>}
    </div>
  );
}
