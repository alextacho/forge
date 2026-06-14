package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binary string

func TestMain(m *testing.M) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		panic(err)
	}
	temp, err := os.MkdirTemp("", "forge-integration-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(temp)
	binary = filepath.Join(temp, "forge")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	command := exec.Command("go", "build", "-o", binary, "./cmd/forge")
	command.Dir = repoRoot
	if output, err := command.CombinedOutput(); err != nil {
		panic(string(output))
	}
	code := m.Run()
	_ = os.RemoveAll(temp)
	os.Exit(code)
}

func TestDefaultAndNamedSnapshots(t *testing.T) {
	root := newProject(t, []string{"workspace", "generated.txt"})
	nested := filepath.Join(root, "nested", "work")
	if err := os.MkdirAll(filepath.Join(root, "workspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workspace", "state.txt"), []byte("default"), 0o644); err != nil {
		t.Fatal(err)
	}

	run(t, nested, true, "save", "--yes")
	if err := os.WriteFile(filepath.Join(root, "workspace", "state.txt"), []byte("named"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, nested, true, "save", "baseline", "--yes")

	if err := os.WriteFile(filepath.Join(root, "workspace", "extra.txt"), []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, nested, true, "load", "--yes")
	assertContent(t, filepath.Join(root, "workspace", "state.txt"), "default")
	if _, err := os.Stat(filepath.Join(root, "workspace", "extra.txt")); !os.IsNotExist(err) {
		t.Fatalf("extra file survived exact load: %v", err)
	}

	run(t, nested, true, "load", "baseline", "--yes")
	assertContent(t, filepath.Join(root, "workspace", "state.txt"), "named")
}

func TestOverwriteNoClobberAndNonInteractiveApproval(t *testing.T) {
	root := newProject(t, []string{"state.txt"})
	if err := os.WriteFile(filepath.Join(root, "state.txt"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, root, true, "save", "baseline", "--yes")
	if err := os.WriteFile(filepath.Join(root, "state.txt"), []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, root, true, "save", "baseline", "--yes")
	output := run(t, root, false, "save", "baseline", "--no-clobber", "--yes")
	if !strings.Contains(output, "already exists") {
		t.Fatalf("output = %q", output)
	}

	if err := os.RemoveAll(filepath.Join(root, ".forge", "snapshots", "default")); err != nil {
		t.Fatal(err)
	}
	output = run(t, root, false, "save")
	if !strings.Contains(output, "use --yes") {
		t.Fatalf("output = %q", output)
	}
	if _, err := os.Stat(filepath.Join(root, ".forge", "snapshots", "default")); !os.IsNotExist(err) {
		t.Fatalf("snapshot created without approval: %v", err)
	}
}

func TestResetFromTemplates(t *testing.T) {
	root := newProject(t, []string{"workspace", "generated.txt"})
	if err := os.MkdirAll(filepath.Join(root, "workspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workspace", "stale.txt"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "generated.txt"), []byte("remove"), 0o644); err != nil {
		t.Fatal(err)
	}
	template := filepath.Join(root, ".forge", "templates", "workspace")
	if err := os.MkdirAll(template, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(template, "clean.txt"), []byte("clean"), 0o644); err != nil {
		t.Fatal(err)
	}

	run(t, root, true, "reset", "--yes")
	assertContent(t, filepath.Join(root, "workspace", "clean.txt"), "clean")
	if _, err := os.Stat(filepath.Join(root, "workspace", "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale file survived reset: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "generated.txt")); !os.IsNotExist(err) {
		t.Fatalf("untemplated path survived reset: %v", err)
	}
}

func TestUnsafeConfigDoesNotMutate(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".forge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".forge", "config.yaml"), []byte("paths: [../outside]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	output := run(t, root, false, "reset", "--yes")
	if !strings.Contains(output, "must not contain .. traversal") {
		t.Fatalf("output = %q", output)
	}
}

func TestExternalSymlinkTargetIsUntouched(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require privileges on Windows")
	}
	root := newProject(t, []string{"link"})
	external := filepath.Join(t.TempDir(), "external.txt")
	if err := os.WriteFile(external, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	run(t, root, true, "save", "--yes")
	if err := os.Remove(filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	run(t, root, true, "load", "--yes")
	target, err := os.Readlink(filepath.Join(root, "link"))
	if err != nil {
		t.Fatal(err)
	}
	if target != external {
		t.Fatalf("target = %q, want %q", target, external)
	}
	assertContent(t, external, "outside")
}

func newProject(t *testing.T, paths []string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".forge"), 0o755); err != nil {
		t.Fatal(err)
	}
	var config bytes.Buffer
	config.WriteString("paths:\n")
	for _, path := range paths {
		config.WriteString("  - " + path + "\n")
	}
	if err := os.WriteFile(filepath.Join(root, ".forge", "config.yaml"), config.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func run(t *testing.T, directory string, success bool, args ...string) string {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Dir = directory
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	if success && err != nil {
		t.Fatalf("forge %v failed: %v\n%s", args, err, output.String())
	}
	if !success && err == nil {
		t.Fatalf("forge %v unexpectedly succeeded\n%s", args, output.String())
	}
	return output.String()
}

func assertContent(t *testing.T, name, want string) {
	t.Helper()
	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != want {
		t.Fatalf("%s = %q, want %q", name, content, want)
	}
}
