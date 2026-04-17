# WebSocket Module Implementation

> 模块版本：v1.4
> 创建日期：2026-04-17
> 基于架构文档：backend-architecture.md v1.4, frontend-architecture.md v1.4

---

## 一、模块概述

WebSocket模块提供终端命令的实时通信功能，支持路径补全和搜索功能。

### 1.1 文件结构

```
internal/server/websocket.go    - WebSocket Handler（HTTP升级、消息处理）
internal/service/completion.go  - Completion Service（路径补全、搜索逻辑）
frontend/components/command/CommandPrompt.tsx - 前端WebSocket客户端
```

### 1.2 WebSocket端点

- **端点**：`/ws/terminal`
- **协议**：WebSocket (JSON消息)
- **功能**：路径补全、文章搜索

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

### 2.2 搜索

**请求**：
```json
{
  "type": "search_request",
  "keyword": "terminal"
}
```

**响应**：
```json
{
  "type": "search_response",
  "results": [
    {"path": "README.md", "title": "Terminalog"}
  ]
}
```

**搜索规则**：
- 搜索文章标题
- 过滤 `_` 开头的特殊文件
- 最多返回10条结果

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
| 超时时间 | 5秒（补全/搜索），60秒（连接） |

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

```typescript
// 前端发送搜索请求
ws.send(JSON.stringify({
  type: 'search_request',
  keyword: 'terminal'
}));

// 接收响应并跳转
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'search_response' && data.results.length > 0) {
    router.push(`/article?path=${data.results[0].path}`);
  }
};
```

---

**文档结束**