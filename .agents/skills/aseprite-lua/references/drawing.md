# Drawing

How to place pixels.
Full docs: <https://www.aseprite.org/api/image> and <https://www.aseprite.org/api/app#appusetool>.

## Direct pixels

This is the most predictable method in batch mode.
Draw on the image of a cel.

```lua
local img = sprite.cels[1].image
local c = Color{ r = 220, g = 60, b = 40, a = 255 }
img:drawPixel(3, 5, c)
```

For a tight loop, compute a raw pixel value once and write it directly:

```lua
local px = app.pixelColor.rgba(220, 60, 40, 255)
for x = 0, img.width - 1 do
  img:drawPixel(x, 0, px)
end
```

Read a pixel back with `img:getPixel(x, y)`.
Decode it with `app.pixelColor.rgbaR(px)`, `rgbaG`, `rgbaB`, and `rgbaA`.

## Shapes and lines with tools

`app.useTool` runs a real Aseprite tool.
It draws on the active cel, so set the active layer and frame first.

```lua
app.useTool{
  tool = "line",                 -- also: pencil, filled_rectangle, ellipse, ...
  color = Color{ r = 20, g = 20, b = 20 },
  brush = Brush(1),
  points = { Point(0, 0), Point(15, 15) },
}
```

Common tools: `pencil`, `line`, `rectangle`, `filled_rectangle`, `ellipse`, `filled_ellipse`, `contour`, and `paint_bucket`.
A filled rectangle from one pair of corner points is the fastest way to block in a shape.

Note for batch mode: some tools behave differently with no GUI.
If a shape does not appear, use direct-pixel drawing.
See [headless.md](headless.md).

## Selections and fills

```lua
sprite.selection:select(Rectangle(2, 2, 10, 10))
app.useTool{ tool = "paint_bucket", color = c, points = { Point(3, 3) } }
sprite.selection:deselect()
```

## Indexed color

For a fixed palette, use `ColorMode.INDEXED` and draw palette indices, not RGB values.

```lua
local s = Sprite(32, 32, ColorMode.INDEXED)
local pal = s.palettes[1]
pal:resize(4)
pal:setColor(0, Color{ r = 0, g = 0, b = 0, a = 0 })   -- index 0 = transparent
pal:setColor(1, Color{ r = 34, g = 32, b = 52 })
s.cels[1].image:drawPixel(1, 1, 1)                      -- writes index 1
s.transparentColor = 0
```

## Dither a ramp

This is a simple checkerboard dither between two indices or colors:

```lua
local img = sprite.cels[1].image
for y = 0, img.height - 1 do
  for x = 0, img.width - 1 do
    local a = ((x + y) % 2 == 0)
    img:drawPixel(x, y, a and 1 or 2)   -- indexed; use Color values for RGB
  end
end
```

Aseprite also has a built-in ordered dither.
It runs when the sprite converts from RGB to indexed.
See [export.md](export.md) and the ConvertColorMode command in the official docs.
