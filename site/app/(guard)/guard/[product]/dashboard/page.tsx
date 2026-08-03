import Link from 'next/link';
import { notFound } from 'next/navigation';
import { productBySlug } from '@/lib/content';
import { getSession } from '@/lib/session';

/**
 * The Dashboard tab.
 *
 * Signed out, this renders a sign-in prompt *in place of the tab content* — it
 * does not redirect. A redirect throws away the page the reader chose, and
 * after signing in they land somewhere else and have to find their way back.
 * Keeping them here means signing in returns them to exactly this tab.
 *
 * The real dashboard arrives in S5, reading from the Go API.
 */
export default async function DashboardTab({
  params,
}: {
  params: Promise<{ product: string }>;
}) {
  const { product } = await params;
  const p = productBySlug(product);
  if (!p || p.secondaryTab.kind !== 'dashboard') notFound();

  const session = await getSession();

  if (!session.authenticated) {
    const next = encodeURIComponent(`/guard/${product}/dashboard`);
    return (
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="max-w-md rounded-xl border border-rule bg-panel p-8">
          <h2 className="font-display text-xl font-semibold tracking-tight">
            Sign in to see your findings
          </h2>
          <p className="mt-3 text-sm text-ink-muted">
            The dashboard shows what {p.name} found in your own AWS account —
            priced, with evidence, and the command that fixes each one.
          </p>

          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              href={`/guard/login?next=${next}`}
              className="rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-2"
            >
              Log in
            </Link>
            <Link
              href={`/guard/signup?next=${next}`}
              className="rounded-lg border border-rule px-4 py-2.5 text-sm font-medium transition-colors hover:bg-surface-2"
            >
              Create an account
            </Link>
          </div>

          <p className="mt-6 border-t border-rule pt-4 text-sm text-ink-muted">
            Not ready to sign up?{' '}
            <Link href={`/guard/${product}/docs`} className="text-accent hover:underline">
              Read the documentation
            </Link>{' '}
            — it explains everything the dashboard does.
          </p>
        </div>
      </div>
    );
  }

  // Signed in, but the findings view has not been ported yet.
  //
  // Rendering nothing here left an authenticated user staring at a blank panel
  // with no explanation — indistinguishable from a page that crashed. Until S5
  // lands, say plainly where the dashboard is and give them the working one.
  return (
    <div className="mx-auto max-w-6xl px-6 py-16">
      <div className="max-w-xl rounded-xl border border-rule bg-panel p-8">
        <p className="font-mono text-[11px] uppercase tracking-[0.16em] text-alert">
          Being rebuilt
        </p>
        <h2 className="mt-3 font-display text-xl font-semibold tracking-tight">
          Your findings are not on this page yet
        </h2>
        <p className="mt-3 text-sm leading-relaxed text-ink-muted">
          Signed in as <span className="text-ink">{session.email}</span>. The
          scanning, pricing and findings all work — this new dashboard view is
          still being wired up to them.
        </p>
        <p className="mt-3 text-sm leading-relaxed text-ink-muted">
          The existing dashboard has everything: connected accounts, priced
          findings with evidence, copyable fix commands and CSV export. It uses
          the same account you just signed in with.
        </p>

        <div className="mt-6 flex flex-wrap gap-3">
          <a
            href="/dashboard"
            className="rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-2"
          >
            Open the current dashboard
          </a>
          <Link
            href={`/guard/${product}/docs`}
            className="rounded-lg border border-rule px-4 py-2.5 text-sm font-medium transition-colors hover:bg-surface-2"
          >
            Read the documentation
          </Link>
        </div>
      </div>
    </div>
  );
}
