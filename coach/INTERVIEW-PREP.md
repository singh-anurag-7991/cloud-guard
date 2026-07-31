# Interview Prep Roadmap — Go Backend / SDE-3 (5 yrs exp)

Banaya: 2026-08-02 · Timeline: **14 din, aggressive** · Target: Go backend / broader SDE role

## Context aur strategy

20+ Go interviews de chuke ho, offer nahi mila, feedback bhi nahi mila. 5 saal ke experience pe interviewers sirf syntax nahi poochte — wo dekhte hain ki tum **design decisions justify kar sakte ho ya nahi**. Isliye is plan mein DSA sirf ek chhota hissa hai; asli weight **Go depth + system design + distributed systems** pe hai, kyunki 5-yr candidates wahi pe sabse zyada reject hote hain.

### 5 saal pe reject hone ki 4 सबसे common wajah

1. **Go concurrency theek se explain nahi kar paana** — goroutine chala dete hain par "iska lifecycle kya hai, leak kaise hoga, context cancel kaise propagate hoga" pe atak jaate hain.
2. **System design mein structure na hona** — seedha database schema pe kood jaate hain bina requirements clarify kiye, scale estimate kiye, ya trade-offs bataye.
3. **"Kyun" ka jawab na hona** — "Kafka use kiya" bol dete hain par "RabbitMQ kyun nahi" ka jawab nahi hota. 5 yrs pe *kyun* hi asli sawal hai.
4. **Apne hi project ko bech na paana** — tumhare paas Cloud Guard hai (Go + AWS + Stripe + auth + scanners, solo built). Ye ek killer talking point hai agar structured tarike se sunao.

### Har topic ka rule (ye sabse important hai)

Har topic ke liye teen cheezein honi chahiye:
- **60-90 second ka bola hua explanation** — jaise interviewer ke saamne bol rahe ho, khud sun ke check karo
- **10-20 line ka code snippet** memory se likha hua (Go topics ke liye)
- **Ek trade-off statement** — "X use karenge jab ye, Y use karenge jab wo, kyunki..."

Sirf padhna kaafi nahi hai. Bolne ki practice karo — mujhse kabhi bhi mock maang lo.

---

## 14-din ka plan

Roz ka Interview Prep block: **60-75 min** (agar zyada time mile to Go/System Design ko do, DSA ko nahi).
Har din ka format: **~20 min DSA + ~45 min core topic**. DSA chhota rakha hai jaan-boojh ke — 5 yrs pe DSA screening clear karne bhar chahiye, usse aage return kam hai.

---

### 🟦 Week 1 — Language depth + data layer + services

**Day 1 — Go Concurrency Part 1 (सबसे important din)**
- Goroutines: kaise schedule hoti hain (G-M-P model ka basic idea), stack growth
- Channels: unbuffered vs buffered — kab deadlock hota hai aur kyun
- `select` statement, `default` case, timeout pattern
- Code likho: worker pool (N workers, jobs channel, results channel, WaitGroup se close)
- Interview Q: "Unbuffered channel pe send kab block karta hai?" · "Closed channel se read karoge to kya hoga?" · "Nil channel pe send/receive?"

**Day 2 — Go Concurrency Part 2 + context**
- `sync` package: Mutex vs RWMutex, WaitGroup, Once, `sync.Map` kab use karna hai
- `context`: cancellation, timeout, deadline, values — aur ye **propagate kaise hota hai** call chain mein
- Goroutine leaks: kaise hote hain, kaise detect karo, kaise rokte hain
- Race conditions: `go run -race`, atomic operations
- Code likho: context ke saath cancellable pipeline (fan-out → fan-in)
- Interview Q: "Goroutine leak ka ek example do" · "Context cancel karne pe running goroutine turant rukti hai kya?" (nahi — cooperative hai)

**Day 3 — Go internals + gotchas**
- Slices: len vs cap, append ka aliasing bug, `copy`, slice of slice
- Maps: internal bucket structure, iteration order random kyun, concurrent map write panic
- Interfaces: implicit implementation, **nil interface vs nil pointer** (classic trap), type assertion vs type switch, empty interface
- Escape analysis: stack vs heap, `go build -gcflags="-m"`
- GC basics: concurrent mark-and-sweep, GOGC, GC pause kyun matter karta hai
- Gotchas: loop variable capture (Go 1.22 se pehle vs baad), `defer` loop mein, defer with named return values
- Interview Q: "Ye function nil return kar raha hai par `err != nil` true aa raha hai, kyun?"

**Day 4 — Databases Part 1: SQL depth**
- Indexes: B-tree kaise kaam karta hai, composite index ka column order kyun matter karta hai, covering index
- `EXPLAIN` / query plan padhna — full scan vs index scan
- Transactions + ACID, isolation levels (Read Committed, Repeatable Read, Serializable) aur har level pe kya anomaly bachti hai (dirty read, non-repeatable read, phantom read)
- Locking: optimistic vs pessimistic, deadlock kaise hota hai
- Connection pooling (Go mein `database/sql` — `SetMaxOpenConns`, `SetMaxIdleConns`) — ye interview mein aata hai
- N+1 query problem
- Interview Q: "Is slow query ko kaise optimize karoge?" · "Isolation level kaunsa choose karoge aur kyun?"

**Day 5 — Databases Part 2: NoSQL + scaling**
- SQL vs NoSQL — **kab kaunsa, aur kyun** (ye asli sawal hai, definition nahi)
- Redis: data structures, caching patterns (cache-aside, write-through, write-behind), TTL, eviction policies, Redis ko distributed lock ki tarah use karna aur uske खतरे
- MongoDB / document DB basics: schema design, embedding vs referencing
- **Scaling:** replication (leader-follower), read replicas aur replication lag, sharding strategies (range, hash, geo), partitioning
- CAP theorem — practical version, na ki ratta
- Cache invalidation strategies, thundering herd / cache stampede
- Interview Q: "Cache aur DB inconsistent ho gaye to kya karoge?" · "Sharding key kaise choose karoge?"

**Day 6 — REST API design + Go web layer**
- REST principles, correct HTTP methods aur status codes (201 vs 200, 202, 204, 409, 422, 429)
- Idempotency — POST ko idempotent kaise banate ho (idempotency key), retries ke saath ye kyun zaruri hai
- Pagination: offset vs cursor-based (aur offset large data pe kyun bekaar hai)
- Versioning strategies (URL, header), backward compatibility
- Auth: JWT vs session, refresh token flow, token expiry, JWT ke risks
- Rate limiting: token bucket, sliding window (tumne **Shield** mein ye already socha hai — ye interview mein bolo)
- Error response design, request validation, timeouts, graceful shutdown Go mein
- REST vs gRPC vs GraphQL — trade-offs
- Interview Q: "Payment API ko duplicate charge se kaise bachaoge?"

**Day 7 — Weekly review + mock #1**
- Week 1 ke topics khud se revise karo — jo bhi bolne pe atka, wo note karo
- **Mujhse full mock lo:** 1 Go concurrency grilling + 1 DB design question
- Cloud Guard ki STAR story likho aur bolke practice karo (neeche template hai)

---

### 🟥 Week 2 — Distributed systems + infra + system design (yahan game jeeta jaata hai)

**Day 8 — Message queues + async architecture**
- Queue kyun chahiye: decoupling, load leveling, retries, spike absorption
- **Kafka:** topics, partitions, consumer groups, offsets, ordering guarantee (sirf partition ke andar), rebalancing
- **RabbitMQ:** exchanges, routing keys, ack/nack — aur Kafka se kab better hai
- Delivery semantics: at-most-once, at-least-once, exactly-once (aur exactly-once practically kitna mushkil hai)
- **Idempotent consumers** — at-least-once ke saath ye compulsory hai
- Dead letter queue, retry with exponential backoff + jitter
- Outbox pattern (DB write + event publish atomically)
- Interview Q: "Message do baar process ho gaya to?" · "Ordering kaise guarantee karoge?"

**Day 9 — Microservices architecture**
- Monolith vs microservices — **kab microservices actually galat choice hai** (ye bolna maturity dikhata hai)
- Service boundaries kaise decide karo (domain-driven, bounded context)
- Inter-service communication: sync (REST/gRPC) vs async (events)
- **Distributed transactions:** two-phase commit kyun avoid karte hain, **Saga pattern** (choreography vs orchestration), compensating transactions
- Resilience: circuit breaker, retry + backoff, bulkhead, timeout budgets, graceful degradation
- Service discovery, API gateway, BFF pattern
- Observability: structured logging, metrics (RED/USE), **distributed tracing** (trace ID propagation)
- Config management, secrets handling
- Interview Q: "Ek service down ho gayi to baaki system pe kya asar?" · "Saga rollback kaise karoge?"

**Day 10 — Docker + containers**
- Image vs container, layers aur layer caching
- **Multi-stage builds** — Go ke liye ye perfect example hai (builder stage + scratch/distroless final image, ~10MB binary). Ye definitely poocha jaata hai.
- Dockerfile best practices: layer order, `.dockerignore`, non-root user, specific tags (latest nahi)
- Networking basics, volumes vs bind mounts
- docker-compose (tumhare cloud-guard repo mein already hai — usko explain kar paana chahiye)
- Container security basics: image scanning, minimal base images
- Interview Q: "Go app ki image size 800MB se 15MB kaise laoge?"

**Day 11 — Kubernetes**
- Core objects: Pod, ReplicaSet, Deployment, Service (ClusterIP/NodePort/LoadBalancer), Ingress
- ConfigMap vs Secret
- **Probes: liveness vs readiness vs startup** — inka difference bahut poocha jaata hai, aur galat probe se outage kaise hota hai
- Resource requests vs limits, OOMKilled, CPU throttling
- HPA (autoscaling), basic scheduling concepts
- StatefulSet vs Deployment — database jaise stateful workload ke liye
- Rolling update mechanics, `maxSurge` / `maxUnavailable`
- Debugging: `kubectl logs`, `describe`, `exec`, CrashLoopBackOff kaise debug karte ho
- Interview Q: "Pod CrashLoopBackOff mein hai — step by step kaise debug karoge?"

**Day 12 — Deployment strategies + CI/CD + reliability**
- **Deployment strategies:** rolling, blue-green, canary, feature flags — har ek ka trade-off (rollback speed vs infra cost vs risk)
- Zero-downtime deployment: graceful shutdown, connection draining, readiness probe ka role
- **Database migrations with zero downtime** — expand-contract pattern (ye senior-level sawal hai, isko zaroor samjho: pehle naya column add karo, dono jagah likho, backfill karo, phir purana hatao)
- CI/CD pipeline design: build → test → scan → deploy, artifact promotion across environments
- Rollback strategy, monitoring aur alerting during deploy
- SLI / SLO / error budget basics
- Incident response: detection → mitigation → root cause → postmortem
- Interview Q: "Deploy ke baad error rate spike ho gaya — ab kya karoge?"

**Day 13 — System Design (sabse zyada weightage) 🔥**

Framework — har design question isi order mein karo, kabhi skip mat karo:
1. **Requirements clarify** (2-3 min): functional + non-functional. Sawal pucho, assume mat karo.
2. **Scale estimate**: DAU, QPS (read vs write ratio), data size, peak vs average
3. **High-level design**: boxes aur arrows — client, LB, services, DB, cache, queue
4. **Deep dive**: API contract, data model, ek-do component ko detail mein
5. **Bottlenecks + scaling**: kahan tootega, kaise fix karoge
6. **Trade-offs bolo** — "Maine X choose kiya kyunki Y, iska cost Z hai"

Core building blocks jinke trade-offs pata hone chahiye: load balancer (L4 vs L7), caching layers, CDN, DB replication/sharding, message queue, consistent hashing, rate limiter, idempotency, leader election, eventual vs strong consistency.

Practice designs (2-3 chuno, poore end-to-end karo):
- URL shortener (classic warm-up — hashing, collisions, read-heavy caching)
- Rate limiter (tumhara **Shield** product literally yahi hai — huge advantage, isko sunao)
- Notification / alerting system (queue, fan-out, retries, dedup — Cloud Guard ke Slack alerts se relate karta hai)
- **Cloud cost monitoring system** ← ye tumhara Cloud Guard hi hai. Isko design question ki tarah practice karo: multi-tenant scanning, scheduled jobs, AWS API rate limits, findings storage, alerting. Agar interview mein ye aa gaya to tum jeet gaye.
- News feed / chat system (agar time bacha to)

**Day 14 — Full mock + behavioral + final revision**
- **Mujhse full mock interview lo**: 1 system design (45 min, poora framework use karke) + Go rapid-fire round
- Behavioral / STAR stories final polish (neeche)
- Weak topics ki list banao — jo bhi abhi tak shaky hai, us par focused revision

---

## Behavioral / STAR stories (roz 5 min, ignore mat karna)

5 saal pe behavioral round bhi reject kar sakta hai. Ye 5 stories **likh ke** rakho, phir bol ke practice karo (STAR: Situation → Task → Action → Result, result mein number ho to best):

1. **Sabse tough technical problem** jo solve kiya (production bug, performance issue, outage)
2. **Ek system jo tumne design/build kiya** → **Cloud Guard** perfect hai: solo built, Go + AWS SDK + SQLite, STS AssumeRole se secure onboarding (static keys nahi), Stripe billing 3 tiers, scanners for EC2/S3/RDS/Cost Explorer, Slack alerts, CloudFormation 1-click setup. Ye ek strong end-to-end ownership ki kahani hai.
3. **Conflict / disagreement** team ya manager ke saath — kaise resolve kiya
4. **Failure ya galti** — kya seekha (honest raho, defensive mat bano)
5. **"Why are you leaving?"** — ye zaroor prepare karo. Current company ko badnaam mat karo. Frame: "growth aur bigger scale ke problems chahiye" — negative nahi, forward-looking.

Aur "Do you have questions for us?" ke liye 3 sawal ready rakho — ye judge hota hai.

---

## Reject hone ke chances kam karne ke 5 quick rules

1. **Bolo, chup mat raho.** Coding round mein soch ke chup ho jaana sabse bada red flag hai. Loud thinking karo.
2. **Pehle clarify karo, phir solve karo.** Especially system design mein — direct solution pe koodna junior signal hai.
3. **Har choice ke saath "kyun" bolo.** "Postgres use karunga" nahi — "Postgres use karunga kyunki relational constraints chahiye aur scale abhi is range mein hai; agar write volume 10x hota to sharding ya Cassandra consider karta."
4. **Interview ke baad turant likho** ki kya poocha gaya aur kahan atke — evening check-in mein ye batao, main pattern nikal ke us topic ko plan mein aage laa dunga.
5. **Har interview ke end mein feedback maango** — "koi area jahan main improve kar sakta hoon?" Kabhi kabhi mil jaata hai, aur wahi asli data hai jo abhi missing hai.

---

## 14 din ke baad

Ye plan khatam hone pe: weak topics revisit karo, har hafte 2 system design mock, aur DSA ko medium level pe le jao. Jab tak offer nahi aata, roz ka rhythm chalta rahega.
