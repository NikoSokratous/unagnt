# Unagnt Web UI

Modern web interface for monitoring and managing agent runs.

## Features

- Real-time run monitoring with SSE streaming
- Interactive run timeline visualization
- Run status dashboard with metrics
- Dark mode UI optimized for terminal users

## Development

```bash
cd web
npm install
npm run dev
```

The dev server runs on `http://localhost:3000` and proxies API requests to `unagntd` on port 8080.

## Build

```bash
npm run build
```

Builds to `dist/` directory, ready for embedding in `unagntd`.

## Tech Stack

- React 18
- TypeScript
- Vite
- TanStack Query (React Query)
- React Router
- Recharts (for future visualizations)
- Lucide Icons
