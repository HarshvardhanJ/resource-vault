# OCI Always Free Deployment & Operations Guide

This guide details deployment of **NITC Resource Vault** on an **Oracle Cloud Infrastructure (OCI) Always Free** Ampere A1 VM.

## 1. Target topology

```text
OCI Ampere A1 VM
├── Caddy :80/:443
├── archive-web :8080
└── PostgreSQL (private Docker network + persistent volume)

archive-web
├── Google OIDC
└── Internet Archive (published PDFs)
```

App and PostgreSQL intentionally live on one VM. Published PDF bytes live in Internet Archive.

## 2. OCI VM provisioning

Use the OCI Always Free Ampere A1 Flex shape. Target 2 OCPUs and 12 GB RAM within the current Always Free allocation, with a modest boot volume.

Use an ARM64-capable Ubuntu image. The production Docker image must build/run on ARM64.

## 3. OCI firewall / network rules

Public ingress:

- TCP 80
- TCP 443
- TCP 22 for SSH; restrict SSH source where practical.

Do **not** expose TCP 5432. PostgreSQL must only be reachable through the Docker network.

Also configure the host firewall consistently with the OCI security list.

## 4. Install Docker

Install Docker Engine, Buildx, and the Docker Compose plugin using Docker's official Ubuntu repository.

After installation, add the deployment user to the `docker` group and reconnect the SSH session.

## 5. Checkout

```bash
git clone https://github.com/HarshvardhanJ/resource-vault.git /opt/resource-vault
cd /opt/resource-vault
```

## 6. Production environment

```bash
cp .env.example .env
chmod 600 .env
```

Set real values:

```text
APP_ENV=production
APP_ADDR=:8080
APP_BASE_URL=https://YOUR_DOMAIN
DATABASE_URL=postgres://archive:STRONG_PASSWORD@postgres:5432/archive?sslmode=disable
POSTGRES_USER=archive
POSTGRES_PASSWORD=STRONG_PASSWORD
POSTGRES_DB=archive
SESSION_SECRET=<long random secret>
MAX_UPLOAD_MB=25
LOG_LEVEL=info

GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=https://YOUR_DOMAIN/auth/google/callback
GOOGLE_HOSTED_DOMAIN=nitc.ac.in

IA_ACCESS_KEY=...
IA_SECRET_KEY=...
IA_COLLECTION=...
IA_IDENTIFIER_PREFIX=nitc-resource-vault
```

Never commit `.env` or real credentials.

## 7. Google OAuth

Configure a Google OAuth/OIDC client with the exact production callback:

`https://YOUR_DOMAIN/auth/google/callback`

The backend must validate the provider's signed identity and enforce the configured NITC Workspace domain. The `GOOGLE_HOSTED_DOMAIN` value is a server-side authorization policy, not merely a UI hint.

## 8. Internet Archive

Create the project's IA collection/item strategy before enabling public contributions.

The app database is authoritative for catalog/moderation state. Internet Archive stores published bytes.

The application must never expose IA credentials to the browser.

An approved resource becomes public only after successful IA ingestion. Failed ingestion remains retryable/reconcilable.

## 9. Caddy

Create `deploy/Caddyfile` from `deploy/Caddyfile.example`, replacing the example hostname.

The production Compose file publishes Caddy on ports 80/443 and keeps the Go app bound to localhost:8080 on the host.

Caddy terminates TLS and reverse proxies to the Go service.

## 10. Start

```bash
docker compose --profile production up -d --build
```

Check:

```bash
docker compose ps
curl -i https://YOUR_DOMAIN/healthz
```

Logs:

```bash
docker compose logs -f archive-web
docker compose logs -f caddy
docker compose logs -f postgres
```

## 11. Migrations

The application contains an embedded migration runner and applies only unapplied migrations.

For an explicit run:

```bash
docker compose exec archive-web /app/archive -migrate
```

Production migrations must never drop/recreate existing data automatically.

## 12. Seed data

Seed the canonical academic-unit list and only verified course data.

```bash
docker compose run --rm archive-web /app/archive -seed
```

Do not seed invented course codes/names merely to make the UI look populated.

## 13. PostgreSQL backups

Create a daily dump outside the database container and replicate it away from the VM.

Example:

```bash
mkdir -p /opt/backups/resource-vault
docker compose exec -T postgres pg_dump -U archive -d archive \
  | gzip > /opt/backups/resource-vault/archive-$(date +%F).sql.gz
```

Use a cron/systemd timer after the manual procedure is verified.

Test restoration periodically.

## 14. Restore

For a clean recovery VM, restore into an initialized PostgreSQL container:

```bash
gunzip < /opt/backups/resource-vault/archive-YYYY-MM-DD.sql.gz \
  | docker compose exec -T postgres psql -U archive -d archive
```

Do not restore over production blindly; verify target database and backup timestamp first.

## 15. Updates

```bash
git pull --ff-only
docker compose --profile production up -d --build
docker compose ps
```

Review recent logs after deployment.

## 16. Destructive command warning

Never run this on production unless intentionally destroying the database:

```bash
docker compose down -v
```

The PostgreSQL data volume is part of the live system and is not a backup.

## 17. Recovery model

The initial architecture is intentionally single-node. If the VM is lost:

1. Provision another OCI A1 VM.
2. Install Docker/Compose.
3. Clone the repository.
4. Restore the PostgreSQL backup.
5. Restore `.env`/secrets from the secure secret source.
6. Start the Compose production profile.
7. Point DNS at the replacement VM.

Published PDFs remain in Internet Archive independent of VM loss.
