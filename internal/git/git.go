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
