"use client";

import { ReactNode } from "react";

/**
 * EmptyState renders a centered "nothing here yet" message with an optional
 * action. Used by dashboard panels when there's no data.
 */
export function EmptyState({
  title,
  message,
  action,
}: {
  title: string;
  message?: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface/50 px-6 py-10 text-center">
      <p className="text-sm font-semibold text-muted">{title}</p>
      {message && <p className="mt-1 max-w-sm text-xs text-muted/70">{message}</p>}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}

/**
 * LoadingState renders a pulsing skeleton placeholder. Used while data fetches
 * are in flight.
 */
export function LoadingState({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-2">
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className="h-12 animate-pulse rounded-lg bg-border/40"
          style={{ animationDelay: `${i * 100}ms` }}
        />
      ))}
    </div>
  );
}

/**
 * ErrorState renders a user-friendly error message with an optional retry.
 */
export function ErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="rounded-xl border border-rose-500/40 bg-rose-500/10 px-4 py-3">
      <p className="text-sm text-rose-400">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="mt-2 rounded-md border border-rose-500/40 px-3 py-1 text-xs text-rose-400 hover:bg-rose-500/20"
        >
          Retry
        </button>
      )}
    </div>
  );
}

/**
 * PanelWrapper provides the standard panel chrome (title + content area) used
 * across the dashboard, with consistent spacing and responsive behavior.
 */
export function PanelWrapper({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle?: string;
  children: ReactNode;
}) {
  return (
    <section className="mt-6 sm:mt-8">
      <h3 className="text-base font-semibold text-foreground">{title}</h3>
      {subtitle && <p className="mt-1 text-xs text-muted">{subtitle}</p>}
      <div className="mt-3">{children}</div>
    </section>
  );
}
