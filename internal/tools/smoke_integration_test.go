//go:build integration

// This opt-in smoke test runs a real Aseprite installation from end to end.
// The installation needs a separate license. Run the test with a build tag and
// an executable path:
//
//	ASEPRITE_PATH=/path/to/aseprite go test -tags integration ./internal/tools/ -run Smoke -v
//
// The default build does not include this test, so CI never needs Aseprite.
package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mattt/aseprite-mcp/internal/aseprite"
	"github.com/mattt/aseprite-mcp/internal/server"
	"github.com/mattt/aseprite-mcp/internal/tools"
	"github.com/mattt/aseprite-mcp/internal/workspace"
)

func TestSmokeRealAseprite(t *testing.T) {
	execPath := os.Getenv("ASEPRITE_PATH")
	if execPath == "" {
		t.Skip("set ASEPRITE_PATH to run the real-Aseprite smoke test")
	}

	root := t.TempDir()
	ws, err := workspace.New(root)
	if err != nil {
		t.Fatal(err)
	}
	deps := tools.Deps{
		Runner:        aseprite.NewRunner(execPath, ws, 60*time.Second, 1<<20),
		Workspace:     ws,
		MaxScriptSize: 1 << 20,
		MaxImageSize:  10 << 20,
	}

	srv := server.New("test", deps)
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "smoke", Version: "test"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	script := `
local dir = app.params["aseprite_mcp_workspace"]
local s = Sprite(32, 32, ColorMode.RGB)
local img = s.cels[1].image
for i = 0, 31 do img:drawPixel(i, i, Color{ r = 255, g = 120, b = 40, a = 255 }) end
s:saveAs(dir .. "/hero.aseprite")
s:saveCopyAs(dir .. "/hero.png")
print("ok")
`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "execute_lua",
		Arguments: map[string]any{"script": script},
	})
	if err != nil {
		t.Fatalf("execute_lua call: %v", err)
	}
	if res.IsError {
		t.Fatalf("execute_lua failed: %+v", res.Content)
	}

	for _, name := range []string{"hero.aseprite", "hero.png"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Errorf("expected %s: %v", name, err)
		}
	}

	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".lua" {
			t.Errorf("temp script %q was not cleaned up", e.Name())
		}
	}

	inspect, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "inspect_export",
		Arguments: map[string]any{"path": "hero.png"},
	})
	if err != nil {
		t.Fatalf("inspect_export call: %v", err)
	}
	if inspect.IsError {
		t.Fatalf("inspect_export failed: %+v", inspect.Content)
	}
	hasImage := false
	for _, c := range inspect.Content {
		if _, ok := c.(*mcp.ImageContent); ok {
			hasImage = true
		}
	}
	if !hasImage {
		t.Error("inspect_export returned no image content")
	}
}
