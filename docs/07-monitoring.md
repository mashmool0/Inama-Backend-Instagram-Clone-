# Monitoring

## The Three Pillars

```
METRICS                      LOGS                         TRACES
───────                      ────                         ──────
What is happening now?       What happened?               Where did the time go?

"feed p95 = 800ms"           "2026-06-12 auth error:      "this request took 800ms:
"500 req/sec on /feed"        invalid OTP for user 42"     - gateway: 5ms
"Postgres CPU 80%"           Structured, searchable        - user svc gRPC: 620ms ← !!
                             Per-service, centralized      - DB query: 180ms"
Prometheus + Grafana         Loki or stdout + Grafana      OpenTelemetry + Tempo
```

In a monolith, logs are enough — one process, one log file.
In microservices, **traces** are essential — a slow request crosses 4 services,
and without traces you cannot tell which one is responsible.

---

## Prometheus + Grafana (Metrics)

### How Prometheus Works

Prometheus **scrapes** (pulls) metrics from your services on a schedule.
Each service exposes a `/metrics` endpoint.

```
Every 15 seconds:
  Prometheus → GET auth:8080/metrics
  Prometheus → GET user:8080/metrics
  Prometheus → GET posts:8080/metrics
  ...stores the numbers in its time-series database
```

Your services expose metrics using a Prometheus client library.
In Go: `github.com/prometheus/client_golang`
In Python: `prometheus-fastapi-instrumentator` (auto-instruments FastAPI)

### What to Measure

```
For every service:
  http_requests_total{service, method, path, status}    ← request count
  http_request_duration_seconds{...}                    ← latency histogram
  active_connections                                    ← concurrent load

For Postgres:
  pg_stat_activity_count                                ← active queries
  pg_replication_lag                                    ← replica lag

For the queue:
  kafka_consumer_lag                                    ← how far behind consumers are

For the feed fanout worker:
  fanout_duration_seconds                               ← how long fanout takes
  fanout_followers_count                                ← how many writes per post
```

### Grafana Dashboards

Grafana connects to Prometheus and visualizes the metrics.
Pre-built dashboards exist for Postgres, Go services, Kafka — import them, don't build from scratch.

Key dashboards to have:
- Request rate + error rate + latency (RED method) per service
- Postgres connections + slow queries
- Queue consumer lag
- System resources (CPU, memory, disk)

---

## OpenTelemetry (Traces)

### The Problem Traces Solve

```
User complains: "feed is slow"

Without traces:
  → check Grafana: yes, feed is slow (800ms)
  → which service? which query? no idea
  → grep logs across 4 services manually

With traces:
  → open Jaeger/Tempo
  → find that user's request
  → see: gateway 5ms | user-svc 620ms | feed 175ms
  → open user-svc span: "GetFollowers" query took 615ms
  → add index to follows table → fixed
```

### How OpenTelemetry Works

```
Each service instruments its code with the OTel SDK.
Every request gets a trace_id (e.g. "abc-123").
Each service creates a span for its work and attaches the trace_id.
Spans are sent to a collector (OpenTelemetry Collector).
Collector forwards to Tempo (storage) → Grafana (visualization).
```

When gateway calls User service via gRPC, it passes the `trace_id` in metadata.
User service creates a child span. The whole chain is linked.

### Implementation Plan

**Step 1 (Now):** Add Prometheus metrics to each service.
**Step 2 (Before first load test):** Add OpenTelemetry tracing.
**Step 3 (After load test):** Use traces to find the bottleneck.

Don't add tracing after the load test — you won't be able to interpret the results without it.

### Libraries

Go services:
```go
import "go.opentelemetry.io/otel"
import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
// Wrap gRPC server/client with otelgrpc → auto-traces all gRPC calls
```

Python services (FastAPI):
```python
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
FastAPIInstrumentor.instrument_app(app)
# Auto-instruments all routes
```

---

## Stack Summary

```
Tool              Purpose                  Self-hosted   Cost
────────────────────────────────────────────────────────────
Prometheus        Metrics storage + query  Yes           Free
Grafana           Dashboards, alerting     Yes           Free
Loki              Log aggregation          Yes           Free (optional)
Tempo             Trace storage            Yes           Free
OpenTelemetry     Instrumentation SDK      Library       Free
  Collector       Receives + routes data   Yes           Free
```

Everything runs as Docker containers alongside your services.
Zero external services, zero cost, full control.

---

## When to Add Monitoring

**Not at the end — not as an afterthought.**

Add in this order:
1. **Prometheus + basic metrics** — when you build the first service
2. **Grafana dashboard** — when you have 2+ services
3. **OpenTelemetry tracing** — before the first load test
4. **Alerting** — after you know what "normal" looks like

The point: you want monitoring in place *during* development so you develop the habit
of looking at metrics when something is slow, not guessing.
