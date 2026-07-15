# cover-letter-writter

MCP server that generates styled PDF cover letters. Agents manage user profiles (CRUD), draft the body text, and generate cover letters that reference a profile. History of every generation is tracked on disk.

## Architecture

```
Agent (MCP client over stdio)
  │
  ▼
cover-letter-writter (Go binary)
  ├── create_profile / get / list / update / delete  ── reads+writes profiles.json
  ├── generate_cover_letter                          ── builds PDF, writes to disk, records history
  └── list_history                                   ── reads history.json
```

The server is stdio-based. `profiles.json` and `history.json` are stored next to the executable and loaded on every tool call — update them without rebuilding.

## MCP Tools

### `create_profile`

Create a user profile storing contact info used to generate cover letters. Returns the profile with its ID.

| Param | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Full name as it appears on the cover letter |
| `email` | string | yes | Email shown in the contact column |
| `address` | string | no | Address shown in the contact column |
| `phone` | string | no | Phone shown in the contact column |
| `label` | string | no | Short label (e.g. "default", "tech") |
| `outputDir` | string | no | Default PDF output directory. Defaults to ~/Downloads. |
| `filename` | string | no | Default PDF filename. Defaults to `<Name>NoSpacesCoverLetter.pdf`. |

### `get_profile`

Fetch a single profile by ID.

| Param | Type | Required | Description |
|---|---|---|---|
| `id` | string | yes | Profile ID |

### `list_profiles`

List all stored profiles sorted by creation time. No params.

### `update_profile`

Update fields on an existing profile. Only non-empty fields are applied.

| Param | Type | Required | Description |
|---|---|---|---|
| `id` | string | yes | Profile ID to update |
| `name` / `email` / `address` / `phone` / `label` / `outputDir` / `filename` | string | no | Fields to update |

### `delete_profile`

Delete a profile by ID. History entries are preserved.

| Param | Type | Required | Description |
|---|---|---|---|
| `id` | string | yes | Profile ID to delete |

### `generate_cover_letter`

Generate a styled PDF cover letter. Requires a profile and body text. Saves to the profile's `outputDir` (or override), defaulting to `~/Downloads`. Records the generation in history.

| Param | Type | Required | Description |
|---|---|---|---|
| `profileId` | string | yes | Profile ID (from `create_profile` or `list_profiles`) |
| `body` | string | yes | Cover letter body text — the agent drafts this |
| `to` | string | no | Recipient salutation line (e.g. `Dear Hiring Manager,`). Defaults to `To Whom it May Concern,` |
| `outputDir` | string | no | Override profile's outputDir |
| `filename` | string | no | Override profile's filename |

Returns the absolute file path of the saved PDF.

### `list_history`

List past cover letter generations, newest first.

| Param | Type | Required | Description |
|---|---|---|---|
| `profileId` | string | no | Filter by profile ID |

## Build

```bash
go build -o cover-letter-writter .
```

## Docker

```bash
docker build -t cover-letter-writter .
```

## Run

```bash
./cover-letter-writter
```

No env vars required. The server starts immediately over stdio.

## MCP Config

Copy `mcp-config.json` into your agent's MCP config. For Docker, mount a volume so `profiles.json` and `history.json` persist across runs.

## Files

| File | Purpose |
|---|---|
| `main.go` | Entry point, resolves data dir next to executable |
| `internal/mcpserver/server.go` | MCP server setup, tool handlers |
| `internal/profile/profile.go` | Profile struct, disk-backed store |
| `internal/history/history.go` | History struct, disk-backed store |
| `internal/generate/generate.go` | PDF generation logic, sanitize, layout |
| `Dockerfile` | Multi-stage Go build → Alpine runtime |
| `mcp-config.json` | MCP config snippet |
