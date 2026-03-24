package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func PrepareDataDir(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("data dir is required")
	}

	path = filepath.Clean(path)

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

func normalizeDataDir(path, nodeID string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	path = strings.ReplaceAll(path, "${node_id}", nodeID)
	return filepath.Clean(path)
}

func normalizePeers(peers []Peer) ([]Peer, error) {
	if len(peers) == 0 {
		return peers, nil
	}

	normalized := make([]Peer, len(peers))
	for i, peer := range peers {
		peer.NodeID = strings.TrimSpace(peer.NodeID)

		value, err := NormalizeHostPort(peer.RaftAddress)
		if err != nil {
			return nil, fmt.Errorf("normalize peer %q raft address %q: %w", peer.NodeID, peer.RaftAddress, err)
		}
		peer.RaftAddress = value
		normalized[i] = peer
	}

	return normalized, nil
}

func normalizeConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be empty")
	}

	cfg.NodeID = strings.TrimSpace(cfg.NodeID)
	cfg.DataDir = normalizeDataDir(cfg.DataDir, cfg.NodeID)

	if strings.TrimSpace(cfg.ClientAddress) != "" {
		value, err := NormalizeHostPort(cfg.ClientAddress)
		if err != nil {
			return fmt.Errorf("normalize client address %q: %w", cfg.ClientAddress, err)
		}
		cfg.ClientAddress = value
	}

	peers, err := normalizePeers(cfg.Peers)
	if err != nil {
		return err
	}
	cfg.Peers = peers

	return nil
}

func Load(path string) (*Config, error) {
	var cfg Config

	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("config path is required")
	}

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(yamlFile))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config file %q: %w", path, err)
	}

	if err := normalizeConfig(&cfg); err != nil {
		return nil, fmt.Errorf("normalize config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config file %q: %w", path, err)
	}

	if err := PrepareDataDir(cfg.DataDir); err != nil {
		return nil, fmt.Errorf("prepare data dir %q: %w", cfg.DataDir, err)
	}

	return &cfg, nil
}
