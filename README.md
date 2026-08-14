# aseprite-mcp

[![CI][ci badge]][ci]
[![License][license badge]][license]

aseprite-mcp is a [Model Context Protocol][mcp] server that makes pixel art with [Aseprite][aseprite].
Aseprite includes a complete Lua scripting API and can run it from the command line,
so the most direct way to give an agent the whole editor is to let it write and run scripts.
The server exposes that as two tools,
one to run a Lua script and
one to return an exported PNG,
with a reference for the API.

For each task,
the agent writes one Lua script that
draws, animates, and exports sprites in a single Aseprite process.
It then looks at the exported PNG and corrects its work.
This code-first approach keeps the tool count small and the context short.

> [!IMPORTANT]
> You must supply your own licensed copy of Aseprite.
> This project does not include, download, or redistribute it.

---

## Requirements

- Go 1.25 or later, to build the server.
- Aseprite 1.3 or later, installed and licensed by you (see the note above).

## Installation

```sh
go install github.com/mattt/aseprite-mcp/cmd/aseprite-mcp@latest
```

## Usage

Start the server and set the workspace directory:

```sh
aseprite-mcp --workspace /path/to/project
```

The server speaks over stdio.
Log messages go to stderr, so the protocol stream on stdout stays clean.

### Tools

- `execute_lua` runs a Lua script through Aseprite in batch mode.
  One Aseprite process runs for each call.
  The server passes the workspace path to the script as the parameter `aseprite_mcp_workspace`.
  The tool returns stdout, stderr, and the exit status.
- `inspect_export` returns an exported PNG from the workspace as image content.
  The agent sees the result and corrects the sprite.

### Configuration

A flag overrides its environment variable, which overrides the default.

| Flag | Environment | Default | Purpose |
| ------ | ------------- | --------- | --------- |
| `--aseprite` | `ASEPRITE_PATH` | found automatically | Path to the Aseprite executable. |
| `--workspace` | `ASEPRITE_MCP_WORKSPACE` | current directory | Directory that confines scripts and inspected exports. |
| `--timeout` | | `60s` | Maximum time for one Aseprite process. |
| `--max-output-bytes` | | `1048576` | Maximum captured bytes for each stream. |
| `--max-script-bytes` | | `1048576` | Maximum size of the Lua source. |
| `--max-image-bytes` | | `10485760` | Maximum size of an inspected PNG. |

When you do not set the Aseprite path, the server looks on `PATH` and in the standard install locations for each platform, including the common Steam paths.
If the server finds nothing, it stops and tells you to set the path.

### MCP client configuration

Use a stdio server entry like this:

```json
{
  "mcpServers": {
    "aseprite": {
      "command": "aseprite-mcp",
      "args": ["--workspace", "/path/to/project"],
      "env": { "ASEPRITE_PATH": "/Applications/Aseprite.app/Contents/MacOS/aseprite" }
    }
  }
}
```

## The Aseprite Lua Skill

The skill at [`.agents/skills/aseprite-lua`](.agents/skills/aseprite-lua) teaches the create, inspect, and refine workflow.
It also teaches the Aseprite Lua API through references that load only when a task needs them.

A compatible agent finds skills in `.agents/skills/` in a project.
To use the skill in another project or for all projects, copy the directory:

```sh
cp -r .agents/skills/aseprite-lua ~/.agents/skills/aseprite-lua
```

## Security

> [!WARNING]
> `execute_lua` runs arbitrary Lua with full access to Aseprite and the file system.
> Run it only with code that you trust.

The workspace check confines the paths that the tools accept and inspect to one directory.
It does not sandbox the Lua interpreter.
For stronger isolation, run the server inside a container or a restricted account.

## Development

```sh
go test ./...
go vet ./...
```

The tests compile a fake Aseprite executable, so the suite runs on Linux, macOS, and Windows with no external program.
To test against a real installation:

```sh
ASEPRITE_PATH=/path/to/aseprite go test -tags integration ./internal/tools/ -run Smoke -v
```

## License

aseprite-mcp is available under the Apache 2.0 license.
See the [LICENSE](LICENSE) file for more info.
Aseprite is separate software under its own license.

## Contact

Mattt ([@mattt](https://twitter.com/mattt))

[ci]: https://github.com/mattt/aseprite-mcp/actions
[ci badge]: https://github.com/mattt/aseprite-mcp/workflows/CI/badge.svg
[license]: https://www.apache.org/licenses/LICENSE-2.0
[license badge]: https://img.shields.io/badge/license-Apache%202.0-blue.svg?style=flat
[mcp]: https://modelcontextprotocol.io
[aseprite]: https://www.aseprite.org
