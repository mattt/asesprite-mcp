// Package server builds the MCP server and registers the two tools.
package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mattt/aseprite-mcp/internal/tools"
)

// New builds an MCP server with execute_lua and inspect_export.
func New(version string, deps tools.Deps) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "aseprite-mcp",
		Title:   "Aseprite Code Mode",
		Version: version,
	}, nil)
	tools.AddExecuteTool(s, deps)
	tools.AddInspectTool(s, deps)
	return s
}
