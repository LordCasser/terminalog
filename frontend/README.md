## Terminalog Frontend

Next.js App Router frontend for Terminalog. The app is statically exported and then embedded into the Go backend binary.

## Development

Start the frontend dev server:

```bash
npm run dev
```

When running against a separate backend in debug mode, set:

```env
NEXT_PUBLIC_API_BASE=http://localhost:18085
```

## Build

```bash
npm run build
```

The root `Makefile` copies the exported frontend output into `pkg/embed/static`.

## Rendering Stack

- `react-markdown`
- `remark-gfm`
- `rehype-highlight`
- `remark-math`
- `rehype-katex`

## Key Paths

- `app/`: routes and layouts
- `components/`: UI and terminal interaction
- `lib/api/`: backend API clients
- `types/`: shared frontend response types

## Notes

- Search uses REST: `GET /api/v1/search`
- Path completion uses WebSocket: `/ws/terminal`
- Markdown content is fetched from the backend and rendered client-side
