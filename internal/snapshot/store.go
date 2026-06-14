package snapshot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alextacho/forge/internal/tree"
)

const snapshotsPath = ".forge/snapshots"

type SavePlan struct {
	ProjectRoot string
	Name        string
	Paths       []string
	Entries     []Entry
	Exists      bool
	Present     int
	Absent      int
}

type Loaded struct {
	ProjectRoot string
	Name        string
	Directory   string
	Manifest    Manifest
	Present     int
	Absent      int
}

type Result struct {
	Saved    int
	Removed  int
	Restored int
	Absent   int
	Replaced bool
}

func PlanSave(projectRoot, name string, paths []string, noClobber bool) (SavePlan, error) {
	if err := ValidateName(name); err != nil {
		return SavePlan{}, err
	}
	snapshotsDir := filepath.Join(projectRoot, snapshotsPath)
	if err := validateOptionalDirectory(snapshotsDir, "snapshots directory"); err != nil {
		return SavePlan{}, err
	}
	destination := filepath.Join(snapshotsDir, name)
	info, err := os.Lstat(destination)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return SavePlan{}, fmt.Errorf("inspect snapshot %q: %w", name, err)
	}
	if exists && !info.IsDir() {
		return SavePlan{}, fmt.Errorf("snapshot %q is not a directory", name)
	}
	if exists && noClobber {
		return SavePlan{}, fmt.Errorf("snapshot %q already exists", name)
	}

	project, err := os.OpenRoot(projectRoot)
	if err != nil {
		return SavePlan{}, fmt.Errorf("open project root: %w", err)
	}
	defer project.Close()

	plan := SavePlan{
		ProjectRoot: projectRoot,
		Name:        name,
		Paths:       append([]string(nil), paths...),
		Exists:      exists,
	}
	for _, managedPath := range paths {
		inspection, err := tree.ValidateCopySource(project, managedPath)
		if err != nil {
			return SavePlan{}, err
		}
		entry := Entry{
			Path:    managedPath,
			Present: inspection.Kind != tree.KindAbsent,
			Kind:    inspection.Kind,
			Nodes:   inspection.Nodes,
		}
		if entry.Present {
			plan.Present++
		} else {
			entry.Kind = ""
			entry.Nodes = nil
			plan.Absent++
		}
		plan.Entries = append(plan.Entries, entry)
	}
	return plan, nil
}

func Save(plan SavePlan) (Result, error) {
	snapshotsDir := filepath.Join(plan.ProjectRoot, snapshotsPath)
	if err := os.MkdirAll(snapshotsDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create snapshots directory: %w", err)
	}
	if err := validateDirectory(snapshotsDir, "snapshots directory"); err != nil {
		return Result{}, err
	}
	stageDir, err := os.MkdirTemp(snapshotsDir, ".tmp-"+plan.Name+"-")
	if err != nil {
		return Result{}, fmt.Errorf("create snapshot staging directory: %w", err)
	}
	stagePublished := false
	defer func() {
		if !stagePublished {
			_ = os.RemoveAll(stageDir)
		}
	}()

	stage, err := os.OpenRoot(stageDir)
	if err != nil {
		return Result{}, fmt.Errorf("open snapshot staging directory: %w", err)
	}
	defer stage.Close()
	if err := stage.Mkdir("data", 0o755); err != nil {
		return Result{}, fmt.Errorf("create snapshot data directory: %w", err)
	}
	project, err := os.OpenRoot(plan.ProjectRoot)
	if err != nil {
		return Result{}, fmt.Errorf("open project root: %w", err)
	}
	defer project.Close()
	for _, entry := range plan.Entries {
		if !entry.Present {
			continue
		}
		if err := tree.CopyForStorage(project, entry.Path, stage, filepath.ToSlash(filepath.Join("data", entry.Path))); err != nil {
			return Result{}, err
		}
	}
	manifest := Manifest{Version: manifestVersion, Name: plan.Name, Entries: plan.Entries}
	if err := writeManifest(stage, manifest); err != nil {
		return Result{}, err
	}
	if err := validateManifest(manifest, plan.Name, plan.Paths, stage); err != nil {
		return Result{}, fmt.Errorf("validate staged snapshot: %w", err)
	}
	if err := stage.Close(); err != nil {
		return Result{}, fmt.Errorf("close staged snapshot: %w", err)
	}

	destination := filepath.Join(snapshotsDir, plan.Name)
	var backup string
	if plan.Exists {
		backupDirectory, err := os.MkdirTemp(snapshotsDir, ".backup-"+plan.Name+"-")
		if err != nil {
			return Result{}, fmt.Errorf("reserve snapshot backup path: %w", err)
		}
		if err := os.Remove(backupDirectory); err != nil {
			return Result{}, fmt.Errorf("prepare snapshot backup path: %w", err)
		}
		backup = backupDirectory
		if err := os.Rename(destination, backup); err != nil {
			return Result{}, fmt.Errorf("move existing snapshot aside: %w", err)
		}
	}
	if err := os.Rename(stageDir, destination); err != nil {
		if plan.Exists {
			_ = os.Rename(backup, destination)
		}
		return Result{}, fmt.Errorf("publish snapshot: %w", err)
	}
	stagePublished = true
	if plan.Exists {
		if err := os.RemoveAll(backup); err != nil {
			return Result{}, fmt.Errorf("snapshot replaced, but old backup could not be removed: %w", err)
		}
	}
	return Result{Saved: plan.Present, Absent: plan.Absent, Replaced: plan.Exists}, nil
}

func Load(projectRoot, name string, configuredPaths []string) (Loaded, error) {
	if err := ValidateName(name); err != nil {
		return Loaded{}, err
	}
	snapshotsDir := filepath.Join(projectRoot, snapshotsPath)
	if err := validateDirectory(snapshotsDir, "snapshots directory"); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Loaded{}, fmt.Errorf("snapshot %q does not exist", name)
		}
		return Loaded{}, err
	}
	directory := filepath.Join(snapshotsDir, name)
	if err := validateDirectory(directory, fmt.Sprintf("snapshot %q", name)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Loaded{}, fmt.Errorf("snapshot %q does not exist", name)
		}
		return Loaded{}, err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Loaded{}, fmt.Errorf("snapshot %q does not exist", name)
		}
		return Loaded{}, fmt.Errorf("open snapshot %q: %w", name, err)
	}
	defer root.Close()
	manifest, err := readManifest(root)
	if err != nil {
		return Loaded{}, err
	}
	if err := validateManifest(manifest, name, configuredPaths, root); err != nil {
		return Loaded{}, err
	}
	loaded := Loaded{
		ProjectRoot: projectRoot,
		Name:        name,
		Directory:   directory,
		Manifest:    manifest,
	}
	for _, entry := range manifest.Entries {
		if entry.Present {
			loaded.Present++
		} else {
			loaded.Absent++
		}
	}
	return loaded, nil
}

func Restore(loaded Loaded) (Result, error) {
	project, err := os.OpenRoot(loaded.ProjectRoot)
	if err != nil {
		return Result{}, fmt.Errorf("open project root: %w", err)
	}
	defer project.Close()
	source, err := os.OpenRoot(loaded.Directory)
	if err != nil {
		return Result{}, fmt.Errorf("open snapshot %q: %w", loaded.Name, err)
	}
	defer source.Close()

	result := Result{Absent: loaded.Absent}
	for _, entry := range loaded.Manifest.Entries {
		inspection, err := tree.Inspect(project, entry.Path)
		if err != nil {
			return result, err
		}
		if inspection.Kind != tree.KindAbsent {
			if err := tree.Remove(project, entry.Path); err != nil {
				return result, partialError(err)
			}
			result.Removed++
		}
	}
	for _, entry := range loaded.Manifest.Entries {
		if !entry.Present {
			continue
		}
		if err := tree.Copy(source, filepath.ToSlash(filepath.Join("data", entry.Path)), project, entry.Path); err != nil {
			return result, partialError(err)
		}
		if err := tree.ApplyModes(project, entry.Path, entry.Nodes); err != nil {
			return result, partialError(err)
		}
		result.Restored++
	}
	return result, nil
}

func ValidateName(name string) error {
	if name == "" || name == "." || name == ".." {
		return errors.New("snapshot name must not be empty")
	}
	if strings.ContainsAny(name, `/\:`) || filepath.Base(name) != name {
		return fmt.Errorf("snapshot name %q must be a single path segment", name)
	}
	if strings.TrimRight(name, " .") != name {
		return fmt.Errorf("snapshot name %q must not end with a space or dot", name)
	}
	for _, character := range name {
		if character < 0x20 || character == 0x7f {
			return fmt.Errorf("snapshot name %q contains control characters", name)
		}
	}
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, ".tmp-") || strings.HasPrefix(lower, ".backup-") {
		return fmt.Errorf("snapshot name %q uses a reserved prefix", name)
	}
	upper := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	switch upper {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return fmt.Errorf("snapshot name %q is reserved on Windows", name)
	}
	return nil
}

func partialError(err error) error {
	return fmt.Errorf("%w; managed paths may be partially restored", err)
}

func validateDirectory(name, label string) error {
	info, err := os.Lstat(name)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", label)
	}
	return nil
}

func validateOptionalDirectory(name, label string) error {
	err := validateDirectory(name, label)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
