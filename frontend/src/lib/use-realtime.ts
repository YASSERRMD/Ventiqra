"use client";

import { useEffect, useRef, useState } from "react";
import { getToken } from "@/lib/auth";

export type RealtimeMessage = {
  type: string; // "tick" | "event" | "decision" | "funding" | "hello"
  company_id: string;
  payload: Record<string, unknown> | undefined;
  sent_at: string;
};

type Options = {
  onMessage?: (msg: RealtimeMessage) => void;
  enabled?: boolean;
};

/**
 * useRealtime opens a WebSocket to the backend realtime endpoint, authenticated
 * via the stored JWT query param. It reconnects with exponential backoff and
 * surfaces connection status plus an optional per-message callback.
 */
export function useRealtime({ onMessage, enabled = true }: Options = {}) {
  const [status, setStatus] = useState<"connecting" | "open" | "closed">("closed");
  const [lastMessage, setLastMessage] = useState<RealtimeMessage | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const retryRef = useRef(0);
  const closedByUnmount = useRef(false);

  useEffect(() => {
    if (!enabled) return;
    const token = getToken();
    if (!token) return;

    closedByUnmount.current = false;

    const connect = () => {
      if (closedByUnmount.current) return;
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const url = `${proto}//${window.location.host}/api/v1/realtime?token=${encodeURIComponent(token)}`;
      setStatus("connecting");

      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        retryRef.current = 0;
        setStatus("open");
      };
      ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data) as RealtimeMessage;
          setLastMessage(msg);
          onMessage?.(msg);
        } catch {
          // Ignore malformed frames.
        }
      };
      ws.onclose = () => {
        setStatus("closed");
        if (closedByUnmount.current) return;
        // Exponential backoff capped at ~30s.
        const delay = Math.min(1000 * 2 ** retryRef.current, 30_000);
        retryRef.current += 1;
        setTimeout(connect, delay);
      };
      ws.onerror = () => {
        // Let onclose drive reconnection.
        ws.close();
      };
    };

    connect();

    return () => {
      closedByUnmount.current = true;
      wsRef.current?.close();
    };
    // onMessage is intentionally omitted: callers should memoize it.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled]);

  return { status, lastMessage };
}
