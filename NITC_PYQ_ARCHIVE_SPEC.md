# NITC PYQ Archive — Product & Engineering Specification

**Version:** 1.2  
**Status:** V1 implementation handoff / coding-agent specification  
**Target:** NIT Calicut student/community archive  
**Primary resource type in V1:** Previous-Year Question Papers (PYQs)  
**Hosting target:** Oracle Cloud Infrastructure (OCI) Always Free  
**Storage target:** Internet Archive  
**Frontend:** Server-rendered HTML + HTMX + custom CSS  
**Backend:** Go + Chi  
**Database:** PostgreSQL  
**Containerization:** Docker / Docker Compose

---

## 0. Agent Brief — Read This First

This document is the source of truth for the initial implementation of the **NITC PYQ Archive**.

Build a small, maintainable, student/community-run academic archive. The product is intentionally **not** a startup-style dashboard, SPA, or social platform. It should feel like a fast, well-organized academic archive with a slightly retro / early-2000s web aesthetic.

### Non-negotiable product principles

1. **Reading is public.** Visitors do not need an account to browse, preview, or download published PYQs.
2. **Uploading requires NITC identity.** Contributors authenticate with Google OAuth/OIDC and must belong to the configured NITC Google Workspace domain.
3. **PostgreSQL owns metadata and application state.** Do not use Internet Archive as the source of truth for application taxonomy, moderation, users, or permissions.
4. **Internet Archive owns published file bytes.** Do not permanently store published PDFs on the app container.
5. **SSR first.** Render pages on the server. Use HTMX only for small interactive areas. Do not introduce React, Next.js, Vue, Svelte, or another SPA framework in V1.
6. **One Go modular monolith.** Do not split the application into microservices.
7. **Docker everywhere.** Local development and production should use the same basic container model.
8. **Design for migration.** File storage must sit behind an `ObjectStore` interface so Internet Archive can later be replaced by MinIO, S3/R2, or another archival backend without rewriting the catalog domain.
9. **Moderation-first.** Uploaded resources should not become publicly listed merely because an upload request succeeded. V1 should support a pending/review flow.
10. **Keep the dependency count low.** Prefer standard Go packages and a small number of focused libraries over a large framework ecosystem.

### Do not build in V1 unless explicitly requested

- Notes/lecture notes as first-class resources
- Solutions
- Comments/forums
- User-to-user messaging
- Ratings/reputation systems
- Recommendation/AI systems
- Native mobile apps
- Elasticsearch/OpenSearch
- Complex JavaScript state management
- Microservices
- Serverless architecture
- Paid cloud object storage as a requirement
- Institutional LMS integration

---

# 1. Product Definition

## 1.1 Problem

Students at NIT Calicut often need previous-year question papers but the collection can be scattered across drives, chats, personal folders and informal archives. The goal is to create one simple public archive where papers are discoverable by branch, course, semester, academic year and exam type.

## 1.2 V1 Goal

A student should be able to go from the homepage to the exact PYQ they need with minimal friction, ideally within three primary navigation steps or fewer.

Example:

```text
Homepage
  -> CSE
  -> CS201
  -> End Sem 2025
  -> View / Download PDF
```

Searching should also be available directly from the homepage.

## 1.3 Core value proposition

> **Open to read. Verified to contribute. Built to last.**

---

# 2. Users and Roles

## 2.1 Guest / Student

Can:

- Browse branches
- Browse courses
- Search courses/resources
- Filter PYQs
- Open resource detail pages
- Preview published PDFs in the browser
- Download published PDFs
- Report a resource/problem without necessarily signing in

Cannot:

- Upload
- Edit metadata
- Moderate content
- Access admin routes

## 2.2 Contributor

A contributor is an authenticated user whose Google identity passes the NITC Workspace-domain check.

Can:

- Access contribution UI
- Submit PYQs
- See their own submissions/status

V1 contributor permissions should be intentionally narrow. Contributors do not get course-taxonomy administration or moderation permissions.

## 2.3 Reviewer

Can:

- Access moderation queue
- Review pending resources
- Publish/reject resources
- Request correction if that workflow is implemented
- Handle reports

## 2.4 Admin

Can additionally:

- Manage branches/courses
- Manage users and roles
- Block a contributor
- Remove/restore published resources
- View operational/archive status
- Inspect audit events

Use explicit RBAC in application code.

---

# 3. Information Architecture

```text
HOME
├── Search
├── Branches
│   ├── Branch page
│   │   └── Course list
│   │       └── Course page
│   │           └── PYQ/resource list
│   │               └── Resource detail / preview
├── Contribute
├── Login / Logout
└── Admin
    ├── Dashboard
    ├── Moderation queue
    ├── Reports
    ├── Courses
    ├── Branches
    ├── Users
    └── Archive health
```

---

# 4. Primary User Journeys

## 4.1 Guest browse

1. Open homepage.
2. Choose a branch OR enter a course code/name in search.
3. Open a course page.
4. Filter/select academic year and exam type if necessary.
5. Open resource.
6. Preview or download PDF.

No login interruption is allowed.

## 4.2 Contributor upload

1. Click **Contribute**.
2. Sign in with Google.
3. Backend validates OIDC response.
4. Confirm account belongs to the configured NITC Workspace domain.
5. Create a local application session.
6. Select branch/course/exam metadata.
7. Submit PDF.
8. Server validates the PDF and metadata.
9. Compute SHA-256 checksum.
10. Reject or warn on exact duplicate.
11. Persist submission as `PENDING_REVIEW`.
12. Moderator approves.
13. Application uploads approved PDF to Internet Archive.
14. Application records Internet Archive identifier/file reference.
15. Resource becomes `PUBLISHED` only after successful archival ingestion.

## 4.3 Moderation

1. Reviewer opens moderation queue.
2. Inspect title + metadata + PDF.
3. Publish, reject, or remove.
4. Every state-changing moderation action creates an audit event.
5. Failed Internet Archive ingestion remains visible to admins and can be retried.

---

# 5. Functional Requirements

| ID | Requirement |
|---|---|
| FR-01 | Public users can browse without authentication. |
| FR-02 | Users can search by course code and course name. |
| FR-03 | Users can filter by branch, semester, academic year, and exam type. |
| FR-04 | Published resources expose human-readable metadata and download/preview actions. |
| FR-05 | Only authorized NITC Google identities can enter contributor routes. |
| FR-06 | Uploads are PDF-only in V1. |
| FR-07 | Server validates PDF magic bytes, not merely extension/MIME type. |
| FR-08 | Server computes SHA-256 and detects exact duplicate files. |
| FR-09 | Uploads are subject to size/rate/concurrency limits. |
| FR-10 | Moderators can publish, reject, remove and restore resources. |
| FR-11 | Guests can report resources. |
| FR-12 | Admins can manage the branch/course taxonomy. |
| FR-13 | Security-sensitive and moderation state changes are audited. |
| FR-14 | Removed resources cannot be accessed through public application routes. |
| FR-15 | Application provides `/healthz`. |
| FR-16 | Core flows are keyboard accessible and use semantic HTML. |
| FR-17 | Published PDFs are not persistently stored in the app container. |
| FR-18 | The application can be started locally with Docker Compose. |
| FR-19 | The production stack can run on one Oracle VM with Docker Compose. |

---

# 6. Non-Functional Requirements

## 6.1 Performance

The archive is expected to be mostly read-heavy. V1 does not need distributed caching or a CDN.

Targets:

- Server-rendered normal catalog page should be lightweight.
- Avoid unnecessary JavaScript.
- Use database indexes for primary discovery/filter paths.
- Use pagination where lists could become large.
- Do not load full PDF contents through Postgres.

## 6.2 Availability

The initial deployment is intentionally simple: one OCI VM. Single-node failure is therefore possible. The design should minimize recovery difficulty rather than pretending to provide high availability.

## 6.3 Maintainability

- Clear package boundaries.
- Small interfaces.
- SQL migrations committed to Git.
- Structured logging.
- Tests for security-sensitive code and storage lifecycle.
- No hidden platform-specific state.

---

# 7. Recommended Technology Stack

## Backend

- **Go**
- **Chi** for routing
- Go `html/template` or **templ** for HTML rendering
- **HTMX** for partial page interactions
- **pgx** for PostgreSQL
- **sqlc** preferred once the SQL schema stabilizes
- **goose** or **golang-migrate** for migrations
- Standard `crypto/sha256`, `crypto/rand`, `net/http`, `context`, etc. where practical

## Frontend

- Server-rendered HTML
- Custom CSS
- Minimal JavaScript
- HTMX only where it materially improves UX

Do not introduce a frontend build system unless there is a clear requirement for one.

## Database

- PostgreSQL
- Relational schema
- B-tree indexes for common filters
- PostgreSQL full-text/trigram search only if simple indexed SQL search proves insufficient

## File archive

- Internet Archive for published PDFs
- Application-defined `ObjectStore` interface
- Local filesystem implementation may be used for development/testing
- Optional S3-compatible implementation may be added later

## Authentication

- Google OAuth 2.0 / OpenID Connect
- Server-side token validation
- NITC Workspace domain restriction
- Google `sub` as immutable external identity key

## Deployment

- Oracle Cloud Infrastructure Always Free
- Ubuntu ARM64 VM on OCI Ampere A1
- Docker Compose
- Caddy for HTTPS/reverse proxy
- PostgreSQL container
- Go application container

---

# 8. Why SSR Is the Default

Server-side rendering is intentional, not a temporary shortcut.

The archive primarily consists of:

- catalogs
- lists
- filters
- forms
- resource detail pages
- authentication flows
- moderation tables

These are naturally represented as server-rendered HTML.

Benefits for this project:

- less frontend code
- no SPA routing layer
- less client-side state
- excellent initial load performance for simple pages
- straightforward accessibility
- simple deployment
- simple caching model
- easier retro/early-web visual design
- one primary programming language
- easier codebase for future contributors to understand

HTMX can provide dynamic behavior without converting the application into a client-rendered SPA. Example: changing a search filter can request an HTML table fragment rather than a JSON response plus client-side rendering.

### Example principle

Prefer:

```text
GET /search?q=signals
       -> Go handler
       -> SQL query
       -> HTML fragment
       -> HTMX swaps result section
```

over:

```text
GET /api/search?q=signals
       -> JSON
       -> JavaScript state management
       -> virtual DOM rendering
```

Use JavaScript only for interactions where it clearly improves usability.

---

# 9. UI / Design Specification

## 9.1 Visual goal

The site should intentionally reject the current generic SaaS visual language.

### Avoid

- gradients
- glassmorphism
- giant hero sections
- oversized rounded cards
- floating blobs
- excessive shadows
- animated backgrounds
- neon/cyberpunk styling
- excessive iconography
- dashboard-style visual noise

### Prefer

- warm off-white or light neutral background
- near-black text
- restrained single accent color
- thin borders
- compact spacing
- dense information hierarchy
- rectangular or mildly rounded controls
- monospace for course codes, metadata and utility labels where appropriate
- visible focus outlines
- subtle hover states
- tables/list rows over card grids
- simple typography

The overall reference should be **early-2000s academic web / Internet archive / university website**, but cleaned up enough to remain pleasant and accessible today.

## 9.2 Homepage concept

```text
+--------------------------------------------------------------+
| NITC PYQ ARCHIVE                         [Search] [Contribute] |
+--------------------------------------------------------------+
|                                                              |
| Previous-year question papers, organized by course.          |
|                                                              |
| [ Search course code / course name / year................. ] |
|                                                              |
| BRANCHES                                                     |
| CSE     ECE     EEE     ME     CE     CH     BT     ...      |
|                                                              |
| RECENTLY ADDED                                               |
| 2025 | EC301 | Signals & Systems | ENDSEM | PDF              |
| 2025 | CS201 | Data Structures | MIDSEM   | PDF              |
| 2024 | ME203 | Thermodynamics | ENDSEM   | PDF              |
|                                                              |
+--------------------------------------------------------------+
```

## 9.3 Resource listing

Use dense rows rather than giant cards.

Example:

```text
YEAR     EXAM       RESOURCE                         FILE
2025     ENDSEM     Signals & Systems                PDF
2024     ENDSEM     Signals & Systems                PDF
2024     MIDSEM     Signals & Systems                PDF
2023     ENDSEM     Signals & Systems                PDF
```

## 9.4 Mobile

Desktop-first is acceptable, but the layout must remain usable on phones. Tables should collapse into stacked metadata blocks where required.

---

# 10. System Architecture

```text
                         INTERNET
                            |
                            v
                     +--------------+
                     |    Caddy     |
                     | TLS + proxy  |
                     +------+-------+
                            |
                            v
                    +---------------+
                    |   Go App      |
                    | SSR + HTMX    |
                    +---+-------+---+
                        |       |
                        |       +----------------+
                        v                        v
                +---------------+       +----------------+
                | PostgreSQL    |       | Google OIDC    |
                | metadata/auth |       | authentication |
                +---------------+       +----------------+
                        |
                        |
                        v
                +-------------------+
                | Storage Adapter   |
                +---------+---------+
                          |
                          v
                +-------------------+
                | Internet Archive  |
                | Published PDFs    |
                +-------------------+

            All application components run on ONE OCI VM.
```

The archive file bytes are intentionally outside the application VM for durable/public archival storage.

---

# 11. Oracle Cloud Deployment Architecture

## 11.1 Target infrastructure

Use **Oracle Cloud Infrastructure (OCI) Always Free**, specifically an **OCI Ampere A1 Flex** VM.

Current Oracle documentation states that the Always Free A1 allocation provides a total of **2 OCPUs and 12 GB RAM** for Always Free tenancies. Oracle also provides **200 GB total Always Free block volume**, though the exact usable allocation depends on how it is divided among resources. The free resources are tied to the tenancy's home region. Oracle notes that A1 instances can encounter temporary out-of-capacity errors when provisioning. See the official OCI Always Free documentation before deployment.

Recommended single-VM allocation:

```text
OCI Ampere A1 Flex
  OCPU: 2
  RAM: 12 GB
  OS: Ubuntu ARM64
  Boot volume: ~50 GB initially
```

This is substantially more than the application needs for early usage.

## 11.2 Why one VM

The user explicitly prefers hosting app + database in one place and the archive is expected to be small initially.

A single VM avoids:

- managed database fees
- multiple cloud services
- multiple deployment surfaces
- network/service configuration complexity
- unnecessary managed infrastructure for V1

## 11.3 Docker Compose production topology

```text
OCI VM
|
+-- Caddy container
|     |
|     +-- public :80/:443
|
+-- archive-web container
|     |
|     +-- Go SSR application
|
+-- postgres container
|     |
|     +-- persistent Docker volume
|
+-- optional future worker container
      |
      +-- IA ingestion / scans / cleanup
```

## 11.4 Network exposure

Only expose publicly:

- TCP 80
- TCP 443
- TCP 22 (SSH), preferably restricted by source IP if practical

Do **not** expose PostgreSQL to the public internet.

The app should reach PostgreSQL over the private Docker network using the Compose service name, for example:

```text
postgres:5432
```

---

# 12. Oracle Deployment and Operations Requirements

The coding agent should produce documentation for:

1. Creating the OCI VM.
2. Creating the VCN/subnet/security rules.
3. Installing Docker + Compose.
4. Setting up SSH.
5. Pointing a domain/subdomain at the VM public IP.
6. Configuring Caddy for automatic HTTPS.
7. Creating production environment secrets.
8. Running migrations.
9. Starting the stack.
10. Updating the stack.
11. Inspecting logs.
12. Backing up PostgreSQL.
13. Restoring PostgreSQL.
14. Retrying failed Internet Archive ingestions.

The application repository should contain a concise `docs/deployment-oci.md` as part of the implementation.

---

# 13. Database Design

PostgreSQL is the source of truth for the application catalog.

## 13.1 `users`

Suggested fields:

```sql
id                uuid primary key

google_sub        text not null unique
email             text not null

display_name      text
role              text not null
status            text not null

created_at        timestamptz not null
eupdated_at        timestamptz not null
last_login_at     timestamptz
```

Suggested roles:

```text
contributor
reviewer
admin
```

Suggested statuses:

```text
active
blocked
disabled
```

Do not use Google email as the immutable identity key. Store Google `sub` as the stable identity identifier.

## 13.2 `branches`

```sql
id                uuid primary key
code              text not null unique
name              text not null
active            boolean not null default true
created_at        timestamptz not null
```

Examples:

```text
CSE
ECE
EEE
ME
CE
CH
BT
```

Do not assume this list is complete; the admin UI should manage it.

## 13.3 `courses`

```sql
id                uuid primary key
code              text not null unique
name              text not null
description       text
active            boolean not null default true
created_at        timestamptz not null
updated_at        timestamptz not null
```

## 13.4 `course_branches`

Many-to-many mapping because a course can be common to multiple branches.

```sql
course_id         uuid not null references courses(id)
branch_id         uuid not null references branches(id)
primary key (course_id, branch_id)
```

## 13.5 `resources`

Suggested fields:

```sql
id                    uuid primary key
course_id             uuid not null references courses(id)
uploader_id           uuid references users(id)

resource_type         text not null
semester              text not null
academic_year_start   integer not null
exam_type             text not null

title                 text not null

file_size_bytes       bigint not null
sha256                text not null

storage_provider      text not null
storage_identifier    text
storage_filename      text

status                text not null
archive_status        text not null
scan_status           text

created_at            timestamptz not null
updated_at            timestamptz not null
published_at          timestamptz
removed_at            timestamptz
```

V1 values:

```text
resource_type:
  PYQ

exam_type:
  MIDSEM
  ENDSEM
  QUIZ
  OTHER

status:
  PENDING_REVIEW
  PUBLISHED
  REJECTED
  REMOVED

archive_status:
  NOT_APPLICABLE
  PENDING
  UPLOADING
  UPLOADED
  FAILED

scan_status:
  NOT_RUN
  PENDING
  PASSED
  FAILED
```

### Important

Use an integer `academic_year_start` such as:

```text
2025
```

and derive display text as:

```text
2025-26
```

Do not store multiple inconsistent strings such as `25-26`, `2025/26`, `2025-2026`.

## 13.6 `reports`

```sql
id                uuid primary key
resource_id       uuid not null references resources(id)
reporter_id       uuid references users(id)
category          text not null
message           text
status            text not null
created_at        timestamptz not null
resolved_at       timestamptz
resolved_by       uuid references users(id)
```

Allow anonymous reports, but add rate limiting and anti-abuse controls.

## 13.7 `audit_events`

```sql
id                uuid primary key
actor_user_id     uuid references users(id)
action            text not null
entity_type       text not null
entity_id         uuid
metadata_json     jsonb
created_at        timestamptz not null
```

Audit examples:

```text
LOGIN_SUCCESS
LOGIN_REJECTED
UPLOAD_CREATED
UPLOAD_VALIDATED
RESOURCE_PUBLISHED
RESOURCE_REJECTED
RESOURCE_REMOVED
RESOURCE_RESTORED
USER_BLOCKED
COURSE_CREATED
COURSE_UPDATED
```

## 13.8 `sessions`

Use a server-side session table unless there is a compelling reason to choose another approach.

Suggested fields:

```sql
id                text primary key
user_id           uuid not null references users(id)
expires_at        timestamptz not null
created_at        timestamptz not null
last_seen_at      timestamptz
ip_hash           text
user_agent        text
```

Store only a session identifier in the browser cookie.

---

# 14. Database Indexes

At minimum:

```text
courses(code unique)
resources(course_id, academic_year_start, exam_type)
resources(status, created_at)
course_branches(branch_id, course_id)
resources(sha256)
resources(uploader_id, created_at)
reports(status, created_at)
```

Search strategy for V1:

1. exact/prefix course-code search
2. `ILIKE` course-name search
3. add `pg_trgm` later if needed

Do not add Elasticsearch/OpenSearch in V1.

---

# 15. Authentication Specification

## 15.1 Google OIDC flow

```text
Browser
  -> /login/google
  -> Google authorization
  -> /auth/google/callback
  -> token exchange
  -> ID-token validation
  -> local user lookup by Google sub
  -> local session creation
  -> redirect to /contribute
```

## 15.2 Required validation

Validate the returned identity server-side:

- issuer (`iss`)
- audience (`aud`)
- expiry (`exp`)
- signature
- `email_verified == true`
- hosted domain (`hd`) equals configured NITC Workspace domain

The `hd` value sent to Google as a login hint is **not sufficient** by itself. Authorization must be based on the validated identity returned by Google.

## 15.3 Environment variable

```text
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=
GOOGLE_HOSTED_DOMAIN=
```

The exact NITC Google Workspace domain must be confirmed before production configuration.

## 15.4 Session rules

Cookie should be:

```text
Secure
HttpOnly
SameSite=Lax (or stricter where compatible)
```

Use expiration and server-side revocation.

---

# 16. Upload and Validation Pipeline

## 16.1 Important storage rule

The application must not treat its container filesystem as permanent storage.

Temporary files are allowed only for bounded validation/scanning and must be removed after processing.

## 16.2 V1 upload lifecycle

```text
DRAFT
  |
  v
UPLOAD RECEIVED
  |
  v
VALIDATING
  |
  +--> invalid --> REJECTED
  |
  v
DUPLICATE CHECK
  |
  +--> duplicate --> REJECT / WARN
  |
  v
PENDING_REVIEW
  |
  +--> rejected --> REJECTED
  |
  v
APPROVED
  |
  v
IA_INGESTION_PENDING
  |
  v
UPLOADING_TO_IA
  |
  +--> failure --> IA_FAILED --> retry
  |
  v
IA_UPLOADED
  |
  v
PUBLISHED
```

## 16.3 PDF validation

V1 accepts only PDFs.

Validate:

1. upload size <= configured limit
2. file begins with PDF magic bytes (`%PDF-`)
3. declared MIME type is not trusted by itself
4. SHA-256 checksum is computed
5. optionally parse enough PDF structure to reject obviously corrupt files
6. optional malware scanning before publication

Initial configurable limit:

```text
25 MB
```

Do not hard-code the value throughout the application.

## 16.4 Duplicate detection

Exact duplicates are determined by SHA-256.

Example:

```text
sha256 = 7a3b...e91
```

Create a uniqueness strategy that prevents duplicate active resources while allowing an administrator to handle unusual historical cases.

Do not rely only on filenames for duplicate detection.

---

# 17. Internet Archive Integration

## 17.1 Why Internet Archive

Internet Archive is being used because this project is fundamentally an archive and the primary files are public academic documents.

Internet Archive provides programmatic tooling for creating/managing items, uploading files, modifying metadata and downloading files. Its documented upload mechanism uses an S3-like interface and requires server-side archive credentials.

The application should use the official API/tooling as documented at implementation time rather than depending on undocumented web behavior.

## 17.2 Architectural rule

**Postgres is the application source of truth. Internet Archive is the published-file archive.**

Do not read branch/course/moderation state from IA metadata.

## 17.3 Storage abstraction

Define a small interface, conceptually:

```go
type ObjectStore interface {
    Put(ctx context.Context, object Object) (StoredObject, error)
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

The exact interface may be adjusted to reflect the actual IA workflow.

Implementations:

```text
storage/
  storage.go
  internetarchive.go
  local.go
```

Future:

```text
  s3.go
  minio.go
```

## 17.4 Internet Archive item strategy

Start with **one IA item per individual published resource** unless implementation testing demonstrates a better grouping model.

Example conceptual identifier:

```text
nitc-pyq-cs201-2025-endsem-<short-unique-id>
```

Do not make the identifier depend solely on user-controlled strings.

Upload a clean server-generated filename, e.g.:

```text
cs201-2025-endsem.pdf
```

Store in Postgres:

```text
storage_provider = internet_archive
storage_identifier = <IA identifier>
storage_filename = <archive filename>
sha256 = <hash>
```

## 17.5 IA metadata

The item should receive useful archive metadata, where appropriate, such as:

- title
- description
- subject/category
- date/year
- creator/uploader attribution if policy permits
- mediatype appropriate for documents/text

Do not store private user/session data in IA metadata.

## 17.6 Credentials

Never expose Internet Archive access keys to browsers.

Use server-side environment variables or secret storage.

Suggested variables:

```text
IA_ACCESS_KEY=
IA_SECRET_KEY=
IA_COLLECTION=
IA_IDENTIFIER_PREFIX=nitc-pyq
```

Use the minimum permissions/configuration supported by the IA account/workflow.

## 17.7 Publication ordering

Preferred V1 sequence:

```text
moderator approves
      -> IA upload succeeds
      -> DB records IA reference
      -> resource becomes PUBLISHED
```

If the IA upload fails:

```text
resource remains unavailable publicly
archive_status = FAILED
```

An admin/reconciliation command should be able to retry.

Do not mark an item public merely because the local upload file exists.

## 17.8 Public download

Preferred V1 behavior:

```text
Application resource page
        |
        +--> canonical Internet Archive file link
```

A backend proxy may be introduced later for:

- centralized rate limiting
- consistent response headers
- analytics
- access control
- replacing storage providers without changing public URLs

Do not build the proxy in V1 unless it is needed for a concrete requirement.

---

# 18. HTTP Routes

The app is not API-first, but route boundaries should remain clean.

## Public

```text
GET  /
GET  /branches/{branch}
GET  /courses/{course}
GET  /resources/{id}
GET  /resources/{id}/download
GET  /search
GET  /reports/new?resource={id}
POST /reports
GET  /healthz
```

## Authentication

```text
GET  /login/google
GET  /auth/google/callback
POST /logout
```

## Contributor

```text
GET  /contribute
POST /uploads/init
POST /uploads/complete
GET  /submissions
GET  /submissions/{id}
```

Exact upload route structure may be simplified if the final upload pipeline does not require a two-step browser flow.

## Admin/Reviewer

```text
GET  /admin
GET  /admin/resources
GET  /admin/resources/{id}
POST /admin/resources/{id}/publish
POST /admin/resources/{id}/reject
POST /admin/resources/{id}/remove
POST /admin/resources/{id}/restore
GET  /admin/reports
POST /admin/reports/{id}/resolve
GET  /admin/courses
POST /admin/courses
POST /admin/courses/{id}/edit
GET  /admin/branches
POST /admin/branches
POST /admin/branches/{id}/edit
GET  /admin/users
POST /admin/users/{id}/block
POST /admin/users/{id}/role
GET  /admin/archive-health
```

Use standard redirects after successful form submissions.

---

# 19. HTML / Template Structure

Recommended:

```text
templates/
  layouts/
    base.html
    admin.html
  pages/
    home.html
    branch.html
    course.html
    resource.html
    search.html
    contribute.html
    login.html
    submissions.html
    admin_dashboard.html
    admin_resource.html
    admin_reports.html
    admin_courses.html
    admin_branches.html
    admin_users.html
  partials/
    search_results.html
    resource_table.html
    flash.html
    pagination.html
    moderation_row.html
```

The exact templating technology may be either `html/template` or `templ`, but choose one and use it consistently.

---

# 20. Suggested Go Repository Layout

```text
nitc-pyq-archive/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── google.go
│   │   ├── session.go
│   │   └── middleware.go
│   ├── catalog/
│   │   ├── service.go
│   │   ├── handler.go
│   │   └── repository.go
│   ├── resources/
│   │   ├── service.go
│   │   ├── handler.go
│   │   └── repository.go
│   ├── uploads/
│   │   ├── service.go
│   │   ├── validation.go
│   │   └── handler.go
│   ├── moderation/
│   │   ├── service.go
│   │   └── handler.go
│   ├── reports/
│   │   ├── service.go
│   │   └── handler.go
│   ├── storage/
│   │   ├── storage.go
│   │   ├── internetarchive.go
│   │   └── local.go
│   ├── db/
│   │   ├── queries/
│   │   └── generated/
│   ├── audit/
│   ├── web/
│   │   ├── renderer.go
│   │   └── middleware.go
│   └── config/
├── templates/
├── static/
│   ├── css/
│   │   └── main.css
│   └── js/
│       └── app.js
├── migrations/
├── tests/
├── scripts/
├── docs/
│   └── deployment-oci.md
├── Dockerfile
├── compose.yaml
├── .dockerignore
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

Do not cargo-cult every directory above if it increases complexity. The architecture should remain understandable to a student contributor.

---

# 21. Docker Specification

## 21.1 Application image

Use a multi-stage build.

Conceptually:

```dockerfile
FROM golang:<stable>-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o /bin/archive ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/archive /archive
COPY --from=build /app/templates /templates
COPY --from=build /app/static /static
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/archive"]
```

The actual image may use a current Go base and an architecture-neutral build strategy if buildx is used. Production target is OCI ARM64.

## 21.2 Compose services

Minimum:

```text
postgres
archive-web
caddy
```

Optional later:

```text
worker
```

## 21.3 Persistent storage

Postgres gets a named volume.

Example:

```text
postgres_data:/var/lib/postgresql/data
```

Do not create a persistent PDF volume for production.

---

# 22. Environment Variables

Create `.env.example` containing placeholders only.

```text
APP_ENV=development
APP_ADDR=:8080
APP_BASE_URL=http://localhost:8080

DATABASE_URL=postgres://archive:archive@postgres:5432/archive?sslmode=disable

SESSION_SECRET=

GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
GOOGLE_HOSTED_DOMAIN=

IA_ACCESS_KEY=
IA_SECRET_KEY=
IA_COLLECTION=
IA_IDENTIFIER_PREFIX=nitc-pyq

MAX_UPLOAD_MB=25

LOG_LEVEL=info
```

Secrets must never be committed.

---

# 23. Security Specification

## Authentication / authorization

- Validate Google tokens server-side.
- Check issuer.
- Check audience.
- Check expiry.
- Require `email_verified`.
- Require validated hosted domain.
- Use Google `sub` as stable identity.
- Enforce role checks in middleware/service layer.

## Session security

- Secure cookie
- HttpOnly
- SameSite
- Expiration
- Revocation
- Regenerate session on login

## Upload security

- strict maximum size
- PDF signature check
- SHA-256
- optional malware scanner
- rate limiting
- concurrency limit
- no shell execution on user-provided filenames
- never construct storage paths directly from untrusted raw input

## CSRF

Protect state-changing browser requests, including:

- uploads
- reports if applicable
- admin actions
- logout
- taxonomy changes

## SQL injection

Use parameterized SQL or sqlc. Never concatenate user-provided strings into SQL.

## Secrets

- environment variables for V1
- secret manager can be added later
- never store secrets in Git
- never put IA or Google secrets into browser JavaScript

## Rate limiting

At minimum rate-limit:

- Google auth initiation/callback abuse
- report submission
- upload creation
- upload completion
- login/session-related endpoints

V1 can use an in-process token bucket because the application is single-node. If deployment later becomes multi-instance, move rate-limit state to shared storage.

---

# 24. Copyright / Takedown / Community Policy

The site should not assume that every uploaded academic paper is freely licensed simply because a student possessed a copy.

V1 should display a concise contribution policy stating that contributors are responsible for uploading material they are allowed to share and that the archive can remove disputed material.

Do not claim official NITC institutional ownership, sponsorship, or endorsement unless such a relationship actually exists.

Provide:

- Report resource
- contact/takedown path
- admin removal capability
- audit trail for removals

The public site can describe itself as a student/community archive.

---

# 25. Search and Discovery

V1 should remain PostgreSQL-only.

Search fields:

- course code
- course name
- title
- branch
- academic year
- semester
- exam type

Recommended starting behavior:

```text
course code:
  exact + prefix match

course name/title:
  ILIKE '%term%'

filters:
  branch
  semester
  year
  exam type
```

Use pagination.

Add PostgreSQL `pg_trgm` if simple `ILIKE` becomes slow or poor. Do not introduce a separate search engine in V1.

---

# 26. Resource Detail Page

Every public resource page should show:

```text
Course code
Course name
Branch / common course
Semester
Academic year
Exam type
Title
File type
File size
Publication status
```

Actions:

```text
[ Preview PDF ]
[ Download PDF ]
[ Report this resource ]
```

The page should not require login.

---

# 27. Moderation UI

The moderation queue should prioritize utility over visual polish.

Suggested row:

```text
PENDING
CS201 | Data Structures
2025-26 | S5 | ENDSEM
Submitted by: <display name>
File: 4.8 MB

[ Review ]
```

Review page:

```text
Metadata
  course
  branch
  semester
  year
  exam type
  filename

File
  preview/download

Checks
  PDF signature: PASS
  SHA256: ...
  Duplicate: NO
  Scan: PASS / NOT RUN

Actions
  [ Publish ] [ Reject ] [ Request Correction ]
```

---

# 28. Audit / Operational Visibility

Admin dashboard should expose at least:

- pending submissions
- failed IA ingestions
- recent reports
- published count
- total resources
- recently published resources
- blocked users

Use structured JSON logs in production.

Include request IDs for important actions.

---

# 29. Backups and Recovery

Postgres is the most important data that needs independent backup because published PDFs are archived externally.

V1 minimum:

```text
Daily pg_dump
     |
     v
Separate backup location
```

Keep several generations.

Document how to restore from a dump.

Do not claim a backup strategy is complete unless restoration has actually been tested.

The Oracle VM itself should not be treated as the only copy of the database.

---

# 30. Deployment Strategy

## Development

```bash
docker compose up -d postgres
# run Go app locally
```

Optional:

```bash
docker compose --profile dev up
```

for auxiliary local services.

## Production

```text
GitHub
  |
  v
OCI VM
  |
  +-- git checkout / deployment mechanism
  |
  +-- docker compose pull/build
  |
  +-- docker compose up -d
```

A GitHub Actions workflow may be added for:

- go test
- lint
- Docker build
- deployment via SSH or an OCI-friendly deployment method

Do not introduce a complicated Kubernetes setup.

---

# 31. Suggested CI Pipeline

On every pull request:

```text
gofmt check
 go test ./...
 static analysis/lint
 Docker build
```

On main:

```text
build
push/deploy
```

The first version may simply build the Docker image directly on the OCI host if setting up a registry is unnecessary.

---

# 32. Testing Strategy

## Unit tests

At minimum cover:

- PDF validation
- SHA-256 duplicate detection
- academic year parsing
- role authorization
- OIDC claim validation logic
- resource state transitions
- IA storage adapter error handling

## Integration tests

Cover:

- Postgres migrations
- public resource browsing
- contributor authorization
- upload metadata lifecycle
- moderation publish/reject/remove
- report flow

## Security tests

Explicitly test:

- non-NITC Google identity rejected
- manipulated email/domain payload cannot bypass auth
- unauthenticated POST to contributor/admin routes rejected
- contributor cannot call admin endpoints
- removed resource cannot be downloaded
- duplicate SHA-256 is handled
- oversized upload rejected
- non-PDF disguised as `.pdf` rejected
- CSRF-protected requests cannot be forged

---

# 33. Implementation Phases

## Phase 0 — Foundation

Deliver:

- repository
- Go app skeleton
- config
- Dockerfile
- Compose
- Postgres
- migrations
- base templates
- base CSS
- health endpoint

Exit condition:

```text
docker compose up
-> app opens
-> DB connects
-> migrations apply
-> /healthz succeeds
```

## Phase 1 — Public archive

Deliver:

- seeded branches
- seeded courses
- homepage
- branch pages
- course pages
- search
- filters
- resource detail pages
- sample PDFs / fixture metadata

Exit condition:

A guest can navigate the full browse/download experience.

## Phase 2 — Google authentication

Deliver:

- Google OIDC
- session management
- contributor role
- domain restriction
- login/logout

Exit condition:

Valid NITC user can enter contributor area; external Google account cannot.

## Phase 3 — Upload + validation

Deliver:

- upload UI
- PDF validation
- SHA-256
- duplicate detection
- upload state machine
- temporary file cleanup

Exit condition:

Valid PDF becomes a pending submission without becoming public automatically.

## Phase 4 — Internet Archive

Deliver:

- `ObjectStore` interface
- IA implementation
- item creation/upload
- IA identifier storage
- ingestion status
- retry path
- public download linking

Exit condition:

Approved PDF is uploaded to IA and the database records the canonical archive reference.

## Phase 5 — Moderation / reports / audit

Deliver:

- reviewer/admin roles
- moderation queue
- publish/reject/remove/restore
- report flow
- audit events
- admin dashboard

Exit condition:

The project can be safely opened for student submissions.

## Phase 6 — OCI production

Deliver:

- OCI VM
- Docker Compose production config
- Caddy HTTPS
- firewall/security configuration
- production secrets
- Postgres backups
- deployment docs
- logging/monitoring basics

Exit condition:

Public beta works from a fresh OCI VM and recovery procedures are documented.

## Phase 7 — Future resource generalization

Only after PYQ workflow is stable:

```text
resource_type:
  PYQ
  NOTE
  SLIDE
  ASSIGNMENT
  SOLUTION
```

Do not let future flexibility make V1 complicated.

---

# 34. Definition of Done — First Public Beta

The system is ready when all of the following are true:

1. A guest can browse branches/courses without login.
2. A guest can search and filter resources.
3. A guest can preview/download a published PDF.
4. A valid NITC Google user can authenticate.
5. A non-NITC Google account cannot authenticate into contributor functionality.
6. User identity is based on validated Google `sub`.
7. Uploaded non-PDF files are rejected.
8. Oversized uploads are rejected.
9. Exact duplicate files are detected using SHA-256.
10. Uploaded PDFs are not permanently stored on the application container.
11. Resources are reviewed before publication.
12. Approved resources are archived to Internet Archive.
13. IA upload failures are visible and retryable.
14. Postgres stores the resource metadata and IA reference.
15. Reviewers/admins can remove public resources.
16. Removed resources are inaccessible through public application routes.
17. Reports work.
18. State-changing admin actions are audited.
19. Postgres backups exist and restoration has been tested.
20. The full application stack runs from Docker Compose.
21. The production stack runs on one OCI Always Free VM.
22. HTTPS is enabled in production.
23. No secrets are committed to Git.
24. The visual design follows the archive/retro direction and does not look like a generic modern SaaS dashboard.
25. The documentation explains how another contributor can run the project locally and deploy it.

---

# 35. Agent Implementation Rules

The coding agent should follow these rules while implementing:

### Rule 1 — Prefer simplicity over abstraction

Do not create abstractions merely because they are theoretically reusable. The `ObjectStore` interface is required because storage migration is an explicit product concern; most other abstractions should earn their existence.

### Rule 2 — Keep domain state in Postgres

Do not use filenames, IA metadata, browser state, or Google profile data as substitutes for the local resource model.

### Rule 3 — Server-rendered by default

When adding a page, first ask whether a normal SSR route solves it. Add HTMX only where it produces a meaningful interaction improvement.

### Rule 4 — No SPA creep

Do not add React/Next/Vue/Svelte because a single interaction appears difficult. Prefer a normal HTML form + server route + HTMX fragment.

### Rule 5 — No public file storage on the VM

The VM is application infrastructure, not the archive. Published PDFs belong in Internet Archive.

### Rule 6 — Validate on the server

Never trust client-side MIME types, extensions, hidden form fields, or UI restrictions for security.

### Rule 7 — Do not auto-publish by accident

Make the moderation state machine explicit. A successful upload request is not equivalent to publication.

### Rule 8 — Make failures visible

External IA failures must produce a durable local status such as `FAILED`, an error log/event, and a retry path.

### Rule 9 — Keep migrations deterministic

A clean empty PostgreSQL database must become the correct schema by running migrations in order.

### Rule 10 — Document every operational secret and external dependency

Add `.env.example` and deployment documentation.

---

# 36. Initial Seed Data

The implementation should include a seed mechanism for development.

Example branches:

```text
CSE
ECE
EEE
ME
CE
CH
BT
```

Example courses can be placeholders or actual verified courses supplied by the project owner. Do not invent a detailed official NITC curriculum and present it as authoritative.

Development fixtures should clearly be marked as test data.

---

# 37. Future Enhancements

Potential post-V1 work:

- notes
- slides
- assignments
- solutions
- topic/unit tags
- contributor history
- course popularity
- download analytics
- duplicate/similar-resource detection
- OCR/text indexing
- full-text question search
- API
- MinIO/self-hosted storage option
- alternate frontend only if SSR becomes insufficient

The data model should support these without forcing them into V1.

---

# 38. Current Platform Notes

These platform details were checked against official documentation in September 2026 and should be rechecked before production deployment because cloud policies and limits can change.

## OCI

Oracle's current Always Free documentation states:

- Always Free compute resources are available for the life of the account within the published limits.
- OCI Ampere A1 Flex provides a total Always Free allocation equivalent to **2 OCPUs and 12 GB RAM** for Always Free tenancies.
- Always Free resources must be provisioned in the tenancy's home region.
- There is **200 GB total** Always Free block volume storage across the tenancy's eligible resources.
- Oracle provides Always Free object/archive storage allowances, but this project intentionally does not depend on OCI object storage for the primary published PDF archive.
- Oracle notes that Always Free A1 provisioning can encounter temporary out-of-capacity errors in a region.
- Oracle may reclaim idle Always Free compute instances under its published idle-resource policy; monitor the instance and platform status.

Primary reference:

- https://docs.oracle.com/en-us/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm

## Internet Archive

Internet Archive's developer tooling documents:

- programmatic item access
- upload/download operations
- metadata operations
- S3-like upload credentials/interface

References:

- https://internetarchive.readthedocs.io/en/stable/api.html
- https://internetarchive.readthedocs.io/en/stable/cli.html

## Google identity

Google's OpenID Connect guidance should be used to validate the returned ID token and to use `sub` as the stable identifier. Workspace-domain restrictions should use validated identity information such as `hd`, not merely a UI hint.

References:

- https://developers.google.com/identity/openid-connect/reference
- https://developers.google.com/identity/sign-in/web/backend-auth

---

# 39. Suggested First-Agent Prompt

The following can be given to the coding agent together with this specification:

> Implement the NITC PYQ Archive according to `NITC_PYQ_ARCHIVE_SPEC.md`.
>
> Start with Phase 0 and Phase 1 only unless instructed otherwise. Build a small Go modular monolith using Chi, PostgreSQL, SSR HTML templates, HTMX, and Docker Compose. Do not introduce React/Next/Vue/Svelte or microservices.
>
> Make the repository runnable locally from a clean checkout. Create migrations, seed support, base templates, a retro/early-2000s archive-style CSS system, public branch/course browsing, search, filters, and resource pages. Include tests for the critical domain behavior.
>
> Design the code so Phase 2 Google OIDC and Phase 4 Internet Archive integration can be added without restructuring the application. In particular, keep authentication, catalog, resources, moderation, and storage behind clear package boundaries.
>
> Implement the `ObjectStore` interface early, with a local development implementation first. Do not add an Internet Archive production implementation until the core resource lifecycle is stable.
>
> Keep the application stateless except for PostgreSQL and temporary bounded upload files. Published PDFs must not be stored permanently in the web container.
>
> At the end of each phase, update the README with exact commands for running, testing, migrating, and seeding the application.

---

# 40. Final Architecture Summary

```text
                  PUBLIC INTERNET
                        |
                        v
                 +-------------+
                 |    Caddy    |
                 | HTTPS / TLS |
                 +------+------+
                        |
                        v
                 +-------------+
                 |   Go SSR    |
                 | Chi + HTMX  |
                 +------+------+ 
                        |
             +----------+----------+
             |                     |
             v                     v
      +-------------+       +--------------+
      | PostgreSQL  |       | Google OIDC  |
      | metadata    |       | contributor  |
      +-------------+       +--------------+
             |
             v
      +-------------+
      | ObjectStore |
      |  interface  |
      +------+------+ 
             |
             v
      +----------------+
      | Internet       |
      | Archive        |
      | published PDFs |
      +----------------+

All compute/database components:
  ONE OCI Ampere A1 VM
  ONE Docker Compose stack

Published file bytes:
  Internet Archive

Application truth:
  PostgreSQL
```

**Core product principle:** keep the application small, the archive durable, the interface fast, and the contribution process accountable.
