"use client";

import { useRealtime } from "@/lib/use-realtime";

const TONE: Record<string, string> = {
  open: "bg-emerald-500",
  connecting: "bg-amber-500 animate-pulse",
  closed: "bg-rose-500",
};

/**
 * LiveBadge subscribes to the realtime WebSocket and renders a small connection
 * indicator. When a tick message arrives it dispatches a window event so panels
 * can refresh their data without polling.
 */
export function LiveBadge() {
  const { status, lastMessage } = useRealtime({
    onMessage: (msg) => {
      // Tell the rest of the dashboard that fresh state arrived.
      window.dispatchEvent(new CustomEvent("ventiqra:tick", { detail: msg }));
    },
  });

  // If a decision/event message arrives, refresh the pending-decision panel.
  const live = lastMessage ? ` · day ${(lastMessage.payload?.day as number) ?? "—"}` : "";

  return (
    <span className="inline-flex items-center gap-2 rounded-full border border-border bg-surface px-3 py-1 text-xs text-muted">
      <span className={`h-2 w-2 rounded-full ${TONE[status]}`} />
      {status === "open" ? "Live" : status === "connecting" ? "Connecting" : "Offline"}
      <span className="text-muted/70">{live}</span>
    </span>
  );
}
