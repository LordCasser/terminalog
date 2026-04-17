# Terminalog - 系统架构总览

> 文档版本：v2.1
> 创建日期：2026-04-15
> 最后更新：2026-04-18
> 基于需求文档：requirements.md v1.6

---

## 一、架构概览

Terminalog 采用**前后端分离 + 单文件部署**的架构模式：

- **前端**：Next.js 静态导出，生成纯静态 HTML/CSS/JS
- **后端**：Go HTTP 服务，提供 API 和 Git Smart HTTP 服务
- **部署**：前端静态资源通过 Go embed 嵌入二进制文件，实现单文件启动

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Terminalog 可执行文件                          │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    前端静态资源（embed）                          │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │ │
│  │  │   HTML       │  │   CSS        │  │   JavaScript         │ │ │
│  │  │   (页面)      │  │   (样式)      │  │   (交互逻辑)          │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                         后端服务                                 │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │ │
│  │  │  HTTP Server │  │  REST API    │  │  Git Smart HTTP      │ │ │
│  │  │  (路由分发)    │  │  (文章/资源) │  │  (Git 协议)          │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Git 仓库（用户指定目录）                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐ │
│  │  .git/       │  │  *.md        │  │  assets/                 │ │
│  │  (Git 历史)   │  │  (Markdown)  │  │  (图片资源)               │ │
│  └──────────────┘  └──────────────┘  └──────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.2 运行时交互模型

```
┌──────────────┐                    ┌──────────────┐
│    用户       │                    │  反向代理     │
│  (浏览器)     │───── HTTPS ──────▶│  (Nginx)     │
└──────────────┘                    └──────────────┘
       │                                    │
       │ HTTP                               │ HTTP
       │                                    │
       ▼                                    ▼
┌──────────────────────────────────────────────────────┐
│                   Terminalog 服务                      │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │  静态资源     │  │  REST API    │  │  Git API   │ │
│  │  (前端页面)   │  │  (文章数据)   │  │  (Git 操作) │ │
│  └──────────────┘  └──────────────┘  └────────────┘ │
└──────────────────────────────────────────────────────┘
       │                    │                │
       │                    │                │
       ▼                    ▼                ▼
┌──────────────────────────────────────────────────────┐
│                    Git 仓库                            │
│            (Blog 内容目录 + Git 历史)                   │
└──────────────────────────────────────────────────────┘
```

---

## 二、子系统划分

系统分为两大子系统：

### 2.1 前端子系统

详细设计见：[frontend-architecture.md](./frontend-architecture.md)

| 模块 | 职责 |
|------|------|
| Brutalist UI | Brutalist 风格 UI 组件，Dracula Spectrum 配色，Glass 效果，0px 圆角 |
| Command Parser | 命令行输入解析与执行（**无 clear 命令**） |
| Sort Manager | 排序状态管理（表头点击 + 命令行排序共用） |
| Markdown Renderer | Markdown 内容渲染（代码高亮、公式、Mermaid，Inter 字体） |
| Article Viewer | 文章详情页展示（版本号、折叠式历史、EOF、标签） |
| About Me | About Me 页面展示（读取 `_ABOUTME.md`） |
| API Client | 与后端 API 通信 |
| Path Transformer | 图片路径转换 |

### 2.2 后端子系统

详细设计见：[backend-architecture.md](./backend-architecture.md)

| 模块 | 职责 |
|------|------|
| HTTP Server | HTTP 路由分发，静态资源服务 |
| Article Service | 文章列表、内容读取、元数据获取（**过滤 `_` 开头文件**） |
| About Me Service | 读取并返回 `_ABOUTME.md` 内容 |
| Version Service | 基于行数变化计算语义版本号 |
| Git Service | Git 历史查询，Smart HTTP 协议实现 |
| File Service | 文件系统操作，目录扫描，特殊文件过滤 |
| Auth Service | 用户认证校验，密码验证 |
| Asset Service | 图片等静态资源读取与响应 |
| Config Manager | TOML 配置文件解析与管理 |

---

## 三、模块依赖关系

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

## 四、API 概览

详细接口定义见：[api-spec.md](./api-spec.md)

### 4.1 REST API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/articles` | GET | 获取文章列表（**排除 `_` 开头文件**，支持 `?sort=created|edited&order=asc|desc`） |
| `/api/v1/articles/{path}` | GET | 获取文章内容 |
| `/api/v1/articles/{path}/timeline` | GET | 获取文章 Git 时间线 |
| `/api/v1/articles/{path}/version` | GET | 获取文章版本号及历史（v1.2 新增） |
| `/api/v1/articles/search` | GET | 搜索文章标题（**排除 `_` 开头文件**） |
| `/api/v1/tree` | GET | 获取目录树结构（**排除 `_` 开头文件**） |
| `/api/v1/assets/{path}` | GET | 获取图片等静态资源 |
| `/api/v1/special/aboutme` | GET | 获取 About Me 内容（v1.2 新增） |
| `/api/v1/settings` | GET | 获取前端配置 |
| `/api/v1/healthz` | GET | 服务健康状态 |
| `/api/v1/readyz` | GET | 服务就绪状态 |
| `/api/v1/livez` | GET | 服务存活状态 |
| `/api/v1/status` | GET | 详细状态信息 |

### 4.2 Git Smart HTTP API

**Git Clone URL**: `http://{host}:{port}/api/v1/git/`

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/git/info/refs?service=git-upload-pack` | GET | Git Clone refs |
| `/api/v1/git/git-upload-pack` | POST | Git Clone packfile |
| `/api/v1/git/info/refs?service=git-receive-pack` | GET | Git Push refs（需认证） |
| `/api/v1/git/git-receive-pack` | POST | Git Push 数据（需认证） |

---

## 五、项目目录结构

```
terminalog/
├── cmd/
│   └── terminalog/
│       └── main.go              # 入口文件
│
├── internal/                    # 内部模块（不对外暴露）
│   ├── config/                  # 配置管理
│   ├── service/                 # 业务服务
│   ├── handler/                 # HTTP Handler
│   ├── model/                   # 数据模型
│   └── server/                  # HTTP Server
│
├── pkg/                         # 公共包（可对外暴露）
│   └── embed/                   # 嵌入的静态资源
│       └── static/              # 前端构建产物
│       └── embed.go
│   └── utils/                   # 工具函数
│
├── frontend/                    # 前端源码（独立子项目）
│   ├── app/                     # Next.js App Router
│   ├── components/              # React 组件
│   ├── lib/                     # API、Hooks、工具
│   ├── styles/                  # 样式文件
│   ├── types/                   # TypeScript 类型
│   └── next.config.js           # Next.js 配置
│
├── docs/                        # 文档
│   ├── requirements.md          # 需求文档
│   ├── architecture.md          # 系统架构总览（本文件）
│   ├── frontend-architecture.md # 前端架构详细设计
│   ├── backend-architecture.md  # 后端架构详细设计
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

## 六、技术选型总览

### 6.1 前端技术栈

| 组件 | 技术选型 |
|------|---------|
| 框架 | Next.js 14+（静态导出） |
| UI 库 | shadcn/ui |
| 样式 | Tailwind CSS |
| Markdown | react-markdown + rehype-highlight + KaTeX + Mermaid |
| 语言 | TypeScript 5+ |
| 标题字体 | Space Grotesk（Google Fonts） |
| 正文字体 | Inter（Google Fonts） |
| UI 字体 | JetBrains Mono（Google Fonts） |

### 6.2 后端技术栈

| 组件 | 技术选型 |
|------|---------|
| 语言 | Go 1.21+ |
| HTTP 路由 | chi v5 |
| Git 操作 | go-git/v5 |
| TOML 解析 | pelletier/go-toml/v2 |
| 日志 | slog（标准库） |
| 密码哈希 | bcrypt |

---

## 七、构建流程

### 7.1 开发流程

```bash
# 前端开发
cd frontend
pnpm install
pnpm dev         # Next.js 开发模式 (localhost:3000)

# 后端开发
go run cmd/terminalog/main.go --port 8080 --config config.toml
```

### 7.2 生产构建

```bash
# 使用 Makefile
make build       # 前端 + 后端完整构建

# 或手动构建
cd frontend
pnpm build       # 静态导出到 out/
cp -r out/* ../pkg/embed/static/
cd ..
go build -o bin/terminalog cmd/terminalog/main.go
```

### 7.3 Makefile 命令

```makefile
make frontend    # 构建前端
make backend     # 构建后端
make build       # 完整构建
make run         # 运行服务
make clean       # 清理
make release     # 跨平台构建
```

---

## 八、安全设计

### 8.1 认证与授权

| 操作 | 认证要求 | 授权逻辑 |
|------|---------|---------|
| Blog 页面访问 | 无 | 公开 |
| 文章 API (GET) | 无 | 公开 |
| 资源 API (GET) | 无 | 公开 |
| Git Clone | 无 | 公开 |
| Git Push | Basic Auth | 仅配置文件中用户 |

### 8.2 安全措施

- **路径校验**：防止目录遍历攻击
- **文件类型限制**：仅允许访问 Markdown 和图片文件
- **密码哈希**：使用 bcrypt 存储密码
- **.git 保护**：拒绝 API 访问 `.git` 目录

---

## 九、风险与缓解

### 9.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| go-git Smart HTTP 实现复杂 | Git 协议兼容性问题 | 先实现基础功能；可 fallback 到系统 git 命令 |
| Git 历史查询性能 | 大仓库查询慢 | 内存缓存 + 并行查询 |
| 跨平台路径处理 | Windows 路径分隔符问题 | 统一使用 filepath 包 |

### 9.2 架构风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 无数据库 | 元数据查询性能依赖 Git | 缓存 + 限制仓库规模 |
| 单进程 | 无法水平扩展 | 明确 MVP 边界，文档说明 |
| Git 仓库损坏 | 数据丢失 | 建议用户定期备份 |

---

## 十、后续迭代规划

### 10.1 MVP（当前版本）

- ✅ 基础 Blog 展示（Brutalist 编辑器 UI，Dracula Spectrum 配色）
- ✅ 版本号自动生成（基于行数变化计算语义版本号）
- ✅ About Me 页面（从 `_ABOUTME.md` 读取）
- ✅ 特殊文件处理（`_` 开头文件不参与列表展示）
- ✅ 鼠标交互（顶部导航 + 底部单行 prompt）
- ✅ 命令行交互（cd, view, search, help，支持 `cd ..`/`cd .`/`cd` 空）
- ❌ 移除命令：ls、clear、exit
- ✅ 文章列表 5 列表格（Created/Updated/Editors/Filename/Latest Commit）
- ✅ 表头点击排序（与命令行排序共用逻辑）
- ✅ 三字体系统（Space Grotesk + JetBrains Mono + Inter）
- ✅ 0px 圆角 Brutalist 风格
- ✅ Git HTTP 协议（Clone/Push）
- ✅ Markdown 渲染（代码高亮、公式、Mermaid、图片）
- ✅ 文章元数据（Git 历史）
- ❌ 不支持移动端响应式

### 10.2 后续迭代

| 功能 | 优先级 | 依赖 |
|------|--------|------|
| SSH Git 协议 | 中 | SSH Server 实现 |
| 命令历史上下键 | 中 | 前端状态管理 |
| 命令自动补全 | 低 | Tab 补全逻辑 |
| RSS 订阅 | 低 | 无 |
| 全文搜索 | 低 | 搜索引擎集成 |
| Web 管理后台 | 低 | 需求变更 |

---

## 十一、相关文档

| 文档 | 说明 |
|------|------|
| [requirements.md](./requirements.md) | 原始需求文档 |
| [frontend-architecture.md](./frontend-architecture.md) | 前端架构详细设计 |
| [backend-architecture.md](./backend-architecture.md) | 后端架构详细设计 |
| [api-spec.md](./api-spec.md) | API 接口文档 |

---

**文档结束**

> 本系统架构总览基于 requirements.md v1.2（Brutalist 编辑器 + 版本号规则 + 特殊文件处理）
> 下一步：进入实现阶段（Coder 模式）