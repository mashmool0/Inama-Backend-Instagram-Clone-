# libs/

Shared conventions so every service looks and behaves the same. Not business
logic — the common skeleton (CLAUDE.md / roadmap step 0.5).

```
libs/
  go/       Go module  github.com/mashmool0/inama/libs   (gateway, user, posts, feed, notifications)
  python/   Package     inama_libs                        (auth, search)
```

Both provide the same three things:

| Concern | Go | Python |
|---|---|---|
| Structured JSON logging | `logging.New("user")` | `inama_libs.logging.new("auth")` |
| Env-var config | `config.LoadBase("user")` | `inama_libs.config.load_base("auth")` |
| Identity (`x-user-id`) | `identity.*` | `inama_libs.identity.*` |

## Identity — the important one

The verified user id travels in gRPC metadata under `x-user-id`. Never read
identity from a client-supplied request field — always from here.

**Go — service side:**
```go
server := grpc.NewServer(grpc.UnaryInterceptor(identity.UnaryServerInterceptor()))
// in a handler:
userID, err := identity.RequireUserID(ctx)
```
**Go — caller side (Gateway):**
```go
ctx = identity.WithUserID(ctx, verifiedUserID)
resp, err := postsClient.CreatePost(ctx, req)
```

**Python — service side:**
```python
from inama_libs import identity
user_id = identity.require_user_id(context)   # aborts UNAUTHENTICATED if missing
```
**Python — caller side:**
```python
stub.CreatePost(req, metadata=identity.with_user_id(user_id))
```

## How a service depends on libs

**Go** — in the service's `go.mod`:
```
require github.com/mashmool0/inama/libs v0.0.0
replace github.com/mashmool0/inama/libs => ../../libs/go
```
`logging` and `config` are pure stdlib; `identity` pulls in grpc (resolved by
`go mod tidy` the first time you build a service that imports it).

**Python** — in the service's `requirements.txt` or Dockerfile:
```
pip install -e ../../libs/python
```
(That switches to a repo-root Docker build context — set up when the Python
services start using gRPC, same as for proto/gen.)
