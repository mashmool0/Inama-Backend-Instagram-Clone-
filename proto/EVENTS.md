# Async Event Contracts (RabbitMQ)

These are contracts too. Changing a payload = changing an interface both devs
depend on — agree together, same as a `.proto` change.

- **Serialization:** JSON (readable, easy to debug; revisit if throughput demands it)
- **Broker:** RabbitMQ
- **Topology:** one topic exchange `inama.events`. Publishers publish with the
  event name as the routing key. Each consumer declares its **own** queue and
  binds it to the routing keys it cares about — that's how one event fans out
  to several services.

```
                          ┌──────────────────────────┐
post.created ──publish──▶ │  exchange: inama.events   │
                          │        (topic)            │
                          └──────────────────────────┘
                             │route: post.created
              ┌──────────────┼───────────────┐
              ▼              ▼                ▼
        feed.q          search.q         notif.q
      (fanout worker)   (indexer)      (notifications)
```

## Envelope

Every event shares this envelope; `data` holds the event-specific payload.

```json
{
  "event_id": "uuid",           // unique per event, for dedupe/idempotency
  "event_type": "post.created",
  "occurred_at": "2026-07-11T12:00:00Z",  // RFC3339 UTC
  "data": { }
}
```

## Events

### `post.created` — publisher: Posts → consumers: Feed, Search, Notifications
```json
{
  "post_id": "uuid",
  "author_id": "uuid",
  "caption": "text with #hashtags",
  "media_url": "https://... or empty",
  "created_at": "2026-07-11T12:00:00Z"
}
```

### `post.liked` — publisher: Posts → consumer: Notifications
```json
{
  "post_id": "uuid",
  "post_author_id": "uuid",   // notification recipient
  "actor_id": "uuid"          // who liked
}
```

### `comment.created` — publisher: Posts → consumer: Notifications
```json
{
  "comment_id": "uuid",
  "post_id": "uuid",
  "post_author_id": "uuid",   // notification recipient
  "actor_id": "uuid"          // who commented
}
```

### `user.followed` — publisher: User → consumers: Notifications, Feed
```json
{
  "follower_id": "uuid",
  "followee_id": "uuid"       // notification recipient
}
```

### `user.updated` — publisher: User → consumer: Posts (denorm sync)
```json
{
  "user_id": "uuid",
  "username": "new_name",
  "avatar_url": "https://..."
}
```

## Rules

1. **Idempotent consumers.** Events can be delivered more than once. Dedupe on
   `event_id` (or a natural key). Never assume exactly-once delivery.
2. **Publish via the Outbox pattern** (see docs/04-communication.md) — the event
   row is written in the same DB transaction as the state change, then a worker
   publishes it. No event is lost if RabbitMQ is briefly down.
3. **Additive changes only** without a heads-up. Adding a field is safe; removing
   or renaming one breaks a consumer — coordinate first.
