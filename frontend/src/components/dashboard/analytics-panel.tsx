"use client";

import { useEffect, useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatCents } from "@/lib/format";
import type { Analytics } from "@/lib/types";

const tooltipStyle = {
  background: "#0f172a",
  border: "1px solid #1e293b",
  borderRadius: 8,
  fontSize: 12,
};

export function AnalyticsPanel() {
  const [data, setData] = useState<Analytics | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    api
      .get<Analytics>("/api/v1/companies/me/analytics", { token })
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 404)) {
          setData(null);
        }
      });
  }, []);

  if (!data || data.cash.length < 2) return null;

  const series = data.cash;

  return (
    <section className="mt-8">
      <h3 className="text-base font-semibold text-foreground">Analytics</h3>
      <p className="mt-1 text-xs text-muted">
        Headline metrics over time. Now on day {data.day}.
      </p>

      <div className="mt-3 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ChartCard title="Cash" series={series} dataKey="cash_cents" color="#6366f1" format />
        <ChartCard title="Daily revenue" series={series} dataKey="revenue_cents" color="#10b981" format />
        <ChartCard title="Monthly burn" series={series} dataKey="monthly_burn" color="#f43f5e" format />
        <ChartCard title="Valuation" series={series} dataKey="valuation_cents" color="#0ea5e9" format />
        <ChartCard title="Customers" series={series} dataKey="customers" color="#a855f7" />
        <ChartCard title="Cash trend" series={series} dataKey="cash_cents" color="#6366f1" format area />
      </div>
    </section>
  );
}

function ChartCard({
  title,
  series,
  dataKey,
  color,
  format,
  area,
}: {
  title: string;
  series: { day: number;[k: string]: number }[];
  dataKey: string;
  color: string;
  format?: boolean;
  area?: boolean;
}) {
  return (
    <div className="rounded-xl border border-border bg-surface p-3">
      <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted">{title}</p>
      <div className="h-40 w-full">
        <ResponsiveContainer width="100%" height="100%">
          {area ? (
            <AreaChart data={series} margin={{ top: 4, right: 8, bottom: 0, left: -8 }}>
              <defs>
                <linearGradient id={`g-${dataKey}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor={color} stopOpacity={0.5} />
                  <stop offset="100%" stopColor={color} stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
              <XAxis dataKey="day" tick={{ fill: "#94a3b8", fontSize: 11 }} />
              <YAxis tick={{ fill: "#94a3b8", fontSize: 11 }} tickFormatter={(v) => (format ? `$${Math.round(v / 100)}k` : v)} width={48} />
              <Tooltip contentStyle={tooltipStyle} formatter={(v) => (format ? formatCents(Number(v)) : String(v))} labelFormatter={(d) => `Day ${d}`} />
              <Area type="monotone" dataKey={dataKey} stroke={color} strokeWidth={2} fill={`url(#g-${dataKey})`} />
            </AreaChart>
          ) : (
            <LineChart data={series} margin={{ top: 4, right: 8, bottom: 0, left: -8 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
              <XAxis dataKey="day" tick={{ fill: "#94a3b8", fontSize: 11 }} />
              <YAxis tick={{ fill: "#94a3b8", fontSize: 11 }} tickFormatter={(v) => (format ? `$${Math.round(v / 100)}k` : v)} width={48} />
              <Tooltip contentStyle={tooltipStyle} formatter={(v) => (format ? formatCents(Number(v)) : String(v))} labelFormatter={(d) => `Day ${d}`} />
              <Line type="monotone" dataKey={dataKey} stroke={color} strokeWidth={2} dot={false} />
            </LineChart>
          )}
        </ResponsiveContainer>
      </div>
    </div>
  );
}
