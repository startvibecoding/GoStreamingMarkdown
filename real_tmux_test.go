//go:build realtest

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const realTmuxTimeout = 10 * time.Second

func TestRealTmuxOneShotRender(t *testing.T) {
	requireTmux(t)

	bin := buildRealTestBinary(t)
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.md")
	writeFile(t, inputPath, []byte(`# Real Terminal

This paragraph has **bold**, *italic*, and `+"`inline code`"+`.

- [x] Done
- [ ] Pending

| Name | Status |
|------|--------|
| Ada  | OK     |

`+"```go"+`
fmt.Println("tmux")
`+"```"+`
`), 0o644)

	const marker = "__GSM_REAL_DONE__"
	scriptPath := filepath.Join(dir, "oneshot.sh")
	writeFile(t, scriptPath, []byte(fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
TERM=xterm-256color %q -w 48 %q
printf '\n%s\n'
sleep 30
`, bin, inputPath, marker)), 0o755)

	session := newTmuxSession(t, scriptPath, 80, 24)
	screen := waitForPaneText(t, session, marker)

	assertPaneContains(t, screen,
		"# Real Terminal",
		"bold",
		"italic",
		"inline",
		"code.",
		"Done",
		"Pending",
		"Name",
		"Ada",
		"fmt.Println",
	)
	assertPaneDoesNotContain(t, screen, "\x1b[")
}

func TestRealTmuxStreamRenderKeepsLatestFrame(t *testing.T) {
	requireTmux(t)

	bin := buildRealTestBinary(t)
	dir := t.TempDir()
	fifoPath := filepath.Join(dir, "stream.fifo")
	scriptPath := filepath.Join(dir, "stream.sh")
	writeFile(t, scriptPath, []byte(fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
mkfifo %q
(
	exec 3>%q
	printf '# Stream Frame\n' >&3
	sleep 0.1
	printf '\nProcessing **delta** chunks.\n' >&3
	sleep 0.1
	printf '\n- [x] first\n- [ ] second\n' >&3
	sleep 0.1
	printf '\n`+"```go"+`\nfmt.Println("stream")\n`+"```"+`\n' >&3
	sleep 30
) &
TERM=xterm-256color %q --stream --delay 5ms -w 52 < %q
`, fifoPath, fifoPath, bin, fifoPath)), 0o755)

	session := newTmuxSession(t, scriptPath, 80, 24)
	screen := waitForPaneText(t, session, `fmt.Println("stream")`)

	assertPaneContains(t, screen,
		"# Stream Frame",
		"Processing",
		"delta",
		"first",
		"second",
		"fmt.Println",
	)
	assertPaneDoesNotContain(t, screen, "\x1b[")
}

func requireTmux(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux is required for real terminal tests")
	}
}

func buildRealTestBinary(t *testing.T) string {
	t.Helper()

	goTool := os.Getenv("GO")
	if goTool == "" {
		goTool = "go"
	}

	binPath := filepath.Join(t.TempDir(), "GoStreamingMarkdown")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, goTool, "build", "-o", binPath, ".")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build real-test binary: %v\n%s", err, out)
	}
	return binPath
}

func newTmuxSession(t *testing.T, scriptPath string, width, height int) string {
	t.Helper()

	session := fmt.Sprintf("gsm-real-%d-%d", os.Getpid(), time.Now().UnixNano())
	tmuxRun(t, "new-session", "-d", "-x", fmt.Sprint(width), "-y", fmt.Sprint(height), "-s", session, scriptPath)
	t.Cleanup(func() {
		_ = exec.Command("tmux", "kill-session", "-t", session).Run()
	})
	return session
}

func waitForPaneText(t *testing.T, session, want string) string {
	t.Helper()

	deadline := time.Now().Add(realTmuxTimeout)
	var screen string
	for time.Now().Before(deadline) {
		screen = capturePane(t, session)
		if strings.Contains(screen, want) {
			return screen
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %q in tmux pane:\n%s", want, screen)
	return ""
}

func capturePane(t *testing.T, session string) string {
	t.Helper()
	return tmuxRun(t, "capture-pane", "-p", "-t", session, "-S", "-2000")
}

func tmuxRun(t *testing.T, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), realTmuxTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tmux", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("tmux %s: %v\nstdout:\n%s\nstderr:\n%s", strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
	return stdout.String()
}

func writeFile(t *testing.T, path string, data []byte, perm os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, data, perm); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertPaneContains(t *testing.T, screen string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(screen, fragment) {
			t.Fatalf("tmux pane is missing %q:\n%s", fragment, screen)
		}
	}
}

func assertPaneDoesNotContain(t *testing.T, screen, fragment string) {
	t.Helper()
	if strings.Contains(screen, fragment) {
		t.Fatalf("tmux pane unexpectedly contains %q:\n%s", fragment, screen)
	}
}
