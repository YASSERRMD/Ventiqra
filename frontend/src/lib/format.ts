// Formatting helpers for display.

// formatCents renders an integer-cent value as a USD currency string.
export function formatCents(cents: number): string {
  const sign = cents < 0 ? "-" : "";
  const abs = Math.abs(cents);
  const dollars = Math.floor(abs / 100);
  const remainder = abs % 100;
  return `${sign}$${dollars.toLocaleString("en-US")}.${remainder.toString().padStart(2, "0")}`;
}

// formatNumber adds thousands separators to an integer.
export function formatNumber(n: number): string {
  return n.toLocaleString("en-US");
}
