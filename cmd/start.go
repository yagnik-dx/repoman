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
	Short: "Select repos to start, open their VS Code workspace with integrated terminals",
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
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("[%s] skipped: directory not found\n", name)
				continue
			}
			repoCfg := cfg.RepoConfig[name]
			fmt.Printf("[%s] injecting tasks...\n", name)
			if err := injectRepomanTasks(cfg.Workspace, path, name, repoCfg.Start); err != nil {
				fmt.Printf("[%s] warning: %v\n", name, err)
			}
		}

		fmt.Println("Opening workspace...")
		if err := exec.Command("code", cfg.Workspace).Start(); err != nil {
			return fmt.Errorf("could not open VS Code: %w", err)
		}

		return nil
	},
}

func injectRepomanTasks(workspacePath, repoName, repoDir string, commands []string) error {
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

	// Preserve non-repoman tasks, replace repoman ones for this repo.
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

	for _, command := range commands {
		kept = append(kept, map[string]interface{}{
			"label":   fmt.Sprintf("repoman: %s — %s", repoName, command),
			"type":    "shell",
			"command": command,
			"options": map[string]interface{}{"cwd": repoDir},
			"runOptions": map[string]interface{}{
				"runOn": "folderOpen",
			},
			"presentation": map[string]interface{}{
				"panel":            "new",
				"focus":            false,
				"showReuseMessage": false,
			},
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
