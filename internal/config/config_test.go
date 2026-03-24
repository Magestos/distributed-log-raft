package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validConfig(dataDir string) *Config {
	return &Config{
		NodeID:        "node-1",
		ClientAddress: "127.0.0.1:8080",
		RaftAddress:   "127.0.0.1:8081",
		Peers: []string{
			"127.0.0.1:8081",
			"127.0.0.1:8082",
		},
		DataDir:       dataDir,
		ElectionMinMS: 150,
		ElectionMaxMS: 300,
		HeartbeatMS:   50,
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	return configPath
}

func TestNormalizeHostPort(t *testing.T) {
	t.Run("normalizes trimmed ipv4 address", func(t *testing.T) {
		got, err := NormalizeHostPort(" 127.0.0.1:8080 ")
		if err != nil {
			t.Fatalf("NormalizeHostPort() returned error: %v", err)
		}

		if got != "127.0.0.1:8080" {
			t.Fatalf("NormalizeHostPort() returned %q, want %q", got, "127.0.0.1:8080")
		}
	})

	t.Run("normalizes hostname to lowercase", func(t *testing.T) {
		got, err := NormalizeHostPort(" LOCALHOST:8080 ")
		if err != nil {
			t.Fatalf("NormalizeHostPort() returned error: %v", err)
		}

		if got != "localhost:8080" {
			t.Fatalf("NormalizeHostPort() returned %q, want %q", got, "localhost:8080")
		}
	})

	t.Run("normalizes ipv6 address", func(t *testing.T) {
		got, err := NormalizeHostPort(" [::1]:8080 ")
		if err != nil {
			t.Fatalf("NormalizeHostPort() returned error: %v", err)
		}

		if got != "[::1]:8080" {
			t.Fatalf("NormalizeHostPort() returned %q, want %q", got, "[::1]:8080")
		}
	})

	t.Run("rejects zero port", func(t *testing.T) {
		_, err := NormalizeHostPort("127.0.0.1:0")
		if err == nil || !strings.Contains(err.Error(), "port out of range") {
			t.Fatalf("NormalizeHostPort() error = %v, want port out of range", err)
		}
	})

	t.Run("rejects malformed address", func(t *testing.T) {
		_, err := NormalizeHostPort("127.0.0.1")
		if err == nil {
			t.Fatal("NormalizeHostPort() returned nil error for malformed address")
		}
	})
}

func TestPrepareDataDir(t *testing.T) {
	t.Run("creates missing directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "data")
		if err := PrepareDataDir(path); err != nil {
			t.Fatalf("PrepareDataDir() returned error: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat() returned error: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("PrepareDataDir() created non-directory at %q", path)
		}
	})

	t.Run("accepts existing directory", func(t *testing.T) {
		path := t.TempDir()
		if err := PrepareDataDir(path); err != nil {
			t.Fatalf("PrepareDataDir() returned error: %v", err)
		}
	})

	t.Run("rejects file path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatalf("WriteFile() returned error: %v", err)
		}

		err := PrepareDataDir(path)
		if err == nil || !strings.Contains(err.Error(), "is not a directory") {
			t.Fatalf("PrepareDataDir() error = %v, want not-a-directory error", err)
		}
	})
}

func TestPrepareDataDir(t *testing.T) {
	t.Run("creates missing directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "data")
		if err := PrepareDataDir(path); err != nil {
			t.Fatalf("PrepareDataDir() returned error: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat() returned error: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("PrepareDataDir() created non-directory at %q", path)
		}
	})

	t.Run("accepts existing directory", func(t *testing.T) {
		path := t.TempDir()
		if err := PrepareDataDir(path); err != nil {
			t.Fatalf("PrepareDataDir() returned error: %v", err)
		}
	})

	t.Run("rejects file path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatalf("WriteFile() returned error: %v", err)
		}

		err := PrepareDataDir(path)
		if err == nil || !strings.Contains(err.Error(), "is not a directory") {
			t.Fatalf("PrepareDataDir() error = %v, want not-a-directory error", err)
		}
	})
}

func TestConfigValidate(t *testing.T) {
	testCases := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "valid config",
		},
		{
			name: "missing data dir on disk is allowed",
			mutate: func(cfg *Config) {
				cfg.DataDir = filepath.Join(t.TempDir(), "missing")
			},
		},
		{
			name:    "nil config",
			wantErr: "config cannot be empty",
		},
		{
			name: "missing node id",
			mutate: func(cfg *Config) {
				cfg.NodeID = " "
			},
			wantErr: "node id is required",
		},
		{
			name: "missing data dir",
			mutate: func(cfg *Config) {
				cfg.DataDir = ""
			},
			wantErr: "data dir is required",
		},
		{
			name: "invalid client address",
			mutate: func(cfg *Config) {
				cfg.ClientAddress = "127.0.0.1"
			},
			wantErr: "invalid address or port value on client address",
		},
		{
			name: "invalid raft address",
			mutate: func(cfg *Config) {
				cfg.RaftAddress = "127.0.0.1"
			},
			wantErr: "invalid address or port value on raft address",
		},
		{
			name: "empty peers",
			mutate: func(cfg *Config) {
				cfg.Peers = nil
			},
			wantErr: "peers cannot be less than 1",
		},
		{
			name: "invalid peer",
			mutate: func(cfg *Config) {
				cfg.Peers = []string{"127.0.0.1"}
			},
			wantErr: "invalid peer address",
		},
		{
			name: "missing raft address in peers",
			mutate: func(cfg *Config) {
				cfg.Peers = []string{"127.0.0.1:8082"}
			},
			wantErr: "peers must contain raft address",
		},
		{
			name: "duplicate peers",
			mutate: func(cfg *Config) {
				cfg.Peers = []string{
					"127.0.0.1:8081",
					" 127.0.0.1:8081 ",
				}
			},
			wantErr: "peers contain duplicates",
		},
		{
			name: "election min less than one",
			mutate: func(cfg *Config) {
				cfg.ElectionMinMS = 0
			},
			wantErr: "ElectionMinMS or ElectionMaxMS cannot be less than 1",
		},
		{
			name: "election min greater than max",
			mutate: func(cfg *Config) {
				cfg.ElectionMinMS = 301
			},
			wantErr: "ElectionMinMS cannot be more than ElectionMaxMS",
		},
		{
			name: "heartbeat less than one",
			mutate: func(cfg *Config) {
				cfg.HeartbeatMS = 0
			},
			wantErr: "heartbeat must be greater than 0",
		},
		{
			name: "heartbeat not less than election min",
			mutate: func(cfg *Config) {
				cfg.HeartbeatMS = 150
			},
			wantErr: "heartbeat must be less than ElectionMinMS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantErr == "config cannot be empty" {
				var cfg *Config
				err := cfg.Validate()
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("Validate() error = %v, want substring %q", err, tc.wantErr)
				}
				return
			}

			cfg := validConfig(filepath.Join(t.TempDir(), "data"))
			if tc.mutate != nil {
				tc.mutate(cfg)
			}

			err := cfg.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() returned error: %v", err)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("loads valid config and normalizes fields", func(t *testing.T) {
		baseDir := t.TempDir()
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: \" LOCALHOST:8080 \"",
			"raft_address: \" LOCALHOST:8081 \"",
			"peers:",
			"  - \" localhost:8081 \"",
			"  - \" LOCALHOST:8082 \"",
			"data_dir: \" " + filepath.Join(baseDir, "${node_id}") + " \"",
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
		}, "\n"))

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() returned error: %v", err)
		}

		if cfg.NodeID != "node-1" {
			t.Fatalf("Load() NodeID = %q, want %q", cfg.NodeID, "node-1")
		}
		if cfg.ClientAddress != "localhost:8080" {
			t.Fatalf("Load() ClientAddress = %q, want %q", cfg.ClientAddress, "localhost:8080")
		}
		if cfg.RaftAddress != "localhost:8081" {
			t.Fatalf("Load() RaftAddress = %q, want %q", cfg.RaftAddress, "localhost:8081")
		}

		wantPeers := []string{"localhost:8081", "localhost:8082"}
		for i, peer := range wantPeers {
			if cfg.Peers[i] != peer {
				t.Fatalf("Load() Peers[%d] = %q, want %q", i, cfg.Peers[i], peer)
			}
		}

		wantDataDir := filepath.Join(baseDir, "node-1")
		if cfg.DataDir != wantDataDir {
			t.Fatalf("Load() DataDir = %q, want %q", cfg.DataDir, wantDataDir)
		}

		info, err := os.Stat(cfg.DataDir)
		if err != nil {
			t.Fatalf("Stat() returned error: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("Load() created non-directory data dir at %q", cfg.DataDir)
		}
	})

	t.Run("creates missing data dir only after successful validation", func(t *testing.T) {
		missingDataDir := filepath.Join(t.TempDir(), "new-data-dir")
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"raft_address: 127.0.0.1:8081",
			"peers:",
			"  - 127.0.0.1:8081",
			"  - 127.0.0.1:8082",
			"data_dir: " + missingDataDir,
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
		}, "\n"))

		if _, err := os.Stat(missingDataDir); !os.IsNotExist(err) {
			t.Fatalf("Stat() error = %v, want not-exist before Load()", err)
		}

		_, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() returned error: %v", err)
		}

		info, err := os.Stat(missingDataDir)
		if err != nil {
			t.Fatalf("Stat() returned error: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("Load() created non-directory data dir at %q", missingDataDir)
		}
	})

	t.Run("creates missing data dir", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yml")
		missingDataDir := filepath.Join(t.TempDir(), "new-data-dir")
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"raft_address: 127.0.0.1:8081",
			"peers:",
			"  - 127.0.0.1:8081",
			"  - 127.0.0.1:8082",
			"data_dir: " + missingDataDir,
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
		}, "\n"))

		if _, err := os.Stat(missingDataDir); !os.IsNotExist(err) {
			t.Fatalf("Stat() error = %v, want not-exist before Load()", err)
		}

		_, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() returned error: %v", err)
		}

		info, err := os.Stat(missingDataDir)
		if err != nil {
			t.Fatalf("Stat() returned error: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("Load() created non-directory data dir at %q", missingDataDir)
		}
	})

	t.Run("returns read error", func(t *testing.T) {
		_, err := Load(filepath.Join(t.TempDir(), "missing.yml"))
		if err == nil || !strings.Contains(err.Error(), "read config file") {
			t.Fatalf("Load() error = %v, want read error", err)
		}
	})

	t.Run("returns unmarshal error", func(t *testing.T) {
		configPath := writeConfigFile(t, "node_id: [")

		_, err := Load(configPath)
		if err == nil || !strings.Contains(err.Error(), "unmarshal config file") {
			t.Fatalf("Load() error = %v, want unmarshal error", err)
		}
	})

	t.Run("rejects unknown fields", func(t *testing.T) {
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"raft_address: 127.0.0.1:8081",
			"peers:",
			"  - 127.0.0.1:8081",
			"  - 127.0.0.1:8082",
			"data_dir: " + t.TempDir(),
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
			"unknown_field: value",
		}, "\n"))

		_, err := Load(configPath)
		if err == nil || !strings.Contains(err.Error(), "field unknown_field not found") {
			t.Fatalf("Load() error = %v, want unknown field error", err)
		}
	})

	t.Run("returns validation error without creating data dir", func(t *testing.T) {
		missingDataDir := filepath.Join(t.TempDir(), "new-data-dir")
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"raft_address: 127.0.0.1:8081",
			"peers:",
			"  - 127.0.0.1:8082",
			"data_dir: " + missingDataDir,
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
		}, "\n"))

		_, err := Load(configPath)
		if err == nil || !strings.Contains(err.Error(), "validate config file") {
			t.Fatalf("Load() error = %v, want validation error", err)
		}

		if _, err := os.Stat(missingDataDir); !os.IsNotExist(err) {
			t.Fatalf("Stat() error = %v, want not-exist after failed Load()", err)
		}
	})

	t.Run("returns validation error for missing heartbeat", func(t *testing.T) {
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"raft_address: 127.0.0.1:8081",
			"peers:",
			"  - 127.0.0.1:8081",
			"  - 127.0.0.1:8082",
			"data_dir: " + t.TempDir(),
			"election_min_ms: 150",
			"election_max_ms: 300",
		}, "\n"))

		_, err := Load(configPath)
		if err == nil || !strings.Contains(err.Error(), "heartbeat must be greater than 0") {
			t.Fatalf("Load() error = %v, want missing heartbeat validation error", err)
		}
	})

	t.Run("returns prepare error for file data dir", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "data-file")
		if err := os.WriteFile(filePath, []byte("x"), 0o600); err != nil {
			t.Fatalf("WriteFile() returned error: %v", err)
		}

		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"raft_address: 127.0.0.1:8081",
			"peers:",
			"  - 127.0.0.1:8081",
			"  - 127.0.0.1:8082",
			"data_dir: " + filePath,
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
		}, "\n"))

		_, err := Load(configPath)
		if err == nil || !strings.Contains(err.Error(), "prepare data dir") {
			t.Fatalf("Load() error = %v, want prepare data dir error", err)
		}
	})
}
