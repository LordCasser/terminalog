# Server 模块

提供 HTTP/HTTPS 服务器功能，基于 `go-chi/chi/v5` 路由和标准库 `net/http`。

## 功能特性

- **HTTP 服务**：标准 HTTP 服务（默认）
- **HTTPS/TLS 支持**：可选 TLS 加密传输
- **证书自动检测**：启用 TLS 时自动搜索默认路径的证书文件
- **AutoCert 开发模式**：自动生成自签名证书，便于本地开发
- **自动重定向**：TLS 在标准端口 443 时，自动启用 HTTP(:80)→HTTPS 重定向
- **307 重定向**：使用 Temporary Redirect（参考 multifile），保留 HTTP 方法，不被浏览器缓存
- **HSTS 安全头**：TLS 启用时自动添加 Strict-Transport-Security 响应头
- **优雅关闭**：支持 graceful shutdown
- **中间件栈**：请求日志、Gzip 压缩、CORS、超时控制、HSTS 等

## 快速开始

### HTTP 模式（默认）

```go
srv := server.NewServer(
    "0.0.0.0:8080",
    handlers,
    logger,
    embedFS,
    debug,
    server.TLSConfig{}, // TLS 禁用
)

if err := srv.Start(); err != nil {
    log.Fatal(err)
}
```

### HTTPS 模式（手动指定证书）

```go
tlsConfig := server.TLSConfig{
    Enabled:  true,
    CertFile: "/path/to/cert.pem",
    KeyFile:  "/path/to/key.pem",
    HSTS:     true, // 启用 HSTS 安全头
}

srv := server.NewServer(
    "0.0.0.0:443",
    handlers,
    logger,
    embedFS,
    debug,
    tlsConfig,
)

// 启动 HTTP→HTTPS 重定向（307 Temporary Redirect）
go srv.StartRedirect()

if err := srv.Start(); err != nil {
    log.Fatal(err)
}
```

### HTTPS 模式（AutoCert 开发自签名证书）

```go
tlsConfig := server.TLSConfig{
    Enabled:  true,
    AutoCert: true, // 证书不存在时自动生成自签名证书
    HSTS:     true,
}

srv := server.NewServer(...)
```

## 配置说明

### TOML 配置示例

```toml
[server]
host = "0.0.0.0"

# 端口：0=自动选择（HTTP→8080, HTTPS→443），或显式指定
port = 0

debug = false

# TLS 配置
tls_enabled = true

# 证书路径（不设置时自动检测默认路径）
# 自动检测路径（按顺序）：resources/https.crt, resources/cert.pem, cert.pem
# cert_file = "/etc/terminalog/cert.pem"

# 私钥路径（不设置时自动检测默认路径）
# 自动检测路径（按顺序）：resources/https.key, resources/key.pem, key.pem
# key_file = "/etc/terminalog/key.pem"

# HTTP→HTTPS 重定向地址
# TLS+443 默认 ":80"，非标准端口默认不启动
# 设为 "-" 禁用重定向
# http_redirect_addr = ":80"

# 自动生成自签名证书（仅开发用！）
# auto_cert = false
```

### 证书自动检测路径

启用 `tls_enabled = true` 但未设置 `cert_file`/`key_file` 时，系统按以下顺序搜索：

| 类型 | 搜索路径（按优先级） |
|------|---------------------|
| 证书 | `resources/https.crt` → `resources/cert.pem` → `cert.pem` |
| 私钥 | `resources/https.key` → `resources/key.pem` → `key.pem` |

### 自签名证书生成

**方式一：AutoCert（推荐用于开发）**

在配置文件中设置 `auto_cert = true`，系统会自动生成自签名证书到 `resources/` 目录。

**方式二：手动 openssl**

```bash
# 生成 365 天有效期的自签名证书
openssl req -x509 -newkey rsa:4096 \
    -keyout resources/https.key -out resources/https.crt \
    -days 365 -nodes \
    -subj "/CN=localhost"
```

**方式三：Go 代码调用**

```go
import "terminalog/pkg/utils"

err := utils.GenerateSelfSignedCert("resources/https.crt", "resources/https.key", "localhost")
```

## API 设计

### TLSConfig 结构体

```go
type TLSConfig struct {
    Enabled          bool   // 是否启用 TLS
    CertFile         string // 证书文件路径（支持自动检测）
    KeyFile          string // 私钥文件路径（支持自动检测）
    HTTPRedirectAddr string // HTTP→HTTPS 重定向地址（默认 ":80"，"-" 禁用）
    HSTS             bool   // 启用 Strict-Transport-Security 响应头
    AutoCert         bool   // 证书是否为自动生成的自签名证书（开发用）
}
```

### 核心方法

| 方法 | 说明 |
|------|------|
| `NewServer(addr, handlers, logger, embedFS, debug, tls)` | 创建服务器实例 |
| `Start() error` | 启动服务器（HTTP 或 HTTPS） |
| `StartRedirect() error` | 启动 HTTP→HTTPS 重定向服务器（307） |
| `Stop(ctx) error` | 优雅关闭服务器 |

### Config 层 TLS 解析方法

| 方法 | 说明 |
|------|------|
| `ResolveDefaultPort()` | 智能默认端口：TLS→443, HTTP→8080 |
| `ResolveTLSSettings()` | 自动检测证书/私钥路径 |
| `ResolveHTTPRedirectAddr()` | 根据端口自动确定重定向地址 |

## 路由结构

```
/api/v1
├── /healthz, /readyz, /livez, /status  # 健康检查
├── /articles/*                         # 文章 API
├── /search                             # 搜索 API
├── /tree                               # 目录树 API
├── /assets/*                           # 静态资源
├── /special/aboutme                    # 关于页面
├── /settings                           # 前端配置
└── /git/*                              # Git Smart HTTP

/ws/terminal                            # WebSocket 终端
/info/refs, /git-upload-pack, /git-receive-pack  # Git HTTP
/*                                      # 前端静态文件
```

## 中间件栈

1. **RequestID** - 请求追踪 ID
2. **RealIP** - 真实 IP 获取
3. **loggingMiddleware** - 请求日志
4. **Gzip** - 响应压缩
5. **Recoverer** - 异常恢复
6. **Timeout** - 请求超时（60s）
7. **corsMiddleware** - 跨域支持（仅 debug 模式）
8. **hstsMiddleware** - Strict-Transport-Security（仅 TLS+HSTS 模式）

## TLS 安全注意事项

- **307 重定向**：使用 `Temporary Redirect` 而非 `Moved Permanently`，保留 HTTP 方法，不被浏览器缓存（参考 multifile 项目）
- **HSTS 头**：`max-age=31536000; includeSubDomains`，告知浏览器始终使用 HTTPS
- **AutoCert**：自签名证书仅用于开发，浏览器会显示安全警告
- **私钥保护**：自动生成的私钥文件权限为 0600
- **WebSocket**：TLS 下使用 `wss://` 协议

## 变更日志

- **v1.0** - 基础 HTTP 服务器
- **v1.6** - 新增 HTTPS/TLS 支持，HTTP→HTTPS 重定向（301）
- **v1.7** - TLS 服务改进（参考 multifile 项目）
  - 证书自动检测：支持默认路径搜索（resources/https.crt 等）
  - 智能默认端口：TLS→443, HTTP→8080
  - 重定向改为 307 Temporary Redirect（保留方法，不缓存）
  - 新增 HSTS 安全头（Strict-Transport-Security）
  - 新增 AutoCert 模式：开发时自动生成自签名证书
  - 非标准端口默认不启动重定向（需显式配置）
