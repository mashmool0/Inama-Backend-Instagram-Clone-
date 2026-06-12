# Implementation Roadmap

Two developers, working in parallel. This roadmap is built so that after a shared
foundation, each developer owns separate services and rarely blocks the other.

> This is a direction, not a contract. Phases will shift as we learn. The goal is
> a clear path and clean parallel work — not a fixed schedule.

---

## The Core Idea: Contract-First Parallel Work

The single thing that lets two people work independently:

```
Phase 0 defines ALL the contracts (proto files + event schemas) up front.

After that, each developer codes against the agreed interface.
When Dev B needs Dev A's service that isn't built yet → mock it.
The contract is the source of truth, not the running service.
```

Without this, Dev B's Feed service waits for Dev A's Posts service, and parallel
work collapses into sequential work. With it, both build against the same `.proto`
and only sync at phase boundaries.

---

## Service Ownership

Split so each developer gets both languages, balanced complexity, and minimal
cross-blocking. Dependencies flow so each phase's prerequisites are already done.

```
Dev A track                        Dev B track
───────────                        ───────────
Auth         (FastAPI)             User / Social Graph  (Go)
Posts        (Go)                  Feed + Explore       (Go)
Search       (FastAPI)             Notifications        (Go)

Gateway — shared. Skeleton built in Phase 0.
          Each dev adds their own service's routes as they go.
          (Different path prefixes → near-zero merge conflicts.)
```

Why this split:
- **No cross-blocking within a phase** — the two services in each phase are independent
- **Dependencies satisfied in advance** — Feed (B, Phase 2) needs User (B, Phase 1, already done) and Posts events (A, Phase 2, contract mocked until integration)
- **Both learn both stacks** — each dev writes Go and Python
- **Balanced** — Dev B owns Feed (hardest service) but lighter Auth-equivalent; Dev A owns Posts (outbox complexity) plus two medium services

---

## Dependency Map

```
            ┌─────────┐
            │  Auth   │  (no deps — pure identity)
            └─────────┘
            ┌─────────┐
            │  User   │  (no deps — profiles + follows)
            └────┬────┘
                 │ GetFollowers (gRPC)
                 ▼
┌─────────┐  events   ┌─────────┐
│  Posts  │──────────▶│  Feed   │  (needs Posts events + User)
└────┬────┘           └─────────┘
     │ events
     ├──────────▶ ┌──────────────┐
     │            │ Notifications│  (needs Posts + User events)
     │            └──────────────┘
     └──────────▶ ┌─────────┐
                  │ Search  │  (needs Posts events)
                  └─────────┘
```

Read top to bottom = build order. Nothing below can be integrated until what it
points from exists. But all can be *developed* in parallel against contracts.

---

## Timeline at a Glance

```
Phase 0   Foundation            BOTH together      ~1-2 weeks
Phase 1   Identity & Profiles   A: Auth | B: User  ~2 weeks
Phase 2   Content & Feed        A: Posts | B: Feed ~2-3 weeks
Phase 3   Reactions & Discovery A: Search | B: Notif ~2 weeks
Phase 4   Hardening + Load Test BOTH together      ~2 weeks
Phase 5   Frontend & E2E        A: UI | B: backend polish ~2-3 weeks
                                                   ─────────────
                                                   ~11-14 weeks
```

---

## BLOCKERS — Open Decisions That Gate Phases

Two decisions are still TBD and **must be made before the phase that needs them**:

| Decision | Needed by | Why |
|---|---|---|
| **Kafka vs RabbitMQ** | Start of Phase 2 | First real events flow in Phase 2 (post.created). Cannot build the consumer without the broker chosen. |
| **feed_db: Postgres vs Redis** | Middle of Phase 2 | Feed Reader and fanout writer need to know the storage. |

Decide Kafka/RabbitMQ during Phase 0 or 1 so Phase 2 is not blocked.

---

# Phase 0 — Foundation

**Both developers, together. Do not split yet.**
Everything after this depends on getting the skeleton right once.

### Goal
`docker compose up` starts the whole stack. One gRPC call works end to end.
CI is green. Both developers can run everything locally and agree on all contracts.

### Steps

**0.1 — Repo scaffold**
```
inama/
├── services/{gateway,auth,user,posts,feed,notifications,search}/
├── proto/
├── infra/{nginx,postgres,prometheus,grafana}/
├── libs/            # shared conventions (see 0.5)
├── docker-compose.yml
├── docker-compose.dev.yml
└── .github/workflows/
```

**0.2 — Define ALL proto contracts (most important step)**
Write every `.proto` now, even for services not built yet. Agree on them together.
```
proto/auth.proto     — Register, VerifyOTP, Login, RefreshToken
proto/user.proto     — GetProfile, UpdateProfile, Follow, Unfollow, GetFollowers
proto/posts.proto    — CreatePost, DeletePost, GetPost, LikePost, AddComment
proto/feed.proto     — GetFeed, GetExplore
proto/notif.proto    — GetNotifications, MarkAsRead
proto/search.proto   — Search
```
Also define **event schemas** (the async contracts):
```
events: post.created, post.liked, comment.created, user.followed, user.updated
```
Document each event's payload in `docs/04-communication.md`. These are contracts too.

**0.3 — Infrastructure in Docker Compose**
Postgres (with `init.sql` creating all 5 databases), Redis, the chosen message
broker, Elasticsearch, Prometheus, Grafana, Nginx. All as containers.

**0.4 — Two "hello world" services through the gateway**
- One Go service (e.g. a stub `user`) exposing one gRPC method
- One Python service (e.g. a stub `auth`) exposing one gRPC method
- Gateway routes an HTTP request → gRPC → response
- Proves: proto generation works for both languages, service discovery works,
  the gateway pattern works

**0.5 — Shared conventions (`libs/`)**
Agree once so both devs write consistent code:
- Structured logging format (JSON, with trace_id field reserved)
- Config loading (env vars → typed config struct)
- Error handling pattern (gRPC status codes, error wrapping)
- The Handler → Service → Repository skeleton as a template per language
- Health check endpoint convention (`/health`)
- Metrics endpoint convention (`/metrics`)

**0.6 — CI/CD pipeline**
- GitHub Actions: on PR → lint + unit tests per changed service (path filters)
- On merge to main → deploy to VPS (SSH + `docker compose up --build -d`)
- Branch protection: no direct push to main, PR + 1 review required

**0.7 — Monitoring base**
- Prometheus scraping config (even with two stub services)
- Grafana running with Prometheus as data source
- One basic dashboard (request rate, even if near zero)

### Exit Criteria (Definition of Done)
- [ ] `docker compose up` → entire stack healthy
- [ ] HTTP request → gateway → Go service → response works
- [ ] HTTP request → gateway → Python service → response works
- [ ] All proto files written and code-generating for Go + Python
- [ ] All event schemas documented
- [ ] CI runs tests on PR, deploys on merge
- [ ] Grafana shows metrics from the stub services
- [ ] Kafka vs RabbitMQ decided (or scheduled before Phase 2)

---

# Phase 1 — Identity & Profiles

**Parallel. Zero cross-dependency — the cleanest parallel phase.**

```
Dev A: Auth Service (FastAPI)        Dev B: User / Social Graph (Go)
```

### Per-Service Step Pattern (both follow this)
This is the rhythm for every service from here on:
```
1. Schema + migrations       → design tables, write migration files
2. Repository layer          → DB access only, unit tested against real Postgres
3. Service layer             → business logic, unit tested with mocked repo
4. Handler layer             → gRPC handlers, thin — translate proto ↔ service
5. Wire gRPC server          → expose the proto, register handlers
6. Metrics + logging         → /metrics, structured logs, health check
7. Dockerfile                → containerize
8. Integration test          → full service against real DB in CI
9. Gateway routes            → add HTTP→gRPC routes for this service
```

### Dev A — Auth Service
```
Steps:
  1. auth_db schema: users, otp_codes, refresh_tokens, password_reset_tokens
  2. Repositories: UserRepository, TokenRepository
  3. Services: OTPService (Redis TTL), JWTService, PasswordService (bcrypt)
  4. Handlers: Register, VerifyOTP, Login, RefreshToken, RequestReset, ResetPassword
  5. SMS client (mock provider in dev), Email client (mock in dev)
  6. Unit tests: OTP expiry, JWT issue/validate, bcrypt verify
  7. Integration test: full register → OTP → login flow against real auth_db + Redis
  8. Gateway: add /auth/* routes
```
**Security focus** — this is the most security-critical service. Review carefully.

### Dev B — User / Social Graph Service
```
Steps:
  1. user_db schema: users, follows (with composite indexes)
  2. Repositories: UserRepository, FollowRepository
       → FollowRepository.GetFollowers(user_id) — Feed will depend on this
  3. Services: ProfileService, FollowService (duplicate-follow guard, counters)
  4. Handlers: GetProfile, UpdateProfile, Follow, Unfollow, GetFollowers, GetFollowing
  5. Unit tests: follow/unfollow idempotency, counter correctness
  6. Integration test: follow flow + GetFollowers against real user_db
  7. Gateway: add /users/* routes
```

### Integration Point (end of phase, both together)
- Auth issues a JWT → Gateway's JWT middleware verifies it → user_id reaches User service
- Manual test: register → login → get token → view/edit profile → follow someone

### Exit Criteria
- [ ] Register, OTP verify, login, refresh all work end to end
- [ ] Profile view/edit works with a real JWT
- [ ] Follow/unfollow works, follower counts correct
- [ ] `GetFollowers` returns correct IDs (Feed depends on this next phase)
- [ ] Both services have unit + integration tests passing in CI
- [ ] Both services expose metrics

---

# Phase 2 — Content & Feed

**Parallel. This is where the first async event flows — the integration is bigger.**

```
Dev A: Posts + Interactions (Go)     Dev B: Feed + Explore (Go)
```

> ⚠️ Requires Kafka/RabbitMQ decided, and feed_db storage decided.

### Dev A — Posts + Interactions
```
Steps:
  1. posts_db schema: posts, likes, comments, OUTBOX table
  2. Repositories: PostRepository, LikeRepository, CommentRepository, OutboxRepository
  3. Services:
       PostService    — create (writes post + outbox event in ONE transaction)
       LikeService    — atomic counter (UPDATE ... like_count + 1), idempotent
       CommentService — insert + atomic comment_count
  4. S3 client — upload media to S3/Liara (mock in dev)
  5. Outbox Worker — polls outbox table, publishes events to queue, marks published
  6. Handlers: CreatePost, DeletePost, GetPost, LikePost, UnlikePost, AddComment, GetComments
  7. Unit tests: atomic counter (no race), outbox written in same TX
  8. Integration test: create post → event appears in queue
  9. Gateway: add /posts/* routes
```
**Pattern focus** — the Outbox pattern. Get the dual-write guarantee right.

### Dev B — Feed + Explore
```
Steps (Phase 1 version = fan-out-on-read, intentionally simple):
  1. feed_db schema: feed_items (for the fan-out-on-write migration later)
  2. Feed Reader — fan-out-on-read query (join followed users' posts, paginate)
       → mock post data until Posts integration; code against posts.proto
  3. Explore Service — random recent public posts from Redis cache (TTL 5min)
  4. Queue Consumer + Fanout Worker — consumes post.created
       → calls User.GetFollowers (real, from Phase 1)
       → writes feed_items (skeleton; not the active read path yet)
  5. Handlers: GetFeed (cursor pagination), GetExplore
  6. Unit tests: pagination cursor correctness, explore cache fallback
  7. Integration test: GetFeed returns followed users' posts
  8. Gateway: add /feed and /explore routes
```

### Integration Point (the big one)
```
Dev A creates a post
  → post.created event published (outbox worker)
  → Dev B's fanout worker consumes it
  → calls GetFollowers, writes feed_items
  → GetFeed returns the post
```
This is the first full async flow across two developers' services. Test it together.

### First Load Test (end of phase)
```
Load test GET /feed with fan-out-on-read.
Expected: it struggles as followee count grows — this is the planned bottleneck.
Record the numbers. This is the baseline for the Phase 4 migration.
```

### Exit Criteria
- [ ] Create post (text + media) works, event published via outbox
- [ ] Likes atomic and idempotent, counts correct
- [ ] Comments work
- [ ] Fanout worker consumes post.created and calls GetFollowers
- [ ] GetFeed returns correct paginated feed
- [ ] Explore returns cached recent posts
- [ ] First load test run and numbers recorded

---

# Phase 3 — Reactions & Discovery

**Parallel. Both are async consumers of events that already flow from Phase 2.**

```
Dev A: Search Indexer (FastAPI)      Dev B: Notifications (Go)
```

### Dev A — Search Indexer
```
Steps:
  1. Elasticsearch index mapping (posts: caption, hashtags, author, created_at)
  2. Queue Consumer — consumes post.created
  3. Indexer Service + Hashtag Extractor — build ES document, index it
  4. Search Service — build ES queries (by username, hashtag, full-text)
  5. Handler: Search(query, type) via gRPC
  6. ES client (elasticsearch-py)
  7. Unit tests: hashtag extraction, query building
  8. Integration test: post created → indexed → searchable
  9. Gateway: add /search route
```

### Dev B — Notifications
```
Steps:
  1. notif_db schema: notifications (type, actor, recipient, post_id, read)
  2. Queue Consumer — consumes post.liked, comment.created, user.followed
  3. NotificationService — create record, determine recipient, dedupe
  4. Push Client — FCM/APNs (mock in dev), async delivery
  5. Handlers: GetNotifications, MarkAsRead, MarkAllRead
  6. Unit tests: dedupe logic, recipient resolution
  7. Integration test: like a post → notification created + push attempted
  8. Gateway: add /notifications/* routes
```

### Integration Point
- Like a post → notification appears for the post author
- Create a post → it becomes searchable by hashtag within seconds
- Both consume from the same event stream Posts/User already publish

### Exit Criteria
- [ ] Liking/commenting/following creates the right notification
- [ ] Push delivery attempted async (failure doesn't break the event)
- [ ] Mark-as-read works
- [ ] Search by username, hashtag, and text returns correct results
- [ ] Posts are searchable shortly after creation

---

# Phase 4 — Hardening & Load Testing

**Both together. The system is feature-complete; now make it hold under load.**

### Goal
Hit the test target: **1,000 req/sec sustained, p95 < 500ms.**
Learn where it breaks, fix it, re-test.

### Steps
```
4.1  Add OpenTelemetry tracing to ALL services
       → trace a request across gateway → service → DB
       → REQUIRED before serious load testing (to see where time goes)

4.2  Build the load test suite (k6)
       → feed read, post like, full flow, search
       → open model (constant arrival rate) for realistic load

4.3  Load test → find bottlenecks (predicted order):
       1. Feed query (fan-out-on-read) — slowest, fix first
       2. Missing DB indexes
       3. Connection pool exhaustion
       4. Replica lag under write load
       5. Fanout worker backlog

4.4  THE BIG MIGRATION: Feed fan-out-on-read → fan-out-on-write
       → activate feed_items as the read path
       → fanout worker becomes the primary write path
       → re-test → confirm feed reads are now constant-time

4.5  Add read replicas, route reads → replicas, writes → primary
       → handle read-your-own-writes (route author's own reads to primary)

4.6  Re-test until target met. Document the before/after numbers.
```

### Exit Criteria
- [ ] OpenTelemetry tracing live across all services
- [ ] Feed migrated to fan-out-on-write
- [ ] Read/write split with replicas
- [ ] 1k req/sec at p95 < 500ms achieved
- [ ] Before/after numbers documented in docs/08-load-testing.md

---

# Phase 5 — Frontend & End-to-End

**Can overlap with Phase 4. Backend is stable; frontend consumes it.**

```
Dev A: Next.js frontend              Dev B: backend polish + E2E tests
```

### Steps
```
5.1  Next.js app: auth flow, feed, profile, post create, notifications
5.2  Connect to the real API Gateway (cursor pagination, JWT refresh)
5.3  End-to-end tests (Playwright): register → post → feed → like → notification
5.4  Final full-system load test with realistic traffic mix
5.5  Trace a slow request end to end using OpenTelemetry — close the loop
```

### Exit Criteria
- [ ] Full user journey works in the browser
- [ ] E2E tests cover the critical flows
- [ ] Final load test passes
- [ ] You can debug a slow request via traces

---

## Collaboration Rules (How Two Devs Stay Unblocked)

```
1. Contract-first.      Proto + event schemas agreed in Phase 0. Never change a
                        shared contract without telling the other dev.

2. Own your services.   Each dev is the owner of their services. The other reviews
                        PRs but doesn't edit without asking.

3. Mock across the gap. Need a service that isn't built yet? Mock it against the
                        agreed proto. Integrate at the phase boundary.

4. Branch per feature.  feature/<service>-<thing>. Service folders rarely collide,
                        so merge conflicts are rare. proto/ changes need a heads-up.

5. PR + 1 review.       No direct push to main. The reviewer learns the other half
                        of the system — that's a feature, not overhead.

6. Sync at boundaries.  Each phase ends with a joint integration test. That's the
                        moment the two tracks reconnect.

7. Update docs/.        A decision changes → update the doc in the same PR. Stale
                        docs break the contract-first model.

8. Shared things change together. proto/, libs/, docker-compose.yml, init.sql —
                        touching these needs both devs aware (they affect everyone).
```

---

## What Will Probably Change (and That's Fine)

- Phase durations — first service always takes longer than expected
- feed_db choice — may flip after the Phase 2 load test
- Kafka vs RabbitMQ — pick early, but the consumer abstraction should make it swappable
- Celebrity-hybrid fanout — may not be needed at this scale; revisit after Phase 4
- The ownership split — if one track runs ahead, rebalance the next phase

The roadmap is the map. The territory will surprise us. Update the map as we go.
