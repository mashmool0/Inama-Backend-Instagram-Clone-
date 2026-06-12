# Load Testing

## Target

```
Must pass on VPS:    1,000 req/sec sustained, p95 < 500ms
Design target:       10,000 req/sec (architectural reasoning)
```

The gap between these two is not failure — it's intentional.
Build for 10k (the patterns), test at 1k (the reality on one VPS).

---

## Why Load Testing Is the Learning

The process is more valuable than the number:

```
1. Build the system
2. Load test it
3. Watch where it breaks (it will break — that's the point)
4. Read the traces: which service, which query
5. Fix the bottleneck
6. Re-test → higher ceiling
7. Repeat
```

Each cycle teaches something concrete:
- Cycle 1: probably the feed query (fan-out-on-read)
- Cycle 2: probably missing DB index
- Cycle 3: probably replica lag or connection pool exhaustion
- Cycle 4: probably the fanout worker backlog

You cannot learn these lessons from a book. You must observe them.

---

## Tools

### k6 (Primary Load Testing Tool)

k6 is a JavaScript-based load testing tool. Clean syntax, good output, free.

```javascript
// k6 test: simulate 1000 users requesting their feed
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 100 },   // ramp up to 100 users
    { duration: '3m', target: 1000 },  // ramp up to 1000 users
    { duration: '5m', target: 1000 },  // hold at 1000 for 5 min
    { duration: '1m', target: 0 },     // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // p95 must be < 500ms
    http_req_failed:   ['rate<0.01'],  // error rate < 1%
  },
};

export default function () {
  const res = http.get('https://inama.yourdomain.com/api/feed', {
    headers: { Authorization: `Bearer ${__ENV.JWT_TOKEN}` },
  });
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(1);
}
```

Run:
```bash
k6 run --env JWT_TOKEN=yourtoken feed_test.js
```

Output shows: req/sec, p50/p95/p99 latency, error rate, data transfer.

---

## Open Model vs Closed Model

Important concept for accurate load testing:

```
CLOSED MODEL (default, simpler)     OPEN MODEL (more realistic)
────────────────────────────────    ────────────────────────────
N virtual users, each waits         Arrivals independent of responses
for response before sending next    New requests arrive at fixed rate
                                    regardless of how many are pending

If response is slow (800ms):        If response is slow (800ms):
→ user waits 800ms before next      → new requests keep arriving
→ req/sec drops automatically       → queue builds up → system degrades
→ load test hides the problem       → more realistic real-world behavior
```

For the feed endpoint, use open model to simulate Instagram-like load:
```javascript
export const options = {
  scenarios: {
    feed_load: {
      executor: 'constant-arrival-rate',  // open model
      rate: 1000,                         // 1000 arrivals per second
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 2000,
    },
  },
};
```

---

## What to Load Test (and in What Order)

Test endpoints from most to least critical:

```
Priority 1 — Feed read (GET /feed)
  → most frequent, fan-out-on-read is the first bottleneck
  → test with 50, 100, 500, 1000 req/sec

Priority 2 — Post like (POST /posts/:id/like)
  → most frequent write, atomic counter stress test
  → test concurrent likes on the same post

Priority 3 — Full flow (register → post → like → feed)
  → end-to-end latency including all service hops
  → where gRPC overhead shows up

Priority 4 — Search
  → Elasticsearch query latency under load

Priority 5 — Fanout throughput
  → create posts rapidly, measure how far behind the fanout worker falls
  → measure queue consumer lag in Grafana
```

---

## Interpreting Results

```
p50 = 50% of requests took less than this  (typical user experience)
p95 = 95% of requests took less than this  (your SLA target)
p99 = 99% of requests took less than this  (worst case)

If p50 = 100ms and p99 = 4000ms:
  Most users are fine. A few are waiting 4 seconds.
  That gap means something is blocking some requests.
  Open traces → find the outliers → they tell you why.

Target: p95 < 500ms at 1000 req/sec
```

## Bottlenecks You Will Find (Prediction)

In order of likelihood:

1. **Feed query slow** — fan-out-on-read with missing composite index on `posts(author_id, created_at)`
2. **DB connection pool exhaustion** — default pool too small for concurrent load
3. **Replica lag** — reads from replica see stale data under write load
4. **Fanout worker falling behind** — queue consumer lag grows, followers see posts late
5. **Redis connection limit** — rate limiter in gateway exhausting Redis connections
6. **gRPC timeout misconfiguration** — default timeouts too short, cascading failures

Each one is a lesson. Finding them through load testing is the point.
