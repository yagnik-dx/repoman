package git

import (
	"os"
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
