"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import type { Achievement } from "@/lib/types";

const ALL_KEYS = [
  { key: "first_launch", name: "First Launch", icon: "🚀" },
  { key: "first_funding", name: "First Funding", icon: "💰" },
  { key: "profitability", name: "Profitable", icon: "📈" },
  { key: "unicorn", name: "Unicorn", icon: "🦄" },
  { key: "hired_ten", name: "Growing Team", icon: "👥" },
  { key: "thousand_customers", name: "Viral", icon: "🔥" },
];

export function AchievementsPanel() {
  const [awarded, setAwarded] = useState<Achievement[] | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<Achievement[]>("/api/v1/companies/me/achievements", { token })
      .then(setAwarded)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) setAwarded(null);
      });
  }, []);

  if (!awarded) return null;
  const awardedKeys = new Set(awarded.map((a) => a.key));

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">
        Achievements <span className="text-xs font-normal text-muted">· {awarded.length}/{ALL_KEYS.length}</span>
      </h3>
      <div className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-3">
        {ALL_KEYS.map((a) => {
          const got = awardedKeys.has(a.key);
          const detail = awarded.find((x) => x.key === a.key);
          return (
            <div
              key={a.key}
              className={`rounded-xl border p-4 text-center ${
                got ? "border-emerald-500/40 bg-emerald-500/10" : "border-border bg-surface opacity-60"
              }`}
            >
              <div className="text-2xl">{got ? a.icon : "🔒"}</div>
              <p className={`mt-1 text-xs font-semibold ${got ? "text-emerald-400" : "text-muted"}`}>{a.name}</p>
              {got && detail && <p className="text-xs text-muted/70">day {detail.awarded_day}</p>}
            </div>
          );
        })}
      </div>
    </section>
  );
}
