# Research: Implementing Incidents (Similar to Uptime Kuma)

**Date**: 2026-03-04T08:43:46Z
**Git Commit**: 6232a5138814412de6f35b5205b364085e1a3c27
**Branch**: main
**Repository**: peekaping

## What Uptime Kuma's Incidents Feature Is

Incidents are **manually authored announcements** that appear on public status pages. They are _not_ triggered automatically by monitor failures. An admin creates an incident to communicate service disruptions to visitors. Each incident has a title, Markdown content, a color style (info/warning/danger/primary/light/dark), and lifecycle states (active+pinned → unpinned → resolved → deleted).

Key facts from Uptime Kuma:

- Incidents belong to a **status page** (not to individual monitors)
- Only one incident can be pinned/active at a time (a known limitation)
- Active/pinned incidents appear at the **top** of the status page as colored banners
- Resolved incidents appear in a scrollable **incident history** section
- Content supports **Markdown** (rendered to HTML, sanitized for XSS)

### Uptime Kuma Data Model

| DB Column          | Type         | Nullable | Default     | Notes                          |
| ------------------ | ------------ | -------- | ----------- | ------------------------------ |
| `id`               | INTEGER      | NO       | auto-inc    | PK                             |
| `title`            | VARCHAR(255) | NO       | —           | Required                       |
| `content`          | TEXT         | NO       | —           | Markdown supported             |
| `style`            | VARCHAR(30)  | NO       | `"warning"` | One of 6 Bootstrap variants    |
| `created_date`     | DATETIME     | NO       | `now()`     | Set once on creation           |
| `last_updated_date`| DATETIME     | YES      | NULL        | Set on edit or resolve         |
| `pin`              | BOOLEAN      | NO       | `true`      | Whether shown in active area   |
| `active`           | BOOLEAN      | NO       | `true`      | Whether incident is open       |
| `status_page_id`   | INTEGER      | YES      | NULL        | FK to status_page              |

### Uptime Kuma Incident Lifecycle

```
[CREATE]
  → active = true, pin = true
  → Appears at the top of status page
       |
       ├──► [EDIT]     → Updates title, content, style, pin
       ├──► [UNPIN]    → pin = false, still active, moves out of top section
       └──► [RESOLVE]  → active = false, pin = false, moves to history

[DELETE] (available at any stage) → Hard-deleted from DB
```

### Uptime Kuma CRUD Operations

| Operation | Description                                |
| --------- | ------------------------------------------ |
| Create    | New incident with `pin=true`, `active=true`|
| Read      | Active via status page data; history paginated (10/page) |
| Edit      | Updates title, content, style, pin         |
| Resolve   | Sets `active=false`, `pin=false`           |
| Unpin     | Sets `pin=false` only                      |
| Delete    | Hard-delete                                |

---

## What Peekaping Has Today

- **No incident module exists** — zero incident-related code in the codebase
- The public status page (`/status/:slug`) renders: icon, title, description, overall status, monitor cards with heartbeat bars, footer, and auto-refresh countdown
- There is no place for manually authored messages or announcements

### Current Module Pattern

Every module in Peekaping follows this structure:

```
apps/server/internal/modules/<name>/
├── <name>.model.go             — domain model structs
├── <name>.dto.go               — request/response DTOs
├── <name>.repository.go        — repository interface
├── <name>.sql.repository.go    — SQL implementation (bun ORM)
├── <name>.mongo.repository.go  — MongoDB implementation
├── <name>.service.go           — Service interface + ServiceImpl
├── <name>.controller.go        — HTTP handlers
├── <name>.route.go             — Route registration
└── <name>.dig.go               — dig container DI wiring
```

### Current Public Status Page Render Order

1. Optional icon image
2. Page title
3. Optional description
4. Theme toggle button
5. Computed overall status line (All Operational / Partial Outage / Under Maintenance)
6. Monitor cards (status icon, name, 24h uptime %, heartbeat bar chart)
7. Optional footer text
8. Auto-refresh countdown + manual refresh button
9. "Powered by Peekaping" attribution

---

## What Would Need to Be Built

### 1. Database Migration

New file: `apps/server/cmd/bun/migrations/YYYYMMDDHHMMSS_add_incidents.tx.up.sql`

```sql
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY,
    status_page_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    style VARCHAR(30) NOT NULL DEFAULT 'warning',
    pin BOOLEAN NOT NULL DEFAULT true,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE
);

CREATE INDEX idx_incidents_status_page_id ON incidents(status_page_id);
CREATE INDEX idx_incidents_active_pin ON incidents(active, pin);
```

And corresponding `.down.sql`:

```sql
DROP TABLE IF EXISTS incidents;
```

### 2. Backend Module

New directory: `apps/server/internal/modules/incident/`

| File                        | Purpose                                                                 |
| --------------------------- | ----------------------------------------------------------------------- |
| `incident.model.go`         | `Model` struct (ID, StatusPageID, Title, Content, Style, Pin, Active, CreatedAt, UpdatedAt) + `UpdateModel` with pointer fields |
| `incident.dto.go`           | `CreateIncidentDTO`, `UpdateIncidentDTO`, response DTOs                 |
| `incident.repository.go`    | Repository interface: Create, FindByID, FindByStatusPageID (paginated), FindActiveByStatusPageID, Update, Delete |
| `incident.sql.repository.go`| SQL implementation using `bun` ORM with `sqlModel` struct               |
| `incident.mongo.repository.go` | MongoDB implementation (required — all modules have dual backends)   |
| `incident.service.go`       | Service interface + `ServiceImpl` — CRUD, Resolve, Unpin               |
| `incident.controller.go`    | HTTP handlers with swagger annotations                                  |
| `incident.route.go`         | Route registration — auth-protected CRUD + public read endpoint         |
| `incident.dig.go`           | `RegisterDependencies` wiring for `dig` container                       |

#### API Routes

| Method | Path                                     | Auth          | Handler              |
| ------ | ---------------------------------------- | ------------- | -------------------- |
| POST   | `/incidents`                             | AllAuth       | Create               |
| GET    | `/incidents/:id`                         | AllAuth       | FindByID             |
| PATCH  | `/incidents/:id`                         | AllAuth       | Update               |
| DELETE | `/incidents/:id`                         | AllAuth       | Delete               |
| PATCH  | `/incidents/:id/resolve`                 | AllAuth       | Resolve              |
| GET    | `/status-pages/slug/:slug/incidents`     | None (public) | FindByStatusPageSlug |

### 3. API Server Wiring

**`apps/server/cmd/api/main.go`** — add:

```go
import "peekaping/internal/modules/incident"

// in the RegisterDependencies block (~line 120):
incident.RegisterDependencies(container, internalCfg)
```

**`apps/server/internal/server.go`** — add to `ProvideServer` params:

```go
incidentRoute *incident.Route,
incidentController *incident.Controller,
```

And in the route connection block:

```go
incidentRoute.ConnectRoute(router, incidentController)
```

### 4. Swagger/OpenAPI Regeneration

Add swagger doc comments to controller methods (project uses `swaggo/gin-swagger`). Then regenerate:

```bash
swag init  # regenerates apps/server/docs/swagger.json
```

### 5. Frontend API Client Regeneration

The frontend uses `@hey-api/openapi-ts` with `apps/server/docs/swagger.json` as input (`apps/web/openapi-ts.config.ts`). After backend swagger docs regenerate:

- Run the openapi-ts codegen to produce updated `types.gen.ts`, `sdk.gen.ts`, and `@tanstack/react-query.gen.ts`
- This gives the frontend typed hooks like `useGetIncidentsBySlug`, `usePostIncidents`, etc.

### 6. Frontend — Public Status Page Changes

**File:** `apps/web/src/app/status/[slug]/page.tsx`

Two new sections needed:

1. **Active incidents banner** — between the "overall status" line (~line 259) and the monitor cards list (~line 264). Fetch active/pinned incidents for the slug and render as colored alert boxes with Markdown-rendered content.

2. **Incident history section** — between the monitor cards and the footer (~line 351). Paginated list of all incidents (including resolved), with style-colored borders, title, content, and timestamps.

This requires:

- A new `useQuery` call for `GET /status-pages/slug/:slug/incidents`
- A Markdown rendering library (e.g., `react-markdown` or `marked` + `DOMPurify`)
- New UI components: `IncidentBanner` and `IncidentHistory`

### 7. Frontend — Admin Incident Management

Two approaches:

**Option A (simpler, like Uptime Kuma):** Embed incident management inline on the status page edit form. Add an "Incidents" section in `create-edit-form.tsx` with a list of existing incidents and a "Create Incident" button that opens a dialog.

**Option B (consistent with existing patterns):** Create a full admin page at `/incidents` with its own sidebar entry (`apps/web/src/components/app-sidebar.tsx`), list view, and create/edit forms. This follows how maintenance windows are managed.

Either way, the admin UI needs:

- An incident form with: title (text input), content (textarea/Markdown editor), style (dropdown: info/warning/danger/primary/light/dark)
- A list view showing incidents grouped by status page
- Actions: Edit, Resolve, Delete

### 8. Internationalization

Add translation keys to all 46 locale files in `apps/web/src/i18n/locales/`. At minimum the `en` locale, with keys like:

- `incidents.title`, `incidents.create`, `incidents.resolve`, `incidents.delete`
- `incidents.style.*` (info, warning, danger, etc.)
- `incidents.no_active`, `incidents.history`
- `status.incidents` (for the public page section headers)

### 9. Status Page Cascade Deletion

When a status page is deleted, its incidents must also be deleted. The SQL migration handles this via `ON DELETE CASCADE`, but the MongoDB path requires explicit cleanup in `status_page.service.go`'s `Delete` method — similar to how it already cleans up `monitor_status_page` and `domain_status_page` records.

---

## Summary of Effort by Layer

| Layer               | Files to Create     | Files to Modify                |
| ------------------- | ------------------- | ------------------------------ |
| Database migration  | 2 (up + down SQL)   | 0                              |
| Backend module      | 9 (model, dto, repo interface, sql repo, mongo repo, service, controller, route, dig) | 2 (main.go, server.go) |
| Swagger/API         | 0 (annotations in controller) | regenerate docs         |
| Frontend API        | 0 (auto-generated)  | regenerate client              |
| Frontend public page| 2-3 components       | 1 (status page.tsx)           |
| Frontend admin      | 3-5 pages/components | 1 (sidebar)                   |
| i18n                | 0                   | 46 locale files                |
| **Total**           | **~16-19 new files** | **~50 modified files**         |

The backend module is the most mechanical part — it follows an established pattern exactly. The frontend public page rendering and the Markdown support are the areas requiring the most design decisions.

---

## References

- [Uptime Kuma - GitHub](https://github.com/louislam/uptime-kuma)
- [Uptime Kuma incident model](https://github.com/louislam/uptime-kuma/blob/master/server/model/incident.js)
- [Uptime Kuma DB schema](https://github.com/louislam/uptime-kuma/blob/master/db/knex_init_db.js)
- [Uptime Kuma status page socket handlers](https://github.com/louislam/uptime-kuma/blob/master/server/socket-handlers/status-page-socket-handler.js)
- [Feature Request: Multiple Incidents - Issue #2726](https://github.com/louislam/uptime-kuma/issues/2726)
- [Feature Request: Better Incident System - Issue #5967](https://github.com/louislam/uptime-kuma/issues/5967)
