package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

var nonAlphanumRe = regexp.MustCompile(`[^a-zA-Z0-9]`)

func flagFile(repoName string, index int) string {
	safe := nonAlphanumRe.ReplaceAllString(repoName, "_")
	return filepath.Join(os.TempDir(), fmt.Sprintf("repoman_%s_%d.flag", safe, index))
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

	for i, command := range commands {
		flag := flagFile(repoName, i)

		// Create the trigger flag — task checks for this on every workspace open.
		// If missing (normal VS Code open), the task exits immediately and the
		// terminal closes. Only repoman start creates the flag.
		if err := os.WriteFile(flag, []byte{}, 0644); err != nil {
			return fmt.Errorf("could not create flag for %q: %w", command, err)
		}

		// cmd /K keeps the terminal open after the command exits so the user
		// can see output. exit 0 still closes cmd, so the "no flag" path closes
		// the terminal cleanly when combined with "close": true.
		taskCmd := fmt.Sprintf(
			`IF EXIST "%s" (del "%s" & cd /d "%s" & %s) ELSE (exit 0)`,
			flag, flag, repoDir, command,
		)

		kept = append(kept, map[string]interface{}{
			"label":   fmt.Sprintf("repoman: %s — %s", repoName, command),
			"type":    "shell",
			"command": taskCmd,
			"options": map[string]interface{}{
				"shell": map[string]interface{}{
					"executable": "cmd.exe",
					"args":       []string{"/D", "/K"},
				},
			},
			"runOptions": map[string]interface{}{
				"runOn": "folderOpen",
			},
			"presentation": map[string]interface{}{
				"panel":            "new",
				"focus":            false,
				"close":            true,
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
