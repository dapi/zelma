package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvironmentSmokeDiagnosticsE2E(t *testing.T) {
	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	repoRoot := t.TempDir()
	runCommand(t, repoRoot, "git", "init", "--quiet")

	first := runZelma(t, bin, repoRoot, nil, "setup", "--json")
	if first.code != 0 {
		t.Fatalf("first setup code = %d, want 0; stderr = %q", first.code, first.stderr)
	}
	firstSetup := decodeSetupResult(t, first.stdout)
	if !firstSetup.Changed || !firstSetup.GitignoreChanged || !firstSetup.ZelmaDirCreated {
		t.Fatalf("first setup = %+v, want all changed flags", firstSetup)
	}
	assertDir(t, filepath.Join(repoRoot, ".zelma"))
	assertOneZelmaGitignoreEntry(t, filepath.Join(repoRoot, ".gitignore"))

	second := runZelma(t, bin, repoRoot, nil, "setup", "--json")
	if second.code != 0 {
		t.Fatalf("second setup code = %d, want 0; stderr = %q", second.code, second.stderr)
	}
	secondSetup := decodeSetupResult(t, second.stdout)
	if secondSetup.Changed || secondSetup.GitignoreChanged || secondSetup.ZelmaDirCreated {
		t.Fatalf("second setup = %+v, want idempotent unchanged flags", secondSetup)
	}
	assertOneZelmaGitignoreEntry(t, filepath.Join(repoRoot, ".gitignore"))

	list := runZelma(t, bin, repoRoot, nil, "sessions", "list", "--json")
	if list.code != 0 {
		t.Fatalf("list code = %d, want 0; stderr = %q", list.code, list.stderr)
	}
	if strings.TrimSpace(list.stderr) != "" {
		t.Fatalf("list stderr = %q, want empty", list.stderr)
	}
	if strings.TrimSpace(list.stdout) != `{
  "version": 1,
  "sessions": []
}` {
		t.Fatalf("list stdout = %q, want empty schema v1 registry JSON", list.stdout)
	}

	missingZellij := filepath.Join(t.TempDir(), "missing-zellij")
	detect := runZelma(t, bin, repoRoot, []string{"ZELMA_ZELLIJ_BIN=" + missingZellij}, "sessions", "detect", "--json")
	if detect.code != 1 {
		t.Fatalf("detect code = %d, want 1", detect.code)
	}
	if strings.TrimSpace(detect.stdout) != "" {
		t.Fatalf("detect stdout = %q, want empty on diagnostic failure", detect.stdout)
	}
	for _, want := range []string{
		"zelma sessions detect:",
		"zellij_missing_binary",
		"install zellij or configure the adapter binary path",
	} {
		if !strings.Contains(detect.stderr, want) {
			t.Fatalf("detect stderr = %q, want substring %q", detect.stderr, want)
		}
	}
}

type setupResult struct {
	GitignorePath    string `json:"gitignore_path"`
	ZelmaDirPath     string `json:"zelma_dir_path"`
	Changed          bool   `json:"changed"`
	GitignoreChanged bool   `json:"gitignore_changed"`
	ZelmaDirCreated  bool   `json:"zelma_dir_created"`
}

type commandResult struct {
	code   int
	stdout string
	stderr string
}

func projectRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func runZelma(t *testing.T, bin, dir string, env []string, args ...string) commandResult {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("%s %s: %v", bin, strings.Join(args, " "), err)
		}
		code = exitErr.ExitCode()
	}
	return commandResult{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func runCommand(t *testing.T, dir, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %s: %v\nstdout:\n%s\nstderr:\n%s", name, strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
}

func decodeSetupResult(t *testing.T, data string) setupResult {
	t.Helper()

	var result setupResult
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("decode setup JSON: %v; data = %q", err, data)
	}
	if result.GitignorePath == "" || result.ZelmaDirPath == "" {
		t.Fatalf("setup result = %+v, want stable paths", result)
	}
	return result
}

func assertDir(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}

func assertOneZelmaGitignoreEntry(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == ".zelma" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("%s has %d .zelma entries, want 1; content = %q", path, count, string(data))
	}
}
