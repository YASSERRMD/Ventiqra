import { Sidebar } from "@/components/layout/sidebar";

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen w-full bg-background">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex items-center justify-between border-b border-border bg-surface/60 px-6 py-4 backdrop-blur">
          <div>
            <h1 className="text-lg font-semibold text-foreground">Ventiqra</h1>
            <p className="text-xs text-muted">
              Build a company. Ship products. Survive the market.
            </p>
          </div>
          <a
            href="https://github.com/YASSERRMD/Ventiqra"
            target="_blank"
            rel="noreferrer"
            className="rounded-md border border-border px-3 py-1.5 text-xs text-muted transition-colors hover:text-foreground"
          >
            GitHub
          </a>
        </header>
        <main className="flex-1 overflow-y-auto px-6 py-6">{children}</main>
      </div>
    </div>
  );
}
