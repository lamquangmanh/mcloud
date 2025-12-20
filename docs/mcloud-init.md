# MCloud Init Feature

## Overview

The `mcloud init` feature allows you to initialize a new MCloud cluster with LXD, Ceph, and OVN networking support. This creates the initial cluster configuration, sets up the leader node, generates certificates, and creates a bootstrap token for other nodes to join.

## Architecture

### Components

1. **mcloudctl** - CLI tool for cluster management
2. **mcloudd** - HTTP server that manages cluster state
3. **SQLite Database** - Stores cluster configuration, nodes, certificates, and tokens

### Flow

```
mcloudctl init → HTTP POST → mcloudd → Database + LXD
                              ↓
                         Cluster Created
                              ↓
                    Bootstrap Token Generated
```

## Usage

### Initialize a Cluster

```bash
mcloudctl init --name <cluster-name> --address <ip:port> [--server <server-url>]
```

**Parameters:**

- `--name` - Name of the cluster (default: "mcloud-cluster")
- `--address` - Advertise address for the cluster (required, e.g., "192.168.1.10:8443")
- `--server` - mcloudd server URL (default: "http://localhost:8080")

**Example:**

```bash
mcloudctl init --name production-cluster --address 192.168.1.100:8443
```

**Output:**

```
✓ Cluster initialized successfully!

Cluster ID:   c7d4c6f0-2881-4190-9508-04d05a1b7596
Cluster Name: production-cluster
Leader Node:  server1 (192.168.1.100:8443)

Bootstrap Token (save this to join other nodes):
  mcloud-c7d4c6f0-yEEoUQOx6QBpQ7dC

To join a node, run:
  mcloudctl join --token mcloud-c7d4c6f0-yEEoUQOx6QBpQ7dC --server http://localhost:8080
```

## What Gets Created

### 1. Cluster Record

- Unique cluster ID (UUID)
- Cluster name
- State: "active"

### 2. Leader Node

- Node ID (UUID)
- Hostname (from system)
- IP address (from --address parameter)
- Role: "leader"
- Status: "online"

### 3. Certificate Authority

- Self-signed CA certificate
- RSA 2048-bit key pair
- Valid for 10 years

### 4. Bootstrap Token

- Secure random token
- Valid for 24 hours
- Used for joining new nodes

### 5. Configuration Store (Key-Value)

**LXD Configuration:**

- `lxd.cluster.name` - Cluster name
- `lxd.cluster.address` - Advertise address

**Ceph Configuration:**

- `ceph.enabled` - Set to "true"
- `ceph.cluster.name` - Ceph cluster name (cluster-name + "-ceph")

**OVN Configuration:**

- `ovn.enabled` - Set to "true"
- `ovn.network.name` - OVN network name (cluster-name + "-ovn")

## Database Schema

The init command creates records in the following tables:

- `clusters` - Cluster information
- `nodes` - Node details (leader node)
- `certificate_authorities` - CA cert and key
- `bootstrap_tokens` - Token for joining
- `kv_store` - Configuration key-value pairs

## Testing

Run the test script to verify the init feature:

```bash
./test-init.sh
```

This script will:

1. Clean up any existing database
2. Build the binaries
3. Start the mcloudd server
4. Run the init command
5. Display the results
6. Clean up

### Manual Testing

```bash
# Start the server
./mcloudd

# In another terminal, initialize the cluster
./mcloudctl init --name my-cluster --address 192.168.1.10:8443

# Inspect the database
sqlite3 mcloud.db 'SELECT * FROM clusters;'
sqlite3 mcloud.db 'SELECT * FROM nodes;'
sqlite3 mcloud.db 'SELECT * FROM bootstrap_tokens;'
sqlite3 mcloud.db 'SELECT * FROM kv_store;'
```

## Configuration

Server configuration is loaded from `internal/config/config.yaml`:

```yaml
manager:
  host: '127.0.0.1'
  port: 9028

agent:
  manager_url: 'http://127.0.0.1:9028'

store:
  db_path: 'mcloud.db'
```

## Implementation Details

### Files Changed/Created

**CLI (`cmd/mcloudctl/`):**

- `init.go` - Init command implementation
- `join.go` - Join command stub
- `main.go` - CLI entry point

**Server (`internal/cluster/`):**

- `service.go` - Init logic with transaction handling
- `handler.go` - HTTP handler for /cluster/init
- `init_module.go` - Route registration

**Database (`internal/database/`):**

- All repository files updated to use `sqlExecutor` interface
- This allows repositories to work with both `*sql.DB` and `*sql.Tx`

**Security (`internal/`):**

- `cert/ca.go` - CA certificate generation (RSA 2048)
- `auth/token.go` - Secure token generation

**LXD Integration (`internal/lxd/`):**

- `client.go` - LXD client interface
- `cluster.go` - Cluster initialization (with graceful fallback)
- `mock_client.go` - Mock client for testing

## Error Handling

The init command handles various error scenarios:

- **Cluster already exists**: Returns 409 error
- **Invalid parameters**: Returns 400 error
- **LXD not available**: Uses mock data (for development)
- **Database errors**: Automatic rollback via transaction
- **Network errors**: Clear error message to user

## Security Considerations

1. **Certificates**: Self-signed CA with RSA 2048-bit encryption
2. **Tokens**: Cryptographically secure random tokens
3. **Token Expiry**: Bootstrap tokens expire after 24 hours
4. **Database**: Uses SQLite with WAL mode for concurrent access

## Future Enhancements

- [ ] JWT-based authentication tokens
- [ ] Certificate rotation
- [ ] Multi-master support
- [ ] Real LXD cluster integration
- [ ] Ceph pool creation
- [ ] OVN network setup
- [ ] Node join implementation
- [ ] Web UI for cluster management

## Troubleshooting

### Database locked error

```
Error: database is locked (5) (SQLITE_BUSY)
```

**Solution:**

1. Kill any existing mcloudd processes: `pkill -f mcloudd`
2. Remove WAL files: `rm -f mcloud.db-shm mcloud.db-wal`
3. Restart the server

### Address already in use

```
Error: bind: address already in use
```

**Solution:**
Kill the existing server process: `pkill -f mcloudd`

### LXD not available

The init command will work even if LXD is not installed. It uses mock data for development purposes.

## References

- [Repository Structure](docs/repo-structure.md)
- [Requirements](requirements/mcloudd.txt)
- Inspired by:
  - [MicroCloud](https://github.com/canonical/microcloud)
  - [LXD](https://github.com/canonical/lxd)
  - [MicroCluster](https://github.com/canonical/microcluster)
