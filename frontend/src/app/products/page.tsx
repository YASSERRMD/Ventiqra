"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Product, ProductStage } from "@/lib/types";
import { PageHeader } from "@/components/layout/page-header";

type State =
  | { kind: "loading" }
  | { kind: "unauth" }
  | { kind: "no-company" }
  | { kind: "ready"; products: Product[] }
  | { kind: "error"; message: string };

const STAGES: ProductStage[] = ["idea", "building", "launched", "retired"];

export default function ProductsPage() {
  const [state, setState] = useState<State>(() =>
    getToken() ? { kind: "loading" } : { kind: "unauth" },
  );

  useEffect(() => {
    if (state.kind !== "loading") return;
    const token = getToken();
    if (!token) return;
    api
      .get<Product[]>("/api/v1/companies/me/products", { token })
      .then((products) => setState({ kind: "ready", products }))
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          clearToken();
          setState({ kind: "unauth" });
        } else if (err instanceof ApiError && err.status === 404) {
          setState({ kind: "no-company" });
        } else {
          setState({
            kind: "error",
            message: err instanceof Error ? err.message : "Could not load products.",
          });
        }
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function refresh() {
    const token = getToken();
    if (!token) {
      setState({ kind: "unauth" });
      return;
    }
    try {
      const products = await api.get<Product[]>("/api/v1/companies/me/products", { token });
      setState({ kind: "ready", products });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearToken();
        setState({ kind: "unauth" });
      } else if (err instanceof ApiError && err.status === 404) {
        setState({ kind: "no-company" });
      } else {
        setState({
          kind: "error",
          message: err instanceof Error ? err.message : "Could not load products.",
        });
      }
    }
  }

  if (state.kind === "loading") {
    return <p className="text-sm text-muted">Loading…</p>;
  }

  if (state.kind === "unauth") {
    return (
      <div>
        <PageHeader
          title="Products"
          subtitle="Build, launch, and iterate on your products."
        />
        <div className="rounded-xl border border-border bg-surface p-8 text-center">
          <p className="text-sm text-muted">
            You need an account to manage products.{" "}
            <Link href="/register" className="text-brand hover:underline">
              Register
            </Link>{" "}
            or{" "}
            <Link href="/login" className="text-brand hover:underline">
              log in
            </Link>
            .
          </p>
        </div>
      </div>
    );
  }

  if (state.kind === "no-company") {
    return (
      <div>
        <PageHeader
          title="Products"
          subtitle="Build, launch, and iterate on your products."
        />
        <div className="rounded-xl border border-border bg-surface p-8 text-center">
          <p className="text-sm text-muted">
            Create a company first to start building products.{" "}
            <Link href="/company" className="text-brand hover:underline">
              Found a company
            </Link>
            .
          </p>
        </div>
      </div>
    );
  }

  if (state.kind === "error") {
    return (
      <div>
        <PageHeader title="Products" />
        <div className="rounded-xl border border-border bg-surface p-6 text-sm text-rose-400">
          {state.message}
        </div>
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="Products"
        subtitle="Build, launch, and iterate on your products."
      />
      <NewProduct onCreated={refresh} />
      {state.products.length === 0 ? (
        <div className="mt-6 rounded-xl border border-dashed border-border bg-surface/40 px-6 py-12 text-center">
          <h3 className="text-base font-medium text-foreground">No products yet</h3>
          <p className="mt-2 text-sm text-muted">
            Start by adding your first product above.
          </p>
        </div>
      ) : (
        <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {state.products.map((p) => (
            <ProductCard key={p.id} product={p} onChanged={refresh} />
          ))}
        </div>
      )}
    </div>
  );
}

const STAGE_TONE: Record<ProductStage, string> = {
  idea: "bg-sky-500/15 text-sky-300",
  building: "bg-amber-500/15 text-amber-300",
  launched: "bg-emerald-500/15 text-emerald-300",
  retired: "bg-zinc-500/15 text-zinc-300",
};

function ProductCard({
  product,
  onChanged,
}: {
  product: Product;
  onChanged: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const nextStage =
    product.stage === "retired"
      ? null
      : STAGES[STAGES.indexOf(product.stage) + 1];

  async function advance() {
    if (!nextStage) return;
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await api.patch(`/api/v1/products/${product.id}/stage`, { stage: nextStage }, { token });
      onChanged();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not advance stage.");
    } finally {
      setBusy(false);
    }
  }

  const progress = Math.max(0, Math.min(100, product.dev_progress));

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-5">
      <div className="flex items-start justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold text-foreground">{product.name}</h3>
          <p className="text-xs text-muted">{product.slug}</p>
        </div>
        <span
          className={`rounded-full px-2.5 py-0.5 text-xs font-medium capitalize ${STAGE_TONE[product.stage]}`}
        >
          {product.stage}
        </span>
      </div>

      <div>
        <div className="flex items-center justify-between text-xs text-muted">
          <span>Development</span>
          <span>{progress.toFixed(0)}%</span>
        </div>
        <div className="mt-1 h-2 w-full overflow-hidden rounded-full bg-background">
          <div
            className="h-full rounded-full bg-brand"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      <div className="text-sm text-muted">
        Price:{" "}
        <span className="text-foreground">
          {product.price_cents != null ? formatCents(product.price_cents) : "—"}
        </span>
      </div>

      {nextStage ? (
        <button
          type="button"
          onClick={advance}
          disabled={busy}
          className="rounded-md border border-border bg-background px-3 py-1.5 text-xs font-semibold text-foreground capitalize outline-none transition hover:border-brand disabled:opacity-60"
        >
          {busy ? "Updating…" : `Advance to ${nextStage}`}
        </button>
      ) : (
        <span className="text-xs text-muted">This product is retired.</span>
      )}

      {error && <p className="text-xs text-rose-400">{error}</p>}
    </div>
  );
}

function NewProduct({ onCreated }: { onCreated: () => void }) {
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (trimmed.length < 1 || trimmed.length > 120) {
      setError("Name must be 1–120 characters.");
      return;
    }
    const token = getToken();
    if (!token) {
      setError("You are no longer logged in.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await api.post("/api/v1/companies/me/products", { name: trimmed }, { token });
      setName("");
      onCreated();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not create product.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form
      onSubmit={onSubmit}
      className="flex max-w-xl items-end gap-3 rounded-xl border border-border bg-surface p-4"
    >
      <label className="flex flex-1 flex-col gap-1 text-sm">
        <span className="text-muted">New product name</span>
        <input
          className="rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-brand"
          value={name}
          onChange={(e) => setName(e.target.value)}
          maxLength={120}
          placeholder="e.g. Orbit Analytics"
        />
      </label>
      <button
        type="submit"
        disabled={busy}
        className="rounded-md bg-brand px-4 py-2 text-sm font-semibold text-surface-muted disabled:opacity-60"
      >
        {busy ? "Adding…" : "Add product"}
      </button>
      {error && <p className="text-sm text-rose-400">{error}</p>}
    </form>
  );
}
