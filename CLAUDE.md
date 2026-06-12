# Inama — Project Rules

## Locked Decisions — Never Change Without Discussion

| Concern | Decision |
|---|---|
| Architecture | Microservices, mono-repo |
| Repo | One repo, `services/` per service |
| Deployment | Docker Compose, single VPS, GitHub Actions CI/CD |
| Databases | Per-service (auth_db, user_db, posts_db, feed_db, notif_db) on one Postgres instance |
| DB topology | 1 primary + 2 read replicas |
| Sync calls | gRPC only |
| Async events | Message queue (Kafka or RabbitMQ — **TBD**) |
| Languages | Go: Gateway, User, Posts, Feed, Notifications — FastAPI: Auth, Search |
| Feed (now) | Fan-out-on-read + cursor pagination — planned bottleneck, migrate after load test |
| Feed (target) | Fan-out-on-write with fanout worker |
| Load target | Test: 1k req/sec, p95 < 500ms — Design: 10k req/sec |
| Monitoring | Prometheus + Grafana + OpenTelemetry (tracing before first load test) |

## Open Decisions — Must Discuss Before Implementing

- Kafka vs RabbitMQ (message broker choice)
- feed_db storage: Postgres feed_items table vs Redis lists

## Current Build Phase

**Phase 0 — Foundation** (not started)
Mono-repo scaffold → Docker Compose → CI/CD → shared proto → one working gRPC call end-to-end.
Full phase plan: [docs/](docs/)

## Guardrails — Check These Before Every Action

1. **Discuss before implement.** No code without explicit agreement on design first.
2. **No new libraries or tools** without discussion — not even "small" ones.
3. **Service boundary.** Each service reads/writes only its own database. No cross-service SQL. Ever.
4. **Communication boundary.** Caller waits → gRPC. Fire and forget → queue. Never mix for the same flow.
5. **Check open decisions.** If the task touches Kafka/RabbitMQ or feed_db — flag that it's TBD first.
6. **Locked decision conflict.** If a request contradicts anything in the Locked Decisions table above → correct it before proceeding, not after.
7. **Docs stay current.** If a decision changes mid-conversation → update the relevant file in `docs/`.
8. **Phase check.** If asked to implement something from Phase 2+ while Phase 0 is incomplete → flag the dependency.

## Internal Component Pattern (All Services)

```
gRPC Server / Queue Consumer
        ↓
    Handler
        ↓
    Service (business logic)
        ↓
  Repository (DB access only)
```

No handler touches a DB directly. No service calls another service's repository.

## Architecture Docs Reference

| Question | Read |
|---|---|
| Which service does what + language | [docs/02-services.md](docs/02-services.md) |
| Database ownership + denormalization | [docs/03-database-strategy.md](docs/03-database-strategy.md) |
| gRPC vs queue — full call map | [docs/04-communication.md](docs/04-communication.md) |
| Feed fanout strategy | [docs/05-feed-strategy.md](docs/05-feed-strategy.md) |
| Mono-repo layout + CI/CD | [docs/06-infrastructure.md](docs/06-infrastructure.md) |
| Monitoring stack | [docs/07-monitoring.md](docs/07-monitoring.md) |
| Load test targets + tools | [docs/08-load-testing.md](docs/08-load-testing.md) |

Read the relevant doc before answering architecture questions — do not answer from memory alone.
