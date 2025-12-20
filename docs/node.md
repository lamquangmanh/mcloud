# Node management

```bash
# Init cluster
mcloud init

# join cluster
mcloud join <token>

# leave cluster
mcloud leave

# get current status
mcloud status

# get list nodes of cluster
mcloud cluster list
```

# Design

```bash
┌────────────┐
│  mcloudctl │  (CLI)
└─────┬──────┘
      │ REST / gRPC
      ▼
┌────────────────────┐
│   MCloud Manager   │  (Control Plane)
│   (single leader)  │
└─────┬──────────────┘
      │
      │ TLS + cert
      ▼
┌─────────────────────────────────────┐
│           MCloud Agent               │
│   (run on every node)                │
│  - join / leave                      │
│  - system info                       │
│  - install deps                      │
│  - talk to LXD / K8s                 │
└─────────────────────────────────────┘

```

# Detail

- Init cluster

```bash
mcloud init
 ↓
Manager:
- generate cluster_id
- generate CA
- store cluster state
- mark leader

Sequence

mcloudctl
   |
   | POST /cluster/init
   |
mcloudd (HTTP)
   |
   | cluster.Init()
   |
   |-- validate request
   |-- check no cluster exists
   |
   |-- generate CA
   |-- generate bootstrap token
   |
   |-- lxd.InitCluster()
   |      |
   |      |-- lxd init
   |      |-- set node as leader
   |
   |-- DB TRANSACTION
   |      |-- insert cluster
   |      |-- insert leader node
   |      |-- insert CA
   |      |-- insert token
   |-- COMMIT
   |
   | return result


```

- Join node

```bash
Node B:
mcloud join <token>
 ↓
Agent:
- validate token
- fetch CA cert
- register node
- install LXD
- join LXD cluster

```

- Leave node

```bash
mcloud leave
 ↓
Manager:
- drain node
- remove from LXD cluster
- cleanup cert

```

# Security

- CA (Certificate Authority) for cluster
- Node cert signed by CA
- gRPC mTLS
- The token is used only during the bootstrap process
