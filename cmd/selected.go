package cmd

import (
	"fmt"

	"repoman/internal/config"

	"github.com/spf13/cobra"
)

var selectedCmd = &cobra.Command{
	Use:     "selected",
	Short:   "List currently selected repos",
	Example: "  repoman selected",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		for _, r := range cfg.SelectedRepos {
			fmt.Println(r)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(selectedCmd)
}
