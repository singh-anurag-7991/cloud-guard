import type { Metadata } from 'next';
import './globals.css';

/**
 * Root layout.
 *
 * Deliberately thin: it emits <html> and <body> and nothing else. All colour,
 * typography and chrome belong to the two route-group layouts, because the
 * moment something visual lives here it is shared by both identities — which
 * is how a two-brand site quietly collapses back into one.
 */

export const metadata: Metadata = {
  // Per-page metadata overrides this. It exists so a route that forgets to set
  // its own still ships something truthful rather than "Create Next App".
  title: 'Guard Infra',
  metadataBase: new URL(
    process.env.NEXT_PUBLIC_SITE_ORIGIN ?? 'https://guardinfra.duckdns.org'
  ),
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
