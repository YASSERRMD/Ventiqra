"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { NAV_ITEMS } from "@/lib/nav-items";

function isActive(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/";
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="flex w-64 shrink-0 flex-col border-r border-border bg-surface-muted">
      <div className="flex items-center gap-2 px-5 py-5">
        <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-brand text-surface-muted font-bold">
          V
        </span>
        <div className="leading-tight">
          <p className="text-sm font-semibold text-foreground">Ventiqra</p>
          <p className="text-xs text-muted">Startup Simulator</p>
        </div>
      </div>

      <nav className="flex-1 px-3 py-2">
        <ul className="space-y-1">
          {NAV_ITEMS.map((item) => {
            const active = isActive(pathname, item.href);
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  className={`flex flex-col rounded-md px-3 py-2 transition-colors ${
                    active
                      ? "bg-brand-soft text-brand"
                      : "text-muted hover:bg-surface hover:text-foreground"
                  }`}
                >
                  <span className="text-sm font-medium">{item.label}</span>
                  <span className="text-xs text-muted/80">{item.description}</span>
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="border-t border-border px-5 py-4 text-xs text-muted">
        v0.1.0 &middot; pre-alpha
      </div>
    </aside>
  );
}
