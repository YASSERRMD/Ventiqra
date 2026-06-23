"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { LeaderboardEntry } from "@/lib/types";
import { PageHeader } from "@/components/layout/page-header";

const OUTCOME_TONE: Record<string, string> = {
  bankrupt: "text-rose-400",
  thriving: "text-emerald-400",
  acquired: "text-sky-400",
};

export default function LeaderboardPage() {
  const [entries, setEntries] = useState<LeaderboardEntry[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<LeaderboardEntry[]>("/api/v1/leaderboard", { token })
      .then(setEntries)
      .catch((err: unknown) => {
        if (err instanceof ApiError) setError(err.message);
        setEntries([]);
      });
  }, []);

  return (
    <div>
      <PageHeader title="Leaderboard" subtitle="Local high scores across all your runs." />
      {error && <p className="mb-4 text-sm text-rose-400">{error}</p>}
      {!entries ? (
        <p className="text-sm text-muted">Loading…</p>
      ) : entries.length === 0 ? (
        <p className="text-sm text-muted">
          No scores yet. Play a company to bankruptcy (or thrive) to land on the board.
        </p>
      ) : (
        <table className="w-full overflow-hidden rounded-xl border border-border bg-surface text-sm">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted">
              <th className="px-4 py-2">#</th>
              <th className="px-4 py-2">Company</th>
              <th className="px-4 py-2">Score</th>
              <th className="px-4 py-2">Days</th>
              <th className="px-4 py-2">Peak valuation</th>
              <th className="px-4 py-2">Outcome</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((e, i) => (
              <tr key={e.id} className="border-b border-border/50 last:border-0">
                <td className="px-4 py-2 text-muted">{i + 1}</td>
                <td className="px-4 py-2 font-semibold text-foreground">{e.company_name}</td>
                <td className="px-4 py-2 font-mono text-brand">{e.score.toLocaleString("en-US")}</td>
                <td className="px-4 py-2 text-muted">{e.days_survived}</td>
                <td className="px-4 py-2 text-muted">{formatCents(e.peak_valuation)}</td>
                <td className={`px-4 py-2 text-xs font-medium uppercase ${OUTCOME_TONE[e.outcome] ?? "text-muted"}`}>
                  {e.outcome}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
