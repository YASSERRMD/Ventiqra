import { PageHeader } from "@/components/layout/page-header";
import { BackendStatus } from "@/components/dashboard/backend-status";
import { SpeedControls } from "@/components/dashboard/speed-controls";
import { LiveBadge } from "@/components/dashboard/live-badge";
import { MetricsCards } from "@/components/dashboard/metrics-cards";
import { CompetitorPanel } from "@/components/dashboard/competitor-panel";
import { MarketPanel } from "@/components/dashboard/market-panel";
import { MarketingPanel } from "@/components/dashboard/marketing-panel";
import { ReputationPanel } from "@/components/dashboard/reputation-panel";
import { EventsPanel } from "@/components/dashboard/events-panel";
import { TimelinePanel } from "@/components/dashboard/timeline-panel";
import { RoadmapPanel } from "@/components/dashboard/roadmap-panel";
import { AnalyticsPanel } from "@/components/dashboard/analytics-panel";
import { DecisionModal } from "@/components/dashboard/decision-modal";

export default function DashboardPage() {
  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle="Your company at a glance. Metrics come online as you build."
        action={
          <div className="flex flex-wrap items-center gap-3">
            <SpeedControls />
            <LiveBadge />
            <BackendStatus />
          </div>
        }
      />

      <MetricsCards />
      <MarketPanel />
      <MarketingPanel />
      <ReputationPanel />
      <CompetitorPanel />
      <EventsPanel />
      <TimelinePanel />
      <RoadmapPanel />
      <AnalyticsPanel />
      <DecisionModal />
    </div>
  );
}
