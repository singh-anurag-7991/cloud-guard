# Guard Platform — Build Plan

**Created:** 1 August 2026
**Decisions locked:** single Go app with path-based products · Data Guard = honest "Coming Soon" for now · fix bugs and remove overclaims before building the portfolio

---

## Target architecture

One app, one login, one database. Products live behind paths, not subdomains — DuckDNS can't issue wildcard TLS certs, and separate apps would mean 2× infra for zero customers.

```
/                        Portfolio (Anurag) — bio, achievements, products box
/products                Public product catalogue
/login  /signup  /logout Shared auth
/app                     Post-login hub — shows only products this user can access
/app/cloud-guard         Cloud Guard dashboard (today's /dashboard)
/app/data-guard          Data Guard — Coming Soon + waitlist
/healthz                 Health check
```

**Entitlements:** new table `entitlements(tenant_id, product, status, granted_at)`. `status` ∈ `active | trial | none`. `/app` reads it to decide which product cards are live vs locked. Cloud Guard defaults to `trial` on signup; Data Guard to `none` with a waitlist join.

Nothing here blocks moving a product to its own subdomain or repo later — the DB and auth stay the same.

---

## Phase 1 — Make what exists actually work

**1.1 Fix the scan failure**
Two accounts, both failing AssumeRole. Diagnose from container logs, then fix. Likely causes, in order: role name not matching the `CloudGuardReadOnlyRole-*` pattern the app is permitted to assume; ExternalId mismatch; trust policy pointing at the wrong account.

**1.2 Fix the error-page data loss (confirmed bug)**
`handleScan` and `handleConnect` error paths build `dashboardData` with only `Error`, `TenantID`, `Findings` — so accounts, counts and savings render as 0 and it *looks* like data was deleted. Extract one `buildDashboardData(tenantID)` helper and use it everywhere, including error paths.

**1.3 Surface per-scanner results**
Right now a scan that half-fails looks identical to one that fully succeeded. Persist per-scanner status (`ok` / `failed` + reason) per scan, and show it on the dashboard. Cost Explorer failing should read *"Cost Explorer not enabled — enable it in Billing"*, not silence.

**1.4 Validate the role at connect time, not scan time**
Confirmed real-world failure: the app accepted `arn:aws:iam::…:role/CloudFormation-role-439d95ff` — a CloudFormation *service* role whose trust policy only allows `cloudformation.amazonaws.com` — and only failed later, during the scan, with an opaque "2 errors" message. If it confused the person who built it, it will confuse every customer.

On `POST /connect`, perform a real `sts:AssumeRole` before saving. Reject immediately with a specific reason:
- trust policy doesn't include this SaaS account → *"This role doesn't trust Cloud Guard. Use the template's trust policy."*
- ExternalId mismatch → *"ExternalId doesn't match."*
- name outside `CloudGuardReadOnlyRole-*` → *"Role name must start with CloudGuardReadOnlyRole-."*

**1.5 Cost Explorer onboarding**
Detect the `not enabled for cost explorer access` error specifically and render a fix-it card with the exact steps and the 24h data delay warning.

---

## Phase 2 — Delete the lies

Every item here is something a technical buyer can disprove in one click. For a security product that costs more than the feature earns.

**2.1 Replace hardcoded stats with real data**
`landing.html` lines ~470/472/475/485/493/737 hardcode `847 resources`, `$620/mo`, `99.9% SLA`. Replace with real platform aggregates from the DB (total scans run, total findings, total accounts connected). When the number is zero, say so or omit the stat — do not invent one.

**2.2 Fix the "1-click CloudFormation" claim**
Either host the template in a public S3 bucket and set `CF_TEMPLATE_S3_URL` (making it genuinely 1-click), or change the wording everywhere to match the real 2-step flow.

**2.3 Mark Data Guard and Shield honestly**
Both currently read as shipping products. Data Guard → "In development", Shield → link to the actual repo or drop it. No pricing tiers on things that don't exist.

**2.4 Remove the unenforced trial claim**
"14-day free trial" appears on the pricing page but nothing in code enforces it. Either implement it in the entitlements table or remove the words.

---

## Phase 3 — Portfolio homepage

**3.1 `/` becomes the portfolio**
Bio, experience, skills, achievements, contact. Cinematic feel — scroll-reveal, subtle motion — matching the existing dark theme so it doesn't look bolted on. Current landing page content moves to `/products`.

**3.2 Products box on the portfolio**
Card grid: name, one-line description, real status badge (`Live` / `In development` / `Open source`), link. This is the bridge from "who is Anurag" to "what he ships".

**3.3 Real visitor counter**
New table `page_views(path, day, count)`. Middleware increments on GET of public HTML pages only — excluding `/healthz`, static assets, and authenticated app routes. Bots filtered by user-agent. Display total on the portfolio. This must be a genuine count; a fake counter is exactly the kind of thing this phase exists to remove.

**3.4 `/app` product hub**
Post-login page listing products with entitlement state. Accessible ones link through; locked ones show why and offer the waitlist.

---

## Phase 4 — Data Guard placeholder

**4.1 `/app/data-guard` — Coming Soon**
Honest page: what it will do (GCP — BigQuery cost, GCS permissions, IAM over-privilege), current status, waitlist email capture into a `waitlist` table.

**4.2 Public product page**
`/products/data-guard` with the same honesty. No pricing until it exists.

Full Data Guard is a second product — GCP auth, three scanners, its own onboarding. That is weeks of work and is deliberately out of scope here.

---

## Deferred (tracked, not now)

- Stripe live keys + real payment walkthrough
- Trial enforcement in code
- Monitoring/alerting on the app itself
- `DISABLE_SIGNUP=1` once real customers exist
- Delete test artifacts: doubled-name test role, `e2e-*` / `iso-*` accounts
- Regenerate the DuckDNS token

---

## Sequencing note

`coach/PLAN.md` still has job search as priority #1 on a weeks-level timeline, with Cloud Guard on weekend-only maintenance. This plan is large. Phases 1 and 2 are small and high-value — they make the product true. Phase 3 (portfolio) arguably *serves* the job search directly. Phase 4 is deliberately minimal.

If time is short, Phase 1 + 2.1 + 2.3 alone gets the product to "honest and working", which is the state worth pausing in.
