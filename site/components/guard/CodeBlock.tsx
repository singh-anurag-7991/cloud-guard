'use client';

import { useState } from 'react';

/**
 * A copyable code block.
 *
 * Documentation that shows a command you cannot copy is documentation that
 * makes people retype commands, and retyped commands get typos in resource IDs.
 */
export function CodeBlock({
  children,
  label,
}: {
  children: string;
  label?: string;
}) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    const done = () => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    };
    // navigator.clipboard is undefined on plain HTTP and in older browsers.
    // Falling back to a prompt beats a button that silently does nothing.
    if (navigator.clipboard && window.isSecureContext) {
      try {
        await navigator.clipboard.writeText(children);
        done();
        return;
      } catch {
        /* fall through */
      }
    }
    window.prompt('Copy this command:', children);
  };

  return (
    <figure className="overflow-hidden rounded-lg border border-rule bg-surface-2">
      <figcaption className="flex items-center justify-between gap-3 border-b border-rule px-4 py-2">
        <span className="font-mono text-[11px] uppercase tracking-[0.1em] text-ink-muted">
          {label ?? 'Shell'}
        </span>
        <button
          type="button"
          onClick={copy}
          className="rounded-md px-2 py-1 font-mono text-[11px] text-accent transition-colors hover:bg-accent/10"
        >
          {copied ? 'Copied' : 'Copy'}
        </button>
      </figcaption>
      <pre className="overflow-x-auto px-4 py-3 font-mono text-[13px] leading-relaxed text-ink">
        <code>{children}</code>
      </pre>
    </figure>
  );
}
