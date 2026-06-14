package app

import (
	"fmt"

	"github.com/alextacho/forge/internal/project"
	"github.com/alextacho/forge/internal/snapshot"
)

func runSave(found project.Project, opts options, env Environment) error {
	plan, err := snapshot.PlanSave(found.Root, opts.name, found.Config.Paths, opts.noClobber)
	if err != nil {
		return err
	}
	action := "create"
	if plan.Exists {
		action = "replace"
	}
	fmt.Fprintf(env.Stdout, "Save snapshot %q: %s, %d present, %d absent.\n", plan.Name, action, plan.Present, plan.Absent)
	if err := approve(opts, env); err != nil {
		return err
	}
	result, err := snapshot.Save(plan)
	if err != nil {
		return err
	}
	fmt.Fprintf(env.Stdout, "Saved snapshot %q: %d saved, %d absent", plan.Name, result.Saved, result.Absent)
	if result.Replaced {
		fmt.Fprint(env.Stdout, ", replaced previous snapshot")
	}
	fmt.Fprintln(env.Stdout, ".")
	return nil
}
