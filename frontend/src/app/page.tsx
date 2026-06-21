import { PageHeader } from "@/components/layout/page-header";
import { BackendStatus } from "@/components/dashboard/backend-status";
import { MetricsCards } from "@/components/dashboard/metrics-cards";
import { CompetitorPanel } from "@/components/dashboard/competitor-panel";

export default function DashboardPage() {
  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle="Your company at a glance. Metrics come online as you build."
        action={<BackendStatus />}
      />

      <MetricsCards />
      <CompetitorPanel />
    </div>
  );
}
