'use client';

import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { useEffect, useMemo, useRef } from 'react';

import * as THREE from 'three';

/**
 * Long-exposure star trails.
 *
 * A starfield of dots is the default 3D hero and it is what everyone ships.
 * What a real long exposure of the night sky actually produces is *arcs* —
 * every star sweeping a circle around the celestial pole, warm at the core,
 * with the tail fading behind it. That is the image this scene draws.
 *
 * How it stays cheap enough for a mid-range Android:
 *
 *  - Every arc lives in ONE BufferGeometry drawn as a single LineSegments.
 *    One draw call for the whole sky.
 *  - Trail length is driven by a shader uniform, not by rebuilding geometry.
 *    Scrolling changes one float; the CPU does no work per frame.
 *  - No postprocessing. The warm-core glow is done with additive blending on
 *    the lines themselves, which costs nothing extra.
 *  - Mobile builds a third of the arcs at two-thirds the segments — a
 *    genuinely smaller scene, not the desktop scene scaled down.
 */

interface Config {
  stars: number;
  segments: number;
}

const DESKTOP: Config = { stars: 900, segments: 26 };
const MOBILE: Config = { stars: 300, segments: 16 };

/* ── Shaders ───────────────────────────────────────────────────────────── */

const vertexShader = /* glsl */ `
  attribute float aT;       // 0..1 position along this star's arc
  attribute float aBright;  // per-star brightness
  attribute float aWarm;    // 0 = cool white-blue, 1 = warm amber core

  uniform float uExposure;  // how much of each arc is currently drawn

  varying float vT;
  varying float vBright;
  varying float vWarm;

  void main() {
    vT = aT;
    vBright = aBright;
    vWarm = aWarm;
    gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
  }
`;

const fragmentShader = /* glsl */ `
  precision mediump float;

  uniform float uExposure;
  uniform vec3  uWarm;
  uniform vec3  uCool;

  varying float vT;
  varying float vBright;
  varying float vWarm;

  void main() {
    // Nothing beyond the current exposure has been recorded yet.
    if (vT > uExposure) discard;

    // The head of the trail is bright, the tail fades behind it.
    //
    // This curve was squared at first, which killed the effect: at the top of
    // the page only ~12% of each arc is exposed, and squaring pushed almost all
    // of that to zero alpha, so the sky rendered but looked empty. A gentler
    // ramp with a floor keeps the whole drawn segment visible while the head
    // still clearly leads.
    float age   = uExposure > 0.0 ? clamp(vT / uExposure, 0.0, 1.0) : 0.0;
    float ramp  = 0.30 + 0.70 * pow(age, 1.4);
    float alpha = ramp * vBright;

    vec3 colour = mix(uCool, uWarm, vWarm);

    // Boost the leading end toward white so the star itself reads as a point of
    // light with a tail, rather than a uniformly coloured stroke.
    colour = mix(colour, colour + vec3(0.35), smoothstep(0.82, 1.0, age));

    gl_FragColor = vec4(colour, alpha);
  }
`;

/* ── Geometry ──────────────────────────────────────────────────────────── */

function useTrailGeometry(cfg: Config) {
  return useMemo(() => {
    const { stars, segments } = cfg;
    const vertsPerStar = segments * 2; // LineSegments needs vertex pairs
    const total = stars * vertsPerStar;

    const positions = new Float32Array(total * 3);
    const ts = new Float32Array(total);
    const brights = new Float32Array(total);
    const warms = new Float32Array(total);

    let v = 0;

    for (let s = 0; s < stars; s++) {
      // Each star sits at a fixed angular distance from the pole and sweeps a
      // circle. Distributing the radius as sqrt(random) keeps the density even
      // across the disc — a plain random radius crowds everything at the pole.
      const radius = Math.sqrt(Math.random()) * 26 + 1.2;
      const start = Math.random() * Math.PI * 2;

      // Arcs further from the pole travel further in the same exposure, which
      // is exactly what happens on a real long exposure.
      const sweep = (0.10 + Math.random() * 0.05) * (1 + radius / 26);

      const z = -8 - Math.random() * 26;

      // Bright stars are rare. A flat distribution gives a uniform haze; this
      // gives a few standouts and a lot of faint background, like a real sky.
      // Bright stars are rare, but the floor cannot be near-zero or most of the
      // sky renders as nothing at all. 0.22 is the point where a faint trail is
      // still perceptible against #0b1020 without flattening the range.
      const bright = Math.pow(Math.random(), 2.0) * 0.78 + 0.22;

      // Most stars read white-blue; a minority are warm. The warm ones are what
      // stop this from being the usual cold-blue space scene.
      const warm = Math.random() < 0.34 ? 0.55 + Math.random() * 0.45 : Math.random() * 0.25;

      for (let i = 0; i < segments; i++) {
        const t0 = i / segments;
        const t1 = (i + 1) / segments;

        const a0 = start + t0 * sweep;
        const a1 = start + t1 * sweep;

        positions[v * 3] = Math.cos(a0) * radius;
        positions[v * 3 + 1] = Math.sin(a0) * radius;
        positions[v * 3 + 2] = z;
        ts[v] = t0;
        brights[v] = bright;
        warms[v] = warm;
        v++;

        positions[v * 3] = Math.cos(a1) * radius;
        positions[v * 3 + 1] = Math.sin(a1) * radius;
        positions[v * 3 + 2] = z;
        ts[v] = t1;
        brights[v] = bright;
        warms[v] = warm;
        v++;
      }
    }

    const geo = new THREE.BufferGeometry();
    geo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geo.setAttribute('aT', new THREE.BufferAttribute(ts, 1));
    geo.setAttribute('aBright', new THREE.BufferAttribute(brights, 1));
    geo.setAttribute('aWarm', new THREE.BufferAttribute(warms, 1));
    return geo;
  }, [cfg]);
}

/* ── Scene ─────────────────────────────────────────────────────────────── */

function Trails({ cfg }: { cfg: Config }) {
  const group = useRef<THREE.Group>(null);
  const geometry = useTrailGeometry(cfg);
  const three = useThree();
  const { camera } = three;

  // Temporary instrumentation. A canvas that mounts, holds a live GL context,
  // and issues zero draw calls gives you nothing to go on from the outside —
  // this hands the whole R3F state to the console so the loop, the scene graph
  // and the renderer stats can be read directly.
  const frames = useRef(0);
  useEffect(() => {
    (window as unknown as Record<string, unknown>).__r3f = three;
    (window as unknown as Record<string, unknown>).__frames = frames;
  }, [three]);

  const uniforms = useMemo(
    () => ({
      // Start with a third of each arc drawn. Any lower and the first view is
      // effectively an empty sky, which defeats the point of the scene being
      // the first thing a visitor sees.
      uExposure: { value: 0.34 },
      // Sodium amber and a muted periwinkle — the personal identity's accents,
      // not the electric cyan every space scene reaches for.
      uWarm: { value: new THREE.Color('#e8a33d') },
      uCool: { value: new THREE.Color('#a9bde0') },
    }),
    []
  );

  const material = useMemo(
    () =>
      new THREE.ShaderMaterial({
        uniforms,
        vertexShader,
        fragmentShader,
        transparent: true,
        depthWrite: false,
        blending: THREE.AdditiveBlending,
      }),
    [uniforms]
  );

  // Scroll drives two things: how much of each arc has been "exposed", and how
  // far the camera has travelled. Read in a passive listener into a ref so no
  // React state is touched on a scroll frame.
  const scroll = useRef(0);
  useEffect(() => {
    const read = () => {
      const max = document.body.scrollHeight - window.innerHeight;
      scroll.current = max > 0 ? window.scrollY / max : 0;
    };
    read();
    window.addEventListener('scroll', read, { passive: true });
    window.addEventListener('resize', read);
    return () => {
      window.removeEventListener('scroll', read);
      window.removeEventListener('resize', read);
    };
  }, []);

  const eased = useRef(0);

  useFrame((_state, delta) => {
    frames.current++;
    // Clamp delta: after a background tab wakes up, delta can be several
    // seconds and the sky would visibly jump.
    const dt = Math.min(delta, 0.05);

    eased.current += (scroll.current - eased.current) * Math.min(1, dt * 3);

    // The exposure opens as you read down the page: short arcs at the top,
    // full trails by the bottom.
    uniforms.uExposure.value = 0.34 + eased.current * 0.66;

    if (group.current) {
      // The sky turns. Slowly — this is ambience, not a screensaver.
      group.current.rotation.z += dt * 0.012;
      // A gentle tilt keeps the pole off-centre so the composition is not a
      // bullseye.
      group.current.rotation.x = -0.22 + eased.current * 0.12;
    }

    // Scroll moves the camera *through* the field rather than moving the page
    // past a static backdrop.
    camera.position.z = 14 - eased.current * 9;
    camera.position.y = eased.current * 2.2;
    camera.lookAt(0, 0, -18);
  });

  return (
    <group ref={group}>
      <lineSegments geometry={geometry} material={material} frustumCulled={false} />
    </group>
  );
}

export default function StarTrails() {
  // Decide the scene size once, from the viewport at mount. Resizing a phone
  // into a desktop mid-session is not a case worth rebuilding geometry for.
  const cfg = useMemo<Config>(() => {
    if (typeof window === 'undefined') return MOBILE;
    const small = window.matchMedia('(max-width: 768px)').matches;
    // A device reporting few cores is usually a low-end phone regardless of
    // screen size — a large cheap tablet should get the small scene too.
    const weak = (navigator.hardwareConcurrency ?? 4) <= 4;
    return small || weak ? MOBILE : DESKTOP;
  }, []);

  return (
    <Canvas
      // No sizing classes here. R3F gives the Canvas wrapper an inline
      // width/height of 100%, and the parent .scene-slot is fixed inset-0, so
      // it already fills the viewport. Adding `!absolute inset-0` on top would
      // be fighting inline styles with !important for no gain.
      //
      // Cap dpr at 1.5 rather than the device's full ratio. On a 3x phone the
      // extra pixels are invisible at these line widths but cost real fill rate.
      dpr={[1, 1.5]}
      gl={{ antialias: true, alpha: true, powerPreference: 'high-performance' }}
      camera={{ fov: 62, position: [0, 0, 14], near: 0.1, far: 120 }}
      // Nothing in this scene reacts to the pointer, so let every event fall
      // through to the page beneath.
      style={{ pointerEvents: 'none' }}
    >
      <Trails cfg={cfg} />
    </Canvas>
  );
}
