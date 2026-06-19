import { PageHeader, Placeholder } from "@/components/layout/page-header";

export default function ProductsPage() {
  return (
    <div>
      <PageHeader
        title="Products"
        subtitle="Build, launch, and iterate on your products."
      />
      <Placeholder
        title="Product lifecycle"
        hint="Products arrive in Phase 9."
      />
    </div>
  );
}
