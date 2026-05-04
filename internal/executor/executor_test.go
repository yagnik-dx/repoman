package executor

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestPrefixWriter(t *testing.T) {
	var buf bytes.Buffer
	w := &prefixWriter{prefix: "[myrepo] ", w: &buf}

	input := "line one\nline two\n"
	w.Write([]byte(input))

	got := buf.String()
	if !strings.Contains(got, "[myrepo] line one") {
		t.Errorf("missing prefix on line one: %q", got)
	}
	if !strings.Contains(got, "[myrepo] line two") {
		t.Errorf("missing prefix on line two: %q", got)
	}
}

func TestRunStreaming(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	var command string
	if runtime.GOOS == "windows" {
		command = "echo hello"
	} else {
		command = "echo hello"
	}

	err := runStreamingTo(dir, "test", command, &buf)
	if err != nil {
		t.Fatalf("RunStreaming: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if !strings.Contains(got, "[test]") {
		t.Errorf("expected prefixed output, got: %q", got)
	}
}

func TestPrefixWriterEmptyTrailingLine(t *testing.T) {
	var buf bytes.Buffer
	w := &prefixWriter{prefix: "[r] ", w: &buf}
	w.Write([]byte("hello\n"))
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 || lines[0] != "[r] hello" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}
