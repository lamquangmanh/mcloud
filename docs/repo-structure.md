# 🧱 Structure repo MCloud (mono-repo)

```bash
mcloud/
├── README.md
├── Makefile
├── go.work
├── go.sum
├── go.mod
│
├── cmd/                        # Entry points (binary)
│   ├── mcloudctl/              # CLI (init / join / leave)
│   │   └── main.go
│   │
│   ├── mcloudd/                # Control Plane (Manager) start http server
│   │   └── main.go
│   │
│   └── mcloud-agent/           # Node Agent
│       └── main.go
│
├── internal/                   # Core logic (NOT reusable)
│   ├── cluster/                # cluster state, membership
│   │   ├── router.go           # routers for cluster
│   │   ├── init.go
│   │   ├── service.go
│   │   ├── join.go
│   │   ├── leave.go
│   │   └── state.go
│   │
│   ├── node/                   # node lifecycle
│   │   ├── register.go
│   │   ├── health.go
│   │   └── drain.go
│   │
│   ├── agent/                  # agent handlers
│   │   ├── join.go
│   │   ├── leave.go
│   │   └── exec.go
│   │
│   ├── controller/             # control loops
│   │   ├── node_controller.go
│   │   └── health_controller.go
│   │
│   ├── cert/                   # CA, cert, rotation
│   │   ├── ca.go
│   │   └── issue.go
│   │
│   ├── storage/                # cluster metadata storage
│   │   └── init.go
│   │
│   ├── store/                  # database
│   │   └── store.go            # First use sqlite, after that change to use dqlite + etcd
│   │
│   ├── lxd/                    # LXD client wrapper
│   │   ├── client.go
│   │   └── cluster.go
│   │
│   ├── auth/                   # token / bootstrap auth
│   │   └── token.go
│   │
│   └── config/
│       ├── config.yaml
│       └── config.go
│
├── pkg/                        # Reusable libs (public)
│   ├── api/                    # API models
│   │   └── types.go
│   │
│   ├── utils/
│   │   └── exec.go
│   │
│   └── logger/
│       └── logger.go
│
├── proto/                      # gRPC definitions
│   ├── agent.proto
│   └── cluster.proto
│
├── web/                        # UI Console
│   ├── README.md
│   ├── package.json
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── api/
│   │   └── layouts/
│   └── tailwind.config.js
│
├── scripts/                    # Dev / install helpers
│   ├── install-agent.sh
│   ├── bootstrap.sh
│   └── dev.sh
│
├── deploy/                     # future: systemd, helm
│   ├── systemd/
│   │   ├── mcloudd.service
│   │   └── mcloud-agent.service
│   └── docker/
│
└── docs/
    ├── architecture.md
    ├── join-flow.md
    └── security.md

```
