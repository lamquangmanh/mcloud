package state

import (
	"errors"
	"os"
	"time"

	"mcloud/internal/config"

	"gopkg.in/yaml.v3"
)

type Node struct {
	ID            string    `json:"id"`              // Unique identifier for the node
	Hostname      string    `json:"hostname"`        // Node's hostname
	IP            string    `json:"ip"`              // Node's IP address
	Role          string    `json:"role"`            // Node role (e.g., "leader", "worker")
	Status        string    `json:"status"`          // Current status (e.g., "online", "offline")
	InitializedAt time.Time `json:"initialized_at"`  // Timestamp when node was initialized
}

type Cluster struct {
	ID   string `json:"id"`   // Unique identifier for the cluster
	Name string `json:"name"` // Cluster name
	AdvertiseAddr string `json:"advertise_addr"`  // Address used for cluster communication
}

type Flags struct {
	Initialized bool `json:"initialized"` // Whether the node has been initialized
}

// State represents the persistent state of the node in the cluster.
// It stores essential information about the node's identity, cluster membership, and initialization status.
// This state is persisted to disk as a YAML file for recovery across restarts.
type State struct {
	Version string `yaml:"version"` // State schema version for compatibility tracking

	// Node contains information about this specific node
	Node Node `yaml:"node"`

	// Cluster contains information about the cluster this node belongs to
	Cluster Cluster `yaml:"cluster"`

	// Flags contains boolean state indicators
	Flags Flags `yaml:"flags"`
}

// NewState creates and returns a new State instance with default values.
// This is used when initializing a fresh node state.
//
// Returns:
//   A pointer to a new State with version set to "1.0.0"
//
// Example Output:
//   &State{
//     Version: "1.0.0",
//     Node: {...},      // empty struct fields
//     Cluster: {...},   // empty struct fields
//     Flags: {...},     // empty struct fields
//   }
func NewState() *State {
	return &State{
		Version: "1.0.0",
	}
}

// Initialize persists the given state to disk, effectively initializing the node.
// This function should only be called once when the node joins a cluster for the first time.
//
// Parameters:
//   initS - The initial state to persist (contains cluster ID, node info, etc.)
//
// Returns:
//   - current: The state that was persisted
//   - err: An error if the node is already initialized or if file operations fail
//
// Example Input:
//   initS = &State{
//     Version: "1.0.0",
//     Node: {
//       ID: "node-123",
//       Hostname: "server1",
//       IP: "192.168.1.10",
//       Role: "leader",
//       Status: "online",
//       InitializedAt: time.Now(),
//     },
//     Cluster: {
//       ID: "cluster-abc",
//       AdvertiseAddr: "192.168.1.10:8443",
//     },
//     Flags: {
//       Initialized: true,
//     },
//   }
//
// Example Output (Success):
//   current = initS (same as input)
//   err = nil
//
// Example Output (Already Initialized):
//   current = nil
//   err = "node already initialized"
//
// Side Effect:
//   Creates a YAML file at cfg.StatePath containing:
//   version: 1.0.0
//   node:
//     id: node-123
//     hostname: server1
//     ip: 192.168.1.10
//     role: leader
//     status: online
//     initialized_at: 2025-12-29T10:30:00Z
//   cluster:
//     id: cluster-abc
//     advertise_addr: 192.168.1.10:8443
//   flags:
//     initialized: true
func (s *State) Initialize(initS *State) (current *State, err error) {
	// Load configuration to get the state file path
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Check if state file already exists (node already initialized)
	// os.Stat returns an error if file doesn't exist, which is what we want
	_, err = os.Stat(cfg.StatePath)
	if err == nil {
		// File exists - node is already initialized
		return nil, errors.New("node already initialized")
	}

	// Create new state file
	file, err := os.Create(cfg.StatePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Serialize state to YAML format
	data, err := yaml.Marshal(initS)
	if err != nil {
		return nil, err
	}

	// Write YAML data to file
	_, err = file.Write(data)
	if err != nil {
		return nil, err
	}

	return initS, nil
}

// LoadState reads and deserializes the node's state from disk.
// This function is used to restore the node's state after a restart.
//
// Returns:
//   - A pointer to the loaded State
//   - An error if the file doesn't exist, can't be read, or contains invalid YAML
//
// Example Input:
//   State file at cfg.StatePath contains:
//   version: 0.1.0
//   node:
//     id: node-123
//     hostname: server1
//     ip: 192.168.1.10
//     role: leader
//     status: online
//     initialized_at: 2025-12-29T10:30:00Z
//   cluster:
//     id: cluster-abc
//     advertise_addr: 192.168.1.10:8443
//   flags:
//     initialized: true
//
// Example Output (Success):
//   &State{
//     Version: "0.1.0",
//     Node: {
//       ID: "node-123",
//       Hostname: "server1",
//       IP: "192.168.1.10",
//       Role: "leader",
//       Status: "online",
//       InitializedAt: time.Parse(..., "2025-12-29T10:30:00Z"),
//     },
//     Cluster: {
//       ID: "cluster-abc",
//       AdvertiseAddr: "192.168.1.10:8443",
//     },
//     Flags: {
//       Initialized: true,
//     },
//   }
//   err = nil
//
// Example Output (File Not Found):
//   state = nil
//   err = "open /path/to/state.yaml: no such file or directory"
func LoadState() (*State, error) {
	// Load configuration to get the state file path
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Read the entire state file into memory
	data, err := os.ReadFile(cfg.StatePath)
	if err != nil {
		return nil, err
	}

	// Deserialize YAML data into State struct
	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	
	return &s, nil
}

// SaveState updates the state file on disk with the provided state data.
// This function is used to persist state changes after modifications (e.g., status updates, role changes).
// Unlike Initialize, this function overwrites an existing state file.
//
// Parameters:
//   data - The updated state to persist
//
// Returns:
//   - true if the state was successfully saved
//   - false if any error occurred during the save operation
//
// Example Input:
//   data = State{
//     Version: "0.1.0",
//     Node: {
//       ID: "node-123",
//       Hostname: "server1",
//       IP: "192.168.1.10",
//       Role: "worker",        // Changed from "leader" to "worker"
//       Status: "offline",     // Changed from "online" to "offline"
//       InitializedAt: time.Parse(..., "2025-12-29T10:30:00Z"),
//     },
//     Cluster: {
//       ID: "cluster-abc",
//       AdvertiseAddr: "192.168.1.10:8443",
//     },
//     Flags: {
//       Initialized: true,
//     },
//   }
//
// Example Output (Success):
//   return true
//
// Example Output (Error - Config Load Failed):
//   return false
//
// Example Output (Error - File Write Failed):
//   return false
//
// Side Effect:
//   Overwrites the YAML file at cfg.StatePath with:
//   version: 0.1.0
//   node:
//     id: node-123
//     hostname: server1
//     ip: 192.168.1.10
//     role: worker
//     status: offline
//     initialized_at: 2025-12-29T10:30:00Z
//   cluster:
//     id: cluster-abc
//     advertise_addr: 192.168.1.10:8443
//   flags:
//     initialized: true
func (s *State) SaveState(data State) (success bool, err error) {
	// Load configuration to get the state file path
	cfg, err := config.Load()
	if err != nil {
		return false, err
	}

	// Create or overwrite the state file
	file, err := os.Create(cfg.StatePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Serialize state to YAML format
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return false, err
	}

	// Write YAML data to file
	_, err = file.Write(yamlData)
	if err != nil {
		return false, err
	}

	return true, nil
}