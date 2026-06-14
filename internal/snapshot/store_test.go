package snapshot

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveAndRestoreSnapshot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "workspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workspace", "saved.txt"), []byte("saved"), 0o640); err != nil {
		t.Fatal(err)
	}
	paths := []string{"workspace", "missing.txt"}
	plan, err := PlanSave(root, "default", paths, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Save(plan); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "workspace", "extra.txt"), []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "missing.txt"), []byte("later"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(root, "default", paths)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Restore(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if result.Restored != 1 || result.Absent != 1 {
		t.Fatalf("result = %+v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "workspace", "extra.txt")); !os.IsNotExist(err) {
		t.Fatalf("extra file still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "missing.txt")); !os.IsNotExist(err) {
		t.Fatalf("absent file still exists: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, "workspace", "saved.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "saved" {
		t.Fatalf("content = %q", content)
	}
}

func TestSaveAndRestoreNestedModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows permission bits differ")
	}
	root := t.TempDir()
	nested := filepath.Join(root, "workspace", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(nested, "run.sh")
	if err := os.WriteFile(file, []byte("run"), 0o751); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(nested, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(nested, 0o700)
	})
	plan, err := PlanSave(root, "default", []string{"workspace"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Save(plan); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(root, "default", []string{"workspace"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(loaded); err != nil {
		t.Fatal(err)
	}
	directoryInfo, err := os.Stat(nested)
	if err != nil {
		t.Fatal(err)
	}
	if directoryInfo.Mode().Perm() != 0o500 {
		t.Fatalf("directory mode = %o, want 500", directoryInfo.Mode().Perm())
	}
	fileInfo, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Mode().Perm() != 0o751 {
		t.Fatalf("file mode = %o, want 751", fileInfo.Mode().Perm())
	}
}

func TestNoClobberPreservesSnapshot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "state.txt"), []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanSave(root, "baseline", []string{"state.txt"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Save(plan); err != nil {
		t.Fatal(err)
	}
	if _, err := PlanSave(root, "baseline", []string{"state.txt"}, true); err == nil {
		t.Fatal("expected no-clobber error")
	}
}

func TestLoadRejectsConfigMismatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "state.txt"), []byte("state"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanSave(root, "default", []string{"state.txt"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Save(plan); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root, "default", []string{"other.txt"}); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestValidateName(t *testing.T) {
	for _, name := range []string{
		"", ".", "..", "../x", "a/b", "a:b", "trailing.", "trailing ",
		".tmp-x", ".TMP-x", ".backup-x", "NUL", "con.txt", "con.foo.bar", "bad\nname",
	} {
		if err := ValidateName(name); err == nil {
			t.Fatalf("%q: expected error", name)
		}
	}
	if err := ValidateName("default"); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRejectsSymlinkedSnapshotDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require privileges on Windows")
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, snapshotsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, snapshotsPath, "default")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root, "default", []string{"state.txt"}); err == nil {
		t.Fatal("expected symlink error")
	}
}

func TestSaveRejectsSymlinkedSnapshotsRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require privileges on Windows")
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".forge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, snapshotsPath)); err != nil {
		t.Fatal(err)
	}
	if _, err := PlanSave(root, "default", []string{"state.txt"}, false); err == nil {
		t.Fatal("expected symlink error")
	}
}
