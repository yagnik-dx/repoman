package cmd

import (
	"fmt"
	"os"
	"strings"

	"repoman/internal/config"
	"repoman/internal/git"
	"repoman/internal/ui"

	"github.com/spf13/cobra"
)

var cleanLocalCmd = &cobra.Command{
	Use:   "clean-local",
	Short: "Delete all local branches except the currently checked-out one (with confirmation)",
	Long: `For each repo:
  1. Lists all local branches except the current one
  2. Shows the list and prompts for confirmation
  3. On confirmation, force-deletes each branch with git branch -D`,
	Example: `  repoman clean-local
  repoman clean-local --only backend`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		repos, err := resolveRepos(cfg.BasePath, cfg.SelectedRepos)
		if err != nil {
			return err
		}

		var results []repoResult
		for _, name := range repos {
			path := repoPath(cfg.BasePath, name)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("[%s] skipped: directory not found\n", name)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}

			branches, err := git.LocalBranches(path)
			if err != nil {
				fmt.Printf("[%s] skipped: could not list branches: %v\n", name, err)
				results = append(results, repoResult{repo: name, success: false})
				continue
			}
			if len(branches) == 0 {
				fmt.Printf("[%s] nothing to delete\n", name)
				results = append(results, repoResult{repo: name, success: true})
				continue
			}

			current, _ := git.CurrentBranch(path)
			fmt.Printf("[%s] current branch: %s\n", name, current)
			fmt.Printf("[%s] branches to delete:\n", name)
			for _, b := range branches {
				fmt.Printf("  - %s\n", b)
			}

			ok, err := ui.Confirm(fmt.Sprintf("Delete these %d branches in %s?", len(branches), name))
			if err != nil {
				return err
			}
			if !ok {
				fmt.Printf("[%s] skipped\n", name)
				results = append(results, repoResult{repo: name, success: true})
				continue
			}

			deleted := []string{}
			failed := []string{}
			for _, b := range branches {
				if err := git.DeleteBranch(path, b); err != nil {
					fmt.Printf("[%s] failed to delete %s: %v\n", name, b, err)
					failed = append(failed, b)
				} else {
					fmt.Printf("[%s] deleted: %s\n", name, b)
					deleted = append(deleted, b)
				}
			}
			if len(failed) > 0 {
				fmt.Printf("[%s] deleted %d, failed %d (%s)\n", name, len(deleted), len(failed), strings.Join(failed, ", "))
				results = append(results, repoResult{repo: name, success: false})
			} else {
				results = append(results, repoResult{repo: name, success: true})
			}
		}

		printSummary(results)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanLocalCmd)
}
