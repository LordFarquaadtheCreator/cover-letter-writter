# cover-letter-writter

MCP server for generating styled PDF cover letters with profile management and history.

## How it works

This is a stdio-based MCP server written in Go. It exposes seven tools:

- `create_profile` — stores contact info (name, address, email, phone) on disk, returns profile with ID
- `get_profile` — fetches a single profile by ID
- `list_profiles` — lists all profiles sorted by creation time
- `update_profile` — applies non-empty field updates to a profile
- `delete_profile` — removes a profile (history preserved)
- `generate_cover_letter` — builds a styled PDF using a profile + body text, saves to disk, records history
- `list_history` — lists past generations, optionally filtered by profile ID

`profiles.json` and `history.json` are stored next to the executable and loaded on every tool call — not compiled in. The path is resolved via `os.Executable()`, so the binary works regardless of the agent's working directory.

## Architecture

No external API dependencies. The server is self-contained: profile/history stores are JSON files on disk, PDF generation is in-process via `go-pdf/fpdf`.

```
Agent ──stdio──► MCP Server
                   create/get/list/update/delete_profile  ──► profiles.json
                   generate_cover_letter                   ──► PDF + history.json
                   list_history                            ──► history.json
```

## Tests

```bash
go test ./... -count=1
```

Tests across 4 packages:

| Package | Tests cover |
|---|---|
| `internal/profile` | Create validation, get/list/update/delete, not-found cases, disk persistence across instances, malformed/empty JSON |
| `internal/history` | Add validation, list with filter, newest-first ordering, disk persistence, malformed JSON |
| `internal/generate` | Empty body, PDF written, default filename, nested output dir creation, override precedence, smart-quote sanitization |
| `internal/mcpserver` | All 7 handler integration paths, history recording on generate, profile CRUD round-trips, error propagation |

## Build

```bash
go build -o cover-letter-writter .
```

## Run

No env vars required.

```bash
./cover-letter-writter
```

## MCP Config

Copy `mcp-config.json` into the agent's MCP config.

## Key files

| File | Purpose |
|---|---|
| `main.go` | Entry point, resolves data dir via `os.Executable()` |
| `internal/mcpserver/server.go` | MCP server setup, tool handlers, JSON result helper |
| `internal/profile/profile.go` | Profile struct, disk-backed Store (CRUD) |
| `internal/history/history.go` | History Entry struct, disk-backed Store |
| `internal/generate/generate.go` | PDF generation: green sidebar, two-column layout, sanitize, output dir resolution |
| `mcp-config.json` | MCP config snippet |
| `internal/profile/profile_test.go` | Profile store tests |
| `internal/history/history_test.go` | History store tests |
| `internal/generate/generate_test.go` | PDF generation tests |
| `internal/mcpserver/server_test.go` | Handler integration tests |
