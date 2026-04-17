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
	if runtime.GOOS == "windows" {
		t.Skip("echo behaves differently on Windows without shell")
	}
	dir := t.TempDir()
	var buf bytes.Buffer

	err := runStreamingTo(dir, "test", "echo hello", &buf)
	if err != nil {
		t.Fatalf("RunStreaming: %v", err)
	}
	if !strings.Contains(buf.String(), "[test] hello") {
		t.Errorf("expected prefixed output, got: %q", buf.String())
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
