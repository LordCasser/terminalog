# RESTful API Migration Implementation

> Date: 2026-04-17
> Author: Architecture alignment based on current runtime routes

---

## Summary

This document records the migration from legacy `/api/*` endpoints and query-style frontend routes to the current versioned REST API and path-based page routes.

## Changes Overview

### Backend Routes (`internal/server/server.go`)

| Old Endpoint | New Endpoint | Notes |
|-------------|--------------|-------|
| `/api/articles` | `/api/v1/articles` | Added version prefix |
| `/api/articles/*` | `/api/v1/articles/*` | Path parameter routing |
| `/api/search` | `/api/v1/search` | Independent search resource |
| `/api/tree` | `/api/v1/tree` | Added version prefix |
| `/api/assets/*` | `/api/v1/assets/*` | Added version prefix |
| `/api/aboutme` | `/api/v1/special/aboutme` | Moved to Special resource |
| `/api/config` | `/api/v1/settings` | Renamed to Settings |
| - | `/api/v1/resources/*` | **New**: Frontend static resources |

### Frontend API Files (`frontend/lib/api/`)

| File | Changes |
|------|---------|
| `articles.ts` | All paths updated to `/api/v1/articles/*` |
| `aboutme.ts` | Path updated to `/api/v1/special/aboutme` |
| `search.ts` | Path updated to `/api/v1/search` |
| `tree.ts` | Path updated to `/api/v1/tree` |
| `settings.ts` | **New**: `/api/v1/settings` endpoint |

### Frontend Components

| File | Changes |
|------|---------|
| `ArticleTable.tsx` | Link href changed from `/article?path=` to `/article/{path}` |
| `MarkdownRenderer.tsx` | Image path transformation uses `/api/v1/assets/` |
| `CommandPrompt.tsx` | Router navigation uses path parameter format |

### Frontend Routes

| Old Route | New Route |
|-----------|-----------|
| `/article?path=xxx` | `/article/{path}` |
| `article/page.tsx` | `article/[...slug]/page.tsx` |

## Implementation Details

### Backend Handler Addition

Added `ServeResources` method to `StaticHandler` for serving frontend compiled resources (JS/CSS) via `/api/v1/resources/*` endpoint.

```go
func (h *StaticHandler) ServeResources(w http.ResponseWriter, r *http.Request) {
    path := chi.URLParam(r, "*")
    // Serve from embedded static directory
}
```

### Frontend Dynamic Route

Created catch-all route `[...slug]` to handle nested article paths:

```tsx
const params = useParams();
const slug = params.slug as string[];
const articlePath = slug.join("/");
```

## Testing

1. Backend build: `go build ./...` - ✅ Pass
2. Frontend build: `npm run build -- --webpack` - ✅ Pass

## References

- `docs/api-spec.md` - Complete API specification
- `docs/architecture.md` - System architecture
