# repoman

A minimal Go CLI for managing multiple local git repositories. Uses a rebase-only strategy — never touches remotes destructively.

## Install

```bash
go install github.com/yagnik-dx/repoman@latest
```

Or build from source:

```bash
git clone https://github.com/yagnik-dx/repoman.git
cd repoman
go install .
```

## Setup

Run the interactive wizard once to configure your repos:

```bash
repoman setup
```

It will ask for your base directory (the folder containing all your repos), let you pick which repos to manage, and configure a target branch and start commands for each.

Config is saved to `~/.repoman/config.json`.

## Commands

| Command | Description |
|---|---|
| `repoman setup` | Interactive wizard — configure repos, branches, and start commands |
| `repoman select` | Change which repos are active |
| `repoman list` | List all git repos found under `basePath` |
| `repoman selected` | List currently active repos |
| `repoman sync` | `git fetch --prune` + `git rebase origin/<branch>` for each repo |
| `repoman clean` | `git fetch --prune` for each repo (no rebase) |
| `repoman clean-local` | Delete all local branches except the current one (with confirmation) |
| `repoman start` | Run configured start commands with proceed/skip/abort prompts |

### Global flag

```bash
--only repo1,repo2   # Override selected repos for a single command
```

## How sync works

For each repo, `repoman sync`:
1. Checks for uncommitted changes — skips if dirty
2. `git fetch origin --prune`
3. Verifies the target branch exists on remote
4. `git rebase origin/<branch>` — on conflict, runs `git rebase --abort` and skips

No merge commits, no force pushes, no remote branch deletion.

## Config

`~/.repoman/config.json`:

```json
{
  "basePath": "/path/to/your/repos",
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
