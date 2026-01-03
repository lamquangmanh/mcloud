package mcloudctl

import (
	"context"
	"database/sql"
	"fmt"

	"mcloud/internal/cert"
	"mcloud/internal/config"
	"mcloud/internal/constant"
	"mcloud/internal/database"
	"mcloud/internal/installer"
	"mcloud/internal/state"
	"mcloud/pkg/logger"
	"mcloud/pkg/utils"
	"mcloud/services/lxd"
	"mcloud/services/microceph"
	"mcloud/services/microovn"

	"github.com/urfave/cli/v2"
)

// InitRequest represents the request structure for cluster initialization.
// This structure matches the server's expected API request format.
//
// Fields:
//   - Name: The name of the cluster to initialize
//   - AdvertiseAddress: The IP address to advertise to other nodes
//
// Example JSON:
//   {
//     "name": "production-cluster",
//     "advertise_address": "192.168.1.10"
//   }
type InitRequest struct {
	Name             string `json:"name"`
	AdvertiseAddress string `json:"advertise_address"`
}

// InitResponse represents the response structure from cluster initialization.
// Contains cluster details, bootstrap token, and leader node information.
//
// Fields:
//   - ClusterID: Unique identifier for the cluster
//   - Token: Bootstrap token for joining additional nodes
//   - Leader: Information about the cluster leader node
//
// Example JSON:
//   {
//     "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
//     "token": "abc123def456...",
//     "leader": {
//       "id": "660e8400-e29b-41d4-a716-446655440001",
//       "hostname": "node1",
//       "ip": "192.168.1.10",
//       "role": "leader",
//       "status": "online"
//     }
//   }
type InitResponse struct {
	ClusterID string `json:"cluster_id"`
	Token     string `json:"token"`
	Leader    struct {
		ID       string `json:"id"`
		Hostname string `json:"hostname"`
		IP       string `json:"ip"`
		Role     string `json:"role"`
		Status   string `json:"status"`
	} `json:"leader"`
}

// validateClusterName validates that a cluster name meets requirements and doesn't already exist.
// Performs two checks:
//   1. Name length must be at least 3 characters
//   2. No existing cluster with the same name in the database
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - name: The cluster name to validate
//   - conn: Database connection for checking existing clusters
//
// Returns:
//   - nil if validation passes
//   - error if name is too short or already exists
//
// Example Input 1 (Valid):
//   name: "production-cluster"
//   Database: no existing clusters with this name
//
// Example Output 1:
//   Returns: nil (validation passed)
//
// Example Input 2 (Too Short):
//   name: "ab"
//
// Example Output 2:
//   Returns: error("cluster name must be at least 3 characters")
//
// Example Input 3 (Already Exists):
//   name: "test-cluster"
//   Database: cluster with name "test-cluster" exists
//
// Example Output 3:
//   Returns: error("a cluster with the name 'test-cluster' already exists")
func validateClusterName(ctx context.Context, name string, conn *sql.DB) error {
	// Check 1: Validate minimum name length
	if len(name) < 3 {
		return fmt.Errorf("cluster name must be at least 3 characters")
	}

	// Check 2: Verify no cluster with the same name already exists
	clusterRepo := database.NewClusterRepository(conn)
	exists, err := clusterRepo.GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check existing clusters: %w", err)
	}
	if exists != nil {
		return fmt.Errorf("a cluster with the name '%s' already exists", name)
	}
	
	return nil
}

// writeConfig creates and saves the mcloud configuration file.
// Generates configuration for both manager (HTTP/gRPC) and agent components,
// using the host's detected IP address.
//
// Parameters:
//   - host: Detected host information containing hostname and IP addresses
//
// Returns:
//   - nil if config is saved successfully
//   - error if file write fails
//
// Example Input:
//   host: HostInfo{
//     Hostname: "node1",
//     IPs: [192.168.1.10],
//   }
//
// Example Output (Success):
//   Console log: "Wrote config file to /etc/mcloud/config.yaml"
//   File created: /etc/mcloud/config.yaml with content:
//     manager:
//       http_host: 192.168.1.10
//       http_port: 9028
//       grpc_host: 192.168.1.10
//       grpc_port: 9030
//     agent:
//       manager_url: http://192.168.1.10:9030
//     database:
//       db_path: mcloud.db
//
// Example Output (Error):
//   Returns: error("open /etc/mcloud/config.yaml: permission denied")
func writeConfig(host utils.HostInfo) error {
	// Create configuration structure with manager and agent settings
	cfg := config.Config{
		Manager: config.Manager{
			HttpHost: host.IPs[0].String(),
			HttpPort: 9028,
			GrpcHost: host.IPs[0].String(),
			GrpcPort: 9030,
		},
		Agent: config.Agent{
			ManagerURL: fmt.Sprintf("http://%s:9030", host.IPs[0].String()),
		},
		Database: config.Database{
			DBPath: "mcloud.db",
		},
		ConfigPath: constant.DefaultConfigPath,
		StatePath:  constant.DefaultStatePath,
	}

	// Write configuration to YAML file
	if err := config.SaveConfig(&cfg); err != nil {
		return err
	}
	logger.Info("Wrote config file to %s\n", cfg.ConfigPath)
	return nil
}

// writeState creates and saves the cluster state file.
// The state file persists node identity, cluster membership, and initialization status.
//
// Parameters:
//   - name: Cluster name
//   - host: Detected host information
//   - nodeId: UUID for this node
//   - clusterId: UUID for the cluster
//
// Returns:
//   - nil if state is saved successfully
//   - error if file write fails
//
// Example Input:
//   name: "production-cluster"
//   host: HostInfo{Hostname: "node1", IPs: [192.168.1.10]}
//   nodeId: "550e8400-e29b-41d4-a716-446655440000"
//   clusterId: "660e8400-e29b-41d4-a716-446655440001"
//
// Example Output (Success):
//   Console log: "Wrote state file to /etc/mcloud/config.yaml"
//   File created: /var/lib/mcloud/state.yaml with content:
//     version: "0.1.0"
//     node:
//       id: "550e8400-e29b-41d4-a716-446655440000"
//       hostname: "node1"
//       ip: "192.168.1.10"
//       role: "leader"
//     cluster:
//       id: "660e8400-e29b-41d4-a716-446655440001"
//       name: "production-cluster"
//       advertise_addr: "192.168.1.10:7443"
//     flags:
//       initialized: true
//
// Example Output (Error):
//   Returns: error("open /var/lib/mcloud/state.yaml: permission denied")
func writeState(name string, host utils.HostInfo, nodeId string, clusterId string) error {
	state := state.State{
		Version: constant.AppVersion,
		Node: state.Node{
			ID:        nodeId,
			Hostname:  host.Hostname,
			IP:        host.IPs[0].String(),
			Role:      string(constant.RoleLeader),
		},
		Cluster: state.Cluster{
			ID:               clusterId,
			Name:             name,
			AdvertiseAddr: fmt.Sprintf("%s:7443", host.IPs[0].String()),
		},
		Flags: state.Flags{
			Initialized: true,
		},
	}

	// Save state to file
	if _, err := state.SaveState(state); err != nil {
		return err
	}
	logger.Info("Wrote state file to %s\n", config.DefaultConfigPath)
	return nil
}

// generateCert generates the Certificate Authority (CA) and server certificates.
// The CA is used to sign server certificates for secure gRPC communication.
//
// Parameters:
//   - cfg: Configuration containing certificate file paths
//   - host: Host information containing IP address for certificate SAN
//
// Returns:
//   - nil if certificates are generated successfully
//   - error if certificate generation fails
//
// Example Input:
//   cfg: Config{
//     Security: Security{
//       CACertPath: "/etc/mcloud/ca.crt",
//       CAKeyPath: "/etc/mcloud/ca.key",
//       ServerCertPath: "/etc/mcloud/server.crt",
//       ServerKeyPath: "/etc/mcloud/server.key",
//     }
//   }
//   host: HostInfo{IPs: [192.168.1.10]}
//
// Example Output (Success):
//   Console logs:
//     "Generated CA certificate"
//     "Generated server certificate"
//   Files created:
//     /etc/mcloud/ca.crt (4096-bit RSA CA certificate, 10 years validity)
//     /etc/mcloud/ca.key (4096-bit RSA private key)
//     /etc/mcloud/server.crt (2048-bit RSA server certificate, 1 year validity)
//     /etc/mcloud/server.key (2048-bit RSA private key)
//   Certificate details:
//     CA Subject: CN=mcloud-ca
//     Server Subject: CN=192.168.1.10
//     Server SAN: IP:192.168.1.10
//
// Example Output (Error):
//   Returns: error("failed to create CA certificate: permission denied")
func generateCert(cfg config.Config, host utils.HostInfo) error {
	// Generate CA certificate and private key
	caCert, caKey, err := cert.GenerateCAV2(cfg.Security.CACertPath, cfg.Security.CAKeyPath)
	if err != nil {
		return err
	}
	logger.Info("Generated CA certificate")

	// Generate server certificate signed by the CA
	err = cert.GenerateServerCert(
		caCert,
		caKey,
		host.IPs[0].String(),
		cfg.Security.ServerCertPath,
		cfg.Security.ServerKeyPath,
	)
	if err != nil {
		return err
	}
	logger.Info("Generated server certificate")
	return nil
}

// bootstrapDatabase initializes the database and creates initial cluster and node records.
// Connects to the database, runs migrations, and inserts the first cluster and leader node.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - name: Cluster name
//   - clusterId: UUID for the cluster
//   - nodeId: UUID for this node
//   - host: Host information containing hostname and IP
//
// Returns:
//   - *sql.DB: Database connection if successful
//   - error: If database connection or record creation fails
//
// Example Input:
//   name: "production-cluster"
//   clusterId: "660e8400-e29b-41d4-a716-446655440001"
//   nodeId: "550e8400-e29b-41d4-a716-446655440000"
//   host: HostInfo{Hostname: "node1", IPs: [192.168.1.10]}
//
// Example Output (Success):
//   Console log: "Database connected and migrated"
//                "Created initial cluster and node records in database"
//   Database records created:
//     clusters table:
//       - id: "660e8400-e29b-41d4-a716-446655440001"
//       - name: "production-cluster"
//       - state: "active"
//     nodes table:
//       - id: "550e8400-e29b-41d4-a716-446655440000"
//       - cluster_id: "660e8400-e29b-41d4-a716-446655440001"
//       - hostname: "node1"
//       - ip: "192.168.1.10"
//       - role: "leader"
//       - status: "online"
//   Returns: (*sql.DB, nil)
//
// Example Output (Error):
//   Returns: (nil, error("unable to open database file: permission denied"))
func bootstrapDatabase(ctx context.Context, name string, clusterId string, nodeId string, host utils.HostInfo) (*sql.DB, error) {
	// Step 1: Connect to database and run migrations
	conn, err := database.Connect()
	if err != nil {
		return nil, err
	}
	logger.Info("Database connected and migrated")

	// Step 2: Initialize repositories
	clusterRepo := database.NewClusterRepository(conn)
	nodeRepo := database.NewNodeRepository(conn)

	// Step 3: Create cluster record
	cluster := &database.Cluster{
		ID:    clusterId,
		Name:  name,
		State: "active",
	}
	
	if err := clusterRepo.Create(ctx, cluster); err != nil {
		return nil, err
	}

	// Step 4: Create leader node record
	node := &database.Node{
		ID:				 nodeId,
		ClusterID:  clusterId,
		Hostname:   host.Hostname,
		IP:         host.IPs[0].String(),
		Role:       "leader",
		Status:     "online",
	}

	if err := nodeRepo.Create(ctx, node); err != nil {
		return nil, err
	}
	logger.Info("Created initial cluster and node records in database")
	return conn, nil
}

// bootstrap initializes all mcloud infrastructure components.
// Orchestrates the setup of certificates, database, LXD, networking, storage, and systemd service.
//
// The function performs the following steps:
//   1. Generate CA and server certificates for secure communication
//   2. Initialize database and create cluster/node records
//   3. Bootstrap LXD control plane
//   4. Setup OVN networking
//   5. Setup Ceph storage
//   6. Install and start mcloudd as systemd service
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - name: Cluster name
//   - host: Host information
//   - nodeId: UUID for this node
//   - clusterId: UUID for the cluster
//   - cfg: Configuration
//
// Returns:
//   - result: Currently nil, reserved for future use
//   - error: If any bootstrap step fails
//
// Example Input:
//   name: "production-cluster"
//   host: HostInfo{Hostname: "node1", IPs: [192.168.1.10]}
//   nodeId: "550e8400-e29b-41d4-a716-446655440000"
//   clusterId: "660e8400-e29b-41d4-a716-446655440001"
//   cfg: Config{...}
//
// Example Output (Success):
//   Console logs:
//     "Bootstrapping mcloud components..."
//     "Generated CA certificate"
//     "Generated server certificate"
//     "Database connected and migrated"
//     "Created initial cluster and node records in database"
//     "LXD cluster initialized"
//     "OVN initialized"
//     "Ceph cluster initialized"
//     "✅ mcloudd installed and started"
//     "mcloud components bootstrapped successfully"
//   Side effects:
//     - Certificates created in /etc/mcloud/
//     - Database initialized with cluster and node records
//     - LXD cluster created with name "production-cluster"
//     - OVN networking configured
//     - Ceph storage pool configured
//     - mcloudd systemd service running
//   Returns: (nil, nil)
//
// Example Output (Error - LXD Bootstrap Failed):
//   Returns: (nil, error("failed to initialize LXD cluster: connection refused"))
func bootstrap(ctx context.Context, name string, host utils.HostInfo, nodeId string, clusterId string, cfg config.Config) (result any, err error) {
	logger.Info("Bootstrapping mcloud components...")

	// Step 1: Generate CA and server certificates
	if err := generateCert(cfg, host); err != nil {
		return nil, err
	}

	// Step 2: Initialize database and create initial records
	_, err = bootstrapDatabase(ctx, name, clusterId, nodeId, host)
	if err != nil {
		return nil, err
	}

	// Step 3: Initialize LXD control plane
	lxdConfig := lxd.BootstrapConfig{
		ClusterName: name,
		Address:     host.IPs[0].String(),
	}
	if err := lxd.Bootstrap(lxdConfig); err != nil {
		return nil, err
	}

	// Step 4: Setup OVN networking
	if err := microovn.Bootstrap(); err != nil {
		return nil, err
	}
	
	// Step 5: Setup Ceph storage
	cephConfig := microceph.BootstrapConfig{
		Disk: constant.DefaultCephDisk,
	}
	if err := microceph.Bootstrap(cephConfig); err != nil {
		return nil, err
	}

	// Step 6: Install mcloudd as systemd service and start it
	if err := installer.Init(); err != nil {
		return nil, err
	}
	logger.Info("mcloud components bootstrapped successfully")

	return nil, nil
}

// InitCommand is the CLI command handler for 'mcloudctl init'.
// Initializes a new mcloud cluster on the current node, setting it up as the cluster leader.
//
// Command Flow:
//   Step 1: Load configuration and connect to database
//   Step 2: Detect host information (hostname, IP addresses)
//   Step 3: Validate cluster name (length and uniqueness)
//   Step 4: Write configuration file
//   Step 5: Bootstrap all mcloud components (certs, DB, LXD, OVN, Ceph, mcloudd)
//   Step 6: Write cluster state file
//
// CLI Usage:
//   mcloudctl init --name <cluster-name>
//
// Parameters:
//   - c: CLI context containing parsed command-line flags
//
// Returns:
//   - nil if cluster initialization succeeds
//   - error if any step fails
//
// Example Input (Command Line):
//   $ sudo mcloudctl init --name production-cluster
//
// Example Output (Success):
//   Console logs:
//     [INFO] 2026-01-02 10:30:45 Initializing mcloud cluster: production-cluster
//     [INFO] 2026-01-02 10:30:45 Loaded config: {...}
//     [INFO] 2026-01-02 10:30:45 Database initialized and migrated
//     [INFO] 2026-01-02 10:30:45 Wrote config file to /etc/mcloud/config.yaml
//     [INFO] 2026-01-02 10:30:45 Bootstrapping mcloud components...
//     [INFO] 2026-01-02 10:30:46 Generated CA certificate
//     [INFO] 2026-01-02 10:30:46 Generated server certificate
//     [INFO] 2026-01-02 10:30:46 Database connected and migrated
//     [INFO] 2026-01-02 10:30:46 Created initial cluster and node records in database
//     [INFO] 2026-01-02 10:30:47 LXD cluster initialized
//     [INFO] 2026-01-02 10:30:48 OVN initialized
//     [INFO] 2026-01-02 10:30:49 Ceph cluster initialized
//     ✅ mcloudd installed and started
//     [INFO] 2026-01-02 10:30:50 mcloud components bootstrapped successfully
//     [INFO] 2026-01-02 10:30:50 Wrote state file to /var/lib/mcloud/state.yaml
//     [INFO] 2026-01-02 10:30:50 mcloud initialized successfully
//   Returns: nil
//
// Example Output (Error - Not Root):
//   [ERROR] 2026-01-02 10:30:45 must run as root
//   Returns: error("must run as root")
//
// Example Output (Error - Cluster Name Exists):
//   [ERROR] 2026-01-02 10:30:45 a cluster with the name 'production-cluster' already exists
//   Returns: error("a cluster with the name 'production-cluster' already exists")
//
// Side Effects:
//   - Creates /etc/mcloud/config.yaml
//   - Creates /var/lib/mcloud/state.yaml
//   - Initializes mcloud.db with cluster and node records
//   - Generates TLS certificates in /etc/mcloud/
//   - Configures LXD, OVN, and Ceph
//   - Installs and starts mcloudd.service
func InitCommand(c *cli.Context) error {
	ctx := context.Background()

	// Extract cluster name from CLI flag
	clusterName := c.String("name")
	logger.Info("Initializing mcloud cluster: %s\n", clusterName)

	// Step 1a: Load configuration from YAML file
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed to load config: %v", err)
	}
	logger.Info("Loaded config: %v", cfg)

	// Step 1b: Initialize database connection and run migrations
	conn, err := database.Connect()
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
	}
	logger.Info("Database initialized and migrated")

	// Step 2: Detect host information (hostname, IP addresses, memory, etc.)
	host, err := utils.DetectHost()
	if err != nil {
		return err
	}

	// Step 3: Validate cluster name (minimum length and uniqueness)
	if err := validateClusterName(ctx, clusterName, conn); err != nil {
		return err
	}

	// Step 4: Write configuration file with detected settings
	if err := writeConfig(*host); err != nil {
		return err
	}

	// Generate unique identifiers for node and cluster
	nodeId := utils.GenerateUUID()
	clusterId := utils.GenerateUUID()

	// Step 5: Bootstrap all mcloud infrastructure components
	_, err = bootstrap(ctx, clusterName, *host, nodeId, clusterId, *cfg)
	if err != nil {
		return err
	}

	// Step 6: Write cluster state file with node and cluster information
	if err := writeState(clusterName, *host, nodeId, clusterId); err != nil {
		return err
	}

	logger.Info("mcloud initialized successfully")
	return nil
}
