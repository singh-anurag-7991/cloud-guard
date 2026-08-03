import { notFound } from 'next/navigation';
import { CodeBlock } from '@/components/guard/CodeBlock';
import { Figure } from '@/components/guard/Figure';
import { productBySlug } from '@/lib/content';

/**
 * Product documentation.
 *
 * Content is authored here for the static shell. S4 replaces the bodies with
 * MDX from content/docs/, keeping this sidebar-plus-prose layout — the shape of
 * the page is deliberate and should survive the swap.
 */

interface Section {
  id: string;
  title: string;
  body: React.ReactNode;
}

export function ProductDocs({ slug }: { slug: string }) {
  const product = productBySlug(slug);
  if (!product) notFound();

  const sections = SECTIONS[slug];
  if (!sections) notFound();

  return (
    <div className="mx-auto grid max-w-6xl gap-10 px-6 py-12 lg:grid-cols-[200px_1fr]">
      {/* On mobile this sits above the content as an inline contents list
          rather than a sticky rail that would eat a third of a 360px screen. */}
      <nav aria-label="On this page" className="lg:sticky lg:top-24 lg:self-start">
        <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-ink-muted">
          On this page
        </p>
        <ul className="mt-3 space-y-2">
          {sections.map((s) => (
            <li key={s.id}>
              <a
                href={`#${s.id}`}
                className="text-sm text-ink-muted transition-colors hover:text-accent"
              >
                {s.title}
              </a>
            </li>
          ))}
        </ul>
      </nav>

      <article className="min-w-0 max-w-prose">
        {sections.map((s) => (
          <section key={s.id} id={s.id} className="scroll-mt-28 pb-12">
            <h2 className="font-display text-xl font-semibold tracking-tight">
              {/* Anchored heading: the link is the heading, so a reader can copy
                  a URL that lands on the exact section they are reading. */}
              <a href={`#${s.id}`} className="group inline-flex items-baseline gap-2">
                {s.title}
                <span
                  aria-hidden="true"
                  className="text-accent opacity-0 transition-opacity group-hover:opacity-100"
                >
                  #
                </span>
              </a>
            </h2>
            <div className="mt-4 space-y-4 text-[15px] leading-relaxed text-ink-muted">
              {s.body}
            </div>
          </section>
        ))}
      </article>
    </div>
  );
}

const B = ({ children }: { children: React.ReactNode }) => (
  <strong className="font-semibold text-ink">{children}</strong>
);

const Note = ({ children }: { children: React.ReactNode }) => (
  <div className="rounded-lg border border-alert/25 bg-alert/5 p-4 text-sm">
    {children}
  </div>
);

const SECTIONS: Record<string, Section[]> = {
  /* ── Cloud Guard ─────────────────────────────────────────────────────── */
  'cloud-guard': [
    {
      id: 'what-it-does',
      title: 'What it does',
      body: (
        <>
          <p>
            Cloud waste is rarely dramatic. It is a 500 GiB volume left behind
            when someone terminated an instance. An Elastic IP nobody released.
            Snapshots a backup script created in 2024 and never cleaned up.
          </p>
          <p>
            Individually invisible on a bill. Together, real money every month —{' '}
            <B>indefinitely</B>, because nothing in AWS will ever suggest you
            delete them.
          </p>
          <p>
            Cloud Guard reads your account through a read-only role, scans every
            region you have enabled, and reports what is costing money for no
            reason.
          </p>
        </>
      ),
    },
    {
      id: 'connect',
      title: 'Connecting an account',
      body: (
        <>
          <ol className="ml-4 list-decimal space-y-2">
            <li>Download the CloudFormation template from your dashboard.</li>
            <li>
              Create the stack in your AWS account and tick the IAM
              acknowledgement.
            </li>
            <li>
              Copy <code className="rounded bg-surface-2 px-1.5 py-0.5 font-mono text-[13px]">RoleARN</code>{' '}
              from the stack Outputs and paste it back.
            </li>
          </ol>
          <p>
            We validate the ARN by actually assuming the role, so a wrong value
            fails immediately with a reason — rather than saving cleanly and
            returning nothing at scan time.
          </p>
          <p>
            The role grants <B>read-only access</B>. No access keys are shared,
            and nothing we hold can change your infrastructure.
          </p>
        </>
      ),
    },
    {
      id: 'checks',
      title: 'What it checks',
      body: (
        <>
          <p>Nine checks. Four of them attach a priced saving:</p>
          <ul className="ml-4 list-disc space-y-2">
            <li>
              <B>Unattached EBS volumes</B> — state is{' '}
              <code className="rounded bg-surface-2 px-1.5 py-0.5 font-mono text-[13px]">available</code>,
              serving no instance, billed every hour.
            </li>
            <li>
              <B>Unassociated Elastic IPs</B> — AWS charges $0.005/hour
              precisely because the address is idle.
            </li>
            <li>
              <B>gp2 volumes that should be gp3</B> — 20% cheaper per GiB, with
              3000 IOPS baseline included. Online change, no downtime.
            </li>
            <li>
              <B>Snapshots past your retention window</B> — with an AMI check
              built into the fix command.
            </li>
          </ul>
          <p>
            The other five cover idle and stopped instances, public S3 buckets,
            over-provisioned RDS, and a weekly spend summary from Cost Explorer.
          </p>
        </>
      ),
    },
    {
      id: 'confidence',
      title: 'Confidence levels',
      body: (
        <>
          <p>
            An unattached volume is provably serving nothing — that is{' '}
            <B>high</B> confidence, and you can act on it. A 90-day-old snapshot
            might back an AMI or satisfy a retention policy — that is{' '}
            <B>check first</B>, and the command we give you runs the AMI lookup
            before the delete.
          </p>
          <p>
            The distinction exists because being wrong confidently costs far more
            trust than being cautious. A tool that once told you to delete
            something you needed never gets opened again.
          </p>
        </>
      ),
    },
    {
      id: 'acting',
      title: 'Acting on a finding',
      body: (
        <>
          <p>
            Every priced finding carries the command that resolves it. Copy it
            from the dashboard and run it — the resource ID and region are
            already filled in.
          </p>
          <CodeBlock label="Release an idle Elastic IP">
            aws ec2 release-address --allocation-id eipalloc-0677… --region us-east-1
          </CodeBlock>
          <CodeBlock label="Delete an unattached volume, backup first">
            {`aws ec2 create-snapshot --volume-id vol-0ab40… --region us-east-1 \\
  --description 'pre-delete backup' && \\
aws ec2 delete-volume --volume-id vol-0ab40… --region us-east-1`}
          </CodeBlock>
          <p>
            Or export the full list as CSV — saving, annual saving, evidence and
            command per row — and send it to whoever owns the account.
          </p>
        </>
      ),
    },
    {
      id: 'pricing-basis',
      title: 'How savings are calculated',
      body: (
        <>
          <div className="grid gap-6 sm:grid-cols-2">
            <Figure
              value="$0.10"
              label="gp2, per GiB-month"
              source="AWS us-east-1 list"
              asOf="1 Aug 2026"
              size="sm"
            />
            <Figure
              value="$0.08"
              label="gp3, per GiB-month"
              source="AWS us-east-1 list"
              asOf="1 Aug 2026"
              size="sm"
            />
          </div>
          <Note>
            <B>What this does not model.</B> Figures use on-demand list prices.
            Savings Plans, Reserved Instances and volume discounts are not
            applied, so your real saving may be lower than quoted. Prices are a
            us-east-1 snapshot; other regions cost more, which makes estimates
            conservative rather than optimistic.
          </Note>
        </>
      ),
    },
  ],

  /* ── Shield ──────────────────────────────────────────────────────────── */
  shield: [
    {
      id: 'what-it-does',
      title: 'What it does',
      body: (
        <>
          <p>
            Every public API eventually meets a client with a retry loop and no
            backoff, a credential-stuffing script, or one customer who
            accidentally sends 40× their normal traffic on a Tuesday morning.
          </p>
          <p>
            Without a limiter, a single caller degrades the service for{' '}
            <B>everyone else</B>. The database saturates, latency climbs, and the
            incident looks like a capacity problem when it was one bad client.
          </p>
          <p>
            Shield sits in front of your API and decides, per API key and per
            endpoint, whether a request is served or gets a{' '}
            <code className="rounded bg-surface-2 px-1.5 py-0.5 font-mono text-[13px]">429</code>.
          </p>
        </>
      ),
    },
    {
      id: 'algorithms',
      title: 'Choosing an algorithm',
      body: (
        <>
          <ul className="ml-4 list-disc space-y-3">
            <li>
              <B>Token bucket</B> — refills at a fixed rate, allows short bursts.
              The default, because bursty traffic is normal traffic.
            </li>
            <li>
              <B>Leaky bucket</B> — drains at a constant rate. For when the thing
              behind your API cannot absorb bursts at all.
            </li>
            <li>
              <B>Fixed window</B> — cheapest to compute. Allows up to double the
              limit across a window boundary, which is fine when the limit is
              generous.
            </li>
            <li>
              <B>Sliding log</B> — most accurate, most memory. No boundary
              loophole to exploit.
            </li>
          </ul>
        </>
      ),
    },
    {
      id: 'responses',
      title: 'Response headers',
      body: (
        <>
          <p>
            Shield returns headers a well-behaved client can act on, so it backs
            off on schedule instead of guessing and retrying into the same wall.
          </p>
          <CodeBlock label="A rejected request">
            {`HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1735689201`}
          </CodeBlock>
        </>
      ),
    },
    {
      id: 'performance',
      title: 'Measured throughput',
      body: (
        <>
          <div className="grid gap-6 sm:grid-cols-2">
            <Figure value="~55K req/s" label="Token bucket" source="local benchmark, wrk -t12 -c400" size="sm" />
            <Figure value="~45K req/s" label="Sliding log" source="local benchmark, wrk -t12 -c400" size="sm" />
          </div>
          <Note>
            <B>Measured locally, not on a production cluster.</B> Distributed
            Redis state, the PostgreSQL rule engine and the gRPC API are on the
            roadmap and not shipped. Treat these numbers as an order of
            magnitude, not a capacity guarantee.
          </Note>
        </>
      ),
    },
    {
      id: 'quick-start',
      title: 'Quick start',
      body: (
        <CodeBlock label="Run it locally">
          {`git clone https://github.com/singh-anurag-7991/shield.git
cd shield && go mod tidy
go run cmd/server/main.go   # listens on :8080`}
        </CodeBlock>
      ),
    },
  ],

  /* ── Data Guard ──────────────────────────────────────────────────────── */
  'data-guard': [
    {
      id: 'status',
      title: 'Status',
      body: (
        <Note>
          <B>Nothing to try yet.</B> This page describes a design, not a
          product. There is no demo, no signup and no waitlist. When it runs,
          this page will say so and the tab will open something.
        </Note>
      ),
    },
    {
      id: 'problem',
      title: 'The problem',
      body: (
        <>
          <p>
            Pipelines rarely fail loudly. A column changes type upstream. A job
            silently stops at 3am. A join starts dropping 12% of rows.
          </p>
          <p>
            Nothing errors. The dashboard keeps rendering. People keep making
            decisions from it — for a week, until somebody notices a number looks
            off and the team spends two days working out how far back the damage
            goes.
          </p>
          <p>
            The failures that cost the most are the ones that{' '}
            <B>look like success</B>.
          </p>
        </>
      ),
    },
    {
      id: 'planned-checks',
      title: 'What it will watch',
      body: (
        <ul className="ml-4 list-disc space-y-2">
          <li>
            <B>Schema drift</B> — a column added, removed or retyped upstream,
            caught before it breaks the job downstream.
          </li>
          <li>
            <B>Freshness</B> — the table that should update hourly and last
            updated eleven hours ago.
          </li>
          <li>
            <B>Row-count anomalies</B> — a load that normally brings 2M rows and
            quietly brought 240K.
          </li>
          <li>
            <B>Null and distribution shifts</B> — a field that was 2% null
            yesterday and is 60% null today.
          </li>
        </ul>
      ),
    },
    {
      id: 'approach',
      title: 'Approach',
      body: (
        <p>
          Built on the same Spark and Iceberg foundation as the DataInsight
          Engine, reusing Cloud Guard&apos;s model of a finding: the observation,
          the severity, and something specific you can do about it. Same idea, a
          different layer of the stack.
        </p>
      ),
    },
  ],
};
