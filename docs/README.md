# Inama — Project Documentation

Instagram-like social platform. University learning project.
Goal: deep understanding of microservice architecture through real implementation.

---

## Documents

| File | What it covers |
|---|---|
| [01-architecture-overview.md](01-architecture-overview.md) | Big picture, service map, C4 thinking |
| [02-services.md](02-services.md) | Each service, language, database, responsibilities |
| [03-database-strategy.md](03-database-strategy.md) | Per-service DBs, primary + replicas, read/write split |
| [04-communication.md](04-communication.md) | gRPC (sync) vs queue (async) — which calls are which |
| [05-feed-strategy.md](05-feed-strategy.md) | Fanout strategy, the feed table, celebrity problem |
| [06-infrastructure.md](06-infrastructure.md) | Mono-repo, Docker, VPS deployment, CI/CD |
| [07-monitoring.md](07-monitoring.md) | Prometheus, Grafana, OpenTelemetry tracing |
| [08-load-testing.md](08-load-testing.md) | Target: 1k req/sec, tools, methodology |

---

## Non-Functional Targets

- **Test target:** ~1,000 req/sec sustained, p95 < 500ms (on 1–2 VPS)
- **Design target:** reason about 10,000 req/sec (architectural thinking)
- **Background jobs:** must not be lost if a worker crashes
- **Deployment:** single VPS, Docker Compose, automated via GitHub Actions
