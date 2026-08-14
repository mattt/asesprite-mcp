package aseprite

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDiscoverFindsOnPath(t *testing.T) {
	dir := t.TempDir()
	name := "aseprite"
	if runtime.GOOS == "windows" {
		name = "aseprite.exe"
	}
	exe := filepath.Join(dir, name)
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	got, err := Discover(runtime.GOOS)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if !strings.Contains(got, "aseprite") {
		t.Errorf("Discover = %q, want a path containing aseprite", got)
	}
}

func TestDiscoverReportsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := Discover("plan9-unknown")
	if err == nil {
		t.Fatal("Discover with empty PATH: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "ASEPRITE_PATH") {
		t.Errorf("error %q should mention ASEPRITE_PATH", err)
	}
}

func TestStandardLocationsPerOS(t *testing.T) {
	for _, goos := range []string{"darwin", "windows", "linux"} {
		if len(standardLocations(goos)) == 0 {
			t.Errorf("standardLocations(%q) returned none", goos)
		}
	}
}
