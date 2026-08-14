// Package workspace confines file paths to one root directory. It stops path
// traversal and symlink escapes, so scripts and inspected exports stay inside
// the workspace.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// Workspace is a canonical root directory. Each resolved path must stay inside
// it.
type Workspace struct {
	root string
}

// New makes root into an absolute directory with the symlinks resolved. The
// directory must already exist.
func New(root string) (*Workspace, error) {
	if root == "" {
		return nil, fmt.Errorf("workspace: root is empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("workspace: resolve %q: %w", root, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("workspace: %q must be an existing directory: %w", abs, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, fmt.Errorf("workspace: stat %q: %w", resolved, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("workspace: %q is not a directory", resolved)
	}
	return &Workspace{root: resolved}, nil
}

// Root returns the canonical workspace directory.
func (w *Workspace) Root() string { return w.root }

// contains reports whether resolved is the root or is inside it. resolved must
// already be an absolute, clean path.
func (w *Workspace) contains(resolved string) bool {
	if resolved == w.root {
		return true
	}
	rel, err := filepath.Rel(w.root, resolved)
	if err != nil {
		return false
	}
	if filepath.IsAbs(rel) {
		return false
	}
	// A leading ".." segment means resolved is outside the root. This check also
	// rejects a sibling directory that only shares the root as a string prefix
	// (for example /work and /work-other), because filepath.Rel gives
	// "../work-other".
	sep := string(os.PathSeparator)
	return rel != ".." && (len(rel) < 3 || rel[:3] != ".."+sep)
}

// ResolveExisting resolves path against the workspace and confirms that it
// stays inside the root after the symlinks are followed. The path must exist.
// The function joins a relative path to the root and uses an absolute path as
// given.
func (w *Workspace) ResolveExisting(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("workspace: path is empty")
	}
	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(w.root, candidate)
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("workspace: %q must exist inside the workspace: %w", path, err)
	}
	if !w.contains(resolved) {
		return "", fmt.Errorf("workspace: %q resolves to %q, outside workspace %q", path, resolved, w.root)
	}
	return resolved, nil
}
