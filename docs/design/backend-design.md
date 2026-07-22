# Terminalog - 后端架构设计文档

> 文档版本：v2.1
> 创建日期：2026-04-15
> 最后更新：2026-04-18
> 基于需求文档：requirements.md v1.5
> 关联文档：frontend-architecture.md, api-spec.md, architecture.md

---

## 一、架构概览

### 1.1 后端定位

Terminalog 后端是一个 **Go HTTP + WebSocket 服务**，提供以下核心功能：
- RESTful API（文章列表、内容、搜索、目录树、About Me、版本号）
- **WebSocket API（路径补全实时通信，v1.4新增）**
- 静态资源服务（前端页面，通过 embed）
- Git Smart HTTP 服务（Git Clone/Push）
- 图片资源服务（从 Git 仓库读取）
- **特殊文件过滤**：以 `_` 开头的文件不出现在列表中
- **版本号自动生成**：基于行数变化计算语义版本号

### 1.2 后端架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         后端服务                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                         HTTP Server                             │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │ │
│  │  │  Static      │  │  REST API    │  │  Git Smart HTTP      │ │ │
│  │  │  Handler     │  │  Handler     │  │  Handler             │ │ │
│  │  │  (embed)     │  │              │  │                      │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │ │
│  │  ┌──────────────────────────────────────────────────────────┐ │ │
│  │  │  WebSocket Handler (v1.4新增)                             │ │ │
│  │  │  - 路径补全实时通信                                        │ │ │
│  │  │  - 端点: /ws/terminal                                      │ │ │
│  │  └──────────────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                         Service Layer                           │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │ │
│  │  │  Article     │  │  Git         │  │  File                │ │ │
│  │  │  Service     │  │  Service     │  │  Service             │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │ │
│  │  │  Asset       │  │  Auth        │  │  Config              │ │ │
│  │  │  Service     │  │  Service     │  │  Manager             │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │ │
│  │  ┌──────────────────────────────────────────────────────────┐ │ │
│  │  │  WebSocket Service (v1.4新增)                             │ │ │
│  │  │  - 连接管理                                                │ │ │
│  │  │  - 路径补全请求处理                                        │ │ │
│  │  └──────────────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                         Data Layer                              │ │
│  │  ┌──────────────────────────────────────────────────────────┐ │ │
│  │  │                    Git Repository                          │ │ │
│  │  │  (用户指定的内容目录 + Git 历史)                             │ │ │
│  │  └──────────────────────────────────────────────────────────┘ │ │
│  │  ┌──────────────────────────────────────────────────────────┐ │ │
│  │  │                    TOML Config File                        │ │ │
│  │  └──────────────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 二、模块划分与边界定义

### 2.1 后端模块总览

| 模块 | 职责边界 | 依赖 |
|------|---------|------|
| HTTP / WebSocket Server | 路由、连接、静态前端 | chi, gorilla/websocket |
| Article Service | 当前 HEAD 文章索引及其派生视图 | Git Service |
| About Me / Asset Service | 当前 HEAD 特殊文件和二进制资源 | Git Service |
| Version / Completion Service | 版本计算、路径补全 | Git / Article Service |
| Git Service | HEAD tree/blob、历史、差异、Smart HTTP、发布事务 | go-git + 系统 Git |
| Auth Service | Push 认证 | Config Manager |
| Config Manager | TOML 配置 | go-toml |

### 2.2 模块依赖关系

```
HTTP / WebSocket Handlers
          │
          ├── Article / Version / Completion
          ├── About Me / Asset
          └── Git Smart HTTP
                       │
                       ▼
                  Git Service
          ┌────────────┴────────────┐
          ▼                         ▼
     go-git 只读模型             系统 Git 子进程
    HEAD/历史/差异              clone/fetch/push
          └────────────┬────────────┘
                       ▼
                 Git Repository
```

公开内容不经过服务器工作区读取；工作区只服务于 `updateInstead` 的非 bare 发布流程。

---

## 三、模块职责详解

### 3.1 HTTP Server

**职责**：
- 监听 HTTP 请求，路由分发
- 静态资源服务（前端页面，通过 embed）
- API 请求转发到对应 Handler
- Git Smart HTTP 协议路由
- 请求日志记录

**边界**：
- 不处理业务逻辑，仅负责路由分发
- 不直接操作 Git 仓库
- 不负责数据解析

**接口契约**：
```go
// internal/server/server.go

type Server struct {
    addr    string
    router  *chi.Mux
    handlers *Handlers
    logger  *slog.Logger
}

func NewServer(addr string, handlers *Handlers) *Server

func (s *Server) Start() error

func (s *Server) Stop(ctx context.Context) error

func (s *Server) setupRoutes()
```

### 3.1.1 WebSocket Server

WebSocket Server 只管理 `/ws/terminal` 连接、JSON 消息和生命周期。补全结果由 Completion Service 从 Article Service 的当前 HEAD 文章索引派生，不直接访问仓库或工作区。

### 3.2 Article Service

**职责**：
- 从 Git Service 枚举当前 HEAD 下的 Markdown 文件
- 一次 commit walk 批量生成文章元数据
- 从同一文章集合派生直接目录、目录树、搜索和补全输入
- 从 HEAD blob 读取正文；不读取服务器工作区
- 维护带独立 TTL 和 generation 的内容缓存

**边界**：
- 不处理 Markdown 渲染、Git push、图片或认证
- 未提交路径与不存在路径统一视为不可见

**关键接口**：
```go
func (s *ArticleService) ListArticles(ctx context.Context, opts ListOptions) ([]model.Article, error)
func (s *ArticleService) ListDirectory(ctx context.Context, dir string, sort, order string) ([]model.Article, error)
func (s *ArticleService) GetArticle(ctx context.Context, path string) (*model.ArticleDetail, error)
func (s *ArticleService) GetTimeline(ctx context.Context, path string) ([]model.CommitInfo, error)
func (s *ArticleService) GetTree(ctx context.Context, dir string) (*model.TreeNode, error)
func (s *ArticleService) Search(ctx context.Context, query, dir string) ([]model.SearchResult, error)
```

### 3.3 Git Service

**职责**：
- 查询当前 HEAD 中的文件可见性、文件 Git 历史和版本差异
- 实现 Git Smart HTTP 协议（**使用系统 git 子进程**）：
  - `git-upload-pack --stateless-rpc`（Clone/Fetch）
  - `git-receive-pack --stateless-rpc`（Push）
  - `git {service} --stateless-rpc --advertise-refs`（refs advertisement）
- 将 receive-pack、工作区更新、go-git 重载和业务缓存失效组织为一次发布事务

**架构说明**：
- Smart HTTP 的写路径和 clone/fetch 使用系统 Git；go-git/v5 仅用于提交遍历、历史和差异等只读业务查询
- 非 bare 内容仓库固定使用 `receive.denyCurrentBranch=updateInstead`；禁止用 push 后 `reset --hard` 修补工作区
- receive-pack 响应先缓冲，完成仓库重载和缓存失效后再返回客户端
- receive-pack 串行执行，避免多个 push 的发布阶段交错
- 当前 HEAD tree 是公开内容的可见性边界；历史上出现过但当前已删除的路径不可见
- Git RPC 单独延长读写 deadline；不使用未被处理器协作遵循的全局 context timeout

**边界**：
- 负责提供 HEAD tree/blob 这一基础读能力，但不负责文章业务组合
- 不负责文章业务逻辑（Article Service 负责）
- 不负责用户管理（Auth Service 负责）

**关键接口**：
```go
func (s *GitService) GetInfoRefs(service string) ([]byte, error)
func (s *GitService) ServiceRPC(service string, reqBody io.Reader, respWriter io.Writer) error
func (s *GitService) ReceivePack(reqBody io.Reader, onRepoUpdate func()) ([]byte, error)
func (s *GitService) ReloadRepo() error
func (s *GitService) CurrentHead() (string, error)
func (s *GitService) GetFileHistory(ctx context.Context, path string) (*model.FileHistory, error)
func (s *GitService) GetFileHistories(ctx context.Context, paths []string) (map[string]*model.FileHistory, error)
func (s *GitService) NodeTypeAtHead(path string) (model.NodeType, error)
func (s *GitService) ReadFileAtHead(path string) ([]byte, error)
func (s *GitService) ListMarkdownFilesAtHead(dir string) ([]string, error)
```

### 3.3.1 Completion Service

Completion Service 不持有文件系统或 Git 依赖，只从 Article Service 的发布文章集合计算候选项。文件补全不带斜杠，目录补全带斜杠；特殊文件与资源目录在 Git Service 枚举 HEAD 时已被过滤。



### 3.4 Auth Service

**职责**：
- 解析 TOML 配置中的用户认证信息
- 验证 HTTP Basic Auth（用户名 + 密码）
- 首次启动时自动生成默认用户（admin + 随机密码）
- 密码哈希处理（bcrypt）

**边界**：
- 不负责 Git 协议处理
- 不负责配置文件整体解析（Config Manager 负责）
- 不负责权限管理（简单模型，用户即可 push）

**接口契约**：
```go
// internal/service/auth.go

type AuthService interface {
    // 验证用户认证信息
    Validate(username, password string) (bool, error)
    
    // 获取所有用户列表
    GetUsers() []User
    
    // 生成默认用户（首次启动时）
    GenerateDefaultUser() (*User, error)
    
    // 哈希密码
    HashPassword(password string) (string, error)
    
    // 验证密码哈希
    VerifyPassword(hashedPassword, password string) bool
}

type User struct {
    Username     string
    PasswordHash string // bcrypt 哈希值
}

type AuthInfo struct {
    Username string
    Password string
}
```

### 3.5 Asset Service

**职责**：
- 处理图片等静态资源请求
- 从当前 HEAD blob 读取图片文件
- 设置正确的 Content-Type
- 安全路径校验

**边界**：
- 不负责 Markdown 文件读取
- 不负责前端静态资源（由 embed 处理）
- 不负责认证（资源公开）

**接口契约**：
```go
// internal/service/asset.go

type AssetService interface {
    // 获取图片资源
    GetAsset(ctx context.Context, path string) (*Asset, error)
}

type Asset struct {
    Data        []byte
    ContentType string // MIME 类型
    Size        int64
	ETag        string
}
```

### 3.6 Config Manager

**职责**：
- 解析 TOML 配置文件
- 提供配置项访问接口
- 配置变更时保存文件
- 配置验证

**边界**：
- 不负责业务逻辑
- 不负责用户认证验证（Auth Service 负责）

**接口契约**：
```go
// internal/config/manager.go

type ConfigManager interface {
    // 加载配置文件
    Load(path string) error
    
    // 获取内容目录
    GetContentDir() string
    
    // 获取用户列表
    GetUsers() []User
    
    // 保存配置
    Save(path string) error
    
    // 验证配置
    Validate() error
}

type Config struct {
    Blog struct {
        ContentDir string `toml:"content_dir"`
    } `toml:"blog"`
    Auth struct {
        Users []UserConfig `toml:"users"`
    } `toml:"auth"`
}

type UserConfig struct {
    Username string `toml:"username"`
    Password string `toml:"password"` // 可为明文或哈希
}
```

---

## 四、技术选型

### 4.1 核心技术栈

| 组件 | 推荐方案 | 版本 | 理由 |
|------|---------|------|------|
| HTTP 路由 | **chi** | v5 | 轻量、RESTful 友好、兼容 net/http |
| Git Smart HTTP | **系统 git 子进程** | 系统版本 | 原生 packfile 处理，delta/gzip/大文件支持，Gitea架构 |
| Git 读操作 | **go-git/v5** | v5 | 纯 Go commit遍历和文件历史查询，无需系统 git |
| TOML 解析 | **pelletier/go-toml/v2** | v2 | 性能好，API 简洁 |
| 日志 | **slog** | Go 1.21+ 标准库 | 结构化日志，无额外依赖 |
| 密码哈希 | **bcrypt** | 标准库 | 安全、简单 |

### 4.2 Go 版本要求

最低版本：**Go 1.21**（支持 slog 结构化日志）

---

## 五、项目结构

```
terminalog/
├── cmd/
│   └── terminalog/
│       └── main.go              # 入口文件
│
├── internal/
│   ├── config/
│   │   ├── config.go            # 配置结构定义
│   │   ├── loader.go            # 配置加载
│   │   └── manager.go           # 配置管理
│   │
│   ├── service/
│   │   ├── article.go           # Article Service 实现
│   │   ├── git.go               # Git Service 实现
│   │   ├── asset.go             # Asset Service 实现
│   │   └── auth.go              # Auth Service 实现
│   │
│   ├── handler/
│   │   ├── article.go           # 文章 API Handler
│   │   ├── asset.go             # 资源 API Handler
│   │   ├── git.go               # Git Smart HTTP Handler
│   │   ├── search.go            # 搜索 API Handler
│   │   ├── tree.go              # 目录树 API Handler
│   │   └── static.go            # 静态资源 Handler (embed)
│   │
│   ├── model/
│   │   ├── article.go           # Article 数据结构
│   │   ├── commit.go            # CommitInfo 数据结构
│   │   ├── tree.go              # TreeNode 数据结构
│   │   ├── user.go              # User 数据结构
│   │   └── errors.go            # 错误定义
│   │
│   └── server/
│       ├── server.go            # HTTP Server 主逻辑
│       └── router.go            # 路由注册
│       └── middleware/
│           ├── logging.go       # 日志中间件
│           ├── auth.go          # 认证中间件
│           └── recovery.go      # 错误恢复中间件
│
├── pkg/
│   └── embed/
│       └── static/
│       │   └── ...              # 前端构建产物（embed）
│       └── embed.go             # embed 定义
│
│   └── utils/
│       ├── path.go              # 路径处理工具
│       ├── mime.go              # MIME 类型工具
│       └── response.go          # HTTP 响应工具
│
├── frontend/
│   └── ...                      # 前端源码（独立子项目）
│
├── docs/
│   ├── requirements.md          # 需求文档
│   ├── frontend-architecture.md # 前端架构文档
│   ├── backend-architecture.md  # 后端架构文档（本文件）
│   └── api-spec.md              # API 接口文档
│
├── configs/
│   └ config.toml.example        # 配置示例
│
├── go.mod
├── go.sum
├── Makefile                     # 构建脚本
└── README.md
```

---

## 六、核心流程设计

### 6.1 启动流程

```
解析参数与配置
    → 确保内容目录为 Git 仓库
    → NewGitService：打开 go-git 读视图并配置 updateInstead
    → 构造 Article/Auth/Asset/AboutMe/Version/Completion Service
    → 构造 Handler 与 HTTP Server
    → readiness 检查 Git HEAD 读能力
    → 启动并监听优雅关闭信号
```

Git 配置失败属于启动失败，避免服务以“能访问但不能正确发布”的降级状态运行。

### 6.2 文章列表获取流程

```
GET /api/v1/articles/{dir}
    → Article Handler 解析排序
    → Article Service 查询缓存
    → Git Service 从当前 HEAD tree 枚举 Markdown
    → 单次 commit walk 批量构建历史
    → 派生直接文件与子目录并排序
    → 按 generation 写缓存
    → JSON Response
```

正文、目录、搜索、树和补全共享同一个发布文章集合，不再分别扫描工作区。

### 6.3 Git Smart HTTP 流程

#### 6.3.1 Git Clone (upload-pack)

```
GET /info/refs?service=git-upload-pack
    → Git Handler 校验 service
    → GitService.GetInfoRefs(upload-pack)
    → 返回 refs advertisement

POST /git-upload-pack
    → Git Handler 校验 Content-Type
    → GitService.ServiceRPC(upload-pack) 流式传输 packfile
```

Clone/Fetch 是公开读操作，由系统 Git 完整处理 wire protocol。

#### 6.3.2 Git Push (receive-pack)

```
GET /info/refs?service=git-receive-pack
    → Basic Auth
    → GitService.GetInfoRefs(receive-pack)

POST /git-receive-pack
    → Basic Auth 与 Content-Type 校验
    → 串行执行 GitService.ReceivePack
    → 系统 Git 原子更新 ref 与工作区
    → 重新打开并验证 go-git 读视图
    → 推进 Article Cache generation 并清空缓存
    → 最后写出缓冲的 receive-pack 响应
```

因此客户端看到 `git push` 成功时，新提交已经能被文章、目录、搜索、补全、About Me 与资源接口读取。

## 七、Handler 实现

### 7.1 Article Handler

Article Handler 解码通配路径并处理三类资源：

- 目录：调用 `ResolveNodeType` 与 `ListDirectory`。
- Markdown：调用 `GetArticle`。
- `/timeline`、`/version` 后缀：分别调用 Article / Version Service。

路径不存在、未提交或不属于公开 Markdown 时统一返回 404；非法路径返回 400。Handler 不执行文件系统探测，所有可见性都由当前 HEAD 读模型决定。

### 7.2 Git Handler

Git Handler 只负责 HTTP 协议边界，不自行解析 packfile：

- `InfoRefs` 校验 `service`，receive-pack 额外执行 Basic Auth，然后调用 `GitService.GetInfoRefs`。
- `UploadPack` 将请求体和响应流交给 `GitService.ServiceRPC`，clone/fetch 可以流式传输。
- `ReceivePack` 校验认证与 Content-Type 后调用 `GitService.ReceivePack`。该调用返回前必须完成工作区更新、go-git 重载和 Article Cache 失效；Handler 此后才写出缓冲的协议响应。
- 失败响应不伪装成 Git 成功包；服务端日志保留系统 Git 的 stderr 以便定位拒绝 push 的原因。

这种边界把“push 返回成功”的语义固定为“新的内容读视图已经发布”。

### 7.3 静态资源 Handler（embed）

```go
// pkg/embed/embed.go

package embed

import "embed"

//go:embed static/*
var StaticFS embed.FS

// GetStaticFS 返回 embed 的静态文件系统
func GetStaticFS() embed.FS {
    return StaticFS
}
```

```go
// internal/handler/static.go

package handler

import (
    "net/http"
    "strings"
    
    "terminalog/pkg/embed"
)

type StaticHandler struct {
    fs http.FileSystem
}

func NewStaticHandler() *StaticHandler {
    return &StaticHandler{
        fs: http.FS(embed.GetStaticFS()),
    }
}

func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 处理路径
    path := r.URL.Path
    
    // 根路径 → index.html
    if path == "/" || path == "" {
        path = "/static/index.html"
    } else {
        // 其他路径 → static 目录
        path = "/static" + path
    }
    
    // 尝试直接访问文件
    f, err := h.fs.Open(path)
    if err == nil {
        f.Close()
        http.FileServer(h.fs).ServeHTTP(w, r)
        return
    }
    
    // 文件不存在，尝试 .html 扩展名（Next.js trailingSlash）
    if !strings.HasSuffix(path, ".html") {
        path = path + ".html"
        f, err = h.fs.Open(path)
        if err == nil {
            f.Close()
            http.FileServer(h.fs).ServeHTTP(w, r)
            return
        }
    }
    
    // 最终尝试 index.html（SPA fallback）
    http.ServeFile(w, r, "/static/index.html")
}
```

---

## 八、Service 实现

### 8.1 Article Service 实现

Article Service 只依赖 Git Service。列表先调用 `ListMarkdownFilesAtHead` 与 `GetFileHistories`，目录、树、搜索和补全再从这批已发布文章派生；正文调用 `ReadFileAtHead`。因此没有“文件系统存在但 HEAD 不可见”的第二套判定。

目录列表不会再次遍历 Git 历史：它按请求目录截取文章相对路径，将直接文件原样返回，并用最近编辑的子文章元数据代表直接子目录。

### 8.2 Git Service 实现

Git Service 使用两套能力但只保留一个仓库真相源：

- 系统 Git 负责 Smart HTTP、packfile、ref 与非 bare 工作区更新。
- go-git 负责 HEAD tree/blob、提交遍历、文件历史、版本差异等只读业务查询。
- 构造时强制设置 `receive.denyCurrentBranch=updateInstead`；配置失败即启动失败。
- `ReceivePack` 使用互斥锁串行化完整发布事务，先缓冲系统 Git 响应，再重新打开并验证 go-git 视图，最后触发业务缓存失效。
- `ReloadRepo` 先成功打开新视图再加锁替换，避免失败时破坏仍可用的读视图。
- 文章、About Me 与资源内容直接读取 HEAD blob；工作区修改不会进入公开页面。

不在应用层重新实现 Git wire protocol，也不在 push 后执行 `reset --hard` 或全量 `git gc`。前者容易产生协议兼容问题，后两者会扩大成功响应与实际可见内容之间的竞态窗口。

## 九、路由设计

### 9.1 Router 实现

```go
// internal/server/router.go

package server

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    
    "terminalog/internal/handler"
)

type Handlers struct {
    Article *handler.ArticleHandler
    Asset   *handler.AssetHandler
    Git     *handler.GitHandler
    Search  *handler.SearchHandler
    Tree    *handler.TreeHandler
    Static  *handler.StaticHandler
}

func (s *Server) setupRoutes() {
    r := chi.NewRouter()
    
    // 全局中间件
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(s.loggingMiddleware)
    r.Use(middleware.Recoverer)
    
    // API 路由
    r.Route("/api", func(r chi.Router) {
        // 文章 API
        r.Get("/articles", h.Article.List)
        r.Get("/articles/{path}", h.Article.Get)
        r.Get("/articles/{path}/timeline", h.Article.Timeline)
        
        // 目录树 API
        r.Get("/tree", h.Tree.Get)
        
        // 搜索 API
        r.Get("/search", h.Search.Search)
        
        // 资源 API
        r.Get("/assets/{path}", h.Asset.Get)
    })
    
    // Git Smart HTTP 路由
    r.Get("/info/refs", h.Git.UploadPackRefs)
    r.Post("/git-upload-pack", h.Git.UploadPack)
    r.Get("/info/refs", h.Git.ReceivePackRefs) // 需区分 service 参数
    r.Post("/git-receive-pack", h.Git.ReceivePack)
    
    // 静态资源（前端页面）
    r.Handle("/*", h.Static)
    
    s.router = r
}
```

---

## 十、安全设计

### 10.1 认证与授权

| 操作 | 认证要求 | 授权逻辑 |
|------|---------|---------|
| Blog 页面访问 | 无 | 公开 |
| 文章 API (GET) | 无 | 公开 |
| 资源 API (GET) | 无 | 公开 |
| Git Clone (upload-pack) | 无 | 公开 |
| Git Push (receive-pack) | Basic Auth | 仅配置文件中用户 |

### 10.2 安全措施

```go
// pkg/utils/path.go

package utils

import (
    "path/filepath"
    "strings"
)

// ValidatePath 防止目录遍历攻击
func ValidatePath(baseDir, requestedPath string) (string, error) {
    // 清理路径
    fullPath := filepath.Join(baseDir, requestedPath)
    fullPath = filepath.Clean(fullPath)
    
    // 确保路径在 baseDir 内
    if !strings.HasPrefix(fullPath, filepath.Clean(baseDir)) {
        return "", fmt.Errorf("path traversal attempt detected")
    }
    
    // 防止访问 .git 目录
    if strings.Contains(requestedPath, ".git") {
        return "", fmt.Errorf("access to .git directory denied")
    }
    
    return fullPath, nil
}
```

### 10.3 文件类型限制

```go
// pkg/utils/mime.go

package utils

import (
    "mime"
    "path/filepath"
)

var allowedExtensions = map[string]bool{
    ".md":   true,
    ".png":  true,
    ".jpg":  true,
    ".jpeg": true,
    ".gif":  true,
    ".svg":  true,
    ".webp": true,
}

func IsAllowedExtension(ext string) bool {
    return allowedExtensions[ext]
}

func GetMimeType(path string) string {
    ext := filepath.Ext(path)
    return mime.TypeByExtension(ext)
}
```

---

## 十一、性能优化

### 11.1 缓存设计

Article Cache 同时使用两种边界：

- 每个条目保存独立过期时间，写入其他 key 不会延长旧条目的 TTL。
- 全局 `generation` 表示内容仓库快照。读取流程开始时记录 generation，写缓存时必须仍与当前 generation 相同。
- Git push 发布完成时清空所有条目并推进 generation。

这保证 push 前已开始的慢请求即使在失效后才结束，也不能把旧文章列表、目录树、文章元数据或时间线重新写回缓存。

### 11.2 单次历史扫描

文章列表、目录、搜索和路径补全共享 Article Service 构建的“当前 HEAD 可见文章集合”。一次列表构建先扫描 Markdown 路径，再调用 `GetFileHistories` 进行一次 commit walk，并从同一批历史结果派生所有文章元数据。

这比“每个文件启动一个 goroutine 并分别遍历完整提交历史”更稳定：后者只是并行放大重复 I/O 和对象解析，仓库越大越容易产生 CPU、内存尖峰。缓存命中后则直接复用按目录和排序参数保存的结果。

## 十二、构建流程

### 12.1 Makefile

```makefile
# Makefile

.PHONY: all frontend backend build clean run test

# 默认目标
all: build

# 前端构建
frontend:
	cd frontend && npm install && npm run build
	cp -r frontend/out/* pkg/embed/static/

# 后端构建（不包含前端）
backend:
	go build -o bin/terminalog cmd/terminalog/main.go

# 完整构建（包含前端 embed）
build: frontend backend

# 运行
run:
	go run cmd/terminalog/main.go --port 8080 --config config.toml

# 测试
test:
	go test -v ./...

# 清理
clean:
	rm -rf frontend/out
	rm -rf bin/*
	go clean

# 跨平台构建
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/terminalog-linux-amd64 cmd/terminalog/main.go

build-darwin-arm:
	GOOS=darwin GOARCH=arm64 go build -o bin/terminalog-darwin-arm64 cmd/terminalog/main.go

build-darwin-amd:
	GOOS=darwin GOARCH=amd64 go build -o bin/terminalog-darwin-amd64 cmd/terminalog/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/terminalog-windows-amd64.exe cmd/terminalog/main.go

# 发布（所有平台）
release: frontend build-linux build-darwin-arm build-darwin-amd build-windows
```

---

## 十三、错误处理

### 13.1 错误定义

```go
// internal/model/errors.go

package model

import "errors"

var (
    ErrNotFound      = errors.New("resource not found")
    ErrNotCommitted  = errors.New("file not committed to git")
    ErrInvalidPath   = errors.New("invalid path")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
)
```

### 13.2 HTTP 响应工具

```go
// pkg/utils/response.go

package utils

import (
    "encoding/json"
    "net/http"
)

func respondJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{
        "error": message,
    })
}
```

---

## 十四、风险与缓解

### 14.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| go-git Smart HTTP 实现复杂 | Git 协议兼容性问题 | 先实现基础功能；可 fallback 到调用系统 git 命令 |
| Git 历史查询性能 | 大仓库查询慢 | 单次 commit walk + 按快照缓存 |
| 跨平台路径处理 | Windows 路径分隔符问题 | 统一使用 filepath 包 |

### 14.2 架构风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 无数据库 | 元数据查询性能依赖 Git | 缓存；限制仓库规模 |
| 单进程 | 无法水平扩展 | 文档说明；明确 MVP 边界 |
| Git 仓库损坏 | 数据丢失 | 建议用户定期备份 |

---

## 十五、后续迭代规划

### 15.1 MVP（当前版本）

- ✅ REST API（文章、目录树、搜索、资源）
- ✅ Git Smart HTTP（upload-pack, receive-pack）
- ✅ 静态资源服务（embed）
- ✅ Basic Auth 认证
- ✅ Git 历史查询

### 15.2 后续迭代

| 功能 | 优先级 | 说明 |
|------|--------|------|
| SSH Git 协议 | 中 | 需要 SSH Server 实现 |
| Git hook 支持 | 低 | post-receive hook |
| WebSocket 支持 | 低 | 实时更新通知 |
| 缓存优化 | 中 | 更智能的缓存策略 |

---

**文档结束**

> 本后端架构设计基于 requirements.md v1.1
> 关联文档：frontend-architecture.md, api-spec.md
> 下一步：进入实现阶段（Coder 模式）
