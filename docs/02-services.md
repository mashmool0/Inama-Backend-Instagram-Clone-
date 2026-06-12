# Services

## Decision Framework

Two axes drive language choice per service:

**Traffic pattern:**
- Synchronous (gRPC) — caller waits, latency matters directly
- Asynchronous (queue consumer) — no one waiting, throughput matters

**Hot path vs cold path:**
- Hot path = touched by every user request → Go
- Cold path = called only for specific actions → FastAPI fine

---

## Service Map

### 1. API Gateway
| | |
|---|---|
| **Language** | Go |
| **Protocol** | HTTP in, gRPC out |
| **Responsibilities** | Rate limiting, JWT verification, request routing |
| **Database** | None (Redis for rate limit counters) |
| **Why Go** | Every single request passes through here. Python's overhead would show on 100% of traffic. |

```
Overhead target: < 5ms added per request
Work: Redis lookup (rate limit) → JWT verify → proxy to service
```

---

### 2. Auth Service
| | |
|---|---|
| **Language** | FastAPI (Python) |
| **Protocol** | gRPC |
| **Responsibilities** | Register, login, OTP verify, password reset, JWT issue |
| **Database** | `auth_db` (Postgres) |
| **Why FastAPI** | Cold path (called rarely). Security-critical logic — use familiar language. bcrypt is CPU-bound by design so language speed doesn't help. |

```
Flows:
  register → validate → send OTP (SMS) → verify OTP → create user → issue JWT
  login    → verify credentials → issue JWT + refresh token
  reset    → send email → verify token → update password
```

---

### 3. User / Social Graph Service
| | |
|---|---|
| **Language** | Go |
| **Protocol** | gRPC |
| **Responsibilities** | Profile CRUD, follow/unfollow, followers/following lists |
| **Database** | `user_db` (Postgres) |
| **Why Go** | Feed fanout calls this on EVERY post to get follower IDs. Hot path from the fanout worker. |

```
Key tables:
  users(id, username, bio, avatar_url, created_at)
  follows(follower_id, followee_id, created_at)

Key queries:
  "who follows user X?" → WHERE followee_id = X   (fanout uses this)
  "who does user X follow?" → WHERE follower_id = X
```

---

### 4. Posts + Interactions Service
| | |
|---|---|
| **Language** | Go |
| **Protocol** | gRPC (reads/writes) + publishes to queue (post.created event) |
| **Responsibilities** | Create/delete/read posts, likes, comments |
| **Database** | `posts_db` (Postgres) |
| **Why Go** | Likes are the highest-frequency write in the system. Needs to handle atomic increments under concurrent load. |

```
Key tables:
  posts(id, author_id, caption, media_url, like_count, comment_count, created_at)
  likes(post_id, user_id, created_at)
  comments(id, post_id, author_id, body, created_at)

On post created → publish event to queue → fanout worker + search indexer consume it
Like count → atomic: UPDATE posts SET like_count = like_count + 1
```

---

### 5. Feed + Explore Service
| | |
|---|---|
| **Language** | Go |
| **Protocol** | gRPC (serve feed) + queue consumer (fanout worker) |
| **Responsibilities** | Serve home feed, serve explore page, run fanout worker |
| **Database** | `feed_db` (Postgres or Redis lists — TBD) |
| **Why Go** | Hottest read service. Fanout worker spawns concurrent goroutines per follower batch. |

```
Two internal components:
  1. Feed Reader (sync)  → serve paginated feed to user
  2. Fanout Worker (async) → consume post.created → write to each follower's feed

Feed strategy (current): fan-out-on-read + pagination
  → on read: query posts from followed users, paginated 20 at a time
  → known planned bottleneck — will migrate to fan-out-on-write after load test reveals it

Explore: one shared Redis cache of recent public posts, served to all users
```

---

### 6. Notifications Service
| | |
|---|---|
| **Language** | Go |
| **Protocol** | Queue consumer |
| **Responsibilities** | Consume like/comment/follow events, write to DB, push via FCM/APNs |
| **Database** | `notifications_db` (Postgres) |
| **Why Go** | Concurrent push delivery — goroutines per notification batch. |

```
Flow:
  event consumed (like / comment / follow)
  → write notification record to DB
  → call FCM / APNs async
  → user marks as read via gRPC call
```

---

### 7. Search Indexer
| | |
|---|---|
| **Language** | FastAPI (Python) |
| **Protocol** | Queue consumer (indexing) + gRPC (search queries) |
| **Responsibilities** | Index posts into Elasticsearch, serve search queries |
| **Database** | Elasticsearch (external) |
| **Why FastAPI** | Async worker — no one waiting. Python's `elasticsearch-py` client is mature and ergonomic. Complex indexing logic benefits from familiar language. |

```
Indexing flow:
  post.created event → extract hashtags → index into Elasticsearch

Search flow:
  search query (gRPC) → ES query → return results
  search types: by username, by hashtag, full-text caption
```

---

## Language Summary

```
Service               Language     Hot Path   Why
────────────────────────────────────────────────────────────
API Gateway           Go           Yes        Every request
Auth                  FastAPI      No         Complex logic, cold path
User / Social Graph   Go           Yes        Fanout queries this constantly
Posts + Interactions  Go           Yes        Likes = highest freq write
Feed + Explore        Go           Yes        Hottest read + fanout concurrency
Notifications         Go           No (async) Push concurrency with goroutines
Search Indexer        FastAPI      No (async) Best ES client, complex logic
```
