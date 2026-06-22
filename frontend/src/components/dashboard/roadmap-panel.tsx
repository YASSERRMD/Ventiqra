"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import type { RoadmapFeature } from "@/lib/types";

const STATUS_TONE: Record<RoadmapFeature["status"], string> = {
  backlog: "text-muted",
  developing: "text-amber-400",
  shipped: "text-emerald-400",
};

export function RoadmapPanel() {
  const [features, setFeatures] = useState<RoadmapFeature[] | null>(null);
  const [name, setName] = useState("");
  const [priority, setPriority] = useState(0);
  const [value, setValue] = useState(10);
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  function refresh() {
    const token = getToken();
    if (!token) return;
    api
      .get<RoadmapFeature[]>("/api/v1/companies/me/features", { token })
      .then(setFeatures)
      .catch(() => setFeatures([]));
  }
  useEffect(refresh, []);

  function create(e: React.FormEvent) {
    e.preventDefault();
    const token = getToken();
    if (!token || !name) return;
    setError(null);
    api
      .post<RoadmapFeature>("/api/v1/companies/me/features", { name, priority, value_points: value }, { token })
      .then(() => {
        setName("");
        refresh();
      })
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Could not create feature."));
  }

  function develop(id: string) {
    const token = getToken();
    if (!token) return;
    setBusy(id);
    setError(null);
    api
      .post(`/api/v1/companies/me/features/${id}/develop`, { points: 25 }, { token })
      .then(() => refresh())
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Develop failed."))
      .finally(() => setBusy(null));
  }

  function remove(id: string) {
    const token = getToken();
    if (!token) return;
    api.del(`/api/v1/companies/me/features/${id}`, { token }).then(refresh).catch(() => {});
  }

  if (!features) return null;
  const shippedValue = features.filter((f) => f.status === "shipped").reduce((s, f) => s + f.value_points, 0);

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">
        Roadmap <span className="text-xs font-normal text-muted">· {shippedValue} value shipped</span>
      </h3>

      <form onSubmit={create} className="mt-3 flex flex-wrap items-end gap-2 rounded-xl border border-border bg-surface p-4">
        <label className="flex flex-col gap-1 text-xs text-muted">
          Feature name
          <input className="rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={name} maxLength={80} onChange={(e) => setName(e.target.value)} />
        </label>
        <label className="flex flex-col gap-1 text-xs text-muted">
          Priority
          <input type="number" className="w-20 rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={priority} onChange={(e) => setPriority(Number(e.target.value))} />
        </label>
        <label className="flex flex-col gap-1 text-xs text-muted">
          Value pts
          <input type="number" className="w-20 rounded-md border border-border bg-background px-3 py-1.5 text-foreground" value={value} onChange={(e) => setValue(Number(e.target.value))} />
        </label>
        <button type="submit" disabled={!name} className="rounded-md bg-brand px-4 py-1.5 text-xs font-semibold text-surface-muted disabled:opacity-50">
          Add
        </button>
      </form>

      {error && <p className="mt-2 text-sm text-rose-400">{error}</p>}

      {features.length > 0 && (
        <ul className="mt-3 space-y-2">
          {features.map((f) => (
            <li key={f.id} className="rounded-xl border border-border bg-surface px-4 py-3">
              <div className="flex items-center justify-between gap-2">
                <p className="text-sm font-semibold text-foreground">{f.name}</p>
                <span className={`text-xs font-medium uppercase ${STATUS_TONE[f.status]}`}>{f.status}</span>
              </div>
              <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-background">
                <div className="h-full rounded-full bg-brand" style={{ width: `${f.progress}%` }} />
              </div>
              <div className="mt-2 flex items-center justify-between text-xs text-muted">
                <span>priority {f.priority} · {f.value_points} value pts · {f.progress}%</span>
                <div className="flex gap-2">
                  {f.status !== "shipped" && (
                    <button onClick={() => develop(f.id)} disabled={busy === f.id} className="rounded-md bg-brand/80 px-3 py-1 font-semibold text-surface-muted hover:bg-brand disabled:opacity-50">
                      {busy === f.id ? "…" : "Develop +25"}
                    </button>
                  )}
                  <button onClick={() => remove(f.id)} className="rounded-md border border-rose-500/40 px-3 py-1 text-rose-400 hover:bg-rose-500/10">
                    Delete
                  </button>
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
