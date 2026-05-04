# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o repoman .

# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/config/...
go test ./internal/executor/...
go test ./internal/git/...

# Run a single test by name
go test ./internal/config/... -run TestSaveAndLoad

# Lint (standard Go tooling)
go vet ./...
```

## Architecture

`repoman` is a multi-repo workflow manager using a **rebase-only, never-destructive-to-remotes** strategy.

**Pattern:** Cobra CLI with a thin `cmd/` layer — each `cmd/*.go` handles only flag parsing and output formatting. All logic lives in `internal/` packages. There is no service layer between them.

### Data flow

```
cmd layer (parse flags) → load config → call internal/* → print [repo]-prefixed output → printSummary()
```

### Internal packages

- **`internal/config`** — loads/saves `~/.repoman/config.json` (JSON, marshaled with `encoding/json`). Only `setup` and `select` commands write config; all others are read-only.
- **`internal/git`** — thin wrappers around `git` CLI: `ScanRepos` finds subdirs with a `.git` dir, `Fetch`/`Rebase`/`RebaseAbort`/`IsDirty`/`BranchExistsOnRemote`/`LocalBranches`/`DeleteBranch` cover all git operations needed.
- **`internal/executor`** — `RunStreaming` runs a shell command string via `sh -c` (or `cmd /C` on Windows), prefixing every output line with `[repoName]`.
- **`internal/ui`** — `survey`-based interactive prompts (multiselect, text input) used by `setup` and `select`.

### Config schema (`~/.repoman/config.json`)

```json
{
  "basePath": "/path/to/repos",
  "selectedRepos": ["backend", "frontend"],
  "repoConfig": {
    "backend": { "branch": "develop", "start": ["redis-server", "npm run dev"] },
    "frontend": { "branch": "main", "start": ["npm run dev"] }
  }
}
```

### Global `--only` flag

Defined in `cmd/root.go`, consumed via `resolveRepos()`. Overrides `selectedRepos` for any command. Validated upfront against `ScanRepos` — if any name is unknown, the command exits before processing any repo.

### Multi-repo execution model

All commands process repos **sequentially**. Failure in one repo never stops the others — log and continue. Each command ends with `printSummary()` showing total/succeeded/skipped counts.

### Adding a new command

1. Create `cmd/<name>.go`, define a `cobra.Command`, wire it in `init()` with `rootCmd.AddCommand(...)`.
2. Call `config.Load()` → `resolveRepos()` → iterate repos using `repoPath(cfg.BasePath, name)`.
3. Use `git.*` for git operations and `executor.RunStreaming` for arbitrary shell commands.
4. Print progress as `[repoName] <message>` and call `printSummary(results)` at the end.
