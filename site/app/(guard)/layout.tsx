import type { Metadata } from 'next';
import { guardFontVars } from '../fonts';

/**
 * Guard Infra identity shell.
 *
 * Light surface, deep teal accent, its own typefaces and its own favicon.
 * Nothing here is shared with the personal site beyond the Next.js runtime.
 */

export const metadata: Metadata = {
  title: {
    default: 'Guard Infra — Cloud cost and security intelligence',
    template: '%s — Guard Infra',
  },
  description:
    'Guard Infra finds the cloud spend you forgot about and the exposure you did not know you had. Every figure traced to a resource.',
  icons: { icon: '/guard/favicon.svg' },
  openGraph: {
    siteName: 'Guard Infra',
    type: 'website',
    images: ['/guard/og.png'],
  },
};

export default function GuardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div
      data-identity="guard"
      className={`${guardFontVars} min-h-screen bg-surface text-ink`}
    >
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-accent focus:px-4 focus:py-2 focus:font-medium focus:text-white"
      >
        Skip to content
      </a>
      {children}
    </div>
  );
}
