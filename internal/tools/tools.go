// Package tools registers the two MCP tools. execute_lua runs a Lua script
// through Aseprite. inspect_export returns an exported PNG as image content, so
// the agent can see the result and correct it.
package tools

import (
	"github.com/mattt/aseprite-mcp/internal/aseprite"
	"github.com/mattt/aseprite-mcp/internal/workspace"
)

// Deps holds the dependencies that the tool handlers share.
type Deps struct {
	Runner        *aseprite.Runner
	Workspace     *workspace.Workspace
	MaxScriptSize int64
	MaxImageSize  int64
}
