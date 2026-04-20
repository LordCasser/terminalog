# Markdown Theme System Guide

本文档详细介绍 Terminalog 的 Markdown 渲染主题系统，包括如何编写、应用和切换主题。

## 一、系统架构

Markdown 渲染系统采用完全模块化的 CSS 变量架构。所有 Markdown 内容的视觉属性均由 `--md-*` CSS 自定义属性控制，这些变量定义在 `globals.css` 的 `:root` 选择器下。

### 1.1 变量层级结构

```
:root {
  --md-* 变量         ← 主题令牌（颜色、间距、字体）
}

.markdown-body        ← 使用主题令牌的容器
.markdown-code-block  ← 代码块组件
.markdown-blockquote  ← 引用块组件
.markdown-alert       ← GitHub 风格警告框组件
.markdown-list*       ← 列表组件
.markdown-table*      ← 表格组件
.markdown-inline-code ← 行内代码组件
.hljs*                ← 语法高亮令牌
```

### 1.2 核心组件关系

| 组件 | CSS 类 | 负责内容 |
|---|---|---|
| MarkdownRenderer | `.markdown-body` | Markdown 容器 |
| Code Block | `.markdown-code-block` | 代码块外层容器 |
| Inline Code | `.markdown-inline-code` | 行内代码 |
| Blockquote | `.markdown-blockquote` | 引用块 |
| Alert / Callout | `.markdown-alert` | GitHub 风格警告框 |
| List | `.markdown-list` | 有序/无序列表 |
| Table | `.markdown-table-shell` | 表格容器 |

---

## 二、如何编写主题

### 2.1 创建主题文件

在 `frontend/app/` 目录下创建一个 CSS 文件，命名格式为 `markdown-theme-{name}.css`。

**示例文件结构**：
```
frontend/app/
├── globals.css              ← 基础主题（Dracula Spectrum）
├── markdown-theme-light.css ← Light 主题覆盖
├── markdown-theme-github.css ← GitHub 风格主题
└── layout.tsx               ← 导入主题文件
```

### 2.2 编写主题内容

主题文件只需覆盖 `--md-*` 变量即可。**不需要复制所有变量**，只需覆盖你想要改变的变量。

**Light 主题示例**：

```css
/* frontend/app/markdown-theme-light.css */

:root {
  /* --- 正文颜色 --- */
  --md-color-text: #1a1a2e;
  --md-color-heading: #0d0d1a;
  --md-color-heading-2: #6b2fa0;
  --md-color-strong: #8b5cf6;
  --md-color-link: #7c3aed;
  --md-color-link-hover: #6d28d9;
  --md-color-emphasis: #64748b;
  
  /* --- 代码块 --- */
  --md-code-block-bg: #f8fafc;
  --md-code-block-text: #1e293b;
  --md-code-block-padding: 1rem;
  --md-code-block-margin: 1.5rem 0;
  
  /* --- 行内代码 --- */
  --md-inline-code-bg: #e2e8f0;
  --md-inline-code-color: #dc2626;
  --md-inline-code-border: 1px solid #cbd5e1;
  
  /* --- 引用块 --- */
  --md-blockquote-border: 4px solid #7c3aed;
  --md-blockquote-bg: linear-gradient(90deg, rgba(248,250,252,0.95), rgba(248,250,252,0.6));
  --md-blockquote-color: #7c3aed;
  
  /* --- 语法高亮 --- */
  --md-hl-comment: #94a3b8;
  --md-hl-keyword: #dc2626;
  --md-hl-title: #8b5cf6;
  --md-hl-string: #059669;
  --md-hl-number: #ea580c;
  --md-hl-builtin: #0891b2;
  --md-hl-function: #16a34a;
  --md-hl-meta: #0ea5e9;
  
  /* --- 表格 --- */
  --md-table-bg: #f8fafc;
  --md-table-border: 2px solid #e2e8f0;
  --md-table-header-bg: #f1f5f9;
  --md-table-header-color: #6b2fa0;
  --md-table-row-odd: rgba(248,250,252,0.98);
  --md-table-row-even: rgba(241,245,249,0.98);
  --md-table-row-hover: rgba(226,232,240,0.98);
}
```

### 2.3 完整主题变量参考

以下是可以覆盖的全部变量列表。**建议只覆盖必要的变量**，其他变量会自动继承默认值。

#### Typography（排版）
| 变量 | 默认值 | 说明 |
|---|---|---|
| `--md-font-size` | `1.05rem` | 正文字号 |
| `--md-line-height` | `1.9` | 正文行高 |
| `--md-font-body` | `var(--font-body)` | 正文字体 |
| `--md-font-headline` | `var(--font-headline)` | 标题字体 |
| `--md-font-mono` | `var(--font-mono)` | 代码字体 |

#### Colors — Text（文本颜色）
| 变量 | 默认值 | 说明 |
|---|---|---|
| `--md-color-text` | `var(--color-on-surface-variant)` | 正文颜色 |
| `--md-color-heading` | `var(--color-on-surface)` | H1/H3/H4 标题颜色 |
| `--md-color-heading-2` | `var(--color-secondary-fixed-dim)` | H2 标题颜色 |
| `--md-color-strong` | `var(--color-primary-fixed)` | 粗体颜色 |
| `--md-color-link` | `var(--color-primary)` | 链接颜色 |
| `--md-color-link-hover` | `var(--color-secondary)` | 链接悬停颜色 |
| `--md-color-emphasis` | `var(--color-outline)` | 斜体颜色 |

#### Code Blocks（代码块）
| 变量 | 默认值 | 说明 |
|---|---|---|
| `--md-code-block-bg` | `var(--color-surface-lowest)` | 代码块背景 |
| `--md-code-block-text` | `var(--color-on-surface)` | 代码文本颜色 |
| `--md-code-block-label-color` | `var(--color-outline-variant)` | 语言标签颜色 |
| `--md-code-block-padding` | `1.25rem` | 代码块内边距 |
| `--md-code-block-margin` | `1.5rem 0` | 代码块外边距 |
| `--md-code-font-size` | `0.875rem` | 代码字号 |
| `--md-code-line-height` | `1.85` | 代码行高 |

#### Inline Code（行内代码）
| 变量 | 默认值 | 说明 |
|---|---|---|
| `--md-inline-code-bg` | `rgba(39,41,53,0.92)` | 背景色 |
| `--md-inline-code-color` | `var(--color-tertiary-fixed)` | 文字颜色 |
| `--md-inline-code-border` | `1px solid rgba(74,68,81,0.7)` | 边框 |
| `--md-inline-code-padding` | `0.125rem 0.45rem` | 内边距 |
| `--md-inline-code-font-size` | `0.9rem` | 字号 |
| `--md-inline-code-shadow` | `inset 0 -1px 0 rgba(225,225,241,0.06)` | 阴影 |

#### Blockquotes（引用块）
| 变量 | 默认值 | 说明 |
|---|---|---|
| `--md-blockquote-border` | `4px solid var(--color-primary)` | 左边框 |
| `--md-blockquote-bg` | `linear-gradient(...)` | 背景 |
| `--md-blockquote-color` | `var(--color-primary)` | 文字颜色 |
| `--md-blockquote-padding` | `1rem 0 1rem 1.5rem` | 内边距 |
| `--md-blockquote-margin` | `1.5rem 0` | 外边距 |

#### Alerts / Callouts（警告框）
警告框使用 GitHub 风格的 `> [!NOTE]` 语法，由 `remark-github-blockquote-alert` 插件支持。

支持五种类型：`[!NOTE]`、`[!TIP]`、`[!IMPORTANT]`、`[!WARNING]`、`[!CAUTION]`

| 变量 | 默认值 | 说明 |
|---|---|---|
| `--md-alert-note-border` | `var(--color-primary-container)` | NOTE 边框色（紫色） |
| `--md-alert-note-bg` | `rgba(189,147,249,0.08)` | NOTE 背景色 |
| `--md-alert-note-title` | `var(--color-primary-fixed)` | NOTE 标题色 |
| `--md-alert-tip-border` | `var(--color-tertiary)` | TIP 边框色（绿色） |
| `--md-alert-tip-bg` | `rgba(49,227,104,0.08)` | TIP 背景色 |
| `--md-alert-tip-title` | `var(--color-tertiary-fixed)` | TIP 标题色 |
| `--md-alert-important-border` | `var(--color-secondary-fixed-dim)` | IMPORTANT 边框色（粉色） |
| `--md-alert-important-bg` | `rgba(255,175,215,0.08)` | IMPORTANT 背景色 |
| `--md-alert-important-title` | `var(--color-secondary-fixed)` | IMPORTANT 标题色 |
| `--md-alert-warning-border` | `#ffd6a5` | WARNING 边框色（琥珀色） |
| `--md-alert-warning-bg` | `rgba(255,214,165,0.08)` | WARNING 背景色 |
| `--md-alert-warning-title` | `#ffd6a5` | WARNING 标题色 |
| `--md-alert-caution-border` | `#ff6e6e` | CAUTION 边框色（红色） |
| `--md-alert-caution-bg` | `rgba(255,110,110,0.08)` | CAUTION 背景色 |
| `--md-alert-caution-title` | `#ff6e6e` | CAUTION 标题色 |
| `--md-alert-padding` | `1rem 1.25rem` | 警告框内边距 |
| `--md-alert-margin` | `1.5rem 0` | 警告框外边距 |
| `--md-alert-border-width` | `4px` | 左边框宽度 |
| `--md-alert-title-font-size` | `0.85rem` | 标题字号 |
| `--md-alert-body-font-size` | `0.94rem` | 正文字号 |

#### Syntax Highlighting（语法高亮）
| 变量 | 默认值 | 说明 |
|---|---|---|
| `--md-hl-comment` | `var(--color-outline)` | 注释 |
| `--md-hl-keyword` | `var(--color-secondary-fixed-dim)` | 关键字 |
| `--md-hl-title` | `var(--color-primary-fixed)` | 函数名/标题 |
| `--md-hl-string` | `#ffd6a5` | 字符串 |
| `--md-hl-number` | `#ff9e64` | 数字 |
| `--md-hl-builtin` | `#8be9fd` | 内置类型 |
| `--md-hl-function` | `var(--color-tertiary-fixed)` | 函数/属性 |
| `--md-hl-meta` | `#7dcfff` | 元信息 |

#### Tables（表格）
| 变量 | 默认值 | 说明 |
|---|---|---|
| `--md-table-bg` | `var(--color-surface-lowest)` | 表格背景 |
| `--md-table-border` | `2px solid var(--color-outline-variant)` | 边框 |
| `--md-table-header-bg` | `var(--color-surface-container-highest)` | 表头背景 |
| `--md-table-header-color` | `var(--color-secondary-fixed-dim)` | 表头颜色 |
| `--md-table-row-odd` | `rgba(29,31,43,0.92)` | 奇数行背景 |
| `--md-table-row-even` | `rgba(11,14,24,0.98)` | 偶数行背景 |
| `--md-table-row-hover` | `rgba(50,52,64,0.98)` | 悬停行背景 |

---

## 三、如何应用主题

### 3.1 方法一：直接导入（推荐）

在 `frontend/app/layout.tsx` 中导入主题文件，**必须在 `globals.css` 之后导入**。

```tsx
// frontend/app/layout.tsx

import "./globals.css";
import "./markdown-theme-light.css";  // ← 主题覆盖文件
```

**原理**：CSS 后导入的文件会覆盖先导入文件中相同选择器的属性。由于主题文件只重新定义 `:root` 中的 `--md-*` 变量，其他样式保持不变。

### 3.2 方法二：运行时切换（动态主题）

如果需要运行时切换主题（如用户选择 Light/Dark），可以通过 JavaScript 动态修改 CSS 变量：

```tsx
// 切换主题的函数
function applyTheme(theme: 'dark' | 'light') {
  const root = document.documentElement;
  
  if (theme === 'light') {
    root.style.setProperty('--md-color-text', '#1a1a2e');
    root.style.setProperty('--md-code-block-bg', '#f8fafc');
    root.style.setProperty('--md-inline-code-bg', '#e2e8f0');
    // ... 更多变量
  } else {
    // 重置为默认值（删除自定义值）
    root.style.removeProperty('--md-color-text');
    root.style.removeProperty('--md-code-block-bg');
    root.style.removeProperty('--md-inline-code-bg');
    // ...
  }
}
```

### 3.3 方法三：CSS 类切换

另一种动态切换方式是使用 CSS 类选择器：

```css
/* globals.css */

:root {
  --md-color-text: var(--color-on-surface-variant);  /* Dark 默认 */
}

:root.light-theme {
  --md-color-text: #1a1a2e;
  --md-code-block-bg: #f8fafc;
  /* ... */
}
```

```tsx
// 切换主题
function setTheme(theme: 'dark' | 'light') {
  document.documentElement.classList.remove('light-theme');
  if (theme === 'light') {
    document.documentElement.classList.add('light-theme');
  }
}
```

---

## 四、Alerts / Callouts 使用方法

Markdown 渲染支持 GitHub 风格的警告框语法，使用 blockquote 语法加特殊标记实现。

### 4.1 语法格式

在 Markdown 中使用 `> [!TYPE]` 语法创建警告框：

```markdown
> [!NOTE]
> 这是一条提示信息。

> [!TIP]
> 这是一个技巧建议。

> [!IMPORTANT]
> 这是一条重要信息。

> [!WARNING]
> 这是一个警告提醒。

> [!CAUTION]
> 这是一个危险警示。
```

### 4.2 支持的警告类型

| 类型 | 用途 | 色系 |
|---|---|---|
| `NOTE` | 有用的提示信息 | 紫色（Primary） |
| `TIP` | 有帮助的建议 | 绿色（Tertiary） |
| `IMPORTANT` | 需要关注的关键信息 | 粉色（Secondary） |
| `WARNING` | 警告性建议 | 琥珀色 |
| `CAUTION` | 潜在危险警示 | 红色 |

### 4.3 技术实现

- **插件**：`remark-github-blockquote-alert`（remark 插件）
- **HTML 输出**：`<div class="markdown-alert markdown-alert-{type}">`
- **CSS 类**：`.markdown-alert`（容器）、`.markdown-alert-title`（标题行）、`.markdown-alert-{note|tip|important|warning|caution}`（类型变体）
- **SVG 图标**：插件自动注入 GitHub Octicon SVG

---

## 五、主题编写最佳实践

### 5.1 保持一致性

- **颜色体系**：选择一套协调的颜色方案（如 Tailwind 颜色调色板）
- **对比度**：确保正文与背景对比度至少 4.5:1（WCAG AA 标准）
- **代码高亮**：语法高亮颜色应与整体色调一致

### 5.2 最小覆盖原则

只覆盖真正需要改变的变量。例如，如果只想改变代码块背景色：

```css
:root {
  --md-code-block-bg: #1e1e2e;
}
```

其他变量自动继承默认值。

### 5.3 语义化命名

使用有意义的前缀区分不同主题：

```
markdown-theme-dark.css     ← 深色主题（默认）
markdown-theme-light.css    ← 浅色主题
markdown-theme-github.css   ← GitHub 风格
markdown-theme-monokai.css  ← Monokai 风格
```

---

## 六、已修复的渲染问题

### 6.1 CSS 特异性问题

**问题**：`.markdown-body pre` 选择器（特异性 0-1-1）覆盖了 `.markdown-code-block__pre`（0-1-0），导致代码块 `<pre>` 元素获得额外的 `2rem` padding，总内边距达到 64px。

**修复**：将 `.markdown-body pre` 重置为 `padding: 0; margin: 0;`，由外层 `.markdown-code-block` 统一控制间距。

### 6.2 无语言代码块白框问题

**问题**：没有语言标注的代码块（如 ` ``` ` 空代码块）被错误判定为"行内代码"，获得 `markdown-inline-code` 类，导致内部出现 `1px solid` 边框。

**修复**：`MarkdownRenderer.tsx` 的 `code` 组件现在检测内容是否包含 `\n`，多行内容视为代码块而非行内代码。

---

## 七、示例主题

### 7.1 GitHub Light 风格

```css
:root {
  --md-color-text: #24292f;
  --md-color-heading: #24292f;
  --md-color-heading-2: #24292f;
  --md-color-strong: #24292f;
  --md-color-link: #0969da;
  --md-color-link-hover: #0550ae;
  
  --md-code-block-bg: #f6f8fa;
  --md-code-block-text: #24292f;
  --md-code-block-padding: 16px;
  
  --md-inline-code-bg: #afb8c133;
  --md-inline-code-color: #24292f;
  --md-inline-code-border: none;
  
  --md-blockquote-border: 4px solid #d0d7de;
  --md-blockquote-color: #57606a;
  
  --md-hl-keyword: #cf222e;
  --md-hl-string: #0a3069;
  --md-hl-comment: #6e7781;
}
```

### 7.2 Monokai 风格

```css
:root {
  --md-code-block-bg: #272822;
  --md-code-block-text: #f8f8f2;
  
  --md-inline-code-bg: #272822;
  --md-inline-code-color: #f8f8f2;
  --md-inline-code-border: 1px solid #49483e;
  
  --md-hl-keyword: #f92672;
  --md-hl-string: #e6db74;
  --md-hl-number: #ae81ff;
  --md-hl-comment: #75715e;
  --md-hl-title: #a6e22e;
  --md-hl-builtin: #66d9ef;
  --md-hl-function: #fd971f;
}
```

---

## 八、相关文档

- [Frontend Architecture](./architecture.md) — 前端架构设计
- [Frontend Implementation](./implementation.md) — 前端实现笔记
- [Architecture Refactoring](./architecture-refactoring.md) — 前端重构记录