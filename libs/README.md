# libs/

Shared conventions both developers agree on once, so all services look the same
(CLAUDE.md / roadmap step 0.5). Not business logic — just the common skeleton:

- Structured logging format (JSON, with a reserved `trace_id` field)
- Config loading (env vars → typed config)
- Error handling pattern (gRPC status codes, error wrapping)
- The `Handler → Service → Repository` template per language
- Health check convention (`/health`)
- Metrics convention (`/metrics`)

Fill in during Phase 0.
