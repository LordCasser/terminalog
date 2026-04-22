# Server 模块

提供 HTTP/HTTPS 服务器功能，基于 `go-chi/chi/v5` 路由和标准库 `net/http`。

## 功能特性

- **HTTP 服务**：标准 HTTP 服务（默认）
- **HTTPS/TLS 支持**：可选 TLS 加密传输
- **自动重定向**：启用 TLS 时，HTTP(80) 自动重定向到 HTTPS
- **优雅关闭**：支持 graceful shutdown
- **中间件栈**：请求日志、Gzip 压缩、CORS、超时控制等

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

### HTTPS 模式

```go
tlsConfig := server.TLSConfig{
    Enabled:  true,
    CertFile: "/path/to/cert.pem",
    KeyFile:  "/path/to/key.pem",
}

srv := server.NewServer(
    "0.0.0.0:8443",
    handlers,
    logger,
    embedFS,
    debug,
    tlsConfig,
)

// 启动 HTTP 重定向服务器（端口 80）
go srv.StartRedirect()

if err := srv.Start(); err != nil {
    log.Fatal(err)
}
```

## 配置说明

### TOML 配置示例

```toml
[server]
host = "0.0.0.0"
port = 8443
debug = false

# TLS 配置（可选，默认关闭）
tls_enabled = true
cert_file = "/etc/terminalog/cert.pem"
key_file = "/etc/terminalog/key.pem"
```

### 自签名证书生成

```bash
# 生成 365 天有效期的自签名证书
openssl req -x509 -newkey rsa:4096 \
    -keyout key.pem -out cert.pem \
    -days 365 -nodes \
    -subj "/CN=localhost"
```

## API 设计

### Server 结构体

```go
type Server struct {
    addr     string           // 监听地址
    router   *chi.Mux         // 路由
    server   *http.Server     // HTTP 服务器
    logger   *slog.Logger     // 日志
    Handlers *Handlers        // 处理器集合
    debug    bool             // 调试模式
    tls      TLSConfig        // TLS 配置
    redirectServer *http.Server // HTTP 重定向服务器
}
```

### TLSConfig 结构体

```go
type TLSConfig struct {
    Enabled  bool   // 是否启用 TLS
    CertFile string // 证书文件路径
    KeyFile  string // 私钥文件路径
}
```

### 核心方法

| 方法 | 说明 |
|------|------|
| `NewServer(addr, handlers, logger, embedFS, debug, tls)` | 创建服务器实例 |
| `Start() error` | 启动服务器（HTTP 或 HTTPS） |
| `StartRedirect() error` | 启动 HTTP→HTTPS 重定向服务器 |
| `Stop(ctx) error` | 优雅关闭服务器 |

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

## 注意事项

- TLS 启用时，HTTP(80) 自动重定向到 HTTPS，无需额外配置
- 自签名证书会导致浏览器安全警告，内部使用可接受
- WebSocket 在 TLS 下使用 `wss://` 协议
- 重定向服务器使用短超时（5s 读/写，10s 空闲）防止资源占用

## 变更日志

- **v1.0** - 基础 HTTP 服务器
- **v1.6** - 新增 HTTPS/TLS 支持，自动 HTTP→HTTPS 重定向
