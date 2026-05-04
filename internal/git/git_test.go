package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
)

func TestScanRepos(t *testing.T) {
	base := t.TempDir()
	// repos with .git
	os.MkdirAll(filepath.Join(base, "backend", ".git"), 0755)
	os.MkdirAll(filepath.Join(base, "frontend", ".git"), 0755)
	// not a repo
	os.MkdirAll(filepath.Join(base, "docs"), 0755)
	// file, not dir
	os.WriteFile(filepath.Join(base, "README.md"), []byte("hi"), 0644)

	repos, err := ScanRepos(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sort.Strings(repos)
	if len(repos) != 2 {
		t.Fatalf("got %d repos, want 2: %v", len(repos), repos)
	}
	if repos[0] != "backend" || repos[1] != "frontend" {
		t.Errorf("got %v, want [backend frontend]", repos)
	}
}

func TestScanReposMissingBase(t *testing.T) {
	_, err := ScanRepos(filepath.Join(t.TempDir(), "nonexistent"))
	if err == nil {
		t.Fatal("expected error for missing basePath")
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "t@t.com"},
		{"git", "config", "user.name", "T"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("init repo: %v", err)
		}
	}
	// initial commit
	f := filepath.Join(dir, "README.md")
	os.WriteFile(f, []byte("# repo"), 0644)
	exec.Command("git", "-C", dir, "add", ".").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	return dir
}

func TestIsDirty(t *testing.T) {
	dir := initRepo(t)

	dirty, err := IsDirty(dir)
	if err != nil {
		t.Fatalf("IsDirty clean: %v", err)
	}
	if dirty {
		t.Error("expected clean repo to not be dirty")
	}

	// modify a file
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed"), 0644)
	dirty, err = IsDirty(dir)
	if err != nil {
		t.Fatalf("IsDirty dirty: %v", err)
	}
	if !dirty {
		t.Error("expected modified repo to be dirty")
	}
}

func TestCurrentBranch(t *testing.T) {
	dir := initRepo(t)
	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch == "" {
		t.Error("expected non-empty branch name")
	}
}

func TestLocalBranches(t *testing.T) {
	dir := initRepo(t)

	// create two extra branches
	exec.Command("git", "-C", dir, "branch", "feature-a").Run()
	exec.Command("git", "-C", dir, "branch", "feature-b").Run()

	branches, err := LocalBranches(dir)
	if err != nil {
		t.Fatalf("LocalBranches: %v", err)
	}
	// current branch is excluded, so we expect feature-a and feature-b
	if len(branches) != 2 {
		t.Fatalf("got %d branches, want 2: %v", len(branches), branches)
	}
}

func TestDeleteBranch(t *testing.T) {
	dir := initRepo(t)
	exec.Command("git", "-C", dir, "branch", "to-delete").Run()

	if err := DeleteBranch(dir, "to-delete"); err != nil {
		t.Fatalf("DeleteBranch: %v", err)
	}

	branches, _ := LocalBranches(dir)
	for _, b := range branches {
		if b == "to-delete" {
			t.Error("branch still exists after deletion")
		}
	}
}
