# CommandPrompt 组件实现文档

## 概述

CommandPrompt 是终端风格的命令输入组件，固定在页面底部，提供命令行交互体验。

## 功能特性

### 1. 终端输入功能
- 真实的输入框（不只是静态闪烁光标）
- 路径显示：`guest@blog:~/lordcasser$`
- 命令输入和执行

### 2. 键盘聚焦机制
- **全局键盘监听**：按下任意键自动聚焦输入框
- **忽略特殊键**：导航键（Tab, Arrow）、功能键（F1-F12）、修饰键（Meta, Ctrl, Alt）
- **智能聚焦**：仅在未聚焦状态下响应键盘事件

### 3. 搜索icon交互
- 导航栏搜索icon点击触发 `FOCUS_COMMAND_INPUT` 自定义事件
- CommandPrompt监听事件并聚焦输入框

### 4. 核心命令解析
- `search <query>`: 搜索功能（跳转到搜索页面）
- `open <article>`: 打开指定文章
- `cd <path>`: 跳转到指定目录

## 技术实现

### 自定义事件机制
```typescript
// 导出事件名
export const FOCUS_COMMAND_INPUT = "focusCommandInput";

// Navbar触发事件
window.dispatchEvent(new Event(FOCUS_COMMAND_INPUT));

// CommandPrompt监听事件
window.addEventListener(FOCUS_COMMAND_INPUT, handleFocusEvent);
```

### 全局键盘监听
```typescript
useEffect(() => {
  const handleKeyDown = (e: KeyboardEvent) => {
    if (isFocused || e.metaKey || e.ctrlKey || e.altKey) return;
    
    const ignoreKeys = ["Tab", "Escape", "F1", ..., "ArrowUp", ...];
    if (ignoreKeys.includes(e.key)) return;
    
    inputRef.current?.focus();
  };

  window.addEventListener("keydown", handleKeyDown);
  return () => window.removeEventListener("keydown", handleKeyDown);
}, [isFocused]);
```

### 命令解析
```typescript
const executeCommand = useCallback((cmd: string) => {
  const trimmedCmd = cmd.trim().toLowerCase();
  
  if (trimmedCmd.startsWith("search ")) {
    const query = cmd.trim().slice(7);
    router.push(`/?search=${encodeURIComponent(query)}`);
    return;
  }
  
  if (trimmedCmd.startsWith("open ")) {
    const article = cmd.trim().slice(5);
    router.push(`/article?path=${encodeURIComponent(article)}`);
    return;
  }
  
  if (trimmedCmd.startsWith("cd ")) {
    const path = cmd.trim().slice(3);
    router.push(`/?dir=${encodeURIComponent(path)}`);
    return;
  }
  
  // ... 其他命令处理
}, [router]);
```

## 样式约束

- **字体**: JetBrains Mono (font-mono)
- **大小**: text-sm
- **颜色**: 
  - 路径显示：`text-tertiary` + `text-secondary`
  - 输入框：`text-on-surface`
- **位置**: fixed bottom-0
- **光标**: 闪烁光标动画（仅在聚焦且空输入时显示）

## 架构集成

- **layout.tsx**: 集成到 RootLayout 底部，替换静态 footer
- **Navbar.tsx**: 搜索icon点击触发聚焦事件
- **全局键盘**: 整个页面响应键盘输入自动聚焦

## 未来扩展

- 命令历史记录（上下键导航）
- 自动补全功能
- Tab补全路径和文章名