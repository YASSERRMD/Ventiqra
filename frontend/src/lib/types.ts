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

export type ProductStage = "idea" | "building" | "launched" | "retired";

export type Product = {
  id: string;
  name: string;
  slug: string;
  stage: ProductStage;
  dev_progress: number;
  price_cents: number | null;
  created_at: string;
  updated_at: string;
};

export type EmployeeRole =
  | "engineer"
  | "designer"
  | "sales"
  | "marketing"
  | "support"
  | "operations";

export const EMPLOYEE_ROLES: EmployeeRole[] = [
  "engineer",
  "designer",
  "sales",
  "marketing",
  "support",
  "operations",
];

export type Employee = {
  id: string;
  name: string;
  role: EmployeeRole;
  salary_cents: number;
  skill: number;
  morale: number;
  hired_at: string;
  created_at: string;
  updated_at: string;
};

export type EmployeeCreate = {
  name: string;
  role: EmployeeRole;
  salary_cents?: number;
  skill?: number;
  morale?: number;
};
