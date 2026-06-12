# Database Strategy

## Core Rule: Database-Per-Service

Every service owns its data. No other service touches its tables directly.
This is enforced in code (connection strings), not hardware.

```
auth_db      ← only Auth service has the connection string
user_db      ← only User service has the connection string
posts_db     ← only Posts service has the connection string
feed_db      ← only Feed service has the connection string
notif_db     ← only Notifications service has the connection string
```

All five databases run on **one Postgres instance** on the VPS.
Physical separation happens later if one database outgrows the machine.

---

## Why Not a Shared Database

A shared database couples every service through the schema.
What looks like microservices becomes a distributed monolith:
- Schema change in one service → breaks another
- No service can be deployed independently
- Cross-service JOINs make refactoring impossible
- All the network complexity of microservices, none of the independence

The per-service rule forces you to learn:
- **Denormalization** — store `username` copies in posts_db
- **Event-driven sync** — keep those copies up to date via events
- **No cross-service JOINs** — design data ownership carefully

These are the core microservice lessons. A shared DB hides all of them.

---

## Primary + Replicas

```
         Writes
           │
     ┌─────▼──────┐
     │   PRIMARY  │  ← all writes go here
     └─────┬──────┘
           │ WAL streaming replication
     ┌─────┴──────┐
     │  REPLICA 1 │  ← read queries
     └────────────┘
     │  REPLICA 2 │  ← read queries
     └────────────┘
```

**Writes → Primary always.**
**Reads → Replicas** (with one exception below).

### Replica Lag and Read-Your-Own-Writes

Replicas lag behind primary by milliseconds to seconds under load.
Globally acceptable — eventual consistency is fine for "see someone else's post."

One case it's NOT acceptable: a user posts, then views their own profile, and
their post is "missing" because the replica hasn't caught up yet.

Fix (cheap, no architecture change): for a short window after a user writes,
route **that user's own reads** to the primary. Everyone else reads from replicas.
Implementation: store `last_write_ts` in JWT or Redis, compare on read.

### Which databases need replicas?

Start with replicas on `posts_db` and `user_db` — highest read volume.
`auth_db` and `notif_db` can run on primary-only initially.

---

## The Microservice Tax on Joins

In a monolith you'd write:
```sql
SELECT posts.*, users.username, users.avatar_url
FROM posts
JOIN users ON posts.author_id = users.id
WHERE posts.id = $1
```

With database-per-service, `posts_db` and `user_db` are separate.
That JOIN is impossible. Your two options:

**Option A — Service call (N+1 risk):**
```
Posts service fetches post → calls User service for author details
Problem: loading a feed of 20 posts = 20 separate gRPC calls to User service
```

**Option B — Denormalization (correct approach):**
```
Store author_username and author_avatar_url directly in posts_db.posts
When a user changes their username → publish user.updated event
→ Posts service consumes it → updates its local copy

One query, no cross-service calls.
Cost: data duplication + sync complexity.
```

Option B is the real microservice pattern. Option A is the N+1 problem across a network —
worse than N+1 in a monolith because each call has network latency.

---

## Database Assignments

| Service | Database | Key Tables |
|---|---|---|
| Auth | `auth_db` | users, otp_codes, refresh_tokens, password_reset_tokens |
| User | `user_db` | users, follows |
| Posts + Interactions | `posts_db` | posts, likes, comments |
| Feed + Explore | `feed_db` | feed_items (per-user feed rows, when fan-out-on-write is adopted) |
| Notifications | `notif_db` | notifications |
| Search | Elasticsearch | posts index (separate from Postgres entirely) |
