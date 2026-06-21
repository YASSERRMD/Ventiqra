"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { LoadResult, SaveSlot } from "@/lib/types";
import { PageHeader } from "@/components/layout/page-header";

export default function SavesPage() {
  const [slots, setSlots] = useState<SaveSlot[] | null>(null);
  const [hasCompany, setHasCompany] = useState<boolean | null>(null);
  const [slotName, setSlotName] = useState("");
  const [label, setLabel] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  function refresh() {
    const token = getToken();
    if (!token) return;
    api
      .get<SaveSlot[]>("/api/v1/saves", { token })
      .then(setSlots)
      .catch(() => setSlots([]));
    api
      .get("/api/v1/companies/me", { token })
      .then(() => setHasCompany(true))
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) setHasCompany(false);
      });
  }

  useEffect(refresh, []);

  function save() {
    const token = getToken();
    if (!token) return;
    if (!/^[a-z0-9_-]{1,32}$/.test(slotName)) {
      setError("Slot name must be 1-32 lowercase letters, digits, - or _.");
      return;
    }
    setBusy(true);
    setError(null);
    setMessage(null);
    api
      .post<SaveSlot>("/api/v1/saves", { slot: slotName, label }, { token })
      .then(() => {
        setBusy(false);
        setSlotName("");
        setLabel("");
        setMessage(`Saved to slot "${slotName}".`);
        refresh();
      })
      .catch((err: unknown) => {
        setBusy(false);
        setError(err instanceof ApiError ? err.message : "Could not save.");
      });
  }

  function load(slot: string) {
    const token = getToken();
    if (!token) return;
    if (!confirm(`Load slot "${slot}"? Current unsaved progress will be lost.`)) return;
    setBusy(true);
    setError(null);
    api
      .post<LoadResult>(`/api/v1/saves/${slot}/load`, undefined, { token })
      .then(() => {
        setBusy(false);
        setMessage(`Loaded slot "${slot}".`);
        refresh();
      })
      .catch((err: unknown) => {
        setBusy(false);
        setError(err instanceof ApiError ? err.message : "Could not load.");
      });
  }

  function remove(slot: string) {
    const token = getToken();
    if (!token) return;
    if (!confirm(`Delete slot "${slot}"?`)) return;
    api
      .del(`/api/v1/saves/${slot}`, { token })
      .then(refresh)
      .catch((err: unknown) => setError(err instanceof ApiError ? err.message : "Could not delete."));
  }

  return (
    <div>
      <PageHeader
        title="Saves"
        subtitle="Capture your run into a named slot and restore it later. Up to 5 slots."
      />

      {hasCompany === false && (
        <div className="mb-6 rounded-xl border border-amber-500/40 bg-amber-500/10 p-4 text-sm text-amber-300">
          You need a company before you can save or load.
        </div>
      )}
      {error && <p className="mb-4 text-sm text-rose-400">{error}</p>}
      {message && <p className="mb-4 text-sm text-emerald-400">{message}</p>}

      <div className="flex max-w-md flex-col gap-3 rounded-xl border border-border bg-surface p-5">
        <p className="text-sm font-semibold text-foreground">New save</p>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Slot name (lowercase, digits, - or _)</span>
          <input
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={slotName}
            maxLength={32}
            placeholder="e.g. run-1"
            onChange={(e) => setSlotName(e.target.value)}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Label (optional)</span>
          <input
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={label}
            maxLength={80}
            placeholder="e.g. before series A"
            onChange={(e) => setLabel(e.target.value)}
          />
        </label>
        <button
          onClick={save}
          disabled={busy || hasCompany === false || !slotName}
          className="rounded-md bg-brand px-4 py-2 text-sm font-semibold text-surface-muted disabled:opacity-50"
        >
          {busy ? "Saving…" : "Save"}
        </button>
      </div>

      {slots && slots.length > 0 && (
        <ul className="mt-6 space-y-2">
          {slots.map((s) => (
            <li
              key={s.id}
              className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-surface px-4 py-3"
            >
              <div>
                <p className="text-sm font-semibold text-foreground">{s.slot}</p>
                <p className="text-xs text-muted">
                  {s.label ? `${s.label} · ` : ""}Day {s.day} · {formatCents(s.cash_cents)} · {s.status}
                </p>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => load(s.slot)}
                  disabled={busy}
                  className="rounded-md bg-brand/80 px-3 py-1 text-xs font-semibold text-surface-muted hover:bg-brand disabled:opacity-50"
                >
                  Load
                </button>
                <button
                  onClick={() => remove(s.slot)}
                  className="rounded-md border border-rose-500/40 px-3 py-1 text-xs text-rose-400 hover:bg-rose-500/10"
                >
                  Delete
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
