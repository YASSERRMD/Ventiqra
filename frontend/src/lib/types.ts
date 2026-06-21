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
  health?: string;
  status?: string;
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

export type Candidate = {
  index: number;
  role: EmployeeRole;
  name: string;
  quality: "weak" | "average" | "strong";
  skill: number;
  salary_expectation_cents: number;
  hiring_fee_cents: number;
  acceptance_chance: number;
};

export type CandidatePool = {
  day: number;
  candidates: Candidate[];
};

export type HireResult = {
  accepted: boolean;
  message: string;
  employee?: Employee;
};

export type LaunchResult = {
  readiness: number;
  initial_customers: number;
  product: Product;
};

export type LaunchEvent = {
  id: string;
  product_id: string;
  product_name: string;
  readiness: number;
  initial_customers: number;
  launched_at: string;
};

export type CustomerState = {
  product_id: string;
  product_name: string;
  total_customers: number;
  mau: number;
  churned: number;
  satisfaction: number;
  updated_at: string;
};

export type PricingExperiment = {
  id: string;
  product_id: string;
  product_name: string;
  old_price_cents: number | null;
  new_price_cents: number;
  sim_day: number;
  created_at: string;
};

export type FinanceBreakdown = {
  base_cents: number;
  salary_cents: number;
  infra_cents: number;
  marketing_cents: number;
  total_burn_cents: number;
};

export type Finance = {
  marketing_budget_cents: number;
  monthly_revenue_cents: number;
  burn: FinanceBreakdown;
  profit_loss_cents: number;
  day: number;
};

export type FundingRound = {
  id: string;
  round_name: string;
  amount_cents: number;
  pre_money_cents: number;
  equity_percent: number;
  sim_day: number;
  created_at: string;
};

export type FundingSummary = {
  pre_money_cents: number;
  founder_equity_percent: number;
  investor_interest: number;
  rounds_raised: number;
  rounds: FundingRound[];
};

export type InvestorOffer = {
  id: string;
  investor_name: string;
  amount_cents: number;
  equity_percent: number;
  status: string;
  created_at: string;
};

export type NegotiateResult = {
  withdrawn: boolean;
  message: string;
  offer?: InvestorOffer;
};

export type Competitor = {
  id: string;
  name: string;
  strength: number;
  market_share: number;
  last_launch_day: number;
  updated_at: string;
};

export type Market = {
  tam: number;
  growth_rate: number;
  trend_multiplier: number;
  updated_at: string;
};

export type MarketingChannel = {
  name: string;
  weight: number;
  conversion: number;
};

export type Marketing = {
  monthly_budget_cents: number;
  daily_conversions: number;
  cac_cents: number;
  conversion_rate: number;
  channels: MarketingChannel[];
};

export type ReputationEvent = {
  id: string;
  event: string;
  delta: number;
  sim_day: number;
  created_at: string;
};

export type Reputation = {
  score: number;
  growth_multiplier: number;
  events: ReputationEvent[];
};

export type MoraleSummary = {
  headcount: number;
  average_morale: number;
  at_risk: number;
  burnt_out: number;
};

export type GameEvent = {
  id: string;
  kind: "positive" | "negative" | "neutral" | "crisis";
  title: string;
  description: string;
  cash_delta: number;
  reputation_delta: number;
  morale_delta: number;
  sim_day: number;
  created_at: string;
};

export type DecisionChoice = {
  id: string;
  label: string;
  description: string;
  cash_delta: number;
  reputation_delta: number;
  morale_delta: number;
  success_chance: number;
  fail_cash_delta: number;
  fail_reputation_delta: number;
  fail_morale_delta: number;
  recurring_cash_delta: number;
  duration_days: number;
};

export type PendingDecision = {
  id: string;
  decision_id: string;
  title: string;
  description: string;
  category: string;
  sim_day_offered: number;
  choices: DecisionChoice[];
};

export type DecisionOutcome = {
  id: string;
  outcome: "success" | "failure";
  applied_cash_delta: number;
  applied_reputation_delta: number;
  applied_morale_delta: number;
  recurring_cash_delta: number;
  remaining_days: number;
};

export type ScenarioMarket = {
  tam: number;
  growth_rate: number;
  trend_multiplier: number;
};

export type Scenario = {
  id: string;
  name: string;
  category: string;
  description: string;
  difficulty: "easy" | "normal" | "hard" | "brutal";
  industry: string;
  starting_cash_cents: number;
  starting_burn_cents: number;
  market: ScenarioMarket;
};

export type ApplyScenarioResult = {
  scenario: Scenario;
  company: Company;
  applied_at: string;
};

export type CustomScenario = {
  id: string;
  name: string;
  description: string;
  difficulty: Scenario["difficulty"];
  industry: string;
  starting_cash_cents: number;
  starting_burn_cents: number;
  market: ScenarioMarket;
  created_at: string;
  updated_at: string;
};

export type CustomScenarioInput = {
  name: string;
  description?: string;
  difficulty?: Scenario["difficulty"];
  industry?: string;
  starting_cash_cents: number;
  starting_burn_cents: number;
  market_tam: number;
  market_growth_rate: number;
  market_trend: number;
};
