"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";

type DifficultyResp = {
  level: string;
  multipliers: {
    BurnMultiplier: number;
    ChurnMultiplier: number;
    FundingChanceMult: number;
    EventSeverityMult: number;
    StartingCashMult: number;
    AcquisitionRateMult: number;
  };
};

const LEVELS = ["easy", "normal", "hard", "brutal", "custom"] as const;
const TONE: Record<string, string> = {
  easy: "text-emerald-400",
  normal: "text-sky-400",
  hard: "text-amber-400",
  brutal: "text-rose-400",
  custom: "text-violet-400",
};

export function DifficultySelector() {
  const [data, setData] = useState<DifficultyResp | null>(null);
  const [error, setError] = useState<string | null>(null);

  function refresh() {
    const token = getToken();
    if (!token) return;
    api
      .get<DifficultyResp>("/api/v1/companies/me/difficulty", { token })
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) setData(null);
      });
  }
  useEffect(refresh, []);

  function setLevel(level: string) {
    const token = getToken();
    if (!token) return;
    setError(null);
    api
      .post<DifficultyResp>("/api/v1/companies/me/difficulty", { level }, { token })
      .then(setData)
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Failed."));
  }

  if (!data) return null;
  const m = data.multipliers;

  return (
    <section className="mt-6 rounded-xl border border-border bg-surface p-4">
      <h3 className="text-sm font-semibold text-foreground">Difficulty</h3>
      <div className="mt-3 flex flex-wrap gap-2">
        {LEVELS.map((lvl) => (
          <button
            key={lvl}
            onClick={() => setLevel(lvl)}
            className={`rounded-md px-3 py-1.5 text-xs font-semibold uppercase ${
              data.level === lvl
                ? "bg-brand text-surface-muted"
                : "border border-border text-muted hover:bg-background"
            }`}
          >
            {lvl}
          </button>
        ))}
      </div>
      <div className="mt-3 grid grid-cols-2 gap-x-4 gap-y-1 text-xs text-muted sm:grid-cols-3">
        <span>Burn ×{m.BurnMultiplier.toFixed(1)}</span>
        <span>Churn ×{m.ChurnMultiplier.toFixed(1)}</span>
        <span>Funding ×{m.FundingChanceMult.toFixed(1)}</span>
        <span>Events ×{m.EventSeverityMult.toFixed(1)}</span>
        <span>Start cash ×{m.StartingCashMult.toFixed(2)}</span>
        <span>Acquisition ×{m.AcquisitionRateMult.toFixed(1)}</span>
      </div>
      <p className={`mt-2 text-xs font-medium ${TONE[data.level] ?? "text-muted"}`}>
        Current: {data.level}
      </p>
      {error && <p className="mt-2 text-xs text-rose-400">{error}</p>}
    </section>
  );
}
