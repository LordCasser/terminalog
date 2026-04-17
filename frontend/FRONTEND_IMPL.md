# Frontend Module - Terminalog 前端模块

> 版本: v1.1.0
> 创建日期: 2026-04-17
> 最后更新: 2026-04-17
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
| Tailwind CSS | 4.x | CSS-first 配置 |
| ESLint | 9.x | 代码检查 |

### 设计系统

采用 **Dracula Spectrum** 配色 + **Brutalist** 风格：

- **Surface 层级**: lowest(#0b0e18) → highest(#323440)
- **Primary**: #d7baff (淡紫)
- **Secondary**: #ffafd7 (粉色)
- **Tertiary**: #31e368 (绿色成功)
- **0px 圆角**: 全系统无圆角
- **三字体**: Space Grotesk(标题) + Inter(正文) + JetBrains Mono(UI)

---

## 二、目录结构

```
frontend/
├── app/                           # Next.js App Router
│   ├── layout.tsx                 # 根布局（固定底部命令行）
│   ├── page.tsx                   # 主页（文章列表表格）
│   ├── globals.css                # 全局样式（Dracula Spectrum）
│   ├── article/                   # 文章详情页（查询参数路由）
│   │   └── page.tsx               # 文章查看器
│   └── aboutme/                   # About Me 页面
│       └── page.tsx               # _ABOUTME.md 渲染
│
├── components/                    # React 组件
│   ├── brutalist/                 # Brutalist 风格组件
│   │   ├── Navbar.tsx             # 顶部导航栏
│   │   ├── ArticleTable.tsx       # 文章列表表格（5列）
│   │   └── index.ts               # 导出索引
│   │
│   └── ui/                        # shadcn/ui 组件（预留）
│
├── lib/                           # 库/工具
│   ├── api/                       # API 客户端
│   │   ├── client.ts              # 基础 HTTP 客户端
│   │   ├── articles.ts            # 文章 API
│   │   ├── tree.ts                # 目录树 API
│   │   ├── search.ts              # 搜索 API
│   │   └── aboutme.ts             # About Me API
│   │
│   ├── hooks/                     # React Hooks（预留）
│   └── utils/                     # 工具函数（预留）
│
├── types/                         # TypeScript 类型定义
│   └── index.ts                   # 全局类型
│
├── public/                        # 静态资源
│
├── next.config.ts                 # Next.js 配置（静态导出）
├── package.json                   # 项目依赖
└── out/                           # 静态导出产物（构建生成）
```

---

## 三、已实现功能

### 3.1 主页面

- ✅ 顶部 Navbar（Logo + 导航 + 搜索）
- ✅ 文章列表表格（5列：Created/Updated/Editors/Filename/Latest Commit）
- ✅ 表头点击排序（Created/Updated/Filename）
- ✅ 文件类型图标（folder/description/settings/code/image）
- ✅ 编辑者标签（contributor pills）
- ✅ 相对时间格式（2h ago/10m ago/1d ago）
- ✅ 固定底部命令行提示（guest@blog: ~/lordcasser $）

### 3.2 文章详情页

- ✅ 标签 + 版本号 + 日期 header
- ✅ 大标题（Space Grotesk，8xl）
- ✅ 引用块（blockquote，primary边框）
- ✅ Markdown 内容渲染（简化版）
- ✅ EOF 分隔线
- ✅ Git History 折叠展示
- ✅ 导航栏（~/path/filename.md）

### 3.3 About Me 页面

- ✅ 从 _ABOUTME.md 渲染内容
- ✅ 未找到时提示创建文件

### 3.4 API Client

- ✅ `/api/articles` - 文章列表
- ✅ `/api/articles/{path}` - 文章内容
- ✅ `/api/articles/{path}/timeline` - Git 历史
- ✅ `/api/articles/{path}/version` - 版本信息
- ✅ `/api/tree` - 目录树
- ✅ `/api/search` - 搜索
- ✅ `/api/aboutme` - About Me

---

## 四、构建流程

### 4.1 构建命令

```bash
# 开发模式
cd frontend && npm run dev

# 静态导出构建
cd frontend && npm run build

# 输出目录: frontend/out/
```

### 4.2 集成构建（Makefile）

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

### 4.3 构建流程图

```
frontend/npm run build → frontend/out/ → pkg/embed/static/ → go build → bin/terminalog
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

## 六、CSS 配置（Tailwind CSS 4）

`globals.css` 使用 CSS-first 配置：

```css
@import url('https://fonts.googleapis.com/css2?family=Space+Grotesk...');
@import url('https://fonts.googleapis.com/css2?family=Inter...');
@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono...');

@import "tailwindcss";

@theme inline {
  --color-surface-lowest: #0b0e18;
  --color-surface: #11131e;
  --color-primary: #d7baff;
  --color-secondary: #ffafd7;
  --color-tertiary: #31e368;
  --font-headline: "Space Grotesk";
  --font-body: "Inter";
  --font-mono: "JetBrains Mono";
}
```

---

## 七、类型定义

`types/index.ts` 包含以下类型：

| 类型 | 说明 |
|------|------|
| `Article` | 文章元数据 |
| `ArticleListResponse` | 文章列表 API 响应 |
| `ArticleResponse` | 文章详情 API 响应 |
| `CommitInfo` | Git commit 信息 |
| `TreeNode` | 目录树节点 |
| `VersionInfo` | 版本号信息 |
| `VersionHistoryEntry` | 版本历史条目 |
| `AboutMeResponse` | About Me API 响应 |

---

## 八、路由设计

由于静态导出不支持动态路由 `generateStaticParams()`，使用查询参数方式：

| 页面 | 路径 | 说明 |
|------|------|------|
| 主页 | `/` | 文章列表 |
| 文章详情 | `/article?path=xxx.md` | 查询参数路由 |
| About Me | `/aboutme` | 静态页面 |

---

## 九、注意事项

1. **pkg/embed/static/ 被 gitignore** - 构建时生成，不提交
2. **Node 版本要求** - 需要 Node 20.19+ 或 22.13+
3. **useSearchParams 需 Suspense** - 避免静态导出错误
4. **Mock 数据** - 构建时使用 mock 数据预览

---

## 十、后续迭代

- [ ] 完整 Markdown 渲染（react-markdown + KaTeX + Mermaid）
- [ ] 命令行交互（cd, view, search, help）
- [ ] Tab 补全
- [ ] 搜索功能
- [ ] 响应式优化（MVP 不支持）

---

**文档结束**

> ✅ **前端 v1.1 实现完成** - 主页面、文章页、AboutMe页已实现
> 构建验证: `npm run build` 成功，输出静态页面