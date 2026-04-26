package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

		type entry struct {
			label   string
			batPath string
			repoDir string
		}
		var entries []startEntry

		for _, name := range chosen {
			dir := repoPath(cfg.BasePath, name)
			for i, command := range cfg.RepoConfig[name].Start {
				batPath := batchFilePath(name, i)
				bat := fmt.Sprintf("@echo off\r\ncd /d \"%s\"\r\n%s\r\n", dir, command)
				if err := os.WriteFile(batPath, []byte(bat), 0644); err != nil {
					fmt.Printf("[%s] warning: could not write batch file: %v\n", name, err)
					continue
				}
				entries = append(entries, startEntry{
					label:   fmt.Sprintf("repoman: %s — %s", name, command),
					batPath: batPath,
					repoDir: dir,
				})
			}
		}

		if len(entries) == 0 {
			fmt.Println("Nothing to start.")
			return nil
		}

		if err := injectTasks(cfg.Workspace, entries); err != nil {
			return fmt.Errorf("could not update workspace: %w", err)
		}

		fmt.Println("Opening workspace...")
		if err := exec.Command("code", cfg.Workspace).Start(); err != nil {
			return fmt.Errorf("could not open VS Code: %w", err)
		}

		// Give VS Code a moment to load, then trigger the repoman build task.
		fmt.Println("Waiting for VS Code to load...")
		time.Sleep(5 * time.Second)
		_ = exec.Command("code", "--command", "workbench.action.tasks.build").Start()

		return nil
	},
}

var safeNameRe = regexp.MustCompile(`[^a-zA-Z0-9]`)

func batchFilePath(repoName string, index int) string {
	safe := safeNameRe.ReplaceAllString(repoName, "_")
	return filepath.Join(os.TempDir(), fmt.Sprintf("repoman_%s_%d.bat", safe, index))
}

type startEntry struct {
	label   string
	batPath string
	repoDir string
}

func injectTasks(workspacePath string, entries []startEntry) error {
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

	// Keep non-repoman tasks, drop old repoman tasks.
	var kept []interface{}
	if arr, ok := tasksSection["tasks"].([]interface{}); ok {
		for _, t := range arr {
			if tm, ok := t.(map[string]interface{}); ok {
				if label, _ := tm["label"].(string); strings.HasPrefix(label, "repoman:") {
					continue
				}
			}
			kept = append(kept, t)
		}
	}

	// Add one task per command (no runOn — never auto-starts).
	var labels []string
	for _, e := range entries {
		labels = append(labels, e.label)
		kept = append(kept, map[string]interface{}{
			"label":   e.label,
			"type":    "shell",
			"command": e.batPath,
			"options": map[string]interface{}{
				"shell": map[string]interface{}{
					"executable": "cmd.exe",
					"args":       []string{"/D", "/K"},
				},
			},
			"presentation": map[string]interface{}{
				"panel":            "new",
				"focus":            false,
				"showReuseMessage": false,
			},
			"problemMatcher": []interface{}{},
		})
	}

	// Compound task that launches everything at once, set as the default
	// build task so "code --command workbench.action.tasks.build" triggers it.
	kept = append(kept, map[string]interface{}{
		"label":        "repoman: start all",
		"dependsOn":    labels,
		"dependsOrder": "parallel",
		"group": map[string]interface{}{
			"kind":      "build",
			"isDefault": true,
		},
		"problemMatcher": []interface{}{},
	})

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
