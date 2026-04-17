# Terminalog - Terminal-style Blog System

> A blog system with terminal-style UI, powered by Next.js frontend and Go backend.

## Features

- **Terminal-style UI**: Dracula color scheme with blinking cursor effect
- **Dual Interaction**: Command line (cd, ls, view, search, help, clear, exit) + mouse click
- **Git-based Content**: All articles stored in a Git repository
- **Markdown Rendering**: Code highlighting, math formulas (KaTeX), Mermaid diagrams, images
- **Article Metadata**: Creation time, edit time, contributors from Git history
- **Single Binary Deployment**: Frontend embedded in Go binary

## Quick Start

### 1. Build

```bash
make build
```

### 2. Initialize Content Directory

```bash
mkdir content
cd content
git init
echo "# My First Article" > first-article.md
git add .
git commit -m "Initial commit"
```

### 3. Run

```bash
./bin/terminalog
```

On first run, a default `config.toml` will be created. Edit it to set your content directory.

### 4. Access

Open http://localhost:8080 in your browser.

## Configuration

Edit `config.toml`:

```toml
[blog]
content_dir = "./content"

[server]
host = "0.0.0.0"
port = 8080

[auth]
users = [
  { username = "admin", password = "your-password" },
]
```

## Git Push

To push new articles to the blog:

```bash
git remote add terminalog http://localhost:8080/.git
git push terminalog master
```

You'll need to authenticate with the username/password from `config.toml`.

## Project Structure

```
terminalog/
├── cmd/terminalog/main.go      # Entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── handler/                # HTTP handlers
│   ├── model/                  # Data models
│   ├── server/                 # HTTP server
│   └── service/                # Business logic
├── pkg/
│   ├── embed/static/           # Frontend build output
│   └── utils/                  # Utility functions
├── frontend/                   # Next.js frontend (to be implemented)
├── docs/                       # Architecture documentation
├── configs/                    # Configuration examples
├── Makefile
└── README.md
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/articles` | GET | List articles |
| `/api/articles/{path}` | GET | Get article content |
| `/api/articles/{path}/timeline` | GET | Get article timeline |
| `/api/tree` | GET | Get directory tree |
| `/api/search?q=query` | GET | Search articles |
| `/api/assets/{path}` | GET | Get asset (image) |
| `/info/refs` | GET | Git refs advertisement |
| `/git-upload-pack` | POST | Git clone |
| `/git-receive-pack` | POST | Git push (auth required) |

## Development

```bash
# Run in development mode
make dev

# Build frontend (requires Next.js setup)
make frontend

# Run tests
make test
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

## License

MIT