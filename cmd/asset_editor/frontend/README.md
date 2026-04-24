# Asset Editor Frontend

React + TypeScript + Vite frontend for the asset editor. See
`cmd/asset_editor/CLAUDE.md` for full architecture docs.

## Development

```bash
npm install
npm run dev     # starts Vite on :5173, proxies /api to Go server on :8090
```

The Go backend must be running separately with `-dev` flag.

## Build

```bash
npm run build   # outputs to dist/, embedded by Go binary
```

## Stack

- React 19, TypeScript, Vite
- Tailwind CSS v4 (`@tailwindcss/vite` plugin)
- shadcn/ui components (backed by `@base-ui/react`)
- TanStack Query for server state
- React Router for client routing
