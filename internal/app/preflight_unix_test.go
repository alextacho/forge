//go:build !windows

package app

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestLoadPreflightsAllCurrentPathsBeforeRemoval(t *testing.T) {
	root := newProject(t, []string{"first.txt", "second"})
	if err := os.WriteFile(filepath.Join(root, "first.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "second"), []byte("saved"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, stderr, code := runForTest(t, root, []string{"save", "--yes"})
	if code != 0 {
		t.Fatalf("save code=%d stderr=%s", code, stderr)
	}
	if err := os.Remove(filepath.Join(root, "second")); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(root, "second"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, code = runForTest(t, root, []string{"load", "--yes"})
	if code == 0 {
		t.Fatal("expected preflight failure")
	}
	content, err := os.ReadFile(filepath.Join(root, "first.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "keep" {
		t.Fatalf("first path changed: %q", content)
	}
}

func TestResetPreflightsUnreadableTemplatesBeforeRemoval(t *testing.T) {
	root := newProject(t, []string{"first.txt", "second.txt"})
	if err := os.WriteFile(filepath.Join(root, "first.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	templateRoot := filepath.Join(root, ".forge", "templates")
	if err := os.MkdirAll(templateRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	unreadable := filepath.Join(templateRoot, "second.txt")
	if err := os.WriteFile(unreadable, []byte("template"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(unreadable, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o600) })

	var stdout, stderr bytes.Buffer
	code := Run([]string{"reset", "--yes"}, Environment{
		Stdin:      strings.NewReader(""),
		Stdout:     &stdout,
		Stderr:     &stderr,
		Getwd:      func() (string, error) { return root, nil },
		IsTerminal: func(io.Reader) bool { return false },
	})
	if code == 0 {
		t.Fatalf("expected preflight failure, stdout=%q", stdout.String())
	}
	content, err := os.ReadFile(filepath.Join(root, "first.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "keep" {
		t.Fatalf("first path changed: %q", content)
	}
}
