# SniffOps Web UI Backend

HTTP API server for SniffOps Web UI Dashboard.

## Architecture

```
internal/web/
├── server.go   - HTTP server initialization, routing, CORS
├── api.go      - REST API handlers (/api/traces, /api/stats, etc.)
├── embed.go    - Go embed.FS for serving frontend assets
└── dist/       - Embedded frontend build output (copied from web/dist)
```

## API Endpoints

| Method | Path | Description | Query Parameters |
|--------|------|-------------|------------------|
| `GET` | `/api/traces` | List traces with filtering | `?tool=...&namespace=...&risk=...&limit=50&offset=0&start=unix_ms&end=unix_ms` |
| `GET` | `/api/traces/:id` | Get trace by ID | - |
| `GET` | `/api/stats` | Aggregated statistics | `?period=24h` |
| `GET` | `/api/namespaces` | List distinct namespaces | - |
| `GET` | `/api/tools` | List distinct tools | - |
| `GET` | `/` | Serve embedded frontend | - |

## Usage

```bash
# Build backend + frontend
make build-all

# Run web server (default port: 3000)
./bin/sniffops web

# Run on custom port
./bin/sniffops web --port 8080
```

## Development

```bash
# Backend only (uses placeholder frontend)
make build-backend

# Frontend only (React/Vite)
make build-web

# Run backend with hot-reload (requires air or similar)
make web
```

## CORS

Development CORS is enabled for:
- `http://localhost:5173` (Vite dev server)
- `http://127.0.0.1:5173`

Production builds are served directly from the embedded filesystem.

## Implementation Notes

1. **Lightweight**: Uses standard `net/http` (no Gin/Echo)
2. **Embedded FS**: Frontend is bundled into the Go binary
3. **SQLite**: Shares the same trace database as MCP server
4. **Graceful shutdown**: Handles SIGINT/SIGTERM properly

## Build Requirements

- Go 1.23+ (embed.FS, generics)
- Node.js 18+ (for frontend build)
- npm or yarn

## Frontend Build Process

The frontend (React + Vite) is built separately and copied into `internal/web/dist/`:

```bash
cd web/
npm install
npm run build         # -> web/dist/
cp -r dist ../internal/web/dist/  # -> embedded by embed.go
```

This is automated by `make build-all`.
