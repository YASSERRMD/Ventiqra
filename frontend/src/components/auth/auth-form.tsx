"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ApiError } from "@/lib/api";
import { login, register } from "@/lib/auth";

export function AuthForm({ mode }: { mode: "login" | "register" }) {
  const router = useRouter();
  const isRegister = mode === "register";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  function validate(): string | null {
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return "Enter a valid email.";
    if (password.length < 8) return "Password must be at least 8 characters.";
    if (isRegister && name.trim().length < 1) return "Name is required.";
    return null;
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const v = validate();
    if (v) {
      setError(v);
      return;
    }
    setBusy(true);
    try {
      if (isRegister) {
        await register(email, password, name.trim());
      } else {
        await login(email, password);
      }
      router.push("/company");
      router.refresh();
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Something went wrong. Is the backend running?",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-sm">
      <form
        onSubmit={onSubmit}
        className="flex flex-col gap-4 rounded-xl border border-border bg-surface p-6"
      >
        <h2 className="text-lg font-semibold text-foreground">
          {isRegister ? "Create your account" : "Log in to Ventiqra"}
        </h2>

        {isRegister && (
          <label className="flex flex-col gap-1 text-sm">
            <span className="text-muted">Name</span>
            <input
              className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoComplete="name"
            />
          </label>
        )}

        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Email</span>
          <input
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
          />
        </label>

        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">Password</span>
          <input
            type="password"
            className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete={isRegister ? "new-password" : "current-password"}
          />
        </label>

        {error && <p className="text-sm text-rose-400">{error}</p>}

        <button
          type="submit"
          disabled={busy}
          className="mt-1 rounded-md bg-brand px-4 py-2 text-sm font-semibold text-surface-muted disabled:opacity-60"
        >
          {busy ? "Please wait…" : isRegister ? "Register" : "Log in"}
        </button>

        <p className="text-center text-xs text-muted">
          {isRegister ? (
            <>
              Already have an account?{" "}
              <Link href="/login" className="text-brand hover:underline">
                Log in
              </Link>
            </>
          ) : (
            <>
              New here?{" "}
              <Link href="/register" className="text-brand hover:underline">
                Create an account
              </Link>
            </>
          )}
        </p>
      </form>
    </div>
  );
}
