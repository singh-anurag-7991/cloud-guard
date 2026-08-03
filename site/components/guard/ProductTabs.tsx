'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { Product } from '@/lib/content';

/**
 * Tab bar for a product page.
 *
 * Documentation is always first and always the landing tab. Someone arriving at
 * /guard/cloud-guard wants to know what it does before they are asked to sign
 * in — putting a locked dashboard first would turn every first visit into a
 * dead end.
 *
 * Real links, not click handlers: each tab is a URL a reader can bookmark, open
 * in a new tab, or send to a colleague.
 */

interface Props {
  slug: string;
  secondary: Product['secondaryTab'];
}

export function ProductTabs({ slug, secondary }: Props) {
  const pathname = usePathname();

  const docsHref = `/guard/${slug}/docs`;
  // The bare product URL renders the docs tab, so both count as "on docs".
  const onDocs = pathname === `/guard/${slug}` || pathname === docsHref;

  const secondaryHref =
    secondary.kind === 'dashboard' ? `/guard/${slug}/dashboard` : secondary.href;
  const onSecondary =
    secondary.kind === 'dashboard' && pathname === `/guard/${slug}/dashboard`;

  const base =
    'relative -mb-px border-b-2 px-1 pb-3 text-sm font-medium transition-colors';
  const on = 'border-accent text-ink';
  const off = 'border-transparent text-ink-muted hover:text-ink';

  return (
    <nav aria-label="Product sections" className="mt-8 flex gap-7 border-b border-rule">
      <Link href={docsHref} className={`${base} ${onDocs ? on : off}`} aria-current={onDocs ? 'page' : undefined}>
        Documentation
      </Link>

      {secondary.kind === 'dashboard' ? (
        <Link
          href={secondaryHref}
          className={`${base} ${onSecondary ? on : off}`}
          aria-current={onSecondary ? 'page' : undefined}
        >
          Dashboard
        </Link>
      ) : (
        <a
          href={secondaryHref}
          target="_blank"
          rel="noopener noreferrer"
          className={`${base} ${off} inline-flex items-center gap-1.5`}
        >
          Repository
          <span aria-hidden="true" className="text-[11px]">
            ↗
          </span>
          <span className="sr-only">(opens in a new tab)</span>
        </a>
      )}
    </nav>
  );
}
