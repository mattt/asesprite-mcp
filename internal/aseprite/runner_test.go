package aseprite_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mattt/aseprite-mcp/internal/aseprite"
	"github.com/mattt/aseprite-mcp/internal/asepritetest"
	"github.com/mattt/aseprite-mcp/internal/workspace"
)

func newRunner(t *testing.T, root string, timeout time.Duration, maxOutput int64) *aseprite.Runner {
	t.Helper()
	fake := asepritetest.Build(t)
	ws, err := workspace.New(root)
	if err != nil {
		t.Fatal(err)
	}
	return aseprite.NewRunner(fake, ws, timeout, maxOutput)
}

func TestExecutePassesArgsAndWorkspace(t *testing.T) {
	root := t.TempDir()
	argsFile := filepath.Join(root, "args.txt")
	scriptCopy := filepath.Join(root, "script.lua")
	t.Setenv("FAKE_ARGS_FILE", argsFile)
	t.Setenv("FAKE_SCRIPT_COPY", scriptCopy)

	r := newRunner(t, root, 10*time.Second, 1<<20)
	res, err := r.Execute(context.Background(), "print('hi')")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.ExitCode != 0 || res.TimedOut {
		t.Fatalf("unexpected result: %+v", res)
	}

	argv, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	got := string(argv)
	if !strings.Contains(got, "-b") {
		t.Errorf("argv %q missing -b", got)
	}
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	if !strings.Contains(got, "aseprite_mcp_workspace="+resolvedRoot) {
		t.Errorf("argv %q missing workspace param %q", got, resolvedRoot)
	}
	if !strings.Contains(got, "--script") {
		t.Errorf("argv %q missing --script", got)
	}

	copied, err := os.ReadFile(scriptCopy)
	if err != nil {
		t.Fatal(err)
	}
	if string(copied) != "print('hi')" {
		t.Errorf("script content = %q, want print('hi')", copied)
	}
}

func TestExecuteWorkspaceWithSpaces(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "my project")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	argsFile := filepath.Join(root, "args.txt")
	t.Setenv("FAKE_ARGS_FILE", argsFile)

	r := newRunner(t, root, 10*time.Second, 1<<20)
	if _, err := r.Execute(context.Background(), "x=1"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	argv, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	if !strings.Contains(string(argv), resolvedRoot) {
		t.Errorf("argv %q missing spaced workspace %q", argv, resolvedRoot)
	}
}

func TestExecuteWritesFileInWorkspace(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FAKE_WRITE_FILE", "out.txt")
	r := newRunner(t, root, 10*time.Second, 1<<20)
	if _, err := r.Execute(context.Background(), "x=1"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	if _, err := os.Stat(filepath.Join(resolvedRoot, "out.txt")); err != nil {
		t.Errorf("expected out.txt in workspace: %v", err)
	}
}

func TestExecuteNonZeroExit(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FAKE_EXIT", "7")
	t.Setenv("FAKE_STDERR", "boom")
	r := newRunner(t, root, 10*time.Second, 1<<20)
	res, err := r.Execute(context.Background(), "error('x')")
	if err != nil {
		t.Fatalf("Execute returned error, want captured result: %v", err)
	}
	if res.ExitCode != 7 {
		t.Errorf("ExitCode = %d, want 7", res.ExitCode)
	}
	if res.Stderr != "boom" {
		t.Errorf("Stderr = %q, want boom", res.Stderr)
	}
}

func TestExecuteTimeout(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FAKE_SLEEP_MS", "2000")
	r := newRunner(t, root, 100*time.Millisecond, 1<<20)
	res, err := r.Execute(context.Background(), "x=1")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !res.TimedOut {
		t.Errorf("TimedOut = false, want true (result %+v)", res)
	}
}

func TestExecuteCancellation(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FAKE_SLEEP_MS", "2000")
	r := newRunner(t, root, 10*time.Second, 1<<20)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	if _, err := r.Execute(ctx, "x=1"); err != nil {
		// A cancelled process can return either an error or a result. Both are
		// correct, if the call returns soon.
		_ = err
	}
	if time.Since(start) > 1500*time.Millisecond {
		t.Error("Execute did not return promptly after cancellation")
	}
}

func TestExecuteOutputLimit(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FAKE_STDOUT_BYTES", "5000")
	r := newRunner(t, root, 10*time.Second, 1000)
	res, err := r.Execute(context.Background(), "x=1")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if int64(len(res.Stdout)) > 1000 {
		t.Errorf("captured %d bytes, want <= 1000", len(res.Stdout))
	}
	if !res.StdoutTruncated {
		t.Error("StdoutTruncated = false, want true")
	}
}

func TestExecuteCleansUpTempScript(t *testing.T) {
	root := t.TempDir()
	r := newRunner(t, root, 10*time.Second, 1<<20)
	if _, err := r.Execute(context.Background(), "x=1"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "aseprite-mcp-") {
			t.Errorf("temp script %q was not removed", e.Name())
		}
	}
}

func TestExecuteStartFailure(t *testing.T) {
	root := t.TempDir()
	ws, err := workspace.New(root)
	if err != nil {
		t.Fatal(err)
	}
	r := aseprite.NewRunner(filepath.Join(root, "does-not-exist"), ws, time.Second, 1<<20)
	if _, err := r.Execute(context.Background(), "x=1"); err == nil {
		t.Error("Execute with missing executable: expected error, got nil")
	}
}
