import Link from 'next/link';
import { CoordinateFrame } from '@/components/personal/CoordinateFrame';
import { SceneMount } from '@/components/SceneMount';
import { LINKS, PROJECTS, ROLES } from '@/lib/content';

/**
 * The poster: a still frame of what the scene resolves into. Deep navy at the
 * pole warming outward, so the moment before WebGL loads already looks
 * deliberate rather than like a page with a missing image.
 */
const STAR_POSTER =
  'radial-gradient(120% 90% at 50% 8%, #1a1030 0%, #0e1428 38%, #0b1020 70%, #070b17 100%)';

/**
 * The personal site. One page, four sections, one way out — into Guard Infra.
 *
 * Copy is written for a hiring CTO: what was built, what it had to survive,
 * what it was measured on. No "passionate about clean code".
 */

export default function PersonalHome() {
  return (
    <>
      <SceneMount scene="star-trails" poster={STAR_POSTER} />
      <CoordinateFrame />

      <div className="above-scene">
        <Nav />

        <main id="main">
          <Hero />
          <WhatIWorkOn />
          <SelectedWork />
          <Experience />
          <EnterGuardInfra />
          <Contact />
        </main>

        <footer className="border-t border-rule px-6 py-8 md:px-[10%]">
          <p className="font-mono text-xs text-ink-muted">
            © 2026 Anurag Kumar Singh · {LINKS.location}
          </p>
        </footer>
      </div>
    </>
  );
}

/* ── Nav ───────────────────────────────────────────────────────────────── */

function Nav() {
  return (
    <nav className="sticky top-0 z-30 border-b border-rule bg-surface/70 backdrop-blur-md">
      <div className="mx-auto flex h-[92px] max-w-[84%] items-center justify-between md:max-w-none md:px-[10%]">
        <Link href="/" className="flex items-center gap-3">
          <Monogram />
          <span className="font-display text-sm tracking-wide">
            Anurag Kumar Singh
          </span>
        </Link>
        <div className="flex items-center gap-1 text-sm">
          <NavLink href="#work">Work</NavLink>
          <NavLink href="#experience">Experience</NavLink>
          <NavLink href="#contact">Contact</NavLink>
        </div>
      </div>
    </nav>
  );
}

function NavLink({ href, children }: { href: string; children: string }) {
  return (
    <a
      href={href}
      className="rounded-md px-3 py-2 text-ink-muted transition-colors hover:bg-white/5 hover:text-ink"
    >
      {children}
    </a>
  );
}

/**
 * Personal monogram — a star-chart aperture: a circle with a crosshair,
 * initials inside. Authored here as inline SVG rather than an image so it
 * inherits the identity's accent colour and stays sharp at any size.
 */
function Monogram() {
  return (
    <svg
      viewBox="0 0 32 32"
      className="h-8 w-8"
      role="img"
      aria-label="Anurag Kumar Singh"
    >
      <circle
        cx="16"
        cy="16"
        r="14"
        fill="none"
        stroke="var(--accent)"
        strokeWidth="1"
        opacity="0.55"
      />
      <path
        d="M16 1v5M16 26v5M1 16h5M26 16h5"
        stroke="var(--accent)"
        strokeWidth="1"
        opacity="0.55"
      />
      <text
        x="16"
        y="20.5"
        textAnchor="middle"
        fontSize="11"
        fontFamily="var(--font-display)"
        fill="var(--ink)"
      >
        AS
      </text>
    </svg>
  );
}

/* ── Hero ──────────────────────────────────────────────────────────────── */

function Hero() {
  return (
    <section className="px-6 pb-24 pt-24 md:px-[10%] md:pb-32 md:pt-32">
      <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-accent">
        Backend · Go · Distributed systems
      </p>

      {/* Wide and light rather than heavy and tight. A big bold headline shouts;
          a wide light one reads as scale, which is the feeling this page wants. */}
      <h1 className="mt-6 max-w-[16ch] font-display text-[clamp(2.6rem,8vw,6rem)] font-normal leading-[0.95] tracking-[-0.045em]">
        I build systems that stay up when the traffic arrives.
      </h1>

      <p className="mt-8 max-w-prose text-lg text-ink-muted">
        Four years of backend work in Go, Java and Python. Most recently an
        event-driven automation platform running{' '}
        <span className="text-ink">a million tasks a month</span> at{' '}
        <span className="text-ink">under 200 milliseconds</span>, and the data
        pipeline that feeds its dashboards.
      </p>

      <div className="mt-10 flex flex-wrap gap-3">
        <a
          href="#work"
          className="rounded-lg bg-accent px-5 py-3 text-sm font-medium text-surface transition-transform hover:-translate-y-0.5"
        >
          See the work
        </a>
        <a
          href={`mailto:${LINKS.email}`}
          className="rounded-lg border border-rule px-5 py-3 text-sm font-medium transition-colors hover:bg-white/5"
        >
          Email me
        </a>
      </div>
    </section>
  );
}

/* ── What I work on ────────────────────────────────────────────────────── */

const FOCUS = [
  {
    title: 'Event-driven services',
    body: 'Queues, workers and the failure modes that only appear at volume — redelivery, ordering, back-pressure, the retry storm that takes down the thing it was retrying.',
  },
  {
    title: 'Data pipelines',
    body: 'Spark and Iceberg for the processing, and the unglamorous part: knowing when a job silently produced the wrong answer instead of failing.',
  },
  {
    title: 'Cloud cost and posture',
    body: 'What AWS is billing you for that nothing is using, and which of it is safe to delete. This became a product.',
  },
];

function WhatIWorkOn() {
  return (
    <section className="border-t border-rule px-6 py-20 md:px-[10%]">
      <h2 className="font-display text-3xl font-normal tracking-tight">
        What I work on
      </h2>
      <div className="mt-10 grid gap-10 md:grid-cols-3">
        {FOCUS.map((f) => (
          <div key={f.title}>
            <h3 className="font-display text-lg font-medium">{f.title}</h3>
            <p className="mt-3 text-sm leading-relaxed text-ink-muted">
              {f.body}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}

/* ── Selected work ─────────────────────────────────────────────────────── */

function SelectedWork() {
  return (
    <section id="work" className="border-t border-rule px-6 py-20 md:px-[10%]">
      <div className="flex items-baseline justify-between gap-6">
        <h2 className="font-display text-3xl font-normal tracking-tight">
          Selected work
        </h2>
        <span className="font-mono text-xs text-ink-muted">
          {PROJECTS.length} systems
        </span>
      </div>

      <ol className="mt-10 divide-y divide-rule border-t border-rule">
        {PROJECTS.map((p, i) => (
          <li
            key={p.name}
            className="grid gap-4 py-8 md:grid-cols-[3rem_1fr_14rem] md:gap-8"
          >
            <span className="font-mono text-xs text-ink-muted">
              {String(i + 1).padStart(2, '0')}
            </span>

            <div>
              <h3 className="font-display text-xl font-medium">{p.name}</h3>
              <p className="mt-0.5 text-sm text-accent-2">{p.kind}</p>
              <p className="mt-3 max-w-prose text-sm leading-relaxed text-ink-muted">
                {p.summary}
              </p>
              <p className="mt-4 font-mono text-[11px] text-ink-muted">
                {p.stack.join(' · ')}
              </p>
            </div>

            <div className="md:text-right">
              {p.metric ? (
                <>
                  <div className="font-mono text-2xl tabular-nums text-accent">
                    {p.metric.value}
                  </div>
                  <div className="mt-1 text-xs text-ink-muted">
                    {p.metric.of}
                  </div>
                </>
              ) : null}
              <div className="mt-3 font-mono text-[11px] text-ink-muted">
                {p.org}
                <br />
                <span className="opacity-70">{p.period}</span>
              </div>
            </div>
          </li>
        ))}
      </ol>
    </section>
  );
}

/* ── Experience ────────────────────────────────────────────────────────── */

function Experience() {
  return (
    <section
      id="experience"
      className="border-t border-rule px-6 py-20 md:px-[10%]"
    >
      <h2 className="font-display text-3xl font-normal tracking-tight">
        Experience
      </h2>
      <div className="mt-10 space-y-8">
        {ROLES.map((r) => (
          <div
            key={r.org}
            className="grid gap-2 border-t border-rule pt-6 md:grid-cols-[1fr_14rem]"
          >
            <div>
              <h3 className="font-display text-lg font-medium">{r.org}</h3>
              <p className="text-sm text-ink-muted">{r.title}</p>
              <p className="mt-2 max-w-prose text-sm text-ink-muted">{r.note}</p>
            </div>
            <p className="font-mono text-xs text-ink-muted md:text-right">
              {r.period}
            </p>
          </div>
        ))}
      </div>

      <p className="mt-10 font-mono text-xs text-ink-muted">
        MCA 2020–22 · BCA 2015–18 · GCP Associate Cloud Engineer · Azure AZ-900
      </p>
    </section>
  );
}

/* ── The one door into Guard Infra ─────────────────────────────────────── */

function EnterGuardInfra() {
  return (
    <section className="border-t border-rule px-6 py-20 md:px-[10%]">
      <Link
        href="/guard"
        className="group block max-w-3xl border border-rule p-8 transition-colors hover:border-accent md:p-12"
      >
        <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-accent">
          Also building
        </p>
        <h2 className="mt-5 font-display text-4xl font-normal tracking-tight md:text-5xl">
          Guard Infra
        </h2>
        <p className="mt-4 max-w-prose text-ink-muted">
          A company for the things infrastructure should tell you and does not.
          Cloud Guard finds the AWS spend nothing is using. Shield keeps one
          client from taking down your API. Data Guard is next.
        </p>
        <span className="mt-8 inline-flex items-center gap-2 text-sm font-medium text-accent">
          Enter Guard Infra
          <span className="transition-transform group-hover:translate-x-1">→</span>
        </span>
      </Link>
    </section>
  );
}

/* ── Contact ───────────────────────────────────────────────────────────── */

function Contact() {
  const items = [
    { label: 'Email', value: LINKS.email, href: `mailto:${LINKS.email}` },
    { label: 'Phone', value: LINKS.phone, href: `tel:${LINKS.phone.replace(/\s/g, '')}` },
    { label: 'GitHub', value: 'anurag-singh-code', href: LINKS.github },
    { label: 'LinkedIn', value: 'anuragsingh-7762811a3', href: LINKS.linkedin },
  ];

  return (
    <section id="contact" className="border-t border-rule px-6 py-20 md:px-[10%]">
      <h2 className="font-display text-3xl font-normal tracking-tight">
        Get in touch
      </h2>
      <p className="mt-4 max-w-prose text-ink-muted">
        Open to senior backend roles, and to freelance work where the problem is
        interesting. Based in {LINKS.location}, IST.
      </p>

      <dl className="mt-10 grid gap-px border border-rule bg-rule sm:grid-cols-2">
        {items.map((c) => (
          <a
            key={c.label}
            href={c.href}
            {...(c.href.startsWith('http')
              ? { target: '_blank', rel: 'noopener noreferrer' }
              : {})}
            className="bg-surface p-5 transition-colors hover:bg-surface-2"
          >
            <dt className="font-mono text-[10px] uppercase tracking-[0.2em] text-ink-muted">
              {c.label}
            </dt>
            <dd className="mt-1.5 text-sm">{c.value}</dd>
          </a>
        ))}
      </dl>
    </section>
  );
}
