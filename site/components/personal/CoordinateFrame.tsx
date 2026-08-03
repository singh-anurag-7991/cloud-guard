'use client';

import { useEffect, useRef } from 'react';

/**
 * The personal site's signature element.
 *
 * A hairline frame with right-ascension / declination readouts that change as
 * you scroll. The page stops feeling like a document and starts feeling like a
 * viewport on something that is moving — which is the "vast, exploratory"
 * brief, delivered by typography and one thin rule rather than by another
 * particle field.
 *
 * Implementation notes that matter:
 *  - It is `aria-hidden` and `pointer-events-none`. This is atmosphere. A
 *    screen reader announcing "RA 04h 17m" between paragraphs would be noise.
 *  - The scroll handler writes to refs and updates text inside a rAF callback,
 *    never through React state. Re-rendering a component on every scroll frame
 *    is the classic way a beautiful page becomes a janky one.
 */
export function CoordinateFrame() {
  const raRef = useRef<HTMLSpanElement>(null);
  const decRef = useRef<HTMLSpanElement>(null);
  const depthRef = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    // Respect the OS setting: freeze the readout at its start value.
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    let frame = 0;
    let queued = false;

    const paint = () => {
      queued = false;
      const max = document.body.scrollHeight - window.innerHeight;
      const t = max > 0 ? Math.min(1, Math.max(0, window.scrollY / max)) : 0;

      // Arbitrary but internally consistent: RA sweeps through a couple of
      // hours over the length of the page, Dec climbs, depth increases. The
      // numbers are decorative, so they are formatted plausibly rather than
      // computed from a real ephemeris — and they are never presented as data.
      const raH = 2 + Math.floor(t * 3);
      const raM = Math.floor((t * 180) % 60);
      const dec = 12 + t * 47;
      const depth = 0.4 + t * 12.6;

      if (raRef.current) {
        raRef.current.textContent = `${String(raH).padStart(2, '0')}h ${String(raM).padStart(2, '0')}m`;
      }
      if (decRef.current) {
        decRef.current.textContent = `+${dec.toFixed(0)}° ${String(Math.floor((dec % 1) * 60)).padStart(2, '0')}′`;
      }
      if (depthRef.current) {
        depthRef.current.textContent = `${depth.toFixed(1)} Mly`;
      }
    };

    const onScroll = () => {
      if (queued) return;
      queued = true;
      frame = requestAnimationFrame(paint);
    };

    paint();
    if (!reduced) {
      window.addEventListener('scroll', onScroll, { passive: true });
    }
    return () => {
      window.removeEventListener('scroll', onScroll);
      cancelAnimationFrame(frame);
    };
  }, []);

  return (
    <div
      aria-hidden="true"
      className="pointer-events-none fixed inset-0 z-[2] hidden select-none md:block"
    >
      {/* Hairline verticals. Placed off the golden-ish thirds rather than at
          25/50/75 so the frame reads as an instrument overlay, not a grid. */}
      <div className="absolute inset-y-0 left-[8%] w-px bg-rule opacity-40" />
      <div className="absolute inset-y-0 right-[8%] w-px bg-rule opacity-40" />
      <div className="absolute inset-x-0 top-[92px] h-px bg-rule opacity-25" />
      <div className="absolute inset-x-0 bottom-[64px] h-px bg-rule opacity-25" />

      {/* Corner ticks — short marks where the rules cross. */}
      {[
        'left-[8%] top-[92px]',
        'right-[8%] top-[92px]',
        'left-[8%] bottom-[64px]',
        'right-[8%] bottom-[64px]',
      ].map((pos) => (
        <span
          key={pos}
          className={`absolute ${pos} h-2 w-2 -translate-x-1/2 -translate-y-1/2 border-l border-t border-accent opacity-50`}
        />
      ))}

      <div className="absolute left-[8%] top-[92px] -translate-y-full pb-2 pl-3 font-mono text-[10px] uppercase tracking-[0.2em] text-ink-muted">
        RA <span ref={raRef} className="text-accent">02h 00m</span>
      </div>

      <div className="absolute right-[8%] top-[92px] -translate-y-full pb-2 pr-3 text-right font-mono text-[10px] uppercase tracking-[0.2em] text-ink-muted">
        DEC <span ref={decRef} className="text-accent">+12° 00′</span>
      </div>

      <div className="absolute bottom-[64px] left-[8%] translate-y-full pl-3 pt-2 font-mono text-[10px] uppercase tracking-[0.2em] text-ink-muted">
        DEPTH <span ref={depthRef} className="text-accent">0.4 Mly</span>
      </div>

      <div className="absolute bottom-[64px] right-[8%] translate-y-full pr-3 pt-2 text-right font-mono text-[10px] uppercase tracking-[0.2em] text-ink-muted opacity-70">
        Long exposure · 1/1
      </div>
    </div>
  );
}
