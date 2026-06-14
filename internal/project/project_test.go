package project

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindFromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".forge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ConfigPath), []byte("paths: [workspace]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Find(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got.Root != root {
		t.Fatalf("root = %q, want %q", got.Root, root)
	}
}

func TestFindMissing(t *testing.T) {
	if _, err := Find(t.TempDir()); err == nil {
		t.Fatal("expected error")
	}
}

func TestFindRejectsSymlinkedForgeDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require privileges on Windows")
	}
	root := t.TempDir()
	external := t.TempDir()
	if err := os.WriteFile(filepath.Join(external, "config.yaml"), []byte("paths: [workspace]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root, ".forge")); err != nil {
		t.Fatal(err)
	}
	if _, err := Find(root); err == nil {
		t.Fatal("expected symlink error")
	}
}

func TestFindRejectsSymlinkedConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require privileges on Windows")
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".forge"), 0o755); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(external, []byte("paths: [workspace]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root, ConfigPath)); err != nil {
		t.Fatal(err)
	}
	if _, err := Find(root); err == nil {
		t.Fatal("expected symlink error")
	}
}
