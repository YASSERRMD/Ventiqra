"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import type { Competitor } from "@/lib/types";

export function CompetitorPanel() {
  const [competitors, setCompetitors] = useState<Competitor[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<Competitor[]>("/api/v1/companies/me/competitors", { token })
      .then(setCompetitors)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) {
          setCompetitors([]);
        } else {
          setError(err instanceof Error ? err.message : "Could not load competitors.");
        }
      });
  }, []);

  if (error) return <p className="text-sm text-rose-400">{error}</p>;
  if (!competitors) return null;
  if (competitors.length === 0) return null;

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Competitors</h3>
      <p className="mt-1 text-xs text-muted">
        Rivals erode your acquisition. Their pressure scales with strength.
      </p>
      <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {competitors.map((c) => {
          const strength = Math.max(0, Math.min(100, c.strength));
          return (
            <div key={c.id} className="rounded-xl border border-border bg-surface p-4">
              <p className="text-sm font-semibold text-foreground">{c.name}</p>
              <p className="text-xs text-muted">
                Market share {(c.market_share * 100).toFixed(1)}%
              </p>
              <div className="mt-3">
                <div className="flex items-center justify-between text-xs text-muted">
                  <span>Strength</span>
                  <span>{strength}</span>
                </div>
                <div className="mt-1 h-2 w-full overflow-hidden rounded-full bg-background">
                  <div
                    className="h-full rounded-full bg-rose-500"
                    style={{ width: `${strength}%` }}
                  />
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}
