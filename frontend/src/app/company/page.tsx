import { PageHeader, Placeholder } from "@/components/layout/page-header";

export default function CompanyPage() {
  return (
    <div>
      <PageHeader
        title="Company"
        subtitle="Found your startup and configure its starting conditions."
      />
      <Placeholder
        title="Company creation"
        hint="The company creation flow arrives in Phase 6."
      />
    </div>
  );
}
