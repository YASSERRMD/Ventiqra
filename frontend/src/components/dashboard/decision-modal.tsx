"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { DecisionOutcome, PendingDecision } from "@/lib/types";

// A small delta pill that shows a signed effect with the right tone.
function Delta({ value, label }: { value: number; label: string }) {
  if (value === 0) return null;
  const tone = value >= 0 ? "text-emerald-400" : "text-rose-400";
  return (
    <span className={tone}>
      {label} {value >= 0 ? "+" : ""}
      {label === "cash" ? formatCents(value) : value}
    </span>
  );
}

export function DecisionModal() {
  const [decision, setDecision] = useState<PendingDecision | null>(null);
  const [outcome, setOutcome] = useState<DecisionOutcome | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Poll for a pending decision whenever no card is shown. The simulation
  // offers cards on a fixed cadence, so a light interval is enough.
  useEffect(() => {
    let cancelled = false;
    const token = getToken();
    if (!token) return;

    const check = () => {
      if (cancelled || decision) return;
      api
        .get<PendingDecision>("/api/v1/companies/me/decisions/pending", { token })
        .then((d) => {
          if (!cancelled) {
            setDecision(d);
            setOutcome(null);
            setError(null);
          }
        })
        .catch((err: unknown) => {
          // 404 just means no pending card; anything else is ignored here.
          if (err instanceof ApiError && err.status !== 404) {
            // ignore — keep polling
          }
        });
    };
    check();
    const id = setInterval(check, 5000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [decision]);

  if (!decision) return null;

  const resolve = (choiceId: string) => {
    const token = getToken();
    if (!token) return;
    setBusy(choiceId);
    setError(null);
    api
      .post<DecisionOutcome>(
        `/api/v1/companies/me/decisions/${decision.id}/resolve`,
        { choice_id: choiceId },
        { token },
      )
      .then((res) => {
        setOutcome(res);
        setBusy(null);
      })
      .catch((err: unknown) => {
        setBusy(null);
        if (err instanceof ApiError) {
          setError(err.message);
        } else {
          setError("Could not resolve decision.");
        }
      });
  };

  const close = () => {
    setDecision(null);
    setOutcome(null);
    setError(null);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-lg rounded-2xl border border-border bg-surface p-6 shadow-xl">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-indigo-400">
              {decision.category} · day {decision.sim_day_offered}
            </p>
            <h3 className="mt-1 text-lg font-semibold text-foreground">{decision.title}</h3>
          </div>
        </div>
        <p className="mt-2 text-sm text-muted">{decision.description}</p>

        {outcome ? (
          <div className="mt-5 rounded-xl border border-border bg-black/20 p-4">
            <p className="text-sm font-semibold text-foreground">
              {outcome.outcome === "success" ? "✅ It paid off" : "⚠️ It backfired"}
            </p>
            <div className="mt-2 flex flex-wrap gap-3 text-xs text-muted">
              <Delta value={outcome.applied_cash_delta} label="cash" />
              <Delta value={outcome.applied_reputation_delta} label="rep" />
              <Delta value={outcome.applied_morale_delta} label="morale" />
            </div>
            {outcome.recurring_cash_delta !== 0 && outcome.remaining_days > 0 && (
              <p className="mt-2 text-xs text-muted">
                Recurring {outcome.recurring_cash_delta >= 0 ? "+" : ""}
                {formatCents(outcome.recurring_cash_delta)}/day for {outcome.remaining_days} days.
              </p>
            )}
            <button
              onClick={close}
              className="mt-4 w-full rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
            >
              Continue
            </button>
          </div>
        ) : (
          <div className="mt-5 space-y-3">
            {decision.choices.map((c) => {
              const riskPct = Math.round((1 - c.success_chance) * 100);
              return (
                <button
                  key={c.id}
                  disabled={busy !== null}
                  onClick={() => resolve(c.id)}
                  className="block w-full rounded-xl border border-border bg-black/20 p-4 text-left transition hover:border-indigo-500/50 hover:bg-black/30 disabled:opacity-50"
                >
                  <div className="flex items-center justify-between gap-2">
                    <p className="text-sm font-semibold text-foreground">{c.label}</p>
                    {c.success_chance < 1 && (
                      <span className="text-xs font-medium text-amber-400">{riskPct}% risk</span>
                    )}
                  </div>
                  <p className="mt-1 text-xs text-muted">{c.description}</p>
                  <div className="mt-2 flex flex-wrap gap-3 text-xs text-muted">
                    {c.success_chance < 1 ? (
                      <>
                        <span className="text-emerald-400">
                          win: <Delta value={c.cash_delta} label="cash" />{" "}
                          <Delta value={c.reputation_delta} label="rep" />{" "}
                          <Delta value={c.morale_delta} label="morale" />
                        </span>
                        <span className="text-rose-400">
                          miss: <Delta value={c.fail_cash_delta} label="cash" />{" "}
                          <Delta value={c.fail_reputation_delta} label="rep" />{" "}
                          <Delta value={c.fail_morale_delta} label="morale" />
                        </span>
                      </>
                    ) : (
                      <>
                        <Delta value={c.cash_delta} label="cash" />
                        <Delta value={c.reputation_delta} label="rep" />
                        <Delta value={c.morale_delta} label="morale" />
                      </>
                    )}
                    {c.recurring_cash_delta !== 0 && c.duration_days > 0 && (
                      <span className="text-indigo-400">
                        {c.recurring_cash_delta >= 0 ? "+" : ""}
                        {formatCents(c.recurring_cash_delta)}/day × {c.duration_days}d
                      </span>
                    )}
                  </div>
                </button>
              );
            })}
            {error && <p className="text-xs text-rose-400">{error}</p>}
          </div>
        )}
      </div>
    </div>
  );
}
