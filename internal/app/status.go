package app

import (
	"fmt"

	"github.com/alextacho/forge/internal/discovery"
	"github.com/alextacho/forge/internal/project"
)

func runStatus(found project.Project, env Environment) error {
	status, err := discovery.Inspect(found)
	if err != nil {
		return err
	}
	fmt.Fprint(env.Stdout, discovery.FormatStatus(status))
	return nil
}
