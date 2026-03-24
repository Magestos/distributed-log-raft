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
		Peers: []Peer{
			{
				NodeID:      "node-1",
				RaftAddress: "127.0.0.1:8081",
			},
			{
				NodeID:      "node-2",
				RaftAddress: "127.0.0.1:8082",
			},
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

func TestConfigSelfPeer(t *testing.T) {
	t.Run("returns current peer and raft address", func(t *testing.T) {
		cfg := validConfig(filepath.Join(t.TempDir(), "data"))

		peer, ok := cfg.SelfPeer()
		if !ok {
			t.Fatal("SelfPeer() did not find current node")
		}

		if peer.NodeID != "node-1" {
			t.Fatalf("SelfPeer() NodeID = %q, want %q", peer.NodeID, "node-1")
		}
		if peer.RaftAddress != "127.0.0.1:8081" {
			t.Fatalf("SelfPeer() RaftAddress = %q, want %q", peer.RaftAddress, "127.0.0.1:8081")
		}

		address, ok := cfg.RaftAddress()
		if !ok {
			t.Fatal("RaftAddress() did not resolve current node address")
		}
		if address != "127.0.0.1:8081" {
			t.Fatalf("RaftAddress() = %q, want %q", address, "127.0.0.1:8081")
		}
	})

	t.Run("returns false when current peer is missing", func(t *testing.T) {
		cfg := validConfig(filepath.Join(t.TempDir(), "data"))
		cfg.NodeID = "node-9"

		if _, ok := cfg.SelfPeer(); ok {
			t.Fatal("SelfPeer() found unexpected current node")
		}

		if _, ok := cfg.RaftAddress(); ok {
			t.Fatal("RaftAddress() resolved unexpected current node address")
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
			name: "empty peers",
			mutate: func(cfg *Config) {
				cfg.Peers = nil
			},
			wantErr: "peers cannot be less than 1",
		},
		{
			name: "missing peer node id",
			mutate: func(cfg *Config) {
				cfg.Peers[0].NodeID = " "
			},
			wantErr: "peer node id is required",
		},
		{
			name: "invalid peer raft address",
			mutate: func(cfg *Config) {
				cfg.Peers[0].RaftAddress = "127.0.0.1"
			},
			wantErr: "invalid address or port value on peer raft address",
		},
		{
			name: "missing current node in peers",
			mutate: func(cfg *Config) {
				cfg.Peers = []Peer{
					{
						NodeID:      "node-2",
						RaftAddress: "127.0.0.1:8082",
					},
				}
			},
			wantErr: "peers must contain current node id",
		},
		{
			name: "duplicate peer node id",
			mutate: func(cfg *Config) {
				cfg.Peers[1].NodeID = "node-1"
			},
			wantErr: "peers contain duplicate node id",
		},
		{
			name: "duplicate peer raft address",
			mutate: func(cfg *Config) {
				cfg.Peers[1].RaftAddress = "127.0.0.1:8081"
			},
			wantErr: "peers contain duplicate raft address",
		},
		{
			name: "client address equals current raft address",
			mutate: func(cfg *Config) {
				cfg.ClientAddress = "127.0.0.1:8081"
			},
			wantErr: "client address must differ from current node raft address",
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
			"peers:",
			"  - node_id: \" node-1 \"",
			"    raft_address: \" LOCALHOST:8081 \"",
			"  - node_id: \" node-2 \"",
			"    raft_address: \" LOCALHOST:8082 \"",
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

		raftAddress, ok := cfg.RaftAddress()
		if !ok {
			t.Fatal("Load() did not resolve current node raft address")
		}
		if raftAddress != "localhost:8081" {
			t.Fatalf("Load() RaftAddress() = %q, want %q", raftAddress, "localhost:8081")
		}

		wantPeers := []Peer{
			{
				NodeID:      "node-1",
				RaftAddress: "localhost:8081",
			},
			{
				NodeID:      "node-2",
				RaftAddress: "localhost:8082",
			},
		}
		for i, peer := range wantPeers {
			if cfg.Peers[i] != peer {
				t.Fatalf("Load() Peers[%d] = %+v, want %+v", i, cfg.Peers[i], peer)
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
			"peers:",
			"  - node_id: node-1",
			"    raft_address: 127.0.0.1:8081",
			"  - node_id: node-2",
			"    raft_address: 127.0.0.1:8082",
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

	t.Run("rejects unknown top-level fields", func(t *testing.T) {
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"peers:",
			"  - node_id: node-1",
			"    raft_address: 127.0.0.1:8081",
			"  - node_id: node-2",
			"    raft_address: 127.0.0.1:8082",
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

	t.Run("rejects unknown nested peer fields", func(t *testing.T) {
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"peers:",
			"  - node_id: node-1",
			"    raft_address: 127.0.0.1:8081",
			"    extra: value",
			"data_dir: " + t.TempDir(),
			"election_min_ms: 150",
			"election_max_ms: 300",
			"heartbeat_ms: 50",
		}, "\n"))

		_, err := Load(configPath)
		if err == nil || !strings.Contains(err.Error(), "field extra not found") {
			t.Fatalf("Load() error = %v, want nested unknown field error", err)
		}
	})

	t.Run("returns validation error without creating data dir", func(t *testing.T) {
		missingDataDir := filepath.Join(t.TempDir(), "new-data-dir")
		configPath := writeConfigFile(t, strings.Join([]string{
			"node_id: node-1",
			"client_address: 127.0.0.1:8080",
			"peers:",
			"  - node_id: node-2",
			"    raft_address: 127.0.0.1:8082",
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
			"peers:",
			"  - node_id: node-1",
			"    raft_address: 127.0.0.1:8081",
			"  - node_id: node-2",
			"    raft_address: 127.0.0.1:8082",
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
			"peers:",
			"  - node_id: node-1",
			"    raft_address: 127.0.0.1:8081",
			"  - node_id: node-2",
			"    raft_address: 127.0.0.1:8082",
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
