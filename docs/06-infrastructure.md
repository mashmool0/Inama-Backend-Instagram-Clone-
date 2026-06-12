# Infrastructure

## Mono-Repo vs Multi-Repo

### The Question

With multiple services, do you put them in separate Git repositories or one?

```
Multi-repo:                    Mono-repo:
  inama-auth/                    inama/
  inama-user/                      services/
  inama-posts/                       auth/
  inama-feed/                        user/
  inama-notifications/               posts/
  inama-search/                      feed/
  inama-gateway/                     notifications/
                                     search/
                                     gateway/
                                   proto/         ← shared .proto files
                                   docker-compose.yml
                                   .github/workflows/
```

### Decision: Mono-Repo

For this project: **mono-repo, unambiguously.**

| Concern | Multi-Repo | Mono-Repo |
|---|---|---|
| Shared `.proto` files | Copy between repos (drift risk) or a 3rd "protos" repo | One `proto/` folder, referenced by all services |
| Cross-service changes (rename a field) | Open PRs in multiple repos, coordinate merges | One PR, one review, one merge |
| Running all services locally | Clone 7 repos, configure each | One clone, `docker-compose up` |
| CI/CD | 7 separate pipelines | One pipeline, conditional per service |
| Team overhead | Manageable with big teams | Easier for small/solo |
| Independent deployment | Easier per-repo (GitHub release per service) | Handled with path filters in CI |

Multi-repo pays off when:
- Different teams own different services (access control)
- Services have completely independent release cycles
- The repo grows so large that cloning it is slow

None of these apply to Inama. Multi-repo here is complexity without benefit.

---

## Directory Structure

```
inama/
├── services/
│   ├── gateway/          # Go
│   ├── auth/             # Python / FastAPI
│   ├── user/             # Go
│   ├── posts/            # Go
│   ├── feed/             # Go
│   ├── notifications/    # Go
│   └── search/           # Python / FastAPI
├── proto/                # shared .proto definitions
│   ├── auth.proto
│   ├── user.proto
│   ├── posts.proto
│   └── feed.proto
├── infra/
│   ├── nginx/
│   │   └── nginx.conf
│   ├── postgres/
│   │   └── init.sql      # creates all databases
│   └── prometheus/
│       └── prometheus.yml
├── docker-compose.yml    # runs the whole system
├── docker-compose.dev.yml
└── .github/
    └── workflows/
        └── deploy.yml
```

---

## Docker and Docker Compose

Each service is one Docker container. `docker-compose.yml` defines the whole system.

Why Docker Compose (not Kubernetes):
- Kubernetes is the right tool for multi-machine orchestration
- On one VPS, Docker Compose is simpler, faster, and teaches the same container concepts
- You can migrate to Kubernetes later (the Dockerfiles stay the same)

```yaml
# docker-compose.yml (simplified)
services:
  nginx:
    image: nginx:alpine
    ports: ["80:80", "443:443"]
    depends_on: [gateway]

  gateway:
    build: ./services/gateway
    environment:
      AUTH_SERVICE_ADDR: auth:50051
      USER_SERVICE_ADDR: user:50051

  auth:
    build: ./services/auth
    environment:
      DATABASE_URL: postgres://postgres:secret@postgres/auth_db

  user:
    build: ./services/user
    environment:
      DATABASE_URL: postgres://postgres:secret@postgres/user_db

  postgres:
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./infra/postgres/init.sql:/docker-entrypoint-initdb.d/init.sql

  redis:
    image: redis:7-alpine

  kafka:                  # or rabbitmq — TBD
    image: confluentinc/cp-kafka:latest

  elasticsearch:
    image: elasticsearch:8.13.0

  prometheus:
    image: prom/prometheus
    volumes:
      - ./infra/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports: ["3000:3000"]
```

---

## Deployment: GitHub Actions

Automated deploy on push to `main`.

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to VPS
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: |
            cd /opt/inama
            git pull origin main
            docker compose pull
            docker compose up --build -d
            docker compose ps
```

**What this does on every push to main:**
1. GitHub Actions runner SSHes into your VPS
2. Pulls latest code
3. Rebuilds changed containers
4. Restarts with zero-downtime (Docker Compose handles it)

**Secrets stored in GitHub:**
- `VPS_HOST` — your server IP
- `VPS_USER` — SSH username
- `VPS_SSH_KEY` — private key (never in code)

---

## Path-Based Deploys (Optimization)

When only one service changes, rebuild only that service:

```yaml
- name: Detect changed services
  id: changes
  run: |
    CHANGED=$(git diff --name-only HEAD~1 HEAD | grep '^services/' | cut -d/ -f2 | sort -u)
    echo "services=$CHANGED" >> $GITHUB_OUTPUT

- name: Rebuild only changed services
  run: |
    for svc in ${{ steps.changes.outputs.services }}; do
      docker compose build $svc
      docker compose up -d $svc
    done
```

This matters when you have 7 services and you're only changing one.
Rebuilding all 7 on every push wastes 5 minutes.

---

## VPS Setup Checklist

On first setup, run once:

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh

# Clone repo
git clone git@github.com:yourname/inama.git /opt/inama

# Set environment variables
cp .env.example .env
# edit .env with real secrets

# Start everything
docker compose up -d

# Set up SSL (Let's Encrypt)
certbot --nginx -d yourdomain.com
```

---

## Service Discovery (How Services Find Each Other)

On one VPS with Docker Compose, service discovery is automatic:
Docker creates a network where each service is reachable by its name.

```
gateway connects to auth:50051    ← "auth" resolves to auth container's IP
auth connects to postgres/auth_db ← "postgres" resolves to postgres container's IP
```

No Consul, no Kubernetes DNS, no config files.
Docker Compose's internal network handles it.
This is why the environment variable is `AUTH_SERVICE_ADDR: auth:50051`
and not an IP address — the name is stable, the IP is not.
