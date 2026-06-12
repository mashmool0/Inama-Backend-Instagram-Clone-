# Feed Strategy

## The Core Problem

A user opens their home feed. They follow 300 people.
How do you efficiently show them the latest 20 posts from those 300 people?

This is harder than it looks. Two fundamentally different approaches exist.

---

## Fan-Out-On-Read (Current Plan)

On every feed request, compute the feed live:

```
User requests feed
  → "give me posts from everyone I follow"
  → SELECT follower list (300 users)
  → SELECT recent posts WHERE author_id IN (300 ids)
  → sort by created_at DESC
  → return page 1 (20 posts)
```

**What pagination actually does:**
```
"give me 20" is the LAST step.
The DB still scans posts from all 300 followees to find the newest 20.
Deep scroll (page 30) = scan further back through all 300 followees.
```

| ✅ Advantages | ❌ Disadvantages |
|---|---|
| 1 write per post (simple) | Expensive read — grows with followee count |
| No feed storage needed | Repeated work on every scroll |
| Always fresh (no lag) | Deep scroll gets slow |
| Good for low follower counts | Bad at scale with many followees |

**Why start here:** it will break under load test. That breakage is the lesson.
When the feed query shows up as the bottleneck in profiling → migrate to fan-out-on-write.

---

## Fan-Out-On-Write (Planned Migration After Load Test)

On every post creation, pre-compute the feed for every follower:

```
User creates post
  → post.created event published
  → Fanout Worker consumes it
  → fetch all follower IDs (e.g. 500 followers)
  → INSERT into feed_items for each follower:
       (follower_id, post_id, author_id, created_at)
  → done

User requests feed
  → SELECT * FROM feed_items WHERE user_id = me
      ORDER BY created_at DESC LIMIT 20
  → one simple indexed query
```

| ✅ Advantages | ❌ Disadvantages |
|---|---|
| Feed read = one simple query | Write amplification: 1 post = 500 DB writes |
| Constant read cost regardless of followees | Feed storage grows (one row per follower per post) |
| Scales to heavy read load | Stale if worker is slow (lag) |
| Precomputed = fast | Celebrity problem (see below) |

---

## The Celebrity Problem

Fan-out-on-write breaks for accounts with millions of followers.
One post = 1,000,000 INSERT operations.
The fanout worker would take hours. The queue would back up.

**Standard solution: hybrid fan-out**
```
Normal users (< 10,000 followers)  → fan-out-on-write (fast, precomputed)
Celebrity users (> 10,000)         → fan-out-on-read for their posts only

Feed read:
  → read precomputed feed (fan-out-on-write entries)
  → MERGE with: latest posts from any celebrities you follow (live query)
  → sort combined list → return 20

Instagram's actual approach. Twitter used a similar hybrid.
```

For Inama: not needed at current scale. Worth knowing it exists.
Build it after fan-out-on-write is working and load-tested.

---

## Feed Table Structure (Fan-Out-On-Write)

```sql
CREATE TABLE feed_items (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,      -- whose feed this entry belongs to
    post_id     BIGINT NOT NULL,
    author_id   BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL, -- when the POST was created (not the feed entry)
    inserted_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_feed_user_created ON feed_items (user_id, created_at DESC);
-- This makes "give me user X's feed, newest first, paginated" fast.
```

The `created_at` is the post's timestamp (not insertion time) so the feed is sorted
by when content was posted, not when the fanout ran.

---

## Explore (Simple Version)

Explore is not a per-user feed — it's a shared pool of recent public content.

```
Implementation:
  → Cache a list of recent public post IDs in Redis (TTL: 5 minutes)
  → On explore request: return random 20 from that cached list
  → Refresh cache every 5 minutes from posts_db

Cost: near zero. One Redis read, shared by all users.
No algorithm, no ML, no per-user state. Simple and fast.
```

Upgrade path (not now): personalization, engagement-based ranking, ML models.
For learning purposes, the simple version is correct to start.

---

## Migration Plan

```
Phase 1 (now):     Fan-out-on-read + pagination
                   Build it, run load test, watch it break

Phase 2 (after):   Fan-out-on-write
                   Fanout worker, feed_items table, one-query reads

Phase 3 (later):   Celebrity hybrid
                   Flag accounts above threshold, merge on read

Phase 4 (optional): Feed pruning
                   Keep only last N items per user in feed_items
                   Old entries deleted by a cleanup job
```
