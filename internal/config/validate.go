package config

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

func NormalizeHostPort(address string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", errors.New("address is required")
	}

	host, portValue, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}

	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("host is required")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil {
		return "", fmt.Errorf("invalid port %q", portValue)
	}
	if port < 1 || port > 65535 {
		return "", errors.New("port out of range")
	}

	if parsedIP, err := netip.ParseAddr(host); err == nil {
		host = parsedIP.String()
	} else {
		host = strings.ToLower(host)
	}

	return net.JoinHostPort(host, strconv.Itoa(port)), nil
}

func ValidateHostPort(address string) error {
	_, err := NormalizeHostPort(address)
	return err
}

func requireNonEmpty(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func validateAddress(name, value string) error {
	if err := requireNonEmpty(name, value); err != nil {
		return err
	}

	if err := ValidateHostPort(value); err != nil {
		return fmt.Errorf("invalid address or port value on %s: %w", name, err)
	}
	return nil
}

func validatePeers(peers []Peer, selfNodeID string) error {
	if len(peers) < 1 {
		return errors.New("peers cannot be less than 1")
	}
	seenNodeIDs := make(map[string]struct{}, len(peers))
	seenRaftAddresses := make(map[string]struct{}, len(peers))
	hasSelf := false

	for _, peer := range peers {
		if err := requireNonEmpty("peer node id", peer.NodeID); err != nil {
			return err
		}

		if err := validateAddress("peer raft address", peer.RaftAddress); err != nil {
			return err
		}

		if _, ok := seenNodeIDs[peer.NodeID]; ok {
			return fmt.Errorf("peers contain duplicate node id %q", peer.NodeID)
		}
		seenNodeIDs[peer.NodeID] = struct{}{}

		if _, ok := seenRaftAddresses[peer.RaftAddress]; ok {
			return fmt.Errorf("peers contain duplicate raft address %q", peer.RaftAddress)
		}
		seenRaftAddresses[peer.RaftAddress] = struct{}{}

		if peer.NodeID == selfNodeID {
			hasSelf = true
		}
	}

	if !hasSelf {
		return errors.New("peers must contain current node id")
	}

	return nil
}

func validateTimeouts(electionMin, electionMax, heartbeat int) error {
	if electionMin < 1 || electionMax < 1 {
		return errors.New("ElectionMinMS or ElectionMaxMS cannot be less than 1")
	}

	if electionMin > electionMax {
		return errors.New("ElectionMinMS cannot be more than ElectionMaxMS")
	}

	if heartbeat < 1 {
		return errors.New("heartbeat must be greater than 0")
	}

	if heartbeat >= electionMin {
		return errors.New("heartbeat must be less than ElectionMinMS")
	}

	return nil
}

func (conf *Config) Validate() error {
	if conf == nil {
		return errors.New("config cannot be empty")
	}

	if err := requireNonEmpty("node id", conf.NodeID); err != nil {
		return err
	}

	if err := requireNonEmpty("data dir", conf.DataDir); err != nil {
		return err
	}

	if err := validateAddress("client address", conf.ClientAddress); err != nil {
		return err
	}

	if err := validatePeers(conf.Peers, conf.NodeID); err != nil {
		return err
	}

	raftAddress, ok := conf.RaftAddress()
	if !ok {
		return errors.New("peers must contain current node id")
	}
	if conf.ClientAddress == raftAddress {
		return errors.New("client address must differ from current node raft address")
	}

	if err := validateTimeouts(conf.ElectionMinMS, conf.ElectionMaxMS, conf.HeartbeatMS); err != nil {
		return err
	}

	return nil
}
