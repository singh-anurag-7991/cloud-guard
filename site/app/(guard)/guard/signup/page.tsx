import Link from 'next/link';
import type { Metadata } from 'next';
import {
  AuthShell,
  Field,
  FormError,
  GoogleButton,
} from '@/components/guard/AuthShell';

export const metadata: Metadata = {
  title: 'Sign up',
  robots: { index: false, follow: false },
};

export default async function SignupPage({
  searchParams,
}: {
  searchParams: Promise<{ next?: string; error?: string }>;
}) {
  const { next, error } = await searchParams;
  const safeNext = next && next.startsWith('/') && !next.startsWith('//') ? next : '/guard';

  return (
    <AuthShell
      title="Create an account"
      subtitle="Connect an AWS account and see what it is paying for. Free while Cloud Guard is in early access."
      footer={
        <>
          Already have an account?{' '}
          <Link
            href={`/guard/login${next ? `?next=${encodeURIComponent(safeNext)}` : ''}`}
            className="text-accent hover:underline"
          >
            Log in
          </Link>
        </>
      }
    >
      <FormError message={ERRORS[error ?? '']} />

      <form action="/api/auth/signup" method="post" className="space-y-4">
        <input type="hidden" name="next" value={safeNext} />

        <Field label="Email" name="email" type="email" autoComplete="email" />
        <Field
          label="Password"
          name="password"
          type="password"
          autoComplete="new-password"
          hint="At least 10 characters. Length beats complexity — a passphrase is fine."
        />

        <button
          type="submit"
          className="w-full rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-2"
        >
          Create account
        </button>
      </form>

      <div className="my-5 flex items-center gap-3 text-xs text-ink-muted">
        <span className="h-px flex-1 bg-rule" />
        or
        <span className="h-px flex-1 bg-rule" />
      </div>

      <GoogleButton next={safeNext} />

      <p className="mt-6 text-xs leading-relaxed text-ink-muted">
        Signing up creates an organisation for you. Cloud Guard only ever reads
        your AWS account, through a role you create and can revoke at any time.
      </p>
    </AuthShell>
  );
}

const ERRORS: Record<string, string | undefined> = {
  taken: 'An account already exists for that email. Log in instead, or reset your password.',
  weak: 'That password is under 10 characters. Add a few more — a short phrase works well.',
  email: 'That does not look like an email address. Check for a typo.',
  google: 'Google sign-in did not complete. Try again, or create an account with email and password.',
};
