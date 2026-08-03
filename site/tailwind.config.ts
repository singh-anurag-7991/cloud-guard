import type { Config } from 'tailwindcss';

/**
 * Tailwind reads colours and typefaces from CSS custom properties rather than
 * hard-coding hex values here.
 *
 * That is the whole mechanism keeping the two identities apart: `bg-surface`
 * resolves to navy inside the personal route group and to cool paper inside
 * Guard Infra, because each group's layout sets a different value for
 * `--surface`. If the palettes lived in this file we would need two prefixed
 * sets of utilities and they would leak into each other within a week.
 */
const config: Config = {
  content: [
    './app/**/*.{ts,tsx,mdx}',
    './components/**/*.{ts,tsx,mdx}',
    './content/**/*.mdx',
  ],
  theme: {
    extend: {
      colors: {
        surface: 'var(--surface)',
        'surface-2': 'var(--surface-2)',
        panel: 'var(--panel)',
        ink: 'var(--ink)',
        'ink-muted': 'var(--ink-muted)',
        rule: 'var(--rule)',
        accent: 'var(--accent)',
        'accent-2': 'var(--accent-2)',
        alert: 'var(--alert)',
      },
      fontFamily: {
        display: 'var(--font-display)',
        body: 'var(--font-body)',
        mono: 'var(--font-mono)',
      },
      maxWidth: {
        prose: '68ch',
      },
    },
  },
  plugins: [],
};

export default config;
