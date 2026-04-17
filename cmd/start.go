package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"repoman/internal/config"
	"repoman/internal/ui"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Select repos to start, open their VS Code workspace and launch start commands in new terminals",
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

		// Filter to repos that have start commands configured
		var startable []string
		for _, name := range repos {
			rc, ok := cfg.RepoConfig[name]
			if ok && len(rc.Start) > 0 {
				startable = append(startable, name)
			}
		}
		if len(startable) == 0 {
			fmt.Println("No repos with start commands configured.")
			return nil
		}

		// Let user pick which repos to start
		chosen, err := ui.MultiSelect("Select repos to start:", startable, startable)
		if err != nil {
			return err
		}
		if len(chosen) == 0 {
			fmt.Println("Nothing selected.")
			return nil
		}

		for _, name := range chosen {
			path := repoPath(cfg.BasePath, name)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("[%s] skipped: directory not found\n", name)
				continue
			}

			repoCfg := cfg.RepoConfig[name]

			// Open VS Code workspace if configured
			if repoCfg.Workspace != "" {
				fmt.Printf("[%s] opening workspace...\n", name)
				if err := exec.Command("code", repoCfg.Workspace).Start(); err != nil {
					fmt.Printf("[%s] warning: could not open VS Code: %v\n", name, err)
				}
			}

			// Spawn a new terminal window per start command
			for _, command := range repoCfg.Start {
				fmt.Printf("[%s] launching: %s\n", name, command)
				if err := spawnTerminal(path, name, command); err != nil {
					fmt.Printf("[%s] warning: could not launch terminal: %v\n", name, err)
				}
			}
		}

		return nil
	},
}

// spawnTerminal opens a new terminal window that runs command in dir.
func spawnTerminal(dir, repoName, command string) error {
	var cmd *exec.Cmd
	title := fmt.Sprintf("repoman — %s", repoName)

	if runtime.GOOS == "windows" {
		// Try Windows Terminal first, fall back to cmd
		if _, err := exec.LookPath("wt"); err == nil {
			cmd = exec.Command("wt", "--title", title,
				"--startingDirectory", dir,
				"cmd", "/K", command)
		} else {
			cmd = exec.Command("cmd", "/C", "start",
				fmt.Sprintf("repoman — %s", repoName),
				"cmd", "/K",
				fmt.Sprintf("cd /d %s && %s", dir, command))
		}
	} else {
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf(`osascript -e 'tell app "Terminal" to do script "cd %s && %s"'`, dir, command))
	}

	return cmd.Start()
}

func init() {
	rootCmd.AddCommand(startCmd)
}
