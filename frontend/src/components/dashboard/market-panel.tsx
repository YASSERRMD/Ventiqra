"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatNumber } from "@/lib/format";
import type { Market } from "@/lib/types";

export function MarketPanel() {
  const [market, setMarket] = useState<Market | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<Market>("/api/v1/companies/me/market", { token })
      .then(setMarket)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) {
          setMarket(null);
        }
      });
  }, []);

  if (!market) return null;

  const trendPct = (market.trend_multiplier - 1) * 100;

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Market</h3>
      <p className="mt-1 text-xs text-muted">
        The addressable market grows over time; the trend multiplier cycles demand.
      </p>
      <div className="mt-3 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs uppercase tracking-wide text-muted">Total addressable market</p>
          <p className="mt-2 text-xl font-semibold text-foreground">
            {formatNumber(market.tam)}
          </p>
        </div>
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs uppercase tracking-wide text-muted">Monthly growth</p>
          <p className="mt-2 text-xl font-semibold text-foreground">
            {(market.growth_rate * 100).toFixed(1)}%
          </p>
        </div>
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs uppercase tracking-wide text-muted">Trend multiplier</p>
          <p
            className={`mt-2 text-xl font-semibold ${trendPct >= 0 ? "text-emerald-400" : "text-rose-400"}`}
          >
            {market.trend_multiplier.toFixed(3)}× ({trendPct >= 0 ? "+" : ""}{trendPct.toFixed(1)}%)
          </p>
        </div>
      </div>
    </section>
  );
}
