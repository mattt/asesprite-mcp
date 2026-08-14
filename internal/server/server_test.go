package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mattt/aseprite-mcp/internal/aseprite"
	"github.com/mattt/aseprite-mcp/internal/asepritetest"
	"github.com/mattt/aseprite-mcp/internal/server"
	"github.com/mattt/aseprite-mcp/internal/tools"
	"github.com/mattt/aseprite-mcp/internal/workspace"
)

func TestNewRegistersBothTools(t *testing.T) {
	fake := asepritetest.Build(t)
	ws, err := workspace.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	deps := tools.Deps{
		Runner:        aseprite.NewRunner(fake, ws, time.Second, 1<<20),
		Workspace:     ws,
		MaxScriptSize: 1 << 20,
		MaxImageSize:  1 << 20,
	}
	srv := server.New("test", deps)

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
	defer cs.Close()

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(res.Tools) != 2 {
		t.Errorf("registered %d tools, want 2", len(res.Tools))
	}
}
