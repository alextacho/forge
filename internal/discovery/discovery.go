package discovery

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/alextacho/forge/internal/project"
	"github.com/alextacho/forge/internal/snapshot"
)

type ManagedPath struct {
	Path  string `json:"path"`
	State string `json:"state"`
}

type Status struct {
	ProjectRoot string        `json:"projectRoot"`
	ConfigPath  string        `json:"configPath"`
	Paths       []ManagedPath `json:"paths"`
	Snapshots   []string      `json:"snapshots"`
	Templates   []string      `json:"templates"`
}

func Inspect(found project.Project) (Status, error) {
	status := Status{
		ProjectRoot: found.Root,
		ConfigPath:  filepath.Join(found.Root, project.ConfigPath),
	}
	for _, managedPath := range found.Config.Paths {
		state, err := inspectManagedPath(found.Root, managedPath)
		if err != nil {
			return Status{}, err
		}
		status.Paths = append(status.Paths, ManagedPath{Path: managedPath, State: state})
	}
	snapshots, err := ListSnapshots(found.Root)
	if err != nil {
		return Status{}, err
	}
	status.Snapshots = snapshots
	templates, err := ListConfiguredTemplates(found.Root, found.Config.Paths)
	if err != nil {
		return Status{}, err
	}
	status.Templates = templates
	return status, nil
}

func FormatStatus(status Status) string {
	var out bytes.Buffer
	fmt.Fprintf(&out, "Project: %s\n", status.ProjectRoot)
	fmt.Fprintf(&out, "Config: %s\n", status.ConfigPath)
	fmt.Fprintf(&out, "Managed paths: %d\n", len(status.Paths))
	for _, managedPath := range status.Paths {
		fmt.Fprintf(&out, "  - %s (%s)\n", managedPath.Path, managedPath.State)
	}
	fmt.Fprintf(&out, "Snapshots: %d\n", len(status.Snapshots))
	for _, name := range status.Snapshots {
		fmt.Fprintf(&out, "  - %s\n", name)
	}
	fmt.Fprintf(&out, "Templates: %d configured paths present\n", len(status.Templates))
	for _, name := range status.Templates {
		fmt.Fprintf(&out, "  - %s\n", name)
	}
	return out.String()
}

func ListSnapshots(projectRoot string) ([]string, error) {
	directory := filepath.Join(projectRoot, ".forge", "snapshots")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read snapshots directory: %w", err)
	}
	var snapshots []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if err := snapshot.ValidateName(name); err != nil {
			continue
		}
		snapshots = append(snapshots, name)
	}
	sort.Strings(snapshots)
	return snapshots, nil
}

func ListConfiguredTemplates(projectRoot string, paths []string) ([]string, error) {
	var templates []string
	for _, managedPath := range paths {
		templatePath := filepath.Join(projectRoot, ".forge", "templates", filepath.FromSlash(managedPath))
		if _, err := os.Lstat(templatePath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("inspect template %s: %w", managedPath, err)
		}
		templates = append(templates, managedPath)
	}
	return templates, nil
}

func inspectManagedPath(projectRoot, managedPath string) (string, error) {
	info, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(managedPath)))
	if err != nil {
		if os.IsNotExist(err) {
			return "absent", nil
		}
		return "", fmt.Errorf("inspect %s: %w", managedPath, err)
	}
	if info.IsDir() {
		return "directory", nil
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "symlink", nil
	}
	return "file", nil
}
