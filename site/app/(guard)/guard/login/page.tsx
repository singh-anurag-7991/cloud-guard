import Link from 'next/link';
import type { Metadata } from 'next';
import {
  AuthShell,
  Field,
  FormError,
  GoogleButton,
} from '@/components/guard/AuthShell';

export const metadata: Metadata = {
  title: 'Log in',
  // Sign-in pages have nothing to offer a search engine and everything to lose
  // if one indexes a URL carrying a ?next= parameter.
  robots: { index: false, follow: false },
};

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ next?: string; error?: string }>;
}) {
  const { next, error } = await searchParams;

  // Only relative paths are honoured. Reflecting an arbitrary ?next= into a
  // redirect is an open-redirect: an attacker sends a link to our real login
  // page and bounces the victim to a lookalike afterwards.
  const safeNext = next && next.startsWith('/') && !next.startsWith('//') ? next : '/guard';

  return (
    <AuthShell
      title="Log in"
      subtitle="Your findings and connected accounts are waiting."
      footer={
        <>
          No account yet?{' '}
          <Link
            href={`/guard/signup${next ? `?next=${encodeURIComponent(safeNext)}` : ''}`}
            className="text-accent hover:underline"
          >
            Create one
          </Link>
        </>
      }
    >
      <FormError message={ERRORS[error ?? '']} />

      <form action="/api/auth/login" method="post" className="space-y-4">
        <input type="hidden" name="next" value={safeNext} />

        <Field label="Email" name="email" type="email" autoComplete="email" />
        <Field
          label="Password"
          name="password"
          type="password"
          autoComplete="current-password"
        />

        <button
          type="submit"
          className="w-full rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-2"
        >
          Log in
        </button>
      </form>

      <div className="my-5 flex items-center gap-3 text-xs text-ink-muted">
        <span className="h-px flex-1 bg-rule" />
        or
        <span className="h-px flex-1 bg-rule" />
      </div>

      <GoogleButton next={safeNext} />

      <p className="mt-5 text-sm">
        <Link href="/guard/forgot-password" className="text-ink-muted hover:text-accent">
          Forgot your password?
        </Link>
      </p>
    </AuthShell>
  );
}

/**
 * Error copy.
 *
 * Each one says what happened and what to do next. Note that a wrong email and
 * a wrong password give the same message on purpose — telling an attacker which
 * half was correct confirms whether an address is registered.
 */
const ERRORS: Record<string, string | undefined> = {
  invalid: 'That email and password combination did not match. Check both and try again.',
  expired: 'Your session ended after seven days. Log in again to continue.',
  google: 'Google sign-in did not complete. Try again, or use your email and password.',
};
