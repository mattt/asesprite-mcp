# Headless limits

Scripts run through `aseprite -b --script`, with no GUI.
Full CLI docs: <https://www.aseprite.org/docs/cli/>.

## The GUI is absent

- `app.isUIAvailable` is false.
  `Dialog`, `app.alert`, and any tool that opens a window do nothing useful and can stop the process.
  Do not use them.
- Report status with a print to stdout.
  `execute_lua` returns stdout and stderr, so a final `print` is your feedback channel.

## No text API

The Lua API has no glyph or font drawing.
To place text, draw the pixels directly.
You can also make a text image outside Aseprite and import it with `Image{ fromFile = path }`, then copy it onto a cel.

## Some tools behave differently

`app.useTool` depends on editor state that is thinner in batch mode.
If a shape or fill does not appear as you expect, use direct `image:drawPixel` drawing, which is deterministic.
See [drawing.md](drawing.md).

## Errors and diagnostics

- A Lua error causes a non-zero exit.
  `execute_lua` returns the error as a tool error with the message in stderr.
  Read it and correct the script.
- Use `pcall` when you want to print a clean message before you exit:

```lua
local ok, err = pcall(function()
  -- risky work
end)
if not ok then
  io.stderr:write("failed: " .. tostring(err) .. "\n")
  os.exit(1)
end
```

- Save in steps.
  If a long script fails late, an earlier `saveAs` still leaves a file that you can inspect.

## Version notes

- This project needs Aseprite 1.3 or later.
- Some batch-mode behavior depends on the build.
  If the output looks wrong, note the Aseprite version (the server logs it at startup).
  Then prefer direct-pixel drawing and an explicit `saveCopyAs` over a tool-dependent path.
- Check any command flag against the official CLI and API docs for the installed version.
