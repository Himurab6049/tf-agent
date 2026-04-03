# tf-agent — Claude Code Guide

## Build & Run

```bash
make build          # builds client + server binary → .bin/tf-agent-server
make run            # build + start server on :8080
go build ./...      # Go only
cd client && npm run build   # client only
```

Kill / restart the server:
```bash
lsof -ti :8080 | xargs kill -9
.bin/tf-agent-server &
```

## Project layout

```
tf-agent/
├── cmd/server/            entry point (main.go, embed.go)
├── internal/
│   ├── agent/             core LLM loop (loop.go, prompt.go, compact.go)
│   ├── commands/          slash-command registry + handlers
│   ├── config/            YAML config loader
│   ├── db/                SQLite store interface (store.go) + impl (sqlite.go)
│   ├── hooks/             pre/post-tool hook runner
│   ├── llm/               provider abstraction (Anthropic, Bedrock, mock)
│   ├── permissions/       tool allow/deny policy checker
│   ├── queue/             in-memory task queue
│   ├── server/            HTTP handlers, SSE hub, task runner, crypto, metrics
│   ├── session/           conversation history (file + memory backends)
│   ├── skills/            RepoScan, Clarifier, Generate, Validate, SecurityScan, CreatePR, JiraFetch, DriftDetect
│   ├── taskctx/           context helpers (credentials, ask_user callback)
│   └── tools/             Read, Write, Edit, Glob, Grep, Ls, Bash, Agent, AskUser, Task, WebFetch, WebSearch
├── client/
│   └── src/
│       ├── components/    React components
│       ├── hooks/         useTaskRunner, …
│       ├── lib/api.ts     typed API client
│       └── types.ts       shared TypeScript types
├── Makefile
├── Dockerfile
└── go.mod                 module: github.com/tf-agent/tf-agent
```

## Key conventions

- **Go 1.22+ routing** — method+path patterns on root mux, e.g. `PATCH /v1/admin/users/{id}`. Use `r.PathValue("id")` — never nest admin routes in a sub-mux or path values won't populate.
- **Token format** — `tfa-` + 40 hex chars from `crypto/rand`. Only SHA-256 hash stored; raw token shown once.
- **Sensitive fields** — GitHub / Atlassian tokens stored AES-256-GCM encrypted. Decrypt in `task_runner.go` before use.
- **Import cycle prevention** — `internal/tools` must NOT import `internal/agent`. Use injected function types (e.g. `SubAgentRunner`) wired in `server/task_runner.go`.
- **Task output** — accumulated from `TurnEventText` events in `task_runner.go`, stored in `tasks.output` column.

## Database

SQLite file at `~/.tf-agent/tf-agent.db` by default. Schema in `internal/db/schema.sql`.

When adding a column to an existing table use `ALTER TABLE ADD COLUMN` in a migration — do not re-create tables.

## Auth

`Authorization: Bearer <token>` header. Admin endpoints additionally checked by `adminMiddleware`. Self-action guards prevent admins from deleting/deactivating their own account.

## Frontend hash routing

Top-level views use URL hash (`#form`, `#history`, `#settings`). Settings section persisted in `sessionStorage` key `tf_settings_section`. Both survive browser refresh.

## Running tests

```bash
go test ./...
cd client && npm test
```
