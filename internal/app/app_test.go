package app

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCommands(t *testing.T) {
	tests := []struct {
		args    []string
		command string
		name    string
		yes     bool
	}{
		{[]string{"save"}, "save", "default", false},
		{[]string{"save", "--yes", "baseline"}, "save", "baseline", true},
		{[]string{"save", "baseline", "--yes"}, "save", "baseline", true},
		{[]string{"load", "default"}, "load", "default", false},
		{[]string{"reset", "--yes"}, "reset", "", true},
		{[]string{"version"}, "version", "", false},
		{[]string{"--version"}, "version", "", false},
		{[]string{"instructions"}, "instructions", "", false},
		{[]string{"status"}, "status", "", false},
		{[]string{"mcp"}, "mcp", "", false},
	}
	for _, test := range tests {
		got, err := parse(test.args)
		if err != nil {
			t.Fatalf("%v: %v", test.args, err)
		}
		if got.command != test.command || got.name != test.name || got.yes != test.yes {
			t.Fatalf("%v: got %+v", test.args, got)
		}
	}
}

func TestParseRejectsInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		{"unknown"},
		{"save", "one", "two"},
		{"load", "--no-clobber"},
		{"reset", "name"},
		{"version", "extra"},
		{"instructions", "extra"},
		{"status", "--yes"},
		{"mcp", "extra"},
	} {
		if _, err := parse(args); err == nil {
			t.Fatalf("%v: expected error", args)
		}
	}
}

func TestHelpAndInstructionsDoNotRequireProject(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"version"}, {"instructions"}} {
		var stdout, stderr bytes.Buffer
		code := Run(args, Environment{
			Stdin:  strings.NewReader(""),
			Stdout: &stdout,
			Stderr: &stderr,
			Getwd: func() (string, error) {
				return "", os.ErrNotExist
			},
		})
		if code != 0 {
			t.Fatalf("%v: code=%d stderr=%s", args, code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "forge") && !strings.Contains(stdout.String(), "Forge") {
			t.Fatalf("%v: stdout=%q", args, stdout.String())
		}
	}
}

func TestStatus(t *testing.T) {
	root := newProject(t, []string{"workspace", "generated.txt"})
	if err := os.MkdirAll(filepath.Join(root, "workspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workspace", "state.txt"), []byte("state"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".forge", "templates", "workspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	runForTest(t, root, []string{"save", "baseline", "--yes"})

	stdout, stderr, code := runForTest(t, root, []string{"status"})
	if code != 0 {
		t.Fatalf("status code=%d stderr=%s", code, stderr)
	}
	for _, want := range []string{"Managed paths: 2", "workspace (directory)", "generated.txt (absent)", "baseline", "Templates: 1"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("status missing %q in:\n%s", want, stdout)
		}
	}
}

func TestSaveLoadAndReset(t *testing.T) {
	root := newProject(t, []string{"workspace", "generated.txt"})
	if err := os.MkdirAll(filepath.Join(root, "workspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workspace", "saved.txt"), []byte("saved"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runForTest(t, root, []string{"save", "--yes"})
	if code != 0 {
		t.Fatalf("save code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `Saved snapshot "default"`) {
		t.Fatalf("stdout = %q", stdout)
	}

	if err := os.WriteFile(filepath.Join(root, "workspace", "extra.txt"), []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "generated.txt"), []byte("later"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, stderr, code = runForTest(t, root, []string{"load", "default", "--yes"})
	if code != 0 {
		t.Fatalf("load code=%d stderr=%s", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(root, "workspace", "extra.txt")); !os.IsNotExist(err) {
		t.Fatalf("extra file still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "generated.txt")); !os.IsNotExist(err) {
		t.Fatalf("absent file still exists: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(root, ".forge", "templates", "workspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".forge", "templates", "workspace", "template.txt"), []byte("template"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, stderr, code = runForTest(t, root, []string{"reset", "--yes"})
	if code != 0 {
		t.Fatalf("reset code=%d stderr=%s", code, stderr)
	}
	content, err := os.ReadFile(filepath.Join(root, "workspace", "template.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "template" {
		t.Fatalf("content = %q", content)
	}
}

func TestConfirmationRequiredAndDecline(t *testing.T) {
	root := newProject(t, []string{"state.txt"})
	if err := os.WriteFile(filepath.Join(root, "state.txt"), []byte("state"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"save"}, Environment{
		Stdin:      strings.NewReader("y\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		Getwd:      func() (string, error) { return root, nil },
		IsTerminal: func(io.Reader) bool { return false },
	})
	if code == 0 || !strings.Contains(stderr.String(), "use --yes") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".forge", "snapshots", "default")); !os.IsNotExist(err) {
		t.Fatalf("snapshot created without terminal confirmation: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"save"}, Environment{
		Stdin:      strings.NewReader("no\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		Getwd:      func() (string, error) { return root, nil },
		IsTerminal: func(io.Reader) bool { return true },
	})
	if code == 0 || !strings.Contains(stderr.String(), "operation cancelled") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".forge", "snapshots", "default")); !os.IsNotExist(err) {
		t.Fatalf("snapshot created after decline: %v", err)
	}
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

func runForTest(t *testing.T, root string, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, Environment{
		Stdin:      strings.NewReader(""),
		Stdout:     &stdout,
		Stderr:     &stderr,
		Getwd:      func() (string, error) { return root, nil },
		IsTerminal: func(io.Reader) bool { return false },
	})
	return stdout.String(), stderr.String(), code
}
