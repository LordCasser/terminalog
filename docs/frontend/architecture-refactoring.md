# Frontend Architecture Refactoring - Implementation Guide

> Document Version: v2.0
> Created Date: 2026-04-17
> Purpose: Record frontend architecture changes (public components + Markdown rendering)

---

## 一、架构改动概述

### 1.1 公共组件整合

**改动目标**：
- 顶部导航栏（Navbar）作为所有页面的公共组件，集成在 `layout.tsx`
- 底部终端输入栏（CommandPrompt）作为所有页面的公共组件，已集成在 `layout.tsx`

**改动内容**：
1. **Navbar组件** (`components/brutalist/Navbar.tsx`)
   - 移除旧的 `currentPath` 入参，改为由共享配置和当前路由推导导航路径
   - 统一为公共组件，不再在各页面单独导入
   
2. **Layout集成** (`app/layout.tsx`)
   - 导入 `Navbar` 作为顶部公共组件
   - 导入 `CommandPrompt` 作为底部公共组件（已有）
   - 页面内容通过 `children` 传递
   
3. **页面改动**
   - `page.tsx`（首页）：移除Navbar导入和使用
   - `article/page.tsx`：移除Navbar导入和使用
   - `aboutme/page.tsx`：移除独立navbar实现，使用统一样式

### 1.2 Markdown渲染实现

**改动目标**：
- 使用 `react-markdown` 渲染从后端 API 获取的 Markdown 内容
- Markdown样式遵循Dracula Spectrum设计系统

**改动内容**：
1. **依赖安装** (`package.json`)
   ```
   react-markdown remark-gfm rehype-highlight remark-math rehype-katex highlight.js
   ```

2. **配置文件** (`next.config.mjs`)
   - 启用静态导出（`output: 'export'`）

3. **MarkdownRenderer组件** (`components/article/MarkdownRenderer.tsx`)
   - 用于渲染从API获取的远程Markdown内容
   - 支持GitHub Flavored Markdown（remark-gfm）
   - 支持代码语法高亮（rehype-highlight + Dracula theme）
   - 支持数学公式渲染（remark-math + rehype-katex）
   - 支持图片路径转换（相对路径 → API路径）
   - 应用统一的Dracula Spectrum排版样式

---

## 二、架构设计原则

### 2.1 公共组件原则

**约束**：
- 所有页面共享同一导航栏（统一视觉体验）
- Navbar 和 CommandPrompt 显示同一份当前目录信息
- Blog owner 通过 `/api/v1/settings` 下发，当前目录由路由解析得到

**优点**：
1. 架构统一，所有页面共享同一 UI 框架
2. 减少代码重复，提高维护性
3. 当前目录显示和路由状态一致，不再依赖页面手动同步

### 2.2 Markdown渲染原则

**约束**：
- 样式遵循Dracula Spectrum设计系统
- 字体约束：
  - 标题：Space Grotesk（text-3xl for h2）
  - 正文：Inter（text-lg）
  - 代码：JetBrains Mono（text-sm）
  - 引用：JetBrains Mono（text-lg）
  - 列表：JetBrains Mono（text-base）

**优点**：
1. 渲染链路简单，和后端 API 返回的 Markdown 内容直接对接
2. 统一样式，确保视觉一致性
3. 功能完整（GFM、语法高亮、数学公式、图片路径转换）

---

## 三、技术实现细节

### 3.1 Layout架构

```tsx
// app/layout.tsx
export default function RootLayout({ children }) {
  return (
    <html lang="en" className="dark">
      <body>
        {/* Top Navigation Bar - Public Component */}
        <Navbar />
        
        {/* Main Content */}
        <main className="flex-grow pb-24 pt-20">
          {children}
        </main>
        
        {/* Bottom Command Prompt - Public Component */}
        <CommandPrompt />
      </body>
    </html>
  );
}
```

### 3.2 Next.js配置

```mjs
// next.config.mjs
const nextConfig = {
  output: 'export',
  trailingSlash: true,
  images: { unoptimized: true },
  pageExtensions: ['js', 'jsx', 'ts', 'tsx'],
}

export default nextConfig
```

### 3.3 MarkdownRenderer使用示例

```tsx
// article/page.tsx
import { MarkdownRenderer } from "@/components/article/MarkdownRenderer";

// Extract base path for image transformation
const basePath = decodedPath.replace(/\/[^\/]+\.md$/, '');

// Render Markdown content
<MarkdownRenderer content={content} basePath={basePath} />
```

---

## 四、架构约束遵守

### 4.1 架构约束

✅ **遵守以下约束**：
- 遵循YAGNI原则：只实现必要功能，不编写不需要的接口
- 遵循Dracula Spectrum设计系统
- 遵循UI视觉统一性约束（导航栏统一、Markdown渲染样式统一）
- 遵循模块化封装原则（MarkdownRenderer独立组件）

### 4.2 未来扩展

**已支持功能**：
1. ✅ GitHub Flavored Markdown（GFM）
2. ✅ 代码语法高亮（Dracula theme）
3. ✅ 数学公式渲染（KaTeX）
4. ✅ 图片路径转换（相对 → API）
5. ✅ 统一Markdown样式（Dracula Spectrum）

**可选扩展**（未来需要时实现）：
1. Mermaid流程图渲染（需要安装mermaid-react）
2. 命令历史记录（上下键导航）
3. 自动补全功能（Tab补全路径和文章名）

---

## 五、测试验证

### 5.1 公共组件测试

**测试项**：
- Navbar在所有页面顶部显示
- CommandPrompt在所有页面底部显示
- 导航栏样式统一（JetBrains Mono、uppercase、tracking-tighter）
- POSTS和ABOUTME链接正常工作
- 搜索icon点击触发CommandPrompt聚焦

### 5.2 Markdown渲染测试

**测试项**：
- 标题使用Space Grotesk字体
- 正文使用Inter字体
- 代码块使用JetBrains Mono字体+Dracula语法高亮
- 引用块使用JetBrains Mono字体+border-l-4样式
- 列表使用JetBrains Mono字体+tertiary箭头符号
- 图片路径正确转换（相对路径 → API路径）
- 数学公式正确渲染（KaTeX）
- GFM特性正常工作（表格、任务列表等）

---

## 六、Git提交记录

**提交信息示例**：
```
refactor(frontend): integrate Navbar as public component in layout

- Remove deprecated currentPath prop from Navbar component
- Integrate Navbar in layout.tsx as top public component
- Remove independent Navbar implementations from pages
- Update pages to use layout.tsx architecture
- Follow YAGNI principle and UI visual consistency constraints
```

---

**文档结束**

> 本实现文档记录前端架构改动（公共组件整合 + Markdown渲染实现）
> 关联文档：frontend-architecture.md, requirements.md, testing.md
