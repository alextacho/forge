package config

import (
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Paths []string `yaml:"paths"`
}

func Decode(r io.Reader) (Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Config{}, errors.New("config must contain exactly one YAML document")
		}
		return Config{}, fmt.Errorf("decode trailing config content: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func validate(cfg *Config) error {
	if len(cfg.Paths) == 0 {
		return errors.New("config paths must contain at least one entry")
	}

	seen := make(map[string]struct{}, len(cfg.Paths))
	for i, raw := range cfg.Paths {
		normalized, err := normalizePath(raw)
		if err != nil {
			return fmt.Errorf("config path %q: %w", raw, err)
		}
		cfg.Paths[i] = normalized
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("config path %q is duplicated", normalized)
		}
		seen[key] = struct{}{}
	}

	for i, candidate := range cfg.Paths {
		for j, other := range cfg.Paths {
			if i == j {
				continue
			}
			if strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(other)+"/") {
				return fmt.Errorf("config paths %q and %q overlap", other, candidate)
			}
		}
	}
	return nil
}

func normalizePath(value string) (string, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if value == "" {
		return "", errors.New("path is empty")
	}
	if strings.HasPrefix(value, "/") || hasVolumePrefix(value) {
		return "", errors.New("absolute paths are not allowed")
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == ".." {
			return "", errors.New("path must not contain .. traversal")
		}
	}

	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", errors.New("path must stay inside the project")
	}
	lower := strings.ToLower(cleaned)
	if lower == ".forge" || strings.HasPrefix(lower, ".forge/") {
		return "", errors.New(".forge is reserved for CLI state")
	}
	return cleaned, nil
}

func hasVolumePrefix(value string) bool {
	return len(value) >= 2 &&
		((value[0] >= 'A' && value[0] <= 'Z') || (value[0] >= 'a' && value[0] <= 'z')) &&
		value[1] == ':'
}
