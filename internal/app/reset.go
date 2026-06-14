package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alextacho/forge/internal/project"
	"github.com/alextacho/forge/internal/tree"
)

type resetEntry struct {
	path    string
	present bool
}

func runReset(found project.Project, opts options, env Environment) error {
	templateDirectory := filepath.Join(found.Root, ".forge", "templates")
	if info, err := os.Lstat(templateDirectory); err == nil && !info.IsDir() {
		return fmt.Errorf("templates path is not a directory")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect templates: %w", err)
	}
	templates, err := os.OpenRoot(templateDirectory)
	templatesMissing := false
	if errors.Is(err, os.ErrNotExist) {
		templatesMissing = true
	} else if err != nil {
		return fmt.Errorf("open templates: %w", err)
	} else {
		defer templates.Close()
	}

	entries := make([]resetEntry, 0, len(found.Config.Paths))
	present := 0
	for _, managedPath := range found.Config.Paths {
		entry := resetEntry{path: managedPath}
		if !templatesMissing {
			inspection, err := tree.ValidateCopySource(templates, managedPath)
			if err != nil {
				return err
			}
			entry.present = inspection.Kind != tree.KindAbsent
		}
		if entry.present {
			present++
		}
		entries = append(entries, entry)
	}
	projectRoot, err := os.OpenRoot(found.Root)
	if err != nil {
		return fmt.Errorf("open project root: %w", err)
	}
	defer projectRoot.Close()
	currentPresent := 0
	for _, entry := range entries {
		inspection, err := tree.Inspect(projectRoot, entry.path)
		if err != nil {
			return err
		}
		if inspection.Kind != tree.KindAbsent {
			currentPresent++
		}
	}
	fmt.Fprintf(
		env.Stdout,
		"Reset project: remove %d present paths, restore %d templates, leave %d absent.\n",
		currentPresent,
		present,
		len(entries)-present,
	)
	if err := approve(opts, env); err != nil {
		return err
	}

	removed := 0
	for _, entry := range entries {
		inspection, err := tree.Inspect(projectRoot, entry.path)
		if err != nil {
			return err
		}
		if inspection.Kind != tree.KindAbsent {
			if err := tree.Remove(projectRoot, entry.path); err != nil {
				return fmt.Errorf("%w; managed paths may be partially reset", err)
			}
			removed++
		}
	}
	restored := 0
	for _, entry := range entries {
		if !entry.present {
			continue
		}
		if err := tree.Copy(templates, entry.path, projectRoot, entry.path); err != nil {
			return fmt.Errorf("%w; managed paths may be partially reset", err)
		}
		restored++
	}
	fmt.Fprintf(env.Stdout, "Reset complete: %d removed, %d restored, %d absent.\n", removed, restored, len(entries)-restored)
	return nil
}
