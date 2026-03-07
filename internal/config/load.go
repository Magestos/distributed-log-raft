package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func PrepareDataDir(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("data dir is required")
	}

	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("data dir %q is not a directory", path)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("stat data dir %q: %w", path, err)
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create data dir %q: %w", path, err)
	}

	return nil
}

func Load(path string) (*Config, error) {
	var cfg Config

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config file %q: %w", path, err)
	}

	cfg.DataDir = strings.ReplaceAll(cfg.DataDir, "${node_id}", cfg.NodeID)

	if err := PrepareDataDir(cfg.DataDir); err != nil {
		return nil, fmt.Errorf("prepare data dir %q: %w", cfg.DataDir, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config file %q: %w", path, err)
	}

	return &cfg, nil
}
