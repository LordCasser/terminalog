# Terminalog - 前端架构设计文档

> 文档版本：v1.0
> 创建日期：2026-04-15
> 基于需求文档：requirements.md v1.1
> 关联文档：backend-architecture.md, api-spec.md

---

## 一、架构概览

### 1.1 前端定位

Terminalog 前端采用 **Next.js 静态导出** 模式，生成纯静态 HTML/CSS/JS 文件，最终通过 Go embed 嵌入后端二进制文件，实现单文件部署。

### 1.2 前端子系统架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        前端静态资源                                   │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    Next.js 静态导出产物                          │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │ │
│  │  │   HTML       │  │   CSS        │  │   JavaScript         │ │ │
│  │  │   (页面)      │  │   (样式)      │  │   (交互逻辑)          │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                         核心模块                                 │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │ │
│  │  │  Terminal UI │  │  Command     │  │  Markdown            │ │ │
│  │  │  (终端风格)    │  │  Parser      │  │  Renderer           │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │ │
│  │  │  Article     │  │  API Client  │  │  Path                │ │ │
│  │  │  Viewer      │  │              │  │  Transformer         │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ HTTP API
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         后端服务（Go）                                │
│  REST API: /api/articles, /api/assets, /api/search, /api/tree       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 二、模块划分与边界定义

### 2.1 前端模块总览

| 模块 | 负责人 | 职责边界 | 依赖关系 |
|------|--------|---------|---------|
| **Terminal UI** | 前端 | 终端风格 UI 组件，Dracula 配色实现 | shadcn/ui, Tailwind CSS |
| **Command Parser** | 前端 | 命令行输入解析与执行 | Terminal UI |
| **Markdown Renderer** | 前端 | Markdown 内容渲染（代码高亮、公式、Mermaid） | Markdown 解析库 |
| **Article Viewer** | 前端 | 文章详情页展示（时间线、元数据） | Markdown Renderer |
| **API Client** | 前端 | 与后端 API 通信 | Fetch API |
| **Path Transformer** | 前端 | 图片路径转换（相对路径 → API 路径） | Article Viewer |

### 2.2 模块职责详解

#### 2.2.1 Terminal UI 模块

**职责**：
- 实现终端风格的整体 UI 容器
- 应用 Dracula 配色方案
- 提供终端窗口装饰（标题栏、窗口按钮）
- 实现光标闪烁效果

**边界**：
- 不负责业务逻辑处理
- 不负责 API 调用

**组成组件**：
- `TerminalWindow.tsx`：终端窗口容器
- `TerminalHeader.tsx`：终端标题栏（显示当前路径）
- `TerminalContent.tsx`：终端内容区域（输出显示）
- `CommandInput.tsx`：命令输入框（底部）
- `Cursor.tsx`：闪烁光标组件
- `ArticleList.tsx`：终端风格的文章列表

#### 2.2.2 Command Parser 模块

**职责**：
- 解析用户输入的命令
- 提取命令名、参数、flags
- 执行对应命令逻辑
- 管理命令历史

**边界**：
- 不负责 UI 渲染
- 不负责直接 API 调用（通过 API Client）

**支持命令**：

| 命令 | 功能 | 参数/Flags |
|------|------|-----------|
| `cd <path>` | 切换目录 | `path`: 目标路径 |
| `ls` | 列出内容 | `--sort=created\|edited`, `--asc\|--desc` |
| `view <file>` | 查看文章 | `file`: 文件名 |
| `search <keyword>` | 搜索 | `keyword`: 搜索词 |
| `help` | 显示帮助 | 无 |
| `clear` | 清屏 | 无 |
| `exit` | 退出查看 | 无 |

#### 2.2.3 Markdown Renderer 模块

**职责**：
- 渲染 Markdown 内容
- 代码语法高亮
- 数学公式渲染
- Mermaid 流程图渲染
- 图片路径转换

**边界**：
- 不负责获取 Markdown 内容（由 Article Viewer 负责）
- 不负责 Git 历史展示

#### 2.2.4 Article Viewer 模块

**职责**：
- 文章详情页面展示
- 显示文章元数据（创建/编辑信息）
- 显示编辑时间线（可展开）
- 集成 Markdown Renderer

**边界**：
- 不负责命令解析
- 不负责文章列表展示

#### 2.2.5 API Client 模块

**职责**：
- 封装所有后端 API 调用
- 处理请求/响应错误
- 提供类型安全的 API 接口

**边界**：
- 不负责业务逻辑
- 不负责 UI 渲染

#### 2.2.6 Path Transformer 模块

**职责**：
- 识别 Markdown 中的图片路径类型
- 转换相对路径为 API 资源路径
- 保持外部链接不变

**边界**：
- 不负责图片加载（由浏览器处理）
- 不负责 API 调用

---

## 三、技术选型

### 3.1 核心技术栈

| 组件 | 推荐方案 | 版本 | 理由 |
|------|---------|------|------|
| 框架 | **Next.js** | 14+ | App Router，静态导出支持 |
| UI 库 | **shadcn/ui** | 最新 | 高质量组件，Tailwind 集成 |
| 样式 | **Tailwind CSS** | 3+ | 灵活定制，性能好 |
| Markdown | **react-markdown** | 9+ | 灵活扩展 |
| 代码高亮 | **rehype-highlight** | 最新 | 语法高亮 |
| 数学公式 | **KaTeX** + **rehype-katex** | 最新 | 比 MathJax 更快 |
| Mermaid | **mermaid-react** | 最新 | 流程图渲染 |
| 语言 | **TypeScript** | 5+ | 类型安全 |

### 3.2 辅助工具

| 工具 | 用途 |
|------|------|
| **pnpm** | 包管理（更快、更节省空间） |
| **ESLint** | 代码检查 |
| **Prettier** | 代码格式化 |

---

## 四、项目结构

```
frontend/
├── app/                           # Next.js App Router
│   ├── layout.tsx                 # 根布局（终端容器）
│   ├── page.tsx                   # 主页（文章列表）
│   ├── globals.css                # 全局样式
│   └── article/
│       └── [path]/
│           └── page.tsx           # 文章详情页（动态路由）
│
├── components/                    # React 组件
│   ├── ui/                        # shadcn/ui 组件（自动生成）
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── input.tsx
│   │   └── ...
│   │
│   ├── terminal/                  # 终端风格组件
│   │   ├── TerminalWindow.tsx     # 终端窗口容器
│   │   ├── TerminalHeader.tsx     # 终端标题栏
│   │   ├── TerminalContent.tsx    # 终端内容区域
│   │   ├── CommandInput.tsx       # 命令输入框
│   │   ├── Cursor.tsx             # 闪烁光标
│   │   └── ArticleList.tsx        # 终端风格文章列表
│   │
│   ├── article/                   # 文章相关组件
│   │   ├── ArticleView.tsx        # 文章查看器（全屏）
│   │   ├── ArticleMeta.tsx        # 文章元数据展示
│   │   ├── Timeline.tsx           # 编辑时间线（可展开）
│   │   ├── MarkdownRenderer.tsx   # Markdown 渲染器
│   │   └── PathTransformer.tsx    # 图片路径转换
│   │
│   ├── command/                   # 命令处理
│   │   ├── CommandParser.ts       # 命令解析（纯逻辑）
│   │   └── commands/              # 各命令实现
│   │       ├── cd.ts
│   │       ├── ls.ts
│   │       ├── view.ts
│   │       ├── search.ts
│   │       ├── help.ts
│   │       ├── clear.ts
│   │       └── exit.ts
│   │
│   └── common/                    # 通用组件
│       ├── SearchBar.tsx          # 搜索组件
│       └── DirectoryTree.tsx      # 目录树组件
│
├── lib/                           # 库/工具
│   ├── api/                       # API 客户端
│   │   ├── client.ts              # 基础 HTTP 客户端
│   │   ├── articles.ts            # 文章 API
│   │   ├── assets.ts              # 资源 API
│   │   └── tree.ts                # 目录树 API
│   │
│   ├── hooks/                     # React Hooks
│   │   ├── useArticles.ts         # 文章数据管理
│   │   ├── useCommand.ts          # 命令处理
│   │   ├── useCurrentPath.ts      # 当前路径管理
│   │   └── useTerminalState.ts    # 终端状态管理
│   │
│   └── utils/                     # 工具函数
│   │   ├── path-transform.ts      # 图片路径转换
│   │   ├── markdown-plugins.ts    # Markdown 插件配置
│   │   ├── terminal-format.ts     # 终端输出格式化
│   │   └── date-format.ts         # 日期格式化
│
├── styles/                        # 样式文件
│   ├── dracula.css                # Dracula 配色定义
│   └── terminal.css               # 终端样式
│
├── types/                         # TypeScript 类型定义
│   ├── article.ts                 # Article 类型
│   ├── commit.ts                  # Commit 类型
│   ├── tree.ts                    # Tree 类型
│   ├── command.ts                 # Command 类型
│   └── api.ts                     # API 响应类型
│
├── public/                        # 静态资源
│   └── fonts/                     # 字体文件
│
├── next.config.js                 # Next.js 配置
├── tailwind.config.js             # Tailwind 配置
├── components.json                # shadcn/ui 配置
├── tsconfig.json                  # TypeScript 配置
├── package.json
└── pnpm-lock.yaml
```

---

## 五、核心组件设计

### 5.1 终端窗口（TerminalWindow）

#### 5.1.1 组件结构

```
┌─────────────────────────────────────────────────────┐
│  ●  ●  ●    terminalog ~ ~/articles                 │  ← TerminalHeader
├─────────────────────────────────────────────────────┤
│                                                      │  ← TerminalContent
│  $ ls                                                │     (输出区域)
│  drwxr-xr-x  tech/           2024-01-15 10:30       │
│  -rw-r--r--  welcome.md       2024-01-10 09:00       │
│  -rw-r--r--  about.md         2024-01-08 15:20       │
│                                                      │
│  $ cd tech                                           │
│                                                      │
│  ~/articles/tech $ ls                                │
│  -rw-r--r--  golang.md      2024-01-12 14:00        │
│  -rw-r--r--  rust.md        2024-01-11 12:30        │
│                                                      │
├─────────────────────────────────────────────────────┤
│  $ _                                                │  ← CommandInput + Cursor
│                                                      │     (输入区域)
└─────────────────────────────────────────────────────┘
```

#### 5.1.2 状态管理

```typescript
// types/command.ts

interface TerminalState {
  currentPath: string;           // 当前路径，如 "tech"
  history: string[];             // 命令历史记录
  output: OutputLine[];          // 输出内容列表
  mode: 'list' | 'view';         // 当前模式：列表或查看文章
  viewingArticle?: Article;      // 正在查看的文章（view 模式）
  isLoading: boolean;            // 加载状态
  error?: string;                // 错误信息
}

interface OutputLine {
  id: string;                    // 唯一标识
  type: 'command' | 'result' | 'error' | 'info';
  content: string;               // 内容文本
  timestamp?: Date;              // 时间戳（可选）
}

interface Command {
  name: string;                  // 命令名：cd, ls, view, etc.
  args: string[];                // 参数列表
  flags: Record<string, string | boolean>; // flags
  raw: string;                   // 原始输入字符串
}
```

#### 5.1.3 状态管理 Hook

```typescript
// lib/hooks/useTerminalState.ts

import { useState, useCallback } from 'react';
import { TerminalState, OutputLine, Command } from '@/types/command';
import { parseCommand } from '@/components/command/CommandParser';
import { executeCommand } from '@/components/command/commands';

export function useTerminalState() {
  const [state, setState] = useState<TerminalState>({
    currentPath: '',
    history: [],
    output: [],
    mode: 'list',
    isLoading: false,
  });

  // 执行命令
  const execute = useCallback(async (input: string) => {
    const command = parseCommand(input);
    
    // 添加命令到输出
    addOutput({ type: 'command', content: `$ ${input}` });
    
    // 执行命令
    setState(prev => ({ ...prev, isLoading: true }));
    
    try {
      const newState = await executeCommand(command, state);
      setState(newState);
    } catch (error) {
      addOutput({ type: 'error', content: error.message });
    } finally {
      setState(prev => ({ ...prev, isLoading: false }));
    }
  }, [state]);

  // 添加输出
  const addOutput = useCallback((line: Omit<OutputLine, 'id'>) => {
    setState(prev => ({
      ...prev,
      output: [...prev.output, { ...line, id: generateId() }]
    }));
  }, []);

  // 清屏
  const clear = useCallback(() => {
    setState(prev => ({ ...prev, output: [] }));
  }, []);

  // 切换路径
  const changePath = useCallback((path: string) => {
    setState(prev => ({ ...prev, currentPath: path }));
  }, []);

  return {
    state,
    execute,
    addOutput,
    clear,
    changePath,
  };
}
```

### 5.2 命令解析器（CommandParser）

```typescript
// components/command/CommandParser.ts

import { Command } from '@/types/command';

/**
 * 解析用户输入的命令字符串
 * 
 * @example
 * parseCommand("ls --sort=created --desc")
 * // 返回: { name: 'ls', args: [], flags: { sort: 'created', desc: true }, raw: '...' }
 * 
 * parseCommand("cd tech/blog")
 * // 返回: { name: 'cd', args: ['tech/blog'], flags: {}, raw: '...' }
 * 
 * parseCommand("view article.md")
 * // 返回: { name: 'view', args: ['article.md'], flags: {}, raw: '...' }
 */
export function parseCommand(input: string): Command {
  const trimmed = input.trim();
  
  if (!trimmed) {
    return { name: '', args: [], flags: {}, raw: input };
  }

  // 分割命令
  const parts = trimmed.split(/\s+/);
  const name = parts[0];
  const rest = parts.slice(1);

  // 解析参数和 flags
  const args: string[] = [];
  const flags: Record<string, string | boolean> = {};

  for (const part of rest) {
    if (part.startsWith('--')) {
      // Flag 格式: --key=value 或 --key
      const flagPart = part.slice(2);
      if (flagPart.includes('=')) {
        const [key, value] = flagPart.split('=');
        flags[key] = value;
      } else {
        flags[flagPart] = true;
      }
    } else if (part.startsWith('-')) {
      // 短 flag: -k=value 或 -k
      const flagPart = part.slice(1);
      if (flagPart.includes('=')) {
        const [key, value] = flagPart.split('=');
        flags[key] = value;
      } else {
        flags[flagPart] = true;
      }
    } else {
      // 参数
      args.push(part);
    }
  }

  return {
    name,
    args,
    flags,
    raw: input,
  };
}

/**
 * 生成命令帮助文本
 */
export function getCommandHelp(commandName: string): string {
  const helps: Record<string, string> = {
    cd: 'cd <path> - 切换到指定目录',
    ls: 'ls [--sort=created|edited] [--asc|--desc] - 列出当前目录内容',
    view: 'view <filename> - 全屏查看文章',
    search: 'search <keyword> - 搜索文章标题',
    help: 'help [command] - 显示帮助信息',
    clear: 'clear - 清屏',
    exit: 'exit - 退出文章查看模式',
  };

  return helps[commandName] || `Unknown command: ${commandName}`;
}
```

### 5.3 命令实现示例

```typescript
// components/command/commands/ls.ts

import { Command, TerminalState, OutputLine } from '@/types/command';
import { getArticles } from '@/lib/api/articles';
import { formatArticleList } from '@/lib/utils/terminal-format';

export async function executeLs(
  command: Command,
  state: TerminalState
): Promise<TerminalState> {
  // 解析排序参数
  const sort = (command.flags.sort as string) || 'edited';
  const order = command.flags.desc ? 'desc' : (command.flags.asc ? 'asc' : 'desc');
  
  // 调用 API
  const response = await getArticles({
    dir: state.currentPath,
    sort,
    order,
  });

  // 格式化输出
  const formatted = formatArticleList(response.articles);
  
  // 返回新状态
  return {
    ...state,
    output: [
      ...state.output,
      { id: generateId(), type: 'result', content: formatted }
    ],
  };
}
```

```typescript
// components/command/commands/cd.ts

import { Command, TerminalState } from '@/types/command';
import { getTree } from '@/lib/api/tree';

export async function executeCd(
  command: Command,
  state: TerminalState
): Promise<TerminalState> {
  const targetPath = command.args[0];
  
  if (!targetPath) {
    return {
      ...state,
      output: [
        ...state.output,
        { id: generateId(), type: 'error', content: 'cd: missing directory argument' }
      ],
    };
  }

  // 处理路径
  let newPath: string;
  
  if (targetPath === '..') {
    // 返回上级目录
    newPath = state.currentPath.split('/').slice(0, -1).join('/');
  } else if (targetPath === '/' || targetPath === '~') {
    // 回到根目录
    newPath = '';
  } else {
    // 进入子目录（验证是否存在）
    const checkPath = state.currentPath 
      ? `${state.currentPath}/${targetPath}` 
      : targetPath;
    
    const tree = await getTree({ dir: state.currentPath });
    const exists = tree.children.some(
      c => c.name === targetPath && c.type === 'dir'
    );
    
    if (!exists) {
      return {
        ...state,
        output: [
          ...state.output,
          { id: generateId(), type: 'error', content: `cd: ${targetPath}: No such directory` }
        ],
      };
    }
    
    newPath = checkPath;
  }

  return {
    ...state,
    currentPath: newPath,
    output: [
      ...state.output,
      { id: generateId(), type: 'info', content: '' } // cd 成功无输出
    ],
  };
}
```

### 5.4 Markdown 渲染器

```typescript
// components/article/MarkdownRenderer.tsx

import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import rehypeKatex from 'rehype-katex';
import remarkMath from 'remark-math';
import remarkGfm from 'remark-gfm';
import { Mermaid } from 'mermaid-react';
import { transformImagePath } from '@/lib/utils/path-transform';
import 'highlight.js/styles/dracula.css';
import 'katex/dist/katex.min.css';

interface MarkdownRendererProps {
  content: string;
  basePath: string;  // 文章所在目录路径，用于图片路径转换
}

export function MarkdownRenderer({ content, basePath }: MarkdownRendererProps) {
  return (
    <div className="markdown-body">
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkMath]}
        rehypePlugins={[rehypeHighlight, rehypeKatex]}
        components={{
          // 图片路径转换
          img: ({ src, alt, ...props }) => {
            if (!src) return null;
            const transformedSrc = transformImagePath(src, basePath);
            return (
              <img 
                src={transformedSrc} 
                alt={alt} 
                loading="lazy"
                className="max-w-full h-auto"
                {...props} 
              />
            );
          },
          
          // Mermaid 流程图
          code: ({ className, children, ...props }) => {
            const match = /language-(\w+)/.exec(className || '');
            const lang = match ? match[1] : '';
            
            if (lang === 'mermaid') {
              return (
                <Mermaid 
                  chart={String(children).trim()}
                  config={{
                    theme: 'dark',
                    themeVariables: {
                      darkMode: true,
                      background: '#282a36',
                      primaryColor: '#ff79c6',
                      ...
                    }
                  }}
                />
              );
            }
            
            return (
              <code className={className} {...props}>
                {children}
              </code>
            );
          },
          
          // 链接处理
          a: ({ href, children, ...props }) => {
            if (href?.startsWith('http')) {
              return (
                <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
                  {children}
                </a>
              );
            }
            return <a href={href} {...props}>{children}</a>;
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
```

### 5.5 图片路径转换

```typescript
// lib/utils/path-transform.ts

/**
 * 转换 Markdown 中的图片路径
 * 
 * 规则：
 * 1. 外部链接（http/https）：保持原样
 * 2. 相对路径：转换为 API 资源路径
 * 
 * @example
 * transformImagePath("./images/photo.png", "tech/blog")
 * // 返回: "/api/assets/tech/blog/images/photo.png"
 * 
 * transformImagePath("https://cdn.example.com/img.png", "tech/blog")
 * // 返回: "https://cdn.example.com/img.png"
 */
export function transformImagePath(src: string, basePath: string): string {
  // 外部链接：保持原样
  if (src.startsWith('http://') || src.startsWith('https://')) {
    return src;
  }
  
  // 绝对路径（相对于仓库根）
  if (src.startsWith('/')) {
    return `/api/assets${src}`;
  }
  
  // 相对路径：转换为 API 路径
  // 处理 ./ 和 ../
  let normalizedPath = src;
  
  if (normalizedPath.startsWith('./')) {
    normalizedPath = normalizedPath.slice(2);
  }
  
  // 组合 basePath 和图片路径
  const fullPath = basePath 
    ? `${basePath}/${normalizedPath}` 
    : normalizedPath;
  
  return `/api/assets/${fullPath}`;
}
```

### 5.6 文章查看器（ArticleView）

```typescript
// components/article/ArticleView.tsx

import { Article, CommitInfo } from '@/types/article';
import { MarkdownRenderer } from './MarkdownRenderer';
import { ArticleMeta } from './ArticleMeta';
import { Timeline } from './Timeline';
import { useState } from 'react';

interface ArticleViewProps {
  article: Article;
  content: string;
  timeline: CommitInfo[];
  onClose: () => void;
}

export function ArticleView({ article, content, timeline, onClose }: ArticleViewProps) {
  const [showTimeline, setShowTimeline] = useState(false);
  
  // 从文章路径提取目录路径（用于图片路径转换）
  const basePath = article.path.replace(/\/[^\/]+\.md$/, '');

  return (
    <div className="article-view h-full overflow-auto">
      {/* 关闭按钮 */}
      <button 
        onClick={onClose}
        className="close-btn"
      >
        × (exit)
      </button>
      
      {/* 文章元数据 */}
      <ArticleMeta 
        createdBy={article.createdBy}
        createdAt={article.createdAt}
        editedBy={article.editedBy}
        editedAt={article.editedAt}
        contributors={article.contributors}
      />
      
      {/* 文章内容 */}
      <div className="article-content">
        <MarkdownRenderer 
          content={content} 
          basePath={basePath}
        />
      </div>
      
      {/* 编辑时间线 */}
      <div className="timeline-section">
        <button 
          onClick={() => setShowTimeline(!showTimeline)}
          className="timeline-toggle"
        >
          [{showTimeline ? '收起' : '展开'}] 编辑时间线 ({timeline.length} commits)
        </button>
        
        {showTimeline && (
          <Timeline commits={timeline} />
        )}
      </div>
    </div>
  );
}
```

### 5.7 编辑时间线组件

```typescript
// components/article/Timeline.tsx

import { CommitInfo } from '@/types/article';
import { formatDate } from '@/lib/utils/date-format';

interface TimelineProps {
  commits: CommitInfo[];
}

export function Timeline({ commits }: TimelineProps) {
  return (
    <div className="timeline">
      {commits.map((commit, index) => (
        <div key={commit.hash} className="timeline-item">
          <span className="commit-hash">{commit.hash}</span>
          <span className="separator">|</span>
          <span className="commit-date">{formatDate(commit.timestamp)}</span>
          <span className="separator">|</span>
          <span className="commit-author">{commit.author}</span>
          
          {/* 时间线连接线 */}
          {index < commits.length - 1 && (
            <div className="timeline-line" />
          )}
        </div>
      ))}
    </div>
  );
}
```

---

## 六、API 客户端设计

### 6.1 基础客户端

```typescript
// lib/api/client.ts

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || '';

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  headers?: Record<string, string>;
  body?: unknown;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
  }

  async request<T>(
    path: string,
    options: RequestOptions = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${path}`;
    
    const response = await fetch(url, {
      method: options.method || 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      body: options.body ? JSON.stringify(options.body) : undefined,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  async get<T>(path: string): Promise<T> {
    return this.request<T>(path);
  }

  async post<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>(path, { method: 'POST', body });
  }
}

export const apiClient = new ApiClient();
```

### 6.2 文章 API

```typescript
// lib/api/articles.ts

import { apiClient } from './client';
import { Article, ArticleListResponse, ArticleResponse } from '@/types/article';

interface GetArticlesParams {
  dir?: string;
  sort?: 'created' | 'edited';
  order?: 'asc' | 'desc';
}

export async function getArticles(params: GetArticlesParams = {}): Promise<ArticleListResponse> {
  const query = new URLSearchParams();
  
  if (params.dir) query.set('dir', params.dir);
  if (params.sort) query.set('sort', params.sort);
  if (params.order) query.set('order', params.order);
  
  return apiClient.get<ArticleListResponse>(`/api/articles?${query}`);
}

export async function getArticle(path: string): Promise<ArticleResponse> {
  return apiClient.get<ArticleResponse>(`/api/articles/${encodeURIComponent(path)}`);
}

export async function getArticleTimeline(path: string): Promise<{ commits: CommitInfo[] }> {
  return apiClient.get(`/api/articles/${encodeURIComponent(path)}/timeline`);
}
```

### 6.3 目录树 API

```typescript
// lib/api/tree.ts

import { apiClient } from './client';
import { TreeNode } from '@/types/tree';

interface GetTreeParams {
  dir?: string;
}

export async function getTree(params: GetTreeParams = {}): Promise<{ tree: TreeNode }> {
  const query = params.dir ? `?dir=${encodeURIComponent(params.dir)}` : '';
  return apiClient.get(`/api/tree${query}`);
}
```

### 6.4 搜索 API

```typescript
// lib/api/search.ts

import { apiClient } from './client';

interface SearchResult {
  path: string;
  title: string;
  matchedTitle: string;
}

interface SearchParams {
  q: string;
  dir?: string;
}

export async function searchArticles(params: SearchParams): Promise<{
  results: SearchResult[];
  total: number;
}> {
  const query = new URLSearchParams();
  query.set('q', params.q);
  if (params.dir) query.set('dir', params.dir);
  
  return apiClient.get(`/api/search?${query}`);
}
```

---

## 七、样式设计

### 7.1 Dracula 配色方案

```css
/* styles/dracula.css */

:root {
  /* Dracula Theme Colors */
  --dracula-background: #282a36;
  --dracula-current-line: #44475a;
  --dracula-foreground: #f8f8f2;
  --dracula-comment: #6272a4;
  --dracula-cyan: #8be9fd;
  --dracula-green: #50fa7b;
  --dracula-orange: #ffb86c;
  --dracula-pink: #ff79c6;
  --dracula-purple: #bd93f9;
  --dracula-red: #ff5555;
  --dracula-yellow: #f1fa8c;
}

/* 应用到终端 */
.terminal {
  background-color: var(--dracula-background);
  color: var(--dracula-foreground);
  font-family: 'Fira Code', 'JetBrains Mono', 'Consolas', monospace;
}
```

### 7.2 终端样式

```css
/* styles/terminal.css */

.terminal-window {
  background: var(--dracula-background);
  border-radius: 8px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

.terminal-header {
  background: var(--dracula-current-line);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-button {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.terminal-button-close { background: var(--dracula-red); }
.terminal-button-minimize { background: var(--dracula-yellow); }
.terminal-button-maximize { background: var(--dracula-green); }

.terminal-title {
  color: var(--dracula-comment);
  margin-left: auto;
  font-size: 14px;
}

.terminal-content {
  padding: 16px;
  min-height: 400px;
  max-height: 600px;
  overflow-y: auto;
}

.terminal-input-area {
  display: flex;
  align-items: center;
  padding: 16px;
  background: var(--dracula-background);
}

.terminal-prompt {
  color: var(--dracula-green);
  margin-right: 8px;
}

.terminal-input {
  background: transparent;
  border: none;
  color: var(--dracula-foreground);
  font-family: inherit;
  font-size: inherit;
  width: 100%;
  outline: none;
}

.cursor {
  display: inline-block;
  width: 10px;
  height: 18px;
  background: var(--dracula-foreground);
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}
```

### 7.3 文章列表样式

```css
/* styles/terminal.css (continued) */

.article-list {
  font-size: 14px;
  line-height: 1.6;
}

.article-list-item {
  display: flex;
  gap: 16px;
  padding: 4px 0;
  cursor: pointer;
  transition: background 0.2s;
}

.article-list-item:hover {
  background: var(--dracula-current-line);
}

.article-list-type {
  color: var(--dracula-comment);
  width: 80px;
}

.article-list-name {
  color: var(--dracula-cyan);
  flex: 1;
}

.article-list-name.directory {
  color: var(--dracula-purple);
}

.article-list-date {
  color: var(--dracula-comment);
  width: 150px;
}
```

---

## 八、Next.js 配置

### 8.1 静态导出配置

```javascript
// next.config.js

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',              // 启用静态导出
  trailingSlash: true,           // URL 带 trailing slash
  images: {
    unoptimized: true,           // 静态导出不支持图片优化
  },
  basePath: '',                  // 无 basePath（与后端同源）
  reactStrictMode: true,
  
  // 静态导出时，动态路由需要 generateStaticParams
  // 但我们的文章是动态加载的，所以前端只提供框架
};

module.exports = nextConfig;
```

### 8.2 Tailwind 配置

```javascript
// tailwind.config.js

/** @type {import('tailwind').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    './app/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        // Dracula 配色
        dracula: {
          background: '#282a36',
          current: '#44475a',
          foreground: '#f8f8f2',
          comment: '#6272a4',
          cyan: '#8be9fd',
          green: '#50fa7b',
          orange: '#ffb86c',
          pink: '#ff79c6',
          purple: '#bd93f9',
          red: '#ff5555',
          yellow: '#f1fa8c',
        },
      },
      fontFamily: {
        mono: ['Fira Code', 'JetBrains Mono', 'Consolas', 'monospace'],
      },
    },
  },
  plugins: [],
};
```

---

## 九、数据流与交互流程

### 9.1 浏览文章列表（鼠标点击）

```
┌──────────────┐
│    用户       │
│  点击目录     │
└──────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              TerminalContent                  │
│  onClick 触发                                 │
│  调用 changePath(newPath)                    │
└──────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              useArticles Hook                 │
│  自动触发 API 请求                             │
│  getArticles({ dir: newPath })               │
└──────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              API Client                       │
│  GET /api/articles?dir=newPath               │
└──────────────────────────────────────────────┘
       │
       ▼
返回文章列表 JSON
       │
       ▼
┌──────────────────────────────────────────────┐
│              TerminalContent                  │
│  更新 articles 数据                           │
│  重新渲染列表                                 │
└──────────────────────────────────────────────┘
```

### 9.2 浏览文章列表（命令行）

```
┌──────────────┐
│    用户       │
│ 输入 ls 命令  │
└──────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              CommandInput                     │
│  捕获 input                                   │
│  onSubmit(input)                             │
└──────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              useTerminalState.execute()       │
│  parseCommand("ls --sort=created")           │
└──────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              ls 命令实现                       │
│  getArticles({ sort: 'created', ... })       │
└──────────────────────────────────────────────┘
       │
       ▼
返回文章列表
       │
       ▼
┌──────────────────────────────────────────────┐
│              formatArticleList()              │
│  格式化为终端风格输出                          │
└──────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              addOutput()                      │
│  添加格式化结果到 output                      │
└──────────────────────────────────────────────┘
```

### 9.3 查看文章详情

```
┌──────────────┐
│    用户       │
│ view xxx.md  │
└──────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              view 命令实现                    │
│  state.mode = 'view'                         │
│  调用 API 获取文章内容和时间线                 │
└──────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              API Client                       │
│  GET /api/articles/xxx.md                    │
│  GET /api/articles/xxx.md/timeline           │
└──────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              ArticleView                      │
│  显示文章元数据                               │
│  MarkdownRenderer 渲染内容                    │
│  PathTransformer 转换图片路径                 │
│  Timeline 显示编辑历史                        │
└──────────────────────────────────────────────┘
```

---

## 十、性能优化

### 10.1 首屏加载优化

| 策略 | 实现方式 |
|------|---------|
| 静态资源压缩 | Next.js 构建自动压缩 JS/CSS |
| 代码分割 | Next.js 自动按页面分割 |
| 字体优化 | 使用系统字体或 preload |
| CSS 优化 | Tailwind JIT 模式，仅生成使用到的样式 |

### 10.2 Markdown 渲染优化

| 策略 | 实现方式 |
|------|---------|
| 图片懒加载 | `<img loading="lazy">` |
| Mermaid 延迟渲染 | 可见时才渲染流程图 |
| 大文件分块渲染 | （可选）虚拟滚动 |

---

## 十一、构建流程

### 11.1 开发模式

```bash
cd frontend
pnpm install
pnpm dev        # 启动 Next.js 开发服务器 (localhost:3000)
```

### 11.2 生产构建

```bash
cd frontend
pnpm build      # 静态导出，生成 out/ 目录
```

### 11.3 集成到后端

```bash
# 构建完成后，复制到 Go embed 目录
cp -r frontend/out/* ../pkg/embed/static/
```

---

## 十二、后续迭代规划

### 12.1 MVP（当前版本）

- ✅ 终端风格 UI（Dracula 配色）
- ✅ 命令行交互（cd, ls, view, search, help, clear, exit）
- ✅ 鼠标点击导航
- ✅ Markdown 渲染（代码高亮、公式、Mermaid、图片）
- ✅ 编辑时间线展示
- ✅ 光标闪烁效果

### 12.2 后续迭代

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 命令历史上下键 | 中 | 支持 ↑↓ 浏览历史命令 |
| 命令自动补全 | 低 | Tab 补全路径和命令 |
| 更多终端命令 | 低 | pwd, cat, tree 等 |
| 文章目录 TOC | 低 | 长文章目录导航 |
| 深色/浅色主题切换 | 低 | 支持非终端风格模式 |

---

**文档结束**

> 本前端架构设计基于 requirements.md v1.1
> 关联文档：backend-architecture.md, api-spec.md
> 下一步：进入实现阶段（Coder 模式）