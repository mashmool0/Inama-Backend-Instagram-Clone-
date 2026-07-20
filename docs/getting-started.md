# Getting Started — New Developer Onboarding

Welcome. This guide takes you from a fresh machine to writing your first service.
Read it top to bottom once, then use it as a checklist. It should take ~30–45
minutes to get fully set up.

---

## 0. What you're building (2-minute version)

**Inama** is an Instagram-like social platform built as **7 microservices** in a
single mono-repo, deployed with Docker Compose on one VPS. Services talk to each
other two ways:

- **gRPC** for synchronous calls (caller waits for a result)
- **RabbitMQ events** for asynchronous work (fire-and-forget)

The frontend only ever talks to the **API Gateway**, which routes to the services.

Read these before you write code (in this order):

| Read | For |
|---|---|
| [../CLAUDE.md](../CLAUDE.md) | The locked decisions and guardrails — the rules of the repo |
| [01-architecture-overview.md](01-architecture-overview.md) | The big picture |
| [02-services.md](02-services.md) | What each service does + why each language |
| [04-communication.md](04-communication.md) | gRPC vs queue, the outbox pattern |
| [09-roadmap.md](09-roadmap.md) | The phase plan and who owns what |
| [../proto/README.md](../proto/README.md) | The contracts + how to generate code |
| [../proto/EVENTS.md](../proto/EVENTS.md) | The async event contracts |

**Don't write any code before the design is agreed.** Guardrail #1 in CLAUDE.md:
discuss first, implement second.

---

## 1. Install the tools

| Tool | Version | Why | Install |
|---|---|---|---|
| **Git** | any recent | version control | your OS package manager |
| **Docker + Docker Compose** | v2 | runs the whole stack | https://docs.docker.com/get-docker/ |
| **Go** | 1.23+ | your services (Gateway, User, Posts, Feed, Notifications) | https://go.dev/dl/ |
| **make** | any | runs `make gen`, etc. | usually preinstalled on Linux/macOS |
| **Python** | 3.12+ | only if you work on Auth or Search (FastAPI) | https://www.python.org/downloads/ |

You do **not** need to install `buf` or `protoc` by hand — `make gen` installs
`buf` for you (see step 3).

After installing Go, make sure Go's bin directory is on your `PATH` (this is
where `buf` lands):

```bash
# add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:$(go env GOPATH)/bin"
```

Verify:
```bash
git --version
docker --version && docker compose version
go version          # must be 1.23+
make --version
```

---

## 2. Get the code

```bash
git clone https://github.com/mashmool0/Inama-Backend-Instagram-Clone-.git
cd Inama-Backend-Instagram-Clone-
```

We use **feature branches** — never commit directly to `main`. Create a branch
for whatever you're doing:

```bash
git checkout main
git pull
git checkout -b feature/user-scaffold      # name it feature/<service>-<thing>
```

---

## 3. Generate the contract code (the important step)

The `.proto` files under `proto/` are just contracts — text. They must be turned
into real Go/Python code before you can import them. One command does it:

```bash
make gen
```

This will:
1. Install `buf` if it's missing (into `$(go env GOPATH)/bin`).
2. Generate Go + Python code into `proto/gen/`.
3. Run `go mod tidy` on the generated Go module.

The generated code is **committed to the repo** (it's the single source of
truth), so most of the time it will already be there after you clone. You only
re-run `make gen` when a `.proto` file changes.

> If someone changes a `.proto` and forgets to run `make gen`, CI fails with
> "proto/gen is out of date". The fix is always: `make gen`, then commit
> `proto/gen`.

Other proto commands:
```bash
make lint      # style-check the .proto files (run before pushing proto changes)
make format    # auto-format the .proto files
```

---

## 4. Run the whole stack and verify it works

```bash
docker compose up --build
```

This starts every service plus Postgres, Redis, Prometheus, and Grafana on one
shared Docker network. Give it a minute the first time (it builds images).

In another terminal, hit the health endpoints:

```bash
curl localhost:8080/health     # gateway  -> {"status":"ok","service":"gateway"}
curl localhost:8001/health     # auth
curl localhost:8082/health     # user
curl localhost:8083/health     # posts
curl localhost:8084/health     # feed
curl localhost:8085/health     # notifications
curl localhost:8086/health     # search
```

Other things you can open in a browser:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (login `admin` / `admin`)

Stop everything with `Ctrl+C`, or `docker compose down`.

---

## 5. Run just YOUR service (you don't need the whole stack)

You can run a single service on its own — no need to boot everything:

```bash
docker compose up --build user            # just the user service
curl localhost:8082/health
```

Or a service plus only the infra it needs:

```bash
docker compose up --build user postgres redis
```

Services are intentionally **not** wired to depend on each other yet, so any one
runs alone. This is how you'll work day-to-day on your own service.

Port map (host → service):

| Service | Host port |
|---|---|
| gateway | 8080 |
| auth | 8001 |
| user | 8082 |
| posts | 8083 |
| feed | 8084 |
| notifications | 8085 |
| search | 8086 |
| postgres | 5432 |
| redis | 6379 |
| prometheus | 9090 |
| grafana | 3000 |

---

## 6. The git workflow (how we avoid stepping on each other)

```
1. Branch:   git checkout -b feature/<service>-<thing>
2. Work:     stay inside YOUR service folder (services/<yours>/)
3. Commit:   small commits, one job each (see commit rules below)
4. Push:     git push -u origin feature/<service>-<thing>
5. PR:       open a Pull Request on GitHub -> the other dev reviews -> merge
```

Rules:
- **No direct push to `main`.** Everything goes through a PR with 1 review.
- **Own your services.** Edit your own service folders. Don't edit the other
  dev's service without asking.
- **Shared files need a heads-up.** `proto/`, `docker-compose.yml`,
  `infra/postgres/init.sql`, `libs/`, `buf.*` affect both of us — tell the other
  dev before changing them.

### Commit rules (important, project-specific)

- **One job per commit.** Don't lump unrelated changes together. If you added a
  schema and a repository, that's two commits.
- **Do NOT add a "Co-Authored-By" line** to commits.

Good:
```
add user_db schema and follows table migration
add FollowRepository with GetFollowers query
```

---

## 7. Conventions you MUST follow

These are baked into the contracts and the whole codebase — don't deviate:

- **IDs are UUID strings** everywhere (protos and DB columns).
- **Caller identity comes from gRPC metadata** (`x-user-id`), set by the Gateway.
  Your handlers read the current user from the request context/metadata — never
  from a field the client sends.
- **Timestamps** use `google.protobuf.Timestamp`.
- **Pagination** is cursor-based: request carries `limit` + opaque `cursor`;
  response returns `next_cursor` (empty = no more). No page numbers.
- **Each service owns its own database.** No cross-service SQL, ever. If you need
  another service's data, call it over gRPC or consume its event.
- **Logging** is structured JSON.

---

## 8. The internal pattern every service follows

```
gRPC Server / Queue Consumer
        ↓
    Handler        ← translates proto <-> service types. Thin. No business logic.
        ↓
    Service        ← business logic. Unit-tested with a mocked repository.
        ↓
  Repository       ← database access ONLY. No business rules here.
```

Rules: a handler never touches the DB directly; a service never calls another
service's repository. Keep the layers clean — it's what makes the code testable.

---

## 9. Your first task — Phase 1: the User / Social Graph service (Go)

Phase 1 is the cleanest parallel phase: one dev builds **Auth** (FastAPI), the
other builds **User** (Go). They don't depend on each other, so you can't block
one another.

Build the User service in this order (from [09-roadmap.md](09-roadmap.md)):

```
1. Schema + migrations   → user_db: users, follows (with composite indexes)
2. Repository layer       → UserRepository, FollowRepository
                             FollowRepository.GetFollowers(user_id) — Feed depends
                             on this next phase, so get it right
3. Service layer          → ProfileService, FollowService
                             (duplicate-follow guard, follower/following counters)
4. Handler layer          → GetProfile, UpdateProfile, Follow, Unfollow,
                             GetFollowers, GetFollowing  (thin, proto <-> service)
5. Wire the gRPC server   → expose user.proto, register handlers
6. Metrics + logging      → /metrics, structured logs, /health
7. Dockerfile             → already scaffolded; extend as needed
8. Integration test       → follow flow + GetFollowers against a real user_db
9. Gateway routes         → add /users/* routes to the gateway
```

Your contract is already written: [../proto/user/user.proto](../proto/user/user.proto).
Read it first — it defines exactly the methods and messages you implement.

**Definition of done (Phase 1, your half):**
- [ ] Profile view/edit works with a real JWT
- [ ] Follow/unfollow works, follower counts are correct
- [ ] `GetFollowers` returns correct IDs (Feed will depend on this next phase)
- [ ] Unit + integration tests pass in CI
- [ ] Service exposes `/metrics`

---

## 10. Using the generated proto code in a Go service

Add the shared generated module to your service's `go.mod`:

```
require github.com/mashmool0/inama/proto/gen v0.0.0
replace github.com/mashmool0/inama/proto/gen => ../../proto/gen
```

Then import and implement the generated server interface:

```go
import userpb "github.com/mashmool0/inama/proto/gen/user"

// The generated code gives you userpb.UserServiceServer — an interface with one
// method per rpc in user.proto. You implement it in your handler layer.
type handler struct {
    userpb.UnimplementedUserServiceServer
    svc *ProfileService
}

func (h *handler) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.Profile, error) {
    // read caller from metadata if needed, call h.svc, map result -> *userpb.Profile
}
```

> Note: Docker builds currently use each service's own folder as the build
> context, so a container can't yet see `../../proto/gen`. That's fine while
> services only serve `/health`. When you start importing `proto/gen`, the build
> context will be switched to the repo root — flag it and it'll be set up.

---

## 11. Quick reference — where things live

```
services/<name>/     your service code (main.go / main.py, Dockerfile)
proto/<name>/        the .proto contract for that service
proto/gen/           generated Go + Python code (do not hand-edit)
proto/EVENTS.md      async event contracts (RabbitMQ)
infra/               postgres init.sql, prometheus config, etc.
libs/                shared conventions (logging, config, metadata interceptor)
docker-compose.yml   the whole stack
Makefile             make gen / make lint / make format
docs/                architecture + this guide
```

## If you get stuck

1. Re-read the relevant `docs/` file — the answer is usually there.
2. Check the `.proto` for the exact contract.
3. For anything touching `proto/`, shared infra, or a locked decision — ping the
   other dev before changing it.

Welcome aboard. Start with steps 1–4, confirm the stack runs and `/health`
responds, then read your `user.proto` and begin the schema. 🚀
