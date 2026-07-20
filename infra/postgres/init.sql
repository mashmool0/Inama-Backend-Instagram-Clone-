-- Runs once on first Postgres container start (mounted into
-- /docker-entrypoint-initdb.d). Creates one database per service.
-- Each service owns its own DB — no cross-service SQL (CLAUDE.md guardrail #3).

CREATE DATABASE auth_db;
CREATE DATABASE user_db;
CREATE DATABASE posts_db;
CREATE DATABASE notif_db;

-- feed_db: storage is an OPEN DECISION (Postgres vs Redis lists — see
-- docs/09-roadmap.md). Created here so the Postgres option works out of the
-- box; remove if feed ends up Redis-only.
CREATE DATABASE feed_db;
