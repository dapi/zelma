package repo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRootFromNestedDirectory(t *testing.T) {
	root := newGitRepo(t)
	nested := filepath.Join(root, "a", "b", ".")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveRoot(nested)
	if err != nil {
		t.Fatalf("ResolveRoot() error = %v", err)
	}

	want := cleanEval(t, root)
	if got.Path != want {
		t.Fatalf("ResolveRoot() = %q, want %q", got.Path, want)
	}
}

func TestResolveRootFromLinkedWorktreeMarker(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: /tmp/example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveRoot(nested)
	if err != nil {
		t.Fatalf("ResolveRoot() error = %v", err)
	}

	want := cleanEval(t, root)
	if got.Path != want {
		t.Fatalf("ResolveRoot() = %q, want %q", got.Path, want)
	}
}

func TestResolveRootNormalizesFileStart(t *testing.T) {
	root := newGitRepo(t)
	file := filepath.Join(root, "dir", "file.txt")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveRoot(file)
	if err != nil {
		t.Fatalf("ResolveRoot() error = %v", err)
	}

	want := cleanEval(t, root)
	if got.Path != want {
		t.Fatalf("ResolveRoot() = %q, want %q", got.Path, want)
	}
}

func TestResolveRootOutsideRepoReturnsUnsupported(t *testing.T) {
	start := t.TempDir()

	_, err := ResolveRoot(start)
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("ResolveRoot() error = %v, want ErrUnsupported", err)
	}

	diagnostic := Diagnostic("zelma", err)
	for _, want := range []string{"zelma", "unsupported repo", cleanEval(t, start), "Git repository"} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("Diagnostic() = %q, want substring %q", diagnostic, want)
		}
	}
}

func newGitRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func cleanEval(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(resolved)
}
