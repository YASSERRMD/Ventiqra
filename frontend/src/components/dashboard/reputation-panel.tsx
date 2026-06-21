"use client";

import { useEffect, useMemo, useState } from "react";
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import type { Reputation } from "@/lib/types";

export function ReputationPanel() {
  const [rep, setRep] = useState<Reputation | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<Reputation>("/api/v1/companies/me/reputation", { token })
      .then(setRep)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) {
          setRep(null);
        }
      });
  }, []);

  // Build a cumulative score-over-time series from events (oldest → newest),
  // anchored at the current score.
  const series = useMemo(() => {
    if (!rep || rep.events.length === 0) return [];
    const ordered = [...rep.events].reverse(); // oldest first (API is newest-first)
    let running = rep.score;
    const points = [{ label: "now", score: running }];
    for (let i = ordered.length - 1; i >= 0; i--) {
      running -= ordered[i].delta;
      points.push({ label: `d${ordered[i].sim_day}`, score: running });
    }
    return points.reverse();
  }, [rep]);

  if (!rep) return null;

  const tone =
    rep.score >= 70 ? "text-emerald-400" : rep.score >= 40 ? "text-amber-400" : "text-rose-400";

  return (
    <section className="mt-8">
      <div className="flex items-baseline justify-between">
        <h3 className="text-base font-semibold text-foreground">Brand reputation</h3>
        <p className={`text-sm font-semibold ${tone}`}>{rep.score}/100</p>
      </div>
      <p className="mt-1 text-xs text-muted">
        Reputation modulates acquisition ({(rep.growth_multiplier * 100).toFixed(0)}% growth). Launches
        and funding lift it; failures damage it.
      </p>

      {series.length >= 2 && (
        <div className="mt-3 h-48 w-full rounded-xl border border-border bg-surface p-3">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={series} margin={{ top: 5, right: 10, bottom: 0, left: -20 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
              <XAxis dataKey="label" tick={{ fill: "#94a3b8", fontSize: 11 }} />
              <YAxis domain={[0, 100]} tick={{ fill: "#94a3b8", fontSize: 11 }} />
              <Tooltip
                contentStyle={{
                  background: "#0f172a",
                  border: "1px solid #1e293b",
                  borderRadius: 8,
                  fontSize: 12,
                }}
                labelStyle={{ color: "#e2e8f0" }}
              />
              <Line type="monotone" dataKey="score" stroke="#6366f1" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {rep.events.length > 0 && (
        <ul className="mt-3 divide-y divide-border rounded-xl border border-border bg-surface">
          {rep.events.slice(0, 6).map((e) => (
            <li key={e.id} className="flex items-center justify-between px-4 py-2 text-xs">
              <span className="text-muted">
                <span className="text-foreground">{e.event}</span> · day {e.sim_day}
              </span>
              <span className={e.delta >= 0 ? "text-emerald-400" : "text-rose-400"}>
                {e.delta >= 0 ? "+" : ""}{e.delta}
              </span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
