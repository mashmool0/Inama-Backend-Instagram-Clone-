# Plan — Auth Service, Step 1: Foundation (schema + models + first migration)

## Status: ✅ DONE (branch `feature/auth-scaffold`)

Implemented and committed. Two small deviations from the original plan:
- **`config.py` uses `os.environ` directly** (not `inama_libs.config`) to keep
  Step 1 self-contained; it will converge with the shared helpers when the
  server is wired at Step 5 (where the Docker build context moves to repo root).
- **Branched off `feature/phase0-scaffold`** (not `main`), because Phase 0 isn't
  merged yet and Dev B's branches are already based on it too.

---

## Context

The Inama backend is in Phase 1. Dev A owns the **Auth** service (FastAPI/Python,
gRPC protocol, the most security-critical service). All architecture decisions
are locked and approved by the user:

- **Server:** `grpc.aio` (async)
- **ORM:** SQLAlchemy 2.0 async, driver `psycopg` v3
- **Migrations:** Alembic
- **Security stack:** `bcrypt` (password hash), `PyJWT` + RS256 (tokens),
  `redis` (OTP with TTL)
- **IDs:** UUID strings everywhere; **tokens stored hashed**, never raw
- **auth_db schema** (approved): `users`, `refresh_tokens`, `password_reset_tokens`.
  OTP lives in Redis (TTL), **no `otp_codes` table**.

`auth_db` already exists (created in `infra/postgres/init.sql`). The current
`services/auth/` is only a FastAPI `/health` stub. This step lays the database
foundation for the service — **no business logic yet** (that's Steps 2+).

This is Step 1 of the agreed build order:
`Step 1 schema+models+migration → 2 repositories → 3 services → 4 gRPC handlers →
5 async server wiring → 6 SMS/email mock clients → 7 tests → 8 gateway routes`.

## Approach

Restructure `services/auth/` into an `app/` package, swap the stub's deps for the
async stack, define the three tables as SQLAlchemy models, and add an Alembic
setup with a hand-written initial migration (Alembic can't be run in this
environment, so the migration is authored to match the models exactly).

### Files created / modified

**Modified** `services/auth/requirements.txt` — replaced fastapi/uvicorn with:
```
grpcio
grpcio-tools
sqlalchemy[asyncio]
psycopg[binary]
alembic
bcrypt
pyjwt[crypto]
redis
```
(FastAPI/uvicorn dropped — Auth is a gRPC server, not an HTTP API. A tiny
health/metrics HTTP endpoint is added at Step 5.)

**Created** `services/auth/app/__init__.py` — package marker.

**Created** `services/auth/app/config.py` — settings from env vars:
`DATABASE_URL`, `REDIS_URL`, JWT key paths + TTLs, OTP TTL. Local defaults so
`alembic upgrade head` works out of the box. Only `DATABASE_URL` matters for
this step.

**Created** `services/auth/app/db.py` — async SQLAlchemy setup:
- `Base` = `DeclarativeBase` subclass
- async engine from `DATABASE_URL` (`postgresql+psycopg://...`)
- `async_sessionmaker` + an `async` session provider for repositories to use

**Created** `services/auth/app/models.py` — three SQLAlchemy 2.0 typed models
(`Mapped[...]` / `mapped_column`), matching the approved schema:
- `User`: `id` UUID PK (default `uuid4`), `phone` unique not-null,
  `password_hash`, `is_verified` (default false), `created_at`/`updated_at`
  (server_default `now()`)
- `RefreshToken`: `id` UUID PK, `user_id` FK→users.id, `token_hash` not-null,
  `expires_at`, `revoked` (default false), `created_at`; index on
  `token_hash` and `user_id`
- `PasswordResetToken`: same shape as RefreshToken but with `used` (default
  false) instead of `revoked`

**Created** Alembic setup:
- `services/auth/alembic.ini` — points script location at `migrations/`, reads
  the DB URL from env (not hardcoded)
- `services/auth/migrations/env.py` — **async** env: builds an async engine,
  runs migrations via `connection.run_sync(...)`, uses `app.db.Base.metadata`
  as `target_metadata`
- `services/auth/migrations/script.py.mako` — standard Alembic template
- `services/auth/migrations/versions/0001_initial.py` — hand-written migration
  that creates the three tables (columns, PK, FK, unique/index constraints)
  exactly matching `models.py`

### Patterns to reuse
- `libs/python/inama_libs/config.py` — env-var config helpers (adopt at Step 5)
- UUID-string + `google.protobuf.Timestamp` conventions already defined in
  `proto/auth/auth.proto` — the models mirror these field shapes

### Not in this step (deliberately deferred)
- gRPC handlers / business logic (Steps 2–4)
- RS256 key generation + JWT logic (Step 3)
- The async gRPC server + health/metrics endpoint + `main.py` rewrite (Step 5)
- Switching the Docker build context to repo root so the container can reach
  `proto/gen` and `libs` (done when the server starts importing them)
- `make gen` (prerequisite for Step 4, not Step 1)

## Verification

Environment here has no Postgres/pip deps, so full run happens on your machine:

1. `cd services/auth && pip install -r requirements.txt`
2. Start Postgres: `docker compose up -d postgres` (creates `auth_db`)
3. `export DATABASE_URL=postgresql+psycopg://inama:inama@localhost:5432/auth_db`
4. `alembic upgrade head` → should create the three tables with no error
5. Inspect: `psql "$DATABASE_URL" -c '\dt'` shows `users`, `refresh_tokens`,
   `password_reset_tokens`; `\d users` shows the expected columns/constraints
6. `alembic downgrade base` then `upgrade head` again → confirms the migration is
   reversible and repeatable
7. `python -c "import app.models, app.db, app.config"` → imports cleanly

Done in this session (no DB): `python -m py_compile` on all new files passed, and
`0001_initial.py` was reviewed column-for-column against `models.py`.

## Branch & commit plan

Branched `feature/auth-scaffold` off `feature/phase0-scaffold`.

Commits (one job per commit, no Co-Authored-By):
1. `switch auth to async grpc and sqlalchemy dependency stack` (requirements.txt)
2. `add auth_db sqlalchemy models and async db setup` (app/)
3. `add alembic setup and initial auth_db migration` (alembic.ini, migrations/)
