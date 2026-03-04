package config

type Config struct {
	NodeID        string   `yaml:"node_id"`
	ClientAddress string   `yaml:"client_address"`
	RaftAddress   string   `yaml:"raft_address"`
	Peers         []string `yaml:"peers"`
	DataDir       string   `yaml:"data_dir"`

	ElectionMinMS int `yaml:"election_min_ms"`
	ElectionMaxMS int `yaml:"election_max_ms"`
	HeartbeatMS   int `yaml:"heartbeat_ms"`
}
