# Terminalog - API 接口文档

> 文档版本：v1.2
> 创建日期：2026-04-15
> 最后更新：2026-04-16
> 基于需求文档：requirements.md v1.2
> 关联文档：frontend-architecture.md, backend-architecture.md, architecture.md

---

## 一、API 概述

### 1.1 基本信息

| 项目 | 说明 |
|------|------|
| **协议** | HTTP/1.1 |
| **数据格式** | JSON |
| **字符编码** | UTF-8 |
| **认证方式** | Basic Auth（仅 Git Push） |
| **版本** | v1（无版本号前缀） |

### 1.2 基础 URL

```
http://{host}:{port}
```

示例：
```
http://localhost:8080
http://blog.example.com
```

### 1.3 API 分类

| 类别 | 基础路径 | 说明 |
|------|---------|------|
| REST API | `/api` | 文章、搜索、目录树、资源 |
| Git API | `/`（根路径） | Git Smart HTTP 协议 |
| 静态资源 | `/`（根路径） | 前端页面（embed） |

---

## 二、通用约定

### 2.1 请求格式

| Header | 说明 |
|--------|------|
| `Content-Type` | `application/json`（POST/PUT） |
| `Accept` | `application/json` |
| `Authorization` | Basic Auth（仅 Git Push） |

### 2.2 响应格式

#### 成功响应

```json
{
  // 数据字段（根据 API 不同）
}
```

#### 错误响应

```json
{
  "error": "错误类型",
  "message": "详细描述（可选）"
}
```

### 2.3 HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| `200 OK` | 成功 |
| `400 Bad Request` | 参数错误 |
| `401 Unauthorized` | 认证失败 |
| `404 Not Found` | 资源不存在 |
| `500 Internal Server Error` | 服务器内部错误 |

### 2.4 时间格式

所有时间字段使用 **ISO 8601** 格式：
```
2024-01-15T10:30:00Z
```

### 2.5 分页

当前版本不分页，直接返回所有结果。

---

## 三、REST API

### 3.1 文章 API

---

#### 3.1.1 获取文章列表

**端点**：`GET /api/articles`

**描述**：获取指定目录下的文章列表（仅返回已提交到 Git 的 Markdown 文件，**排除**以 `_` 开头的特殊文件）

**认证**：无

**Query Parameters**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `dir` | string | 否 | `""`（根目录） | 目录路径 |
| `sort` | string | 否 | `"edited"` | 排序字段：`created`（创建时间）、`edited`（编辑时间） |
| `order` | string | 否 | `"desc"` | 排序方向：`asc`（升序）、`desc`（降序） |

**请求示例**：

```bash
# 获取根目录文章列表
curl http://localhost:8080/api/articles

# 获取 tech 目录的文章
curl http://localhost:8080/api/articles?dir=tech

# 按创建时间升序排序
curl http://localhost:8080/api/articles?sort=created&order=asc
```

**响应示例（200 OK）**：

```json
{
  "articles": [
    {
      "path": "tech/golang.md",
      "title": "golang",
      "createdAt": "2024-01-10T09:00:00Z",
      "createdBy": "developer1",
      "editedAt": "2024-01-15T10:30:00Z",
      "editedBy": "developer2",
      "contributors": ["developer1", "developer2", "developer3"]
    },
    {
      "path": "welcome.md",
      "title": "welcome",
      "createdAt": "2024-01-08T15:20:00Z",
      "createdBy": "admin",
      "editedAt": "2024-01-08T15:20:00Z",
      "editedBy": "admin",
      "contributors": ["admin"]
    }
  ],
  "currentDir": "",
  "total": 2
}
```

**错误响应**：

| 状态码 | 错误 | 说明 |
|--------|------|------|
| 500 | `"Failed to scan directory"` | 目录扫描失败 |

---

#### 3.1.2 获取文章内容

**端点**：`GET /api/articles/{path}`

**描述**：获取单篇文章的详细内容和元数据

**认证**：无

**Path Parameters**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 文件路径（如 `tech/golang.md`） |

**请求示例**：

```bash
curl http://localhost:8080/api/articles/tech/golang.md
```

**响应示例（200 OK）**：

```json
{
  "path": "tech/golang.md",
  "title": "golang",
  "content": "# Golang Guide\n\nThis is a guide for Golang...\n\n```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```",
  "createdAt": "2024-01-10T09:00:00Z",
  "createdBy": "developer1",
  "editedAt": "2024-01-15T10:30:00Z",
  "editedBy": "developer2",
  "contributors": ["developer1", "developer2"]
}
```

**错误响应**：

| 状态码 | 错误 | 说明 |
|--------|------|------|
| 400 | `"File not committed"` | 文件未提交到 Git |
| 404 | `"Article not found"` | 文件不存在 |
| 500 | `"Internal server error"` | 服务器内部错误 |

---

#### 3.1.3 获取文章时间线

**端点**：`GET /api/articles/{path}/timeline`

**描述**：获取文章的 Git 提交历史（编辑时间线）

**认证**：无

**Path Parameters**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 文件路径 |

**请求示例**：

```bash
curl http://localhost:8080/api/articles/tech/golang.md/timeline
```

**响应示例（200 OK）**：

```json
{
  "commits": [
    {
      "hash": "a1b2c3d",
      "author": "developer2",
      "timestamp": "2024-01-15T10:30:00Z"
    },
    {
      "hash": "e4f5g6h",
      "author": "developer1",
      "timestamp": "2024-01-14T15:20:00Z"
    },
    {
      "hash": "i7j8k9l",
      "author": "developer1",
      "timestamp": "2024-01-10T09:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 错误 | 说明 |
|--------|------|------|
| 404 | `"Article not found"` | 文件不存在 |
| 500 | `"Internal server error"` | Git 查询失败 |

---

### 3.2 目录树 API

---

#### 3.2.1 获取目录树

**端点**：`GET /api/tree`

**描述**：获取指定目录的树形结构

**认证**：无

**Query Parameters**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `dir` | string | 否 | `""`（根目录） | 目录路径 |

**请求示例**：

```bash
curl http://localhost:8080/api/tree
```

**响应示例（200 OK）**：

```json
{
  "tree": {
    "name": "articles",
    "path": "",
    "type": "dir",
    "children": [
      {
        "name": "tech",
        "path": "tech",
        "type": "dir",
        "children": [
          {
            "name": "golang.md",
            "path": "tech/golang.md",
            "type": "file",
            "children": null
          },
          {
            "name": "rust.md",
            "path": "tech/rust.md",
            "type": "file",
            "children": null
          }
        ]
      },
      {
        "name": "life",
        "path": "life",
        "type": "dir",
        "children": []
      },
      {
        "name": "welcome.md",
        "path": "welcome.md",
        "type": "file",
        "children": null
      },
      {
        "name": "about.md",
        "path": "about.md",
        "type": "file",
        "children": null
      }
    ]
  }
}
```

**TreeNode 结构说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 目录或文件名 |
| `path` | string | 完整路径（相对于根目录） |
| `type` | string | 类型：`"dir"` 或 `"file"` |
| `children` | array | 子节点（仅 `dir` 类型有值） |

---

### 3.3 搜索 API

---

#### 3.3.1 搜索文章

**端点**：`GET /api/search`

**描述**：搜索文章标题（简单匹配，**排除**以 `_` 开头的特殊文件）

**认证**：无

**Query Parameters**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `q` | string | 是 | - | 搜索关键词（匹配标题） |
| `dir` | string | 否 | `""`（根目录） | 限定搜索范围 |

**请求示例**：

```bash
# 搜索包含 "golang" 的文章
curl "http://localhost:8080/api/search?q=golang"

# 在 tech 目录下搜索
curl "http://localhost:8080/api/search?q=golang&dir=tech"
```

**响应示例（200 OK）**：

```json
{
  "results": [
    {
      "path": "tech/golang.md",
      "title": "golang",
      "matchedTitle": "golang"
    },
    {
      "path": "tech/golang-advanced.md",
      "title": "golang-advanced",
      "matchedTitle": "golang-advanced"
    }
  ],
  "total": 2
}
```

**错误响应**：

| 状态码 | 错误 | 说明 |
|--------|------|------|
| 400 | `"Missing search query"` | 未提供搜索关键词 |
| 500 | `"Internal server error"` | 服务器内部错误 |

---

### 3.4 资源 API

---

#### 3.4.1 获取静态资源（图片）

**端点**：`GET /api/assets/{path}`

**描述**：从 Git 仓库获取图片等静态资源

**认证**：无

**Path Parameters**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 资源路径（如 `tech/images/diagram.png`） |

**请求示例**：

```bash
curl http://localhost:8080/api/assets/tech/images/diagram.png
```

**响应示例（200 OK）**：

```
Content-Type: image/png
Content-Length: 12345

<图片二进制数据>
```

**Content-Type 映射**：

| 扩展名 | Content-Type |
|--------|-------------|
| `.png` | `image/png` |
| `.jpg`, `.jpeg` | `image/jpeg` |
| `.gif` | `image/gif` |
| `.svg` | `image/svg+xml` |
| `.webp` | `image/webp` |

**错误响应**：

| 状态码 | 错误 | 说明 |
|--------|------|------|
| 404 | `"Asset not found"` | 资源不存在 |
| 403 | `"Access denied"` | 访问被拒绝（路径非法） |

---

## 四、Git Smart HTTP API

### 4.1 协议概述

Git Smart HTTP 协议用于支持 Git Clone 和 Git Push 操作。

| 操作 | 端点 | 认证 |
|------|------|------|
| Clone/Fetch | `/info/refs` + `/git-upload-pack` | 无 |
| Push | `/info/refs` + `/git-receive-pack` | Basic Auth |

### 4.2 Git Clone (upload-pack)

---

#### 4.2.1 获取 refs advertisement

**端点**：`GET /info/refs?service=git-upload-pack`

**描述**：获取仓库 refs 信息（用于 Clone/Fetch）

**认证**：无（公开）

**Query Parameters**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `service` | string | 是 | 必须为 `git-upload-pack` |

**请求示例**：

```bash
curl "http://localhost:8080/info/refs?service=git-upload-pack"
```

**响应示例（200 OK）**：

```
Content-Type: application/x-git-upload-pack-advertisement
Cache-Control: no-cache

001e# service=git-upload-pack
0000
003fabcd1234abcd1234abcd1234abcd1234abcd1234 HEAD
003fabcd1234abcd1234abcd1234abcd1234abcd1234 refs/heads/main
0000
```

> 注：响应为 Git pkt-line 格式，实际使用时由 Git CLI 解析。

---

#### 4.2.2 执行 upload-pack

**端点**：`POST /git-upload-pack`

**描述**：处理 Git Clone/Fetch 的 pack 数据请求

**认证**：无（公开）

**Request Headers**：

| Header | 值 |
|--------|-----|
| `Content-Type` | `application/x-git-upload-pack-request` |

**Request Body**：Git pack request（pkt-line 格式）

**响应示例（200 OK）**：

```
Content-Type: application/x-git-upload-pack-result

<Git packfile>
```

**使用示例**：

```bash
# Git CLI 使用方式
git clone http://localhost:8080/repo.git

# 或
git clone http://localhost:8080/ blog-content
```

---

### 4.3 Git Push (receive-pack)

---

#### 4.3.1 获取 refs advertisement

**端点**：`GET /info/refs?service=git-receive-pack`

**描述**：获取仓库 refs 信息（用于 Push）

**认证**：**Basic Auth（必须）**

**Query Parameters**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `service` | string | 是 | 必须为 `git-receive-pack` |

**Request Headers**：

| Header | 值 |
|--------|-----|
| `Authorization` | `Basic {base64(username:password)}` |

**请求示例**：

```bash
curl -u "admin:password" \
  "http://localhost:8080/info/refs?service=git-receive-pack"
```

**成功响应（200 OK）**：

```
Content-Type: application/x-git-receive-pack-advertisement
Cache-Control: no-cache

001e# service=git-receive-pack
0000
003fabcd1234abcd1234abcd1234abcd1234abcd1234 HEAD
003fabcd1234abcd1234abcd1234abcd1234abcd1234 refs/heads/main
0000
```

**认证失败响应（401 Unauthorized）**：

```
WWW-Authenticate: Basic realm="Git"
Content-Type: application/json

{
  "error": "Unauthorized"
}
```

---

#### 4.3.2 执行 receive-pack

**端点**：`POST /git-receive-pack`

**描述**：处理 Git Push 的 pack 数据请求

**认证**：**Basic Auth（必须）**

**Request Headers**：

| Header | 值 |
|--------|-----|
| `Authorization` | `Basic {base64(username:password)}` |
| `Content-Type` | `application/x-git-receive-pack-request` |

**Request Body**：Git pack data（pkt-line 格式）

**成功响应（200 OK）**：

```
Content-Type: application/x-git-receive-pack-result

<Git receive result>
```

**认证失败响应（401 Unauthorized）**：

```
WWW-Authenticate: Basic realm="Git"
Content-Type: application/json

{
  "error": "Unauthorized"
}
```

**使用示例**：

```bash
# Git CLI 使用方式
git push http://localhost:8080/repo.git main

# Git 会提示输入用户名和密码
Username: admin
Password: [输入配置文件中的密码]
```

---

### 3.5 About Me API（v1.2 新增）

---

#### 3.5.1 获取 About Me 页面内容

**端点**：`GET /api/aboutme`

**描述**：获取 `_ABOUTME.md` 文件的内容，用于 About Me 页面渲染

**认证**：无

**请求示例**：

```bash
curl http://localhost:8080/api/aboutme
```

**响应示例（200 OK）**：

```json
{
  "path": "_ABOUTME.md",
  "title": "About Me",
  "content": "# About Me\n\nWelcome to my blog...\n"
}
```

**错误响应**：

| 状态码 | 错误 | 说明 |
|--------|------|------|
| 404 | `"About Me not found"` | `_ABOUTME.md` 文件不存在 |
| 500 | `"Internal server error"` | 服务器内部错误 |

---

### 3.6 版本号 API（v1.2 新增）

---

#### 3.6.1 获取文章版本号

**端点**：`GET /api/articles/{path}/version`

**描述**：获取文章当前版本号及历史版本列表。版本号基于行数变化自动计算。

**认证**：无

**Path Parameters**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 文件路径 |

**请求示例**：

```bash
curl http://localhost:8080/api/articles/tech/golang.md/version
```

**响应示例（200 OK）**：

```json
{
  "currentVersion": "v2.0.48",
  "history": [
    {
      "version": "v2.0.48",
      "hash": "a1b2c3d",
      "author": "developer2",
      "timestamp": "2024-01-15T10:30:00Z",
      "linesChanged": 5,
      "changeType": "patch"
    },
    {
      "version": "v2.0.47",
      "hash": "e4f5g6h",
      "author": "developer1",
      "timestamp": "2024-01-14T15:20:00Z",
      "linesChanged": 150,
      "changeType": "minor"
    },
    {
      "version": "v2.0.0",
      "hash": "i7j8k9l",
      "author": "developer1",
      "timestamp": "2024-01-10T09:00:00Z",
      "linesChanged": 500,
      "changeType": "major"
    }
  ]
}
```

**版本号计算规则**：

| 变更类型 | 行数变化 | 版本号变化 |
|---------|---------|-----------|
| `patch` | < 10 行 | 补丁版本 +1（vX.Y.Z → vX.Y.Z+1） |
| `minor` | 10~50% 总行数 | 子版本 +1（vX.Y.Z → vX.(Y+1).0） |
| `major` | > 50% 总行数 | 主版本 +1（vX.Y.Z → v(X+1).0.0） |

---

## 五、数据模型

### 5.1 Article

```typescript
interface Article {
  path: string;          // 文件路径（相对于仓库根目录）
  title: string;         // 文件名（去除 .md 扩展名）
  createdAt: string;     // 创建时间（ISO 8601）
  createdBy: string;     // 创建人
  editedAt: string;      // 最后编辑时间（ISO 8601）
  editedBy: string;      // 最后编辑人
  contributors: string[]; // 所有贡献者
}
```

### 5.2 ArticleDetail

```typescript
interface ArticleDetail extends Article {
  content: string;       // Markdown 内容
}
```

### 5.3 CommitInfo

```typescript
interface CommitInfo {
  hash: string;          // 短格式 commit hash（7 位）
  author: string;        // 作者名
  timestamp: string;     // commit 时间（ISO 8601）
}
```

### 5.4 TreeNode

```typescript
interface TreeNode {
  name: string;          // 目录或文件名
  path: string;          // 完整路径
  type: 'dir' | 'file';  // 类型
  children: TreeNode[] | null; // 子节点（仅 dir 有）
}
```

### 5.5 SearchResult

```typescript
interface SearchResult {
  path: string;          // 文件路径
  title: string;         // 标题
  matchedTitle: string;  // 匹配的标题（用于高亮）
}
```

### 5.6 ArticleListResponse

```typescript
interface ArticleListResponse {
  articles: Article[];
  currentDir: string;
  total: number;
}
```

### 5.7 TimelineResponse

```typescript
interface TimelineResponse {
  commits: CommitInfo[];
}
```

### 5.8 SearchResponse

```typescript
interface SearchResponse {
  results: SearchResult[];
  total: number;
}
```

### 5.9 TreeResponse

```typescript
interface TreeResponse {
  tree: TreeNode;
}
```

### 5.10 ErrorResponse

```typescript
interface ErrorResponse {
  error: string;
  message?: string;
}
```

---

## 六、前端调用示例

### 6.1 TypeScript/JavaScript

```typescript
// lib/api/articles.ts

const API_BASE = ''; // 同源，无需前缀

// 获取文章列表
async function getArticles(options?: {
  dir?: string;
  sort?: 'created' | 'edited';
  order?: 'asc' | 'desc';
}): Promise<ArticleListResponse> {
  const params = new URLSearchParams();
  if (options?.dir) params.set('dir', options.dir);
  if (options?.sort) params.set('sort', options.sort);
  if (options?.order) params.set('order', options.order);
  
  const response = await fetch(`${API_BASE}/api/articles?${params}`);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }
  return response.json();
}

// 获取文章内容
async function getArticle(path: string): Promise<ArticleDetail> {
  const response = await fetch(`${API_BASE}/api/articles/${encodeURIComponent(path)}`);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }
  return response.json();
}

// 获取时间线
async function getTimeline(path: string): Promise<TimelineResponse> {
  const response = await fetch(`${API_BASE}/api/articles/${encodeURIComponent(path)}/timeline`);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }
  return response.json();
}

// 搜索
async function search(query: string, dir?: string): Promise<SearchResponse> {
  const params = new URLSearchParams();
  params.set('q', query);
  if (dir) params.set('dir', dir);
  
  const response = await fetch(`${API_BASE}/api/search?${params}`);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }
  return response.json();
}

// 获取目录树
async function getTree(dir?: string): Promise<TreeResponse> {
  const params = dir ? `?dir=${encodeURIComponent(dir)}` : '';
  const response = await fetch(`${API_BASE}/api/tree${params}`);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }
  return response.json();
}
```

### 6.2 图片路径转换

```typescript
// lib/utils/path-transform.ts

function transformImagePath(src: string, basePath: string): string {
  // 外部链接保持原样
  if (src.startsWith('http://') || src.startsWith('https://')) {
    return src;
  }
  
  // 绝对路径（相对于仓库根）
  if (src.startsWith('/')) {
    return `/api/assets${src}`;
  }
  
  // 相对路径转换为 API 资源路径
  let normalized = src;
  if (normalized.startsWith('./')) {
    normalized = normalized.slice(2);
  }
  
  const fullPath = basePath ? `${basePath}/${normalized}` : normalized;
  return `/api/assets/${fullPath}`;
}

// 使用示例
// Markdown: ![](./images/photo.png)
// basePath: "tech/blog"
// 转换后: /api/assets/tech/blog/images/photo.png
```

---

## 七、Git 使用指南

### 7.1 Clone 仓库

```bash
# Clone 整个仓库
git clone http://localhost:8080/repo.git

# Clone 到指定目录
git clone http://localhost:8080/ my-blog

# 查看仓库内容
cd my-blog
ls
```

### 7.2 Push 更新

```bash
# 添加新文章
echo "# My New Article" > new-article.md
git add new-article.md
git commit -m "Add new article"

# Push 到远程
git push origin main

# 输入认证信息
Username: admin
Password: [配置文件中的密码]
```

### 7.3 更新文章

```bash
# 编辑文章
vim tech/golang.md

# 提交更新
git add tech/golang.md
git commit -m "Update golang article"
git push origin main
```

### 7.4 配置 Git

```bash
# 设置远程仓库 URL
git remote set-url origin http://localhost:8080/repo.git

# 或添加远程
git remote add origin http://localhost:8080/repo.git

# 存储认证信息（避免每次输入密码）
git config credential.helper store

# 或使用 Git credential cache
git config credential.helper cache --timeout=3600
```

---

## 八、错误处理指南

### 8.1 前端错误处理

```typescript
// 统一错误处理
async function apiCall<T>(
  url: string,
  options?: RequestInit
): Promise<T> {
  try {
    const response = await fetch(url, options);
    
    if (!response.ok) {
      const error = await response.json().catch(() => ({
        error: `HTTP ${response.status}`
      }));
      throw new ApiError(error.error, response.status);
    }
    
    return response.json();
  } catch (err) {
    if (err instanceof ApiError) {
      throw err;
    }
    throw new ApiError('Network error', 0);
  }
}

class ApiError extends Error {
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
  
  status: number;
}

// 使用示例
try {
  const articles = await getArticles({ dir: 'tech' });
} catch (err) {
  if (err instanceof ApiError) {
    if (err.status === 404) {
      console.error('Directory not found');
    } else if (err.status === 500) {
      console.error('Server error:', err.message);
    }
  }
}
```

### 8.2 错误类型映射

| 错误类型 | HTTP 状态码 | 前端处理建议 |
|---------|------------|--------------|
| `"Article not found"` | 404 | 提示"文章不存在"，返回列表页 |
| `"File not committed"` | 400 | 提示"文章尚未发布" |
| `"Unauthorized"` | 401 | Git Push 认证失败，提示重新输入密码 |
| `"Asset not found"` | 404 | 图片加载失败，显示占位图 |
| `"Internal server error"` | 500 | 提示"服务器错误"，刷新页面 |

---

## 九、API 测试

### 9.1 curl 测试脚本

```bash
#!/bin/bash
# test-api.sh

BASE_URL="http://localhost:8080"

echo "=== 测试文章列表 ==="
curl -s "$BASE_URL/api/articles" | jq .

echo "=== 测试文章内容 ==="
curl -s "$BASE_URL/api/articles/welcome.md" | jq .

echo "=== 测试时间线 ==="
curl -s "$BASE_URL/api/articles/welcome.md/timeline" | jq .

echo "=== 测试目录树 ==="
curl -s "$BASE_URL/api/tree" | jq .

echo "=== 测试搜索 ==="
curl -s "$BASE_URL/api/search?q=welcome" | jq .

echo "=== 测试资源 ==="
curl -s -I "$BASE_URL/api/assets/images/logo.png"

echo "=== 测试 Git refs ==="
curl -s "$BASE_URL/info/refs?service=git-upload-pack"
```

### 9.2 Postman 配置

**Environment 变量**：
```
BASE_URL: http://localhost:8080
```

**Collection 请求**：
- `GET {{BASE_URL}}/api/articles`
- `GET {{BASE_URL}}/api/articles/:path`
- `GET {{BASE_URL}}/api/articles/:path/timeline`
- `GET {{BASE_URL}}/api/tree`
- `GET {{BASE_URL}}/api/search?q=:query`
- `GET {{BASE_URL}}/api/assets/:path`

---

## 十、版本历史

| 版本 | 日期 | 变更内容 |
|------|------|---------|
| v1.0 | 2026-04-15 | 初始 API 定义 |

---

## 十一、后续迭代规划

### 11.1 MVP 版本（当前）

- ✅ 文章列表 API（排序支持）
- ✅ 文章内容 API
- ✅ 时间线 API
- ✅ 目录树 API
- ✅ 搜索 API（标题搜索）
- ✅ 资源 API（图片）
- ✅ Git Clone API
- ✅ Git Push API

### 11.2 后续版本

| API | 优先级 | 说明 |
|-----|--------|------|
| `/api/articles/{path}/raw` | 低 | 获取原始 Markdown（无元数据） |
| `/api/search?q=:query&type=full` | 低 | 全文搜索 |
| `/api/stats` | 低 | 仓库统计信息 |
| WebSocket `/ws` | 低 | 实时更新通知 |
| `/api/rss` | 低 | RSS/Atom 订阅 |

---

**文档结束**

> 本 API 文档基于 requirements.md v1.1
> 关联文档：frontend-architecture.md, backend-architecture.md
> 下一步：进入实现阶段（Coder 模式）