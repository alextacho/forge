package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/alextacho/forge/internal/tree"
)

const manifestVersion = 1

type Manifest struct {
	Version int     `json:"version"`
	Name    string  `json:"name"`
	Entries []Entry `json:"entries"`
}

type Entry struct {
	Path    string      `json:"path"`
	Present bool        `json:"present"`
	Kind    tree.Kind   `json:"kind,omitempty"`
	Nodes   []tree.Node `json:"nodes,omitempty"`
}

func readManifest(root *os.Root) (Manifest, error) {
	data, err := root.ReadFile("manifest.json")
	if err != nil {
		return Manifest{}, fmt.Errorf("read snapshot manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode snapshot manifest: %w", err)
	}
	return manifest, nil
}

func writeManifest(root *os.Root, manifest Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode snapshot manifest: %w", err)
	}
	data = append(data, '\n')
	if err := root.WriteFile("manifest.json", data, 0o644); err != nil {
		return fmt.Errorf("write snapshot manifest: %w", err)
	}
	return nil
}

func validateManifest(manifest Manifest, name string, configuredPaths []string, root *os.Root) error {
	if manifest.Version != manifestVersion {
		return fmt.Errorf("snapshot format version %d is not supported", manifest.Version)
	}
	if manifest.Name != name {
		return fmt.Errorf("snapshot manifest name %q does not match %q", manifest.Name, name)
	}
	if len(manifest.Entries) != len(configuredPaths) {
		return fmt.Errorf("snapshot paths do not match current configuration")
	}

	expected := append([]string(nil), configuredPaths...)
	actual := make([]string, 0, len(manifest.Entries))
	seen := make(map[string]struct{}, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		if _, ok := seen[entry.Path]; ok {
			return fmt.Errorf("snapshot contains duplicate path %q", entry.Path)
		}
		seen[entry.Path] = struct{}{}
		actual = append(actual, entry.Path)
		if !entry.Present {
			if (entry.Kind != "" && entry.Kind != tree.KindAbsent) || len(entry.Nodes) != 0 {
				return fmt.Errorf("absent snapshot path %q has kind %q", entry.Path, entry.Kind)
			}
			continue
		}
		if entry.Kind != tree.KindFile && entry.Kind != tree.KindDirectory && entry.Kind != tree.KindSymlink {
			return fmt.Errorf("snapshot path %q has invalid kind %q", entry.Path, entry.Kind)
		}
		inspection, err := tree.ValidateCopySource(root, path.Join("data", entry.Path))
		if err != nil {
			return err
		}
		if inspection.Kind != entry.Kind {
			return fmt.Errorf("snapshot path %q is %q, manifest says %q", entry.Path, inspection.Kind, entry.Kind)
		}
		if err := validateNodes(entry.Path, entry.Nodes, inspection.Nodes); err != nil {
			return err
		}
	}
	sort.Strings(expected)
	sort.Strings(actual)
	for i := range expected {
		if expected[i] != actual[i] {
			return fmt.Errorf("snapshot paths do not match current configuration")
		}
	}
	return nil
}

func validateNodes(entryPath string, expected, actual []tree.Node) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("snapshot path %q metadata does not match stored data", entryPath)
	}
	actualByPath := make(map[string]tree.Node, len(actual))
	for _, node := range actual {
		actualByPath[node.Path] = node
	}
	seen := make(map[string]struct{}, len(expected))
	for _, node := range expected {
		if _, ok := seen[node.Path]; ok {
			return fmt.Errorf("snapshot path %q has duplicate metadata for %q", entryPath, node.Path)
		}
		seen[node.Path] = struct{}{}
		stored, ok := actualByPath[node.Path]
		if !ok || stored.Kind != node.Kind {
			return fmt.Errorf("snapshot path %q metadata does not match stored data at %q", entryPath, node.Path)
		}
		if node.Mode.Perm() != node.Mode {
			return fmt.Errorf("snapshot path %q has invalid mode for %q", entryPath, node.Path)
		}
	}
	return nil
}
