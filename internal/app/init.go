package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alextacho/forge/internal/config"
	"github.com/alextacho/forge/internal/project"
)

func runInit(root string, opts options, env Environment) error {
	body, paths, err := initConfig(opts.paths)
	if err != nil {
		return err
	}
	forgeDirectory := filepath.Join(root, ".forge")
	configPath := filepath.Join(root, project.ConfigPath)
	templatesPath := filepath.Join(forgeDirectory, "templates")

	if _, err := os.Lstat(configPath); err == nil {
		return fmt.Errorf("%s already exists", project.ConfigPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect %s: %w", project.ConfigPath, err)
	}
	if err := os.MkdirAll(templatesPath, 0o755); err != nil {
		return fmt.Errorf("create .forge/templates: %w", err)
	}
	if err := os.WriteFile(configPath, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", project.ConfigPath, err)
	}
	updatedIgnore, err := ensureSnapshotsIgnored(filepath.Join(root, ".gitignore"))
	if err != nil {
		return err
	}

	fmt.Fprintf(env.Stdout, "Initialized Forge config at %s\n", project.ConfigPath)
	fmt.Fprintf(env.Stdout, "Managed paths: %d\n", len(paths))
	for _, path := range paths {
		fmt.Fprintf(env.Stdout, "  - %s\n", path)
	}
	if updatedIgnore {
		fmt.Fprintln(env.Stdout, "Updated .gitignore with .forge/snapshots/")
	}
	return nil
}

func initConfig(paths []string) ([]byte, []string, error) {
	var body bytes.Buffer
	body.WriteString("paths:\n")
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			return nil, nil, errors.New("init paths must not be empty")
		}
		body.WriteString("  - ")
		body.WriteString(path)
		body.WriteByte('\n')
	}
	cfg, err := config.Decode(bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, nil, err
	}

	body.Reset()
	body.WriteString("paths:\n")
	for _, path := range cfg.Paths {
		body.WriteString("  - ")
		body.WriteString(path)
		body.WriteByte('\n')
	}
	return body.Bytes(), cfg.Paths, nil
}

func ensureSnapshotsIgnored(path string) (bool, error) {
	const ignoreLine = ".forge/snapshots/"
	content, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("read .gitignore: %w", err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == ignoreLine {
			return false, nil
		}
	}

	var next bytes.Buffer
	next.Write(content)
	if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
		next.WriteByte('\n')
	}
	next.WriteString(ignoreLine)
	next.WriteByte('\n')
	if err := os.WriteFile(path, next.Bytes(), 0o644); err != nil {
		return false, fmt.Errorf("write .gitignore: %w", err)
	}
	return true, nil
}
