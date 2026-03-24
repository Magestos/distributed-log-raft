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
