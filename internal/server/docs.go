package server

// Product documentation shown at /hub/{slug} after signing in.
//
// Everything here is written from the products' own repositories, not from
// memory. An earlier version of the portfolio described Data Guard as "in
// design" for weeks while the repository was finished and had a green CI badge;
// the cost of guessing is that the one place a visitor checks your honesty is
// the place you were careless.

// DocSection is one block of a product's documentation page.
type DocSection struct {
	Heading string
	Body    string
	Bullets []string
}

// DocRule is one check Cloud Guard runs. Priced separates the rules that
// produce a dollar figure from the ones that only produce a risk flag — a
// distinction worth making plainly, because a customer reading "nine checks"
// will otherwise assume all nine save money.
type DocRule struct {
	Name   string
	What   string
	Priced bool
}

// ProductDoc is the full page for one product.
type ProductDoc struct {
	Slug     string
	Name     string
	Tagline  string
	Icon     string
	Summary  string
	Stack    string
	Sections []DocSection
	Rules    []DocRule

	// The single action at the end of the page. Cloud Guard is a running
	// service, so it opens; the other two are repositories, so they link out.
	ActionLabel    string
	ActionURL      string
	ActionExternal bool

	// Caveat is printed in a muted box above the action when there is something
	// a reader deserves to know before clicking.
	Caveat string
}

func productDocs() []ProductDoc {
	return []ProductDoc{
		{
			Slug:    "cloud-guard",
			Name:    "Cloud Guard",
			Tagline: "Find the AWS spend you forgot about",
			Icon:    "🛡️",
			Stack:   "Go · SQLite · AWS SDK v2 · Docker · Caddy",
			Summary: "AWS bills you for resources nobody is using — a volume detached from a deleted instance, an Elastic IP reserved and never attached, snapshots from a machine that no longer exists. None of it appears as a line item saying \"waste\". Cloud Guard reads your account through a read-only role and reports what is costing money for no reason, priced from AWS list prices, with the evidence and the command that fixes it.",
			Sections: []DocSection{
				{
					Heading: "How it works",
					Body:    "You create a read-only IAM role in your own account and give Cloud Guard permission to assume it. Nothing is installed in your account and no long-term credentials ever change hands.",
					Bullets: []string{
						"You launch a CloudFormation template that creates the role. It grants read-only access and nothing else.",
						"The role's trust policy names Cloud Guard's account and a per-tenant ExternalId that is unique to you, so nobody else can point Cloud Guard at your role.",
						"On each scan, Cloud Guard calls sts:AssumeRole for temporary credentials that expire on their own.",
						"It lists every region you have enabled and scans them concurrently, so a resource sitting in ap-south-1 is not missed because you work in us-east-1.",
						"Each resource is run through the rule engine. Anything that matches becomes a finding with a region, a resource ID, and where a price exists, a monthly figure.",
					},
				},
				{
					Heading: "Where the money figures come from",
					Body:    "Every saving is computed from published AWS us-east-1 list prices held in the code, not estimated or scaled from a guess. An unattached 100 GiB gp2 volume is 100 × $0.10 = $10.00 a month, and the finding shows you that arithmetic. If your account has a private pricing agreement the real number will be lower, so treat these as an upper bound.",
				},
				{
					Heading: "Getting started",
					Bullets: []string{
						"Open the dashboard and choose Connect account.",
						"Launch the CloudFormation template it gives you. It is pre-filled with your ExternalId — do not share that value.",
						"Paste the role ARN back into the form.",
						"Run a scan. The first one walks every enabled region, so give it a minute.",
						"Work down the findings, highest saving first. Each carries a copyable CLI command. Read it before running it — Cloud Guard never changes anything in your account.",
						"Export the list as CSV when you need to hand it to someone else.",
					},
				},
				{
					Heading: "What it will not do",
					Body:    "Cloud Guard has read-only access by design. It cannot delete a volume, release an address, or modify anything at all. Every fix is a command you run yourself, after reading it. Scans are scheduled or triggered by you — there is no continuous monitoring and nothing here detects intrusions.",
				},
			},
			Rules: []DocRule{
				{Name: "Unattached EBS volume", What: "A volume in state 'available' — attached to nothing, billed in full.", Priced: true},
				{Name: "Unassociated Elastic IP", What: "AWS charges for a reserved address precisely when it is not in use.", Priced: true},
				{Name: "Stale snapshot", What: "Snapshots older than your retention window. Default 90 days, configurable.", Priced: true},
				{Name: "gp2 volume that should be gp3", What: "gp3 is about 20% cheaper per GiB at the same or better baseline performance.", Priced: true},
				{Name: "Stopped EC2 instance", What: "Compute is not billed, but its attached storage still is.", Priced: false},
				{Name: "Idle EC2 instance", What: "Running with low utilisation — a candidate for downsizing.", Priced: false},
				{Name: "Public S3 bucket", What: "A bucket reachable by anyone. A security finding, not a cost one.", Priced: false},
				{Name: "Over-provisioned RDS", What: "An instance class larger than the workload appears to need.", Priced: false},
				{Name: "Cost summary", What: "Account-level spend pulled from Cost Explorer for context.", Priced: false},
			},
			ActionLabel: "Open the dashboard →",
			ActionURL:   "/dashboard",
		},
		{
			Slug:    "data-guard",
			Name:    "Data Guard",
			Tagline: "A firewall for bad data",
			Icon:    "📊",
			Stack:   "Go · PostgreSQL · Next.js · Docker",
			Summary: "Bad data rarely announces itself. A negative amount, a null where the schema promised a value, a row count that quietly halved — nothing errors, and the wrong number reaches a dashboard or a downstream job that acts on it. Data Guard validates API payloads and database rows against declarative rules and catches those failures before anything else consumes them.",
			Sections: []DocSection{
				{
					Heading: "How it works",
					Bullets: []string{
						"Rules are written as JSON: a field, and one or more checks on it. They are portable and readable without knowing Go.",
						"Data arrives one of two ways — pushed to a webhook at POST /ingest/api, or pulled from Postgres by the connector.",
						"For in-memory data the engine evaluates rules directly. For a database source it translates the rules into a SQL WHERE clause and asks the database for the failures, instead of pulling every row across the network to check it in Go.",
						"Runs, errors and alert state are persisted in Postgres so you have a history rather than a moment.",
						"Alerting is a state machine: it fires on the transition between PASS and FAIL, not on every failing run. A source that has been broken for two days does not produce two days of identical Slack messages.",
					},
				},
				{
					Heading: "A rule looks like this",
					Body:    "POST a source id, the schema you expect, the rules, and the rows. The response tells you which rows failed and why.",
					Bullets: []string{
						"source_id — names the stream, so history and alert state group correctly",
						"schema — the types you expect, e.g. {\"amount\": \"number\"}",
						"rules — e.g. field \"amount\", check {\"op\": \"gt\", \"value\": 0}",
						"data — the rows to validate",
					},
				},
				{
					Heading: "Running it",
					Bullets: []string{
						"Docker is the shortest path: docker build -t dataguard . then run it with DATABASE_URL set.",
						"Or run the Go server directly with go run cmd/server/main.go — it listens on :8080.",
						"Postgres is optional. Without it validation still works; you lose history and alerting.",
						"The Next.js dashboard lives in web/ — npm install then npm run dev, on :3000.",
					},
				},
			},
			Caveat:         "Data Guard is open source and self-hosted. There is no account here to sign into — you run it on your own infrastructure.",
			ActionLabel:    "View on GitHub →",
			ActionURL:      "https://github.com/singh-anurag-7991/data-guard",
			ActionExternal: true,
		},
		{
			Slug:    "shield",
			Name:    "Shield",
			Tagline: "Rate limiting that holds under load",
			Icon:    "🧱",
			Stack:   "Go · Gin · Redis",
			Summary: "An API with no rate limit has no floor: one bad client, one retry loop, one credential-stuffing script, and the service is down for everyone. Shield is a rate limiter in Go built to make the allow-or-deny decision cheap enough to sit in the request path of every call.",
			Sections: []DocSection{
				{
					Heading: "How it works",
					Bullets: []string{
						"Middleware sits in front of your handlers and identifies the caller by IP or API key.",
						"A factory picks the algorithm from configuration, so the choice is a config change rather than a code change.",
						"Four algorithms are implemented: token bucket, leaky bucket, fixed window, and sliding log. They trade memory against precision differently — token bucket absorbs bursts, sliding log is exact but stores a timestamp per request.",
						"The storage layer is separate from the algorithms, so counters can live in memory for a single instance or in Redis when several instances must share one limit.",
						"Rejections return a correct 429 with X-RateLimit-Limit, X-RateLimit-Remaining and X-RateLimit-Reset, so a well-behaved client can back off instead of hammering.",
					},
				},
				{
					Heading: "Measured throughput",
					Body:    "Benchmarked locally with wrk (12 threads, 400 connections, 30 seconds) on a small VPS. These are the repository's own numbers on that machine, not a claim about your hardware.",
					Bullets: []string{
						"Token bucket — about 55k requests/second, 20 MB across 10k keys",
						"Leaky bucket — about 52k requests/second, 22 MB",
						"Sliding log — about 45k requests/second, 35 MB",
					},
				},
				{
					Heading: "Running it",
					Bullets: []string{
						"Clone the repository, go mod tidy, then go run cmd/server/main.go. It starts on :8080.",
						"Hit /api/test once to see the headers, then in a loop to watch it return 429.",
						"Configure limits per API key and per endpoint.",
					},
				},
			},
			Caveat:         "Shield is open source and self-hosted, so there is nothing to sign into here. Note that its README lists Redis-backed distributed state under both shipped features and planned work — check the code before relying on multi-instance behaviour.",
			ActionLabel:    "View on GitHub →",
			ActionURL:      "https://github.com/singh-anurag-7991/shield",
			ActionExternal: true,
		},
	}
}

// productDoc returns the doc for a slug, or false when the slug is unknown.
func productDoc(slug string) (ProductDoc, bool) {
	for _, d := range productDocs() {
		if d.Slug == slug {
			return d, true
		}
	}
	return ProductDoc{}, false
}
