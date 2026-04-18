# Frontend Implementation Notes

## Overview

The frontend is a statically exported Next.js App Router application. It renders blog content by calling the Go backend APIs at runtime.

## Active Rendering Path

- Markdown rendering uses `react-markdown`
- GFM uses `remark-gfm`
- Syntax highlighting uses `rehype-highlight`
- Math rendering uses `remark-math` + `rehype-katex`
- GFM tables are rendered with a custom brutalist data-panel style:
  horizontal scroll wrapper, heavy borders, Dracula-aligned row contrast, and no top table label
- Search uses REST: `GET /api/v1/search`
- Path completion uses WebSocket: `/ws/terminal`

## Routing

- `/` renders the root article list
- `/dir/{path}` renders a directory listing
- `/article/{path}` renders an article
- `/aboutme` renders `_ABOUTME.md`

Static export fallback pages are used for `/dir/[...slug]` and `/article/[...slug]`.

## API Usage

- `GET /api/v1/articles`
- `GET /api/v1/articles/{path}`
- `GET /api/v1/articles/{path}/timeline`
- `GET /api/v1/articles/{path}/version`
- `GET /api/v1/search`
- `GET /api/v1/tree`
- `GET /api/v1/special/aboutme`
- `GET /api/v1/settings`

## Command Prompt

- Supported commands: `search`, `open`, `cd`, `help`, `?`
- Search is resolved through REST results and modal selection
- `open` and `cd` are client-side navigations with API validation where needed
- Tab completion uses WebSocket-backed path completion
- Command history is stored in `localStorage` under `terminalog_command_history`

## Current Constraints

- Mermaid is not implemented
- Fonts are loaded via `<link>` tags in `layout.tsx`
- Markdown images use native `<img>` rather than `next/image`
