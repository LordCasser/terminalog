# About Me Implementation

## Overview

The About Me feature allows users to have a dedicated "About Me" page that is separate from regular articles.

## Location

- **Handler**: `internal/handler/aboutme.go`
- **Route**: `GET /api/aboutme`
- **File**: `_ABOUTME.md` in content directory

## Special File Handling

Files starting with `_` are treated as special files:
- They are excluded from regular article listings
- They can only be accessed via specific APIs
- `_ABOUTME.md` is reserved for the About Me page

## API Response

```json
{
  "path": "_ABOUTME.md",
  "title": "About Me",
  "content": "# About Me\n\nThis is my personal blog..."
}
```

## Error Handling

- If `_ABOUTME.md` does not exist: Returns 404 Not Found
- Regular files (not starting with `_`) cannot be accessed via `ReadSpecialFile`

## Configuration

No additional configuration required. The `_ABOUTME.md` file should be placed in the root of the content directory and committed to Git.

## Frontend Integration

The frontend should:
1. Call `/api/aboutme` when navigating to the About Me page
2. Display the Markdown content using the same renderer as articles
3. The title is always "About Me" (extracted from filename)