import { PageHeader } from "@/components/layout/page-header";
import { AuthForm } from "@/components/auth/auth-form";

export default function RegisterPage() {
  return (
    <div>
      <PageHeader title="Create your account" subtitle="Start building your startup." />
      <AuthForm mode="register" />
    </div>
  );
}
