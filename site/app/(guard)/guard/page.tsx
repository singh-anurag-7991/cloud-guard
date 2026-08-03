import Link from 'next/link';
import { GuardFooter, GuardHeader } from '@/components/guard/Chrome';
import { Figure } from '@/components/guard/Figure';
import { PRODUCTS, STATUS_LABEL, type Product } from '@/lib/content';
import { getSession } from '@/lib/session';

/**
 * Guard Infra home.
 *
 * Same page signed in or signed out — the product cards swap their action and
 * the header grows an account menu. Building a separate "app shell" home would
 * mean maintaining two versions of the same page and letting them drift.
 */

export default async function GuardHome() {
  const session = await getSession();
  const user = session.authenticated ? { email: session.email! } : null;

  return (
    <>
      <GuardHeader user={user} />

      <main id="main">
        <Hero />
        <ProductGrid signedIn={!!user} />
        <HowItWorks />
      </main>

      <GuardFooter />
    </>
  );
}

/* ── Hero ──────────────────────────────────────────────────────────────── */

function Hero() {
  return (
    <section className="mx-auto max-w-6xl px-6 pb-16 pt-20 md:pt-28">
      <p className="font-mono text-[11px] uppercase tracking-[0.18em] text-accent">
        Cloud cost &amp; security intelligence
      </p>

      <h1 className="mt-5 max-w-[18ch] font-display text-[clamp(2.2rem,5.5vw,3.9rem)] font-semibold leading-[1.02] tracking-[-0.035em]">
        Your cloud bill is paying for things nobody uses.
      </h1>

      <p className="mt-6 max-w-prose text-lg text-ink-muted">
        Guard Infra connects to your account with read-only access, finds what is
        costing money for no reason, and tells you exactly what to do about it —
        with the price, the evidence, and the command that fixes it.
      </p>

      <div className="mt-8 flex flex-wrap gap-3">
        <Link
          href="/guard/signup"
          className="rounded-lg bg-accent px-5 py-3 text-sm font-medium text-white transition-colors hover:bg-accent-2"
        >
          Create an account
        </Link>
        <Link
          href="/guard/cloud-guard"
          className="rounded-lg border border-rule bg-panel px-5 py-3 text-sm font-medium transition-colors hover:bg-surface-2"
        >
          Read the documentation
        </Link>
      </div>

      {/* The ledger line, doing real work rather than decorating. Every figure
          here names where it came from — which is the same promise the product
          makes about every finding it reports. */}
      <div className="mt-16 grid gap-8 sm:grid-cols-3">
        <Figure
          value="9"
          label="Cost and security checks"
          source="internal/rules · counted from source"
        />
        <Figure
          value="0"
          label="Credentials we store"
          source="STS AssumeRole, read-only, no keys held"
        />
        <Figure
          value="$0.08"
          label="Per GiB-month, gp3 — the price we cost your volumes at"
          source="AWS us-east-1 list"
          asOf="1 Aug 2026"
        />
      </div>
    </section>
  );
}

/* ── Products ──────────────────────────────────────────────────────────── */

const STATUS_STYLE: Record<Product['status'], string> = {
  live: 'bg-accent/10 text-accent border-accent/25',
  'open-source': 'bg-ink/5 text-ink-muted border-rule',
  'coming-soon': 'bg-alert/8 text-alert border-alert/20',
};

function ProductGrid({ signedIn }: { signedIn: boolean }) {
  return (
    <section className="mx-auto max-w-6xl px-6 py-16">
      <h2 className="font-display text-2xl font-semibold tracking-tight">
        Products
      </h2>

      <div className="mt-8 grid gap-5 lg:grid-cols-3">
        {PRODUCTS.map((p) => (
          <article
            key={p.slug}
            className="flex flex-col rounded-xl border border-rule bg-panel p-6 transition-colors hover:border-accent/40"
          >
            <div className="flex items-start justify-between gap-3">
              <ProductMark slug={p.slug} />
              <span
                className={`rounded-full border px-2.5 py-1 font-mono text-[10px] uppercase tracking-[0.1em] ${STATUS_STYLE[p.status]}`}
              >
                {STATUS_LABEL[p.status]}
              </span>
            </div>

            <h3 className="mt-5 font-display text-xl font-semibold tracking-tight">
              {p.name}
            </h3>
            <p className="mt-2 text-sm text-ink-muted">{p.value}</p>

            <ul className="mt-5 flex-1 space-y-2">
              {p.highlights.map((h) => (
                <li key={h} className="flex gap-2.5 text-sm text-ink-muted">
                  <span aria-hidden="true" className="mt-[7px] h-1 w-1 shrink-0 rounded-full bg-accent" />
                  {h}
                </li>
              ))}
            </ul>

            <div className="mt-6 border-t border-rule pt-4">
              {p.status === 'live' && signedIn ? (
                <Link
                  href={`/guard/${p.slug}/dashboard`}
                  className="inline-flex rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-2"
                >
                  Open dashboard
                </Link>
              ) : (
                <Link
                  href={`/guard/${p.slug}`}
                  className="text-sm font-medium text-accent hover:underline"
                >
                  {p.status === 'coming-soon' ? 'Read the plan' : 'Read the docs'} →
                </Link>
              )}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}

/**
 * Per-product marks, authored here as SVG.
 *
 * No AWS or GCP trademark artwork anywhere — that is a legal requirement, not a
 * style preference. Each platform is evoked through geometry only.
 */
function ProductMark({ slug }: { slug: string }) {
  const common = { width: 40, height: 40, viewBox: '0 0 40 40', 'aria-hidden': true } as const;

  if (slug === 'cloud-guard') {
    // A perimeter with a scan line crossing it.
    return (
      <svg {...common}>
        <rect x="4" y="4" width="32" height="32" rx="9" fill="var(--accent)" opacity="0.08" />
        <rect x="10" y="10" width="20" height="20" rx="5" fill="none" stroke="var(--accent)" strokeWidth="1.8" />
        <path d="M6 20h28" stroke="var(--accent)" strokeWidth="1.8" strokeLinecap="round" />
      </svg>
    );
  }
  if (slug === 'shield') {
    // Three lanes, one blocked — a limiter.
    return (
      <svg {...common}>
        <rect x="4" y="4" width="32" height="32" rx="9" fill="var(--accent)" opacity="0.08" />
        <path d="M11 14h18M11 20h18M11 26h9" stroke="var(--accent)" strokeWidth="1.8" strokeLinecap="round" />
        <path d="M24 23l5 6M29 23l-5 6" stroke="var(--alert)" strokeWidth="1.8" strokeLinecap="round" />
      </svg>
    );
  }
  // Data Guard: a stack with one layer out of alignment.
  return (
    <svg {...common}>
      <rect x="4" y="4" width="32" height="32" rx="9" fill="var(--alert)" opacity="0.07" />
      <rect x="11" y="12" width="18" height="4" rx="2" fill="none" stroke="var(--ink-muted)" strokeWidth="1.6" />
      <rect x="14" y="18" width="18" height="4" rx="2" fill="none" stroke="var(--alert)" strokeWidth="1.6" />
      <rect x="11" y="24" width="18" height="4" rx="2" fill="none" stroke="var(--ink-muted)" strokeWidth="1.6" />
    </svg>
  );
}

/* ── How it works ──────────────────────────────────────────────────────── */

const STEPS = [
  {
    title: 'Give read-only access',
    body: 'Launch a CloudFormation template that creates a role we can assume and nothing more. No access keys change hands, and we validate the role by actually assuming it — so a wrong ARN fails now, not at scan time.',
  },
  {
    title: 'We scan every region you use',
    body: 'Compute, storage, network and billing, across every region you have enabled — five at a time, so your account is never throttled.',
  },
  {
    title: 'You get findings you can check',
    body: 'Each one names the resource, what we observed, what it costs, and how confident we are. Provable waste is marked high; anything inferred says so and tells you what to verify first.',
  },
];

function HowItWorks() {
  return (
    <section className="border-t border-rule bg-surface-2/40">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <h2 className="font-display text-2xl font-semibold tracking-tight">
          How it works
        </h2>

        <ol className="mt-8 grid gap-8 md:grid-cols-3">
          {STEPS.map((s, i) => (
            <li key={s.title}>
              <div className="border-t-2 border-accent pt-4">
                <span className="font-mono text-xs text-accent">
                  {String(i + 1).padStart(2, '0')}
                </span>
                <h3 className="mt-2 font-display text-lg font-semibold">
                  {s.title}
                </h3>
                <p className="mt-2 text-sm leading-relaxed text-ink-muted">
                  {s.body}
                </p>
              </div>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
