import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import { GuardFooter, GuardHeader } from '@/components/guard/Chrome';
import { ProductTabs } from '@/components/guard/ProductTabs';
import { PRODUCTS, STATUS_LABEL, productBySlug } from '@/lib/content';
import { getSession } from '@/lib/session';

/**
 * Shell for every product page: header, product title block, tab bar.
 *
 * The tab bar lives in the layout rather than each page so switching tabs never
 * repaints the product header — and so the tabs cannot drift out of sync
 * between the default view and the /docs view.
 */

export function generateStaticParams() {
  return PRODUCTS.map((p) => ({ product: p.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ product: string }>;
}): Promise<Metadata> {
  const { product } = await params;
  const p = productBySlug(product);
  if (!p) return {};
  return {
    title: p.name,
    description: p.value,
    openGraph: { title: `${p.name} — Guard Infra`, description: p.value },
  };
}

export default async function ProductLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ product: string }>;
}) {
  const { product } = await params;
  const p = productBySlug(product);
  if (!p) notFound();

  const badge =
    p.status === 'live'
      ? 'bg-accent/10 text-accent border-accent/25'
      : p.status === 'coming-soon'
        ? 'bg-alert/8 text-alert border-alert/20'
        : 'bg-ink/5 text-ink-muted border-rule';

  const session = await getSession();

  return (
    <>
      <GuardHeader user={session.authenticated ? { email: session.email! } : null} />

      <div className="border-b border-rule">
        <div className="mx-auto max-w-6xl px-6 pb-0 pt-12">
          <nav aria-label="Breadcrumb" className="font-mono text-xs text-ink-muted">
            <a href="/guard" className="hover:text-ink">
              Guard Infra
            </a>
            <span aria-hidden="true"> / </span>
            <span className="text-ink">{p.name}</span>
          </nav>

          <div className="mt-4 flex flex-wrap items-center gap-4">
            <h1 className="font-display text-[clamp(1.9rem,4vw,2.7rem)] font-semibold tracking-[-0.03em]">
              {p.name}
            </h1>
            <span
              className={`rounded-full border px-2.5 py-1 font-mono text-[10px] uppercase tracking-[0.1em] ${badge}`}
            >
              {STATUS_LABEL[p.status]}
            </span>
          </div>

          <p className="mt-3 max-w-prose text-ink-muted">{p.value}</p>

          <ProductTabs slug={p.slug} secondary={p.secondaryTab} />
        </div>
      </div>

      <main id="main">{children}</main>

      <GuardFooter />
    </>
  );
}
