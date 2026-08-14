package tools

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InspectInput is the argument for inspect_export.
type InspectInput struct {
	Path string `json:"path" jsonschema:"Path to a PNG file inside the workspace, relative to the workspace root or absolute. The file must already exist."`
}

// InspectOutput is the result of a successful inspect_export call.
type InspectOutput struct {
	Path     string `json:"path"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	ByteSize int64  `json:"byte_size"`
}

// AddInspectTool registers inspect_export on the server.
func AddInspectTool(s *mcp.Server, deps Deps) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "inspect_export",
		Description: "Return an exported PNG from the workspace as image content so you can see the result and refine it. The file must exist and stay inside the workspace.",
		Annotations: &mcp.ToolAnnotations{Title: "Inspect Exported PNG", ReadOnlyHint: true},
	}, deps.inspect)
}

func (d Deps) inspect(_ context.Context, _ *mcp.CallToolRequest, in InspectInput) (*mcp.CallToolResult, InspectOutput, error) {
	if strings.TrimSpace(in.Path) == "" {
		return errorResult("path is empty"), InspectOutput{}, nil
	}
	if !strings.EqualFold(fileExt(in.Path), ".png") {
		return errorResult("only PNG files can be inspected"), InspectOutput{}, nil
	}

	resolved, err := d.Workspace.ResolveExisting(in.Path)
	if err != nil {
		return errorResult(err.Error()), InspectOutput{}, nil
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return errorResult("cannot stat file: " + err.Error()), InspectOutput{}, nil
	}
	if info.IsDir() {
		return errorResult("path is a directory, not a PNG file"), InspectOutput{}, nil
	}
	if info.Size() > d.MaxImageSize {
		return errorResult(fmt.Sprintf("file is %d bytes, exceeding the %d byte limit", info.Size(), d.MaxImageSize)),
			InspectOutput{}, nil
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return errorResult("cannot read file: " + err.Error()), InspectOutput{}, nil
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return errorResult("file is not a valid PNG: " + err.Error()), InspectOutput{}, nil
	}

	out := InspectOutput{
		Path:     resolved,
		Width:    cfg.Width,
		Height:   cfg.Height,
		ByteSize: info.Size(),
	}
	content := []mcp.Content{
		&mcp.ImageContent{Data: data, MIMEType: "image/png"},
		&mcp.TextContent{Text: fmt.Sprintf("%s: %dx%d PNG, %d bytes", resolved, out.Width, out.Height, out.ByteSize)},
	}
	return &mcp.CallToolResult{Content: content}, out, nil
}

// fileExt returns the file extension with the dot, or "" for none.
func fileExt(path string) string {
	i := strings.LastIndexByte(path, '.')
	if i < 0 {
		return ""
	}
	// Reject a dot that is part of a parent directory name.
	if strings.ContainsAny(path[i:], "/\\") {
		return ""
	}
	return path[i:]
}
