# Special File Filter Implementation

## Overview

Files starting with underscore (`_`) are treated as special files and excluded from regular article listings.

## Location

- **Service**: `internal/service/file.go`
- **Constant**: `SpecialFilePrefix = "_"`

## Filtering Logic

### ScanMarkdownFiles

Files and directories starting with `_` are automatically filtered:
- `_ABOUTME.md` - About Me page (excluded from listing)
- `_DRAFT.md` - Draft articles (excluded from listing)
- `_templates/` - Template directory (entire directory excluded)

```go
// Skip special files (starting with "_")
if strings.HasPrefix(d.Name(), SpecialFilePrefix) {
    if d.IsDir() {
        return filepath.SkipDir // Skip entire special directory
    }
    return nil // Skip special file
}
```

### ReadSpecialFile

Special files can only be read via `ReadSpecialFile` method:
- Requires filename to start with `_`
- Uses same path validation as regular files
- Returns `ErrNotFound` if file doesn't start with `_`

## Use Cases

1. **About Me Page**: `_ABOUTME.md` - Personal introduction
2. **Drafts**: `_DRAFT_*.md` - Articles in progress
3. **Templates**: `_templates/` - Reusable templates
4. **Configuration**: `_config.md` - Internal configuration

## Security

- Path traversal protection applies to special files
- Special files must still be committed to Git to be accessible
- Regular `ReadFile` method cannot access special files