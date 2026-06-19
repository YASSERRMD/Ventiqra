import { PageHeader, Placeholder } from "@/components/layout/page-header";

export default function FinancePage() {
  return (
    <div>
      <PageHeader
        title="Finance"
        subtitle="Track cash, burn rate, runway, and funding rounds."
      />
      <Placeholder
        title="Finance & funding"
        hint="Finance tooling arrives in Phases 16 and 18."
      />
    </div>
  );
}
