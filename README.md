# Terminalog

Terminalog 是一个用 Go 和 Next.js 构建的终端风格博客系统。内容直接来自 Git 仓库，前端静态资源最终嵌入后端二进制，生产环境只需要运行一个程序。

当前实现特征：
- 文章与目录浏览走 REST API
- 搜索只保留 REST：`GET /api/v1/search`
- WebSocket 只负责命令行路径补全：`/ws/terminal`
- 文章元数据来自 Git 历史，不依赖 frontmatter
- 生产部署形态为单二进制 + 一个内容仓库

## 功能概览

- 终端式博客 UI，支持鼠标点击和命令输入两种交互
- 支持 `cd`、`open`、`search`、`help`、`?`
- Markdown 渲染支持 GFM、GitHub Alerts、代码高亮、KaTeX
- 文章创建时间、更新时间、贡献者、时间线来自 Git 提交历史
- `_ABOUTME.md` 作为 About Me 页面内容
- `.assets/` 可存放文章引用图片等资源
- Git Smart HTTP 支持 clone / push，push 需要 Basic Auth

## 运行要求

- Go 1.25+
- Node.js 22+
- npm 11+
- Git

## 内容仓库约定

Terminalog 读取一个 Git 仓库作为博客内容目录。

规则：
- 只有已经提交到 Git 的 Markdown 文件会被展示
- 以 `_` 开头的 Markdown 文件不会进入文章列表
- `_ABOUTME.md` 会映射为 About Me 页面
- `.assets/` 目录不会进入文章列表，但可被文章正文引用
- 文章元数据完全从 Git 历史推导

一个最小示例：

```text
content/
├── .git/
├── _ABOUTME.md
├── hello-world.md
├── guides/
│   ├── first-post.md
│   └── .assets/
│       └── cover.png
└── tech/
    └── golang.md
```

## 快速开始

### 1. 构建程序

```bash
make build
```

构建完成后会得到：

```text
bin/terminalog
```

### 2. 准备博客内容仓库

```bash
mkdir -p /srv/terminalog/content
cd /srv/terminalog/content
git init

cat > hello-world.md <<'EOF'
# Hello World

This is my first Terminalog post.
EOF

cat > _ABOUTME.md <<'EOF'
# About Me

Terminalog blog owner.
EOF

git add .
git commit -m "Initial blog content"
```

### 3. 准备配置文件

可以直接复制示例：

```bash
cp configs/config.toml.example /srv/terminalog/config.toml
```

然后修改为你的实际路径和站点信息：

```toml
[blog]
content_dir = "/srv/terminalog/content"
owner = "alice"

[server]
host = "0.0.0.0"
port = 8080
debug = false

[auth]
users = [
  { username = "writer", password = "change-this-password" },
]
```

说明：
- `blog.content_dir` 必须指向一个 Git 仓库
- `blog.owner` 会显示在前端导航路径中
- `auth.users` 用于 Git push 认证
- 如果 `auth.users` 为空，首次启动会自动生成默认用户并把随机密码写到配置里

### 4. 启动服务

```bash
./bin/terminalog --config /srv/terminalog/config.toml
```

默认访问地址：

```text
http://127.0.0.1:8080
```

### 5. 验证服务

```bash
curl http://127.0.0.1:8080/api/v1/healthz
curl http://127.0.0.1:8080/api/v1/articles
curl "http://127.0.0.1:8080/api/v1/search?q=hello"
```

## 完整部署指南

### 部署方式

推荐的生产部署结构：

```text
/srv/terminalog/
├── bin/terminalog
├── config.toml
└── content/
    └── .git/
```

其中：
- `bin/terminalog` 是发布程序
- `config.toml` 是运行配置
- `content/` 是你的博客内容 Git 仓库

### 步骤 1：构建发布产物

在项目根目录执行：

```bash
make build
```

如果你要手工分步构建：

```bash
cd frontend
npm install
npm run build
cd ..
env GOCACHE=/tmp/terminalog-gocache make backend
```

### 步骤 2：部署到服务器

把以下文件复制到服务器：

```text
bin/terminalog
config.toml
content/
```

如果服务器上还没有内容仓库，也可以只先上传 `terminalog` 和 `config.toml`，再在服务器上初始化 `content/`。

### 步骤 3：以 systemd 运行

新建 `/etc/systemd/system/terminalog.service`：

```ini
[Unit]
Description=Terminalog Blog
After=network.target

[Service]
Type=simple
WorkingDirectory=/srv/terminalog
ExecStart=/srv/terminalog/bin/terminalog --config /srv/terminalog/config.toml
Restart=always
RestartSec=3
User=www-data
Group=www-data

[Install]
WantedBy=multi-user.target
```

然后启用：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now terminalog
sudo systemctl status terminalog
```

### 步骤 4：接入反向代理

Terminalog 自己提供 HTTP 服务。生产环境建议放在 Nginx 或 Caddy 后面做 HTTPS。

Nginx 示例：

```nginx
server {
    listen 80;
    server_name blog.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

如果你使用 HTTPS，建议由反向代理终止 TLS。

### 步骤 5：配置 Git 推送工作流

服务启动后，你可以把 Terminalog 当作内容仓库远端。

添加远端：

```bash
git remote add blog http://blog.example.com/api/v1/git/
```

推送内容：

```bash
git add .
git commit -m "Publish new post"
git push blog HEAD
```

说明：
- clone / fetch 为公开访问
- push 需要 `config.toml` 中配置的用户名和密码
- 页面内容在 push 后会按后端缓存策略自动更新，无需重新构建前端

### 步骤 6：更新博客内容

推荐工作流：

1. 在本地 clone 你的内容仓库。
2. 新建或修改 Markdown 文件。
3. 把图片放到文章同级或子级的 `.assets/` 目录。
4. `git add`、`git commit`、`git push blog HEAD`。
5. 打开站点确认展示结果。

## 前端页面与交互

页面路由：
- `/`：根目录文章列表
- `/dir/{path}`：目录页
- `/article/{path}`：文章详情页
- `/aboutme`：About Me 页面

命令行支持：
- `cd <path>`
- `cd ..`
- `cd .`
- `open <file>`
- `search <keyword>`
- `help`
- `?`

当前不支持：
- `ls`
- `view`
- `clear`
- `exit`

## API 概览

主要接口：
- `GET /api/v1/articles`
- `GET /api/v1/articles/{path}`
- `GET /api/v1/articles/{path}/timeline`
- `GET /api/v1/articles/{path}/version`
- `GET /api/v1/search`
- `GET /api/v1/tree`
- `GET /api/v1/assets/{path}`
- `GET /api/v1/special/aboutme`
- `GET /api/v1/settings`
- `GET /api/v1/healthz`
- `GET /api/v1/readyz`
- `GET /api/v1/livez`
- `GET /api/v1/status`
- `GET /ws/terminal`
- `GET /info/refs`
- `POST /git-upload-pack`
- `POST /git-receive-pack`
- `GET /api/v1/git/info/refs`
- `POST /api/v1/git/git-upload-pack`
- `POST /api/v1/git/git-receive-pack`

详细说明见 [docs/specs/api-spec.md](docs/specs/api-spec.md)。

## 本地开发

### 后端调试模式

```bash
go run cmd/terminalog/main.go --debug --port 18085 --config ./config.toml
```

或：

```bash
./bin/terminalog --debug --port 18085 --config ./config.toml
```

### 前端开发服务器

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

`frontend/.env.local` 示例：

```env
NEXT_PUBLIC_API_BASE=http://localhost:18085
```

### 常用命令

```bash
make build
make backend
make frontend
make test
```

## 验证状态

当前仓库已经验证通过：
- `npm run lint`
- `npm run build -- --webpack`
- `env GOCACHE=/tmp/terminalog-gocache go test ./internal/service ./internal/server ./internal/handler ./internal/config ./internal/model`
- `env GOCACHE=/tmp/terminalog-gocache make build`

## 文档

- [docs/specs/requirements.md](docs/specs/requirements.md) — 项目需求规格
- [docs/specs/api-spec.md](docs/specs/api-spec.md) — API 接口规格
- [docs/design/architecture.md](docs/design/architecture.md) — 全局架构概览
- [docs/design/frontend-design.md](docs/design/frontend-design.md) — 前端架构设计
- [docs/design/backend-design.md](docs/design/backend-design.md) — 后端架构设计
- [docs/guides/markdown-theme.md](docs/guides/markdown-theme.md) — Markdown 主题系统
- [docs/guides/testing.md](docs/guides/testing.md) — 测试指南
- [docs/guides/debug.md](docs/guides/debug.md) — 调试指南

完整文档索引见 [docs/INDEX.md](docs/INDEX.md)。
