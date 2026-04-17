package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"repoman/internal/git"

	"github.com/spf13/cobra"
)

var onlyFlag []string

var rootCmd = &cobra.Command{
	Use:   "repoman",
	Short: "Multi-repo workflow manager",
	Long:  "repoman manages multiple local git repositories using a rebase-only strategy.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&onlyFlag, "only", nil, "run only on specified repos (comma-separated)")
}

type repoResult struct {
	repo    string
	success bool
	message string
}

func printSummary(results []repoResult) {
	total := len(results)
	success := 0
	for _, r := range results {
		if r.success {
			success++
		}
	}
	fmt.Printf("\n%d repos processed — %d succeeded, %d skipped\n", total, success, total-success)
}

// resolveRepos returns the list of repos to operate on.
// If --only is set, validates each name exists under basePath and returns them.
// Otherwise returns selectedRepos from config.
func resolveRepos(basePath string, selectedRepos []string) ([]string, error) {
	if len(onlyFlag) == 0 {
		return selectedRepos, nil
	}
	all, err := git.ScanRepos(basePath)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(all))
	for _, r := range all {
		set[r] = true
	}
	for _, r := range onlyFlag {
		if !set[r] {
			return nil, fmt.Errorf("unknown repo: %s", r)
		}
	}
	return onlyFlag, nil
}

// repoPath joins basePath and repoName into a full path.
func repoPath(basePath, name string) string {
	return filepath.Join(basePath, name)
}
