# Frontend Module - Terminalog 前端模块

> 版本: v1.0.0
> 创建日期: 2026-04-17
> 技术栈: Next.js 16 + TypeScript + Tailwind CSS 4

---

## 一、模块概述

Terminalog 前端采用 **Next.js 静态导出** 模式，生成纯静态 HTML/CSS/JS 文件，最终通过 Go embed 嵌入后端二进制文件，实现单文件部署。

### 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Next.js | 16.2.4 | App Router，静态导出支持 |
| React | 19.2.4 | UI 框架 |
| TypeScript | 5.x | 类型安全 |
| Tailwind CSS | 4.x | 样式框架 |
| ESLint | 9.x | 代码检查 |

---

## 二、目录结构

```
frontend/
├── app/                    # Next.js App Router
│   ├── layout.tsx          # 根布局
│   ├── page.tsx            # 主页
│   ├── globals.css         # 全局样式
│   ├── article/[path]/     # 文章详情页（待实现）
│   └── aboutme/            # About Me 页面（待实现）
│
├── components/             # React 组件（待实现）
│   ├── ui/                 # shadcn/ui 组件
│   ├── brutalist/          # Brutalist 风格组件
│   ├── article/            # 文章相关组件
│   ├── command/            # 命令处理
│   └── common/             # 通用组件
│
├── lib/                    # 库/工具（待实现）
│   ├── api/                # API 客户端
│   ├── hooks/              # React Hooks
│   └── utils/              # 工具函数
│
├── types/                  # TypeScript 类型定义
│   └── index.ts            # 全局类型
│
├── public/                 # 静态资源
│
├── next.config.ts          # Next.js 配置（静态导出）
├── tailwind.config.ts      # Tailwind 配置
├── tsconfig.json           # TypeScript 配置
├── package.json            # 项目依赖
└ └ node_modules/           # 依赖包（构建时安装）
└ └ out/                    # 静态导出产物（构建生成）
└ └ next/                   # Next.js 缓存（构建生成）
```

---

## 三、构建流程

### 3.1 构建命令

```bash
# 开发模式
cd frontend && npm run dev

# 静态导出构建
cd frontend && npm run build

# 输出目录: frontend/out/
```

### 3.2 集成构建（Makefile）

```bash
# 完整构建：前端 + 复制到 embed 目录 + 后端编译
make build

# 仅前端构建并复制
make web-embed

# 仅后端编译（需先执行 web-embed）
make backend

# 清理构建产物
make clean
```

### 3.3 构建流程图

```
┌──────────────────┐
│  frontend/       │
│  npm run build   │
└──────────────────┘
        │
        ▼ 生成 out/
┌──────────────────┐
│  frontend/out/   │
│  (静态 HTML/JS)  │
└──────────────────┘
        │
        ▼ cp -r
┌──────────────────┐
│ pkg/embed/static │
│  (Go embed 目录) │
└──────────────────┘
        │
        ▼ go build
┌──────────────────┐
│  bin/terminalog  │
│  (嵌入前端二进制)│
└──────────────────┘
```

---

## 四、GoReleaser 配置

`.goreleaser.yml` 定义了自动构建流程：

```yaml
before:
  hooks:
    # 1. 构建前端
    - sh -c "cd frontend && npm install && npm run build"
    # 2. 复制静态文件到 embed 目录
    - sh -c "rm -rf pkg/embed/static && mkdir -p pkg/embed/static && cp -r frontend/out/* pkg/embed/static/"
    # 3. Go mod tidy
    - go mod tidy
```

---

## 五、静态导出配置

`next.config.ts` 关键配置：

```typescript
const nextConfig: NextConfig = {
  output: 'export',              // 启用静态导出
  trailingSlash: true,           // URL 带 trailing slash
  images: {
    unoptimized: true,           // 静态导出不支持图片优化
  },
};
```

---

## 六、类型定义

`types/index.ts` 包含以下类型：

- `Article` - 文章类型
- `ArticleListResponse` - 文章列表响应
- `ArticleResponse` - 文章详情响应
- `CommitInfo` - Commit 信息
- `TreeNode` - 目录树节点
- `Command` - 命令类型
- `SortState` - 排序状态
- `VersionInfo` - 版本信息
- `AboutMeResponse` - About Me 响应

---

## 七、后续开发

根据 `docs/frontend-architecture.md` v1.2，需要实现：

1. **Brutalist UI 组件** - Layout, Navbar, ArticleTable, CommandPrompt
2. **Markdown Renderer** - react-markdown + rehype-highlight + KaTeX
3. **API Client** - 与后端 `/api/*` 通信
4. **Command Parser** - cd, view, search, help 命令
5. **版本号展示** - 基于 `/api/articles/{path}/version`

---

## 八、注意事项

1. **pkg/embed/static/ 目录被 gitignore** - 构建时生成，不提交到仓库
2. **Node 版本要求** - 需要 Node 20.19+ 或 22.13+
3. **静态导出限制** - 不支持动态路由的 generateStaticParams（需前端路由处理）

---

**文档结束**