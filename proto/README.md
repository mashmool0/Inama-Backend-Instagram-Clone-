# proto/

The **contracts**. Every service's gRPC interface lives here as a `.proto` file,
one folder per service.

> ⚠️ SHARED FILE ZONE. Changing anything here affects both developers.
> Agree on a proto change together before committing it (CLAUDE.md / roadmap rule 1).

## Layout

```
proto/
  auth/auth.proto      — Register, VerifyOTP, Login, RefreshToken, password reset
  user/user.proto      — GetProfile, UpdateProfile, Follow, Unfollow, GetFollowers
  posts/posts.proto    — CreatePost, DeletePost, GetPost, Like/Unlike, comments
  feed/feed.proto      — GetFeed, GetExplore  (imports posts/posts.proto)
  notif/notif.proto    — GetNotifications, MarkAsRead
  search/search.proto  — Search
  gen/                 — GENERATED code (Go + Python), committed. Do not edit by hand.
```

Async event contracts (RabbitMQ, JSON) live in [EVENTS.md](EVENTS.md).

## Generating code

From `Inama_backend/`:

```bash
make gen      # installs buf if missing, generates proto/gen/, runs go mod tidy
make lint     # style-check the .proto files
make format   # auto-format the .proto files
```

`make gen` is the only step that turns these contracts into usable code. Run it
after any `.proto` change and **commit the result** — `proto/gen/` is checked in.

## How a service consumes the generated code

**Go service** — in the service's `go.mod`:
```
require github.com/mashmool0/inama/proto/gen v0.0.0
replace github.com/mashmool0/inama/proto/gen => ../../proto/gen
```
then:
```go
import postspb "github.com/mashmool0/inama/proto/gen/posts"
```

**Python service** (auth, search) — the generated `*_pb2.py` / `*_pb2_grpc.py`
live under `proto/gen/<service>/`. Add `proto/gen` to `PYTHONPATH` (wired in
each Python service's Dockerfile when it starts using gRPC).

## Conventions baked into every proto

- IDs are **UUID strings**
- Time is `google.protobuf.Timestamp`
- Pagination is an opaque `cursor` string + `limit`; response returns `next_cursor` (empty = end)
- Caller identity travels in **gRPC metadata** (`x-user-id`), never in a message field
  (except pre-login Auth RPCs)
