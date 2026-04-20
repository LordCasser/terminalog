# Terminalog Documentation Index

本目录包含 Terminalog 项目的全部技术文档，按**开发阶段**分类组织。

---

## 目录结构

```
docs/
├── specs/                  # 需求规格阶段
│   ├── requirements.md     # 项目需求规格（唯一入口）
│   └── api-spec.md         # API 接口规格
│
├── design/                 # 设计阶段
│   ├── architecture.md     # 全局架构概览
│   ├── frontend-design.md  # 前端架构设计
│   ├── backend-design.md   # 后端架构设计
│   ├── components.md       # 组件设计规格
│   └── prototype/          # UI 原型设计
│       └── homepage.md     # 主页面原型
│
├── guides/                 # 使用指南
│   ├── markdown-theme.md   # Markdown 主题系统指南
│   ├── testing.md          # 测试指南
│   └── debug.md            # 调试指南
│
├── history/                # 开发历史记录（归档）
│   ├── frontend-refactoring.md
│   ├── restful-migration.md
│   └── backend-services.md # 合并：version-service + special-file + aboutme + server
│
└── INDEX.md                # 文档索引
```

---

## 文档分类

### 需求规格 (`specs/`)

| 文档 | 内容 |
|---|---|
| [requirements.md](specs/requirements.md) | 项目需求规格与功能约束 |
| [api-spec.md](specs/api-spec.md) | RESTful API 规格定义 |

### 设计文档 (`design/`)

| 文档 | 内容 |
|---|---|
| [architecture.md](design/architecture.md) | 全局架构设计（前后端整合概览） |
| [frontend-design.md](design/frontend-design.md) | 前端架构设计、技术选型、渲染路径 |
| [backend-design.md](design/backend-design.md) | 后端架构设计、模块划分 |
| [components.md](design/components.md) | UI 组件设计规格（Modal、CommandPrompt） |
| [prototype/homepage.md](design/prototype/homepage.md) | 主页面 UI 原型设计稿 |

### 使用指南 (`guides/`)

| 文档 | 内容 |
|---|---|
| [markdown-theme.md](guides/markdown-theme.md) | Markdown 主题系统（编写、应用、切换、警告框） |
| [testing.md](guides/testing.md) | 测试策略、验证项、测试方法 |
| [debug.md](guides/debug.md) | 开发调试指南、问题排查 |

### 开发历史 (`history/`)

| 文档 | 内容 |
|---|---|
| [frontend-refactoring.md](history/frontend-refactoring.md) | 前端架构重构记录 |
| [restful-migration.md](history/restful-migration.md) | RESTful API 迁移历史 |
| [backend-services.md](history/backend-services.md) | 后端服务实现细节（归档） |

---

## 快速导航

### 我想了解...

| 需求 | 推荐文档 |
|---|---|
| 项目整体架构 | [design/architecture.md](design/architecture.md) |
| 项目需求规格 | [specs/requirements.md](specs/requirements.md) |
| 后端 API 接口 | [specs/api-spec.md](specs/api-spec.md) |
| 前端技术选型 | [design/frontend-design.md](design/frontend-design.md) |
| 后端模块设计 | [design/backend-design.md](design/backend-design.md) |
| Markdown 渲染样式 | [guides/markdown-theme.md](guides/markdown-theme.md) |
| 如何自定义主题 | [guides/markdown-theme.md](guides/markdown-theme.md#四alerts-callouts-使用方法) |
| GitHub 警告框语法 | [guides/markdown-theme.md](guides/markdown-theme.md#四alerts-callouts-使用方法) |
| 组件设计规格 | [design/components.md](design/components.md) |
| 测试验证 | [guides/testing.md](guides/testing.md) |
| 调试问题 | [guides/debug.md](guides/debug.md) |

---

## 文档维护规范

### 1. 文档位置规则

- **需求规格** → `docs/specs/`
- **设计文档** → `docs/design/`
- **使用指南** → `docs/guides/`
- **历史记录** → `docs/history/`（归档，不再活跃更新）

### 2. 文档命名规范

- 使用英文小写 + 连字符：`api-spec.md`、`frontend-design.md`
- 系统指南使用描述性名称：`markdown-theme.md`、`debug.md`
- 原型设计使用功能名：`homepage.md`

### 3. 文档更新

- **设计文档**：架构变更需同步更新 `design/` 下对应文档
- **需求规格**：功能扩展需同步更新 `specs/requirements.md`
- **API 规格**：新增接口需同步更新 `specs/api-spec.md`
- **历史记录**：重大重构完成后归档到 `history/`

### 4. 文档分类原则

| 分类 | 性质 | 更新频率 |
|---|---|---|
| `specs/` | 规格定义 | 低（项目稳定后少变动） |
| `design/` | 设计规格 | 中（架构演进时更新） |
| `guides/` | 使用指南 | 高（功能新增时更新） |
| `history/` | 变更记录 | 归档（不活跃更新） |

---

## 整理记录

**2026-04-20**：按开发阶段重新整理文档结构
- 从 20 个文档 → 14 个文档（含 INDEX.md）
- 合并后端服务文档、组件文档
- 移除冗余文档（frontend/readme.md、frontend/implementation.md）
- 统一英文路径命名