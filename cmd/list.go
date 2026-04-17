package cmd

import (
	"fmt"

	"repoman/internal/config"
	"repoman/internal/git"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all git repos found under basePath",
	Example: "  repoman list",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		repos, err := git.ScanRepos(cfg.BasePath)
		if err != nil {
			return fmt.Errorf("scan repos: %w", err)
		}
		for _, r := range repos {
			fmt.Println(r)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
