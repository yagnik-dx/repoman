package cmd

import (
	"fmt"
	"os"

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
