"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Timeline } from "@/lib/types";
import { EmptyState } from "@/components/dashboard/panel-states";

const KIND_TONE: Record<string, string> = {
  milestone: "text-indigo-400",
  launch: "text-emerald-400",
  funding: "text-sky-400",
  decision: "text-amber-400",
  crisis: "text-rose-400",
  event: "text-muted",
  reputation: "text-violet-400",
};

export function TimelinePanel() {
  const [tl, setTl] = useState<Timeline | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<Timeline>("/api/v1/companies/me/timeline", { token })
      .then(setTl)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) {
          setTl(null);
        }
      });
  }, []);

  if (!tl) return null;
  if (tl.entries.length === 0) {
    return (
      <section className="mt-6 sm:mt-8">
        <h3 className="text-base font-semibold text-foreground">Timeline</h3>
        <div className="mt-3">
          <EmptyState
            title="No timeline events yet"
            message="Milestones like founding, launches, and funding rounds will appear here as you play."
          />
        </div>
      </section>
    );
  }

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Timeline</h3>
      <p className="mt-1 text-xs text-muted">
        Company history, newest first. Now on day {tl.day}.
      </p>

      <ol className="mt-3 space-y-2">
        {tl.entries.slice(0, 12).map((e) => (
          <li key={e.id} className="rounded-xl border border-border bg-surface px-4 py-3">
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm font-semibold text-foreground">{e.title}</p>
              <span className={`text-xs font-medium uppercase ${KIND_TONE[e.kind] ?? "text-muted"}`}>
                {e.kind}
              </span>
            </div>
            {e.description && <p className="mt-1 text-xs text-muted">{e.description}</p>}
            <p className="mt-1 text-xs text-muted/70">Day {e.sim_day}</p>
          </li>
        ))}
      </ol>

      {tl.monthly_summary.length > 0 && (
        <div className="mt-5">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted">Monthly summary</p>
          <ul className="mt-2 divide-y divide-border rounded-xl border border-border bg-surface">
            {tl.monthly_summary.slice(-6).map((m) => (
              <li key={m.month} className="flex items-center justify-between px-4 py-2 text-xs">
                <span className="text-muted">
                  Month {m.month} · days {m.start_day}–{m.end_day}
                </span>
                <span className="flex gap-4 text-muted">
                  <span>events {m.events_count}</span>
                  <span className="text-muted/70">burn {formatCents(m.burn_end)}/mo</span>
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  );
}
