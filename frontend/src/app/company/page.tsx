"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Company, CompanyCreate } from "@/lib/types";
import { PageHeader, Placeholder } from "@/components/layout/page-header";
import { DifficultySelector } from "@/components/dashboard/difficulty-selector";

type State =
  | { kind: "loading" }
  | { kind: "unauth" }
  | { kind: "create" }
  | { kind: "viewing"; company: Company }
  | { kind: "error"; message: string };

const CASH_PRESETS = [
  { label: "Bootstrap ($50K)", cents: 5_000_00 },
  { label: "SaaS ($500K)", cents: 500_000_00 },
  { label: "VC-backed ($2M)", cents: 2_000_000_00 },
];

export default function CompanyPage() {
  const router = useRouter();
  const [state, setState] = useState<State>(() =>
    getToken() ? { kind: "loading" } : { kind: "unauth" },
  );

  useEffect(() => {
    if (state.kind !== "loading") return;
    const token = getToken();
    if (!token) return;

    api
      .get<Company>("/api/v1/companies/me", { token })
      .then((company) => setState({ kind: "viewing", company }))
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          clearToken();
          setState({ kind: "unauth" });
        } else if (err instanceof ApiError && err.status === 404) {
          setState({ kind: "create" });
        } else {
          setState({
            kind: "error",
            message: err instanceof Error ? err.message : "Could not load company.",
          });
        }
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (state.kind === "loading") {
    return <p className="text-sm text-muted">Loading…</p>;
  }

  if (state.kind === "unauth") {
    return (
      <div>
        <PageHeader title="Company" subtitle="Found your startup and configure its starting conditions." />
        <div className="rounded-xl border border-border bg-surface p-8 text-center">
          <p className="text-sm text-muted">
            You need an account to create a company.{" "}
            <Link href="/register" className="text-brand hover:underline">
              Register
            </Link>{" "}
            or{" "}
            <Link href="/login" className="text-brand hover:underline">
              log in
            </Link>
            .
          </p>
        </div>
      </div>
    );
  }

  if (state.kind === "error") {
    return (
      <div>
        <PageHeader title="Company" />
        <div className="rounded-xl border border-border bg-surface p-6 text-sm text-rose-400">
          {state.message}
        </div>
      </div>
    );
  }

  if (state.kind === "viewing") {
    return <CompanyProfile company={state.company} />;
  }

  return (
    <CreateCompany
      onCreated={(company) => {
        setState({ kind: "viewing", company });
        router.refresh();
      }}
    />
  );
}

function CompanyProfile({ company }: { company: Company }) {
  return (
    <div>
      <PageHeader
        title={company.name}
        subtitle={`${company.industry || "Industry —"} · ${company.slug}`}
      />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Metric label="Cash" value={formatCents(company.cash_cents)} />
        <Metric label="Status" value={company.status} />
        <Metric label="Founded" value={new Date(company.founded_at).toLocaleDateString()} />
      </div>
      <DifficultySelector />
      {company.description ? (
        <p className="mt-6 max-w-2xl text-sm text-muted">{company.description}</p>
      ) : null}
      <div className="mt-8">
        <Placeholder title="Simulation controls" hint="The simulation engine arrives in Phase 7." />
      </div>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border bg-surface p-4">
      <p className="text-xs uppercase tracking-wide text-muted">{label}</p>
      <p className="mt-2 text-xl font-semibold text-foreground">{value}</p>
    </div>
  );
}

function CreateCompany({ onCreated }: { onCreated: (c: Company) => void }) {
  const [name, setName] = useState("");
  const [industry, setIndustry] = useState("");
  const [description, setDescription] = useState("");
  const [preset, setPreset] = useState(CASH_PRESETS[1].cents);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  function validate(): string | null {
    const trimmed = name.trim();
    if (trimmed.length < 1 || trimmed.length > 120) return "Name must be 1–120 characters.";
    if (description.length > 2000) return "Description must be 2000 characters or fewer.";
    return null;
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const v = validate();
    if (v) return setError(v);

    const token = getToken();
    if (!token) return setError("You are no longer logged in.");

    setBusy(true);
    try {
      const body: CompanyCreate = {
        name: name.trim(),
        industry: industry.trim(),
        description: description.trim(),
        starting_cash_cents: preset,
      };
      const created = await api.post<Company>("/api/v1/companies", body, { token });
      onCreated(created);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not create company.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div>
      <PageHeader title="Found your company" subtitle="Pick a starting position and begin the simulation." />
      <form
        onSubmit={onSubmit}
        className="flex max-w-xl flex-col gap-4 rounded-xl border border-border bg-surface p-6"
      >
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Company name *</span>
          <input
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={120}
          />
        </label>

        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Industry</span>
          <input
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={industry}
            onChange={(e) => setIndustry(e.target.value)}
            placeholder="e.g. SaaS, Aerospace, Fintech"
          />
        </label>

        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Description</span>
          <textarea
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            maxLength={2000}
            rows={3}
          />
        </label>

        <div className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Starting capital</span>
          <select
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={preset}
            onChange={(e) => setPreset(Number(e.target.value))}
          >
            {CASH_PRESETS.map((p) => (
              <option key={p.cents} value={p.cents}>
                {p.label}
              </option>
            ))}
          </select>
        </div>

        {error && <p className="text-sm text-rose-400">{error}</p>}

        <button
          type="submit"
          disabled={busy}
          className="mt-1 rounded-md bg-brand px-4 py-2 text-sm font-semibold text-surface-muted disabled:opacity-60"
        >
          {busy ? "Creating…" : "Found company"}
        </button>
      </form>
    </div>
  );
}
