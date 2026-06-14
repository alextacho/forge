package version

import (
	"strings"
	"testing"
)

func TestStringUsesStampedValues(t *testing.T) {
	originalVersion, originalCommit, originalDate := Version, Commit, Date
	t.Cleanup(func() {
		Version, Commit, Date = originalVersion, originalCommit, originalDate
	})

	Version = "v1.2.3"
	Commit = "abc123"
	Date = "2026-06-14T00:00:00Z"

	got := String()
	want := "forge v1.2.3 (abc123, 2026-06-14T00:00:00Z)"
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestStringReturnsForgeVersion(t *testing.T) {
	got := String()
	if !strings.HasPrefix(got, "forge ") {
		t.Fatalf("String() = %q", got)
	}
	if strings.Contains(got, "(,") || strings.Contains(got, ", )") {
		t.Fatalf("String() has empty metadata fields: %q", got)
	}
}
