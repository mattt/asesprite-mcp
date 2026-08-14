package aseprite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/mattt/aseprite-mcp/internal/workspace"
)

// ScriptParamWorkspace is the --script-param key that gives the workspace root
// to the Lua environment. A script reads it through app.params.
const ScriptParamWorkspace = "aseprite_mcp_workspace"

// Runner runs Lua scripts. It starts one Aseprite batch process for each call.
type Runner struct {
	execPath  string
	ws        *workspace.Workspace
	timeout   time.Duration
	maxOutput int64
}

// NewRunner makes a Runner. timeout and maxOutput must be positive. The caller
// checks these values through config.
func NewRunner(execPath string, ws *workspace.Workspace, timeout time.Duration, maxOutput int64) *Runner {
	return &Runner{execPath: execPath, ws: ws, timeout: timeout, maxOutput: maxOutput}
}

// Result reports the outcome of one batch process.
type Result struct {
	Stdout          string
	Stderr          string
	ExitCode        int
	Duration        time.Duration
	TimedOut        bool
	StdoutTruncated bool
	StderrTruncated bool
}

// Execute writes the script to a temporary file inside the workspace, runs it
// through Aseprite in batch mode, and returns the captured result. Execute
// reports a non-zero exit code in Result, not as an error. It returns an error
// only when it cannot start the process or write the temporary script.
func (r *Runner) Execute(ctx context.Context, script string) (*Result, error) {
	f, err := os.CreateTemp(r.ws.Root(), "aseprite-mcp-*.lua")
	if err != nil {
		return nil, fmt.Errorf("aseprite: create temp script: %w", err)
	}
	scriptPath := f.Name()
	defer os.Remove(scriptPath)

	if _, err := f.WriteString(script); err != nil {
		f.Close()
		return nil, fmt.Errorf("aseprite: write temp script: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("aseprite: close temp script: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	args := []string{
		"-b",
		"--script-param", ScriptParamWorkspace + "=" + r.ws.Root(),
		"--script", scriptPath,
	}
	cmd := exec.CommandContext(ctx, r.execPath, args...)
	cmd.Dir = r.ws.Root()

	stdout := &boundedBuffer{limit: r.maxOutput}
	stderr := &boundedBuffer{limit: r.maxOutput}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("aseprite: start %q: %w", r.execPath, err)
	}
	waitErr := cmd.Wait()
	elapsed := time.Since(start)

	res := &Result{
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		Duration:        elapsed,
		StdoutTruncated: stdout.truncated,
		StderrTruncated: stderr.truncated,
	}
	if ctx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
		res.ExitCode = -1
		return res, nil
	}
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
			return res, nil
		}
		return nil, fmt.Errorf("aseprite: run: %w", waitErr)
	}
	return res, nil
}

// boundedBuffer captures up to limit bytes. It records whether the writer sent
// more.
type boundedBuffer struct {
	limit     int64
	buf       []byte
	truncated bool
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	remaining := b.limit - int64(len(b.buf))
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if int64(len(p)) > remaining {
		b.buf = append(b.buf, p[:remaining]...)
		b.truncated = true
		return len(p), nil
	}
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *boundedBuffer) String() string { return string(b.buf) }
