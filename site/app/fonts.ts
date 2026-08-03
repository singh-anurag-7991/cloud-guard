import {
  Archivo,
  IBM_Plex_Mono,
  Instrument_Sans,
  Inter,
  Inter_Tight,
  JetBrains_Mono,
} from 'next/font/google';

/**
 * Six faces, three per identity, loaded through next/font so they are
 * self-hosted and preloaded — no render-blocking request to Google and no
 * flash of fallback text.
 *
 * The two stacks share no family on purpose. Reusing even the body face
 * across both identities is enough to make them feel like one company with
 * two colour schemes, which is exactly the failure this rebuild exists to fix.
 */

// ── Personal ──────────────────────────────────────────────────────────────
// Archivo has a variable width axis. Set wide and light, a headline reads as
// vast rather than loud, which is the feeling the brief asks for.
export const archivo = Archivo({
  subsets: ['latin'],
  weight: ['400', '500', '600'],
  variable: '--font-archivo',
  display: 'swap',
});

export const interTight = Inter_Tight({
  subsets: ['latin'],
  variable: '--font-inter-tight',
  display: 'swap',
});

// Plex Mono reads like instrument output, which suits the coordinate readouts.
export const plexMono = IBM_Plex_Mono({
  subsets: ['latin'],
  weight: ['400', '500'],
  variable: '--font-plex-mono',
  display: 'swap',
});

// ── Guard Infra ───────────────────────────────────────────────────────────
export const instrumentSans = Instrument_Sans({
  subsets: ['latin'],
  variable: '--font-instrument',
  display: 'swap',
});

export const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
});

// JetBrains Mono has tabular figures, so a column of dollar amounts lines up
// on the decimal point. On a page whose whole argument is "check our numbers",
// misaligned figures undercut the point.
export const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  weight: ['400', '500', '700'],
  variable: '--font-jetbrains',
  display: 'swap',
});

export const personalFontVars = [archivo, interTight, plexMono]
  .map((f) => f.variable)
  .join(' ');

export const guardFontVars = [instrumentSans, inter, jetbrainsMono]
  .map((f) => f.variable)
  .join(' ');
