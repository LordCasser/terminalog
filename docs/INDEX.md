# Terminalog Documentation Index

本目录包含 Terminalog 项目的全部技术文档，按模块分门别类组织。

---

## 目录结构

```
docs/
├── frontend/           # 前端相关文档
│   ├── architecture.md
│   ├── implementation.md
│   ├── architecture-refactoring.md
│   ├── readme.md
│   └── markdown-theme-system.md
│
├── backend/            # 后端相关文档
│   ├── architecture.md
│   ├── api-spec.md
│   ├── server.md
│   ├── restful-migration.md
│   ├── version-service.md
│   ├── special-file-service.md
│   └── aboutme-handler.md
│
├── components/         # UI组件实现文档
│   ├── modal.md
│   └── command.md
│
├── architecture.md     # 全局架构设计
├── requirements.md     # 项目需求规格
├── testing.md          # 测试策略与验证
├── DEBUG.md            # 调试指南
│
└── 主页面原型/          # UI原型设计稿
    └── DESIGN.md
```

---

## 文档分类

### 前端文档 (`frontend/`)

| 文档 | 内容 |
|---|---|
| [architecture.md](frontend/architecture.md) | 前端架构设计、技术选型、约束原则 |
| [implementation.md](frontend/implementation.md) | 前端实现笔记、API调用、渲染路径 |
| [architecture-refactoring.md](frontend/architecture-refactoring.md) | 前端架构重构记录（公共组件整合） |
| [readme.md](frontend/readme.md) | 前端项目说明 |
| [markdown-theme-system.md](frontend/markdown-theme-system.md) | **Markdown主题系统指南**（编写、应用、切换） |

### 后端文档 (`backend/`)

| 文档 | 内容 |
|---|---|
| [architecture.md](backend/architecture.md) | 后端架构设计、模块划分 |
| [api-spec.md](backend/api-spec.md) | RESTful API 规格定义 |
| [server.md](backend/server.md) | HTTP Server 实现细节 |
| [restful-migration.md](backend/restful-migration.md) | RESTful API 迁移记录 |
| [version-service.md](backend/version-service.md) | 版本服务实现 |
| [special-file-service.md](backend/special-file-service.md) | 特殊文件处理（_ABOUTME.md） |
| [aboutme-handler.md](backend/aboutme-handler.md) | ABOUTME 接口处理 |

### 组件文档 (`components/`)

| 文档 | 内容 |
|---|---|
| [modal.md](components/modal.md) | Modal 组件实现（搜索、帮助、路径补全） |
| [command.md](components/command.md) | CommandPrompt 组件实现 |

### 全局文档

| 文档 | 内容 |
|---|---|
| [architecture.md](architecture.md) | 全局架构设计（前后端整合） |
| [requirements.md](requirements.md) | 项目需求规格与功能约束 |
| [testing.md](testing.md) | 测试策略、验证项、测试方法 |
| [DEBUG.md](DEBUG.md) | 开发调试指南、问题排查 |

---

## 快速导航

### 我想了解...

| 需求 | 推荐文档 |
|---|---|
| 项目整体架构 | [architecture.md](architecture.md) |
| 前端技术选型 | [frontend/architecture.md](frontend/architecture.md) |
| 后端API接口 | [backend/api-spec.md](backend/api-spec.md) |
| Markdown渲染样式 | [frontend/markdown-theme-system.md](frontend/markdown-theme-system.md) |
| 如何自定义主题 | [frontend/markdown-theme-system.md](frontend/markdown-theme-system.md#二如何编写主题) |
| 搜索组件实现 | [components/modal.md](components/modal.md) |
| 命令行组件 | [components/command.md](components/command.md) |
| 测试验证 | [testing.md](testing.md) |
| 调试问题 | [DEBUG.md](DEBUG.md) |

---

## 文档维护规范

### 1. 文档位置规则

- **前端文档** → `docs/frontend/`
- **后端文档** → `docs/backend/`
- **组件文档** → `docs/components/`
- **全局文档** → `docs/`（根目录）
- **原型设计** → `docs/主页面原型/` 或 `docs/文章查看页面原型/`

### 2. 文档命名规范

- 使用英文小写 + 连字符：`api-spec.md`、`architecture-refactoring.md`
- 实现文档使用功能名：`server.md`、`version-service.md`
- 系统指南使用描述性名称：`markdown-theme-system.md`

### 3. 文档更新

- 每次功能改动应同步更新相关文档
- 架构变更需在 `architecture.md` 或对应的子架构文档中记录
- 新增组件应在 `components/` 下创建实现文档