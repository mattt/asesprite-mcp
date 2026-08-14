# Export

Save the editable source first.
Then export views and sheets.
Full docs: <https://www.aseprite.org/api/sprite> and <https://www.aseprite.org/api/command/ExportSpriteSheet>.

## Save and export

- `sprite:saveAs(path)` writes the file and marks the sprite as saved.
  Use it for the `.aseprite` source.
- `sprite:saveCopyAs(path)` writes a copy and keeps the saved state.
  Use it for a flat PNG that you give to `inspect_export`.

```lua
local dir = app.params["aseprite_mcp_workspace"]
sprite:saveAs(dir .. "/hero.aseprite")
sprite:saveCopyAs(dir .. "/hero.png")
```

The file format follows the extension: `.aseprite`, `.png`, `.gif`, and others.

## Scaled preview for inspection

Pixel art images are small.
Export a larger PNG so that you can see it, and keep the real sprite at its native size.

```lua
local preview = Sprite(sprite)          -- clone
preview:resize(sprite.width * 8, sprite.height * 8)
preview:saveCopyAs(dir .. "/hero_8x.png")
preview:close()
```

Then inspect `hero_8x.png`.
A resize of a clone uses nearest-neighbor and keeps hard pixel edges.

## Sprite sheet with metadata

`ExportSpriteSheet` writes a packed PNG and a JSON file.
The JSON file describes the frames and tags.
A game-engine importer reads this JSON.

```lua
app.command.ExportSpriteSheet{
  ui = false,
  type = SpriteSheetType.HORIZONTAL,       -- or PACKED, ROWS, COLUMNS
  textureFilename = dir .. "/coin_sheet.png",
  dataFilename = dir .. "/coin_sheet.json",
  dataFormat = SpriteSheetDataFormat.JSON_ARRAY,
  listTags = true,
  listLayers = false,
  listSlices = false,
}
```

Set `ui = false` so that no dialog opens in batch mode.
`JSON_ARRAY` gives an ordered frame list.
`JSON_HASH` keys the frames by name.
Include the tags so that the animation names survive the export.

## Godot handoff

The PNG and JSON pair go into the Godot Aseprite Wizard (<https://github.com/viniciusgerevini/godot-aseprite-wizard>).
The wizard maps each tag to an animation and each frame duration to a time.
Name each tag to match the animation that the game expects.
Export the sheet at the native pixel size, with no upscale.
Keep the larger PNG only for your own inspection.
