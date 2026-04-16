# Repoman Design Spec
Date: 2026-04-16

## Overview

`repoman` is a minimal, safe Go CLI for managing multiple local git repositories. It automates common multi-repo workflows using a rebase-only strategy, never touching remotes destructively.

---

## Architecture

**Pattern:** Cobra + thin cmd layer (Option A)

Each `cmd/*.go` handles CLI parsing only. All logic lives in `internal/` packages. Commands call internal packages directly. No service layer.

### Project Structure

```
repoman/
  main.go
  cmd/
    root.go         # root command, --only flag, help
    setup.go
    select.go
    list.go
    selected.go
    sync.go
    start.go
    clean.go
    clean_local.go
  internal/
    config/
      config.go     # load/save ~/.repoman/config.json
    git/
      git.go        # fetch, rebase, prune, scan for repos
    ui/
      ui.go         # survey-based multiselect + prompts
    executor/
      executor.go   # run shell commands, stream output
```

### Data Flow

`cmd` layer parses flags → loads config → calls `git`/`ui`/`executor` → prints prefixed output → prints summary.

---

## Config

Location: `~/.repoman/config.json`

```json
{
  "basePath": "/Repository",
  "selectedRepos": ["backend", "frontend"],
  "repoConfig": {
    "backend": {
      "branch": "develop",
      "start": ["redis-server", "npm run dev"]
    },
    "frontend": {
      "branch": "main",
      "start": ["npm run dev"]
    }
  }
}
```

Loaded once per command invocation. Written back only by `setup` and `select`.

---

## Commands

### `repoman setup`

Interactive wizard:
1. Ask base path
2. Scan `basePath` for `.git` directories
3. Multiselect repos (arrow + space UI)
4. For each selected repo: ask target branch (default: `develop`), ask start commands (comma-separated on one line)
5. Ask which repos go into `selectedRepos`
6. Write config to `~/.repoman/config.json`

### `repoman select`

Load existing config → show multiselect of all scanned repos with current selections pre-ticked → overwrite `selectedRepos` in config.

### `repoman list`

Scan `basePath` for `.git` directories, print each repo name (directory name), one per line. No config required.

### `repoman selected`

Read `selectedRepos` from config, print each, one per line.

### `repoman sync`

For each repo (sequentially):
1. `git fetch origin --prune`
2. `git rebase origin/<branch>` (stays on current branch)
3. On rebase failure: `git rebase --abort`, mark skipped

### `repoman start`

For each repo, for each command in order (sequentially):
1. Print `[repo] > <command>`
2. Prompt: `Proceed? (y/n/skip)`
   - `y`: run command, wait for completion, continue to next
   - `n`: abort all remaining commands for this repo, move to next repo
   - `skip`: skip this command, continue to next command

### `repoman clean`

For each repo (sequentially):
```bash
git fetch origin --prune
```
Fetches latest and removes stale local remote-tracking refs. Does not touch the remote.

### `repoman clean-local`

For each repo:
1. Detect current checked-out branch
2. List all local branches except current
3. If none: log `[repo] nothing to delete`, continue
4. Show list of branches to be deleted
5. Prompt `"Delete these branches? (y/n)"`
6. On `y`: `git branch -D` each branch, log `[repo] deleted: <branch>`
7. On `n`: skip repo

### `repoman help`

Show all available commands with short descriptions and global flags.

---

## Global Flags

### `--only repo1,repo2`

- Overrides `selectedRepos` for the current command
- Validated before execution: if any name is not found in scanned repos under `basePath`, print error and exit immediately
- No repos are processed if validation fails

---

## UI & Interaction

**Interactive** (survey library):
- `setup` and `select`: arrow keys + space to toggle, Enter to confirm
- `start`: inline `y/n/skip` prompt per command
- `clean-local`: `y/n` confirmation with branch list shown

**Non-interactive output format:**
```
[backend] fetching...
[backend] rebasing onto origin/develop...
[backend] done
[frontend] skipped: uncommitted changes

5 repos processed — 4 succeeded, 1 skipped
```

No color or styling in v1. Plain text only.

---

## Execution Model

- All multi-repo commands run repos **sequentially**
- `start` runs commands within each repo sequentially
- Failure in one repo does not stop others (log and continue)

---

## Error Handling

| Scenario | Behavior |
|---|---|
| Repo directory missing | Skip, log `[repo] skipped: directory not found` |
| Target branch missing on remote | Skip, log `[repo] skipped: branch not found` |
| Rebase conflict | `git rebase --abort`, skip, log `[repo] skipped: rebase conflict` |
| Dirty working tree on sync | Skip, log `[repo] skipped: uncommitted changes` |
| `--only` with unknown repo name | Print error and exit, no repos processed |
| Config file missing | Error: `config not found, run repoman setup` |
| Start command fails | Log `[repo] command failed: <cmd>`, prompt to continue to next command or abort repo |
| `clean-local` with only one branch | Log `[repo] nothing to delete` |

---

## Safety Rules

- No `git pull`
- No merge commits
- No remote branch deletion
- No force push operations
- No automatic branch deletion (always prompt)
- Skip safely on errors, never abort the whole run

---

## Libraries

| Library | Use |
|---|---|
| `cobra` | CLI framework |
| `survey` | Interactive prompts and multiselect |
| `os/exec` | Shell command execution |
| `encoding/json` | Config read/write |

---

## Non-Goals (v1)

- No remote branch deletion
- No merge workflows
- No stash on dirty working tree
- No color output
- No database or persistent state beyond config
- No complex UI
