package config

import (
	"io"
	"testing"
	"time"
)

func emptyEnv(string) string { return "" }

func envMap(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("test", nil, emptyEnv, io.Discard)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AsepritePath != "" {
		t.Errorf("AsepritePath = %q, want empty", cfg.AsepritePath)
	}
	if cfg.Workspace != "" {
		t.Errorf("Workspace = %q, want empty", cfg.Workspace)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %s, want %s", cfg.Timeout, DefaultTimeout)
	}
	if cfg.MaxOutputBytes != DefaultMaxOutputBytes {
		t.Errorf("MaxOutputBytes = %d, want %d", cfg.MaxOutputBytes, DefaultMaxOutputBytes)
	}
	if cfg.MaxScriptBytes != DefaultMaxScriptBytes {
		t.Errorf("MaxScriptBytes = %d, want %d", cfg.MaxScriptBytes, DefaultMaxScriptBytes)
	}
	if cfg.MaxImageBytes != DefaultMaxImageBytes {
		t.Errorf("MaxImageBytes = %d, want %d", cfg.MaxImageBytes, DefaultMaxImageBytes)
	}
}

func TestLoadFlagsOverrideEnv(t *testing.T) {
	env := envMap(map[string]string{
		EnvAsepritePath: "/env/aseprite",
		EnvWorkspace:    "/env/workspace",
	})
	args := []string{
		"-aseprite", "/flag/aseprite",
		"-workspace", "/flag/workspace",
		"-timeout", "30s",
		"-max-output-bytes", "2048",
		"-max-script-bytes", "4096",
		"-max-image-bytes", "8192",
	}
	cfg, err := Load("test", args, env, io.Discard)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AsepritePath != "/flag/aseprite" {
		t.Errorf("AsepritePath = %q, want /flag/aseprite", cfg.AsepritePath)
	}
	if cfg.Workspace != "/flag/workspace" {
		t.Errorf("Workspace = %q, want /flag/workspace", cfg.Workspace)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %s, want 30s", cfg.Timeout)
	}
	if cfg.MaxOutputBytes != 2048 || cfg.MaxScriptBytes != 4096 || cfg.MaxImageBytes != 8192 {
		t.Errorf("limits = %d/%d/%d, want 2048/4096/8192", cfg.MaxOutputBytes, cfg.MaxScriptBytes, cfg.MaxImageBytes)
	}
}

func TestLoadEnvFallback(t *testing.T) {
	env := envMap(map[string]string{
		EnvAsepritePath: "/env/aseprite",
		EnvWorkspace:    "/env/workspace",
	})
	cfg, err := Load("test", nil, env, io.Discard)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AsepritePath != "/env/aseprite" {
		t.Errorf("AsepritePath = %q, want /env/aseprite", cfg.AsepritePath)
	}
	if cfg.Workspace != "/env/workspace" {
		t.Errorf("Workspace = %q, want /env/workspace", cfg.Workspace)
	}
}

func TestLoadInvalidLimits(t *testing.T) {
	cases := [][]string{
		{"-timeout", "0"},
		{"-timeout", "-5s"},
		{"-max-output-bytes", "0"},
		{"-max-script-bytes", "-1"},
		{"-max-image-bytes", "0"},
	}
	for _, args := range cases {
		if _, err := Load("test", args, emptyEnv, io.Discard); err == nil {
			t.Errorf("Load(%v): expected error, got nil", args)
		}
	}
}

func TestLoadUnexpectedArgs(t *testing.T) {
	if _, err := Load("test", []string{"stray"}, emptyEnv, io.Discard); err == nil {
		t.Error("Load with positional arg: expected error, got nil")
	}
}

func TestLoadUnknownFlag(t *testing.T) {
	if _, err := Load("test", []string{"-nope"}, emptyEnv, io.Discard); err == nil {
		t.Error("Load with unknown flag: expected error, got nil")
	}
}
