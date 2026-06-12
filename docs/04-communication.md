# Service Communication

## Two Types of Communication

```
SYNCHRONOUS (gRPC)               ASYNCHRONOUS (Message Queue)
──────────────────               ────────────────────────────
Caller waits for response        Fire and forget — no waiting
Used when: need the result NOW   Used when: "do this eventually"
Latency matters directly         Throughput matters, not latency
Tight coupling in time           Decoupled — producer doesn't know consumers
```

The rule: **if the caller needs the result to continue → gRPC.
If the caller just needs to notify that something happened → queue.**

---

## Why gRPC (not REST) for Internal Calls

```
REST (JSON over HTTP)             gRPC (Protobuf over HTTP/2)
─────────────────────             ───────────────────────────
Human-readable JSON               Binary encoding — 5-10x smaller payload
No contract enforcement           .proto file = strict contract
Manual type mapping               Generated client + server code
One request per connection        Multiplexed over one HTTP/2 connection
Good for: public APIs             Good for: internal service calls
```

For internal service-to-service calls, gRPC wins on every axis.
The `.proto` file is the contract — changing it is a deliberate act, not an accident.

Generated code example (one `.proto` → client code for Go AND Python):
```proto
service UserService {
  rpc GetFollowers(GetFollowersRequest) returns (GetFollowersResponse);
  rpc GetProfile(GetProfileRequest) returns (ProfileResponse);
}

message GetFollowersRequest {
  int64 user_id = 1;
  int32 limit   = 2;
  int32 offset  = 3;
}
```

---

## Sync vs Async — Full Map

### Synchronous (gRPC)

| Caller | Called | Call | Why sync |
|---|---|---|---|
| API Gateway | Auth | VerifyToken | Must verify before routing |
| API Gateway | Any service | Route request | Must wait for response to return |
| Feed service | User service | GetFollowers | Fanout needs follower IDs to proceed |
| Posts service | User service | GetProfile | Need author details to return post |
| Client (via gateway) | Auth | Login/Register | User waits for JWT |
| Client (via gateway) | Posts | CreatePost | User waits for confirmation |
| Client (via gateway) | Feed | GetFeed | User waits for feed |
| Client (via gateway) | Search | Search | User waits for results |

### Asynchronous (Queue Events)

| Publisher | Event | Consumers | Why async |
|---|---|---|---|
| Posts service | `post.created` | Feed (fanout), Search (index), Notifications | Don't block "create post" on all downstream work |
| Posts service | `post.liked` | Notifications | Like succeeds immediately; notification follows |
| Posts service | `comment.created` | Notifications | Same |
| User service | `user.followed` | Notifications, Feed | Follow succeeds; notification follows |
| User service | `user.updated` | Posts (denorm sync) | Username change propagates eventually |

---

## The Dual-Write Problem

When a post is created, two things must happen:
1. Write to `posts_db`
2. Publish `post.created` event to queue

If you do them independently:
```
write to DB  ✓
publish to queue  ✗ (queue down)
→ post exists in DB but fanout never ran — some followers never see it
```

**Solution: Outbox Pattern**

```
Transaction:
  INSERT INTO posts (...)       ← the post
  INSERT INTO outbox (event)    ← the event, in the SAME transaction

Outbox worker (separate process):
  → reads unpublished outbox rows
  → publishes to queue
  → marks as published

Now DB write and event publish are atomic.
If queue is down: outbox fills up, publishes when queue recovers.
No event is lost.
```

This is a known hard problem in microservices. The outbox pattern is the standard solution.
CDC (Change Data Capture) is the alternative — watch the DB's WAL log and publish from there.
Outbox is simpler to start; CDC is more robust at scale.

---

## Message Queue Choice: Kafka vs RabbitMQ

**Decision: TBD — discuss before implementation.**

Quick comparison for this project's needs:

| | Kafka | RabbitMQ |
|---|---|---|
| Model | Append-only log, consumers track offset | Message broker, push to consumers, ack/nack |
| Replay | Yes — re-read old events | No — consumed = gone |
| Multiple consumers of same event | Yes, naturally | Yes, with fanout exchange |
| Operational complexity | Higher | Lower |
| Good for | Event sourcing, audit log, replay | Task queues, reliable job delivery |
| Learning value | High (offset model is different) | High (exchange model is different) |

For Inama: `post.created` needs to fan out to Feed + Search + Notifications simultaneously.
Both handle this; the mental models differ. Worth discussing explicitly.
