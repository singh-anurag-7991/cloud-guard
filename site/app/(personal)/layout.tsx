import type { Metadata } from 'next';
import { personalFontVars } from '../fonts';

/**
 * Personal identity shell.
 *
 * `data-identity="personal"` is what activates the navy/plum/sodium token set
 * from globals.css. Guard Infra's layout sets `data-identity="guard"` and the
 * two never meet, so a colour cannot leak from one brand into the other even
 * if a shared component is used in both.
 */

export const metadata: Metadata = {
  title: {
    default: 'Anurag Singh — Backend engineer',
    template: '%s — Anurag Singh',
  },
  description:
    'Backend engineer. Go, Java, distributed systems. Event-driven platforms at 1M+ tasks a month.',
  icons: { icon: '/personal/favicon.svg' },
  openGraph: {
    siteName: 'Anurag Singh',
    type: 'website',
    images: ['/personal/og.png'],
  },
};

export default function PersonalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div
      data-identity="personal"
      className={`${personalFontVars} min-h-screen bg-surface text-ink`}
    >
      {/* Keyboard users reach this before the nav. Without it, tabbing into a
          page means walking the whole header on every navigation. */}
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-accent focus:px-4 focus:py-2 focus:font-medium focus:text-surface"
      >
        Skip to content
      </a>
      {children}
    </div>
  );
}
