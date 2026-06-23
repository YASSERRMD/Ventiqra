import { describe, it, expect, vi } from "vitest";
import { render } from "@testing-library/react";

// Mock the api module so components that fetch on mount don't make real calls.
vi.mock("@/lib/api", () => ({
  api: {
    get: vi.fn().mockRejectedValue(new Error("no token")),
    post: vi.fn().mockRejectedValue(new Error("no token")),
    patch: vi.fn().mockRejectedValue(new Error("no token")),
    del: vi.fn().mockRejectedValue(new Error("no token")),
  },
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
}));

// Mock getToken to return null (no logged-in user).
vi.mock("@/lib/auth", () => ({
  getToken: vi.fn(() => null),
}));

describe("LiveBadge component", () => {
  it("renders without crashing when no token", async () => {
    const { LiveBadge } = await import("@/components/dashboard/live-badge");
    const { container } = render(<LiveBadge />);
    // With no token, useRealval returns "closed" status; the badge renders.
    expect(container.textContent).toContain("Offline");
  });
});

describe("DifficultySelector component", () => {
  it("renders nothing when no company (no token)", async () => {
    const { DifficultySelector } = await import("@/components/dashboard/difficulty-selector");
    const { container } = render(<DifficultySelector />);
    // With no token, the component returns null.
    expect(container.firstChild).toBeNull();
  });
});

describe("AchievementsPanel component", () => {
  it("renders nothing when no token", async () => {
    const { AchievementsPanel } = await import("@/components/dashboard/achievements-panel");
    const { container } = render(<AchievementsPanel />);
    expect(container.firstChild).toBeNull();
  });
});
