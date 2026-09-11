# NITC PYQ Archive

> **Open to read. Verified to contribute. Built to last.**

The **NITC PYQ Archive** is a fast, clean, and durable community-run academic archive for National Institute of Technology Calicut (NITC) students. It organizes previous-year question papers (PYQs) by academic branch, course code, semester, academic year, and examination type.

Built as a single Go modular monolith with server-side rendering (SSR), PostgreSQL metadata, and early-2000s academic aesthetic.

---

## Non-Negotiable Product Principles

1. **Reading is public**: Anyone can browse, preview, and download published question papers without an account.
2. **Uploading requires verified NITC identity**: Contributions require Google Workspace authentication restricted to the `@nitc.ac.in` domain.
3. **PostgreSQL owns metadata**: Schema, courses, branches, audit logs, and status machines live in PostgreSQL.
4. **Internet Archive owns file bytes**: Published PDFs are archived to the Internet Archive (with `ObjectStore` interface abstraction). The container remains stateless.
5. **SSR First**: Server-rendered HTML with Go templates and minimal vendored HTMX for reactive filtering. No heavy SPA frameworks.

---

## Quickstart (Docker Compose)

The simplest way to run the complete stack:

```bash
# 1. Start the PostgreSQL database and Web application
docker compose up -d

# 2. Populate initial branches, courses, and sample question papers
docker compose run --rm archive-web /app/archive -seed

# 3. Access the archive
open http://localhost:8080
```

To view logs:
```bash
docker compose logs -f archive-web
```

Check health:
```bash
curl http://localhost:8080/healthz
```

---

## Local Development (Native Go)

### Prerequisites
- Go 1.22+ installed
- Docker (for running local PostgreSQL)

### 1. Start Local PostgreSQL
```bash
docker compose up -d postgres
```

### 2. Configure Environment
```bash
cp .env.example .env
```

### 3. Run Migrations and Seed Database
```bash
# Run embedded migrations
go run cmd/server/main.go -migrate

# Seed branches, courses, and sample PYQ documents
go run cmd/server/main.go -seed
```

### 4. Start the Application Server
```bash
go run cmd/server/main.go
```

The web server will listen at `http://localhost:8080`.

---

## Running Tests

Run the test suite covering storage adapters, path traversal protection, academic year math, and duplicate detection:

```bash
go test -v ./...
```

---

## Repository Structure

```text
├── cmd/
│   └── server/main.go          # Application entrypoint & CLI flags (-migrate, -seed)
├── internal/
│   ├── config/                 # Environment configuration loader
│   ├── db/                     # Connection pool, embedded migrations, and seeder
│   │   └── migrations/         # PostgreSQL schema migrations (000001_init.sql)
│   ├── storage/                # ObjectStore interface & LocalStorage implementation
│   ├── catalog/                # Branch and Course domain models, repo, service, handler
│   ├── resources/              # PYQ domain models, search, repo, service, handler
│   ├── reports/                # Issue reporting domain handler
│   └── web/                    # Templates renderer, request logging, security middleware
├── templates/
│   ├── layouts/base.html       # Retro academic layout
│   ├── pages/                  # Home, Branch, Course, Resource, Search, Report, Contribute
│   └── partials/               # Resource table and HTMX search fragments
├── static/
│   ├── css/main.css            # Custom retro early-2000s academic stylesheet
│   ├── js/app.js               # Minimal keyboard helpers
│   └── vendor/htmx.min.js      # Vendored HTMX
├── fixtures/
│   └── sample.pdf              # Sample PDF document for testing
├── tests/                      # Unit and integration tests
├── docs/
│   └── deployment-oci.md       # Oracle Cloud Infrastructure Always Free runbook
├── Dockerfile                  # Multi-stage production build
├── compose.yaml                # Multi-container Compose configuration
├── go.mod
└── go.sum
```

---

## Implementation Status

- [x] **Phase 0 — Foundation**: Modular Go monolith, PostgreSQL schema, embedded migrations runner, healthz probe, Docker Compose, security headers.
- [x] **Phase 1 — Public Archive**: Branch browsing, course catalog, filterable PYQ tables, resource detail cards, inline PDF preview, downloads, live HTMX search, seeded NITC branches and sample courses.
- [ ] **Phase 2 — Google OIDC Authentication**: Institutional `@nitc.ac.in` domain verification and session management.
- [ ] **Phase 3 — Upload & Validation Pipeline**: PDF magic bytes validator, SHA-256 duplicate checking, upload state machine.
- [ ] **Phase 4 — Internet Archive Integration**: S3-compatible archival client implementing `ObjectStore`.
- [ ] **Phase 5 — Moderation & Audit**: Review queue, publishing/rejecting workflow, audit trail.
- [ ] **Phase 6 — OCI Always Free Production**: Caddy automatic TLS and production deployment.
