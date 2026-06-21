export type NavItem = {
  label: string;
  href: string;
  description: string;
};

// Primary navigation for the Ventiqra dashboard shell. Routes without backing
// features yet render a placeholder page; they are filled in by later phases.
export const NAV_ITEMS: NavItem[] = [
  { label: "Dashboard", href: "/", description: "Company overview and live metrics" },
  { label: "Company", href: "/company", description: "Found and manage your company" },
  { label: "Products", href: "/products", description: "Build, launch, and iterate" },
  { label: "Employees", href: "/employees", description: "Hire and manage your team" },
  { label: "Finance", href: "/finance", description: "Cash, burn, runway, and funding" },
  { label: "Saves", href: "/saves", description: "Save and restore simulation runs" },
  { label: "Scenarios", href: "/scenarios", description: "Predefined and custom scenarios" },
];
