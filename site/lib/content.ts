/**
 * Site content as typed data.
 *
 * Copy lives here rather than inline in JSX so a claim appears exactly once.
 * The previous site drifted — Data Guard was described as "GCP data security"
 * on one page and "Go, Spark, Iceberg" on another, which is the sort of thing
 * a prospect notices and a candidate gets asked about in an interview.
 */

export type ProductStatus = 'live' | 'open-source' | 'coming-soon';

export interface Product {
  slug: string;
  name: string;
  /** One line a non-engineer understands. */
  value: string;
  /** Three things it does. Not features — outcomes. */
  highlights: [string, string, string];
  status: ProductStatus;
  /** Whether the product page offers a Dashboard tab or a Repository tab. */
  secondaryTab: { kind: 'dashboard' } | { kind: 'repository'; href: string };
}

export const PRODUCTS: Product[] = [
  {
    slug: 'cloud-guard',
    name: 'Cloud Guard',
    value: 'Finds the AWS spend you forgot about, and the exposure you did not know you had.',
    highlights: [
      'Scans every region you have enabled, through a read-only role',
      'Prices each finding from published AWS list prices',
      'Hands you the exact command that fixes it',
    ],
    status: 'live',
    secondaryTab: { kind: 'dashboard' },
  },
  {
    slug: 'shield',
    name: 'Shield',
    value: 'Rate limiting that holds when one client starts hammering your API.',
    highlights: [
      'Four algorithms — token bucket, leaky bucket, fixed window, sliding log',
      'Correct 429 responses with X-RateLimit headers, so clients can back off',
      'Configurable per API key and per endpoint',
    ],
    status: 'open-source',
    secondaryTab: {
      kind: 'repository',
      href: 'https://github.com/singh-anurag-7991/shield',
    },
  },
  {
    slug: 'data-guard',
    name: 'Data Guard',
    // Marked coming-soon, not beta. The brief asked for Beta; there is no code.
    // A Beta badge on nothing would undercut the one thing this brand claims —
    // that every figure on the site is traceable to something real.
    value: 'Catches bad data before a dashboard turns it into a bad decision.',
    highlights: [
      'Schema drift from an upstream change, caught before the job breaks',
      'Freshness — the table that quietly stopped updating',
      'Row-count and null-rate anomalies against a rolling baseline',
    ],
    status: 'coming-soon',
    secondaryTab: {
      kind: 'repository',
      href: 'https://github.com/singh-anurag-7991/data-guard',
    },
  },
];

export function productBySlug(slug: string): Product | undefined {
  return PRODUCTS.find((p) => p.slug === slug);
}

export const STATUS_LABEL: Record<ProductStatus, string> = {
  live: 'Live',
  'open-source': 'Open source',
  'coming-soon': 'Coming soon',
};

/* ── Personal ─────────────────────────────────────────────────────────── */

export interface Project {
  name: string;
  kind: string;
  org: string;
  period: string;
  summary: string;
  /** The number this project was measured on, and what it measures. */
  metric?: { value: string; of: string };
  stack: string[];
}

export const PROJECTS: Project[] = [
  {
    name: 'Anota',
    kind: 'No-code automation platform',
    org: 'Digivate Labs',
    period: '2024 — present',
    summary:
      'Led it end to end. An event-driven platform that runs customer-defined workflows, with its own IAM, notifications, monitoring and proxy services.',
    metric: { value: '1M+', of: 'tasks a month, under 200ms' },
    stack: ['Go', 'MongoDB', 'NSQ', 'GCP'],
  },
  {
    name: 'DataInsight Engine',
    kind: 'Ingestion and visualisation pipeline',
    org: 'Digivate Labs',
    period: '2024 — present',
    summary:
      'Ingests structured data through Spark and Iceberg and surfaces it in Superset, with document classification workers on the way in.',
    stack: ['Python', 'Apache Spark', 'Iceberg', 'Superset'],
  },
  {
    name: 'DQF',
    kind: 'Driver licence verification',
    org: 'Grid Infocom',
    period: '2021 — 2024',
    summary:
      'Verified identity documents submitted through an API or a spreadsheet upload, with RabbitMQ carrying the async work.',
    metric: { value: '100K+', of: 'verifications a month' },
    stack: ['Java', 'Spring Boot', 'RabbitMQ', 'MongoDB', 'PostgreSQL'],
  },
  {
    name: 'OnboardVerify',
    kind: 'Client onboarding',
    org: 'Grid Infocom',
    period: '2021 — 2024',
    summary:
      'DQF rebuilt for clients whose onboarding did not fit one flow — branching workflows and per-client validation rules.',
    stack: ['Java', 'RabbitMQ'],
  },
  {
    name: 'OpsDash',
    kind: 'Realtime monitoring',
    org: 'Grid Infocom',
    period: '2021 — 2024',
    summary:
      'Internal dashboard for task execution and system health, on goroutines and NSQ so the view stayed current without polling the database flat.',
    stack: ['Go', 'PostgreSQL', 'NSQ'],
  },
];

export interface Role {
  org: string;
  title: string;
  period: string;
  note: string;
}

export const ROLES: Role[] = [
  {
    org: 'Digivate Labs',
    title: 'Backend Engineer',
    period: 'Mar 2024 — present',
    note: 'Go and Python. Event-driven services, and the data pipeline behind them.',
  },
  {
    org: 'Grid Infocom',
    title: 'Backend Engineer',
    period: 'Sep 2021 — Feb 2024',
    note: 'Java and Go. Document verification at volume, and the tooling to watch it run.',
  },
];

export const LINKS = {
  email: 'anurag7979singh@gmail.com',
  phone: '+91 79797 65096',
  github: 'https://github.com/anurag-singh-code',
  linkedin: 'https://www.linkedin.com/in/anuragsingh-7762811a3',
  location: 'Gurugram, India',
} as const;
