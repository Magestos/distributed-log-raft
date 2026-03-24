package config

type Peer struct {
	NodeID      string `yaml:"node_id"`
	RaftAddress string `yaml:"raft_address"`
}

type Config struct {
	NodeID        string `yaml:"node_id"`
	ClientAddress string `yaml:"client_address"`
	Peers         []Peer `yaml:"peers"`
	DataDir       string `yaml:"data_dir"`

	ElectionMinMS int `yaml:"election_min_ms"`
	ElectionMaxMS int `yaml:"election_max_ms"`
	HeartbeatMS   int `yaml:"heartbeat_ms"`
}

func (conf *Config) SelfPeer() (Peer, bool) {
	if conf == nil {
		return Peer{}, false
	}

	for _, peer := range conf.Peers {
		if peer.NodeID == conf.NodeID {
			return peer, true
		}
	}

	return Peer{}, false
}

func (conf *Config) RaftAddress() (string, bool) {
	peer, ok := conf.SelfPeer()
	if !ok {
		return "", false
	}

	return peer.RaftAddress, true
}
