# NITC Resource Vault

> **Open to read. Verified to contribute. Built to last.**

**NITC Resource Vault** is a fast, clean, community-run academic archive for National Institute of Technology Calicut (NITC) students. V1 focuses on previous-year question papers (PYQs), organized by academic unit, course, semester, academic year, and examination type.

Built as a single Go modular monolith with server-side rendering (SSR), PostgreSQL metadata, Internet Archive storage for published PDFs, and an intentionally simple early-2000s academic aesthetic.

---

## Non-Negotiable Product Principles

1. **Reading is public**: Anyone can browse, preview, and download published resources without an account.
2. **Uploading requires verified NITC identity**: Contributions require Google Workspace authentication restricted to the configured NITC domain.
3. **PostgreSQL owns metadata**: Courses, academic units, resources, users, moderation state, and audit logs live in PostgreSQL.
4. **Internet Archive owns published file bytes**: Published PDFs are archived to Internet Archive through the `ObjectStore` abstraction. The app container remains stateless.
5. **SSR first**: Server-rendered HTML with Go templates and minimal HTMX. No heavy SPA framework.
6. **Controlled metadata editing**: Students may suggest catalog corrections; reviewers/admins approve canonical changes.
7. **Moderation first**: An upload is never public merely because it was submitted.

---

## Academic Units

The initial catalog contains all current units supplied for the project:

- B. Arch
- Biotechnology
- Chemical Engineering
- Civil Engineering
- Computer Science and Engineering
- Electrical and Electronics Engineering
- Electronics and Communication Engineering
- Energy Engineering
- Engineering Physics
- Humanities and Social Sciences
- Materials Science and Engineering
- Mechanical Engineering
- Production Engineering
- 4-year Integrated Teacher Education Programme (ITEP) B.Sc–B.Ed

Courses are **not** restricted to one unit. Common/shared courses use a many-to-many course-to-academic-unit mapping.

---

## Quickstart (Docker Compose)

```bash
docker compose up -d
curl http://localhost:8080/healthz
```

For local development:

```bash
cp .env.example .env
docker compose up -d postgres
go run ./cmd/server -migrate
go run ./cmd/server -seed
go run ./cmd/server
```

The default development server listens on `http://localhost:8080`.

---

## Production Target

Production is designed for **Oracle Cloud Infrastructure (OCI) Always Free**:

```text
OCI Ampere A1 VM
├── Caddy
├── Go application
└── PostgreSQL

Published PDFs
└── Internet Archive
```

PostgreSQL is never exposed to the public internet. Only HTTP/HTTPS (and restricted SSH) should be publicly reachable.

See [`docs/deployment-oci.md`](docs/deployment-oci.md) and [`NITC_RESOURCE_VAULT_SPEC.md`](NITC_RESOURCE_VAULT_SPEC.md) for the deployment and implementation specification.

---

## Internet Archive

Internet Archive is the V1 durable archive for published PDFs. The application database remains the source of truth for catalog and moderation state.

The storage layer is provider-independent:

```text
internal/storage/
├── storage.go
├── local.go
└── internetarchive.go   # production backend
```

This lets the project move to MinIO, S3/R2, or another compatible backend later without redesigning the catalog.

Do not commit Internet Archive credentials. Use deployment environment variables/secrets.

---

## Wiki-like Course Metadata

Course information is intentionally community-correctable but not freely writable.

- Any authenticated NITC contributor can **suggest** a correction.
- Moderators/admins review suggestions.
- Admins can directly edit canonical metadata.
- Accepted changes are written to the audit log.

This prevents the catalog from becoming a free-for-all while still allowing students to fix mistakes.

---

## Repository Structure

```text
├── cmd/
│   └── server/main.go
├── internal/
│   ├── auth/                 # Google OIDC and sessions
│   ├── catalog/              # Courses + academic units
│   ├── config/               # Environment configuration
│   ├── db/                   # PostgreSQL + embedded migrations
│   ├── reports/              # Resource reporting
│   ├── resources/            # PYQ resource lifecycle
│   ├── storage/              # ObjectStore + storage adapters
│   └── web/                  # SSR templates + middleware
├── templates/
│   ├── layouts/
│   ├── pages/
│   └── partials/
├── static/
│   ├── css/
│   └── js/
├── fixtures/
├── tests/
├── docs/
│   └── deployment-oci.md
├── Dockerfile
├── compose.yaml
├── NITC_RESOURCE_VAULT_SPEC.md
├── go.mod
└── go.sum
```

---

## Implementation Status

- [x] Go modular monolith foundation
- [x] PostgreSQL + embedded migrations
- [x] SSR + HTMX foundation
- [x] Public branch/course/resource browsing
- [x] NITC academic-unit seed catalog
- [x] ObjectStore abstraction
- [ ] Course/academic-unit migration fully adopted throughout the UI
- [ ] Google OIDC with NITC-domain enforcement
- [ ] Upload validation + SHA-256 deduplication
- [ ] Internet Archive production adapter
- [ ] Moderation/review workflow
- [ ] Metadata suggestion workflow
- [ ] Production OCI hardening
- [ ] Caddy HTTPS deployment
- [ ] PostgreSQL backup/restore procedure

---

## License / Content Notice

The software license and the rights status of individual academic PDFs are separate concerns. The project should not claim ownership of uploaded exam papers or imply official NITC endorsement without an explicit institutional relationship.

A public report/takedown mechanism is required before a broad public launch.
