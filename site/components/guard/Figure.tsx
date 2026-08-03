/**
 * Guard Infra's signature element: the ledger line.
 *
 * Every number on this side of the site sits on a ruled baseline and carries a
 * provenance tag — where the figure came from and when. The brand statement is
 * "we show our working", which is not decoration: it is what the product
 * actually does. Findings carry evidence, a confidence level and a pricing
 * source, so the marketing site is built the same way.
 *
 * The rule is that a <Figure> without a `source` will not compile. If a number
 * cannot say where it came from, it does not belong on the page.
 */

interface FigureProps {
  /** The number itself, already formatted. */
  value: string;
  /** What it measures, in the reader's language. */
  label: string;
  /** Where this came from. Required, deliberately. */
  source: string;
  /** When it was captured or measured. */
  asOf?: string;
  size?: 'sm' | 'md' | 'lg';
}

const SIZE = {
  sm: 'text-xl',
  md: 'text-3xl',
  lg: 'text-5xl',
} as const;

export function Figure({ value, label, source, asOf, size = 'md' }: FigureProps) {
  return (
    <div className="border-t border-rule pt-3">
      <div
        className={`font-mono font-medium tabular-nums leading-none text-ink ${SIZE[size]}`}
      >
        {value}
      </div>
      <div className="mt-2 text-sm text-ink">{label}</div>
      {/* The provenance line. Small, always present, never a tooltip — a fact
          you have to hover to find reads as a disclaimer being hidden. */}
      <div className="mt-1 font-mono text-[11px] leading-relaxed text-ink-muted">
        {source}
        {asOf ? <span className="opacity-70"> · {asOf}</span> : null}
      </div>
    </div>
  );
}

/**
 * Inline variant for figures inside running prose.
 */
export function InlineFigure({
  value,
  source,
}: {
  value: string;
  source: string;
}) {
  return (
    <span className="whitespace-nowrap">
      <span className="font-mono font-medium tabular-nums">{value}</span>
      <span className="ml-1.5 font-mono text-[11px] text-ink-muted">
        ({source})
      </span>
    </span>
  );
}
