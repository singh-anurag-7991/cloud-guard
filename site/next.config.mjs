/**
 * Paths the Go binary still owns.
 *
 * Everything else is served by Next.js. As each of these is ported, delete the
 * entry here *and* the matching rule in the Caddyfile — leaving one behind
 * means a route that works locally and 404s in production, or the reverse.
 */
const LEGACY_GO_PATHS = [
  '/api/:path*',
  '/dashboard',
  '/connect',
  '/disconnect',
  '/scan',
  '/logout',
  '/findings.csv',
  '/cloudformation.yaml',
  '/healthz',
];

/** @type {import('next').NextConfig} */
const nextConfig = {
  // standalone emits a self-contained server bundle with only the node_modules
  // it actually uses. The production image is ~120 MB instead of ~1 GB, which
  // matters on a 2 GB box that is already running Go and Caddy.
  output: 'standalone',

  // Required for React Three Fiber under the App Router.
  //
  // Without this the Canvas mounts, WebGL initialises, the element is the right
  // size — and its children never render. No error, no warning, just an empty
  // canvas issuing zero draw calls. The cause is module duplication: Next
  // bundles its own React copy for server components, and R3F's reconciler can
  // bind to a different React instance than the one rendering the app, so the
  // three.js scene graph is built into a root nothing ever paints.
  //
  // Transpiling these through the app's own build makes them share a single
  // React instance.
  transpilePackages: ['three', '@react-three/fiber'],

  reactStrictMode: true,

  // In development Next runs on :3000 and the Go API on :8080. Rewriting keeps
  // the browser on one origin locally too, so the cookie behaviour developers
  // see is the same behaviour production has. Bugs that only appear in prod
  // because dev used a different origin are expensive to find.
  //
  // The list below is not just /api. Until S5 ports the findings view, the Go
  // dashboard is still the working one, and a link to /dashboard has to reach
  // it — otherwise Next answers with its own 404 for a page that exists. Caddy
  // needs the same set in production; keeping both lists in one place means
  // they can be compared at a glance.
  async rewrites() {
    if (process.env.NODE_ENV !== 'development') return [];
    const go = process.env.GO_API_ORIGIN ?? 'http://127.0.0.1:8080';
    return LEGACY_GO_PATHS.map((source) => ({
      source,
      destination: `${go}${source}`,
    }));
  },

  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          // The 3D scenes need no camera, mic or location. Saying so explicitly
          // means a compromised dependency cannot quietly ask for them.
          { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=()' },
        ],
      },
    ];
  },
};

export default nextConfig;
