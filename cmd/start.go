package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"repoman/internal/config"
	"repoman/internal/ui"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Select repos to start, open VS Code workspace with integrated terminals per start command",
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

		if cfg.Workspace == "" {
			fmt.Println("No workspace configured. Run repoman setup to set one.")
			return nil
		}

		for _, name := range chosen {
			path := repoPath(cfg.BasePath, name)
			repoCfg := cfg.RepoConfig[name]
			fmt.Printf("[%s] injecting tasks...\n", name)
			if err := injectRepomanTasks(cfg.Workspace, path, name, repoCfg.Start); err != nil {
				fmt.Printf("[%s] warning: %v\n", name, err)
			}
		}

		fmt.Println("Opening workspace...")
		return exec.Command("code", cfg.Workspace).Start()
	},
}

func injectRepomanTasks(workspacePath, repoDir, repoName string, commands []string) error {
	data, err := os.ReadFile(workspacePath)
	if err != nil {
		return err
	}

	var ws map[string]interface{}
	if err := json.Unmarshal(data, &ws); err != nil {
		return fmt.Errorf("could not parse workspace file: %w", err)
	}

	tasksSection, _ := ws["tasks"].(map[string]interface{})
	if tasksSection == nil {
		tasksSection = map[string]interface{}{"version": "2.0.0"}
	}

	labelPrefix := fmt.Sprintf("repoman: %s", repoName)
	var kept []interface{}
	if arr, ok := tasksSection["tasks"].([]interface{}); ok {
		for _, t := range arr {
			if tm, ok := t.(map[string]interface{}); ok {
				if label, _ := tm["label"].(string); strings.HasPrefix(label, labelPrefix) {
					continue
				}
			}
			kept = append(kept, t)
		}
	}

	// Embed cd into the command so the terminal lands in the right directory
	// regardless of what cwd VS Code resolves to.
	for _, command := range commands {
		fullCmd := fmt.Sprintf(`cd /d "%s" && %s`, repoDir, command)
		kept = append(kept, map[string]interface{}{
			"label":   fmt.Sprintf("repoman: %s — %s", repoName, command),
			"type":    "shell",
			"command": fullCmd,
			"runOptions": map[string]interface{}{
				"runOn": "folderOpen",
			},
			"presentation": map[string]interface{}{
				"panel":            "new",
				"focus":            false,
				"showReuseMessage": false,
			},
			"problemMatcher": []interface{}{},
		})
	}

	tasksSection["tasks"] = kept
	ws["tasks"] = tasksSection

	out, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(workspacePath, out, 0644)
}

func init() {
	rootCmd.AddCommand(startCmd)
}
