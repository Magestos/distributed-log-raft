# distributed-log-raft

Distributed Log Raft is an experimental distributed event log implemented in Go. The project aims to build a fault-tolerant, strongly consistent append-only log based on the Raft consensus algorithm.

Multiple nodes form a cluster that elects a leader, replicates log entries across a quorum, and guarantees a consistent ordering of committed records even in the presence of node failures. The system is designed to explore the internal mechanics of consensus, replication, write-ahead logging, and recovery, while maintaining a clear and minimal architecture suitable for experimentation and performance evaluation.

Build: `go build ./...`  
Test: `go test ./...`
