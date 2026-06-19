import { PageHeader } from "@/components/layout/page-header";
import { AuthForm } from "@/components/auth/auth-form";

export default function LoginPage() {
  return (
    <div>
      <PageHeader title="Log in" subtitle="Welcome back to Ventiqra." />
      <AuthForm mode="login" />
    </div>
  );
}
