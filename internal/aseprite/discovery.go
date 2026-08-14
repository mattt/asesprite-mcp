// Package aseprite finds a user-installed Aseprite executable and runs Lua
// scripts through its batch mode. This project does not include or download
// Aseprite. The user must supply an installation with a separate license.
package aseprite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Discover finds an Aseprite executable on PATH or in the standard install
// locations for the given GOOS. When it finds nothing, it returns a clear
// error that lists the locations that it searched.
func Discover(goos string) (string, error) {
	names := []string{"aseprite", "Aseprite"}
	for _, name := range names {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	candidates := standardLocations(goos)
	for _, c := range candidates {
		if isExecutableFile(c) {
			return c, nil
		}
	}
	searched := append([]string{"PATH"}, candidates...)
	return "", fmt.Errorf("aseprite: executable not found; searched %s; set --aseprite or the ASEPRITE_PATH environment variable",
		strings.Join(searched, ", "))
}

func standardLocations(goos string) []string {
	switch goos {
	case "darwin":
		locs := []string{
			"/Applications/Aseprite.app/Contents/MacOS/aseprite",
		}
		if home, err := os.UserHomeDir(); err == nil {
			locs = append(locs, filepath.Join(home, "Applications/Aseprite.app/Contents/MacOS/aseprite"))
		}
		locs = append(locs,
			"/Applications/Steam.app/Contents/MacOS/steamapps/common/Aseprite/Aseprite.app/Contents/MacOS/aseprite",
		)
		return locs
	case "windows":
		locs := []string{
			`C:\Program Files\Aseprite\Aseprite.exe`,
			`C:\Program Files (x86)\Steam\steamapps\common\Aseprite\Aseprite.exe`,
		}
		if pf := os.Getenv("ProgramFiles"); pf != "" {
			locs = append(locs, filepath.Join(pf, `Aseprite\Aseprite.exe`))
		}
		return locs
	default: // linux and other unix systems
		locs := []string{
			"/usr/bin/aseprite",
			"/usr/local/bin/aseprite",
			"/var/lib/flatpak/exports/bin/org.aseprite.Aseprite",
		}
		if home, err := os.UserHomeDir(); err == nil {
			locs = append(locs,
				filepath.Join(home, ".steam/steam/steamapps/common/Aseprite/aseprite"),
				filepath.Join(home, ".local/share/Steam/steamapps/common/Aseprite/aseprite"),
			)
		}
		return locs
	}
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

// ProbeVersion runs "aseprite --version" and returns the trimmed output. It is
// a simple health check. The caller can log the error and continue.
func ProbeVersion(ctx context.Context, execPath string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, execPath, "--version")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("aseprite: version probe failed: %w", err)
	}
	return strings.TrimSpace(buf.String()), nil
}
