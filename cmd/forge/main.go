package main

import (
	"os"

	"github.com/alextacho/forge/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], app.Environment{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Getwd:  os.Getwd,
	}))
}
