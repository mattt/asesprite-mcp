# Core API

Summaries of the objects that most scripts use.
Full docs: <https://www.aseprite.org/api/>.

## Sprite

The document.
It holds layers, frames, cels, palettes, and tags.

```lua
local s = Sprite(64, 48, ColorMode.RGB)   -- width, height, color mode
local s = Sprite{ fromFile = path }        -- open an existing file
```

Common members: `s.width`, `s.height`, `s.colorMode`, `s.layers`, `s.frames`, `s.cels`, `s.palettes`, `s.tags`, and `s.filename`.
Resize the canvas with `s:resize(w, h)`.
Crop the document with `s:crop(x, y, w, h)`.

Color modes: `ColorMode.RGB`, `ColorMode.GRAYSCALE`, and `ColorMode.INDEXED`.

## Layer

```lua
local layer = s:newLayer()      -- add a normal layer on top
layer.name = "outline"
layer.opacity = 200             -- 0..255
layer.isVisible = true
s:deleteLayer(layer)
```

Group layers exist.
A layer's `stackIndex` sets its order.
The first layer of a new sprite is `s.layers[1]`.

## Cel

A cel is the image of one layer on one frame.

```lua
local cel = s:newCel(layer, frame)          -- empty cel
local cel = s:newCel(layer, frame, image, Point(x, y))
cel.position = Point(4, 2)                   -- top-left of the image
cel.image = image
```

Get an existing cel with `layer:cel(frame)`, or step through `s.cels`.

## Image

The pixels of one cel.
Coordinates are local to the image, not to the sprite.

```lua
local img = Image(16, 16, ColorMode.RGB)
img:drawPixel(x, y, color)                   -- integer pixel value or Color
local px = img:getPixel(x, y)
img:clear()                                   -- fill with transparent/0
img:resize(w, h)
```

Convert between a `Color` and a raw pixel with `app.pixelColor.rgba(r,g,b,a)`, `app.pixelColor.rgbaR(px)`, and the other component functions.
For shapes and brushes, see [drawing.md](drawing.md).

## Palette and Color

```lua
local pal = s.palettes[1]
pal:resize(8)
pal:setColor(0, Color{ r = 0, g = 0, b = 0, a = 0 })
pal:setColor(1, Color{ r = 34, g = 32, b = 52 })
local c = pal:getColor(1)
```

`Color` accepts `{ r, g, b, a }`, `{ h, s, v }`, `{ gray }`, or a palette index `{ index = n }`.
An indexed sprite draws with palette indices; see [drawing.md](drawing.md).

## Transactions and undo

Group edits so a failure rolls back cleanly and the undo history stays clear.

```lua
app.transaction("build hero", function()
  -- create layers, cels, and pixels here
end)
```

## Script parameters

Each `--script-param k=v` value arrives in `app.params`.
The server always passes the workspace root:

```lua
local dir = app.params["aseprite_mcp_workspace"]
```

Join each output path to `dir`.
Do not write outside it.
