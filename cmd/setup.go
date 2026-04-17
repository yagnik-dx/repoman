package cmd

import (
	"fmt"
	"strings"

	"repoman/internal/config"
	"repoman/internal/git"
	"repoman/internal/ui"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:     "setup",
	Short:   "Interactive setup wizard — configure repos, branches, and start commands",
	Example: "  repoman setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		basePath, err := ui.AskString("Base path (directory containing your repos):", "/Repository")
		if err != nil {
			return err
		}

		all, err := git.ScanRepos(basePath)
		if err != nil {
			return fmt.Errorf("scanning repos: %w", err)
		}
		if len(all) == 0 {
			fmt.Println("No git repos found under", basePath)
			return nil
		}

		toSetup, err := ui.MultiSelect("Select repos to configure:", all, nil)
		if err != nil {
			return err
		}
		if len(toSetup) == 0 {
			fmt.Println("No repos selected.")
			return nil
		}

		// Load existing config to preserve entries for repos not being re-configured
		existing, _ := config.Load()
		repoConfigs := make(map[string]config.RepoConfig)
		if existing != nil {
			for k, v := range existing.RepoConfig {
				repoConfigs[k] = v
			}
		}
		for _, r := range toSetup {
			branch, err := ui.AskString(fmt.Sprintf("[%s] Target branch:", r), "develop")
			if err != nil {
				return err
			}
			raw, err := ui.AskString(fmt.Sprintf("[%s] Start commands (comma-separated, leave blank for none):", r), "")
			if err != nil {
				return err
			}
			var starts []string
			for _, s := range strings.Split(raw, ",") {
				s = strings.TrimSpace(s)
				if s != "" {
					starts = append(starts, s)
				}
			}
			repoConfigs[r] = config.RepoConfig{Branch: branch, Start: starts}
		}

		selected, err := ui.MultiSelect("Which repos should be active (selected)?", toSetup, toSetup)
		if err != nil {
			return err
		}

		cfg := &config.Config{
			BasePath:      basePath,
			SelectedRepos: selected,
			RepoConfig:    repoConfigs,
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Println("Config saved.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
