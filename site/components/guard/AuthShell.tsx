import Link from 'next/link';
import { GuardMark } from '@/components/guard/Chrome';

/**
 * Shared frame for log in, sign up and password reset.
 *
 * Deliberately without the full site header: someone on this page is trying to
 * get in, and a nav bar offering them Products is an invitation to abandon the
 * task. The mark still links home so they are never trapped.
 */
export function AuthShell({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string;
  subtitle: string;
  children: React.ReactNode;
  footer: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-rule">
        <div className="mx-auto flex h-16 max-w-6xl items-center px-6">
          <Link href="/guard" className="flex items-center gap-2.5">
            <GuardMark />
            <span className="font-display text-[15px] font-semibold tracking-tight">
              Guard Infra
            </span>
          </Link>
        </div>
      </header>

      <main id="main" className="flex flex-1 items-center justify-center px-6 py-16">
        <div className="w-full max-w-sm">
          <h1 className="font-display text-2xl font-semibold tracking-tight">
            {title}
          </h1>
          <p className="mt-2 text-sm text-ink-muted">{subtitle}</p>

          <div className="mt-8">{children}</div>

          <div className="mt-8 border-t border-rule pt-5 text-sm text-ink-muted">
            {footer}
          </div>
        </div>
      </main>
    </div>
  );
}

export function Field({
  label,
  name,
  type = 'text',
  autoComplete,
  required = true,
  hint,
}: {
  label: string;
  name: string;
  type?: string;
  autoComplete?: string;
  required?: boolean;
  hint?: string;
}) {
  const id = `field-${name}`;
  const hintId = hint ? `${id}-hint` : undefined;

  return (
    <div>
      <label htmlFor={id} className="block text-sm font-medium">
        {label}
      </label>
      <input
        id={id}
        name={name}
        type={type}
        required={required}
        autoComplete={autoComplete}
        aria-describedby={hintId}
        className="mt-1.5 w-full rounded-lg border border-rule bg-panel px-3 py-2.5 text-[15px] outline-none transition-colors focus:border-accent"
      />
      {hint ? (
        <p id={hintId} className="mt-1.5 text-xs text-ink-muted">
          {hint}
        </p>
      ) : null}
    </div>
  );
}

/**
 * Error banner.
 *
 * The copy rule from the brief: say what went wrong and what to do about it.
 * No apologies, no "oops", nothing vague. "Something went wrong" tells a person
 * nothing and leaves them with no next move.
 */
export function FormError({ message }: { message?: string }) {
  if (!message) return null;
  return (
    <div
      role="alert"
      className="mb-5 rounded-lg border border-alert/30 bg-alert/5 px-4 py-3 text-sm text-alert"
    >
      {message}
    </div>
  );
}

export function GoogleButton({ next }: { next?: string }) {
  const href = next
    ? `/api/auth/google?next=${encodeURIComponent(next)}`
    : '/api/auth/google';

  return (
    <a
      href={href}
      className="flex w-full items-center justify-center gap-2.5 rounded-lg border border-rule bg-panel px-4 py-2.5 text-sm font-medium transition-colors hover:bg-surface-2"
    >
      <svg width="17" height="17" viewBox="0 0 18 18" aria-hidden="true">
        <path fill="#4285F4" d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.92c1.7-1.57 2.68-3.88 2.68-6.62Z" />
        <path fill="#34A853" d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.92-2.26c-.8.54-1.83.86-3.04.86-2.34 0-4.32-1.58-5.03-3.7H.96v2.33A9 9 0 0 0 9 18Z" />
        <path fill="#FBBC05" d="M3.97 10.72a5.4 5.4 0 0 1 0-3.44V4.95H.96a9 9 0 0 0 0 8.1l3.01-2.33Z" />
        <path fill="#EA4335" d="M9 3.58c1.32 0 2.5.45 3.44 1.35l2.58-2.58C13.46.89 11.43 0 9 0A9 9 0 0 0 .96 4.95l3.01 2.33C4.68 5.16 6.66 3.58 9 3.58Z" />
      </svg>
      Continue with Google
    </a>
  );
}
