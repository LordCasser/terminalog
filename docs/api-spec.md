# Terminalog - API 接口文档

> 文档版本：v2.1
> 创建日期：2026-04-15
> 最后更新：2026-04-18
> 基于需求文档：requirements.md v1.5
> 关联文档：frontend-architecture.md, backend-architecture.md, architecture.md

---

## 一、API 概述

### 1.1 基本信息

| 项目 | 说明 |
|------|------|
| **协议** | HTTP/1.1 + WebSocket |
| **数据格式** | JSON |
| **字符编码** | UTF-8 |
| **认证方式** | Basic Auth（仅 Git Push） |
| **API风格** | RESTful |

### 1.2 基础 URL

```
http://{host}:{port}
ws://{host}:{port}  (WebSocket)
```

示例：
```
http://localhost:8080
ws://localhost:8080
http://blog.example.com
ws://blog.example.com
```

### 1.3 API 资源分类

| 类别 | 基础路径 | 说明 |
|------|---------|------|
| **Articles** | `/api/v1/articles` | 文章资源（核心） |
| **Assets** | `/api/v1/assets` | 静态资源（图片） |
| **Tree** | `/api/v1/tree` | 目录树结构 |
| **Special Pages** | `/api/v1/special` | 特殊页面（AboutMe等） |
| **Settings** | `/api/v1/settings` | 前端配置 |
| **WebSocket** | `/ws` | 实时通信 |
| **Git Smart HTTP** | `/` | Git协议端点 |

### 1.4 RESTful 设计原则

| 原则 | 说明 |
|------|------|
| 资源导向 | URL表示资源，不表示动作 |
| HTTP方法语义 | GET读取、POST创建、PUT更新、DELETE删除 |
| 层级路径 | `/api/v1/articles/{path}/timeline` 表示文章的子资源 |
| 统一响应 | 成功返回数据，失败返回`{error, message}` |
| **路径参数** | 资源标识直接嵌入URL路径，不使用GET参数 |
| **前缀分离** | 不同类型资源使用不同前缀 |

### 1.5 URL前缀分类

| 类型 | 前缀 | 用途 | 示例 |
|------|------|------|------|
| **API端点** | `/api/v1/` | 后端API | `/api/v1/articles`, `/api/v1/assets` |
| **图片资源** | `/api/v1/assets/` | 文章引用的图片 | `/api/v1/assets/guides/images/photo.jpg` |
| **静态资源** | `/api/v1/resources/` | 前端编译产物(JS/CSS) | `/api/v1/resources/_next/static/chunks/main.js` |
| **页面路由** | `/` | 前端页面URL | `/article/guides/image-test.md` |

### 1.6 URL设计对比

| 场景 | ❌ 错误方式（GET参数） | ✅ RESTful方式（路径参数） |
|------|----------------------|--------------------------|
| 获取文章 | `/article?path=guides%2Fimage-test.md` | `/article/guides/image-test.md` |
| 获取图片 | `/api/assets?file=images/photo.jpg` | `/api/v1/assets/guides/images/photo.jpg` |
| 获取JS | `/_next/static/chunks/main.js` | `/api/v1/resources/_next/static/chunks/main.js` |

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
| `201 Created` | 创建成功 |
| `204 No Content` | 删除成功（无响应体） |
| `400 Bad Request` | 参数错误 |
| `401 Unauthorized` | 认证失败 |
| `403 Forbidden` | 权限不足 |
| `404 Not Found` | 资源不存在 |
| `500 Internal Server Error` | 服务器内部错误 |

### 2.4 时间格式

所有时间字段使用 **ISO 8601** 格式：
```
2024-01-15T10:30:00Z
```

### 2.5 URL编码规则

路径参数中的文件路径需要进行URL编码：
```
/api/v1/articles/tech%2Fgolang.md
```

---

## 三、Articles API（文章资源）

**资源路径**：`/api/v1/articles`

### 3.1 获取文章列表

---

**端点**：`GET /api/v1/articles`

**描述**：获取指定目录下的文章列表

**Query Parameters**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `dir` | string | 否 | `""`（根目录） | 目录路径 |
| `sort` | string | 否 | `"edited"` | 排序字段：`created`、`edited` |
| `order` | string | 否 | `"desc"` | 排序方向：`asc`、`desc` |

**请求示例**：

```bash
# 获取根目录文章列表
curl http://localhost:8080/api/v1/articles

# 获取 tech 目录的文章
curl http://localhost:8080/api/v1/articles?dir=tech

# 按创建时间升序排序
curl "http://localhost:8080/api/v1/articles?sort=created&order=asc"
```

**响应示例（200 OK）**：

```json
{
  "articles": [
    {
      "path": "tech/golang.md",
      "name": "golang.md",
      "title": "golang",
      "type": "file",
      "createdAt": "2024-01-10T09:00:00Z",
      "createdBy": "developer1",
      "editedAt": "2024-01-15T10:30:00Z",
      "editedBy": "developer2",
      "contributors": ["developer1", "developer2"],
      "latestCommit": "Update golang guide"
    }
  ],
  "currentDir": "",
  "total": 1
}
```

---

### 3.2 搜索文章

---

**端点**：`GET /api/v1/articles/search`

**描述**：搜索文章标题

**Query Parameters**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `q` | string | 是 | - | 搜索关键词 |
| `dir` | string | 否 | `""` | 限定搜索范围 |

**请求示例**：

```bash
curl "http://localhost:8080/api/v1/articles/search?q=golang"
curl "http://localhost:8080/api/v1/articles/search?q=golang&dir=tech"
```

**响应示例（200 OK）**：

```json
{
  "results": [
    {
      "path": "tech/golang.md",
      "title": "golang",
      "matchedTitle": "golang"
    }
  ],
  "total": 1
}
```

---

### 3.3 获取单个文章

---

**端点**：`GET /api/v1/articles/{path}`

**描述**：获取单篇文章的内容和元数据

**Path Parameters**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 文件路径（URL编码） |

**请求示例**：

```bash
curl "http://localhost:8080/api/v1/articles/tech/golang.md"
curl "http://localhost:8080/api/v1/articles/guides%2Fimage-test.md"
```

**响应示例（200 OK）**：

```json
{
  "article": {
    "path": "tech/golang.md",
    "name": "golang.md",
    "title": "golang",
    "type": "file",
    "createdAt": "2024-01-10T09:00:00Z",
    "createdBy": "developer1",
    "editedAt": "2024-01-15T10:30:00Z",
    "editedBy": "developer2",
    "contributors": ["developer1", "developer2"],
    "latestCommit": "Update golang guide"
  },
  "content": "# Golang Guide\n\nThis is a guide..."
}
```

**错误响应**：

| 状态码 | 错误 | 说明 |
|--------|------|------|
| 404 | `"Article not found"` | 文件不存在 |
| 400 | `"File not committed"` | 文件未提交到Git |

---

### 3.4 获取文章时间线

---

**端点**：`GET /api/v1/articles/{path}/timeline`

**描述**：获取文章的Git提交历史

**请求示例**：

```bash
curl "http://localhost:8080/api/v1/articles/tech/golang.md/timeline"
```

**响应示例（200 OK）**：

```json
{
  "commits": [
    {
      "hash": "a1b2c3d",
      "author": "developer2",
      "message": "Update golang guide",
      "timestamp": "2024-01-15T10:30:00Z",
      "linesAdded": 12,
      "linesDeleted": 3
    }
  ]
}
```

---

### 3.5 获取文章版本信息

---

**端点**：`GET /api/v1/articles/{path}/version`

**描述**：获取文章版本号及历史版本列表

**请求示例**：

```bash
curl "http://localhost:8080/api/v1/articles/tech/golang.md/version"
```

**响应示例（200 OK）**：

```json
{
  "version": {
    "current": "v2.0.48",
    "changeType": "patch",
    "baseLines": 200,
    "currentLines": 212,
    "changePercent": 6
  },
  "history": [
    {
      "version": "v2.0.48",
      "hash": "a1b2c3d",
      "author": "developer2",
      "timestamp": "2024-01-15T10:30:00Z",
      "message": "Update golang guide",
      "linesAdded": 12,
      "linesDeleted": 3,
      "changeType": "patch"
    }
  ]
}
```

---

## 四、Assets API（资源文件）

**资源路径**：`/api/v1/assets`

### 4.1 获取静态资源

---

**端点**：`GET /api/v1/assets/{path}`

**描述**：从Git仓库获取图片等静态资源

**Path Parameters**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 资源路径（不含`.assets`层级） |

**请求示例**：

```bash
# 原始文件路径: guides/.assets/images/salvation.jpg
# API路径: guides/images/salvation.jpg
curl "http://localhost:8080/api/v1/assets/guides/images/salvation.jpg"
```

**响应示例（200 OK）**：

```
Content-Type: image/jpeg
Content-Length: 15399857

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
| 403 | `"Access denied"` | 路径非法 |

---

## 五、Tree API（目录树）

**资源路径**：`/api/v1/tree`

### 5.1 获取目录树

---

**端点**：`GET /api/v1/tree`

**描述**：获取指定目录的树形结构

**Query Parameters**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `dir` | string | 否 | `""` | 目录路径 |

**请求示例**：

```bash
curl "http://localhost:8080/api/v1/tree"
curl "http://localhost:8080/api/v1/tree?dir=tech"
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
          }
        ]
      }
    ]
  }
}
```

---

## 六、Special Pages API（特殊页面）

**资源路径**：`/api/v1/special`

### 6.1 获取About Me页面

---

**端点**：`GET /api/v1/special/aboutme`

**描述**：获取`_ABOUTME.md`文件内容

**请求示例**：

```bash
curl "http://localhost:8080/api/v1/special/aboutme"
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
| 404 | `"About Me not found"` | `_ABOUTME.md`不存在 |

---

## 七、Settings API（配置）

**资源路径**：`/api/v1/settings`

### 7.1 获取前端配置

---

**端点**：`GET /api/v1/settings`

**描述**：获取前端可访问的配置信息

**请求示例**：

```bash
curl "http://localhost:8080/api/v1/settings"
```

**响应示例（200 OK）**：

```json
{
  "owner": "lordcasser",
  "title": "Terminalog",
  "description": "A Brutalist Compiler-style terminal blog"
}
```

---

## 八、Health Check API

**资源路径**：`/api/v1/healthz`, `/api/v1/readyz`, `/api/v1/livez`, `/api/v1/status`

### 8.1 健康检查端点

| 端点 | 说明 |
|------|------|
| `GET /api/v1/healthz` | 服务健康状态 |
| `GET /api/v1/readyz` | 服务就绪状态 |
| `GET /api/v1/livez` | 服务存活状态 |
| `GET /api/v1/status` | 详细状态信息 |

---

## 九、WebSocket API

**端点**：`ws://localhost:8080/ws/terminal`

### 9.1 路径补全

**请求格式**：
```json
{
  "type": "completion_request",
  "dir": "",
  "prefix": "go"
}
```

**参数说明**：
- `dir`: 搜索范围目录
  - 空字符串 `""`：全局搜索（`search`命令使用），匹配所有级别的路径名
  - 具体路径如 `"tech"`：当前目录搜索（`open`/`cd`命令使用），只匹配直接子项
- `prefix`: 路径前缀匹配

**响应格式**：
```json
{
  "type": "completion_response",
  "items": ["tech/golang/", "tech/golang/go-guide.md"]
}
```

**返回格式**：
- 全局搜索（dir为空）：返回完整路径（如 `tech/golang/`、`tech/golang/go-guide.md`）
- 当前目录搜索（dir指定）：返回basename（如 `golang/`、`go-guide.md`）
- 文件不带斜杠，文件夹带斜杠

### 9.2 搜索文章

**请求格式**：
```json
{
  "type": "search_request",
  "keyword": "terminal"
}
```

**响应格式**：
```json
{
  "type": "search_response",
  "results": [
    {"path": "README.md", "title": "Terminalog"}
  ]
}
```

---

## 十、Git Smart HTTP API

**资源路径**：`/api/v1/git`

> **Git Clone URL**: `http://{host}:{port}/api/v1/git/`
> 例如：`http://localhost:8080/api/v1/git/`

### 10.1 协议概述

| 操作 | 端点 | 认证 |
|------|------|------|
| Clone/Fetch | `GET /api/v1/git/info/refs?service=git-upload-pack` | 无 |
| Clone/Fetch | `POST /api/v1/git/git-upload-pack` | 无 |
| Push | `GET /api/v1/git/info/refs?service=git-receive-pack` | Basic Auth |
| Push | `POST /api/v1/git/git-receive-pack` | Basic Auth |

### 10.2 Git Clone

```bash
# Clone仓库
git clone http://localhost:8080/api/v1/git/ blog-content
```

### 10.3 Git Push

```bash
# Push更新
git push http://localhost:8080/api/v1/git/ main

Username: admin
Password: [配置文件中的密码]
```

---

## 十一、数据模型

### 11.1 Article

```typescript
interface Article {
  path: string;          // 文件路径
  name: string;          // 文件名
  title: string;         // 标题（去除.md）
  type: 'file' | 'dir';  // 类型
  createdAt: string;     // 创建时间
  createdBy: string;     // 创建人
  editedAt: string;      // 编辑时间
  editedBy: string;      // 编辑人
  contributors: string[];// 贡献者
  latestCommit: string;  // 最新提交信息
}
```

### 11.2 ArticleDetail

```typescript
interface ArticleDetail {
  article: Article;
  content: string;       // Markdown内容
}
```

### 11.3 CommitInfo

```typescript
interface CommitInfo {
  hash: string;          // Commit hash（7位）
  author: string;        // 作者
  message: string;       // 提交信息
  timestamp: string;     // 时间
  linesAdded: number;    // 新增行数
  linesDeleted: number;  // 删除行数
}
```

### 11.4 VersionInfo

```typescript
interface VersionInfo {
  current: string;       // 当前版本号
  changeType: string;    // 变更类型
  baseLines: number;     // 基础行数
  currentLines: number;  // 当前行数
  changePercent: number; // 变化百分比
}
```

### 11.5 TreeNode

```typescript
interface TreeNode {
  name: string;
  path: string;
  type: 'dir' | 'file';
  children: TreeNode[] | null;
}
```

### 11.6 SearchResult

```typescript
interface SearchResult {
  path: string;
  title: string;
  matchedTitle: string;
}
```

### 11.7 ErrorResponse

```typescript
interface ErrorResponse {
  error: string;
  message?: string;
}
```

---

## 十二、前端调用示例

### 12.1 API Client

```typescript
// lib/api/client.ts

const API_BASE = ''; // 同源请求

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
  }

  async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown' }));
      throw new Error(error.error);
    }
    return response.json();
  }

  async getText(path: string): Promise<string> {
    const response = await fetch(`${this.baseUrl}${path}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.text();
  }
}

export const apiClient = new ApiClient();
```

### 12.2 Articles API

```typescript
// lib/api/articles.ts

export async function getArticles(params?: {
  dir?: string;
  sort?: 'created' | 'edited';
  order?: 'asc' | 'desc';
}): Promise<ArticleListResponse> {
  const query = new URLSearchParams();
  if (params?.dir) query.set('dir', params.dir);
  if (params?.sort) query.set('sort', params.sort);
  if (params?.order) query.set('order', params.order);
  
  return apiClient.get(`/api/v1/articles?${query}`);
}

export async function searchArticles(q: string, dir?: string): Promise<SearchResponse> {
  const query = new URLSearchParams();
  query.set('q', q);
  if (dir) query.set('dir', dir);
  
  return apiClient.get(`/api/v1/articles/search?${query}`);
}

export async function getArticle(path: string): Promise<ArticleDetail> {
  return apiClient.get(`/api/v1/articles/${encodeURIComponent(path)}`);
}

export async function getArticleTimeline(path: string): Promise<{ commits: CommitInfo[] }> {
  return apiClient.get(`/api/v1/articles/${encodeURIComponent(path)}/timeline`);
}

export async function getArticleVersion(path: string): Promise<{ version: VersionInfo; history: VersionHistoryEntry[] }> {
  return apiClient.get(`/api/v1/articles/${encodeURIComponent(path)}/version`);
}
```

### 12.3 Assets API

```typescript
// lib/api/assets.ts

export function getAssetUrl(path: string): string {
  // 图片路径转换规则：
  // Markdown: ./assets/images/photo.jpg (实际在 guides/.assets/images/photo.jpg)
  // API URL: /api/v1/assets/guides/images/photo.jpg (去除.assets层级)
  return `/api/v1/assets/${path}`;
}
```

### 12.4 图片路径转换

```typescript
// lib/utils/image-path.ts

/**
 * 转换Markdown图片路径为API资源路径
 * 
 * 规则：
 * 1. 外部链接(http/https): 保持原样
 * 2. 绝对路径(/...): 添加/api/v1/assets前缀
 * 3. 相对路径: 去除.assets层级，添加basePath和/api/v1/assets前缀
 */
export function transformImagePath(src: string, basePath?: string): string {
  // 外部链接
  if (src.startsWith('http://') || src.startsWith('https://')) {
    return src;
  }
  
  // 绝对路径
  if (src.startsWith('/')) {
    return `/api/v1/assets${src}`;
  }
  
  // 相对路径
  let normalized = src.startsWith('./') ? src.slice(2) : src;
  
  // 去除.assets层级
  if (normalized.startsWith('.assets/')) {
    normalized = normalized.slice(8);
  }
  
  // 组合完整路径
  const fullPath = basePath ? `${basePath}/${normalized}` : normalized;
  return `/api/v1/assets/${fullPath}`;
}

// 使用示例
// Markdown: ![](./.assets/images/photo.jpg)
// basePath: "guides" (从文章路径提取)
// 结果: /api/v1/assets/guides/images/photo.jpg
```

---

## 十三、API 端点总览

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/articles` | GET | 获取文章列表 |
| `/api/v1/articles/search` | GET | 搜索文章 |
| `/api/v1/articles/{path}` | GET | 获取单个文章 |
| `/api/v1/articles/{path}/timeline` | GET | 获取时间线 |
| `/api/v1/articles/{path}/version` | GET | 获取版本信息 |
| `/api/v1/assets/{path}` | GET | 获取静态资源 |
| `/api/v1/tree` | GET | 获取目录树 |
| `/api/v1/special/aboutme` | GET | 获取About Me |
| `/api/v1/settings` | GET | 获取配置 |
| `/api/v1/healthz` | GET | 健康检查 |
| `/api/v1/readyz` | GET | 就绪检查 |
| `/api/v1/livez` | GET | 存活检查 |
| `/api/v1/status` | GET | 状态详情 |
| `/api/v1/git/info/refs` | GET | Git refs advertisement |
| `/api/v1/git/git-upload-pack` | POST | Git clone/fetch |
| `/api/v1/git/git-receive-pack` | POST | Git push |
| `/ws/terminal` | WebSocket | 终端通信 |

---

## 十五、前端路由设计

### 15.1 RESTful路由规则

前端页面路由遵循RESTful原则，资源标识嵌入URL路径：

| 页面 | 路由 | 说明 |
|------|------|------|
| 主页 | `/` | 文章列表 |
| 文章详情 | `/article/{path}` | 文章路径参数 |
| About Me | `/aboutme` | 特殊页面 |
| 目录浏览 | `/browse/{dir}` | 目录浏览 |

### 15.2 路由示例

```
文章路径: guides/image-test.md

❌ 错误方式（GET参数）:
/article?path=guides%2Fimage-test.md

✅ RESTful方式（路径参数）:
/article/guides/image-test.md
```

### 15.3 Next.js动态路由实现

```typescript
// app/article/[...slug]/page.tsx

export default function ArticlePage({ params }: { params: { slug: string[] } }) {
  // slug = ['guides', 'image-test.md']
  const articlePath = params.slug.join('/');
  
  // 使用articlePath调用API
  const article = await getArticle(articlePath);
  
  return <ArticleContent article={article} />;
}

// 生成静态路径
export async function generateStaticParams() {
  const articles = await getArticles();
  return articles.articles.map(a => ({
    slug: a.path.split('/')
  }));
}
```

### 15.4 路由导航

```typescript
// 导航到文章页面
function navigateToArticle(path: string) {
  // path = 'guides/image-test.md'
  router.push(`/article/${path}`);
}

// ArticleTable组件中的链接
<Link href={`/article/${article.path}`}>
  {article.name}
</Link>
```

---

## 十六、Debug模式说明

### 16.1 开发环境配置

在Debug模式下：
- 前端独立运行（`npm run dev`，端口3000）
- 后端独立运行（`--debug`参数）
- CORS自动启用
- 静态文件不嵌入后端

### 16.2 启动方式

```bash
# 后端（Debug模式）
./bin/terminalog --debug --port 8080

# 前端（开发服务器）
cd frontend && npm run dev

# 前端访问 http://localhost:3000
# API请求代理到 http://localhost:8080
```

### 16.3 前端API代理配置

```typescript
// frontend/next.config.ts

const config = {
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
    ];
  },
};

export default config;
```

---

## 十七、版本历史

| 版本 | 日期 | 变更内容 |
|------|------|---------|
| v2.1 | 2026-04-18 | **统一API前缀**：Git Smart HTTP移至`/api/v1/git/`，健康检查移至`/api/v1/healthz`等 |
| v2.0 | 2026-04-17 | **RESTful重构**：API添加`/api/v1/`前缀，前端路由改用路径参数 |
| v1.5 | 2026-04-17 | 添加`.assets`隐藏目录支持 |
| v1.0 | 2026-04-15 | 初始API定义 |

---

## 十八、架构变更总结

### 18.1 API变更

| 原端点 | 新端点 | 变化 |
|--------|--------|------|
| `/api/articles` | `/api/v1/articles` | 添加版本前缀 |
| `/api/articles/*` | `/api/v1/articles/{path}` | 明确路径参数 |
| `/api/search` | `/api/v1/articles/search` | 合并到Articles资源 |
| `/api/aboutme` | `/api/v1/special/aboutme` | 合并到Special资源 |
| `/api/config` | `/api/v1/settings` | 重命名为Settings |
| `/api/assets/*` | `/api/v1/assets/{path}` | 添加版本前缀 |
| `/healthz` | `/api/v1/healthz` | 移入API v1命名空间 |
| `/readyz` | `/api/v1/readyz` | 移入API v1命名空间 |
| `/livez` | `/api/v1/livez` | 移入API v1命名空间 |
| `/status` | `/api/v1/status` | 移入API v1命名空间 |
| `/info/refs` | `/api/v1/git/info/refs` | Git移入API v1命名空间 |
| `/git-upload-pack` | `/api/v1/git/git-upload-pack` | Git移入API v1命名空间 |
| `/git-receive-pack` | `/api/v1/git/git-receive-pack` | Git移入API v1命名空间 |
| `.git` (根路径) | `/api/v1/git/` | Git Clone URL统一前缀 |

### 18.2 Git URL变更

| 变更项 | 原值 | 新值 |
|--------|------|------|
| Git Clone URL | `http://localhost:8080/` | `http://localhost:8080/api/v1/git/` |
| Git Push URL | `http://localhost:8080/` | `http://localhost:8080/api/v1/git/` |

### 18.3 前端路由变更

| 原路由 | 新路由 | 变化 |
|--------|--------|------|
| `/article?path=xxx` | `/article/{path}` | 改用路径参数 |
| `/article/page.tsx` | `/article/[...slug]/page.tsx` | 改用动态路由 |

### 18.3 需要修改的文件

**后端文件：**
```
internal/server/server.go
- 更新路由定义：添加/api/v1/前缀
- 添加/api/v1/resources/路由（静态资源）
- 调整资源层级
```

**前端文件：**
```
frontend/app/article/page.tsx → frontend/app/article/[...slug]/page.tsx
- 改用动态路由捕获路径参数

frontend/lib/api/articles.ts
- 更新API路径：/api/v1/articles

frontend/lib/api/assets.ts
- 更新API路径：/api/v1/assets

frontend/components/brutalist/ArticleTable.tsx
- 更新链接href：/article/${path}

frontend/next.config.ts
- 添加API代理配置

frontend/components/article/MarkdownRenderer.tsx
- 更新图片路径转换：/api/v1/assets/
```

### 18.4 资源路径映射

| 原路径 | 新路径 | 处理方式 |
|--------|--------|----------|
| `/_next/static/...` | `/api/v1/resources/_next/static/...` | StaticHandler |
| `/api/assets/...` | `/api/v1/assets/...` | AssetHandler |
| `/article?path=...` | `/article/{path}` | 前端动态路由 |

---

**文档结束**

> 本API文档基于RESTful设计原则
> **请切换到Coder模式实现代码**