// Package asepritetest compiles a fake Aseprite executable for tests. The fake
// takes the place of a licensed Aseprite installation, so the tests for the
// runner and the MCP tools run on any platform with no external program.
package asepritetest

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var (
	once  sync.Once
	path  string
	build error
)

// Build compiles the fake executable one time for each test process and
// returns its path. It fails the test if the build fails.
func Build(t *testing.T) string {
	t.Helper()
	once.Do(func() {
		dir, err := os.MkdirTemp("", "fakeaseprite")
		if err != nil {
			build = err
			return
		}
		name := "fakeaseprite"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		out := filepath.Join(dir, name)
		cmd := exec.Command("go", "build", "-o", out, "github.com/mattt/aseprite-mcp/internal/asepritetest/cmd/fakeaseprite")
		if combined, err := cmd.CombinedOutput(); err != nil {
			build = &buildError{output: string(combined), err: err}
			return
		}
		path = out
	})
	if build != nil {
		t.Fatalf("build fakeaseprite: %v", build)
	}
	return path
}

type buildError struct {
	output string
	err    error
}

func (e *buildError) Error() string { return e.err.Error() + ": " + e.output }
