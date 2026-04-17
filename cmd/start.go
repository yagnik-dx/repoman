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
	Short: "Select repos to start, open VS Code workspace and launch each start command in a new terminal",
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

		chosen, err := ui.MultiSelect("Select repos to start:", startable, startable)
		if err != nil {
			return err
		}
		if len(chosen) == 0 {
			fmt.Println("Nothing selected.")
			return nil
		}

		if cfg.Workspace != "" {
			fmt.Println("Opening workspace...")
			if err := exec.Command("code", cfg.Workspace).Start(); err != nil {
				fmt.Printf("warning: could not open VS Code: %v\n", err)
			}
		}

		for _, name := range chosen {
			path := repoPath(cfg.BasePath, name)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("[%s] skipped: directory not found\n", name)
				continue
			}
			for _, command := range cfg.RepoConfig[name].Start {
				fmt.Printf("[%s] launching: %s\n", name, command)
				if err := spawnTerminal(path, name, command); err != nil {
					fmt.Printf("[%s] warning: could not launch terminal: %v\n", name, err)
				}
			}
		}

		return nil
	},
}

func spawnTerminal(dir, repoName, command string) error {
	title := fmt.Sprintf("repoman — %s", repoName)
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("wt"); err == nil {
			cmd = exec.Command("wt", "--title", title,
				"--startingDirectory", dir,
				"cmd", "/K", command)
		} else {
			cmd = exec.Command("cmd", "/C", "start",
				title, "cmd", "/K",
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
