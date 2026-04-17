package cmd

import (
	"fmt"
	"os"

	"repoman/internal/config"
	"repoman/internal/git"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Fetch and rebase each repo onto its target branch",
	Long: `For each repo:
  1. git fetch origin --prune
  2. git rebase origin/<branch>
  On conflict: git rebase --abort, skip repo.
  On dirty working tree: skip repo.`,
	Example: `  repoman sync
  repoman sync --only backend,frontend`,
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

			dirty, err := git.IsDirty(path)
			if err != nil {
				fmt.Printf("[%s] skipped: could not check status: %v\n", name, err)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}
			if dirty {
				fmt.Printf("[%s] skipped: uncommitted changes\n", name)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}

			repoCfg, ok := cfg.RepoConfig[name]
			if !ok {
				fmt.Printf("[%s] skipped: no config found\n", name)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}

			fmt.Printf("[%s] fetching...\n", name)
			if err := git.Fetch(path); err != nil {
				fmt.Printf("[%s] skipped: fetch failed: %v\n", name, err)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}

			exists, err := git.BranchExistsOnRemote(path, repoCfg.Branch)
			if err != nil {
				fmt.Printf("[%s] skipped: could not check remote branch: %v\n", name, err)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}
			if !exists {
				fmt.Printf("[%s] skipped: branch %q not found on remote\n", name, repoCfg.Branch)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}

			fmt.Printf("[%s] rebasing onto origin/%s...\n", name, repoCfg.Branch)
			if err := git.Rebase(path, repoCfg.Branch); err != nil {
				if abortErr := git.RebaseAbort(path); abortErr != nil {
					fmt.Printf("[%s] warning: rebase --abort failed: %v\n", name, abortErr)
				}
				fmt.Printf("[%s] skipped: rebase conflict\n", name)
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
	rootCmd.AddCommand(syncCmd)
}
