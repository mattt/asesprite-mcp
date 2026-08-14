# Animation

A frame holds time.
A cel holds the image for a layer on a frame.
A tag names a range of frames.
Full docs: <https://www.aseprite.org/api/frame> and <https://www.aseprite.org/api/tag>.

## Frames

```lua
local s = Sprite(32, 32)
local f2 = s:newEmptyFrame()      -- append an empty frame
local f3 = s:newFrame(f2)         -- copy f2's content into a new frame
s.frames[1].duration = 0.1        -- seconds (100 ms)
```

`#s.frames` is the frame count.
Set the time of each frame with `frame.duration` in seconds.

## Cels across frames

Each visible layer needs a cel on each frame that you want it to appear on.

```lua
for i = 1, #s.frames do
  local cel = s:newCel(s.layers[1], s.frames[i])
  cel.image:drawPixel(i - 1, 0, Color{ r = 255, g = 255, b = 255 })
end
```

## Tags

A tag marks a named animation, such as "walk" or "idle", and its play direction.

```lua
local tag = s:newTag(1, 4)                 -- frames 1..4
tag.name = "walk"
tag.aniDir = AniDir.FORWARD                -- FORWARD, REVERSE, PING_PONG
```

A tag becomes an animation name on export and on import into a tool such as the Godot Aseprite Wizard.
Name each tag the way the game expects.

## A four-frame cycle

```lua
local dir = app.params["aseprite_mcp_workspace"]
local s = Sprite(16, 16, ColorMode.RGB)
for i = 2, 4 do s:newEmptyFrame() end
for i = 1, 4 do
  s.frames[i].duration = 0.12
  local cel = s.cels[i] or s:newCel(s.layers[1], s.frames[i])
  cel.image:drawPixel(i - 1, 8, Color{ r = 240, g = 220, b = 60 })
end
local walk = s:newTag(1, 4); walk.name = "walk"
s:saveAs(dir .. "/coin.aseprite")
```

Export the cycle as a sprite sheet with frame times and tags in JSON.
See [export.md](export.md).
