package tools_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mattt/aseprite-mcp/internal/aseprite"
	"github.com/mattt/aseprite-mcp/internal/asepritetest"
	"github.com/mattt/aseprite-mcp/internal/tools"
	"github.com/mattt/aseprite-mcp/internal/workspace"
)

func connect(t *testing.T, deps tools.Deps) *mcp.ClientSession {
	t.Helper()
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	tools.AddExecuteTool(srv, deps)
	tools.AddInspectTool(srv, deps)

	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "test"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func depsForRoot(t *testing.T, root string) tools.Deps {
	t.Helper()
	fake := asepritetest.Build(t)
	ws, err := workspace.New(root)
	if err != nil {
		t.Fatal(err)
	}
	return tools.Deps{
		Runner:        aseprite.NewRunner(fake, ws, 10*time.Second, 1<<20),
		Workspace:     ws,
		MaxScriptSize: 1 << 20,
		MaxImageSize:  10 << 20,
	}
}

func TestListToolsExposesExactlyTwo(t *testing.T) {
	cs := connect(t, depsForRoot(t, t.TempDir()))
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	names := map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
		if tool.InputSchema == nil {
			t.Errorf("tool %q has no input schema", tool.Name)
		}
	}
	if len(res.Tools) != 2 || !names["execute_lua"] || !names["inspect_export"] {
		t.Errorf("tools = %v, want exactly execute_lua and inspect_export", names)
	}
}

func TestExecuteLuaSuccess(t *testing.T) {
	t.Setenv("FAKE_STDOUT", "done")
	cs := connect(t, depsForRoot(t, t.TempDir()))
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "execute_lua",
		Arguments: map[string]any{"script": "print('hi')"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %+v", res.Content)
	}
	if !containsText(res.Content, "done") {
		t.Errorf("content missing stdout: %+v", res.Content)
	}
}

func TestExecuteLuaEmptyScriptIsToolError(t *testing.T) {
	cs := connect(t, depsForRoot(t, t.TempDir()))
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "execute_lua",
		Arguments: map[string]any{"script": "   "},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Error("empty script: expected tool error result")
	}
}

func TestExecuteLuaNonZeroExitIsToolError(t *testing.T) {
	t.Setenv("FAKE_EXIT", "3")
	t.Setenv("FAKE_STDERR", "lua boom")
	cs := connect(t, depsForRoot(t, t.TempDir()))
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "execute_lua",
		Arguments: map[string]any{"script": "error('x')"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatal("nonzero exit: expected tool error result")
	}
	if !containsText(res.Content, "lua boom") {
		t.Errorf("diagnostics missing stderr: %+v", res.Content)
	}
}

func TestInspectExportReturnsImage(t *testing.T) {
	root := t.TempDir()
	pngPath := filepath.Join(root, "sprite.png")
	writePNG(t, pngPath, 12, 8)

	cs := connect(t, depsForRoot(t, root))
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "inspect_export",
		Arguments: map[string]any{"path": "sprite.png"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %+v", res.Content)
	}
	var img *mcp.ImageContent
	for _, c := range res.Content {
		if ic, ok := c.(*mcp.ImageContent); ok {
			img = ic
		}
	}
	if img == nil {
		t.Fatalf("no image content in %+v", res.Content)
	}
	if img.MIMEType != "image/png" {
		t.Errorf("MIMEType = %q, want image/png", img.MIMEType)
	}
	if len(img.Data) == 0 {
		t.Error("image data is empty")
	}
}

func TestInspectExportRejectsNonPNG(t *testing.T) {
	root := t.TempDir()
	txt := filepath.Join(root, "note.txt")
	if err := os.WriteFile(txt, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	cs := connect(t, depsForRoot(t, root))
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "inspect_export",
		Arguments: map[string]any{"path": "note.txt"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Error("non-PNG: expected tool error result")
	}
}

func TestInspectExportRejectsTraversal(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(base, "outside.png")
	writePNG(t, outside, 4, 4)
	root := filepath.Join(base, "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	cs := connect(t, depsForRoot(t, root))
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "inspect_export",
		Arguments: map[string]any{"path": filepath.Join("..", "outside.png")},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Error("traversal path: expected tool error result")
	}
}

func containsText(content []mcp.Content, want string) bool {
	for _, c := range content {
		if tc, ok := c.(*mcp.TextContent); ok && bytes.Contains([]byte(tc.Text), []byte(want)) {
			return true
		}
	}
	return false
}

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
