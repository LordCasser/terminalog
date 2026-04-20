# HelpModal 组件实现文档

## 概述

HelpModal 是命令帮助模态框组件，当用户输入`help`或`?`命令时弹出，展示所有可用命令及功能说明。

## 功能特性

### 1. 触发方式
- 输入`help`命令触发模态框
- 输入`?`命令触发模态框（功能相同）
- 使用自定义事件`SHOW_HELP_MODAL`触发

### 2. 关闭方式
- **10秒自动关闭**：模态框弹出后启动10秒定时器，自动关闭
- **手动关闭**：右上角x按钮点击关闭（清除定时器）
- **Backdrop点击关闭**：点击背景区域关闭（清除定时器）

### 3. 内容展示
- 展示所有可用命令：cd, open, search, help
- 每个命令附带功能说明
- 命令使用JetBrains Mono字体，描述使用text-on-surface-variant颜色

## 技术实现

### 自定义事件机制
```typescript
// 导出事件名
export const SHOW_HELP_MODAL = "showHelpModal";

// CommandPrompt触发事件
window.dispatchEvent(new Event(SHOW_HELP_MODAL));

// HelpModal监听事件
window.addEventListener(SHOW_HELP_MODAL, handleShowModal);
```

### 自动关闭定时器
```typescript
useEffect(() => {
  const handleShowModal = () => {
    setIsVisible(true);
    
    // 10秒自动关闭
    const autoCloseTimer = setTimeout(() => {
      setIsVisible(false);
      setTimer(null);
    }, 10000);
    
    setTimer(autoCloseTimer);
  };
  
  window.addEventListener(SHOW_HELP_MODAL, handleShowModal);
  return () => window.removeEventListener(SHOW_HELP_MODAL, handleShowModal);
}, []);
```

### 手动关闭清除定时器
```typescript
const handleClose = () => {
  if (timer) {
    clearTimeout(timer);  // 清除自动关闭定时器
    setTimer(null);
  }
  setIsVisible(false);
};
```

## 样式约束

- **Glass效果**：bg-surface-container-high/42 + backdrop-blur-lg + border-primary/25
- **字体**：JetBrains Mono (font-mono)
- **大小**：text-sm
- **颜色**：
  - 标题：text-secondary
  - 命令：text-tertiary font-bold
  - 描述：text-on-surface-variant
- **圆角**：0px (Brutalist风格)
- **阴影**：shadow-2xl

## 架构集成

- **layout.tsx**: 集成到 RootLayout 作为全局模态框组件
- **CommandPrompt.tsx**: 输入`help`或`?`命令触发`SHOW_HELP_MODAL`事件
- **事件系统**: 使用window自定义事件实现跨组件通信

## 依赖关系

- 独立组件，无外部依赖
- 通过事件系统与CommandPrompt通信
- 集成到layout.tsx全局布局

---

# CommandPrompt Implementation

## Scope

`CommandPrompt` is the fixed bottom terminal input used across all pages.

## Supported Commands

- `search <keyword>`
- `open <path>`
- `cd <path>`
- `help`
- `?`

## Behavior

- Search uses REST via `GET /api/v1/search`
- Path completion uses WebSocket via `/ws/terminal`
- `open` routes to `/article/{path}`
- `cd` routes to `/dir/{path}` or `/`
- Help opens the modal through `SHOW_HELP_MODAL`

## Interaction Model

- Global keypress focuses the input
- Tab is reserved for command/path completion
- Arrow up/down navigate local command history
- Command history is stored in `localStorage`

## Notes

- Path and navigation helpers are extracted to `components/command/utils.ts`
- Search result selection and path completion selection are coordinated through window events
