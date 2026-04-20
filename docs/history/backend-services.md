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
- Regular `ReadFile` method cannot access special files# About Me Implementation

## Overview

The About Me feature allows users to have a dedicated "About Me" page that is separate from regular articles.

## Location

- **Handler**: `internal/handler/aboutme.go`
- **Route**: `GET /api/v1/special/aboutme`
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
1. Call `/api/v1/special/aboutme` when navigating to the About Me page
2. Display the Markdown content using the same renderer as articles
3. The title is always "About Me" (extracted from filename)
# WebSocket Module Implementation

> 模块版本：v1.4
> 创建日期：2026-04-17
> 基于架构文档：backend-architecture.md v1.4, frontend-architecture.md v1.4

---

## 一、模块概述

WebSocket模块提供终端路径补全的实时通信功能。

### 1.1 文件结构

```
internal/server/websocket.go    - WebSocket Handler（HTTP升级、消息处理）
internal/service/completion.go  - Completion Service（路径补全逻辑）
frontend/components/command/CommandPrompt.tsx - 前端WebSocket客户端
```

### 1.2 WebSocket端点

- **端点**：`/ws/terminal`
- **协议**：WebSocket (JSON消息)
- **功能**：路径补全

---

## 二、消息格式

### 2.1 路径补全

**请求**：
```json
{
  "type": "completion_request",
  "dir": "/",
  "prefix": "RE"
}
```

**响应**：
```json
{
  "type": "completion_response",
  "items": ["README.md", "tech/"]
}
```

**补全规则**：
- 文件不带斜杠（`README.md`）
- 文件夹带斜杠（`tech/`）
- 过滤 `_` 开头的特殊文件

---

## 三、前端实现

### 3.1 WebSocket连接

- 自动连接：组件初始化时建立连接
- 断线重连：3秒后自动重连
- 消息格式：JSON

### 3.2 命令历史

- **存储**：localStorage（key: `terminalog_command_history`）
- **容量**：最多100条
- **导航**：ArrowUp/ArrowDown键

### 3.3 Tab补全

- 命令补全：本地（`search`, `open`, `cd`, `help`）
- 路径补全：WebSocket实时获取

---

## 四、技术选型

| 组件 | 技术选型 |
|------|---------|
| WebSocket库 | `gorilla/websocket` v1.5.3 |
| 消息格式 | JSON |
| 超时时间 | 5秒（补全），60秒（连接） |

---

## 五、使用示例

### 5.1 路径补全

```typescript
// 前端发送补全请求
ws.send(JSON.stringify({
  type: 'completion_request',
  dir: '/',
  prefix: 'RE'
}));

// 接收响应
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'completion_response') {
    console.log('补全结果:', data.items);
  }
};
```

### 5.2 搜索

搜索不通过 WebSocket 发送。当前实现使用 REST API `GET /api/v1/search?q=...`。

---

**文档结束**
