import Link from 'next/link';

/**
 * Guard Infra's header and footer.
 *
 * The one hard rule from the brief lives here: Guard Infra does not link to the
 * personal site anywhere in its navigation or body. The only connection is a
 * single line in the footer. A visitor evaluating a security product should not
 * keep bumping into somebody's CV.
 */

/** The Guard Infra mark: a perimeter with a gap being closed. */
export function GuardMark({ size = 30 }: { size?: number }) {
  return (
    <svg
      viewBox="0 0 32 32"
      width={size}
      height={size}
      role="img"
      aria-label="Guard Infra"
      className="shrink-0"
    >
      <rect
        x="3.5"
        y="3.5"
        width="25"
        height="25"
        rx="6"
        fill="none"
        stroke="var(--accent)"
        strokeWidth="2"
        strokeDasharray="62 38"
        strokeDashoffset="-8"
      />
      <rect x="12" y="12" width="8" height="8" rx="2" fill="var(--accent)" />
    </svg>
  );
}

interface HeaderProps {
  /** Null when signed out. */
  user?: { email: string } | null;
}

export function GuardHeader({ user = null }: HeaderProps) {
  return (
    <header className="sticky top-0 z-40 border-b border-rule bg-surface/85 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link href="/guard" className="flex items-center gap-2.5">
          <GuardMark />
          <span className="font-display text-[15px] font-semibold tracking-tight">
            Guard Infra
          </span>
        </Link>

        <nav className="flex items-center gap-1 text-sm">
          <Link
            href="/guard/cloud-guard"
            className="rounded-md px-3 py-2 text-ink-muted transition-colors hover:bg-surface-2 hover:text-ink"
          >
            Products
          </Link>

          {user ? (
            <details className="relative">
              <summary className="cursor-pointer list-none rounded-md px-3 py-2 text-ink-muted transition-colors hover:bg-surface-2 hover:text-ink">
                {user.email}
              </summary>
              <div className="absolute right-0 mt-1 w-56 rounded-lg border border-rule bg-panel p-1 shadow-lg">
                <Link
                  href="/guard/cloud-guard/dashboard"
                  className="block rounded-md px-3 py-2 text-sm hover:bg-surface-2"
                >
                  Open dashboard
                </Link>
                <form action="/api/auth/logout" method="post">
                  <button
                    type="submit"
                    className="w-full rounded-md px-3 py-2 text-left text-sm hover:bg-surface-2"
                  >
                    Sign out
                  </button>
                </form>
              </div>
            </details>
          ) : (
            <>
              <Link
                href="/guard/login"
                className="rounded-md px-3 py-2 text-ink-muted transition-colors hover:bg-surface-2 hover:text-ink"
              >
                Log in
              </Link>
              <Link
                href="/guard/signup"
                className="rounded-md bg-accent px-4 py-2 font-medium text-white transition-colors hover:bg-accent-2"
              >
                Sign up
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}

export function GuardFooter() {
  return (
    <footer className="mt-24 border-t border-rule">
      <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 px-6 py-8 text-sm text-ink-muted">
        <p>© 2026 Guard Infra</p>
        {/* The only route from Guard Infra to the personal site, per the brief.
            New tab so the visitor does not lose the product they were reading. */}
        <p>
          Built by{' '}
          <a
            href="/"
            target="_blank"
            rel="noopener noreferrer"
            className="text-ink underline decoration-rule underline-offset-4 hover:decoration-accent"
          >
            Anurag Singh
          </a>
        </p>
      </div>
    </footer>
  );
}
