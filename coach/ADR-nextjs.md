# ADR: Next.js front end for Guard Infra + personal site

Status: accepted, 3 Aug 2026
Supersedes: the Go `html/template` marketing pages in `web/`

## Context

The Go app works: bcrypt auth, opaque sessions, tenant isolation, multi-region
AWS scanning, nine priced rules, findings with evidence and confidence, CSV
export. That is the product and it must survive this change untouched.

What does not work is the presentation. Two identities are sharing one
stylesheet, so Guard Infra and the personal site read as one company with two
skins. The brief asks for a real separation, five themed 3D scenes, tabbed
product pages with documentation as the landing tab, and an MDX docs system.
None of that fits comfortably in `html/template`.

## Decision

Next.js 15 (App Router) + TypeScript + Tailwind + React Three Fiber + MDX for
the entire front end. The Go binary becomes a JSON API and keeps ownership of
auth, scanning and data.

### 1. Host Next.js on the existing EC2 behind Caddy — not Vercel

The brief specifies Vercel. Rejecting that, deliberately.

Vercel would put the front end on `*.vercel.app` (or a subdomain) and the Go
API on `guardinfra.duckdns.org`. That is cross-origin, which forces
`SameSite=None; Secure` cookies, a CORS allowlist, and a preflight on every
authenticated request. Cross-origin cookie handling is the single most common
way a working auth flow breaks — Safari's ITP and Chrome's third-party cookie
changes both target exactly this pattern.

Serving both from one origin through Caddy makes the cookie problem disappear
entirely:

```
guardinfra.duckdns.org
  /api/*      → Go      127.0.0.1:8080
  /*          → Next.js 127.0.0.1:3000
```

The goal here is a demo that is always ready. One origin, one cookie, no CORS
is fewer moving parts than a CDN we do not need for this traffic.

Cost of this choice: no Vercel preview deploys, no edge CDN. Accepted. The box
is already paid for and Caddy already terminates TLS.

Build risk is handled the same way as the Go image — built on the Mac, pushed
to Docker Hub, pulled by the box. The t3.small never compiles anything. Next
runs in `output: 'standalone'` mode, roughly 120 MB resident, which fits
alongside Go and Caddy in 2 GB.

### 2. Go stays the auth authority

Rejected Supabase and a standalone Auth.js database.

There is already a `users` table with bcrypt hashes, a `sessions` table with
opaque tokens, and per-user `tenant_id` isolation that every query depends on.
Introducing a second identity store would mean either migrating live accounts
or running two sources of truth. Both are worse than adding JSON endpoints to
code that already works.

Next.js calls `/api/auth/*` on the Go app. Go sets the httpOnly cookie exactly
as it does today. Google OAuth is added *in Go*, creating the same user and
tenant records as email signup, so there is one code path for identity.

### 3. Data Guard is "Coming soon", not "Beta"

The brief asks for a Beta badge. There is no Data Guard code. A Beta badge on
nothing is precisely the claim that would undermine the positioning the rest of
the site is built on — that every number is traceable and nothing is marked
available before it works. Overruled on purpose; flagged to Anurag.

### 4. Cutover is verified before the old pages are retired

The Go HTML templates stay in the repo and keep serving until the Next.js site
is live and checked. Caddy is the switch. Rollback is one line in the Caddyfile,
not a redeploy.

### 5. No AWS or GCP trademark artwork in the 3D scenes

Legal requirement from the brief. Each platform is evoked through authored
geometry, palette and camera behaviour only.

## Consequences

- Two runtimes to deploy instead of one. Mitigated by both being containers
  behind the same Caddy config and one pull-deploy script.
- The dashboard is rebuilt as a new view. The scanning, rules and pricing
  packages are not touched — this is a different client over the same API.
- Node is now a build dependency. Only on the Mac; the server pulls images.
- If Next.js turns out to be too heavy for the t3.small in practice, the
  fallback is Vercel with the cross-origin cookie work done properly. Measure
  before assuming.

## Layout

```
cloud-guard/
  cmd/ internal/ deployments/   unchanged Go application
  web/                          legacy Go templates, retired after cutover
  site/                         Next.js application
    app/(personal)/             anurag — universe identity
    app/(guard)/                guard infra — perimeter identity
    content/docs/               MDX documentation
```

The two route groups carry separate token layers so a colour or typeface from
one identity can never reach the other.
