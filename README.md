# Cloud Guard

**Find the AWS spend nobody is using.**

Cloud Guard reads an AWS account through a read-only IAM role and reports what
is costing money for no reason — priced from AWS list prices, with the evidence
and the exact command that fixes it.

Live: **https://guardinfra.duckdns.org/cloud-guard**

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)
![SQLite](https://img.shields.io/badge/sqlite-%2307405e.svg?style=for-the-badge&logo=sqlite&logoColor=white)

---

## The problem

AWS bills you for resources nobody is using, and none of it appears as a line
item saying "waste". A volume detached from a deleted instance keeps billing at
full rate. An Elastic IP is charged *precisely because* it is not attached to
anything. Snapshots outlive the machines they came from. gp2 volumes keep
costing 25% more than the gp3 that replaced them.

Cost Explorer will tell you that EC2 cost $840 last month. It will not tell you
that $61 of it was for storage attached to nothing.

## What it checks

Nine rules. **Four produce a dollar figure**; the rest flag something worth
looking at but carry no saving.

| Check | Looks for | Priced |
|---|---|:--:|
| Unattached EBS volume | State `available` — attached to nothing, billed in full | ✅ |
| Unassociated Elastic IP | AWS charges for a reserved address when it is *not* in use | ✅ |
| Stale snapshot | Older than the retention window (default 90 days) | ✅ |
| gp2 that should be gp3 | gp3 is ~20% cheaper per GiB at equal or better baseline IOPS | ✅ |
| Stopped EC2 instance | Compute is not billed, but attached storage still is | — |
| Idle EC2 instance | Running with average CPU < 5% over 7 days | — |
| Public S3 bucket | Reachable by anyone; also checks Public Access Block | — |
| Over-provisioned RDS | Instance class larger than the workload appears to need | — |
| Cost summary | Account-level spend from Cost Explorer, for context | — |

### Where the numbers come from

Every saving is computed from published AWS **us-east-1 list prices** held in
[`internal/pricing`](internal/pricing/pricing.go) — not estimated, not scaled
from a guess. An unattached 100 GiB gp2 volume is `100 × $0.10 = $10.00/month`,
and the finding shows that arithmetic alongside the resource ID and region.

If your account has committed-use discounts or a private pricing agreement, the
real figure will be lower. **Treat these as an upper bound.**

## Connecting an AWS account

The dashboard generates a CloudFormation template pre-filled with an ExternalId
that is unique to your tenant. Launch it, then paste the role ARN back.

The template creates a role whose trust policy looks like this:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "AWS": "arn:aws:iam::YOUR_HOSTING_ACCOUNT_ID:root" },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": { "sts:ExternalId": "cg-your-unique-value" }
      }
    }
  ]
}
```

> **The `Condition` block is not optional.** Without it, the trust policy says
> "the Cloud Guard account may assume me" and stops there. Role ARNs are not
> secret — anyone who learns yours could paste it into their own Cloud Guard
> account and read your AWS account through us. The ExternalId is what ties the
> role to one tenant. Use the value the dashboard gives you; do not invent one
> and do not share it.

Attached policies: `ReadOnlyAccess` and `SecurityAudit`.

Once connected, a scan enumerates **every region you have enabled** and walks
them concurrently, so a volume idling in `ap-south-1` is not missed because you
work in `us-east-1`.

## What it will not do

- **It cannot change anything.** Access is read-only by design. Every fix is a
  command you copy and run yourself, after reading it.
- **It does not detect threats.** The S3 and RDS rules are configuration checks,
  not intrusion detection. Nothing here watches traffic.
- **It is not real-time.** Scans run when you trigger them or on a schedule.
  Slack alerts follow a scan; they do not follow an AWS event.

## Running it locally

```bash
git clone https://github.com/singh-anurag-7991/cloud-guard.git
cd cloud-guard
go run ./cmd/server
```

Dashboard at [http://localhost:8080](http://localhost:8080).

With Docker:

```bash
docker build -t cloud-guard .
docker run -p 8080:8080 -v cloudguard-data:/root/data cloud-guard
```

### Configuration

| Variable | Default | Purpose |
|---|---|---|
| `SLACK_WEBHOOK_URL` | — | Post findings to Slack after a scan |
| `CLOUDGUARD_SNAPSHOT_STALE_DAYS` | `90` | Retention window for the stale-snapshot rule |
| `CLOUDGUARD_REGIONS` | all enabled | Comma-separated override, for roles without `ec2:DescribeRegions` |
| `SMTP_HOST` / `SMTP_USERNAME` / `SMTP_PASSWORD` | — | Password-reset delivery. **Unset, reset links are written to the server log instead of being sent.** |

## Testing

```bash
go build ./... && go vet ./... && go test ./...
go test ./... -bench=.
```

There are scripts under `deployments/` that create and remove real billable AWS
resources for end-to-end testing. They act only on resources tagged
`CloudGuardTest=true` — never by name, age, or heuristic. Read
`seed-test-waste.sh` before running it; it costs real money until you tear it
down.

## Architecture

- **Backend** — Go standard library plus AWS SDK v2. No web framework.
- **Database** — SQLite (`modernc.org/sqlite`, pure Go, so `CGO_ENABLED=0`
  static builds work), WAL mode.
- **Frontend** — server-rendered `html/template`. No build step.
- **Auth** — bcrypt, opaque session tokens, per-tenant isolation on every query.
- **Orchestrator** — runs scanners concurrently across regions, bounded at 5.
- **Scanners** — one module per AWS service; rules are separate from collection,
  so a rule can be tested without touching AWS.
- **Deploy** — Docker on EC2 behind Caddy, which terminates TLS and proxies to
  `127.0.0.1:8080`.

## Roadmap

- [x] Priced cost rules with evidence and a fix command
- [x] Multi-region scanning
- [x] Multi-tenant auth with per-tenant ExternalId
- [x] CSV export
- [x] Password reset
- [ ] SMTP configured on the hosted deployment
- [ ] Automated backup of the SQLite volume
- [ ] Rate limiting on the login endpoint
- [ ] Configurable thresholds in the UI (idle CPU %, snapshot age)
- [ ] Savings Plans and Reserved Instance coverage analysis

## Status

Early access, and free while it is. Connected accounts are read-only and
nothing is billed. Backups and email delivery are on the list above and not yet
done — worth knowing before you point it at anything you care about.

## License

MIT. See [LICENSE](LICENSE).
