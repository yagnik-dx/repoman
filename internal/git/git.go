package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ScanRepos returns names of subdirectories under basePath that contain a .git directory.
func ScanRepos(basePath string) ([]string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}
	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(basePath, e.Name(), ".git")); err == nil {
			repos = append(repos, e.Name())
		}
	}
	return repos, nil
}

func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}

func output(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// Fetch runs git fetch origin --prune in repoPath.
func Fetch(repoPath string) error {
	return run(repoPath, "git", "fetch", "origin", "--prune")
}

// Rebase runs git rebase origin/<branch> in repoPath.
func Rebase(repoPath, branch string) error {
	return run(repoPath, "git", "rebase", "origin/"+branch)
}

// RebaseAbort runs git rebase --abort in repoPath.
func RebaseAbort(repoPath string) error {
	return run(repoPath, "git", "rebase", "--abort")
}

// IsDirty returns true if the working tree has uncommitted changes.
func IsDirty(repoPath string) (bool, error) {
	out, err := output(repoPath, "git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(out) > 0, nil
}

// BranchExistsOnRemote returns true if <branch> exists on origin.
func BranchExistsOnRemote(repoPath, branch string) (bool, error) {
	out, err := output(repoPath, "git", "ls-remote", "--heads", "origin", branch)
	if err != nil {
		return false, err
	}
	return len(out) > 0, nil
}
