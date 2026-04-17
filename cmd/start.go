package cmd

import (
	"fmt"
	"os"

	"repoman/internal/config"
	"repoman/internal/executor"
	"repoman/internal/ui"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Run startup commands for each repo, one at a time with proceed/skip prompts",
	Long: `For each repo, for each configured start command:
  - Prompt: proceed / skip / abort
  - proceed: run the command and wait for it to finish
  - skip: skip this command, continue to next
  - abort: skip remaining commands for this repo, move to next repo`,
	Example: `  repoman start
  repoman start --only backend`,
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

			repoCfg, ok := cfg.RepoConfig[name]
			if !ok || len(repoCfg.Start) == 0 {
				fmt.Printf("[%s] no start commands configured\n", name)
				results = append(results, repoResult{repo: name, success: true})
				continue
			}

			repoSuccess := true
		commandLoop:
			for _, command := range repoCfg.Start {
				fmt.Printf("[%s] > %s\n", name, command)
				choice, err := ui.Select("Proceed?", []string{"proceed", "skip", "abort"})
				if err != nil {
					return err
				}
				switch choice {
				case "proceed":
					if err := executor.RunStreaming(path, name, command); err != nil {
						fmt.Printf("[%s] command failed: %s\n", name, command)
						repoSuccess = false
						cont, _ := ui.Confirm("Continue with remaining commands?")
						if !cont {
							break commandLoop
						}
					}
				case "skip":
					fmt.Printf("[%s] skipped: %s\n", name, command)
				case "abort":
					fmt.Printf("[%s] aborted\n", name)
					repoSuccess = false
					break commandLoop
				}
			}
			results = append(results, repoResult{repo: name, success: repoSuccess})
		}

		printSummary(results)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
