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
│  │  │  - 端点: /ws/completion                                    │ │ │
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

| 模块 | 负责人 | 职责边界 | 依赖关系 |
|------|--------|---------|---------|
| **HTTP Server** | 后端 | HTTP 路由分发，静态资源服务 | Go net/http 或 chi |
| **WebSocket Server (v1.4新增)** | 后端 | WebSocket连接管理，路径补全实时通信 | Go gorilla/websocket |
| **Article Service** | 后端 | 文章列表、内容读取、元数据获取（**过滤 `_` 开头文件**） | Git Service, File Service |
| **About Me Service** | 后端 | 读取并返回 `_ABOUTME.md` 内容（v1.2 新增） | File Service, Git Service |
| **Version Service** | 后端 | 基于行数变化计算语义版本号（v1.2 新增） | Git Service |
| **WebSocket Service (v1.4新增)** | 后端 | 路径补全请求处理，实时响应补全结果 | File Service |
| **Git Service** | 后端 | Git 历史查询，Smart HTTP 协议实现 | go-git/v5 |
| **File Service** | 后端 | 文件系统操作，目录扫描，特殊文件过滤 | Go os/fs 包 |
| **Auth Service** | 后端 | 用户认证校验，密码验证 | Config Manager |
| **Asset Service** | 后端 | 图片等静态资源读取与响应 | File Service |
| **Config Manager** | 后端 | TOML 配置文件解析与管理 | pelletier/go-toml/v2 |

### 2.2 模块依赖关系图

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Server                           │
│                     (路由分发 + 静态资源)                       │
└─────────────────────────────────────────────────────────────┘
           │              │              │              │
           │              │              │              │
           ▼              ▼              ▼              ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  Article    │  │   Asset     │  │    Git      │  │  Static     │
│  Service    │  │  Service    │  │  Service    │  │  Handler    │
│             │  │             │  │             │  │  (embed)    │
└─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘
       │                │                │
       │                │                │
       ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                      File Service                            │
│                    (文件系统操作)                              │
└─────────────────────────────────────────────────────────────┘
       │                                              │
       │                                              │
       ▼                                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Git Service                             │
│              (Git 历史查询 + Smart HTTP)                      │
└─────────────────────────────────────────────────────────────┘
       │
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Git Repository                            │
│                  (用户指定的内容目录)                          │
└─────────────────────────────────────────────────────────────┘


认证流程依赖：
┌─────────────┐
│ HTTP Server │───▶ Git Handler ───▶ Auth Service.Validate()
└─────────────┘
```

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

### 3.1.1 WebSocket Server (v1.4新增)

**职责**：
- 监听 WebSocket 连接请求（端点：`/ws/completion`）
- 管理 WebSocket 连接生命周期（连接建立、消息处理、连接关闭）
- 处理路径补全请求，实时响应补全结果
- WebSocket消息解析和序列化（JSON格式）

**边界**：
- 不处理HTTP请求（由HTTP Server负责）
- 不负责文件系统操作（由File Service负责）
- 不持久化WebSocket连接状态

**接口契约**：
```go
// internal/server/websocket.go

type WebSocketHandler struct {
    service    *CompletionService
    logger     *slog.Logger
    upgrader   websocket.Upgrader
    connections map[string]*websocket.Conn
}

func NewWebSocketHandler(service *CompletionService) *WebSocketHandler

func (h *WebSocketHandler) HandleTerminal(ws *websocket.Conn)

// WebSocket消息格式
type CompletionRequest struct {
    Type   string `json:"type"`   // "completion_request"
    Dir    string `json:"dir"`    // 当前目录路径
    Prefix string `json:"prefix"` // 路径前缀
}

type CompletionResponse struct {
    Type  string   `json:"type"`  // "completion_response"
    Items []string `json:"items"` // 补全结果列表
}

type SearchRequest struct {
    Type    string `json:"type"`    // "search_request"
    Keyword string `json:"keyword"` // 搜索关键词
}

type SearchResponse struct {
    Type    string        `json:"type"`    // "search_response"
    Results []SearchResult `json:"results"` // 搜索结果列表
}

type SearchResult struct {
    Path  string `json:"path"`  // 文章路径
    Title string `json:"title"` // 文章标题
}
```

**技术选型**：
- WebSocket库：`gorilla/websocket`（成熟稳定的Go WebSocket库）
- 消息格式：JSON（易于前端解析）
- 连接管理：map存储活跃连接（支持多客户端）

**WebSocket端点**：
- `/ws/terminal`：终端命令实时通信端点（路径补全 + 搜索）

### 3.2 Article Service

**职责**：
- 扫描 Git 仓库目录，获取文章列表
- 过滤未提交的文件（仅返回已提交的 Markdown）
- 从 Git 历史获取文章元数据（创建时间、编辑时间、贡献者）
- 读取 Markdown 文件内容
- 搜索文章标题
- 获取目录树结构

**边界**：
- 不负责 Markdown 渲染（前端负责）
- 不负责 Git push 接收（Git Service 负责）
- 不负责图片资源读取（Asset Service 负责）

**接口契约**：
```go
// internal/service/article.go

type ArticleService interface {
    // 获取文章列表（已提交的 Markdown 文件）
    ListArticles(ctx context.Context, opts ListOptions) ([]Article, error)
    
    // 获取单篇文章内容
    GetArticle(ctx context.Context, path string) (*ArticleDetail, error)
    
    // 获取文章 Git 时间线
    GetTimeline(ctx context.Context, path string) ([]CommitInfo, error)
    
    // 获取目录树结构
    GetTree(ctx context.Context, dir string) (*TreeNode, error)
    
    // 搜索文章标题
    Search(ctx context.Context, query string, dir string) ([]SearchResult, error)
}

type ListOptions struct {
    Dir   string      // 目录路径
    Sort  SortField   // 排序字段：created, edited
    Order SortOrder   // 排序方向：asc, desc
}

type Article struct {
    Path         string    // 文件路径（相对于仓库根目录）
    Title        string    // 文件名（去除 .md 扩展名）
    CreatedAt    time.Time // 创建时间（首个 commit）
    CreatedBy    string    // 创建人（首个 commit 作者）
    EditedAt     time.Time // 最后编辑时间
    EditedBy     string    // 最后编辑人
    Contributors []string  // 所有贡献者
}

type ArticleDetail struct {
    Article
    Content    string    // Markdown 内容
}

type CommitInfo struct {
    Hash      string    // 短格式 commit hash（7 位）
    Author    string    // 作者名
    Timestamp time.Time // commit 时间
}

type TreeNode struct {
    Name     string     // 目录/文件名
    Path     string     // 完整路径
    Type     NodeType   // "dir" 或 "file"
    Children []*TreeNode // 子节点（仅目录有）
}

type SearchResult struct {
    Path         string
    Title        string
    MatchedTitle string
}
```

### 3.3 Git Service

**职责**：
- 查询文件 Git 历史（创建时间、编辑时间、贡献者）
- 检查文件是否已提交到 Git
- 实现 Git Smart HTTP 协议（**使用系统 git 子进程**）：
  - `git-upload-pack --stateless-rpc`（Clone/Fetch）
  - `git-receive-pack --stateless-rpc`（Push）
  - `git {service} --stateless-rpc --advertise-refs`（refs advertisement）
- Push 后 checkout 工作目录同步
- Push 后重新加载 go-git 仓库缓存

**架构说明**：
- Smart HTTP 协议的写路径（push）和读路径（clone/fetch）使用系统 git 子进程处理
- 这遵循 Gitea/Giteria 的架构设计，利用 git 原生的 packfile 处理能力
- 系统 git 子进程正确处理：delta objects 解析、gzip 编码、大文件传输、并发安全
- go-git/v5 仅用于读操作（文件历史查询、commit 遍历等）

**边界**：
- 不负责文件内容读取（File Service 负责）
- 不负责文章业务逻辑（Article Service 负责）
- 不负责用户管理（Auth Service 负责）

**接口契约**：
```go
// internal/service/git.go

type GitService struct {
    repoPath string    // Git 仓库绝对路径
    repo     *git.Repository  // go-git 仓库实例（仅用于读操作）
}

// Smart HTTP 协议（git 子进程）
func (s *GitService) GetInfoRefs(service string) ([]byte, error)
    // 运行: git {upload-pack|receive-pack} --stateless-rpc --advertise-refs .
    // 返回 refs advertisement 数据

func (s *GitService) ServiceRPC(service string, reqBody io.Reader, respWriter io.Writer) error
    // 运行: git {upload-pack|receive-pack} --stateless-rpc .
    // 直接 pipe HTTP body 到子进程 stdin，子进程 stdout 到 HTTP response

// Push 后辅助操作
func (s *GitService) CheckoutWorkingTree() error
    // 运行: git checkout --force（同步工作目录到 HEAD）
    // git receive-pack 只更新 refs/objects，不更新工作目录

func (s *GitService) ReloadRepo() error
    // 重新打开 go-git 仓库（刷新缓存状态）

// 读操作（go-git）
func (s *GitService) GetFileHistory(ctx context.Context, filePath string) (*FileHistory, error)
func (s *GitService) IsFileCommitted(ctx context.Context, filePath string) (bool, error)
```

### 3.3.1 WebSocket Service (v1.4新增)

**职责**：
- 处理路径补全请求，实时响应补全结果
- 处理搜索请求，返回匹配的文章路径列表
- 根据目录路径和前缀匹配文章列表和子目录列表
- **过滤约束**：不返回以 `_` 开头的隐藏文件（如 `_ABOUTME.md`）
- 返回补全结果（文件不带斜杠，文件夹带斜杠）
- 支持实时补全和搜索（WebSocket低延迟通信）

**边界**：
- 不负责WebSocket连接管理（由WebSocket Handler负责）
- 不负责文件系统操作（调用File Service）
- 不持久化补全历史和搜索历史

**接口契约**：
```go
// internal/service/completion.go

type CompletionService interface {
    // 处理路径补全请求
    HandleCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    
    // 处理搜索请求
    HandleSearch(ctx context.Context, req SearchRequest) (*SearchResponse, error)
    
    // 获取目录下的匹配项（文件和文件夹）
    GetMatchingItems(ctx context.Context, dir, prefix string) ([]CompletionItem, error)
}

type CompletionRequest struct {
    Dir    string // 当前目录路径（如 "/" 或 "/tech"）
    Prefix string // 路径前缀（如 "RE" 或 "tec"）
}

type CompletionResponse struct {
    Items []CompletionItem // 补全结果列表
}

type SearchRequest struct {
    Keyword string // 搜索关键词（如 "terminal"）
}

type SearchResponse struct {
    Results []SearchResult // 搜索结果列表
}

type SearchResult struct {
    Path  string // 文章路径（如 "README.md"）
    Title string // 文章标题（去除.md扩展名）
}

type CompletionItem struct {
    Name string // 名称（如 "README.md" 或 "tech/"）
    Type string // 类型（"file" 或 "dir"）
}
```

**补全规则**：
- 文件补全不带斜杠（如 `README.md` 或完整路径 `tech/golang/go-guide.md`）
- 文件夹补全带斜杠（如 `tech/` 或完整路径 `tech/golang/`）
- 仅返回已提交的Markdown文件（过滤未提交文件）
- **过滤以 `_` 开头的特殊文件**（如 `_ABOUTME.md`）
- **全局搜索模式**（dir为空）：匹配所有级别的路径名和目录名，返回完整路径
  - 例如：`prefix="go"` 返回 `tech/golang/` 和 `tech/golang/go-guide.md`
- **当前目录模式**（dir指定）：只匹配当前目录下的直接子项，返回basename
  - 例如：dir="/tech"，prefix="go" 返回 `golang/` 和 `golang/go-guide.md`（相对于tech）

**搜索规则**：
- 搜索文章标题（去除.md扩展名后的文件名）
- **过滤以 `_` 开头的特殊文件**（如 `_ABOUTME.md`）
- 返回最匹配的文章路径列表（最多10条结果）
- 支持模糊匹配（keyword包含在标题中即可匹配）

### 3.4 File Service

**职责**：
- 扫描目录结构
- 读取文件内容
- 检查文件是否存在
- 获取文件 MIME 类型
- 安全路径校验（防止目录遍历攻击）

**边界**：
- 不负责 Git 历史查询
- 不负责认证
- 不负责业务逻辑

**接口契约**：
```go
// internal/service/file.go

type FileService interface {
    // 扫描目录，返回所有 .md 文件路径
    ScanMarkdownFiles(ctx context.Context, dir string) ([]string, error)
    
    // 读取文件内容
    ReadFile(ctx context.Context, path string) ([]byte, error)
    
    // 获取目录树结构（所有文件和子目录）
    GetDirectoryTree(ctx context.Context, dir string) (*TreeNode, error)
    
    // 检查文件是否存在
    FileExists(ctx context.Context, path string) (bool, error)
    
    // 安全路径校验（返回合法的完整路径）
    ValidatePath(requestedPath string) (string, error)
    
    // 获取文件 MIME 类型
    GetMimeType(path string) string
}
```

### 3.5 Auth Service

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

### 3.6 Asset Service

**职责**：
- 处理图片等静态资源请求
- 从 Git 仓库读取图片文件
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
}
```

### 3.7 Config Manager

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
│   │   ├── file.go              # File Service 实现
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

```go
// cmd/terminalog/main.go

package main

import (
    "context"
    "flag"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    
    "terminalog/internal/config"
    "terminalog/internal/server"
    "terminalog/internal/service"
)

func main() {
    // 1. 解析命令行参数
    flags := parseFlags()
    
    // 2. 设置日志
    logger := setupLogger(flags.logLevel)
    slog.SetDefault(logger)
    
    // 3. 加载配置文件
    cfg, err := config.Load(flags.configPath)
    if err != nil {
        slog.Info("Config file not found, creating default config")
        cfg = config.Default()
        if err := cfg.Save(flags.configPath); err != nil {
            slog.Error("Failed to save default config", "error", err)
            os.Exit(1)
        }
        slog.Info("Default config saved", "path", flags.configPath)
    }
    
    // 4. 验证配置
    if err := cfg.Validate(); err != nil {
        slog.Error("Config validation failed", "error", err)
        os.Exit(1)
    }
    
    // 5. 确保 Git 仓库目录存在且是 Git 仓库
    if err := ensureGitRepo(cfg.GetContentDir()); err != nil {
        slog.Error("Failed to initialize Git repository", "error", err)
        os.Exit(1)
    }
    
    // 6. 初始化 Auth Service（处理默认用户）
    authSvc := service.NewAuthService(cfg)
    if len(cfg.GetUsers()) == 0 {
        defaultUser, err := authSvc.GenerateDefaultUser()
        if err != nil {
            slog.Error("Failed to generate default user", "error", err)
            os.Exit(1)
        }
        cfg.AddUser(*defaultUser)
        cfg.Save(flags.configPath)
        slog.Info("Generated default user", 
            "username", defaultUser.Username,
            "password", defaultUser.PasswordHash) // 显示临时密码
    }
    
    // 7. 初始化所有 Services
    fileSvc := service.NewFileService(cfg.GetContentDir())
    gitSvc := service.NewGitService(cfg.GetContentDir())
    articleSvc := service.NewArticleService(fileSvc, gitSvc)
    assetSvc := service.NewAssetService(fileSvc)
    
    // 8. 创建 Handlers
    handlers := &server.Handlers{
        Article: server.NewArticleHandler(articleSvc),
        Asset:   server.NewAssetHandler(assetSvc),
        Git:     server.NewGitHandler(gitSvc, authSvc),
        Search:  server.NewSearchHandler(articleSvc),
        Tree:    server.NewTreeHandler(articleSvc),
        Static:  server.NewStaticHandler(),
    }
    
    // 9. 创建 HTTP Server
    addr := flags.host + ":" + flags.port
    srv := server.NewServer(addr, handlers, logger)
    
    // 10. 启动服务（优雅关闭）
    ctx, cancel := context.WithCancel(context.Background())
    
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        slog.Info("Shutting down server...")
        cancel()
        srv.Stop(ctx)
    }()
    
    slog.Info("Starting server", "addr", addr)
    if err := srv.Start(); err != nil {
        slog.Error("Server error", "error", err)
        os.Exit(1)
    }
}

// 命令行参数
type Flags struct {
    host      string
    port      string
    configPath string
    logLevel  string
}

func parseFlags() *Flags {
    f := &Flags{}
    flag.StringVar(&f.host, "host", "0.0.0.0", "Server host")
    flag.StringVar(&f.port, "port", "8080", "Server port")
    flag.StringVar(&f.configPath, "config", "config.toml", "Config file path")
    flag.StringVar(&f.logLevel, "log", "info", "Log level (debug, info, warn, error)")
    flag.Parse()
    return f
}
```

### 6.2 文章列表获取流程

```
HTTP Request: GET /api/articles?dir=tech&sort=edited&order=desc
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│                Article Handler                       │
│  1. 解析 query params                               │
│  2. 验证参数                                         │
│  3. 调用 ArticleService.ListArticles()              │
└─────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│                Article Service                       │
│  1. 调用 FileService.ScanMarkdownFiles(dir)         │
│  2. 对每个文件：                                     │
│     a. GitService.IsFileCommitted(path)             │
│     b. GitService.GetFileHistory(path)              │
│  3. 过滤未提交文件                                   │
│  4. 构建 Article 结构                               │
│  5. 根据参数排序                                     │
└─────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│                  Git Service                         │
│  使用 go-git 打开仓库，查询文件历史                    │
│  - 遍历 commits                                     │
│  - 提取作者、时间、hash                              │
└─────────────────────────────────────────────────────┘
                    │
                    ▼
JSON Response: { articles: [...], currentDir: "tech", total: 5 }
```

### 6.3 Git Smart HTTP 流程

#### 6.3.1 Git Clone (upload-pack)

```
HTTP Request: GET /info/refs?service=git-upload-pack
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│                 Git Handler                          │
│  1. 验证 service 参数                                │
│  2. 无需认证（公开）                                  │
│  3. 设置 Content-Type                                │
│  4. 调用 GitService.GetUploadPackRefs()              │
└─────────────────────────────────────────────────────┘
                    │
                    ▼
Response Header:
  Content-Type: application/x-git-upload-pack-advertisement


HTTP Request: POST /git-upload-pack
  Content-Type: application/x-git-upload-pack-request
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│                 Git Handler                          │
│  1. 调用 GitService.HandleUploadPack()               │
│  2. 返回 packfile                                    │
└─────────────────────────────────────────────────────┘
                    │
                    ▼
Response:
  Content-Type: application/x-git-upload-pack-result
  Body: Git packfile
```

#### 6.3.2 Git Push (receive-pack)

```
HTTP Request: GET /info/refs?service=git-receive-pack
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│                 Git Handler                          │
│  1. 提取 Authorization header                        │
│  2. 解码 Basic Auth (username:password)             │
│  3. 调用 AuthService.Validate()                      │
│  4. 认证失败 → 401 Unauthorized                      │
│  5. 认证成功 → GitService.GetReceivePackRefs()       │
└─────────────────────────────────────────────────────┘
                    │
                    ▼
认证失败:
  HTTP 401
  WWW-Authenticate: Basic realm="Git"


认证成功:
  Content-Type: application/x-git-receive-pack-advertisement


HTTP Request: POST /git-receive-pack
  Authorization: Basic {credentials}
  Content-Type: application/x-git-receive-pack-request
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│                 Git Handler                          │
│  1. 再次验证认证                                     │
│  2. 调用 GitService.HandleReceivePack()              │
│  3. 更新仓库 refs                                    │
└─────────────────────────────────────────────────────┘
                    │
                    ▼
Response:
  Content-Type: application/x-git-receive-pack-result
  Body: Git receive result
```

---

## 七、Handler 实现

### 7.1 Article Handler

```go
// internal/handler/article.go

package handler

import (
    "encoding/json"
    "net/http"
    
    "terminalog/internal/model"
    "terminalog/internal/service"
)

type ArticleHandler struct {
    svc service.ArticleService
}

func NewArticleHandler(svc service.ArticleService) *ArticleHandler {
    return &ArticleHandler{svc: svc}
}

// GET /api/articles
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 解析参数
    dir := r.URL.Query().Get("dir")
    sort := r.URL.Query().Get("sort")
    order := r.URL.Query().Get("order")
    
    // 默认值
    if sort == "" { sort = "edited" }
    if order == "" { order = "desc" }
    
    // 调用 Service
    opts := service.ListOptions{
        Dir:   dir,
        Sort:  model.ParseSortField(sort),
        Order: model.ParseSortOrder(order),
    }
    
    articles, err := h.svc.ListArticles(ctx, opts)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, model.ArticleListResponse{
        Articles:   articles,
        CurrentDir: dir,
        Total:      len(articles),
    })
}

// GET /api/articles/{path}
func (h *ArticleHandler) Get(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    path := chi.URLParam(r, "path")
    
    article, err := h.svc.GetArticle(ctx, path)
    if err != nil {
        if errors.Is(err, model.ErrNotFound) {
            respondError(w, http.StatusNotFound, "Article not found")
        } else if errors.Is(err, model.ErrNotCommitted) {
            respondError(w, http.StatusBadRequest, "File not committed")
        } else {
            respondError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }
    
    respondJSON(w, http.StatusOK, model.ArticleResponse{
        Path:         article.Path,
        Title:        article.Title,
        Content:      article.Content,
        CreatedAt:    article.CreatedAt,
        CreatedBy:    article.CreatedBy,
        EditedAt:     article.EditedAt,
        EditedBy:     article.EditedBy,
        Contributors: article.Contributors,
    })
}

// GET /api/articles/{path}/timeline
func (h *ArticleHandler) Timeline(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    path := chi.URLParam(r, "path")
    
    commits, err := h.svc.GetTimeline(ctx, path)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, model.TimelineResponse{
        Commits: commits,
    })
}
```

### 7.2 Git Handler

```go
// internal/handler/git.go

package handler

import (
    "bufio"
    "encoding/base64"
    "io"
    "net/http"
    "strings"
    
    "terminalog/internal/service"
)

type GitHandler struct {
    gitSvc service.GitService
    authSvc service.AuthService
}

func NewGitHandler(gitSvc service.GitService, authSvc service.AuthService) *GitHandler {
    return &GitHandler{gitSvc: gitSvc, authSvc: authSvc}
}

// GET /info/refs?service=git-upload-pack
func (h *GitHandler) UploadPackRefs(w http.ResponseWriter, r *http.Request) {
    service := r.URL.Query().Get("service")
    if service != "git-upload-pack" {
        respondError(w, http.StatusBadRequest, "Invalid service")
        return
    }
    
    ctx := r.Context()
    
    refs, err := h.gitSvc.GetUploadPackRefs(ctx)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    // 设置 Git Smart HTTP 响应头
    w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
    w.Header().Set("Cache-Control", "no-cache")
    
    // 写入 pkt-line 格式
    pktLine(w, "# service=git-upload-pack\n")
    pktFlush(w)
    w.Write(refs)
}

// POST /git-upload-pack
func (h *GitHandler) UploadPack(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
    
    result, err := h.gitSvc.HandleUploadPack(ctx, r.Body)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    w.Write(result)
}

// GET /info/refs?service=git-receive-pack
func (h *GitHandler) ReceivePackRefs(w http.ResponseWriter, r *http.Request) {
    service := r.URL.Query().Get("service")
    if service != "git-receive-pack" {
        respondError(w, http.StatusBadRequest, "Invalid service")
        return
    }
    
    // 认证
    auth := h.extractAuth(r)
    if auth == nil {
        w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    
    valid, err := h.authSvc.Validate(auth.Username, auth.Password)
    if err != nil || !valid {
        w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    
    ctx := r.Context()
    
    refs, err := h.gitSvc.GetReceivePackRefs(ctx)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    w.Header().Set("Content-Type", "application/x-git-receive-pack-advertisement")
    w.Header().Set("Cache-Control", "no-cache")
    
    pktLine(w, "# service=git-receive-pack\n")
    pktFlush(w)
    w.Write(refs)
}

// POST /git-receive-pack
func (h *GitHandler) ReceivePack(w http.ResponseWriter, r *http.Request) {
    // 再次认证
    auth := h.extractAuth(r)
    if auth == nil {
        w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    
    valid, err := h.authSvc.Validate(auth.Username, auth.Password)
    if err != nil || !valid {
        w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    
    ctx := r.Context()
    
    w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
    
    result, err := h.gitSvc.HandleReceivePack(ctx, r.Body, auth)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    w.Write(result)
}

// 提取 Basic Auth
func (h *GitHandler) extractAuth(r *http.Request) *service.AuthInfo {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return nil
    }
    
    if !strings.HasPrefix(authHeader, "Basic ") {
        return nil
    }
    
    decoded, err := base64.StdEncoding.DecodeString(authHeader[6:])
    if err != nil {
        return nil
    }
    
    parts := strings.SplitN(string(decoded), ":", 2)
    if len(parts) != 2 {
        return nil
    }
    
    return &service.AuthInfo{
        Username: parts[0],
        Password: parts[1],
    }
}

// pkt-line 格式辅助函数
func pktLine(w io.Writer, data string) {
    size := len(data) + 4
    w.Write([]byte(fmt.Sprintf("%04x%s", size, data)))
}

func pktFlush(w io.Writer) {
    w.Write([]byte("0000"))
}
```

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

```go
// internal/service/article.go

package service

import (
    "context"
    "sort"
    "strings"
    "time"
    
    "terminalog/internal/model"
)

type ArticleService struct {
    fileSvc FileService
    gitSvc  GitService
}

func NewArticleService(fileSvc FileService, gitSvc GitService) *ArticleService {
    return &ArticleService{
        fileSvc: fileSvc,
        gitSvc:  gitSvc,
    }
}

func (s *ArticleService) ListArticles(ctx context.Context, opts ListOptions) ([]model.Article, error) {
    // 1. 扫描 Markdown 文件
    files, err := s.fileSvc.ScanMarkdownFiles(ctx, opts.Dir)
    if err != nil {
        return nil, err
    }
    
    // 2. 过滤并获取历史
    articles := make([]model.Article, 0, len(files))
    
    for _, file := range files {
        // 检查是否已提交
        committed, err := s.gitSvc.IsFileCommitted(ctx, file)
        if err != nil || !committed {
            continue // 忽略未提交的文件
        }
        
        // 获取历史
        history, err := s.gitSvc.GetFileHistory(ctx, file)
        if err != nil {
            continue
        }
        
        // 构建 Article
        article := model.Article{
            Path:         file,
            Title:        extractTitle(file),
            CreatedAt:    history.FirstCommit.Timestamp,
            CreatedBy:    history.FirstCommit.Author,
            EditedAt:     history.LastCommit.Timestamp,
            EditedBy:     history.LastCommit.Author,
            Contributors: history.Contributors,
        }
        
        articles = append(articles, article)
    }
    
    // 3. 排序
    sortArticles(articles, opts.Sort, opts.Order)
    
    return articles, nil
}

func (s *ArticleService) GetArticle(ctx context.Context, path string) (*model.ArticleDetail, error) {
    // 验证路径
    fullPath, err := s.fileSvc.ValidatePath(path)
    if err != nil {
        return nil, model.ErrInvalidPath
    }
    
    // 检查是否已提交
    committed, err := s.gitSvc.IsFileCommitted(ctx, path)
    if err != nil {
        return nil, err
    }
    if !committed {
        return nil, model.ErrNotCommitted
    }
    
    // 读取内容
    content, err := s.fileSvc.ReadFile(ctx, path)
    if err != nil {
        return nil, model.ErrNotFound
    }
    
    // 获取历史
    history, err := s.gitSvc.GetFileHistory(ctx, path)
    if err != nil {
        return nil, err
    }
    
    return &model.ArticleDetail{
        Article: model.Article{
            Path:         path,
            Title:        extractTitle(path),
            CreatedAt:    history.FirstCommit.Timestamp,
            CreatedBy:    history.FirstCommit.Author,
            EditedAt:     history.LastCommit.Timestamp,
            EditedBy:     history.LastCommit.Author,
            Contributors: history.Contributors,
        },
        Content: string(content),
    }, nil
}

func (s *ArticleService) GetTimeline(ctx context.Context, path string) ([]model.CommitInfo, error) {
    history, err := s.gitSvc.GetFileHistory(ctx, path)
    if err != nil {
        return nil, err
    }
    
    return history.AllCommits, nil
}

func (s *ArticleService) GetTree(ctx context.Context, dir string) (*model.TreeNode, error) {
    return s.fileSvc.GetDirectoryTree(ctx, dir)
}

func (s *ArticleService) Search(ctx context.Context, query string, dir string) ([]model.SearchResult, error) {
    // 获取文章列表
    articles, err := s.ListArticles(ctx, ListOptions{Dir: dir})
    if err != nil {
        return nil, err
    }
    
    // 搜索标题
    results := make([]model.SearchResult, 0)
    lowerQuery := strings.ToLower(query)
    
    for _, article := range articles {
        if strings.Contains(strings.ToLower(article.Title), lowerQuery) {
            results = append(results, model.SearchResult{
                Path:         article.Path,
                Title:        article.Title,
                MatchedTitle: article.Title,
            })
        }
    }
    
    return results, nil
}

// 辅助函数
func extractTitle(path string) string {
    // 从路径提取文件名，去除 .md 扩展名
    parts := strings.Split(path, "/")
    name := parts[len(parts)-1]
    return strings.TrimSuffix(name, ".md")
}

func sortArticles(articles []model.Article, sort model.SortField, order model.SortOrder) {
    sort.Slice(articles, func(i, j int) bool {
        var less bool
        
        switch sort {
        case model.SortCreated:
            less = articles[i].CreatedAt.Before(articles[j].CreatedAt)
        case model.SortEdited:
            less = articles[i].EditedAt.Before(articles[j].EditedAt)
        default:
            less = articles[i].EditedAt.Before(articles[j].EditedAt)
        }
        
        if order == model.OrderDesc {
            return !less
        }
        return less
    })
}
```

### 8.2 Git Service 实现

```go
// internal/service/git.go

package service

import (
    "context"
    "fmt"
    "io"
    "sort"
    "strings"
    "time"
    
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
    
    "terminalog/internal/model"
)

type GitService struct {
    repoPath string
    repo     *git.Repository
}

func NewGitService(repoPath string) *GitService {
    repo, err := git.PlainOpen(repoPath)
    if err != nil {
        return nil // 后续处理
    }
    
    return &GitService{
        repoPath: repoPath,
        repo:     repo,
    }
}

func (s *GitService) GetFileHistory(ctx context.Context, filePath string) (*FileHistory, error) {
    if s.repo == nil {
        return nil, fmt.Errorf("repository not initialized")
    }
    
    // 获取所有 commits
    commits, err := s.repo.Log(&git.LogOptions{})
    if err != nil {
        return nil, err
    }
    
    // 过滤涉及该文件的 commits
    fileCommits := make([]model.CommitInfo, 0)
    contributors := make(map[string]bool)
    
    err = commits.ForEach(func(c *object.Commit) error {
        // 检查文件是否在该 commit 中存在或被修改
        file, err := c.File(filePath)
        if err != nil {
            // 文件在该 commit 中不存在，跳过
            return nil
        }
        
        // 记录 commit 信息
        commitInfo := model.CommitInfo{
            Hash:      shortHash(c.Hash.String()),
            Author:    c.Author.Name,
            Timestamp: c.Author.When,
        }
        
        fileCommits = append(fileCommits, commitInfo)
        contributors[c.Author.Name] = true
        
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // 检查是否有历史
    if len(fileCommits) == 0 {
        return nil, model.ErrNotCommitted
    }
    
    // 按时间排序（倒序）
    sort.Slice(fileCommits, func(i, j int) bool {
        return fileCommits[i].Timestamp.After(fileCommits[j].Timestamp)
    })
    
    // 构建返回结构
    history := &FileHistory{
        FirstCommit:  fileCommits[len(fileCommits)-1], // 最早的
        LastCommit:   fileCommits[0],                   // 最新的
        AllCommits:   fileCommits,
        Contributors: mapKeys(contributors),
    }
    
    return history, nil
}

func (s *GitService) IsFileCommitted(ctx context.Context, filePath string) (bool, error) {
    history, err := s.GetFileHistory(ctx, filePath)
    if err != nil {
        if errors.Is(err, model.ErrNotCommitted) {
            return false, nil
        }
        return false, err
    }
    return len(history.AllCommits) > 0, nil
}

func (s *GitService) GetUploadPackRefs(ctx context.Context) ([]byte, error) {
    // 获取 refs
    refs, err := s.repo.References()
    if err != nil {
        return nil, err
    }
    
    var buf bytes.Buffer
    
    err = refs.ForEach(func(ref *plumbing.Reference) error {
        // 写入 HEAD
        if ref.Name() == plumbing.HEAD {
            pktLine(&buf, fmt.Sprintf("%s HEAD\n", ref.Hash().String()))
        }
        // 写入其他 refs
        if ref.Name().IsBranch() {
            pktLine(&buf, fmt.Sprintf("%s %s\n", ref.Hash().String(), ref.Name().String()))
        }
        return nil
    })
    
    pktFlush(&buf)
    
    return buf.Bytes(), nil
}

func (s *GitService) HandleUploadPack(ctx context.Context, body io.Reader) ([]byte, error) {
    // 读取请求
    // 解析 wants 和 have
    // 返回 packfile
    // ...（详细实现需要处理 Git 协议细节）
    return nil, nil
}

func (s *GitService) GetReceivePackRefs(ctx context.Context) ([]byte, error) {
    // 类似 GetUploadPackRefs
    return s.GetUploadPackRefs(ctx)
}

func (s *GitService) HandleReceivePack(ctx context.Context, body io.Reader, auth *AuthInfo) ([]byte, error) {
    // 读取请求
    // 解析 refs 更新请求
    // 验证 auth
    // 更新 refs
    // 返回结果
    // ...（详细实现需要处理 Git 协议细节）
    return nil, nil
}

// 辅助函数
func shortHash(hash string) string {
    if len(hash) >= 7 {
        return hash[:7]
    }
    return hash
}

func mapKeys(m map[string]bool) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}
```

---

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

```go
// internal/service/cache.go

package service

import (
    "sync"
    "time"
    
    "terminalog/internal/model"
)

type ArticleCache struct {
    articles  map[string]*model.Article
    timelines map[string][]model.CommitInfo
    mutex     sync.RWMutex
    ttl       time.Duration
    lastUpdate time.Time
}

func NewArticleCache(ttl time.Duration) *ArticleCache {
    return &ArticleCache{
        articles:  make(map[string]*model.Article),
        timelines: make(map[string][]model.CommitInfo),
        ttl:       ttl,
    }
}

func (c *ArticleCache) Get(path string) (*model.Article, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    // 检查 TTL
    if time.Since(c.lastUpdate) > c.ttl {
        return nil, false
    }
    
    article, ok := c.articles[path]
    return article, ok
}

func (c *ArticleCache) Set(path string, article *model.Article) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.articles[path] = article
    c.lastUpdate = time.Now()
}

func (c *ArticleCache) Clear() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.articles = make(map[string]*model.Article)
    c.timelines = make(map[string][]model.CommitInfo)
}
```

### 11.2 并行查询

```go
// internal/service/article.go (优化版本)

func (s *ArticleService) ListArticles(ctx context.Context, opts ListOptions) ([]model.Article, error) {
    files, err := s.fileSvc.ScanMarkdownFiles(ctx, opts.Dir)
    if err != nil {
        return nil, err
    }
    
    // 并行获取历史
    type result struct {
        article model.Article
        err     error
    }
    
    results := make(chan result, len(files))
    
    for _, file := range files {
        go func(f string) {
            committed, err := s.gitSvc.IsFileCommitted(ctx, f)
            if err != nil || !committed {
                results <- result{err: err}
                return
            }
            
            history, err := s.gitSvc.GetFileHistory(ctx, f)
            if err != nil {
                results <- result{err: err}
                return
            }
            
            results <- result{
                article: model.Article{
                    Path:         f,
                    Title:        extractTitle(f),
                    CreatedAt:    history.FirstCommit.Timestamp,
                    CreatedBy:    history.FirstCommit.Author,
                    EditedAt:     history.LastCommit.Timestamp,
                    EditedBy:     history.LastCommit.Author,
                    Contributors: history.Contributors,
                },
            }
        }(file)
    }
    
    // 收集结果
    articles := make([]model.Article, 0, len(files))
    for i := 0; i < len(files); i++ {
        r := <-results
        if r.err == nil {
            articles = append(articles, r.article)
        }
    }
    
    sortArticles(articles, opts.Sort, opts.Order)
    return articles, nil
}
```

---

## 十二、构建流程

### 12.1 Makefile

```makefile
# Makefile

.PHONY: all frontend backend build clean run test

# 默认目标
all: build

# 前端构建
frontend:
	cd frontend && pnpm install && pnpm build
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
| Git 历史查询性能 | 大仓库查询慢 | 实现缓存；并行查询 |
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