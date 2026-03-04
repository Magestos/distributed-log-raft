package config

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strings"
)

func NormalizeHostPort(address string) (string, error) {
	address = strings.TrimSpace(address)

	addrPort, err := netip.ParseAddrPort(address)

	if err != nil {
		return "", err
	}
	if addrPort.Port() == 0 {
		return "", errors.New("port out of range")
	}

	return addrPort.String(), nil
}

func ValidateDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory %s not found", path)
	}
	return nil
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
		return fmt.Errorf("invalid address or port value on %s", name)
	}
	return nil
}

func validatePeers(peers []string, raftAddress string) error {
	if len(peers) < 1 {
		return errors.New("peers cannot be less than 1")
	}

	normalizedRaftAddress, err := NormalizeHostPort(raftAddress)
	if err != nil {
		return fmt.Errorf("invalid raft address %q: %w", raftAddress, err)
	}

	seen := make(map[string]string, len(peers))
	var duplicates []string
	hasRaftAddress := false

	for _, peer := range peers {
		normalizedPeer, err := NormalizeHostPort(peer)
		if err != nil {
			return fmt.Errorf("invalid peer address %q: %w", peer, err)
		}

		if _, ok := seen[normalizedPeer]; ok {
			duplicates = append(duplicates, peer)
			continue
		}
		seen[normalizedPeer] = peer

		if normalizedPeer == normalizedRaftAddress {
			hasRaftAddress = true
		}
	}

	if !hasRaftAddress {
		return errors.New("peers must contain raft address")
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("peers contain duplicates: %v", duplicates)
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

	if heartbeat < 0 {
		return errors.New("heartbeat cannot be less than 0")
	}

	if heartbeat > electionMin {
		return errors.New("heartbeat cannot be greater than ElectionMinMS")
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

	if err := ValidateDirectory(conf.DataDir); err != nil {
		return err
	}

	if err := validateAddress("client address", conf.ClientAddress); err != nil {
		return err
	}

	if err := validateAddress("raft address", conf.RaftAddress); err != nil {
		return err
	}

	if err := validatePeers(conf.Peers, conf.RaftAddress); err != nil {
		return err
	}

	if err := validateTimeouts(conf.ElectionMinMS, conf.ElectionMaxMS, conf.HeartbeatMS); err != nil {
		return err
	}

	return nil
}
