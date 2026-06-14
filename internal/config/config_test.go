package config

import (
	"strings"
	"testing"
)

func TestDecodeValidConfig(t *testing.T) {
	cfg, err := Decode(strings.NewReader("paths:\n  - workspace/\n  - config/generated.yaml\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"workspace", "config/generated.yaml"}
	for i := range want {
		if cfg.Paths[i] != want[i] {
			t.Fatalf("path %d = %q, want %q", i, cfg.Paths[i], want[i])
		}
	}
}

func TestDecodeRejectsInvalidConfig(t *testing.T) {
	tests := map[string]string{
		"empty":        "paths: []\n",
		"unknown":      "paths: [workspace]\nextra: true\n",
		"absolute":     "paths: [/tmp/data]\n",
		"windows":      "paths: ['C:\\data']\n",
		"traversal":    "paths: [../data]\n",
		"inner walk":   "paths: [data/../other]\n",
		"root":         "paths: [.]\n",
		"reserved":     "paths: [.forge/templates]\n",
		"duplicate":    "paths: [workspace, workspace/]\n",
		"case dup":     "paths: [workspace, WORKSPACE]\n",
		"overlap":      "paths: [workspace, workspace/cache]\n",
		"case overlap": "paths: [WORKSPACE, workspace/cache]\n",
		"empty entry":  "paths: ['']\n",
		"multi doc":    "paths: [workspace]\n---\npaths: [other]\n",
		"case reserve": "paths: [.FORGE/templates]\n",
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode(strings.NewReader(input)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
