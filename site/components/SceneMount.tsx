'use client';

import { useEffect, useRef, useState } from 'react';

/**
 * The gate every 3D scene sits behind.
 *
 * Four jobs, all of them about not making the page worse:
 *
 *  1. Reserve the space first. The poster gradient renders immediately at the
 *     final size, so when the canvas arrives nothing moves. A hero that jumps
 *     when WebGL finishes loading is the layout shift Lighthouse punishes and
 *     readers notice.
 *  2. Never load the scene if the reader asked for reduced motion. Not "load
 *     it and pause it" — never fetch the bundle at all.
 *  3. Only mount once the slot is actually on screen, and unmount the whole
 *     canvas when it leaves, which disposes the WebGL context rather than
 *     leaving it idling behind three screens of text.
 *  4. Wait for the browser to be idle. The scene is atmosphere; text and fonts
 *     get the main thread first.
 *
 * Scenes are selected by key rather than by passing an import function in.
 * Pages are Server Components, and React cannot serialise a function across
 * that boundary — `load={() => import(...)}` fails at runtime with
 * "Functions cannot be passed directly to Client Components". Keeping the
 * loader map inside this client module sidesteps that entirely.
 */

const SCENES = {
  'star-trails': () => import('@/components/personal/StarTrails'),
} as const;

export type SceneKey = keyof typeof SCENES;

interface SceneMountProps {
  scene: SceneKey;
  /** CSS gradient shown before the canvas mounts and whenever it is absent. */
  poster: string;
  className?: string;
}

export function SceneMount({ scene, poster, className = '' }: SceneMountProps) {
  const slot = useRef<HTMLDivElement>(null);
  const [Scene, setScene] = useState<React.ComponentType | null>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    const el = slot.current;
    if (!el) return;

    // Seed the state synchronously from the element's own rect.
    //
    // This slot is fixed inset-0, so it is on screen from the first frame — but
    // relying on IntersectionObserver's first async callback to tell us that
    // proved unreliable: the canvas simply never mounted, silently, with no
    // error anywhere. Measuring once ourselves removes that dependency, and the
    // observer below still handles everything after mount.
    const rect = el.getBoundingClientRect();
    if (rect.bottom > 0 && rect.top < window.innerHeight) setVisible(true);

    const io = new IntersectionObserver(
      ([entry]) => setVisible(entry.isIntersecting),
      // Start a little before it scrolls into view so the fade-in has begun by
      // the time the reader gets there.
      { rootMargin: '200px' }
    );
    io.observe(el);
    return () => io.disconnect();
  }, []);

  useEffect(() => {
    if (!visible || Scene) return;

    let cancelled = false;
    const start = () => {
      SCENES[scene]()
        .then((mod) => {
          if (!cancelled) setScene(() => mod.default);
        })
        // Without this the page looks identical whether the scene loaded or
        // threw — which is exactly how a broken scene stayed broken silently.
        // A failed scene is not worth breaking the page over, but it is
        // absolutely worth saying so.
        .catch((err) => {
          console.error(`[SceneMount] scene "${scene}" failed to load`, err);
        });
    };

    // A plain timeout rather than requestIdleCallback. Idle time is scarce on a
    // dev server that is continuously recompiling, and "load the thing the page
    // is built around" should not be scheduled behind everything else. 250ms is
    // past first paint, which is all the deferral this needs.
    const handle = window.setTimeout(start, 250);

    return () => {
      cancelled = true;
      clearTimeout(handle);
    };
  }, [visible, Scene, scene]);

  return (
    <div ref={slot} className={`scene-slot ${className}`} aria-hidden="true">
      <div
        className="absolute inset-0 transition-opacity duration-1000"
        style={{ background: poster, opacity: Scene && visible ? 0 : 1 }}
      />
      {Scene && visible ? <Scene /> : null}
    </div>
  );
}
