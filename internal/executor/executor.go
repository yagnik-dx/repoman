package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// RunStreaming runs a shell command string in dir, prefixing all output lines with [repoName].
func RunStreaming(dir, repoName, command string) error {
	return runStreamingTo(dir, repoName, command, os.Stdout)
}

func runStreamingTo(dir, repoName, command string, w io.Writer) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	pw := &prefixWriter{prefix: fmt.Sprintf("[%s] ", repoName), w: w}
	cmd.Stdout = pw
	cmd.Stderr = pw
	return cmd.Run()
}

type prefixWriter struct {
	prefix string
	w      io.Writer
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		fmt.Fprintf(p.w, "%s%s\n", p.prefix, line)
	}
	return len(b), nil
}
