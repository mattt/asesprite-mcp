package workspace

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewRejectsMissingDir(t *testing.T) {
	if _, err := New(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("New on missing directory: expected error, got nil")
	}
}

func TestNewRejectsFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := New(file); err == nil {
		t.Error("New on file: expected error, got nil")
	}
}

func TestResolveExistingInside(t *testing.T) {
	dir := t.TempDir()
	ws, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "sub", "art.png")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Relative path.
	got, err := ws.ResolveExisting(filepath.Join("sub", "art.png"))
	if err != nil {
		t.Fatalf("ResolveExisting relative: %v", err)
	}
	want, _ := filepath.EvalSymlinks(target)
	if got != want {
		t.Errorf("relative resolve = %q, want %q", got, want)
	}

	// Absolute path.
	if _, err := ws.ResolveExisting(target); err != nil {
		t.Errorf("ResolveExisting absolute: %v", err)
	}
}

func TestResolveExistingRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "outside.png")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	ws, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.ResolveExisting(filepath.Join("..", "outside.png")); err == nil {
		t.Error("traversal path: expected error, got nil")
	}
}

func TestResolveExistingRejectsPrefixCollision(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "work")
	sibling := filepath.Join(dir, "work-other")
	for _, d := range []string{root, sibling} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(sibling, "art.png")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	ws, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.ResolveExisting(file); err == nil {
		t.Error("prefix-collision sibling: expected error, got nil")
	}
}

func TestResolveExistingRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is restricted on Windows CI")
	}
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(dir, "secret.png")
	if err := os.WriteFile(secret, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.png")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}
	ws, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.ResolveExisting("link.png"); err == nil {
		t.Error("symlink escape: expected error, got nil")
	}
}

func TestResolveExistingRejectsMissing(t *testing.T) {
	ws, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.ResolveExisting("nope.png"); err == nil {
		t.Error("missing file: expected error, got nil")
	}
}
