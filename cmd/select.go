package cmd

import (
	"fmt"

	"repoman/internal/config"
	"repoman/internal/git"
	"repoman/internal/ui"

	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:     "select",
	Short:   "Interactively choose which repos are active",
	Example: "  repoman select",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		all, err := git.ScanRepos(cfg.BasePath)
		if err != nil {
			return fmt.Errorf("scanning repos: %w", err)
		}
		selected, err := ui.MultiSelect("Select active repos:", all, cfg.SelectedRepos)
		if err != nil {
			return err
		}
		cfg.SelectedRepos = selected
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Println("Selected repos updated.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(selectCmd)
}
