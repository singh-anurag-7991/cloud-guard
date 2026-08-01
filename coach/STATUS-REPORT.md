# Cloud Guard — Production Status Report

**Date:** 1 August 2026
**Live URL:** https://guardinfra.duckdns.org
**Instance:** `i-048a0f8e1ff5809b3` (t3.small, us-east-1, Elastic IP `44.216.212.91`)
**Verdict:** Live and functional. 24/24 end-to-end tests passed. Three items remain before selling to customers.

---

## 1. What works right now (verified, not assumed)

Every line below was tested against the live server, not inferred from code.

### Authentication & multi-tenancy
| Test | Result |
|---|---|
| `/dashboard` when logged out | Redirects to `/login` |
| `/api/findings`, `/api/accounts` when logged out | `401` |
| Signup: password under 10 chars | Rejected |
| Signup: password mismatch | Rejected |
| Signup: duplicate email | Rejected ("already registered") |
| Signup: valid | Account created, redirected to login |
| Login: wrong password | Rejected (vague message — doesn't leak whether email exists) |
| Login: correct password | Session cookie set, redirected to dashboard |
| Dashboard after login | `200`, renders app UI |
| Logout | Session destroyed, protected routes blocked again |
| **Second user sees zero data from first user** | **Confirmed — separate tenant IDs** |

Passwords are bcrypt-hashed. Sessions are opaque random tokens in an HttpOnly, Secure, SameSite=Lax cookie with a 7-day expiry.

### AWS onboarding & scanning
| Test | Result |
|---|---|
| CloudFormation stack ARN pasted by mistake | Rejected with the exact fix: *"open your stack's Outputs tab and copy the RoleARN"* |
| Malformed ARN | Rejected, not saved |
| Valid role ARN | Connected |
| **Scan via `sts:AssumeRole`** | **Completed in 3.8s — no AccessDenied** |
| EC2 scanner | Ran clean |
| S3 scanner | Ran clean |
| RDS scanner | Ran clean |
| Cost Explorer scanner | Failed — see Known Issues |
| Scan with no accounts connected | Returns a clear error instead of a fake success |
| Background hourly scheduler | Confirmed running on its own |

"0 findings" is the correct result — this AWS account has no stopped EC2 instances, public S3 buckets, or oversized RDS databases to flag.

### Infrastructure
- HTTPS with a Let's Encrypt certificate, auto-renewing via Caddy (no cron, no certbot)
- Elastic IP — the address no longer changes on restart
- App bound to `127.0.0.1:8080` behind Caddy, not directly exposed
- Security group: 22, 80, 443
- `CloudGuardAppRole` attached to the instance, with `sts:AssumeRole` scoped to `CloudGuardReadOnlyRole-*` only
- Build time reduced from ~10–12 min to ~1–2 min

---

## 2. Bugs found and fixed this session

Eleven issues. Three would have been serious in front of a paying customer.

### Critical

**1. There was no authentication at all.**
With `CLERK_SECRET_KEY` unset, every request was silently granted tenant `"default"`. The dashboard, `/connect`, `/scan` and every `/api/*` route were reachable by anyone on the internet — including the ability to register AWS role ARNs and trigger scans. Verified by fetching `/dashboard` with no credentials and receiving `200`.
*Fixed:* full signup/login with bcrypt + server-side sessions.

**2. Multi-tenancy was fake.**
`parseTokenClaims` never verified the JWT signature and always returned `"default"`. Any random string worked as a token, and all users shared one tenant — so any customer would have seen every other customer's findings.
*Fixed:* each signup gets its own isolated `tenant_id`, verified by test.

**3. The database was not persisting.**
`DB_PATH` defaulted to `cloudguard.db` (→ `/root/cloudguard.db`), but the Docker volume mounts at `/root/data`. **Every redeploy silently wiped all accounts and findings.**
*Fixed:* DB now lives inside the volume; directory auto-created. Also added WAL + `busy_timeout`, which removes a latent "database is locked" risk between the scheduler and web handlers.

### High

**4. `sts:ExternalId` was never sent.** The CloudFormation trust policy requires it, but the code omitted it — so *every* customer's AssumeRole would have failed with AccessDenied. Nothing would ever have scanned.

**5. Placeholder AWS account ID.** The template shipped `123456789012`, meaning generated roles trusted a non-existent account.

**6. `/cloudformation.yaml` returned 404 in production.** The Dockerfile never copied `deployments/`, so the "Download Template" button was dead.

**7. Scan failures were invisible.** If all four scanners failed, the app still reported a successful scan with 0 findings. Now it surfaces a real error.

**8. Dead footer links.** `/docs`, `/blog`, `/contact` all 404'd, plus five `href="#"` placeholders. Now every link resolves.

### Medium

**9. Go build OOM-killed the server.** Compiling `aws-sdk-go-v2/ec2` peaked ~1.5 GB with zero swap; the OOM killer took `dnf` instead of the compiler, so the build appeared to hang forever rather than fail. Replaced `mattn/go-sqlite3` with pure-Go `modernc.org/sqlite` — CGO gone, build 10× faster.

**10. `go build cmd/server/main.go`** — missing `./` made Go treat the path as a stdlib import.

**11. `.gitignore` was excluding source code.** An unanchored `server` pattern silently excluded `cmd/server/` and `internal/server/` from git — core application code was never pushed.

---

## 3. Known issues — must be resolved before selling

### A. Cost Explorer is not enabled (blocks a headline feature)
```
AccessDeniedException: User not enabled for cost explorer access
```
Cost scanning is advertised on your pricing page and is central to the "find wasted spend" pitch, but it currently returns nothing. Each customer must enable Cost Explorer once in **Billing → Cost Explorer**, and it takes up to 24h to populate. This needs to be in your onboarding instructions, or the feature looks broken to every new customer.

### B. Stripe is still on test keys
Nobody can actually pay you. The pricing page advertises $99/$199/$299 tiers with a "14-day free trial" that isn't enforced anywhere in code.

### C. The landing page overstates what exists
- **"1-click CloudFormation"** — AWS only accepts an S3-hosted `templateURL`, so it is currently a 2-step flow (download, then upload). Either host the template in S3 and set `CF_TEMPLATE_S3_URL`, or change the wording.
- **Homepage stats** (`847 resources`, `$620/mo saved`, `99.9% uptime SLA`) are hardcoded placeholders presented as real results.
- **Data Guard / Shield** are marketed as products; neither has code.

For a security product, an overclaim that a technical buyer can disprove in one click costs more trust than the feature would have won.

---

## 4. Smaller cleanups

- Signup is publicly open — set `DISABLE_SIGNUP=1` once you have customers
- Regenerate the DuckDNS token (it was pasted in plaintext during setup)
- Delete the test role `CloudGuardReadOnlyRole-us-east-1CloudGuardReadOnlyRole-us-east-1` (name doubled from a console input quirk)
- Delete test accounts `e2e-*@test.local` and `iso-*@test.local`
- `git config --global user.email` is unset — commits are attributed to `anuragsingh@MacBookAir.lan`
- No monitoring or alerting on the app itself — if it crashes at 3am, nothing tells you

---

## 5. Recommended order

1. **Enable Cost Explorer** and document it in onboarding — it's a headline feature that silently fails today
2. **Fix or soften the landing-page claims** — cheapest credibility win available
3. **Stripe live keys + walk the full payment flow yourself**
4. **Then** marketing and outreach

A note on sequencing: per `coach/PLAN.md`, job search is still priority #1 with a weeks-level timeline, and Cloud Guard is meant to be weekend-only maintenance. This session was a large deviation from that. The product is now in a genuinely sellable state, which is a good place to pause it — resist the pull to keep polishing at the expense of the job search.
