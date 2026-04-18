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
