package config

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		BasePath:      "/repos",
		SelectedRepos: []string{"backend", "frontend"},
		RepoConfig: map[string]RepoConfig{
			"backend":  {Branch: "develop", Start: []string{"redis-server", "npm run dev"}},
			"frontend": {Branch: "main", Start: []string{"npm run dev"}},
		},
	}

	if err := saveToPath(cfg, path); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := loadFromPath(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.BasePath != cfg.BasePath {
		t.Errorf("BasePath: got %q, want %q", loaded.BasePath, cfg.BasePath)
	}
	if len(loaded.SelectedRepos) != 2 {
		t.Errorf("SelectedRepos len: got %d, want 2", len(loaded.SelectedRepos))
	}
	if loaded.RepoConfig["backend"].Branch != "develop" {
		t.Errorf("backend branch: got %q, want develop", loaded.RepoConfig["backend"].Branch)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := loadFromPath(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "config not found") {
		t.Errorf("expected 'config not found' error, got: %v", err)
	}
}

func TestConfigJSON(t *testing.T) {
	cfg := &Config{
		BasePath:      "/repos",
		SelectedRepos: []string{"a"},
		RepoConfig:    map[string]RepoConfig{"a": {Branch: "main", Start: []string{"cmd"}}},
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}
