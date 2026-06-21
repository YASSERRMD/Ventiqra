"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { CustomScenario, CustomScenarioInput } from "@/lib/types";

const DIFFICULTIES: CustomScenario["difficulty"][] = ["easy", "normal", "hard", "brutal"];

const EMPTY: CustomScenarioInput = {
  name: "",
  description: "",
  difficulty: "normal",
  industry: "",
  starting_cash_cents: 1_000_000_00,
  starting_burn_cents: 100_000_00,
  market_tam: 50_000,
  market_growth_rate: 0.08,
  market_trend: 1.0,
};

export function CustomScenarioEditor() {
  const [saved, setSaved] = useState<CustomScenario[] | null>(null);
  const [form, setForm] = useState<CustomScenarioInput>(EMPTY);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function refresh() {
    const token = getToken();
    if (!token) return;
    api
      .get<CustomScenario[]>("/api/v1/scenarios/custom", { token })
      .then(setSaved)
      .catch(() => setSaved([]));
  }

  useEffect(refresh, []);

  function set<K extends keyof CustomScenarioInput>(key: K, value: CustomScenarioInput[K]) {
    setForm((f) => ({ ...f, [key]: value }));
  }

  function reset() {
    setForm(EMPTY);
    setEditingId(null);
    setError(null);
  }

  function edit(s: CustomScenario) {
    setEditingId(s.id);
    setForm({
      name: s.name,
      description: s.description,
      difficulty: s.difficulty,
      industry: s.industry,
      starting_cash_cents: s.starting_cash_cents,
      starting_burn_cents: s.starting_burn_cents,
      market_tam: s.market.tam,
      market_growth_rate: s.market.growth_rate,
      market_trend: s.market.trend_multiplier,
    });
    setError(null);
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      if (editingId) {
        await api.patch<CustomScenario>(`/api/v1/scenarios/custom/${editingId}`, form, { token });
      } else {
        await api.post<CustomScenario>("/api/v1/scenarios/custom", form, { token });
      }
      reset();
      refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save scenario.");
    } finally {
      setBusy(false);
    }
  }

  async function remove(id: string) {
    const token = getToken();
    if (!token) return;
    if (!confirm("Delete this custom scenario?")) return;
    try {
      await api.del(`/api/v1/scenarios/custom/${id}`, { token });
      if (editingId === id) reset();
      refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not delete scenario.");
    }
  }

  async function apply(id: string) {
    const token = getToken();
    if (!token) return;
    setBusy(true);
    setError(null);
    try {
      await api.post(`/api/v1/scenarios/custom/${id}/apply`, undefined, { token });
      window.location.href = "/";
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not apply scenario.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Custom scenarios</h3>
      <p className="mt-1 text-xs text-muted">
        Author your own starting configuration. Cash must be $1,000–$50,000,000; TAM 1,000–10,000,000.
      </p>

      <form
        onSubmit={submit}
        className="mt-4 grid grid-cols-1 gap-3 rounded-xl border border-border bg-surface p-5 sm:grid-cols-2"
      >
        <label className="flex flex-col gap-1 text-sm sm:col-span-2">
          <span className="text-muted">Name *</span>
          <input
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.name}
            maxLength={80}
            onChange={(e) => set("name", e.target.value)}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Industry</span>
          <input
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.industry}
            maxLength={60}
            onChange={(e) => set("industry", e.target.value)}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Difficulty</span>
          <select
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.difficulty}
            onChange={(e) => set("difficulty", e.target.value as CustomScenario["difficulty"])}
          >
            {DIFFICULTIES.map((d) => (
              <option key={d} value={d}>
                {d}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Starting cash ({formatCents(form.starting_cash_cents)})</span>
          <input
            type="number"
            min={100_000}
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.starting_cash_cents}
            onChange={(e) => set("starting_cash_cents", Number(e.target.value))}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Monthly burn ({formatCents(form.starting_burn_cents)})</span>
          <input
            type="number"
            min={10_000}
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.starting_burn_cents}
            onChange={(e) => set("starting_burn_cents", Number(e.target.value))}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Market TAM</span>
          <input
            type="number"
            min={1000}
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.market_tam}
            onChange={(e) => set("market_tam", Number(e.target.value))}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Growth rate (0–0.5)</span>
          <input
            type="number"
            step={0.01}
            min={0}
            max={0.5}
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.market_growth_rate}
            onChange={(e) => set("market_growth_rate", Number(e.target.value))}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm sm:col-span-2">
          <span className="text-muted">Trend multiplier (0.5–2.0)</span>
          <input
            type="number"
            step={0.05}
            min={0.5}
            max={2}
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.market_trend}
            onChange={(e) => set("market_trend", Number(e.target.value))}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm sm:col-span-2">
          <span className="text-muted">Description</span>
          <textarea
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={form.description}
            maxLength={1000}
            rows={2}
            onChange={(e) => set("description", e.target.value)}
          />
        </label>

        {error && <p className="text-sm text-rose-400 sm:col-span-2">{error}</p>}

        <div className="flex gap-2 sm:col-span-2">
          <button
            type="submit"
            disabled={busy}
            className="rounded-md bg-brand px-4 py-2 text-sm font-semibold text-surface-muted disabled:opacity-60"
          >
            {busy ? "Saving…" : editingId ? "Update scenario" : "Save scenario"}
          </button>
          {editingId && (
            <button
              type="button"
              onClick={reset}
              className="rounded-md border border-border px-4 py-2 text-sm text-muted hover:bg-background"
            >
              Cancel
            </button>
          )}
        </div>
      </form>

      {saved && saved.length > 0 && (
        <ul className="mt-4 space-y-2">
          {saved.map((s) => (
            <li key={s.id} className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-surface px-4 py-3">
              <div>
                <p className="text-sm font-semibold text-foreground">{s.name}</p>
                <p className="text-xs text-muted">
                  {s.difficulty} · {s.industry || "—"} · {formatCents(s.starting_cash_cents)} · TAM{" "}
                  {s.market.tam.toLocaleString("en-US")}
                </p>
              </div>
              <div className="flex gap-2">
                <button onClick={() => edit(s)} className="rounded-md border border-border px-3 py-1 text-xs text-muted hover:bg-background">
                  Edit
                </button>
                <button
                  onClick={() => apply(s.id)}
                  disabled={busy}
                  className="rounded-md bg-brand/80 px-3 py-1 text-xs font-semibold text-surface-muted hover:bg-brand disabled:opacity-50"
                >
                  Apply
                </button>
                <button onClick={() => remove(s.id)} className="rounded-md border border-rose-500/40 px-3 py-1 text-xs text-rose-400 hover:bg-rose-500/10">
                  Delete
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
