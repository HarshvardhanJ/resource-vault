# NITC Resource Vault — Product & Engineering Specification

**Status:** Implementation specification  
**Version:** 2.0  
**Repository:** HarshvardhanJ/resource-vault  
**Product name:** NITC Resource Vault

---

## 1. Product

NITC Resource Vault is a student-maintained academic archive for NIT Calicut.

### V1 scope

- Public browsing and downloading of Previous Year Question Papers (PYQs).
- No login required for reading.
- Google OAuth/OIDC required for contributors.
- Only users with a valid NITC Google account may contribute.
- Human moderation before a new resource becomes public.
- Resources organized by academic unit, course, academic year, semester, exam/session, and resource type.
- Course/academic metadata is editable in a controlled, wiki-like manner.
- Actual published PDF bytes are stored in Internet Archive.
- PostgreSQL is the source of truth for application metadata and moderation state.

### Future scope

- Notes, lecture material, assignments, lab material, question banks, etc.
- Better search and filtering.
- Course pages and resource history.
- Community metadata suggestions.
- Optional reputation/contributor history.

Do not implement future scope unless explicitly requested.

---

## 2. Academic Units

Seed these exact academic units.

### Bachelor of Architecture

- **B. Arch**

### Bachelor of Technology

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

### 4-year Integrated Teacher Education Programme

- **4-year Integrated Teacher Education Programme (ITEP) B.Sc–B.Ed**

Suggested stable codes:

| Code | Name |
|---|---|
| BARCH | B. Arch |
| BT | Biotechnology |
| CHE | Chemical Engineering |
| CE | Civil Engineering |
| CSE | Computer Science and Engineering |
| EEE | Electrical and Electronics Engineering |
| ECE | Electronics and Communication Engineering |
| ENE | Energy Engineering |
| EP | Engineering Physics |
| HSS | Humanities and Social Sciences |
| MSE | Materials Science and Engineering |
| ME | Mechanical Engineering |
| PE | Production Engineering |
| ITEP | ITEP B.Sc–B.Ed |

Display names may change without changing internal IDs. Codes are seed data and may be corrected by an administrator.

---

## 3. Catalog Model

Do not assume that one course belongs to exactly one branch.

A course may be common across multiple branches, offered by one academic unit but taken by several units, or renamed/re-coded over time.

Therefore:

- academic_units and courses are separate entities;
- use a many-to-many course_academic_units relationship;
- resource records reference a course by UUID;
- course code and name are metadata, not file-storage paths.

### Who decides course metadata?

1. **Contributor suggestion**
   - Any authenticated NITC contributor may suggest a correction.
   - Examples: incorrect course code, typo, missing academic unit.

2. **Moderator/admin review**
   - Suggested changes enter a review queue.
   - A moderator/admin approves or rejects them.

3. **Admin override**
   - Admins can directly edit canonical metadata.

Every accepted metadata change creates an audit record.

---

## 4. Roles

### Anonymous visitor

Can browse academic units, courses, search resources, view metadata, and download published resources.

Cannot upload, edit metadata, or access moderation tools.

### NITC contributor

Authenticated using Google OAuth/OIDC.

Can upload a resource, suggest metadata corrections, see contribution history, and report a resource.

### Moderator

Can review pending resources, approve/reject uploads, review metadata suggestions, hide/unpublish resources, and inspect contributor information needed for moderation.

### Admin

Everything a moderator can do, plus managing academic units, courses, moderators, application settings, audit logs, and recovery operations.

---

## 5. Authentication

Use Google OAuth 2.0 / OpenID Connect.

Requirements:

- Reading is public.
- Uploading requires authentication.
- Restrict contributor access to the official NITC Google Workspace domain.
- Store Google's stable OIDC sub identifier as the external identity key.
- Store email and display name for convenience/audit.
- Do not store Google access tokens unless a future feature requires them.
- Use secure, HTTP-only session cookies.
- Use CSRF protection for state-changing browser requests.
- The allowed domain must be configurable rather than hard-coded.

Suggested:

    GOOGLE_ALLOWED_DOMAIN=nitc.ac.in

Validate the OIDC claims according to Google's current documentation; do not treat a user-entered email domain as proof of authorization.

---

## 6. Resource Lifecycle

States:

- pending
- processing
- published
- rejected
- failed
- hidden

### Upload flow

1. Contributor signs in.
2. Contributor chooses/enters course metadata.
3. Contributor selects a PDF.
4. Server validates authentication, domain, size, MIME type, PDF magic bytes, and basic readability.
5. Store the upload temporarily.
6. Calculate SHA-256.
7. Check for an identical existing file.
8. Create a pending resource.
9. Moderator reviews it.
10. On approval, publish the file to Internet Archive.
11. Persist the Internet Archive identifier/path/checksum.
12. Mark resource published.
13. Public pages expose it.

If Internet Archive fails, keep the resource retryable. Never mark a resource published before the archive upload succeeds.

---

## 7. Internet Archive Storage

Internet Archive is the V1 archival storage backend.

The application must not depend directly on Internet Archive throughout the codebase.

Use internal/storage with an ObjectStore interface covering:

- Put
- Get
- Delete
- Exists
- URL

The current repository already contains this abstraction. Preserve and extend it rather than scattering provider-specific API calls.

### IA model

Use a dedicated Internet Archive item/collection controlled by the project.

The application database remains the canonical index.

A logical key may look like:

    pyqs/CSE/CS201/2025/endsem.pdf

The exact IA item/file layout is an implementation detail of the adapter.

### Store in PostgreSQL

At minimum:

- storage provider
- archive identifier
- object key / filename
- SHA-256
- file size
- content type
- published timestamp
- storage status

Never make an Internet Archive URL the only identifier stored in the application.

### IA credentials

Keep credentials only in deployment secrets/environment variables.

Suggested variables:

    IA_ACCESS_KEY
    IA_SECRET_KEY
    IA_COLLECTION
    IA_ITEM_PREFIX

Never commit these values.

---

## 8. PostgreSQL

PostgreSQL is the source of truth for application metadata.

### Core tables

#### users

- id UUID PK
- google_sub TEXT UNIQUE NOT NULL
- email TEXT NOT NULL
- display_name TEXT
- role TEXT NOT NULL
- active BOOLEAN NOT NULL DEFAULT TRUE
- created_at
- updated_at
- last_login_at

#### academic_units

- id UUID PK
- code TEXT UNIQUE NOT NULL
- name TEXT NOT NULL
- type TEXT
- active BOOLEAN
- created_at
- updated_at

#### courses

- id UUID PK
- code TEXT NOT NULL
- name TEXT NOT NULL
- description TEXT
- active BOOLEAN
- created_at
- updated_at

Do not make code globally unique unless the chosen catalog model guarantees it.

#### course_academic_units

- course_id UUID FK
- academic_unit_id UUID FK
- primary key (course_id, academic_unit_id)

#### resources

- id UUID PK
- course_id UUID FK
- resource_type TEXT
- title TEXT
- academic_year INT
- semester TEXT
- exam_type TEXT
- original_filename TEXT
- content_type TEXT
- file_size BIGINT
- sha256 TEXT
- status TEXT
- storage_provider TEXT
- storage_key TEXT
- storage_identifier TEXT
- uploaded_by UUID FK
- reviewed_by UUID FK NULL
- rejection_reason TEXT NULL
- created_at
- updated_at
- published_at NULL

#### metadata_suggestions

- id UUID PK
- entity_type TEXT
- entity_id UUID
- proposed_changes JSONB
- reason TEXT
- submitted_by UUID FK
- status TEXT
- reviewed_by UUID FK NULL
- created_at
- reviewed_at NULL

#### audit_log

- id UUID PK
- actor_user_id UUID FK NULL
- action TEXT
- entity_type TEXT
- entity_id UUID NULL
- before JSONB NULL
- after JSONB NULL
- created_at
- request_id TEXT NULL

### Indexes

At minimum:

- courses(code)
- courses(name)
- resources(course_id)
- resources(status)
- resources(academic_year)
- resources(sha256)
- resources(published_at)
- users(google_sub)
- metadata_suggestions(status)

Continue using versioned SQL migrations. The current repository embeds and applies migrations; preserve that mechanism, but production must never use destructive startup migrations.

---

## 9. Duplicate Detection

Use SHA-256 of the uploaded file.

- identical hash + same metadata: warn/avoid duplicate;
- identical PDF + different metadata: allow moderator decision;
- never deduplicate solely on filename.

---

## 10. Metadata / Wiki UX

Course pages should provide a **Suggest an edit** action.

Example:

    Course code: CS201
    Course name: Data Structures

A contributor can suggest:

    CS201 -> CS202

with an optional reason.

Public pages show canonical metadata. Suggestions and internal moderation notes are visible only to authorized reviewers.

Accepted changes should be auditable.

---

## 11. SSR Architecture

Use server-side rendering.

Recommended stack:

- Go
- net/http
- Chi router if already used/desired
- html/template
- HTMX for small interactive enhancements
- vanilla JavaScript only where necessary
- custom CSS

Do not introduce React, Vue, Angular, Next.js, or another large frontend framework for V1.

This application is primarily catalogs, tables, course pages, document listings, forms, and search/filter views, making SSR an excellent fit.

HTMX may be used for search, filtering, pagination, moderation actions, edit forms, and partial updates while the server remains the source of truth.

---

## 12. Routes

Public:

    GET /
    GET /branches
    GET /branches/{id}
    GET /courses/{id}
    GET /courses/{id}/resources
    GET /search
    GET /resources/{id}
    GET /resources/{id}/download

Authentication:

    GET /auth/google
    GET /auth/google/callback
    POST /auth/logout

Contributor:

    GET /contribute
    POST /contribute
    GET /me/resources
    GET /suggest/{entity}/{id}
    POST /suggest/{entity}/{id}

Moderation:

    GET /moderation
    GET /moderation/resources
    POST /moderation/resources/{id}/approve
    POST /moderation/resources/{id}/reject
    GET /moderation/suggestions
    POST /moderation/suggestions/{id}/approve
    POST /moderation/suggestions/{id}/reject

Admin:

    GET /admin
    GET /admin/courses
    POST /admin/courses
    PATCH /admin/courses/{id}
    PATCH /admin/academic-units/{id}

Exact routes may differ if the current implementation has a coherent existing convention.

---

## 13. UI / Brand

Primary product name everywhere:

# NITC Resource Vault

Do not use NITC PYQ Archive as the primary product name.

### Visual direction

Old university portal / early-2000s web archive.

Avoid:

- gradients;
- glassmorphism;
- excessive rounded cards;
- huge hero sections;
- animated blobs;
- excessive shadows;
- AI/SaaS visual language;
- oversized typography;
- unnecessary animation.

Prefer:

- visible borders;
- compact layout;
- blue hyperlinks;
- underlined links;
- muted background;
- dense tables;
- simple buttons;
- clear separators;
- restrained colors;
- information density.

Retro should mean functional and nostalgic, not intentionally unusable. Accessibility and responsive behavior still matter.

---

## 14. File Handling

V1 accepts PDFs only.

Validate both declared MIME type and actual PDF signature (%PDF-).

Suggested initial limit:

    MAX_UPLOAD_BYTES=26214400

Do not load arbitrarily large files entirely into memory.

Use streaming where practical.

Temporary files must be removed after processing.

Never execute uploaded files.

Set correct Content-Type, Content-Disposition, and X-Content-Type-Options: nosniff headers as appropriate.

---

## 15. Security

Required:

- secure cookies;
- HTTP-only session cookies;
- CSRF protection;
- authorization checks on every protected route;
- input validation;
- parameterized SQL;
- HTML escaping through html/template;
- upload size limits;
- PDF validation;
- rate limiting on authentication/upload endpoints;
- request IDs;
- structured logging;
- no secrets in logs;
- no secrets in Git.

Admin/moderator authorization must be enforced server-side.

Do not trust hidden form fields for role/permission decisions.

---

## 16. Abuse / Moderation

Provide:

- pending queue;
- approve/reject;
- rejection reason;
- hide/unpublish;
- report resource;
- audit log.

A contributor must not approve their own submission.

Keep enough audit information to determine who uploaded, who reviewed, what changed, and when.

---

## 17. Docker

Use Docker for local development and production.

Recommended services:

- app
- postgres
- caddy

Internet Archive is external.

Production topology:

    Internet
       |
      Caddy
       |
      Go app
       |
    PostgreSQL

               -> Internet Archive

### Application image

Use a multi-stage build.

Build a static Go binary where practical.

Run the final container as a non-root user.

Do not ship development tooling in the production image.

Expose only the application port internally.

---

## 18. Oracle Cloud Always Free

Production target: Oracle Cloud Infrastructure Always Free.

Use an Ampere A1 ARM VM.

Target allocation:

- 2 OCPUs
- 12 GB RAM
- storage within current Oracle Always Free limits

The application must support ARM64 Docker builds.

### OCI setup

Install:

- Ubuntu or another supported Linux distribution;
- Docker Engine;
- Docker Compose plugin;
- Caddy.

Expose only required public ports:

- 80/tcp
- 443/tcp
- SSH restricted to an administrator's trusted source where practical.

Do not expose PostgreSQL publicly.

PostgreSQL must be reachable only from the Docker network.

---

## 19. Production Environment

Use a server-side .env file or another secret mechanism. Never commit it.

Suggested variables:

    APP_ENV=production
    APP_ADDR=:8080
    DATABASE_URL=postgres://...
    SESSION_SECRET=...
    GOOGLE_CLIENT_ID=...
    GOOGLE_CLIENT_SECRET=...
    GOOGLE_REDIRECT_URL=https://<domain>/auth/google/callback
    GOOGLE_ALLOWED_DOMAIN=nitc.ac.in
    IA_ACCESS_KEY=...
    IA_SECRET_KEY=...
    IA_COLLECTION=...
    IA_ITEM_PREFIX=...
    MAX_UPLOAD_BYTES=26214400
    TRUST_PROXY=true

Generate strong random session secrets.

---

## 20. PostgreSQL Production Configuration

Use a persistent Docker volume.

Never run docker compose down -v on production unless intentionally destroying the database.

Configure:

- strong database password;
- private Docker network;
- persistent volume;
- sensible connection pool limits;
- reasonable timeouts.

The app should fail clearly if the database is unavailable.

---

## 21. Backups

Internet Archive protects the published PDF layer, not PostgreSQL metadata.

Minimum:

- daily pg_dump;
- retain several recent backups;
- store backups outside the live PostgreSQL container;
- ideally store backups outside the OCI VM as well.

Test restoration before calling the system production-ready.

---

## 22. Observability

Use structured logging.

Every request should have a request ID.

Log:

- request start/end;
- status;
- latency;
- route;
- important application errors.

Do not log OAuth tokens, session secrets, database passwords, IA credentials, or uploaded document contents.

Provide:

    GET /healthz

Optionally provide:

    GET /readyz

where readiness includes database connectivity.

---

## 23. CI/CD

GitHub Actions should eventually:

1. run go test ./...;
2. run formatting/static analysis;
3. build the application;
4. build an ARM64 production Docker image;
5. optionally build AMD64 for local development;
6. publish the image;
7. deploy to OCI after an approved change.

Initially, manual deployment is acceptable:

    git pull
    docker compose build
    docker compose up -d

Automate only after manual deployment is reliable.

---

## 24. Repository Structure

Prefer the current structure where it is already sound.

Target conceptual layout:

    cmd/server/main.go

    internal/
      auth/
      catalog/
      config/
      db/migrations/
      reports/
      resources/
      storage/
        storage.go
        local.go
        internetarchive.go
      web/

    templates/
    static/
      css/
      js/

    docs/
      deployment-oci.md

    Dockerfile
    compose.yaml
    go.mod
    README.md
    NITC_RESOURCE_VAULT_SPEC.md

The current repository already has embedded templates, DB migrations, catalog models, middleware, and an ObjectStore abstraction. Preserve these foundations.

---

## 25. Deployment Readiness Checklist

### Database

- [ ] production migrations deterministic
- [ ] no destructive startup migrations
- [ ] indexes exist
- [ ] connection pool configured
- [ ] PostgreSQL private
- [ ] persistent volume
- [ ] backups tested

### Internet Archive

- [ ] dedicated IA item/collection created
- [ ] API credentials configured as secrets
- [ ] upload adapter implemented
- [ ] retries/backoff implemented
- [ ] checksum recorded
- [ ] failed publication retryable
- [ ] published only after successful IA operation

### OAuth

- [ ] Google OAuth client created
- [ ] production callback configured
- [ ] allowed domain configured
- [ ] session secret generated
- [ ] secure cookies enabled
- [ ] CSRF protection enabled

### Application

- [ ] ARM64 image builds
- [ ] app runs as non-root
- [ ] upload limits enabled
- [ ] PDF signature validation enabled
- [ ] rate limiting enabled
- [ ] health endpoints exist
- [ ] structured logs exist
- [ ] errors do not expose internals

### OCI

- [ ] VM provisioned
- [ ] Docker installed
- [ ] ports 80/443 configured
- [ ] SSH restricted
- [ ] DNS configured
- [ ] Caddy configured
- [ ] PostgreSQL not public
- [ ] firewall configured
- [ ] backups configured

---

## 26. Development Phases

### Phase 1 — Catalog foundation

- seed all academic units;
- verify schema;
- course model;
- many-to-many course/unit relation;
- admin course management;
- course pages.

### Phase 2 — Public archive

- resource model;
- resource listing;
- course filtering;
- search;
- public download;
- duplicate detection.

### Phase 3 — Authentication

- Google OIDC;
- NITC-domain restriction;
- sessions;
- contributor role.

### Phase 4 — Upload + moderation

- PDF upload;
- pending queue;
- approval/rejection;
- audit trail.

### Phase 5 — Internet Archive

- IA storage adapter;
- upload;
- checksum;
- retries;
- published-state transitions.

### Phase 6 — Metadata wiki

- correction suggestions;
- moderation;
- audit history.

### Phase 7 — Production

- Docker hardening;
- OCI VM;
- Caddy;
- HTTPS;
- PostgreSQL backup;
- monitoring;
- deployment documentation.

---

## 27. Agent Instructions

This document is the source of truth for implementation.

1. Inspect existing code before modifying it.
2. Prefer incremental changes over rewrites.
3. Preserve working architecture unless the specification explicitly requires a change.
4. Keep the application SSR-first.
5. Do not introduce a frontend framework.
6. Keep provider-specific storage logic behind ObjectStore.
7. Never commit secrets.
8. Never invent academic-unit/course data without a source or explicit instruction.
9. Database changes must be migrations.
10. Add tests for non-trivial business logic.
11. Keep handlers thin; business logic belongs in services.
12. Keep SQL/database code out of templates and HTTP handlers.
13. Return user-friendly errors without exposing internals.
14. Never make an unreviewed upload public.
15. Never mark a resource published until archival storage succeeds.
16. Metadata editing must be auditable.
17. Do not expose PostgreSQL publicly.
18. Production must run correctly on ARM64.
19. Do not optimize prematurely.
20. If existing implementation conflicts with this spec, explain the conflict and make the smallest safe change.

### Definition of Done

A feature is complete only when implementation, migrations (if needed), authorization, validation, error handling, tests, responsive templates, and relevant documentation are complete.

---

## 28. Initial Seed Data

The first migration/seed should create the 14 academic units listed in Section 2.

Do not seed guessed courses yet unless verified curriculum data exists.

Courses should initially be created through the admin catalog interface or a separately verified seed dataset.

---

## 29. Product Principles

1. Archive first.
2. Public by default for published resources.
3. Verified contribution.
4. Human moderation.
5. Simple technology.
6. Open architecture.
7. Retro interface, modern engineering.
8. PostgreSQL is the catalog authority; Internet Archive stores files.
9. Student-maintained collective stewardship.
10. Designed to grow.

---

## 30. Current Recommended Architecture

    Browser
       |
       | HTTPS
       v
    Caddy
       |
       v
    Go SSR application
       |
       +------------------+
       |                  |
       v                  v
    PostgreSQL       Internet Archive
    metadata         published PDFs
       |
       v
    moderation/audit

Everything except published PDF bytes runs on the OCI VM in Docker.

The application remains portable because storage is accessed through ObjectStore.

---

## 31. First Implementation Objective

Before adding large amounts of functionality, align the current repository with this specification:

1. Rename product-facing references to NITC Resource Vault.
2. Seed all 14 academic units.
3. Confirm/fix course-to-academic-unit many-to-many modeling.
4. Ensure course metadata is admin-controlled with suggestion workflow.
5. Verify PostgreSQL migrations and indexes.
6. Verify SSR architecture.
7. Implement/verify Google OIDC and NITC-domain authorization.
8. Implement/verify the Internet Archive storage adapter.
9. Verify upload -> moderation -> IA publication state transitions.
10. Harden Docker/OCI configuration.
11. Add production environment configuration documentation.
12. Run tests and build an ARM64-compatible production image.

Do not build notes, advanced search, recommendation systems, mobile apps, or other V2 features during this phase.
