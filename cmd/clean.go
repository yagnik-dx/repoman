package cmd

import (
	"fmt"
	"os"

	"repoman/internal/config"
	"repoman/internal/git"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Fetch and prune stale remote-tracking refs for each repo",
	Long:  "Runs git fetch origin --prune for each repo, removing stale local remote-tracking refs. Does not modify the remote.",
	Example: `  repoman clean
  repoman clean --only backend`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		repos, err := resolveRepos(cfg.BasePath, cfg.SelectedRepos)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var results []repoResult
		for _, name := range repos {
			path := repoPath(cfg.BasePath, name)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("[%s] skipped: directory not found\n", name)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}
			fmt.Printf("[%s] fetching and pruning...\n", name)
			if err := git.Fetch(path); err != nil {
				fmt.Printf("[%s] skipped: fetch failed: %v\n", name, err)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}
			fmt.Printf("[%s] done\n", name)
			results = append(results, repoResult{repo: name, success: true})
		}

		printSummary(results)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
