// Package config gets the server settings from command-line flags,
// environment variables, and defaults, in that order of precedence.
package config

import (
	"flag"
	"fmt"
	"io"
	"time"
)

// Environment variable names. The server reads these when you do not set a flag.
const (
	EnvAsepritePath = "ASEPRITE_PATH"
	EnvWorkspace    = "ASEPRITE_MCP_WORKSPACE"
)

// Default limits. The server uses these when you do not set a flag.
const (
	DefaultTimeout        = 60 * time.Second
	DefaultMaxOutputBytes = 1 << 20  // 1 MiB for each captured stream
	DefaultMaxScriptBytes = 1 << 20  // 1 MiB of Lua source
	DefaultMaxImageBytes  = 10 << 20 // 10 MiB for each inspected PNG
)

// Config holds the server settings.
type Config struct {
	// AsepritePath is the Aseprite executable. An empty value means the caller
	// must find it.
	AsepritePath string
	// Workspace is the directory that confines scripts and inspected exports.
	// An empty value means the caller uses the current directory.
	Workspace string
	// Timeout limits one Aseprite process.
	Timeout time.Duration
	// MaxOutputBytes limits each captured stream.
	MaxOutputBytes int64
	// MaxScriptBytes limits the Lua source that the server accepts.
	MaxScriptBytes int64
	// MaxImageBytes limits an inspected PNG.
	MaxImageBytes int64
}

// Load reads the settings from args and the environment lookup. A flag takes
// precedence over an environment variable. An environment variable takes
// precedence over a default. The caller finds a missing Aseprite path and sets
// a default workspace, so this function stays deterministic and easy to test.
func Load(name string, args []string, getenv func(string) string, out io.Writer) (*Config, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(out)

	aseprite := fs.String("aseprite", "", "path to the Aseprite executable (env "+EnvAsepritePath+")")
	ws := fs.String("workspace", "", "directory that confines scripts and exports (env "+EnvWorkspace+"); defaults to the current directory")
	timeout := fs.Duration("timeout", DefaultTimeout, "maximum duration for one Aseprite batch process")
	maxOutput := fs.Int64("max-output-bytes", DefaultMaxOutputBytes, "maximum captured bytes per stream")
	maxScript := fs.Int64("max-script-bytes", DefaultMaxScriptBytes, "maximum accepted Lua source size in bytes")
	maxImage := fs.Int64("max-image-bytes", DefaultMaxImageBytes, "maximum inspected PNG size in bytes")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if fs.NArg() > 0 {
		return nil, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	cfg := &Config{
		AsepritePath:   firstNonEmpty(*aseprite, getenv(EnvAsepritePath)),
		Workspace:      firstNonEmpty(*ws, getenv(EnvWorkspace)),
		Timeout:        *timeout,
		MaxOutputBytes: *maxOutput,
		MaxScriptBytes: *maxScript,
		MaxImageBytes:  *maxImage,
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %s", c.Timeout)
	}
	if c.MaxOutputBytes <= 0 {
		return fmt.Errorf("max-output-bytes must be positive, got %d", c.MaxOutputBytes)
	}
	if c.MaxScriptBytes <= 0 {
		return fmt.Errorf("max-script-bytes must be positive, got %d", c.MaxScriptBytes)
	}
	if c.MaxImageBytes <= 0 {
		return fmt.Errorf("max-image-bytes must be positive, got %d", c.MaxImageBytes)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
