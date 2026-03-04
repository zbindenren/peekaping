# Incidents Feature — Implementation Plan

## Overview

Add manually-authored incident announcements to public status pages, similar to Uptime Kuma but with support for multiple active incidents per status page. Incidents let admins communicate service disruptions to visitors via the public status page. Each incident has a title, Markdown content, a style (severity level), and lifecycle states (active → resolved).

## Current State Analysis

- **No incident module exists** — zero incident-related code in the codebase
- The public status page (`apps/web/src/app/status/[slug]/page.tsx`) renders monitors, heartbeats, uptime, footer, and auto-refresh — no place for announcements
- No Markdown rendering library is installed in the frontend
- Backend follows a consistent module pattern (model, dto, repository, sql.repository, mongo.repository, service, controller, route, dig)
- The `maintenance` module is the cleanest reference pattern

### Key Discoveries:
- Module wiring: `RegisterDependencies` in `main.go:102-125`, Route+Controller injected into `ProvideServer` in `server.go:56-83`
- Public routes go BEFORE `middleware.AllAuth()` in route files (see `status_page.route.go:24-27`)
- Migrations: `YYYYMMDDHHMMSS_description.tx.{up,down}.sql` in `apps/server/cmd/bun/migrations/`
- Frontend API client auto-generated from swagger via `openapi-ts` (`npm run gen` in `apps/web`)
- Swagger docs generated via `swag init -d cmd/api,internal` in `apps/server`
- `status_page.service.go:321-334` Delete method cleans up `monitor_status_page` but not `domain_status_page` — incidents need similar MongoDB-path cleanup
- 46 i18n locale JSON files in `apps/web/src/i18n/locales/`

## Desired End State

After implementation:
1. A new `incidents` table exists in the database
2. Full CRUD backend module at `apps/server/internal/modules/incident/`
3. Public endpoint `GET /status-pages/slug/:slug/incidents` returns incidents for a status page
4. The public status page shows active incidents as colored banners above monitor cards, and a history section below monitors
5. Admin can manage incidents via `/incidents` in the dashboard (list, create, edit, resolve, delete)
6. Markdown content is rendered safely (XSS-sanitized) on the public page
7. When a status page is deleted, its incidents are cleaned up (SQL via CASCADE, MongoDB via explicit deletion)
8. All 46 locale files have incident-related translation keys

### Verification:
- Backend compiles: `cd apps/server && go build ./...`
- Swagger regenerates: `cd apps/server && swag init -d cmd/api,internal`
- Frontend builds: `cd apps/web && npm run build`
- API client regenerates: `cd apps/web && npm run gen`

## What We're NOT Doing

- **Automated incidents** triggered by monitor failures (these are manual announcements only)
- **Notification integration** (no email/webhook/Discord when incidents are created — future enhancement)
- **Incident updates/comments** (Uptime Kuma doesn't have these either; a single content field suffices for v1)
- **Custom incident severity levels** (fixed set: info/warning/danger/primary)
- **Pagination metadata** (total count, hasNext) — no existing module returns this; a project-wide concern, not incident-specific
- **Ownership/authorization checks** on CRUD — Peekaping is single-user/single-tenant; all authenticated users are fully authorized (consistent with all other modules)
- **Content length limits** on `content TEXT` — no existing module validates TEXT field length; a project-wide concern, not incident-specific
- **Rate limiting** on public endpoints — no rate limiting exists anywhere in the codebase; a project-wide concern

## Implementation Approach

Follow the existing module pattern exactly. Backend first (migration → module → wiring → swagger), then frontend (API client regen → public page → admin pages → i18n).

---

## Phase 1: Database Migration

### Overview
Create the `incidents` table with all required columns.

### Changes Required:

#### 1. Up Migration
**File**: `apps/server/cmd/bun/migrations/20260304120000_add_incidents.tx.up.sql` (new)

```sql
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY,
    status_page_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    style VARCHAR(30) NOT NULL DEFAULT 'warning',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_incidents_status_page_active ON incidents(status_page_id, active, created_at DESC);
```

#### 2. Down Migration
**File**: `apps/server/cmd/bun/migrations/20260304120000_add_incidents.tx.down.sql` (new)

```sql
DROP INDEX IF EXISTS idx_incidents_status_page_active;
DROP TABLE IF EXISTS incidents;
```

### Success Criteria:

#### Automated Verification:
- [ ] Migration applies cleanly via the bun CLI
- [ ] Rollback migration drops the table cleanly

#### Manual Verification:
- [ ] Table exists with correct columns after migration

---

## Phase 2: Backend Incident Module

### Overview
Create the full backend module following the maintenance module pattern. 9 files in `apps/server/internal/modules/incident/`.

### Changes Required:

#### 1. Domain Model
**File**: `apps/server/internal/modules/incident/incident.model.go` (new)

```go
package incident

import "time"

type Model struct {
	ID           string    `json:"id"`
	StatusPageID string    `json:"status_page_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Style        string    `json:"style"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

#### 2. DTOs
**File**: `apps/server/internal/modules/incident/incident.dto.go` (new)

```go
package incident

type CreateDto struct {
	StatusPageID string `json:"status_page_id" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Content      string `json:"content"`
	Style        string `json:"style" validate:"required,oneof=info warning danger primary"`
}

type UpdateDto struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
	Style   *string `json:"style,omitempty" validate:"omitempty,oneof=info warning danger primary"`
}
```

#### 3. Repository Interface
**File**: `apps/server/internal/modules/incident/incident.repository.go` (new)

```go
package incident

import "context"

type Repository interface {
	Create(ctx context.Context, entity *CreateDto) (*Model, error)
	FindByID(ctx context.Context, id string) (*Model, error)
	FindAll(ctx context.Context, page int, limit int, q string) ([]*Model, error)
	FindByStatusPageID(ctx context.Context, statusPageID string, page int, limit int) ([]*Model, error)
	FindActiveByStatusPageID(ctx context.Context, statusPageID string) ([]*Model, error)
	Update(ctx context.Context, id string, entity *UpdateDto) (*Model, error)
	Resolve(ctx context.Context, id string) (*Model, error)
	Delete(ctx context.Context, id string) error
	DeleteByStatusPageID(ctx context.Context, statusPageID string) error
}
```

#### 4. SQL Repository
**File**: `apps/server/internal/modules/incident/incident.sql.repository.go` (new)

Follow the pattern in `maintenance.sql.repository.go`:
- `sqlModel` struct with bun tags, table name `incidents`, alias `i`
- `toDomainModelFromSQL` converter
- `SQLRepositoryImpl` with `*bun.DB`
- UUID generation via `github.com/google/uuid`
- `Create`: insert with `Returning("*")`
- `FindByID`: select where id
- `FindAll`: paginated with optional search query `q` (matches title), ordered by `created_at DESC`
- `FindByStatusPageID`: paginated, filtered by status page, ordered by `created_at DESC`
- `FindActiveByStatusPageID`: where `active = true`, ordered by `created_at DESC`
- `Update`: partial update using Set calls per non-nil field, always set `updated_at`
- `Resolve`: set `active = false`, `updated_at = now()`
- `Delete`: hard delete by id
- `DeleteByStatusPageID`: delete all incidents for a status page (for MongoDB cleanup path)

#### 5. MongoDB Repository
**File**: `apps/server/internal/modules/incident/incident.mongo.repository.go` (new)

Follow the pattern in `maintenance.mongo.repository.go`:
- `mongoModel` with `primitive.ObjectID` and bson tags
- Collection name: `incidents`
- Create index on `status_page_id`
- All operations mirror the SQL repository

#### 6. Service
**File**: `apps/server/internal/modules/incident/incident.service.go` (new)

```go
package incident

import (
	"context"

	"go.uber.org/zap"
)

type Service interface {
	Create(ctx context.Context, dto *CreateDto) (*Model, error)
	FindByID(ctx context.Context, id string) (*Model, error)
	FindAll(ctx context.Context, page int, limit int, q string) ([]*Model, error)
	FindByStatusPageID(ctx context.Context, statusPageID string, page int, limit int) ([]*Model, error)
	FindActiveByStatusPageID(ctx context.Context, statusPageID string) ([]*Model, error)
	Update(ctx context.Context, id string, dto *UpdateDto) (*Model, error)
	Resolve(ctx context.Context, id string) (*Model, error)
	Delete(ctx context.Context, id string) error
	DeleteByStatusPageID(ctx context.Context, statusPageID string) error
}

type ServiceImpl struct {
	repository Repository
	logger     *zap.SugaredLogger
}

func NewService(repository Repository, logger *zap.SugaredLogger) Service {
	return &ServiceImpl{
		repository: repository,
		logger:     logger.Named("[incident-service]"),
	}
}
```

All service methods are thin wrappers over the repository. The `incident` package has **no dependency** on `status_page` — the FK constraint on `status_page_id` handles referential integrity at the database level, and the public slug-based endpoint is handled via a `StatusPageFinder` interface defined in this package (see Controller below).

> **Note:** WebSocket real-time broadcast for incidents is deferred to a future version. The WebSocket server requires JWT authentication, so public status page visitors (the primary audience) cannot receive these events. The frontend uses polling-based auto-refresh instead, which is sufficient for v1.

#### 7. Controller
**File**: `apps/server/internal/modules/incident/incident.controller.go` (new)

Define a `StatusPageFinder` interface in the `incident` package to avoid importing `status_page`:
```go
// StatusPageFinder resolves a status page slug to its ID.
// Implemented by status_page.Service — avoids a circular import.
type StatusPageFinder interface {
	FindBySlug(ctx context.Context, slug string) (id string, err error)
}
```

> **Note:** `status_page.Service.FindBySlug` returns `(*Model, error)`. The `incident.Controller` constructor accepts a `StatusPageFinder`, and a small adapter in `incident.dig.go` bridges `status_page.Service` to this interface (see DI Wiring below).

Swagger-annotated handlers following the maintenance controller pattern:
- `Create` — POST, bind JSON, validate, return 201
- `FindByID` — GET by `:id`, return 404 if nil
- `FindAll` — GET, paginated, optional `q` search param (admin list page)
- `FindActiveBySlug` — GET by slug (public, no auth), resolves slug via `StatusPageFinder`, returns active incidents
- `Update` — PATCH `:id`, partial update
- `Resolve` — PATCH `:id/resolve`
- `Delete` — DELETE `:id`

`FindActiveBySlug` handler:
```go
// @Router    /status-pages/slug/{slug}/incidents [get]
// @Summary   Get incidents for a status page by slug
// @Tags      Status Pages
// @Produce   json
// @Param     slug path string true "Status page slug"
// @Success   200 {object} utils.ApiResponse[[]Model]
// @Failure   404 {object} utils.APIError[any]
func (c *Controller) FindActiveBySlug(ctx *gin.Context) {
    slug := ctx.Param("slug")
    statusPageID, err := c.statusPageFinder.FindBySlug(ctx, slug)
    if err != nil || statusPageID == "" {
        ctx.JSON(http.StatusNotFound, utils.NewFailResponse("Status page not found"))
        return
    }
    incidents, err := c.service.FindActiveByStatusPageID(ctx, statusPageID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
        return
    }
    ctx.JSON(http.StatusOK, utils.NewSuccessResponse("success", incidents))
}
```

#### 8. Routes
**File**: `apps/server/internal/modules/incident/incident.route.go` (new)

```go
package incident

import (
	"peekaping/internal/modules/middleware"

	"github.com/gin-gonic/gin"
)

type Route struct {
	controller *Controller
	middleware *middleware.AuthChain
}

func NewRoute(controller *Controller, middleware *middleware.AuthChain) *Route {
	return &Route{controller: controller, middleware: middleware}
}

func (r *Route) ConnectRoute(rg *gin.RouterGroup, controller *Controller) {
	// Public route — before auth middleware (same pattern as status_page.route.go)
	sp := rg.Group("status-pages")
	sp.GET("/slug/:slug/incidents", r.controller.FindActiveBySlug)

	// Auth-protected routes
	incidents := rg.Group("incidents")
	incidents.Use(r.middleware.AllAuth())
	incidents.POST("", r.controller.Create)
	incidents.GET("", r.controller.FindAll)
	incidents.GET("/:id", r.controller.FindByID)
	incidents.PATCH("/:id", r.controller.Update)
	incidents.PATCH("/:id/resolve", r.controller.Resolve)
	incidents.DELETE("/:id", r.controller.Delete)
}
```

#### 9. DI Wiring
**File**: `apps/server/internal/modules/incident/incident.dig.go` (new)

```go
package incident

import (
	"context"

	"peekaping/internal/config"
	"peekaping/internal/modules/status_page"
	"peekaping/internal/utils"

	"go.uber.org/dig"
)

// statusPageFinderAdapter adapts status_page.Service to the StatusPageFinder interface.
type statusPageFinderAdapter struct {
	svc status_page.Service
}

func (a *statusPageFinderAdapter) FindBySlug(ctx context.Context, slug string) (string, error) {
	sp, err := a.svc.FindBySlug(ctx, slug)
	if err != nil || sp == nil {
		return "", err
	}
	return sp.ID, nil
}

func RegisterDependencies(container *dig.Container, cfg *config.Config) {
	utils.RegisterRepositoryByDBType(container, cfg, NewSQLRepository, NewMongoRepository)
	container.Provide(NewService)
	container.Provide(func(svc status_page.Service) StatusPageFinder {
		return &statusPageFinderAdapter{svc: svc}
	})
	container.Provide(NewController)
	container.Provide(NewRoute)
}
```

> **Note:** The `status_page` import is isolated to `incident.dig.go` (DI wiring only). The core incident package files (model, dto, repository, service, controller, route) have **zero** imports from `status_page`, keeping the dependency one-directional: `status_page` → `incident` for cascade deletion.

#### 10. Service Tests
**File**: `apps/server/internal/modules/incident/incident.service_test.go` (new)

Follow the pattern in `maintenance.service_test.go`:
- Mock `Repository` using `testify/mock`
- Test `Create` — happy path, repository error
- Test `FindByID` — happy path, not found
- Test `FindActiveByStatusPageID` — returns active incidents
- Test `Update` — happy path, repository error
- Test `Resolve` — happy path
- Test `Delete` — happy path, repository error
- Test `DeleteByStatusPageID` — happy path

### Success Criteria:

#### Automated Verification:
- [ ] Module compiles: `cd apps/server && go build ./...`
- [ ] No lint errors: `cd apps/server && golangci-lint run ./internal/modules/incident/...`
- [ ] Tests pass: `cd apps/server && go test ./internal/modules/incident/...`

---

## Phase 3: API Server Wiring & Swagger

### Overview
Wire the incident module into the API server and regenerate swagger docs.

### Changes Required:

#### 1. main.go
**File**: `apps/server/cmd/api/main.go`
- Add import: `"peekaping/internal/modules/incident"`
- Add after line 119 (after `domain_status_page`): `incident.RegisterDependencies(container, internalCfg)`

#### 2. server.go
**File**: `apps/server/internal/server.go`
- Add import: `"peekaping/internal/modules/incident"`
- Add to `ProvideServer` params: `incidentRoute *incident.Route, incidentController *incident.Controller,`
- Add route connection after line 124: `incidentRoute.ConnectRoute(router, incidentController)`

#### 3. Status Page Service — Cascade Deletion
**File**: `apps/server/internal/modules/status_page/status_page.service.go`
- Add `incidentRepo incident.Repository` to `ServiceImpl` struct (use `Repository` not `Service` — we only need `DeleteByStatusPageID`)
- Add `incidentRepo incident.Repository` param to `NewService`
- Refactor `Delete` method (line 321) to delete children **before** the parent to prevent orphaned records on MongoDB. The existing `monitorStatusPageService` cleanup has the same bug — fix both:
  ```go
  func (s *ServiceImpl) Delete(ctx context.Context, id string) error {
      err := s.incidentRepo.DeleteByStatusPageID(ctx, id)
      if err != nil {
          s.logger.Errorw("Failed to delete incidents for status page", "error", err, "statusPageID", id)
          return err
      }

      err = s.monitorStatusPageService.DeleteAllMonitorsForStatusPage(ctx, id)
      if err != nil {
          s.logger.Errorw("Failed to delete all monitors for status page", "error", err, "statusPageID", id)
          return err
      }

      err = s.domainStatusPageService.DeleteAllDomainsForStatusPage(ctx, id)
      if err != nil {
          s.logger.Errorw("Failed to delete all domains for status page", "error", err, "statusPageID", id)
          return err
      }

      return s.repository.Delete(ctx, id)
  }
  ```

Note: For SQL, `ON DELETE CASCADE` handles children automatically, but MongoDB needs explicit cleanup. Deleting children first ensures no orphans if a step fails. The `domainStatusPageService` cleanup fixes a pre-existing bug where `domain_status_pages` were not cleaned up on delete. The dependency direction is one-way: `status_page` → `incident`.

#### 4. Swagger Regeneration
```bash
cd apps/server && swag init -d cmd/api,internal
```

### Success Criteria:

#### Automated Verification:
- [ ] Server compiles: `cd apps/server && go build ./cmd/api`
- [ ] Swagger regenerates without errors
- [ ] `apps/server/docs/swagger.json` includes incident endpoints

---

## Phase 4: Frontend API Client Regeneration

### Overview
Copy updated swagger.json and regenerate the typed frontend API client.

### Changes Required:

```bash
cp apps/server/docs/swagger.json apps/web/swagger.json
cd apps/web && npm run gen
```

This auto-generates:
- `apps/web/src/api/types.gen.ts` — incident types
- `apps/web/src/api/sdk.gen.ts` — incident API functions
- `apps/web/src/api/@tanstack/react-query.gen.ts` — React Query hooks

### Success Criteria:

#### Automated Verification:
- [ ] `npm run gen` completes without errors
- [ ] Generated types include incident-related types
- [ ] Frontend still builds: `cd apps/web && npm run build`

---

## Phase 5: Frontend — Public Status Page

### Overview
Add active incident banners and incident history to the public status page.

### Changes Required:

#### 1. Install Markdown renderer and sanitizer
```bash
cd apps/web && npm install react-markdown rehype-sanitize
```

#### 2. Active Incident Banner Component
**File**: `apps/web/src/app/status/[slug]/components/incident-banner.tsx` (new)

A colored alert box showing:
- Style-based background color (info=blue, warning=amber, danger=red, primary=indigo)
- Title in bold
- Markdown-rendered content via `react-markdown` with `rehype-sanitize` plugin to prevent XSS
- Created/updated timestamp

#### 3. Incident History Component
**File**: `apps/web/src/app/status/[slug]/components/incident-history.tsx` (new)

A collapsible/paginated list showing all incidents (active and resolved):
- Style-colored left border
- Title, Markdown content, timestamps
- "Resolved" badge for resolved incidents
- Pagination (10 per page)

#### 4. Update Public Status Page
**File**: `apps/web/src/app/status/[slug]/page.tsx`

- Add a `useQuery` call for the public incidents endpoint (`GET /status-pages/slug/:slug/incidents`)
- Between the "Overall Status" section (line 259) and the "Monitors" section (line 265), render `<IncidentBanner>` for each active incident
- Between the monitors section (line 349) and the footer (line 351), render `<IncidentHistory>`
- Also refetch incidents on auto-refresh

### Success Criteria:

#### Automated Verification:
- [ ] Frontend builds: `cd apps/web && npm run build`
- [ ] No TypeScript errors

#### Manual Verification:
- [ ] Active incidents appear as colored banners above monitors
- [ ] Resolved incidents appear in the history section below monitors
- [ ] Markdown renders correctly (bold, links, lists, code)
- [ ] No XSS possible via Markdown content (`rehype-sanitize` strips dangerous HTML even if `rehype-raw` is added later)
- [ ] Incidents refresh on auto-refresh cycle

**Implementation Note**: After completing this phase, pause for manual verification before proceeding.

---

## Phase 6: Frontend — Admin Incident Management

### Overview
Create admin pages for managing incidents, following the maintenance admin pages pattern.

### Changes Required:

#### 1. Add Sidebar Entry
**File**: `apps/web/src/components/app-sidebar.tsx`
- Add after the Status Pages entry (line 53):
  ```tsx
  {
    title: t("navigation.incidents"),
    url: "/incidents",
    icon: AlertTriangle, // from lucide-react
  },
  ```
- Add `AlertTriangle` to lucide-react imports

#### 2. Incident List Page
**File**: `apps/web/src/app/incidents/page.tsx` (new)

Follow the pattern of `apps/web/src/app/maintenance/page.tsx`:
- Page title, "New Incident" button
- List of incidents as cards, grouped or filtered by status page
- Each card shows: title, style badge, status (active/resolved), status page name, timestamps
- Actions: Edit, Resolve, Delete

#### 3. Incident Card Component
**File**: `apps/web/src/app/incidents/components/incident-card.tsx` (new)

Card displaying incident info with action buttons.

#### 4. Create Incident Page
**File**: `apps/web/src/app/incidents/new/page.tsx` (new)

#### 5. Edit Incident Page
**File**: `apps/web/src/app/incidents/edit/page.tsx` (new)

#### 6. Create/Edit Form Component
**File**: `apps/web/src/app/incidents/components/create-edit-form.tsx` (new)

Form fields:
- **Status Page** — dropdown selector (required)
- **Title** — text input (required)
- **Content** — textarea with "Markdown supported" hint
- **Style** — dropdown: Info, Warning, Danger, Primary

Use `react-hook-form` with `zod` validation, following the maintenance form pattern.

### Success Criteria:

#### Automated Verification:
- [ ] Frontend builds: `cd apps/web && npm run build`
- [ ] No TypeScript errors

#### Manual Verification:
- [ ] "Incidents" appears in the sidebar
- [ ] Can create a new incident for a status page
- [ ] Incident list shows all incidents with correct info
- [ ] Can edit an incident (title, content, style)
- [ ] Can resolve an incident
- [ ] Can delete an incident
- [ ] Created incident appears on the public status page

**Implementation Note**: After completing this phase, pause for manual verification before proceeding.

---

## Phase 7: Internationalization

### Overview
Add incident-related translation keys to all 46 locale files.

### Changes Required:

Add the following keys to `en-US.json` (and the same English text as fallback to all other 45 locale files):

```json
{
  "navigation": {
    "incidents": "Incidents"
  },
  "incidents": {
    "title": "Incidents",
    "create": "Create Incident",
    "edit": "Edit Incident",
    "resolve": "Resolve",
    "resolved": "Resolved",
    "delete": "Delete Incident",
    "active": "Active",
    "no_incidents": "No incidents",
    "confirm_delete": "This will permanently delete this incident.",
    "confirm_resolve": "This will mark the incident as resolved.",
    "created_success": "Incident created successfully",
    "updated_success": "Incident updated successfully",
    "resolved_success": "Incident resolved successfully",
    "deleted_success": "Incident deleted successfully",
    "form": {
      "title_label": "Title",
      "title_placeholder": "Incident title",
      "content_label": "Content",
      "content_placeholder": "Describe the incident...",
      "markdown_supported": "Markdown is supported",
      "style_label": "Style",
      "status_page_label": "Status Page",
      "select_status_page": "Select a status page..."
    },
    "style": {
      "info": "Info",
      "warning": "Warning",
      "danger": "Danger",
      "primary": "Primary"
    }
  },
  "status": {
    "incidents": {
      "active_title": "Active Incidents",
      "history_title": "Incident History",
      "no_active": "No active incidents",
      "no_history": "No incident history"
    }
  }
}
```

### Success Criteria:

#### Automated Verification:
- [ ] Frontend builds: `cd apps/web && npm run build`
- [ ] All 46 locale files are valid JSON

---

## Phase 8: Status Page Cascade Deletion (MongoDB path)

### Overview
Ensure incidents are cleaned up when a status page is deleted via MongoDB.

This is already handled in Phase 3 by adding `incidentRepo.DeleteByStatusPageID` to the status page Delete method. The SQL path is covered by `ON DELETE CASCADE`. This phase is just a verification step.

### Success Criteria:

#### Manual Verification:
- [ ] Deleting a status page (SQL backend) removes its incidents
- [ ] Deleting a status page (MongoDB backend) removes its incidents

---

## Summary of Files

| Layer | Files to Create | Files to Modify |
| --- | --- | --- |
| Database migration | 2 (up + down SQL) | 0 |
| Backend module | 9 (model, dto, repo, sql.repo, mongo.repo, service, controller, route, dig) | 0 |
| Server wiring | 0 | 3 (main.go, server.go, status_page.service.go) |
| Swagger/API | 0 | regenerate docs (3 files) |
| Frontend API | 0 | regenerate client (3 files) |
| Frontend public page | 2 components | 1 (page.tsx) |
| Frontend admin | 5 pages/components | 1 (app-sidebar.tsx) |
| i18n | 0 | 46 locale files |
| **Total** | **~18 new files** | **~54 modified files** |

## References

- Research document: `docs/research-incidents.md`
- Reference module: `apps/server/internal/modules/maintenance/`
- Public status page: `apps/web/src/app/status/[slug]/page.tsx`
- Status page routes: `apps/server/internal/modules/status_page/status_page.route.go`
- API wiring: `apps/server/cmd/api/main.go`, `apps/server/internal/server.go`
