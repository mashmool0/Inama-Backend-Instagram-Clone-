# Architecture Overview

## What Inama Is

An Instagram-like social platform built with microservice architecture.
One VPS, real patterns, all complexity intact — missing only multi-machine deployment.

---

## Request Flow (Bird's Eye)

```
Client (Next.js / React)
        │
        ▼
    [ Nginx ]           ← serves static files, reverse proxy (not load balancer)
        │
        ▼
  [ API Gateway ]       ← rate limiting, JWT verification, routing (Go)
        │
   ┌────┴─────────────────────────────────────┐
   ▼            ▼             ▼               ▼
[Auth]     [Posts +      [User /          [Feed +
Service]   Interactions] Social Graph]    Explore]
FastAPI       Go              Go              Go
   │            │               │              │
   │            └──── publishes events ──────→ │
   │                  to message queue         │
   │                       │                  │
   │               ┌───────┘                  │
   │               ▼                          │
   │      [Notifications]   [Search Indexer]  │
   │           Go               FastAPI       │
   │               │                │         │
   └───────────────┴────────────────┴─────────┘
                   │
          External Systems:
          SMS Provider    (OTP)
          Email Provider  (password reset)
          S3 / Liara      (photo storage)
          FCM / APNs      (push notifications)
          Elasticsearch   (search + hashtags)
```

---

## Functional Requirements

### Auth
- Register with phone number
- Login / logout
- SMS OTP verification
- Password reset via email
- Username required after first signup

### User & Profile
- View and edit own profile (bio, avatar, username)
- View other users' profiles
- Follow / unfollow
- Followers and following lists

### Posts
- Create a post: photo/video **or** text-only
- Delete own post
- View post with full details
- Like count and comment count visible
- Image resize worker: **deferred** (out of current scope)

### Interactions
- Like / unlike a post
- Add a comment
- View comments
- Counts visible on post

### Feed + Explore
- Home feed: posts from followed users, paginated (20 per scroll)
- Explore: random/recent public content (simple, no ML ranking)
- Both in the same service, different code paths

### Notifications
- Notify on: like, comment, follow
- Delivered async via push (FCM / APNs)
- Mark notifications as read

### Search
- Search by username
- Search by hashtag
- Powered by Elasticsearch

---

## Architecture Decisions Locked

| Decision | Choice | Reason |
|---|---|---|
| Architecture style | Microservices | Learning goal |
| Repo structure | Mono-repo | One team, one VPS, shared protos |
| Database ownership | Per-service | Core microservice discipline |
| Physical DB | One Postgres instance, separate databases | Logical isolation, cheap on one VPS |
| Sync communication | gRPC | Typed, fast, first-class Go support |
| Async communication | Message queue (Kafka or RabbitMQ — TBD) | Fanout, notifications, search indexing |
| Feed strategy | Fan-out-on-read + pagination (start) | Simple first; planned bottleneck to fix |
| Social graph storage | `follows` table in User service DB | Not Neo4j — simple relational is correct |
| Languages | Go + FastAPI (Python) | See services doc |
| Monitoring | Prometheus + Grafana + OpenTelemetry | Self-hosted, free |
| Deployment | Docker Compose + GitHub Actions | Single VPS, automated |
