package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alextacho/forge/internal/config"
)

const ConfigPath = ".forge/config.yaml"

type Project struct {
	Root   string
	Config config.Config
}

func Find(start string) (Project, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return Project{}, fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		configPath := filepath.Join(current, filepath.FromSlash(ConfigPath))
		forgeDirectory := filepath.Join(current, ".forge")
		forgeInfo, forgeErr := os.Lstat(forgeDirectory)
		if forgeErr == nil && forgeInfo.Mode()&os.ModeSymlink != 0 {
			return Project{}, fmt.Errorf("%s must not be a symlink", forgeDirectory)
		}
		if forgeErr != nil && !errors.Is(forgeErr, os.ErrNotExist) {
			return Project{}, fmt.Errorf("inspect %s: %w", forgeDirectory, forgeErr)
		}
		configInfo, configErr := os.Lstat(configPath)
		if configErr == nil && (configInfo.Mode()&os.ModeSymlink != 0 || !configInfo.Mode().IsRegular()) {
			return Project{}, fmt.Errorf("%s must be a regular file, not a symlink", configPath)
		}
		if configErr != nil && !errors.Is(configErr, os.ErrNotExist) {
			return Project{}, fmt.Errorf("inspect %s: %w", configPath, configErr)
		}
		file, err := os.Open(configPath)
		if err == nil {
			defer file.Close()
			cfg, err := config.Decode(file)
			if err != nil {
				return Project{}, fmt.Errorf("%s: %w", configPath, err)
			}
			return Project{Root: current, Config: cfg}, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return Project{}, fmt.Errorf("open %s: %w", configPath, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return Project{}, errors.New(".forge/config.yaml not found in current directory or any parent")
		}
		current = parent
	}
}
