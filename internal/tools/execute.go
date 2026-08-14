package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mattt/aseprite-mcp/internal/aseprite"
)

// ExecuteInput is the argument for execute_lua.
type ExecuteInput struct {
	Script string `json:"script" jsonschema:"Lua source executed by Aseprite in batch mode. Write one coherent program. Read the workspace root from app.params.aseprite_mcp_workspace and save output beneath it."`
}

// ExecuteOutput is the result of a successful execute_lua call.
type ExecuteOutput struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	Truncated  bool   `json:"truncated"`
}

// AddExecuteTool registers execute_lua on the server.
func AddExecuteTool(s *mcp.Server, deps Deps) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "execute_lua",
		Description: "Run a Lua script through Aseprite in batch mode (one process per call). The workspace root is passed as the script parameter aseprite_mcp_workspace; write files beneath it. Returns stdout, stderr, and exit status. Use inspect_export afterward to view exported PNGs.",
		Annotations: &mcp.ToolAnnotations{Title: "Execute Aseprite Lua"},
	}, deps.execute)
}

func (d Deps) execute(ctx context.Context, _ *mcp.CallToolRequest, in ExecuteInput) (*mcp.CallToolResult, ExecuteOutput, error) {
	if strings.TrimSpace(in.Script) == "" {
		return errorResult("script is empty"), ExecuteOutput{}, nil
	}
	if int64(len(in.Script)) > d.MaxScriptSize {
		return errorResult(fmt.Sprintf("script is %d bytes, exceeding the %d byte limit", len(in.Script), d.MaxScriptSize)),
			ExecuteOutput{}, nil
	}

	res, err := d.Runner.Execute(ctx, in.Script)
	if err != nil {
		return errorResult("could not start Aseprite: " + err.Error()), ExecuteOutput{}, nil
	}

	truncated := res.StdoutTruncated || res.StderrTruncated
	if res.TimedOut {
		return errorResult(diagnostics("Aseprite timed out", res, truncated)), ExecuteOutput{}, nil
	}
	if res.ExitCode != 0 {
		return errorResult(diagnostics(fmt.Sprintf("Aseprite exited with code %d", res.ExitCode), res, truncated)),
			ExecuteOutput{}, nil
	}

	out := ExecuteOutput{
		Stdout:     res.Stdout,
		Stderr:     res.Stderr,
		ExitCode:   res.ExitCode,
		DurationMs: res.Duration.Milliseconds(),
		Truncated:  truncated,
	}
	summary := fmt.Sprintf("Aseprite finished in %d ms.", out.DurationMs)
	if truncated {
		summary += " Output was truncated."
	}
	if res.Stdout != "" {
		summary += "\n--- stdout ---\n" + res.Stdout
	}
	if res.Stderr != "" {
		summary += "\n--- stderr ---\n" + res.Stderr
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: summary}}}, out, nil
}

func diagnostics(headline string, res *aseprite.Result, truncated bool) string {
	var b strings.Builder
	b.WriteString(headline)
	fmt.Fprintf(&b, " after %d ms.", res.Duration.Milliseconds())
	if truncated {
		b.WriteString(" Output was truncated.")
	}
	if res.Stdout != "" {
		b.WriteString("\n--- stdout ---\n")
		b.WriteString(res.Stdout)
	}
	if res.Stderr != "" {
		b.WriteString("\n--- stderr ---\n")
		b.WriteString(res.Stderr)
	}
	return b.String()
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}
