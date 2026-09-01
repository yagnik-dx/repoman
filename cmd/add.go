package cmd

import (
	"fmt"
	"strings"

	"repoman/internal/config"
	"repoman/internal/git"
	"repoman/internal/ui"

	"github.com/spf13/cobra"
)

var (
	addBranch   string
	addStart    string
	addNoSelect bool
)

var addCmd = &cobra.Command{
	Use:   "add [repo...]",
	Short: "Add new repos to the existing config without re-running setup",
	Long: `Adds one or more repos to the existing config, keeping basePath, workspace,
and every already-configured repo untouched.

With no arguments, shows a checkbox list of repos found under basePath that are
not yet configured. Added repos become selected unless --no-select is passed.`,
	Example: `  repoman add
  repoman add newservice
  repoman add newservice --branch main --start "npm run dev"
  repoman add newservice --no-select`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		all, err := git.ScanRepos(cfg.BasePath)
		if err != nil {
			return fmt.Errorf("scanning repos: %w", err)
		}
		known := make(map[string]bool, len(all))
		for _, r := range all {
			known[r] = true
		}
		if cfg.RepoConfig == nil {
			cfg.RepoConfig = make(map[string]config.RepoConfig)
		}

		toAdd := args
		if len(toAdd) == 0 {
			var unconfigured []string
			for _, r := range all {
				if _, ok := cfg.RepoConfig[r]; !ok {
					unconfigured = append(unconfigured, r)
				}
			}
			if len(unconfigured) == 0 {
				fmt.Println("All repos under", cfg.BasePath, "are already configured.")
				return nil
			}
			toAdd, err = ui.MultiSelect("Select repos to add:", unconfigured, nil)
			if err != nil {
				return err
			}
			if len(toAdd) == 0 {
				fmt.Println("No repos selected.")
				return nil
			}
		} else {
			// Validate all names upfront so nothing is written on a typo.
			for _, r := range toAdd {
				if !known[r] {
					return fmt.Errorf("unknown repo: %s (not a git repo under %s)", r, cfg.BasePath)
				}
			}
		}

		if len(toAdd) > 1 && (addBranch != "" || addStart != "") {
			return fmt.Errorf("--branch and --start apply to a single repo; add repos one at a time or omit the flags")
		}

		var added []string
		for _, name := range toAdd {
			if existing, ok := cfg.RepoConfig[name]; ok && addBranch == "" && addStart == "" {
				ok, err := ui.Confirm(fmt.Sprintf("[%s] already configured (branch %s) — reconfigure?", name, existing.Branch))
				if err != nil {
					return err
				}
				if !ok {
					fmt.Printf("[%s] left unchanged\n", name)
					continue
				}
			}

			branch := addBranch
			if branch == "" {
				branch, err = ui.AskString(fmt.Sprintf("[%s] Target branch:", name), "develop")
				if err != nil {
					return err
				}
			}
			rawStart := addStart
			if addStart == "" && addBranch == "" {
				rawStart, err = ui.AskString(fmt.Sprintf("[%s] Start commands (comma-separated, leave blank for none):", name), "")
				if err != nil {
					return err
				}
			}

			cfg.RepoConfig[name] = config.RepoConfig{Branch: branch, Start: parseStartCommands(rawStart)}
			added = append(added, name)
			fmt.Printf("[%s] configured (branch %s)\n", name, branch)
		}

		if len(added) == 0 {
			return nil
		}

		if !addNoSelect {
			cfg.SelectedRepos = appendUnique(cfg.SelectedRepos, added...)
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("\n%d repo(s) added — %d selected total\n", len(added), len(cfg.SelectedRepos))
		return nil
	},
}

// parseStartCommands splits a comma-separated command string into trimmed,
// non-empty entries.
func parseStartCommands(raw string) []string {
	var starts []string
	for _, s := range strings.Split(raw, ",") {
		if s = strings.TrimSpace(s); s != "" {
			starts = append(starts, s)
		}
	}
	return starts
}

// appendUnique appends names to list, skipping any already present.
func appendUnique(list []string, names ...string) []string {
	seen := make(map[string]bool, len(list))
	for _, r := range list {
		seen[r] = true
	}
	for _, n := range names {
		if !seen[n] {
			list = append(list, n)
			seen[n] = true
		}
	}
	return list
}

func init() {
	addCmd.Flags().StringVar(&addBranch, "branch", "", "target branch (skips the interactive prompt)")
	addCmd.Flags().StringVar(&addStart, "start", "", "comma-separated start commands")
	addCmd.Flags().BoolVar(&addNoSelect, "no-select", false, "configure the repo but leave it inactive")
	rootCmd.AddCommand(addCmd)
}
