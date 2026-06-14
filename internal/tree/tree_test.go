package tree

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyDirectoryAndPermissions(t *testing.T) {
	sourceDir := t.TempDir()
	destinationDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(sourceDir, "workspace", "nested"), 0o750); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(sourceDir, "workspace", "nested", "run.sh")
	if err := os.WriteFile(file, []byte("#!/bin/sh\n"), 0o751); err != nil {
		t.Fatal(err)
	}
	source := openRoot(t, sourceDir)
	destination := openRoot(t, destinationDir)

	if err := Copy(source, "workspace", destination, "copy"); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(destinationDir, "copy", "nested", "run.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "#!/bin/sh\n" {
		t.Fatalf("content = %q", content)
	}
	info, err := os.Stat(filepath.Join(destinationDir, "copy", "nested", "run.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o751 {
		t.Fatalf("mode = %o, want 751", info.Mode().Perm())
	}
}

func TestCopyAndRemoveSymlinkWithoutFollowingTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require privileges on Windows")
	}
	sourceDir := t.TempDir()
	destinationDir := t.TempDir()
	externalDir := t.TempDir()
	externalFile := filepath.Join(externalDir, "outside.txt")
	if err := os.WriteFile(externalFile, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(externalFile, filepath.Join(sourceDir, "outside-link")); err != nil {
		t.Fatal(err)
	}
	source := openRoot(t, sourceDir)
	destination := openRoot(t, destinationDir)

	if err := Copy(source, "outside-link", destination, "copied-link"); err != nil {
		t.Fatal(err)
	}
	target, err := os.Readlink(filepath.Join(destinationDir, "copied-link"))
	if err != nil {
		t.Fatal(err)
	}
	if target != externalFile {
		t.Fatalf("target = %q, want %q", target, externalFile)
	}
	if err := Remove(destination, "copied-link"); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(externalFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "keep" {
		t.Fatalf("external content = %q", content)
	}
}

func TestInspectAbsentAndRejectsUnsupportedType(t *testing.T) {
	rootDir := t.TempDir()
	root := openRoot(t, rootDir)
	inspection, err := Inspect(root, "missing")
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Kind != KindAbsent {
		t.Fatalf("kind = %q, want absent", inspection.Kind)
	}

	if runtime.GOOS == "windows" {
		return
	}
	pipe := filepath.Join(rootDir, "pipe")
	if err := createFIFO(pipe); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	if _, err := Inspect(root, "pipe"); err == nil {
		t.Fatal("expected unsupported file type error")
	}
}

func TestRootRejectsEscape(t *testing.T) {
	root := openRoot(t, t.TempDir())
	if _, err := Inspect(root, "../outside"); err == nil {
		t.Fatal("expected escape error")
	}
}

func TestInspectRejectsSymlinkedParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require privileges on Windows")
	}
	rootDir := t.TempDir()
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "state.txt"), []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(rootDir, "linked")); err != nil {
		t.Fatal(err)
	}
	root := openRoot(t, rootDir)
	if _, err := Inspect(root, "linked/state.txt"); err == nil {
		t.Fatal("expected symlinked parent error")
	}
}

func TestCopyRestrictiveDirectoryMode(t *testing.T) {
	sourceDir := t.TempDir()
	destinationDir := t.TempDir()
	restricted := filepath.Join(sourceDir, "restricted")
	if err := os.Mkdir(restricted, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(restricted, "state.txt"), []byte("state"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(restricted, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(restricted, 0o700)
		_ = os.Chmod(filepath.Join(destinationDir, "restricted"), 0o700)
	})
	source := openRoot(t, sourceDir)
	destination := openRoot(t, destinationDir)
	if err := Copy(source, "restricted", destination, "restricted"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(destinationDir, "restricted"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o500 {
		t.Fatalf("mode = %o, want 500", info.Mode().Perm())
	}
}

func TestValidateCopySourceRejectsUnreadableFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows permission bits differ")
	}
	rootDir := t.TempDir()
	file := filepath.Join(rootDir, "state.txt")
	if err := os.WriteFile(file, []byte("state"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(file, 0o600) })
	root := openRoot(t, rootDir)
	if _, err := ValidateCopySource(root, "state.txt"); err == nil {
		t.Fatal("expected unreadable source error")
	}
}

func openRoot(t *testing.T, dir string) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return root
}
