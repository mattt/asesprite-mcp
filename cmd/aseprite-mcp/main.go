// Command aseprite-mcp is a stdio MCP server. It runs Aseprite Lua scripts and
// returns exported PNGs for inspection. It does not include Aseprite. Supply an
// installation with a separate license through --aseprite or ASEPRITE_PATH.
package main

import (
	"context"
	"log"
	"os"
	"runtime"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mattt/aseprite-mcp/internal/aseprite"
	"github.com/mattt/aseprite-mcp/internal/config"
	"github.com/mattt/aseprite-mcp/internal/server"
	"github.com/mattt/aseprite-mcp/internal/tools"
	"github.com/mattt/aseprite-mcp/internal/workspace"
)

// version is set at build time with -ldflags.
var version = "dev"

func main() {
	// Log to stderr, so the stdio MCP stream on stdout stays clean.
	log.SetFlags(0)
	log.SetPrefix("aseprite-mcp: ")

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load("aseprite-mcp", os.Args[1:], os.Getenv, os.Stderr)
	if err != nil {
		return err
	}

	execPath := cfg.AsepritePath
	if execPath == "" {
		execPath, err = aseprite.Discover(runtime.GOOS)
		if err != nil {
			return err
		}
	}

	wsRoot := cfg.Workspace
	if wsRoot == "" {
		wsRoot, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	ws, err := workspace.New(wsRoot)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if v, err := aseprite.ProbeVersion(ctx, execPath); err != nil {
		log.Printf("warning: %v", err)
	} else {
		log.Printf("using Aseprite %q (%s)", v, execPath)
	}
	log.Printf("workspace: %s", ws.Root())

	runner := aseprite.NewRunner(execPath, ws, cfg.Timeout, cfg.MaxOutputBytes)
	deps := tools.Deps{
		Runner:        runner,
		Workspace:     ws,
		MaxScriptSize: cfg.MaxScriptBytes,
		MaxImageSize:  cfg.MaxImageBytes,
	}

	srv := server.New(version, deps)
	return srv.Run(ctx, &mcp.StdioTransport{})
}
