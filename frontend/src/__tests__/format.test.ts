import { describe, it, expect } from "vitest";
import { formatCents, formatNumber } from "@/lib/format";

describe("formatCents", () => {
  it("formats zero", () => {
    expect(formatCents(0)).toBe("$0.00");
  });

  it("formats positive amounts with dollars and cents", () => {
    expect(formatCents(12345)).toBe("$123.45");
    expect(formatCents(100)).toBe("$1.00");
    expect(formatCents(1_000_000_00)).toBe("$1,000,000.00");
  });

  it("formats negative amounts", () => {
    expect(formatCents(-500)).toBe("-$5.00");
  });
});

describe("formatNumber", () => {
  it("formats plain integers", () => {
    expect(formatNumber(42)).toBe("42");
  });

  it("formats large numbers with separators", () => {
    expect(formatNumber(1000000)).toBe("1,000,000");
  });

  it("formats zero", () => {
    expect(formatNumber(0)).toBe("0");
  });
});
