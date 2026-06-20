"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Employee, EmployeeCreate, EmployeeRole, Candidate } from "@/lib/types";
import { EMPLOYEE_ROLES } from "@/lib/types";
import { PageHeader } from "@/components/layout/page-header";

type State =
  | { kind: "loading" }
  | { kind: "unauth" }
  | { kind: "no-company" }
  | { kind: "ready"; employees: Employee[] }
  | { kind: "error"; message: string };

const ROLE_TONE: Record<EmployeeRole, string> = {
  engineer: "bg-sky-500/15 text-sky-300",
  designer: "bg-fuchsia-500/15 text-fuchsia-300",
  sales: "bg-emerald-500/15 text-emerald-300",
  marketing: "bg-amber-500/15 text-amber-300",
  support: "bg-cyan-500/15 text-cyan-300",
  operations: "bg-zinc-500/15 text-zinc-300",
};

export default function EmployeesPage() {
  const [state, setState] = useState<State>(() =>
    getToken() ? { kind: "loading" } : { kind: "unauth" },
  );

  useEffect(() => {
    if (state.kind !== "loading") return;
    void load(setState);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function refresh() {
    await load(setState);
  }

  if (state.kind === "loading") {
    return <p className="text-sm text-muted">Loading…</p>;
  }

  if (state.kind === "unauth") {
    return (
      <div>
        <PageHeader
          title="Employees"
          subtitle="Hire talent and manage salaries, productivity, and morale."
        />
        <div className="rounded-xl border border-border bg-surface p-8 text-center">
          <p className="text-sm text-muted">
            You need an account to manage a team.{" "}
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

  if (state.kind === "no-company") {
    return (
      <div>
        <PageHeader
          title="Employees"
          subtitle="Hire talent and manage salaries, productivity, and morale."
        />
        <div className="rounded-xl border border-border bg-surface p-8 text-center">
          <p className="text-sm text-muted">
            Create a company first to start hiring.{" "}
            <Link href="/company" className="text-brand hover:underline">
              Found a company
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
        <PageHeader title="Employees" />
        <div className="rounded-xl border border-border bg-surface p-6 text-sm text-rose-400">
          {state.message}
        </div>
      </div>
    );
  }

  const totalMonthly = state.employees.reduce(
    (sum, e) => sum + e.salary_cents,
    0,
  );

  return (
    <div>
      <PageHeader
        title="Employees"
        subtitle="Hire talent and manage salaries, productivity, and morale."
        action={
          <div className="rounded-lg border border-border bg-surface px-4 py-2 text-right">
            <p className="text-xs text-muted">Monthly payroll</p>
            <p className="text-sm font-semibold text-foreground">
              {formatCents(totalMonthly)}
            </p>
          </div>
        }
      />
      <HireEmployee onHired={refresh} />
      <HiringMarket onChanged={refresh} />
      {state.employees.length === 0 ? (
        <div className="mt-6 rounded-xl border border-dashed border-border bg-surface/40 px-6 py-12 text-center">
          <h3 className="text-base font-medium text-foreground">No employees yet</h3>
          <p className="mt-2 text-sm text-muted">
            Your team is empty. Hire your first employee above.
          </p>
        </div>
      ) : (
        <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {state.employees.map((e) => (
            <EmployeeCard key={e.id} employee={e} onChanged={refresh} />
          ))}
        </div>
      )}
    </div>
  );
}

async function load(setState: (s: State) => void) {
  const token = getToken();
  if (!token) {
    setState({ kind: "unauth" });
    return;
  }
  try {
    const employees = await api.get<Employee[]>("/api/v1/companies/me/employees", { token });
    setState({ kind: "ready", employees });
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      clearToken();
      setState({ kind: "unauth" });
    } else if (err instanceof ApiError && err.status === 404) {
      setState({ kind: "no-company" });
    } else {
      setState({
        kind: "error",
        message: err instanceof Error ? err.message : "Could not load employees.",
      });
    }
  }
}

function EmployeeCard({
  employee,
  onChanged,
}: {
  employee: Employee;
  onChanged: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [salary, setSalary] = useState(String(employee.salary_cents));
  const token = getToken();

  async function saveSalary() {
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    const cents = Math.max(0, Math.round(Number(salary) || 0));
    setBusy(true);
    setError(null);
    try {
      await api.patch(`/api/v1/employees/${employee.id}/salary`, { salary_cents: cents }, { token });
      setSalary(String(cents));
      onChanged();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not update salary.");
    } finally {
      setBusy(false);
    }
  }

  async function fire() {
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await api.del(`/api/v1/employees/${employee.id}`, { token });
      onChanged();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not fire employee.");
    } finally {
      setBusy(false);
    }
  }

  const skill = Math.max(0, Math.min(100, employee.skill));
  const morale = Math.max(0, Math.min(100, employee.morale));

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-5">
      <div className="flex items-start justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold text-foreground">{employee.name}</h3>
          <p className="text-xs text-muted">
            Hired {new Date(employee.hired_at).toLocaleDateString()}
          </p>
        </div>
        <span
          className={`rounded-full px-2.5 py-0.5 text-xs font-medium capitalize ${ROLE_TONE[employee.role]}`}
        >
          {employee.role}
        </span>
      </div>

      <Meter label="Skill" value={skill} tone="bg-brand" />
      <Meter label="Morale" value={morale} tone="bg-emerald-500" />

      <div className="text-sm text-muted">
        Salary:{" "}
        <span className="text-foreground">{formatCents(employee.salary_cents)}/mo</span>
      </div>

      <div className="flex items-end gap-2">
        <label className="flex flex-1 flex-col gap-1 text-xs text-muted">
          <span>Salary (cents/mo)</span>
          <input
            className="rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground outline-none focus:border-brand"
            value={salary}
            onChange={(e) => setSalary(e.target.value)}
            inputMode="numeric"
          />
        </label>
        <button
          type="button"
          onClick={saveSalary}
          disabled={busy}
          className="rounded-md border border-border bg-background px-3 py-1.5 text-xs font-semibold text-foreground outline-none transition hover:border-brand disabled:opacity-60"
        >
          {busy ? "Saving…" : "Set"}
        </button>
      </div>

      <button
        type="button"
        onClick={fire}
        disabled={busy}
        className="rounded-md border border-rose-500/40 bg-rose-500/10 px-3 py-1.5 text-xs font-semibold text-rose-300 outline-none transition hover:bg-rose-500/20 disabled:opacity-60"
      >
        Fire
      </button>

      {error && <p className="text-xs text-rose-400">{error}</p>}
    </div>
  );
}

function Meter({ label, value, tone }: { label: string; value: number; tone: string }) {
  return (
    <div>
      <div className="flex items-center justify-between text-xs text-muted">
        <span>{label}</span>
        <span>{value.toFixed(0)}</span>
      </div>
      <div className="mt-1 h-2 w-full overflow-hidden rounded-full bg-background">
        <div className={`h-full rounded-full ${tone}`} style={{ width: `${value}%` }} />
      </div>
    </div>
  );
}

function HireEmployee({ onHired }: { onHired: () => void }) {
  const [name, setName] = useState("");
  const [role, setRole] = useState<EmployeeRole>("engineer");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (trimmed.length < 1 || trimmed.length > 120) {
      setError("Name must be 1–120 characters.");
      return;
    }
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const body: EmployeeCreate = { name: trimmed, role };
      await api.post("/api/v1/companies/me/employees", body, { token });
      setName("");
      onHired();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not hire employee.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form
      onSubmit={onSubmit}
      className="flex max-w-2xl flex-wrap items-end gap-3 rounded-xl border border-border bg-surface p-4"
    >
      <label className="flex flex-1 flex-col gap-1 text-sm">
        <span className="text-muted">Name</span>
        <input
          className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
          value={name}
          onChange={(e) => setName(e.target.value)}
          maxLength={120}
          placeholder="e.g. Ada Lovelace"
        />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted">Role</span>
        <select
          className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
          value={role}
          onChange={(e) => setRole(e.target.value as EmployeeRole)}
        >
          {EMPLOYEE_ROLES.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </select>
      </label>
      <button
        type="submit"
        disabled={busy}
        className="rounded-md bg-brand px-4 py-2 text-sm font-semibold text-surface-muted disabled:opacity-60"
      >
        {busy ? "Hiring…" : "Hire"}
      </button>
      {error && <p className="basis-full text-sm text-rose-400">{error}</p>}
    </form>
  );
}

const QUALITY_TONE: Record<Candidate["quality"], string> = {
  weak: "bg-zinc-500/15 text-zinc-300",
  average: "bg-amber-500/15 text-amber-300",
  strong: "bg-emerald-500/15 text-emerald-300",
};

function HiringMarket({ onChanged }: { onChanged: () => void }) {
  const [candidates, setCandidates] = useState<Candidate[] | null>(null);
  const [day, setDay] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ index: number; text: string; ok: boolean } | null>(null);

  async function loadMarket() {
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const pool = await api.get<{ day: number; candidates: Candidate[] }>(
        "/api/v1/companies/me/candidates",
        { token },
      );
      setCandidates(pool.candidates);
      setDay(pool.day);
      setResult(null);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not load candidates.");
    } finally {
      setLoading(false);
    }
  }

  async function makeOffer(c: Candidate) {
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await api.post<{ accepted: boolean; message: string }>(
        `/api/v1/companies/me/candidates/${c.index}/hire`,
        undefined,
        { token },
      );
      setResult({ index: c.index, text: res.message, ok: res.accepted });
      onChanged();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not make offer.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className="mt-6">
      <div className="mb-3 flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-foreground">Hiring market</h3>
          <p className="text-xs text-muted">
            Deterministic candidate pool{day != null ? ` — round day ${day}` : ""}. Offers resolve the same way until the day advances.
          </p>
        </div>
        <button
          type="button"
          onClick={loadMarket}
          disabled={loading}
          className="rounded-md border border-border bg-background px-3 py-1.5 text-xs font-semibold text-foreground outline-none transition hover:border-brand disabled:opacity-60"
        >
          {loading ? "Loading…" : candidates ? "Refresh" : "Show candidates"}
        </button>
      </div>

      {error && <p className="text-sm text-rose-400">{error}</p>}

      {candidates && (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {candidates.map((c) => (
            <div key={c.index} className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-4">
              <div className="flex items-start justify-between gap-2">
                <div>
                  <p className="text-sm font-semibold text-foreground">{c.name}</p>
                  <p className="text-xs text-muted capitalize">{c.role}</p>
                </div>
                <span className={`rounded-full px-2 py-0.5 text-xs font-medium capitalize ${QUALITY_TONE[c.quality]}`}>
                  {c.quality}
                </span>
              </div>
              <div className="text-xs text-muted">Skill: <span className="text-foreground">{c.skill}</span></div>
              <div className="text-xs text-muted">
                Salary: <span className="text-foreground">{formatCents(c.salary_expectation_cents)}/mo</span>
              </div>
              <div className="text-xs text-muted">
                Hiring fee: <span className="text-foreground">{formatCents(c.hiring_fee_cents)}</span>
              </div>
              <div className="text-xs text-muted">
                Acceptance: <span className="text-foreground">{(c.acceptance_chance * 100).toFixed(0)}%</span>
              </div>
              <button
                type="button"
                onClick={() => makeOffer(c)}
                disabled={loading}
                className="mt-1 rounded-md bg-brand px-3 py-1.5 text-xs font-semibold text-surface-muted disabled:opacity-60"
              >
                Make offer
              </button>
              {result && result.index === c.index && (
                <p className={`text-xs ${result.ok ? "text-emerald-400" : "text-rose-400"}`}>{result.text}</p>
              )}
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
