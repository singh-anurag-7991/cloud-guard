# Anurag's Coach Plan

Owner: Anurag Singh · Coach: Claude · Started: 2026-08-01 · Last major revision: 2026-08-01

## ⚠️ Current #1 priority: JOB SEARCH (urgent, weeks-level timeline)

Current company situation is bad — job search is now the top priority above everything else, including Cloud Guard. Target: broader SDE role (any stack), not limited to backend/Go. This overrides the original plan below where they conflict.

**Everything reprioritizes around this until an offer is signed.**

## Mission (in priority order)

1. **Land a new SDE job** — weeks-level urgency.
2. Stay healthy enough to think straight and interview well (exercise, non-negotiable).
3. Keep Cloud Guard alive on low heat — don't abandon it, don't feed it time that should go to #1.

Longer-term (resumes once job search is settled): Guard Platform → 2-5 paying Cloud Guard clients/month, $1000+ MRR, then Data Guard / Shield.

## Repo snapshot (assessed 2026-08-01) — for later reference, not today's focus

Cloud Guard (Go, SQLite, AWS SDK v2) is further along than "just an MVP":
- Landing page, `/products`, `/about` pages live in `web/`
- Stripe billing wired up: Starter $99, Pro $199, Business $299/mo (`internal/billing/stripe.go`)
- Auth middleware, protected dashboard, connect/scan flows
- Scanners: EC2 (stopped/idle), S3 (public buckets), RDS (over-provisioned), Cost Explorer
- Slack alerts, 1-click CloudFormation onboarding (no static AWS keys)
- `products.html` also markets **Data Guard** (GCP) and **Shield** (rate limiter/DDoS) — aspirational/roadmap only, no code yet. Don't touch these at all right now.
- Only 1 git commit so far — no evidence of production deployment or a real customer conversation yet.

The product is basically sellable — but that's a project for after the job search stabilizes. Right now it just needs to not rot: no new features, just don't let it go completely cold (see weekend-only maintenance below).

## ⚠️ Update (2026-08-02): interviews are flowing, but not converting

Pipeline is working — daily interview calls are coming in (20+ interviews given so far, Go-focused). Zero offers yet, and no clear feedback on why. Without signal on which round is failing, prep broadens to cover the most likely gaps for a Go SDE interview: Go language depth (not just DSA), plus light system design and behavioral. Application volume can ease off slightly since the funnel is clearly working — prep time increases instead.

## Daily time budget (2-3 hrs+ total) — reprioritized for interview conversion

| Block | Time | What |
|---|---|---|
| Exercise | 15-20 min | Quick — walk/run/bodyweight. Keeps stress and sleep in check during a high-pressure search. Do it first. |
| Interview Prep (merged) | 60-75 min | Rotates daily between DSA and Go-language deep dives — see rotation below. This is now the highest-leverage block. |
| Job search maintenance | 20-30 min | Lighter than before — pipeline is already flowing. A few applications/day to keep it topped up, respond to recruiters, schedule interviews, log outcomes. |
| Cloud Guard | 0-15 min, weekdays optional | On hold. Only touch it on weekends, and only maintenance (see below), not new features. |

Order: exercise → Interview Prep → job search maintenance is the default. Interview Prep wins over everything except exercise if the day gets squeezed.

### 👉 Interview Prep ab `coach/INTERVIEW-PREP.md` follow karega

Weekly rotation replaced by a structured **14-day aggressive roadmap** — see `coach/INTERVIEW-PREP.md`. Context: 5 years experience, SDE-3 level, interviews already running, 2-week timeline.

Day-by-day: Go concurrency (D1-D2) → Go internals/gotchas (D3) → SQL depth (D4) → NoSQL/caching/scaling (D5) → REST API design (D6) → review + mock (D7) → message queues (D8) → microservices (D9) → Docker (D10) → Kubernetes (D11) → deployment strategies/CI-CD (D12) → **system design (D13, highest weight)** → full mock + behavioral (D14).

Daily format: ~20 min DSA + ~45 min core topic. DSA is deliberately the smaller slice — at 5 years, system design and Go depth are where offers are won and lost.

Ask me anytime for a live mock interview (system design, Go concept grilling, DSA, or behavioral) right here in chat — reps matter more than reading.

## Golang interview roadmap — basic to advanced

Go deep-dive topics for the Tue/Thu slots. Order roughly basic → advanced; skip ahead if something's already solid, but don't skip concurrency — it's the single most common Go interview differentiator.

1. **Fundamentals:** value vs pointer receivers, structs & embedding, slices vs arrays (len/cap, append/aliasing gotchas), maps internals, interfaces & implicit implementation, zero values
2. **Error handling idioms:** `error` interface, wrapping (`errors.Is`/`errors.As`/`%w`), custom error types, panic/recover/defer semantics and when (not) to use them
3. **Concurrency (the big one):** goroutines, unbuffered vs buffered channels, `select`, `sync.Mutex`/`RWMutex`/`WaitGroup`/`Once`, the `context` package (cancellation, timeouts, propagation), common patterns (worker pool, fan-in/fan-out, pipeline), goroutine leaks, deadlocks, race conditions (`go run -race`)
4. **Memory & runtime internals:** stack vs heap and escape analysis, garbage collector basics (concurrent mark-and-sweep, GC tuning knobs), how slices/maps are implemented under the hood
5. **Standard library & tooling:** `net/http` basics, `encoding/json` gotchas, `testing` package (table-driven tests, benchmarks, `t.Parallel`), Go modules, basic `pprof` profiling
6. **Idioms & design:** composition over inheritance, small interfaces ("accept interfaces, return structs"), functional options pattern, dependency injection Go-style
7. **Classic gotcha questions** (these come up a lot — know the "why", not just the fix): nil interface vs nil pointer, loop-variable capture in closures/goroutines (pre-Go 1.22 behavior vs 1.22+), slice append aliasing bugs, channel deadlocks, why `defer` in a loop can bite you

For each topic: be able to explain it out loud in 60-90 seconds AND write a 10-15 line code snippet demonstrating it from memory. That combination is what actually gets tested.

## Job search playbook (the main event right now)

1. **Day 1-2: Foundation.** Rewrite resume for a broad SDE target (impact-focused bullets, metrics, no company-specific jargon). Update LinkedIn headline + "Open to Work" (recruiters only, unless situation is bad enough to go public). Turn on LinkedIn easy-apply alerts for target roles.
2. **Daily, from day 2 onward:**
   - Apply to 5-10 relevant SDE roles/day (quality-tailored over pure volume, but volume matters at "weeks" urgency).
   - 2-3 direct outreach messages/day to people at target companies (alumni, ex-colleagues, recruiters) asking for a referral or 15 min chat.
   - Track every application (company, role, date, status) — ask me to set up a tracker if it gets hard to hold in your head.
3. **Interview prep, daily:** DSA block above doubles as this. Add light behavioral prep (2-3 STAR stories about real projects — Cloud Guard itself is a strong talking point: "built and shipped a full SaaS with billing, auth, AWS integration solo").
4. **Weekly:** review response rate (applications → replies → interviews). If reply rate is near zero after ~50 applications, the resume/targeting needs a rework, not more volume — flag this for a plan adjustment.
5. **Do not neglect the current job** while searching — keep performing there; a bad reference or being let go before you have an offer makes everything harder.

## DSA roadmap — beginner→interview track (LeetCode, Easy→Medium)

Now serving interview prep directly, so pace is faster than the original "someday" version. 1-2 problems/day, write the pattern used in one line after solving.

1. **Arrays & Strings** — two-sum patterns, prefix sums, string manipulation
2. **Hashing** — hashmap/set for frequency, duplicates, lookups
3. **Two Pointers & Sliding Window**
4. **Sorting & Binary Search**
5. **Recursion & Basic Backtracking**
6. **Linked Lists**
7. **Stacks & Queues**
8. **Trees (BFS/DFS, basic traversal)**
9. Start Medium difficulty, revisit weak patterns
10. Light System Design basics (broad SDE interviews often touch this even for mid-level) — scaling, caching, basic API design, once patterns above feel solid.

## Cloud Guard maintenance mode (paused, weekends only)

While job search is #1: no new features, no outreach push. Just don't let it fully die:
- Weekends only, 30-60 min if there's spare time (not required).
- If touched at all: fix anything broken, or nothing. Do not start Data Guard/Shield.
- Full backlog (deploy live, funnel test, outreach, etc.) is parked in the "Cloud Guard backlog (parked)" section below — resume once the job search is done.

### Cloud Guard backlog (parked — resume after job secured)

1. Ship it to a real URL — production deploy, domain, HTTPS, live Stripe keys/webhook.
2. Walk the full funnel yourself end-to-end, fix what's broken.
3. Replace placeholder landing-page numbers with a real scan's findings.
4. Narrow target customer (indie hackers / seed startups / small agencies on AWS).
5. Daily outreach loop once resumed (Reddit, Indie Hackers, X, direct DMs).
6. Build in public.
7. First 5 users free-trial, high-touch.

## Weekly rhythm

- Mon-Fri: exercise + DSA + job search execution, every day, no exceptions.
- Weekend: same, plus optional light Cloud Guard maintenance, plus a **weekly job-search review** (applications sent, replies, interviews landed, what to adjust).

## How the daily check-ins work

- **9:00 AM** — morning plan: today's DSA problem, today's specific job-search actions (applications target, outreach target, or prep task), exercise reminder. Cloud Guard only mentioned on weekends.
- **10:00 PM** — evening check-in: what got done (exercise / DSA / job search actions — apps sent, replies, interviews / Cloud Guard if weekend), logged to `coach/LOG.md` honestly.

## Notes / decisions log

- 2026-08-01: Plan created, initial focus = Cloud Guard as primary business track.
- 2026-08-01 (same day, revised): Job search declared #1 priority, urgent (weeks timeline), target = broader SDE role. Cloud Guard moved to weekend-only maintenance mode. Daily budget reallocated: exercise 15-20min, DSA 45-60min (interview-focused), job search execution 45-60min, Cloud Guard 0-15min weekdays (optional, none required).
- 2026-08-02 (later): Experience confirmed = 5 years (SDE-3 level), timeline = 2 weeks aggressive. Full `coach/INTERVIEW-PREP.md` created: 14-day roadmap covering Go concurrency/internals, SQL + NoSQL + caching + scaling, REST API design, message queues, microservices + Saga/resilience, Docker, Kubernetes, deployment strategies + zero-downtime migrations, and system design (highest weight, with a framework). DSA reduced to ~20min/day since it's the lower-return area at this level. Behavioral/STAR stories added, with Cloud Guard as the flagship project story.
- 2026-08-02: 20+ interviews given (Go-focused), zero offers, no clear rejection feedback. Interview Prep block merged and expanded to 60-75min: Mon/Wed/Fri = DSA, Tue/Thu = Go language deep-dive (roadmap added, basic→advanced, concurrency emphasized), Sat = light system design, Sun = weekly review + behavioral. Job search execution eased to maintenance-level (20-30min) since the pipeline is clearly working. Coach communicates in Hinglish from this point per Anurag's request.
- 2026-08-01 (Cowork session): Anurag asked to resume full Cloud Guard push (deploy, marketing, cold outreach) "even if I'm not around." Confirmed with him: **job search stays #1 priority** — this does NOT override the plan above, Cloud Guard stays weekend-only maintenance, no daily time reallocation. For when weekend slots are used on Cloud Guard, he confirmed the approach below (not to be rushed into weekday time).
  - **Deploy**: via AWS account (has $200 credits) + Chrome browser with the Claude-in-Chrome extension (not Safari — browser automation tools can't click/type in Safari for safety reasons, view-only). Chrome extension not connected yet as of this session — Anurag needs to install Chrome, log into AWS there, and connect the extension before deploy work can proceed.
  - **Outreach (cold DMs to US/Canada businesses on LinkedIn/X/Instagram, selling Cloud Guard + small-website-building services)**: low-volume only, and Claude asks for approval on each message before sending — not bulk/automated, to avoid ToS violations and account bans.
  - **Marketing (X posts, social groups)** and **portfolio website improvements**: parked behind deploy — sequence is deploy → portfolio → marketing → outreach.
  - No work started yet this session beyond this plan update — blocked on Chrome extension connection. Next weekend session: connect Chrome, then execute deploy.
- 2026-08-01 (same session, later): Chrome connected, deploy attempted. Progress: IAM deploy-user `cloudguard-deploy` created (EC2FullAccess), access key generated, all uncommitted code (auth/billing/scheduler/deploy configs) committed and pushed to GitHub (`singh-anurag-7991/cloud-guard`, public). Blockers hit: (1) AWS CloudShell unavailable — new account under AWS's own "account verification" hold, up to 2 days, unrelated to anything we did, resolves on its own; EC2 itself works fine, only CloudShell is blocked. (2) Sandbox environment has no general internet access (allowlist-only), so AWS CLI/GitHub can't be reached directly from there — browser (Chrome) is the only channel. (3) Hit a scroll bug in the EC2 "Launch Instance" console wizard via browser automation — page wouldn't scroll to reach Key pair/Network/User-data sections, so couldn't finish the click-through launch reliably. Resolution: wrote `deployments/deploy-ec2.sh`, a ready-to-run script covering security group + EC2 launch + user-data (installs Docker, clones the repo, builds, runs on port 80). Next step: run it in AWS CloudShell once the account verification clears (check back in ~1-2 days), or Anurag can run it himself now if he has AWS CLI configured locally. No domain yet (deliberately deferred) — will get a raw IP URL first.
- 2026-08-01 (same session, autonomous debug pass while Anurag was away): Anurag ran `deploy-ec2.sh` locally multiple times via his own AWS CLI, pasting console output/errors back for diagnosis. Found and fixed 3 real bugs, each verified via a fresh terminate+relaunch cycle: (1) `Dockerfile`/`README.md` used `go build cmd/server/main.go` (missing `./` prefix) which Go misresolves as a stdlib import path — fixed to `./cmd/server`. (2) `.gitignore` had an unanchored `server` pattern that was silently excluding both `cmd/server/` and `internal/server/` from git entirely — core app code was never actually pushed to GitHub despite being present locally. Fixed by anchoring `.gitignore` patterns with a leading `/` and explicitly re-adding both directories. Confirmed via `aws ec2 get-console-output` showing the exact Docker build failures each time, one bug at a time. After both fixes, the Docker build finally proceeded past the code stage into the CGO/gcc sqlite3 compile step for the first time.
  - Once code was fixed, the running instance (t3.micro, ~912MB usable RAM per boot log) was still unreachable — suspected OOM during the CGO build. Terminated it and relaunched on `t3.small` (2GB RAM) with the existing `cloud-guard-sg` rules replicated (new SG `launch-wizard-1`, ports 22+80 open to 0.0.0.0/0), no key pair (EC2 Instance Connect only), and an improved user-data script that logs to `/var/log/user-data.log` with `set -x` for future debugging. Updated `deployments/deploy-ec2.sh` to use `t3.small` and the same logging improvement (not yet committed/pushed — Anurag needs to `git add/commit/push` this next time he's at the terminal, since Claude's sandbox has no GitHub network access).
  - New instance `i-0aa8b8c830e029375` (t3.small, IP `100.54.110.132`) launched and reached "Running, 3/3 status checks passed" within ~2 min, but the app was still not reachable on port 80 after 25+ minutes of monitoring. Ruled out network-level causes definitively: VPC route table has a correct `0.0.0.0/0 → igw` route (Active), the default Network ACL allows all traffic both ways, and the new security group has both port 22 and 80 open — confirmed via direct navigation to `http://100.54.110.132/` in Chrome, which returned `ERR_CONNECTION_REFUSED` (not a timeout), proving packets ARE reaching the box and it IS alive; nothing is listening on port 80 yet, i.e. this is a slow-build issue, not a network/security issue. The EC2 "Get system log" troubleshooting panel confirmed `dockerd`, `docker-buildx`, and the Go `compile` process were all actively running (not OOM-killed, no crash), so the build was genuinely progressing, just slower than expected (possibly new-account bandwidth throttling on package downloads, even though the account's own "customer verification" AWS Health Event shows SUCCESS from 3 days ago). EC2 Instance Connect (browser-based SSH) was tried ~6 times throughout and consistently either hung on "Establishing Connection..." or failed with a generic "Error establishing SSH connection" — this looks like a flaky/unreliable AWS console feature on this account rather than an instance-health signal (it failed identically on this healthy t3.small as it did on earlier broken instances), so it wasn't useful for deeper diagnosis. IAM user `cloudguard-deploy` has no SSM permissions and the instance has no IAM instance profile attached, so SSM Session Manager wasn't an option either.
  - **RESOLVED — Cloud Guard is LIVE at http://34.229.120.193** (instance `i-048a0f8e1ff5809b3`, t3.small, us-east-1). Landing page, `/products` and nav all confirmed rendering.
  - **Actual root cause of the "unreachable" mystery: OOM during the Go build.** `aws ec2 get-console-output` on the stuck box showed `oom-kill:constraint=CONSTRAINT_NONE ... global_oom`, `Total swap = 0kB`, and two concurrent Go `compile` processes at ~892MB and ~651MB RSS — ~1.5GB of a 1.9GB box. The huge `aws-sdk-go-v2/service/ec2 v1.279.2` package is the memory hog. Critically, the OOM killer killed `dnf` rather than the compiler (compile runs at `oom_score_adj -500`, so it's protected), which is why the build never errored — it just silently stalled forever and the box stayed "Running / 3-3 checks passed" while nothing ever bound to port 80. Browser `ERR_CONNECTION_REFUSED` (not a timeout) was the key clue that the network path was fine all along and only the app was missing.
  - **Fixes applied (all committed + pushed):** (1) 4GB swapfile created in user-data *before* the build — instances ship with zero swap; this alone is what let the OOM fire. (2) `go build -p 1` in the Dockerfile to serialize package compilation and halve peak memory. (3) Root volume bumped 8GB → 20GB so the swapfile + Docker layers fit. (4) Key pair now auto-created at `~/.ssh/cloud-guard-key.pem` and attached at launch — the previous relaunch used "proceed without a key pair", which left EC2 Instance Connect (unreliable on this account, failed ~6/6 times) as the only way in, i.e. no way to inspect a stuck box. Never launch without a key again.
  - Verified live via SSH during the successful build: `Swap: 4.0Gi, 2.0Gi used` — 2GB of swap actively in use, exactly the memory that was previously being OOM-killed. Build took ~10-12 min (slower due to `-p 1` + swapping, but it *completes*, which is the point).
- 2026-08-01 (same session, HTTPS + stable hostname): Site was confirmed publicly reachable (verified by fetching it from an external network), yet Anurag reported it unreachable from every device except his Mac — including devices on the same WiFi, and even when typing `http://` explicitly. Two wrong theories ruled out along the way: carrier/ISP filtering of raw-IP HTTP (killed by the same-WiFi failures) and a crash-looping container (killed by repeated clean external fetches of `/` and `/about`). Surviving explanation: **device-level HTTPS-Only mode** (Safari "Always Use Secure Connections", Chrome/Firefox HTTPS-Only), which force-upgrades `http://` URLs even when the scheme is typed explicitly — device-level, so it fails on every device regardless of network, while the Mac worked because navigation was done programmatically rather than via the address bar.
  - Fix shipped: **https://guardinfra.duckdns.org** — Elastic IP `44.216.212.91` (permanent, replaces the ephemeral IP that had already changed once), port 443 opened, DuckDNS free subdomain, and Caddy in Docker (`--network host`) reverse-proxying to the app on `127.0.0.1:8080` with automatic Let's Encrypt certs and auto-renewal (no certbot/cron). Certificate issued successfully via tls-alpn-01. Verified from an external network over HTTPS; app auto-rewrites all internal links to the new hostname. Confirmed exactly one Elastic IP exists and it is associated (an idle one bills ~$3.60/mo).
  - Rejected sslip.io/nip.io after checking: they are NOT on the Public Suffix List, so Let's Encrypt rate limits are shared across all users of those domains and the quota is currently exhausted — certs would likely have failed. DuckDNS *is* on the PSL, hence reliable.
  - Note: the DuckDNS token was pasted in plaintext during setup — regenerate it at duckdns.org when convenient. The raw-IP URL no longer serves the app by design; everything is on the hostname now.
  - **Known follow-up (deliberately parked, not urgent):** building Go+CGO on a small box is the fragile part of this setup. Better long-term: either swap `mattn/go-sqlite3` → `modernc.org/sqlite` (pure Go, drops CGO/gcc entirely) or build the image on the Mac and push to a registry so the instance only pulls. Either would make deploys fast and boring and allow a smaller/free-tier instance. Also still pending: domain + HTTPS (currently raw IP, http only), and live Stripe keys.
