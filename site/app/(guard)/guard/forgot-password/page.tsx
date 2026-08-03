import Link from 'next/link';
import type { Metadata } from 'next';
import { AuthShell, Field, FormError } from '@/components/guard/AuthShell';

export const metadata: Metadata = {
  title: 'Reset your password',
  robots: { index: false, follow: false },
};

export default async function ForgotPasswordPage({
  searchParams,
}: {
  searchParams: Promise<{ sent?: string; error?: string }>;
}) {
  const { sent, error } = await searchParams;

  if (sent) {
    return (
      <AuthShell
        title="Check your email"
        subtitle="If an account exists for that address, a reset link is on its way. The link works once and expires in an hour."
        footer={
          <Link href="/guard/login" className="text-accent hover:underline">
            Back to log in
          </Link>
        }
      >
        {/* Deliberately the same message whether or not the address is
            registered. Saying "no account found" would let anyone check which
            email addresses have accounts here. */}
        <p className="text-sm text-ink-muted">
          Nothing arrived after a few minutes? Check spam, then try again — the
          address may have a typo.
        </p>
      </AuthShell>
    );
  }

  return (
    <AuthShell
      title="Reset your password"
      subtitle="Enter the email you signed up with and we'll send a reset link."
      footer={
        <Link href="/guard/login" className="text-accent hover:underline">
          Back to log in
        </Link>
      }
    >
      <FormError message={ERRORS[error ?? '']} />

      <form action="/api/auth/forgot-password" method="post" className="space-y-4">
        <Field label="Email" name="email" type="email" autoComplete="email" />
        <button
          type="submit"
          className="w-full rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-2"
        >
          Send reset link
        </button>
      </form>
    </AuthShell>
  );
}

const ERRORS: Record<string, string | undefined> = {
  email: 'That does not look like an email address. Check for a typo.',
  rate: 'A reset link was sent recently. Wait a few minutes before requesting another.',
};
