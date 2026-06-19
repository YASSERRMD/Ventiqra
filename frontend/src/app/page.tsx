import { PageHeader } from "@/components/layout/page-header";
import { BackendStatus } from "@/components/dashboard/backend-status";

const METRICS = [
  { label: "Cash", value: "—", hint: "Available capital" },
  { label: "Burn Rate", value: "—", hint: "Monthly spend" },
  { label: "Runway", value: "—", hint: "Months until broke" },
  { label: "Valuation", value: "—", hint: "Current valuation" },
];

export default function DashboardPage() {
  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle="Your company at a glance. Metrics come online as you build."
        action={<BackendStatus />}
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {METRICS.map((m) => (
          <div
            key={m.label}
            className="rounded-xl border border-border bg-surface p-4"
          >
            <p className="text-xs uppercase tracking-wide text-muted">{m.label}</p>
            <p className="mt-2 text-2xl font-semibold text-foreground">{m.value}</p>
            <p className="mt-1 text-xs text-muted/80">{m.hint}</p>
          </div>
        ))}
      </div>

      <div className="mt-6 rounded-xl border border-dashed border-border bg-surface/40 p-8 text-center">
        <p className="text-sm text-muted">
          No company yet. Found a company in the{" "}
          <a href="/company" className="text-brand hover:underline">
            Company
          </a>{" "}
          section to begin the simulation.
        </p>
      </div>
    </div>
  );
}
