# Terminalog - Debug Mode 开发调试指南

> 文档版本：v1.0
> 创建日期：2026-04-17
> 适用版本：v1.2+

---

## 一、Debug Mode 简介

Debug Mode 是为前后端分离调试设计的开发模式，允许前端和后端独立启动和调试。

### 1.1 核心特性

| 特性 | Debug Mode ON | Debug Mode OFF |
|------|--------------|----------------|
| 前端嵌入 | ❌ 不嵌入前端静态文件 | ✅ 前端静态文件嵌入二进制 |
| CORS 支持 | ✅ 启用 CORS，允许跨域 | ❌ 不启用 CORS |
| 前端运行方式 | 独立 dev server (Next.js) | 从后端服务静态文件 |
| 适用场景 | 开发调试、前后端联调 | 生产部署、单文件启动 |

---

## 二、启用 Debug Mode

### 2.1 命令行方式（推荐）

```bash
# 启动后端（debug mode）
./terminalog --debug --port 18085 --config config.toml

# 或使用 go run
go run cmd/terminalog/main.go --debug --port 18085 --config config.toml
```

**参数说明**：

- `--debug`：启用 debug 模式
- `--port`：后端端口（默认 8080，建议使用 18085 避免与前端冲突）
- `--config`：配置文件路径

### 2.2 配置文件方式

编辑 `config.toml`：

```toml
[server]
host = '0.0.0.0'
port = 18085
debug = true  # 启用 debug 模式
```

**注意**：命令行 `--debug` 参数优先级高于配置文件。

---

## 三、前后端分离调试流程

### 3.1 启动后端

```bash
# 方式 1：编译后运行
make backend
./bin/terminalog --debug --port 18085

# 方式 2：直接运行
go run cmd/terminalog/main.go --debug --port 18085
```

**后端日志示例**：

```
INFO  Terminalog starting version=dev buildDate=unknown
INFO  Server started addr=0.0.0.0:18085
INFO  Enabling debug mode from command line
INFO  Git service initialized repoPath=/tmp/articles
INFO  CORS enabled for frontend dev server
INFO  Access the API at http://0.0.0.0:18085/api
```

### 3.2 启动前端

```bash
# 进入前端目录
cd frontend

# 安装依赖（首次）
pnpm install

# 启动 Next.js dev server
pnpm dev
```

**前端日志示例**：

```
  ▲ Next.js 16.2.4
  - Local:        http://localhost:3000
  - Network:      http://192.168.1.100:3000

 ✓ Ready in 2.3s
```

### 3.3 前端配置 API 基础 URL

前端默认同源访问 API（无 basePath）。在 debug mode 下，前端运行在 `localhost:3000`，后端运行在 `localhost:18085`，需要配置 API 基础 URL。

**方式 1：环境变量**（推荐）

创建 `frontend/.env.local`：

```env
NEXT_PUBLIC_API_BASE=http://localhost:18085
```

**方式 2：修改 API Client**

编辑 `frontend/lib/api/client.ts`：

```typescript
const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:18085';
```

### 3.4 验证联调

访问 `http://localhost:3000`，前端页面应该能正常加载，并且能访问后端 API：

- ✅ 文章列表正常显示
- ✅ API 请求跨域成功（浏览器 Network 面板无 CORS 错误）
- ✅ Git 操作正常

---

## 四、Debug Mode 架构说明

### 4.1 架构对比

#### Production Mode（默认）

```
┌─────────────────────────────────────────┐
│     单一可执行文件（terminalog）          │
│  ┌────────────────────────────────────┐ │
│  │      前端静态资源（embed）            │ │
│  │  - Next.js 静态导出产物              │ │
│  │  - CSS/JS/图片资源                  │ │
│  └────────────────────────────────────┘ │
│  ┌────────────────────────────────────┐ │
│  │      后端 API 服务                   │ │
│  │  - REST API                        │ │
│  │  - Git Smart HTTP                  │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### Debug Mode

```
┌──────────────────┐              ┌──────────────────┐
│  前端 Dev Server   │              │  后端 API Server │
│  (Next.js)        │              │  (Go)            │
│  localhost:3000   │──── CORS ───▶│  localhost:18085 │
│  - React 组件      │              │  - REST API      │
│  - 热更新          │              │  - Git HTTP      │
│  - 开发调试        │              │  - CORS Enabled  │
└──────────────────┘              └──────────────────┘
```

### 4.2 CORS 配置

Debug Mode 启用 CORS，允许前端跨域访问后端 API：

```go
// CORS Headers
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS, PUT, DELETE
Access-Control-Allow-Headers: Accept, Content-Type, Content-Length, Accept-Encoding, Authorization
```

**注意**：生产环境不建议启用 CORS（前端已嵌入后端，无需跨域）。

---

## 五、常见问题

### 5.1 前端无法访问 API

**问题**：浏览器 Network 面板显示 CORS 错误。

**解决**：

1. 确认后端已启用 `--debug` 参数
2. 确认前端 `.env.local` 配置了正确的 API URL
3. 检查后端日志是否显示 "CORS enabled"

### 5.2 前端页面空白

**问题**：前端 dev server 正常启动，但页面空白。

**解决**：

1. 检查浏览器 Console 是否有错误
2. 确认后端 API 正常运行（访问 `http://localhost:18085/api/articles`）
3. 检查前端 API Client 配置

### 5.3 如何切换回 Production Mode

**方式 1：关闭 --debug**

```bash
./terminalog --port 18085  # 不加 --debug
```

**方式 2：修改配置文件**

```toml
[server]
debug = false
```

---

## 六、开发调试最佳实践

### 6.1 推荐调试流程

1. **启动后端**：`go run cmd/terminalog/main.go --debug --port 18085`
2. **启动前端**：`cd frontend && pnpm dev`
3. **浏览器调试**：访问 `http://localhost:3000`
4. **修改代码**：前端热更新，后端需重启

### 6.2 调试技巧

| 场景 | 建议 |
|------|------|
| 前端修改 UI | 直接修改，自动热更新 |
| 前端修改 API 调用 | 修改后刷新页面 |
| 后端修改逻辑 | 重启后端服务 |
| 后端修改路由 | 重启后端服务 |
| Git 操作测试 | 使用 git clone/push 命令测试 |

### 6.3 日志调试

**后端日志级别调整**：

```bash
# 更详细的日志
./terminalog --debug --log debug

# 或修改 config.toml
[server]
debug = true
log_level = "debug"
```

---

## 七、生产部署注意事项

### 7.1 切换回 Production Mode

生产部署时，必须关闭 debug mode：

```bash
# 编译前端
cd frontend
pnpm build

# 复制到 embed 目录
cp -r out/* ../pkg/embed/static/

# 编译后端（不启用 debug）
cd ..
make build

# 启动
./bin/terminalog --port 8080
```

### 7.2 安全提示

⚠️ **Debug Mode 仅用于开发调试，不适合生产环境**：

- CORS 允许所有来源（`Access-Control-Allow-Origin: *`）
- 前端未嵌入，依赖独立 dev server
- 缺少生产级安全配置

---

## 八、相关文档

| 文档 | 说明 |
|------|------|
| [architecture.md](./architecture.md) | 系统架构总览 |
| [frontend-architecture.md](./frontend-architecture.md) | 前端架构设计 |
| [backend-architecture.md](./backend-architecture.md) | 后端架构设计 |

---

**文档结束**

> ✅ Debug Mode 改造完成，支持前后端分离调试。