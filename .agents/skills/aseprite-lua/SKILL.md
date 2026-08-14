---
name: aseprite-lua
description: Create, edit, animate, and export pixel art by writing Aseprite Lua scripts and running them through the execute_lua and inspect_export MCP tools. Use when the task involves pixel art, sprites, sprite sheets, tilesets, palettes, animation frames, .aseprite files, or exporting PNGs for a game engine such as Godot or Unity.
license: Apache-2.0
compatibility: Requires the aseprite-mcp server and a separately licensed Aseprite 1.3 or newer installation. Scripts run in Aseprite batch mode with no GUI.
---

# Aseprite Lua

Make pixel art with one Lua program that runs through Aseprite in batch mode.
Write one script that builds the whole sprite, then look at the result.
Do not send many small drawing calls.

## Trust and scope

`execute_lua` runs real Lua on the local machine.
The script has full access to Aseprite and the file system.
Treat it as trusted local code.
Write files only inside the workspace.

## Workflow

Copy this checklist and track your progress:

```
- [ ] 1. Read the reference for the current step (see the routing list below)
- [ ] 2. Write one Lua script that builds and saves the sprite
- [ ] 3. Run the script with execute_lua
- [ ] 4. Export a PNG and look at it with inspect_export
- [ ] 5. Compare the result with the goal and correct the script
```

1. **Read one reference.**
   Load only the reference file that the step needs.
   Do not read all of the files first.
2. **Write one script.**
   Build the sprite, layers, frames, and palette in one program.
   Read the workspace root from `app.params["aseprite_mcp_workspace"]`.
   Join each output path to that root.
3. **Run the script.**
   Give the script as the `script` argument to `execute_lua`.
   A non-zero exit comes back as a tool error with stdout and stderr.
   Read the message and correct the script.
4. **Look at the result.**
   Export a scaled PNG and call `inspect_export` with its path.
   You cannot judge pixel art without looking at it.
5. **Correct the script.**
   Change the script and run it again.
   Keep the `.aseprite` file so later edits do not start from a flat PNG.

## Minimal example

```lua
local dir = app.params["aseprite_mcp_workspace"]
local sprite = Sprite(32, 32, ColorMode.RGB)
local image = sprite.cels[1].image
image:drawPixel(16, 16, Color{ r = 255, g = 80, b = 40, a = 255 })
sprite:saveAs(dir .. "/hero.aseprite")           -- editable source
sprite:saveCopyAs(dir .. "/hero.png")            -- flat export to inspect
print("saved hero.aseprite and hero.png")
```

Then call `inspect_export` with `hero.png`.

## Reference routing

Read the one file that matches the step.
Each file is a short, self-contained summary with links to the official API.

- Sprites, images, layers, cels, palettes, colors, and undo transactions:
  [references/core-api.md](references/core-api.md)
- Pixels and shapes, indexed color, and dithering:
  [references/drawing.md](references/drawing.md)
- Frames, cels across time, tags, and timing for animation:
  [references/animation.md](references/animation.md)
- Saving `.aseprite`, exporting PNG, and sprite sheets with JSON metadata:
  [references/export.md](references/export.md)
- Batch-mode limits, diagnostics, and version notes:
  [references/headless.md](references/headless.md)

## Rules that prevent wasted runs

- Save an editable `.aseprite` file with `Sprite:saveAs`.
  Export a flat copy with `Sprite:saveCopyAs`.
  Do not use the PNG as your working file.
- Put risky edits inside `app.transaction(function() ... end)`.
  A failure then does not leave a half-built sprite.
- Print a short status line at the end.
  `execute_lua` returns stdout, so a final `print` shows what the script did.
- Do not open a dialog and do not call `app.alert`.
  The GUI is absent in batch mode; see [references/headless.md](references/headless.md).
- There is no text API.
  Draw glyphs pixel by pixel, or import a prepared image.

## Official documentation

The API reference is at <https://github.com/aseprite/api> and <https://www.aseprite.org/api/>.
The command-line reference is at <https://www.aseprite.org/docs/cli/>.
Read them when a reference here does not cover an object that you need.
