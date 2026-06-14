package app

import (
	"fmt"
	"os"

	"github.com/alextacho/forge/internal/project"
	"github.com/alextacho/forge/internal/snapshot"
	"github.com/alextacho/forge/internal/tree"
)

func runLoad(found project.Project, opts options, env Environment) error {
	loaded, err := snapshot.Load(found.Root, opts.name, found.Config.Paths)
	if err != nil {
		return err
	}
	projectRoot, err := os.OpenRoot(found.Root)
	if err != nil {
		return fmt.Errorf("open project root: %w", err)
	}
	defer projectRoot.Close()
	currentPresent := 0
	for _, managedPath := range found.Config.Paths {
		inspection, err := tree.Inspect(projectRoot, managedPath)
		if err != nil {
			return err
		}
		if inspection.Kind != tree.KindAbsent {
			currentPresent++
		}
	}
	fmt.Fprintf(
		env.Stdout,
		"Load snapshot %q: remove %d present paths, restore %d, leave %d absent.\n",
		loaded.Name,
		currentPresent,
		loaded.Present,
		loaded.Absent,
	)
	if err := approve(opts, env); err != nil {
		return err
	}
	result, err := snapshot.Restore(loaded)
	if err != nil {
		return err
	}
	fmt.Fprintf(
		env.Stdout,
		"Loaded snapshot %q: %d removed, %d restored, %d absent.\n",
		loaded.Name,
		result.Removed,
		result.Restored,
		result.Absent,
	)
	return nil
}
