import { PageHeader, Placeholder } from "@/components/layout/page-header";

export default function EmployeesPage() {
  return (
    <div>
      <PageHeader
        title="Employees"
        subtitle="Hire talent and manage salaries, productivity, and morale."
      />
      <Placeholder
        title="Hiring & team"
        hint="Employees arrive in Phase 10."
      />
    </div>
  );
}
