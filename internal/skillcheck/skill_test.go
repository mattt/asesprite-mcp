// Package skillcheck validates the Aseprite Lua skill against the Agent Skills
// format. It checks the frontmatter fields, the match between the name and the
// directory, the required references, the one-level relative links, and the
// absence of absolute local paths.
package skillcheck

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func skillDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	// The path is internal/skillcheck/skill_test.go. The repo root is three
	// directories up.
	root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	return filepath.Join(root, ".agents", "skills", "aseprite-lua")
}

func readSkill(t *testing.T) (dir, body string, front map[string]string) {
	t.Helper()
	dir = skillDir(t)
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	front, body = parseFrontmatter(t, string(data))
	return dir, body, front
}

func parseFrontmatter(t *testing.T, content string) (map[string]string, string) {
	t.Helper()
	if !strings.HasPrefix(content, "---\n") {
		t.Fatal("SKILL.md must start with YAML frontmatter")
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		t.Fatal("SKILL.md frontmatter is not terminated")
	}
	block := content[4 : 4+end]
	rest := content[4+end:]
	if i := strings.Index(rest, "\n"); i >= 0 {
		rest = rest[i+1:]
	}
	fields := map[string]string{}
	for _, line := range strings.Split(block, "\n") {
		if line == "" || strings.HasPrefix(line, " ") {
			continue
		}
		if k, v, ok := strings.Cut(line, ":"); ok {
			fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return fields, rest
}

var nameRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func TestFrontmatterFields(t *testing.T) {
	dir, _, front := readSkill(t)

	name := front["name"]
	if !nameRe.MatchString(name) || len(name) > 64 {
		t.Errorf("name %q is not a valid skill name", name)
	}
	if got := filepath.Base(dir); got != name {
		t.Errorf("name %q must match directory %q", name, got)
	}
	desc := front["description"]
	if desc == "" || len(desc) > 1024 {
		t.Errorf("description length %d is out of range", len(desc))
	}
	for _, kw := range []string{"execute_lua", "inspect_export", "pixel art"} {
		if !strings.Contains(desc, kw) {
			t.Errorf("description should mention %q for discovery", kw)
		}
	}
	if front["license"] != "Apache-2.0" {
		t.Errorf("license = %q, want Apache-2.0", front["license"])
	}
}

func TestRequiredReferencesExist(t *testing.T) {
	dir, body, _ := readSkill(t)
	required := []string{"core-api.md", "drawing.md", "animation.md", "export.md", "headless.md"}
	for _, name := range required {
		rel := "references/" + name
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("missing reference %s: %v", rel, err)
		}
		if !strings.Contains(body, rel) {
			t.Errorf("SKILL.md does not link to %s", rel)
		}
	}
}

var linkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

func TestRelativeLinksResolveOneLevel(t *testing.T) {
	dir := skillDir(t)
	mdFiles := collectMarkdown(t, dir)
	for _, path := range mdFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range linkRe.FindAllStringSubmatch(string(data), -1) {
			target := m[1]
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "#") {
				continue
			}
			if strings.Contains(target, "..") {
				t.Errorf("%s links to %q, which is not one level deep", rel(dir, path), target)
			}
			resolved := filepath.Join(filepath.Dir(path), filepath.FromSlash(target))
			if _, err := os.Stat(resolved); err != nil {
				t.Errorf("%s links to missing file %q", rel(dir, path), target)
			}
		}
	}
}

func TestNoAbsoluteLocalPaths(t *testing.T) {
	dir := skillDir(t)
	banned := []string{"/Users/", "/home/mattt", `C:\Users`}
	for _, path := range collectMarkdown(t, dir) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, b := range banned {
			if strings.Contains(text, b) {
				t.Errorf("%s contains stale local path %q", rel(dir, path), b)
			}
		}
	}
}

func collectMarkdown(t *testing.T, dir string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".md") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func rel(base, path string) string {
	r, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return r
}
