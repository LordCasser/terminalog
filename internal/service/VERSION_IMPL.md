# Version Service Implementation

## Overview

The Version Service provides semantic version calculation for articles based on Git history and line changes.

## Location

- **Service**: `internal/service/version.go`
- **Handler**: Integrated into `internal/handler/article.go`
- **Route**: `GET /api/v1/articles/{path}/version`

## Version Number Rules

Based on requirements v1.2:

| Change Type | Threshold | Version Bump |
|------------|-----------|--------------|
| Patch      | < 10 lines | +0.0.1 |
| Minor      | 10% ~ 50% of total lines | +0.1.0 (patch reset) |
| Major      | > 50% of total lines | +1.0.0 (minor/patch reset) |

- Initial version: `v1.0.0`
- Version history is stored in reverse order (newest first)

## API Response

```json
{
  "currentVersion": "v2.0.48",
  "history": [
    {
      "version": "v2.0.48",
      "hash": "a1b2c3d",
      "author": "John Doe",
      "timestamp": "2024-01-15T10:30:00Z",
      "linesChanged": 5,
      "changeType": "patch"
    }
  ]
}
```

## Implementation Notes

### MVP Version

The current implementation uses estimated line changes based on commit position in history. For production use, this should be enhanced to:

1. Extract actual file content from each commit
2. Calculate real diff between versions
3. Count actual lines added/removed/modified

### Constants

```go
PatchThreshold = 10              // lines
MinorThresholdPercent = 0.5      // 50%
InitialVersion = "v1.0.0"
```

## Dependencies

- `ArticleService`: For article metadata
- `GitService`: For commit history
- `FileService`: For current file content
