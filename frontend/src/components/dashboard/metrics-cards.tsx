"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents, formatNumber } from "@/lib/format";
import type { Metrics } from "@/lib/types";

type State =
  | { kind: "loading" }
  | { kind: "unauth" }
  | { kind: "no-company" }
  | { kind: "ready"; metrics: Metrics }
  | { kind: "error"; message: string };

type Card = { label: string; value: string; hint: string };

const PLACEHOLDER_CARDS: Card[] = [
  { label: "Cash", value: "—", hint: "Available capital" },
  { label: "Revenue", value: "—", hint: "Per-day run rate" },
  { label: "Burn/mo", value: "—", hint: "Monthly spend" },
  { label: "Valuation", value: "—", hint: "Estimated worth" },
  { label: "Runway", value: "—", hint: "Months until broke" },
];

export function MetricsCards() {
  // Lazy initializer mirrors src/app/company/page.tsx: derive the initial auth
  // state from localStorage without calling setState synchronously in an effect.
  const [state, setState] = useState<State>(() =>
    getToken() ? { kind: "loading" } : { kind: "unauth" },
  );

  useEffect(() => {
    if (state.kind !== "loading") return;
    const token = getToken();
    if (!token) return;

    api
      .get<Metrics>("/api/v1/companies/me/metrics", { token })
      .then((metrics) => setState({ kind: "ready", metrics }))
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setState({ kind: "unauth" });
        } else if (err instanceof ApiError && err.status === 404) {
          setState({ kind: "no-company" });
        } else {
          setState({
            kind: "error",
            message: err instanceof Error ? err.message : "Could not load metrics.",
          });
        }
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (state.kind === "unauth" || state.kind === "no-company") {
    return (
      <div className="rounded-xl border border-dashed border-border bg-surface/40 p-8 text-center">
        <p className="text-sm text-muted">
          No company yet. Found a company in the{" "}
          <Link href="/company" className="text-brand hover:underline">
            Company
          </Link>{" "}
          section to begin the simulation.
        </p>
      </div>
    );
  }

  if (state.kind === "error") {
    return (
      <div className="rounded-xl border border-border bg-surface p-6 text-sm text-rose-400">
        {state.message}
      </div>
    );
  }

  if (state.kind === "loading") {
    return <MetricGrid cards={PLACEHOLDER_CARDS} />;
  }

  const m = state.metrics;
  const runway =
    m.runway_months < 0 ? "\u221e" : `${m.runway_months.toFixed(1)} mo`;

  const cards: Card[] = [
    { label: "Cash", value: formatCents(m.cash_cents), hint: "Available capital" },
    { label: "Revenue", value: formatCents(m.revenue_cents), hint: "Per-day run rate" },
    { label: "Burn/mo", value: formatCents(m.burn_cents_per_month), hint: "Monthly spend" },
    { label: "Valuation", value: formatCents(m.valuation_cents), hint: "Estimated worth" },
    { label: "Runway", value: runway, hint: "Months until broke" },
  ];

  return (
    <div>
      <MetricGrid cards={cards} />
      <p className="mt-3 text-xs text-muted/80">
        Simulated day {formatNumber(m.day)}
      </p>
    </div>
  );
}

function MetricGrid({ cards }: { cards: Card[] }) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
      {cards.map((c) => (
        <div key={c.label} className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs uppercase tracking-wide text-muted">{c.label}</p>
          <p className="mt-2 text-2xl font-semibold text-foreground">{c.value}</p>
          <p className="mt-1 text-xs text-muted/80">{c.hint}</p>
        </div>
      ))}
    </div>
  );
}
