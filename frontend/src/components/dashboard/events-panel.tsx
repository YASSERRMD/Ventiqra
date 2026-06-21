"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { GameEvent } from "@/lib/types";

const KIND_TONE: Record<GameEvent["kind"], string> = {
  positive: "border-emerald-500/40 bg-emerald-500/10",
  negative: "border-rose-500/40 bg-rose-500/10",
  neutral: "border-border bg-surface",
};

const KIND_LABEL: Record<GameEvent["kind"], string> = {
  positive: "text-emerald-400",
  negative: "text-rose-400",
  neutral: "text-muted",
};

export function EventsPanel() {
  const [events, setEvents] = useState<GameEvent[] | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<GameEvent[]>("/api/v1/companies/me/events", { token })
      .then(setEvents)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) {
          setEvents(null);
        }
      });
  }, []);

  if (!events || events.length === 0) return null;

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Event log</h3>
      <p className="mt-1 text-xs text-muted">Random events that shaped the company, newest first.</p>
      <ul className="mt-3 space-y-2">
        {events.slice(0, 10).map((e) => (
          <li key={e.id} className={`rounded-xl border px-4 py-3 ${KIND_TONE[e.kind]}`}>
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm font-semibold text-foreground">{e.title}</p>
              <span className={`text-xs font-medium uppercase ${KIND_LABEL[e.kind]}`}>{e.kind}</span>
            </div>
            <p className="mt-1 text-xs text-muted">{e.description}</p>
            <div className="mt-2 flex flex-wrap gap-3 text-xs text-muted">
              <span>Day {e.sim_day}</span>
              {e.cash_delta !== 0 && (
                <span className={e.cash_delta >= 0 ? "text-emerald-400" : "text-rose-400"}>
                  {formatCents(e.cash_delta)}
                </span>
              )}
              {e.reputation_delta !== 0 && (
                <span className={e.reputation_delta >= 0 ? "text-emerald-400" : "text-rose-400"}>
                  rep {e.reputation_delta >= 0 ? "+" : ""}{e.reputation_delta}
                </span>
              )}
              {e.morale_delta !== 0 && (
                <span className={e.morale_delta >= 0 ? "text-emerald-400" : "text-rose-400"}>
                  morale {e.morale_delta >= 0 ? "+" : ""}{e.morale_delta}
                </span>
              )}
            </div>
          </li>
        ))}
      </ul>
    </section>
  );
}
