export type Company = {
  id: string;
  name: string;
  slug: string;
  industry: string;
  description: string;
  founded_at: string;
  cash_cents: number;
  status: string;
  created_at: string;
  updated_at: string;
};

export type CompanyCreate = {
  name: string;
  industry?: string;
  description?: string;
  starting_cash_cents?: number;
};

export type Metrics = {
  cash_cents: number;
  revenue_cents: number;
  burn_cents_per_month: number;
  valuation_cents: number;
  runway_months: number;
  day: number;
};
