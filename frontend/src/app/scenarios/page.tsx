"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { ApplyScenarioResult, Scenario } from "@/lib/types";
import { PageHeader } from "@/components/layout/page-header";
import { CustomScenarioEditor } from "@/components/dashboard/custom-scenario-editor";

const DIFFICULTY_TONE: Record<Scenario["difficulty"], string> = {
  easy: "text-emerald-400 border-emerald-500/40",
  normal: "text-sky-400 border-sky-500/40",
  hard: "text-amber-400 border-amber-500/40",
  brutal: "text-rose-400 border-rose-500/40",
};

export default function ScenariosPage() {
  const router = useRouter();
  const [scenarios, setScenarios] = useState<Scenario[] | null>(null);
  const [hasCompany, setHasCompany] = useState<boolean | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [applied, setApplied] = useState<Scenario | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<Scenario[]>("/api/v1/scenarios", { token })
      .then(setScenarios)
      .catch(() => setScenarios([]));
    api
      .get("/api/v1/companies/me", { token })
      .then(() => setHasCompany(true))
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) setHasCompany(false);
        else setHasCompany(null);
      });
  }, []);

  function apply(scenarioId: string) {
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    setBusy(scenarioId);
    setError(null);
    api
      .post<ApplyScenarioResult>(`/api/v1/scenarios/${scenarioId}/apply`, undefined, { token })
      .then((res) => {
        setApplied(res.scenario);
        setBusy(null);
        // Give the user a moment to read the result, then head to the dashboard.
        setTimeout(() => router.push("/"), 1200);
      })
      .catch((err: unknown) => {
        setBusy(null);
        setError(err instanceof ApiError ? err.message : "Could not apply scenario.");
      });
  }

  return (
    <div>
      <PageHeader
        title="Scenarios"
        subtitle="Choose a predefined starting scenario. Applying one resets your company's cash, industry, and market."
      />

      {hasCompany === false && (
        <div className="mb-6 rounded-xl border border-amber-500/40 bg-amber-500/10 p-4 text-sm text-amber-300">
          You need a company before you can apply a scenario.{" "}
          <Link href="/company" className="underline">
            Create one first
          </Link>
          .
        </div>
      )}

      {applied && (
        <div className="mb-6 rounded-xl border border-emerald-500/40 bg-emerald-500/10 p-4 text-sm text-emerald-300">
          Applied <span className="font-semibold">{applied.name}</span>. Redirecting to the dashboard…
        </div>
      )}

      {error && <p className="mb-4 text-sm text-rose-400">{error}</p>}

      {!scenarios ? (
        <p className="text-sm text-muted">Loading scenarios…</p>
      ) : scenarios.length === 0 ? (
        <p className="text-sm text-muted">No scenarios available.</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {scenarios.map((s) => {
            const isBusy = busy === s.id;
            return (
              <article
                key={s.id}
                className="flex flex-col rounded-2xl border border-border bg-surface p-5"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h3 className="text-base font-semibold text-foreground">{s.name}</h3>
                    <p className="text-xs uppercase tracking-wide text-muted">{s.category}</p>
                  </div>
                  <span
                    className={`rounded-full border px-2 py-0.5 text-xs font-medium uppercase ${DIFFICULTY_TONE[s.difficulty]}`}
                  >
                    {s.difficulty}
                  </span>
                </div>
                <p className="mt-3 flex-1 text-sm text-muted">{s.description}</p>

                <dl className="mt-4 grid grid-cols-2 gap-x-4 gap-y-2 text-xs">
                  <div>
                    <dt className="text-muted">Starting cash</dt>
                    <dd className="font-semibold text-foreground">{formatCents(s.starting_cash_cents)}</dd>
                  </div>
                  <div>
                    <dt className="text-muted">Monthly burn</dt>
                    <dd className="font-semibold text-foreground">{formatCents(s.starting_burn_cents)}</dd>
                  </div>
                  <div>
                    <dt className="text-muted">Industry</dt>
                    <dd className="font-semibold text-foreground">{s.industry}</dd>
                  </div>
                  <div>
                    <dt className="text-muted">Market TAM</dt>
                    <dd className="font-semibold text-foreground">{s.market.tam.toLocaleString("en-US")}</dd>
                  </div>
                </dl>

                <button
                  onClick={() => apply(s.id)}
                  disabled={busy !== null || hasCompany === false}
                  className="mt-5 w-full rounded-lg bg-brand px-4 py-2 text-sm font-semibold text-surface-muted hover:opacity-90 disabled:opacity-50"
                >
                  {isBusy ? "Applying…" : `Apply ${s.name}`}
                </button>
              </article>
            );
          })}
        </div>
      )}

      <CustomScenarioEditor />
    </div>
  );
}
