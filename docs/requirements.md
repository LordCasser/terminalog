# Terminalog - 原始需求文档

> 文档版本：v1.5
> 创建日期：2026-04-15
> 最后更新：2026-04-17
> 项目名称：Terminalog（Terminal + Blog）
> 状态：✅ 需求 v1.5 对齐完成（v1.6.1 功能实现）

---

## 一、项目概述

Terminalog 是一个模拟终端风格的现代化 Blog 系统，采用单文件部署架构，前端编译后通过 Go embed 嵌入后端二进制文件，实现开箱即用的单文件启动体验。

### 1.1 核心特性

- **Brutalist 编辑器 UI**：采用 Dracula Spectrum 配色 + Brutalist 风格，提供现代化的终端美学
- **版本号自动生成**：基于文件行数变化自动计算语义版本号（v10.10.10）
- **单文件部署**：前端资源嵌入 Go 二进制，无需额外部署步骤
- **Git 原生**：内置轻量级 Git 服务，支持 HTTP/HTTPS 协议（SSH 协议后续迭代）
- **动态渲染**：页面访问时动态读取 Markdown 文件，实时响应内容变更
- **三字体系统**：Space Grotesk（标题）+ JetBrains Mono（UI）+ Inter（正文）
- **跨平台支持**：支持 Linux、macOS、Windows

---

## 二、技术架构约束

### 2.1 前端技术栈

| 组件 | 技术选型 | 版本要求 |
|------|---------|---------|
| 框架 | Next.js | 最新稳定版 |
| UI 组件库 | shadcn/ui | - |
| 样式方案 | Tailwind CSS | - |
| Markdown 渲染 | 待定（支持代码高亮、数学公式、Mermaid） | - |

### 2.2 后端技术栈

| 组件 | 技术选型 | 版本要求 |
|------|---------|---------|
| 语言 | Go | ≥ 1.21 |
| HTTP 框架 | 待定 | - |
| Git 服务 | 内置轻量级实现 | - |
| 配置格式 | TOML | - |

### 2.3 部署架构

```
┌─────────────────────────────────────────┐
│         单一可执行文件（terminalog）        │
│  ┌────────────────────────────────────┐ │
│  │      前端静态资源（embed）            │ │
│  │  - Next.js 静态导出产物              │ │
│  │  - CSS/JS/图片资源                  │ │
│  └────────────────────────────────────┘ │
│  ┌────────────────────────────────────┐ │
│  │      后端服务                        │ │
│  │  - HTTP API Server                 │ │
│  │  - Git Smart HTTP Server           │ │
│  │  - Markdown 解析服务                 │ │
│  │  - 静态资源服务                      │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### 2.4 前端构建方式

- **构建模式**：Next.js 静态导出（`output: 'export'`）
- **产物类型**：纯静态 HTML/CSS/JS
- **嵌入方式**：Go embed 嵌入二进制文件

---

## 三、功能需求详述

### 3.1 Blog 内容管理

#### 3.1.1 内容存储

- **存储方式**：纯文件系统 + Git，无数据库
- **内容目录**：用户指定的文件夹即为 Blog 内容根目录，同时也是 Git 仓库根目录
- **文件格式**：Markdown 文件（`.md` 扩展名）
- **组织结构**：直接映射文件系统目录结构，子文件夹即为分类

#### 3.1.2 Markdown 解析功能

| 功能 | 支持要求 |
|------|---------|
| 标准 Markdown | 必须支持 |
| 代码高亮 | 必须支持，支持多种语言 |
| 数学公式 | 必须支持（LaTeX 语法） |
| Mermaid 流程图 | 必须支持 |
| 图片渲染 | 必须支持 |

#### 3.1.3 图片处理规则

**图片存储位置**：图片可存放在 Git 仓库中，与 Markdown 文件同目录或子目录。

**图片路径处理**：

| 图片类型 | 处理方式 |
|---------|---------|
| 相对路径图片（如 `./images/photo.png`） | 前端渲染时，将路径转换为后端 API 访问路径，从 Git 仓库读取图片资源 |
| 外部链接图片（如 `https://cdn.example.com/img.png`） | 保留原始格式，直接加载外部资源 |

**图片访问流程**：
```
Markdown: ![](./images/photo.png)
    ↓ 前端渲染时转换
前端请求: /api/assets/articles/xxx/images/photo.png
    ↓ 后端处理
后端读取: Git 仓库中对应路径的图片文件
    ↓ 返回
返回图片二进制数据
```

#### 3.1.4 特殊文件处理

以下划线 `_` 开头的文件为**特殊文件**，不参与文章列表展示。

| 文件名 | 用途 | 访问方式 |
|--------|------|---------|
| `_ABOUTME.md` | About Me 页面内容 | 通过顶部导航栏 "ABOUTME" 链接访问，或 API `/api/aboutme` |
| 其他 `_*.md` | 预留特殊文件 | 暂不处理 |

**文章列表规则**：**排除**所有以 `_` 开头的 Markdown 文件。例如 `ls` 命令或 `/api/articles` 不返回 `_ABOUTME.md`。

#### 3.1.5 元数据规则

- **无 Frontmatter**：Markdown 文件不包含 YAML/JSON 元数据头
- **元数据来源**：全部从 Git 历史获取
- **元数据类型**：
  - 创建时间：文件的第一个 Git commit 时间
  - 创建人：文件的第一个 Git commit 作者
  - 最后编辑时间：文件的最后一个 Git commit 时间
  - 最后编辑人：文件的最后一个 Git commit 作者
  - 所有贡献者：所有参与该文件编辑的 Git commit 作者

### 3.2 Git 集成服务

#### 3.2.1 Git 协议支持

| 协议 | 支持状态 | 端口配置 | 备注 |
|------|---------|---------|------|
| HTTP/HTTPS | ✅ 必须支持 | 与 Blog 服务共享端口 | 认证支持，push 需验证 |
| SSH | ⏳ 后续迭代 | 独立端口（可配置） | MVP 版本暂不支持 |

**HTTPS 说明**：后端不内置 TLS 支持，由反向代理（Nginx/Caddy）处理 HTTPS 终结。

#### 3.2.2 认证机制

- **认证方式**：用户名 + 密码（TOML 配置文件）
- **配置方式**：
  - 用户手动在配置文件中定义用户名和密码
  - 或系统首次启动时自动生成默认 `admin` 用户和随机密码，写入配置文件
- **权限模型**：
  - Blog 访问：完全公开，无需认证
  - Git Clone：公开，无需认证
  - Git Push：仅配置文件中定义的用户可 push，需 Basic Auth 认证

#### 3.2.3 Git 服务架构

```
┌──────────────────────────────────────────────────────┐
│                   Terminalog 服务                      │
│  ┌──────────────────┐  ┌──────────────────────────┐  │
│  │   Blog Web UI    │  │     Git HTTP API         │  │
│  │   (Next.js)      │  │     (Smart HTTP)          │  │
│  │   - 静态资源服务    │  │     - git-upload-pack    │  │
│  │   - Markdown渲染  │  │     - git-receive-pack   │  │
│  └──────────────────┘  └──────────────────────────┘  │
│         │                          │                   │
│         └──────────────────────────┘                   │
│                      │                                 │
│            ┌─────────▼─────────┐                       │
│            │  Git Repository   │                       │
│            │  (Blog 内容目录)    │                       │
│            │  - Markdown 文件   │                       │
│            │  - 图片资源        │                       │
│            └───────────────────┘                       │
└──────────────────────────────────────────────────────┘
```

#### 3.2.4 文件状态与可见性规则

| 场景 | 处理规则 |
|------|---------|
| 文件存在但无 Git 提交 | **不展示**该文件 |
| 文件成功 push | 正常展示 |
| 文件删除 | 从 Blog 中移除 |
| 文件删除后重新添加（同名） | 视为**新文件**，创建时间重新计算 |
| Git 仓库迁移（保留 .git） | 提交历史保留，日期不变 |
| 仅搬运文件并重新提交 | 全量重算日期 |

### 3.3 前端展示需求

#### 3.3.1 设计风格：Brutalist 编辑器

**配色方案**：Dracula Spectrum（升级自经典 Dracula 10 色）

```
核心颜色:
  background:   #282a36
  surface-1:    #44475a      → surface-5: #6272a4
  comment:      #6272a4
  green/blue:   #8be9fd
  green/dull:   #50fa7b
  orange:       #ffb86c
  pink:         #ff79c6
  pink/soft:    #e589d6
  purple:       #bd93f9
  red:          #ff5555
  yellow:       #f1fa8c

Glass 效果:
  bg: rgba(68, 71, 90, 0.42)
  backdrop-filter: blur(8px)
  border: 1px solid rgba(139, 233, 253, 0.25)
```

**字体系统**（三字体）：

| 用途 | 字体 | 示例 |
|------|------|------|
| 标题/大数字 | Space Grotesk | "Terminalog"、"1,200"、"9" |
| 正文/内容渲染 | Inter | 文章段落文本 |
| UI/等宽元素 | JetBrains Mono | 代码、命令行、标签 |

**设计风格**：
- **0px 圆角**：Brutalist 风格，不使用 `rounded`
- **Glass 效果**：关键 UI 元素使用毛玻璃透明效果
- **无终端窗口装饰**：无标题栏、无窗口按钮装饰
- **渐变下划线**：主标题下方的装饰线使用 `#f1fa8c → #ff79c6 → #bd93f9 → #50fa7b → #8be9fd` 渐变

#### 3.3.2 交互方式：鼠标为主，命令为辅

**交互模型**：
1. **顶部导航栏**：Logo（可点击返回）、路径显示、搜索icon（点击后聚焦终端输入框并自动填入 `search ` 命令）
2. **鼠标点击为主要交互方式**：目录点击、文章点击、排序点击、底部命令行点击
3. **底部单行命令输入区**：显示 `guest@blog: ~/path $ ` 前缀，支持实际输入功能，按 Enter 执行命令
4. **键盘输入默认聚焦**：页面任意位置键盘输入自动聚焦到底部命令输入栏
5. **搜索icon交互**：点击顶部导航栏搜索icon，聚焦终端输入框并自动填入 `search ` 命令，用户直接输入关键词后回车完成搜索

**UI视觉统一性（v1.3新增）**：
1. **导航栏统一**：主页面和文章查看页面的顶部导航栏使用相同的字体（JetBrains Mono）、样式（uppercase、tracking-tight）、大小（text-sm）
2. **导航栏布局**：左侧显示路径 `~/{owner}/{currentDir}`，右侧显示 POSTS 和 ABOUTME 导航链接 + 搜索icon（右对齐）
3. **文章标题优化**：文章查看页面标题使用Space Grotesk，字体大小调整为text-4xl（优化阅读体验）
4. **Markdown渲染样式**：参考原型设计，使用Inter字体，代码块使用JetBrains Mono，配色遵循Dracula Spectrum
5. **底部终端输入栏**：placeholder透明度降低（opacity-30），避免干扰视觉焦点

**支持命令**：

| 命令 | 功能 | 示例 |
|------|------|------|
| `cd <path>` | 切换文章路径 | `cd tech/blog` |
| `cd ..` | 返回上级目录 | `cd ..` |
| `cd .` | 刷新当前目录 | `cd .` |
| `open <filename>` | 打开文章 | `open my-post.md` |
| `search <keyword>` | 搜索文章标题 | `search terminal` |
| `help` 或 `?` | 显示命令帮助模态框 | `help` 或 `?` |

**命令交互增强（v1.3新增）**：
1. **Tab键自动补全**：输入命令前缀后按Tab键自动补全完整命令（如`se`→`search `），禁用浏览器默认Tab键焦点切换行为
2. **路径补全**：Tab键支持补全文章/文件夹路径（如`open RE`→`open README.md`，`cd tec`→`cd tech/`），通过WebSocket实时从后端获取路径信息
3. **键盘输入默认聚焦**：页面任意位置键盘输入自动聚焦到底部命令输入栏
4. **搜索icon交互**：点击顶部导航栏搜索icon，聚焦终端输入框并自动填入 `search ` 命令，用户直接输入关键词后回车执行搜索

**架构约束（v1.4新增 - 架构重大变更）**：

> **变更原因**：底部终端输入框是纯前端UI组件，命令执行不应依赖后端HTTP API，仅路径补全和搜索需要后端支持（WebSocket实时通信）。

1. **前端命令处理**：
   - 底部终端输入框大部分命令不与后端通信（纯前端路由跳转）
   - `open <filename>` → 前端路由跳转到文章页面 `/article?path=filename`
   - `cd <path>` → 前端路由跳转到目录页面 `/?dir=path`
   - `help` / `?` → 触发前端模态框组件显示
   - 不需要HTTP API `/api/command` 端点

2. **WebSocket搜索命令**：
   - `search <keyword>` → WebSocket实时搜索（避免HTTP请求延迟）
   - 后端检索匹配文章标题，返回最匹配的文章路径列表
   - **过滤约束**：不搜索以 `_` 开头的隐藏文件（如 `_ABOUTME.md`）
   - 前端接收结果后直接跳转到第一个匹配结果（或显示列表）
   - WebSocket消息格式：`{"type":"search_request","keyword":"terminal"}`
   - 响应格式：`{"type":"search_response","results":[{"path":"README.md","title":"Terminalog"}]}`

3. **历史记录前端存储**：
   - 命令历史记录存储在localStorage（key: `terminalog_command_history`）
   - 支持上下键导航历史记录（ArrowUp/ArrowDown）
   - 历史记录最多保存100条
   - 历史记录仅在前端保存，不发送到后端

4. **WebSocket路径补全**：
   - Tab键路径补全通过WebSocket实时从后端获取路径信息
   - WebSocket端点：`ws://localhost:18085/ws/terminal`
   - 路径补全消息：`{"type":"completion_request","dir":"/","prefix":"RE"}`
   - 响应格式：`{"type":"completion_response","items":["README.md","tech/"]}`

**导航栏选中状态（v1.6新增，v1.6.1优化）**：
1. **POSTS选中样式**：在首页（`/` 或 `/?dir=`）时，POSTS文字颜色变为强调色（primary-container），并显示下划线
2. **ABOUTME选中样式**：在About Me页面（`/aboutme`）时，ABOUTME文字颜色变为强调色（primary-container），并显示下划线；同时POSTS文字颜色变为非强调色（outline），无下划线
3. **下划线样式**：使用CSS伪元素（after）实现独立下划线，下划线位于文字下方额外位置，**不影响span字体对齐**——POSTS和ABOUTME的span文字始终保持同一基线对齐
4. **未选中链接**：文字颜色为 outline（暗灰色），hover 时背景色变化，无下划线

**搜索结果模态框（v1.6新增）**：
1. **单结果处理**：搜索返回单个结果时，直接跳转到对应文章页面
2. **多结果显示**：返回多个结果时，弹出模态框显示搜索结果列表：
   - 左对齐显示文章标题
   - 右对齐显示文章最后更新日期
3. **交互方式**：
   - 上下方向键（ArrowUp/ArrowDown）选择结果
   - 回车键（Enter）确认跳转
   - 右上角x按钮手动关闭
   - 10秒自动关闭计时器
4. **样式复用**：复用 HelpModal 的 Glass 效果样式

**Tab键路径补全模态框（v1.6新增）**：
1. **指令补全**：Tab键自动补全指令名称（如`se`→`search `），单匹配直接填充
2. **路径补全模态框**：当`open`或`cd`命令后路径有多个匹配时：
   - 弹出模态框显示所有匹配路径
   - 左侧显示文件类型图标（📄文件/📁文件夹）
   - 文件夹路径带斜杠结尾
3. **交互方式**：
   - ArrowUp/ArrowDown选择路径
   - Enter确认并填充到输入框
   - Tab或ESC取消关闭
   - 10秒自动关闭计时器
4. **命令类型过滤**：
   - `open`命令仅显示文件（过滤文件夹）
   - `cd`命令仅显示文件夹（过滤文件）

**无匹配提示（v1.6新增）**：
1. **搜索无结果**：搜索返回空结果时，在光标上方显示1秒提示"没有搜索结果"
2. **路径补全无匹配**：Tab键补全无匹配路径时，在光标上方显示1秒提示"没有匹配内容"
3. **指令无匹配**：Tab键补全指令名称无匹配时，显示1秒提示"没有匹配内容"
4. **提示样式**：显示在输入框右上角上方，不遮挡输入内容，使用error颜色，1秒后自动消失

**帮助模态框设计**：
- **触发方式**：输入`help`或`?`命令后弹出模态框
- **内容**：展示所有可用命令及功能说明
- **关闭方式**：
  1. 3秒自动关闭（定时器）
  2. 右上角x按钮手动关闭
- **样式**：遵循Dracula Spectrum设计系统，Glass效果

**cd 命令特殊规则**：

| 输入 | 行为 |
|------|------|
| `cd ..` | 返回上级目录，自动刷新页面 |
| `cd .` | 刷新当前目录 |
| `cd <路径>` | 切换到指定路径 |
| `cd`（空） | 返回根目录 |

#### 3.3.3 文章页面布局

```
┌── Navbar ───────────────────────────────────┐
│ [Logo]  path/to/article   [Search]  [_]     │
└─────────────────────────────────────────────┘

  Tag1 • Tag2                                    ← 文件类型标签（文章/AUTHOR）
# Article Title                                  ← Space Grotesk, 渐变下划线
> Quote line from the article here.              ← Blockquote 引用块

───────────────────────────────────────────────

[文章内容 - Inter 字体渲染的 Markdown]

───────────────────────────────────────────────

EOF                                              ← EOF 分隔线
Last edited: 22 Apr 2025
v2.0.48                                          ← 版本号
▼ History (折叠/展开)
  ┌────────────────────────────────────────────┐
  │ 07 Apr 25 - v2.0.47 - 2 lines changed      │
  │ 09 Mar 25 - v2.0.46 - 15 lines changed     │
  │ 01 Mar 25 - v2.0.40 - 120 lines changed    │
  └────────────────────────────────────────────┘

> _                                               ← 底部命令输入区
```

#### 3.3.4 版本号自动生成规则

每次 Git push 时自动计算版本号，格式：`v{major}.{minor}.{patch}`，初始版本为 `v1.0.0`。

**版本号计算规则**（基于文件行数变化，优先级从高到低）：

| 变更等级 | 行数变化 | 版本号变化 | 说明 |
|---------|---------|-----------|------|
| 最高优先级 | < 10 行 | 补丁版本 +1（vX.Y.Z→vX.Y.Z+1） | 小修改（错别字、微调格式） |
| 中等变更 | 10~50% 总行数 | 子版本 +1（vX.Y.Z→X.(Y+1).0） | 增加/删除章节 |
| 重大变更 | > 50% 总行数 | 主版本 +1（vX.Y.Z→(X+1).0.0） | 几乎全部重写 |

> 优先级说明：如果同一 push 中涉及多个文件，取**最高优先级**的变更等级作为整体版本号变化。

### 3.4 搜索功能

- **搜索范围**：文章标题
- **搜索方式**：简单匹配，无需全文搜索
- **搜索结果**：显示匹配的文章列表

---

## 四、非功能需求

### 4.1 性能要求

| 指标 | 要求 |
|------|------|
| 预期文章数量 | ≤ 200 篇 |
| 预期并发量 | 100 QPS |
| 首屏加载速度 | 优先优化 |
| Markdown 渲染 | 客户端渲染 |

### 4.2 部署要求

| 配置项 | 配置方式 |
|--------|---------|
| 用户认证信息 | 配置文件（TOML） |
| Git 仓库路径 | 配置文件（TOML） |
| 服务端口 | 命令行参数 |
| 监听地址 | 命令行参数 |
| 配置文件路径 | 命令行参数 |

**命令行参数示例**：
```bash
./terminalog --port 8080 --host 0.0.0.0 --config config.toml
```

**HTTPS 部署说明**：
后端不内置 TLS 支持，推荐通过反向代理实现 HTTPS：
```nginx
# Nginx 示例配置
server {
    listen 443 ssl;
    server_name blog.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 4.3 平台支持

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

---

## 五、数据流设计

### 5.1 文章访问流程

```
用户访问 → 前端请求文章列表 → 后端扫描 Git 仓库目录
                                        ↓
                              过滤已提交的 Markdown 文件
                                        ↓
                              从 Git 历史获取元数据
                                        ↓
                              返回文章列表（JSON）
                                        ↓
                              前端渲染文章列表
```

### 5.2 文章内容获取流程

```
用户点击/命令查看 → 前端请求文章内容 → 后端读取 Markdown 文件
                                           ↓
                                   返回原始 Markdown 文本
                                           ↓
                                   前端渲染 Markdown
                                           ↓
                                   请求 Git 时间线
                                           ↓
                                   后端查询 Git log
                                           ↓
                                   返回 commit 历史
```

### 5.3 Git Push 流程

```
用户 git push → 认证（用户名/密码）→ 验证配置文件中的用户
                                           ↓
                                   执行 Git receive-pack
                                           ↓
                                   文件写入仓库目录
                                           ↓
                                   Push 完成（无需触发回调）
                                           ↓
                                   下次访问时自动获取最新内容
```

---

## 六、配置文件设计

### 6.1 配置文件示例（config.toml）

```toml
# Terminalog 配置文件

[blog]
# Blog 内容目录（也是 Git 仓库根目录）
# 支持相对路径和绝对路径
content_dir = "./content"

# Blog 属主名称（显示在导航栏路径中，如 ~/lordcasser）
# 默认值：lordcasser
owner = "lordcasser"

[auth]
# Git 用户认证列表
# 如果不存在，系统首次启动时会自动生成 admin 用户

[[auth.users]]
username = "admin"
password = "auto-generated-secure-password"

[[auth.users]]
username = "developer"
password = "custom-password-123"
```

### 6.2 配置项说明

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `blog.content_dir` | string | 是 | Blog 内容目录，也是 Git 仓库根目录 |
| `blog.owner` | string | 否 | Blog 属主名称，显示在导航栏路径中（如 `~/lordcasser`），默认值 `lordcasser` |
| `auth.users` | array | 否 | 用户认证列表，为空时自动生成 admin 用户 |
| `auth.users[].username` | string | 是 | 用户名 |
| `auth.users[].password` | string | 是 | 密码 |

---

## 七、API 设计概要

### 7.1 RESTful API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/articles` | GET | 获取文章列表，支持 `?sort=created\|edited&order=asc\|desc`（**排除** `_` 开头的文件） |
| `/api/articles/{path}` | GET | 获取文章内容 |
| `/api/articles/{path}/timeline` | GET | 获取文章 Git 时间线 |
| `/api/articles/{path}/version` | GET | 获取文章版本号及历史版本列表 |
| `/api/tree` | GET | 获取目录树结构（**排除** `_` 开头的文件） |
| `/api/search` | GET | 搜索文章标题（query: `?q=keyword`，**排除** `_` 开头的文件） |
| `/api/assets/{path}` | GET | 获取图片等静态资源（从 Git 仓库读取） |
| `/api/aboutme` | GET | 获取 About Me 页面内容（读取 `_ABOUTME.md`） |

### 7.2 Git Smart HTTP Endpoints

| 端点 | 方法 | 描述 |
|------|------|------|
| `/{repo}/info/refs?service=git-receive-pack` | GET | Git receive-pack 引用 |
| `/{repo}/git-receive-pack` | POST | Git push 接收 |
| `/{repo}/info/refs?service=git-upload-pack` | GET | Git upload-pack 引用 |
| `/{repo}/git-upload-pack` | POST | Git fetch/clone |

---

## 八、需求确认状态

### 8.1 已确认事项 ✓

| 类别 | 决策 | 确认状态 |
|------|------|---------|
| 前端框架 | Next.js + shadcn/ui | ✅ 已确认 |
| Go 版本 | 1.21+ | ✅ 已确认 |
| 存储方式 | 纯文件系统 + Git（无数据库） | ✅ 已确认 |
| Git 协议 | HTTP/HTTPS（SSH 后续迭代） | ✅ 已确认 |
| 认证方式 | 用户名密码（TOML 配置） | ✅ 已确认 |
| HTTPS 支持 | 反向代理处理，后端不内置 TLS | ✅ 已确认 |
| Markdown 扩展 | 代码高亮、数学公式、Mermaid、图片 | ✅ 已确认 |
| 图片处理 | 相对路径转换为 API 访问，外部链接保留 | ✅ 已确认 |
| 文章组织 | 目录结构映射 | ✅ 已确认 |
| 元数据来源 | Git 历史，无 Frontmatter | ✅ 已确认 |
| 文章排序 | 默认编辑时间降序，支持命令行指定 | ✅ 已确认 |
| 前端构建 | Next.js 静态导出 | ✅ 已确认 |
| 部署方式 | 单文件，embed 嵌入 | ✅ 已确认 |
| 访问权限 | Blog 公开，Git Push 需认证 | ✅ 已确认 |
| 功能范围 | 标题搜索、无评论、无后台 | ✅ 已确认 |
| 性能指标 | ≤200 篇文章，100 QPS | ✅ 已确认 |
| UI 风格 | Dracula 配色 + 终端交互 | ✅ 已确认 |
| UI 风格 | Dracula Spectrum + Brutalist（0px圆角、Glass效果） | ✅ 已确认（v1.2） |
| 字体系统 | Space Grotesk + JetBrains Mono + Inter | ✅ 已确认（v1.2） |
| 特殊文件 | `_ABOUTME.md` 作为 About Me，列表排除 `_` 开头文件 | ✅ 已确认（v1.2） |
| 版本号 | 自动计算（基于行数变化，<10行补丁/10~50%子版本/>50%主版本） | ✅ 已确认（v1.2） |
| 交互模式 | 鼠标为主，命令为辅（顶部导航+底部单行Prompt） | ✅ 已确认（v1.2） |
| 命令功能 | 移除 clear，保留完整输出历史 | ✅ 已确认（v1.2） |
| 移动端 | MVP 不支持移动端响应式 | ✅ 已确认（v1.2） |
| SSH 认证 | MVP 暂不支持，后续迭代 | ✅ 已确认 |

### 8.2 需求变更记录

| 版本 | 日期 | 变更内容 |
|------|------|---------|
| v1.0 | 2026-04-15 | 初始需求文档 |
| v1.1 | 2026-04-15 | 确认：SSH 认证暂不支持、HTTPS 由反向代理处理、图片路径处理规则、文章排序规则、Next.js 静态导出 |
| v1.2 | 2026-04-16 | **设计 v1.2 升级**：Brutalist 编辑器（Dracula Spectrum 配色+三字体+0px圆角）、版本号自动计算、About Me页面、特殊文件处理、鼠标为主命令为辅、移除clear命令、不支持移动端 |
| v1.3 | 2026-04-17 | **UI统一性 v1.3**：导航栏路径同步（currentDir状态共享）、blog.owner配置、搜索icon自动填充命令、placeholder透明度降低、HelpModal宽度调整+回车关闭 |
| v1.4 | 2026-04-17 | **架构重构 v1.4**：WebSocket终端端点（路径补全+搜索）、前端命令处理（纯前端路由跳转）、历史记录localStorage存储 |
| v1.5 | 2026-04-17 | **UI改进 v1.5**：导航栏选中状态（下划线不影响字体对齐，使用after伪元素）、MDX样式对齐原型HTML、EDITORS字体增大(10px→12px) |

---

## 九、项目边界与限制

### 9.1 不包含的功能

- ❌ 数据库存储
- ❌ Web 管理后台
- ❌ 评论系统
- ❌ 用户注册系统
- ❌ RSS/Atom 订阅
- ❌ 全文搜索
- ❌ 复杂的 Git 平台功能（PR、Issue、Wiki 等）
- ❌ SSH Git 协议（后续迭代）
- ❌ 内置 TLS/HTTPS（由反向代理处理）

### 9.2 技术限制

- 单进程单实例部署，不支持分布式
- Git 历史查询性能依赖仓库大小
- 不支持多仓库
- **MVP 不支持移动端**（不支持移动端响应式布局，桌面浏览器优先）

---

## 十、验收标准

### 10.1 功能验收

- [ ] 单文件启动成功
- [ ] Blog 页面正常显示 Brutalist 编辑器 UI（Dracula Spectrum 配色，0px 圆角，Glass 效果）
- [ ] 三字体系统正常应用（Space Grotesk 标题、JetBrains Mono UI、Inter 正文）
- [ ] 主页面文章列表为 5 列表格（Created、Updated、Editors、Filename、Latest Commit）
- [ ] 支持表头点击排序和命令行排序，共用同一套排序逻辑
- [ ] 底部命令行保留完整输出历史，不支持 clear 清屏
- [ ] 顶部导航栏正常（Logo、路径显示、搜索框）
- [ ] 底部单行命令输入框正常（Enter 执行，Tab 补全）
- [ ] 文章页面正常显示（标签行、标题+渐变下划线、引用块、EOF 分隔线、折叠式历史）
- [ ] 版本号自动计算显示正确（基于行数变化规则）
- [ ] Markdown 正确渲染（代码高亮、数学公式、Mermaid、图片）
- [ ] Git 仓库中的相对路径图片正确加载
- [ ] 外部链接图片正确加载
- [ ] Git clone 通过 HTTP 协议正常工作（无需认证）
- [ ] Git push 通过 HTTP 协议正常工作（需认证）
- [ ] 文章正确显示创建时间、创建人、编辑时间、编辑人
- [ ] 编辑时间线正确显示（commit hash、时间、作者、行数变化提示）
- [ ] 未提交文件不显示在 Blog 中
- [ ] 以 `_` 开头的文件不显示在文章列表中
- [ ] About Me 页面从 `_ABOUTME.md` 正常渲染
- [ ] 标题搜索功能正常
- [ ] 文章排序功能正常（表头点击 + 命令行参数，共用逻辑）
- [ ] 跨平台运行（Linux/macOS/Windows）

### 10.2 性能验收

- [ ] 首屏加载时间 < 2s（本地测试）
- [ ] 文章列表响应时间 < 200ms（100 篇文章）
- [ ] 单篇文章加载时间 < 100ms

---

**文档结束**

> ✅ **需求对齐完成**：所有约束条件已确认，可以进入架构设计阶段。
> 
> 下一步：进行系统架构设计，包括模块划分、接口定义、数据流设计等。