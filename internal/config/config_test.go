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

func TestNormalizeHostPort(t *testing.T) {
	t.Run("normalizes trimmed address", func(t *testing.T) {
		got, err := NormalizeHostPort(" 127.0.0.1:8080 ")
		if err != nil {
			t.Fatalf("NormalizeHostPort returned error: %v", err)
		}

		if got != "127.0.0.1:8080" {
			t.Fatalf("NormalizeHostPort returned %q, want %q", got, "127.0.0.1:8080")
		}
	})

	t.Run("rejects zero port", func(t *testing.T) {
		_, err := NormalizeHostPort("127.0.0.1:0")
		if err == nil {
			t.Fatal("NormalizeHostPort returned nil error for zero port")
		}
	})

	t.Run("rejects malformed address", func(t *testing.T) {
		_, err := NormalizeHostPort("127.0.0.1")
		if err == nil {
			t.Fatal("NormalizeHostPort returned nil error for malformed address")
		}
	})
}

func TestValidateDirectory(t *testing.T) {
	if err := ValidateDirectory(t.TempDir()); err != nil {
		t.Fatalf("ValidateDirectory returned error for existing directory: %v", err)
	}

	missingDir := filepath.Join(t.TempDir(), "missing")
	if err := ValidateDirectory(missingDir); err == nil {
		t.Fatal("ValidateDirectory returned nil error for missing directory")
	}
}

func TestConfigValidate(t *testing.T) {
	dataDir := t.TempDir()

	testCases := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "valid config",
		},
		{
			name: "nil config",
			mutate: func(_ *Config) {
			},
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
			name: "missing data dir on disk",
			mutate: func(cfg *Config) {
				cfg.DataDir = filepath.Join(dataDir, "missing")
			},
			wantErr: "directory",
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
					"127.0.0.1:8081",
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
			name: "negative heartbeat",
			mutate: func(cfg *Config) {
				cfg.HeartbeatMS = -1
			},
			wantErr: "heartbeat cannot be less than 0",
		},
		{
			name: "heartbeat greater than election min",
			mutate: func(cfg *Config) {
				cfg.HeartbeatMS = 151
			},
			wantErr: "heartbeat cannot be greater than ElectionMinMS",
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

			cfg := validConfig(dataDir)
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
	dataDir := t.TempDir()

	t.Run("loads valid config", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yml")
		content := strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"raft_address: 127.0.0.1:8081",
			"peers:",
			"  - 127.0.0.1:8081",
			"  - 127.0.0.1:8082",
			"data_dir: " + dataDir,
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
		}, "\n")

		if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() returned error: %v", err)
		}

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() returned error: %v", err)
		}

		if cfg.NodeID != "node-1" {
			t.Fatalf("Load() NodeID = %q, want %q", cfg.NodeID, "node-1")
		}
	})

	t.Run("returns read error", func(t *testing.T) {
		_, err := Load(filepath.Join(t.TempDir(), "missing.yml"))
		if err == nil || !strings.Contains(err.Error(), "read config file") {
			t.Fatalf("Load() error = %v, want read error", err)
		}
	})

	t.Run("returns unmarshal error", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yml")
		if err := os.WriteFile(configPath, []byte("node_id: ["), 0o600); err != nil {
			t.Fatalf("WriteFile() returned error: %v", err)
		}

		_, err := Load(configPath)
		if err == nil || !strings.Contains(err.Error(), "unmarshal config file") {
			t.Fatalf("Load() error = %v, want unmarshal error", err)
		}
	})

	t.Run("returns validation error", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yml")
		content := strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"raft_address: 127.0.0.1:8081",
			"peers:",
			"  - 127.0.0.1:8082",
			"data_dir: " + dataDir,
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
		}, "\n")

		if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() returned error: %v", err)
		}

		_, err := Load(configPath)
		if err == nil || !strings.Contains(err.Error(), "validate config file") {
			t.Fatalf("Load() error = %v, want validation error", err)
		}
	})
}
