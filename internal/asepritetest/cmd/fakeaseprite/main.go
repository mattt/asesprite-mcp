// Command fakeaseprite takes the place of the Aseprite executable in tests. It
// runs on any platform and does not start Aseprite. Environment variables set
// its behavior, so the tests for the runner need no licensed Aseprite
// installation.
//
// The command reads these variables:
//
//	FAKE_STDOUT        text written to stdout
//	FAKE_STDERR        text written to stderr
//	FAKE_EXIT          integer exit code (default 0)
//	FAKE_SLEEP_MS      milliseconds to sleep before exiting
//	FAKE_STDOUT_BYTES  number of filler bytes written to stdout
//	FAKE_ARGS_FILE     path to write the received argv, one per line
//	FAKE_SCRIPT_COPY   path to copy the --script file contents into
//	FAKE_WRITE_FILE    path (relative to the workspace param) to create
package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	args := os.Args[1:]

	if p := os.Getenv("FAKE_ARGS_FILE"); p != "" {
		_ = os.WriteFile(p, []byte(strings.Join(args, "\n")), 0o644)
	}

	scriptPath, workspace := parse(args)

	if p := os.Getenv("FAKE_SCRIPT_COPY"); p != "" && scriptPath != "" {
		if data, err := os.ReadFile(scriptPath); err == nil {
			_ = os.WriteFile(p, data, 0o644)
		}
	}

	if name := os.Getenv("FAKE_WRITE_FILE"); name != "" && workspace != "" {
		_ = os.WriteFile(filepath.Join(workspace, name), []byte("created by fakeaseprite"), 0o644)
	}

	if ms := os.Getenv("FAKE_SLEEP_MS"); ms != "" {
		if n, err := strconv.Atoi(ms); err == nil {
			time.Sleep(time.Duration(n) * time.Millisecond)
		}
	}

	if s := os.Getenv("FAKE_STDOUT"); s != "" {
		os.Stdout.WriteString(s)
	}
	if n := os.Getenv("FAKE_STDOUT_BYTES"); n != "" {
		if count, err := strconv.Atoi(n); err == nil && count > 0 {
			os.Stdout.Write([]byte(strings.Repeat("a", count)))
		}
	}
	if s := os.Getenv("FAKE_STDERR"); s != "" {
		os.Stderr.WriteString(s)
	}

	code := 0
	if s := os.Getenv("FAKE_EXIT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			code = n
		}
	}
	os.Exit(code)
}

// parse reads the --script value and the aseprite_mcp_workspace parameter.
func parse(args []string) (scriptPath, workspace string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--script":
			if i+1 < len(args) {
				scriptPath = args[i+1]
			}
		case "--script-param":
			if i+1 < len(args) {
				if k, v, ok := strings.Cut(args[i+1], "="); ok && k == "aseprite_mcp_workspace" {
					workspace = v
				}
			}
		}
	}
	return scriptPath, workspace
}
